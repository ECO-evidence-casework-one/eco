using System.IO.Compression;
using System.Security.Cryptography;
using System.Text;
using System.Xml.Linq;
using Eco.Recovery.Models;

namespace Eco.Recovery.Services;

public sealed class EvidenceIntakeService
{
    private static readonly HashSet<string> PlainTextExtensions = new(StringComparer.OrdinalIgnoreCase)
    {
        ".txt", ".md", ".csv", ".json", ".log", ".xml"
    };

    private static readonly HashSet<string> ImageExtensions = new(StringComparer.OrdinalIgnoreCase)
    {
        ".png", ".jpg", ".jpeg", ".bmp", ".tif", ".tiff"
    };

    private readonly WorkspacePaths _paths;

    public EvidenceIntakeService(WorkspacePaths paths)
    {
        _paths = paths ?? throw new ArgumentNullException(nameof(paths));
    }

    public async Task<EvidenceRecord> PreserveAndExtractAsync(
        string sourcePath,
        MatterWorkspace workspace,
        IProgress<EvidenceImportProgress>? progress = null,
        CancellationToken cancellationToken = default)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(sourcePath);
        ArgumentNullException.ThrowIfNull(workspace);

        var sourceInfo = new FileInfo(Path.GetFullPath(sourcePath));
        if (!sourceInfo.Exists)
        {
            throw new FileNotFoundException("The selected evidence file no longer exists.", sourceInfo.FullName);
        }
        if ((sourceInfo.Attributes & FileAttributes.Directory) != 0)
        {
            throw new IOException("A directory cannot be imported as one evidence item.");
        }

        var evidenceId = "ev-" + Guid.NewGuid().ToString("N");
        var safeName = SanitizeFileName(sourceInfo.Name);
        var stagingRoot = _paths.ImportStagingRoot(workspace.MatterId);
        var finalRoot = _paths.OriginalEvidenceRoot(workspace.MatterId, evidenceId);
        Directory.CreateDirectory(stagingRoot);
        Directory.CreateDirectory(finalRoot);

        var stagingPath = Path.Combine(stagingRoot, evidenceId + ".part");
        var finalPath = Path.Combine(finalRoot, safeName);
        string? hash = null;
        long copiedBytes = 0;

        try
        {
            progress?.Report(new EvidenceImportProgress
            {
                Stage = "Preserving original",
                CompletedBytes = 0,
                TotalBytes = sourceInfo.Length
            });

            await using (var input = new FileStream(
                sourceInfo.FullName,
                FileMode.Open,
                FileAccess.Read,
                FileShare.Read,
                128 * 1024,
                FileOptions.Asynchronous | FileOptions.SequentialScan))
            await using (var output = new FileStream(
                stagingPath,
                FileMode.CreateNew,
                FileAccess.Write,
                FileShare.None,
                128 * 1024,
                FileOptions.Asynchronous | FileOptions.WriteThrough))
            using (var hasher = IncrementalHash.CreateHash(HashAlgorithmName.SHA256))
            {
                var buffer = new byte[128 * 1024];
                while (true)
                {
                    cancellationToken.ThrowIfCancellationRequested();
                    var read = await input.ReadAsync(buffer, cancellationToken);
                    if (read == 0)
                    {
                        break;
                    }
                    await output.WriteAsync(buffer.AsMemory(0, read), cancellationToken);
                    hasher.AppendData(buffer, 0, read);
                    copiedBytes += read;
                    progress?.Report(new EvidenceImportProgress
                    {
                        Stage = "Preserving original",
                        CompletedBytes = copiedBytes,
                        TotalBytes = sourceInfo.Length
                    });
                }

                await output.FlushAsync(cancellationToken);
                output.Flush(flushToDisk: true);
                hash = Convert.ToHexString(hasher.GetHashAndReset()).ToLowerInvariant();
            }

            if (copiedBytes != sourceInfo.Length)
            {
                throw new IOException(
                    $"The selected file changed or ended during preservation. Expected {sourceInfo.Length} bytes; copied {copiedBytes} bytes.");
            }

            File.Move(stagingPath, finalPath);
            var relativePath = _paths.ToMatterRelativePath(workspace.MatterId, finalPath);
            var duplicate = workspace.Evidence.FirstOrDefault(item =>
                item.SizeBytes == copiedBytes && string.Equals(item.Sha256, hash, StringComparison.Ordinal));

            var record = new EvidenceRecord
            {
                Id = evidenceId,
                OriginalName = sourceInfo.Name,
                PreservedRelativePath = relativePath,
                Sha256 = hash,
                SizeBytes = copiedBytes,
                State = EvidenceState.Extracting,
                DuplicateOfEvidenceId = duplicate?.Id
            };

            progress?.Report(new EvidenceImportProgress
            {
                Stage = "Reading preserved copy",
                CompletedBytes = copiedBytes,
                TotalBytes = copiedBytes
            });
            await ExtractAsync(finalPath, record, cancellationToken);
            return record;
        }
        catch (OperationCanceledException)
        {
            TryDelete(stagingPath);
            TryDelete(finalPath);
            TryDeleteEmptyDirectory(finalRoot);
            throw;
        }
        catch
        {
            TryDelete(stagingPath);
            if (hash is null)
            {
                TryDelete(finalPath);
                TryDeleteEmptyDirectory(finalRoot);
            }
            throw;
        }
    }

    public void RollBackPreservedEvidence(string matterId, EvidenceRecord record)
    {
        ArgumentNullException.ThrowIfNull(record);
        var path = _paths.ResolveMatterRelativePath(matterId, record.PreservedRelativePath);
        TryDelete(path);
        TryDeleteEmptyDirectory(Path.GetDirectoryName(path)!);
    }

    private static async Task ExtractAsync(
        string preservedPath,
        EvidenceRecord record,
        CancellationToken cancellationToken)
    {
        var extension = Path.GetExtension(preservedPath);
        try
        {
            if (PlainTextExtensions.Contains(extension))
            {
                var text = await ReadBoundedTextAsync(preservedPath, cancellationToken);
                AddSegments(record, text, ExtractionKind.TextLayer);
                record.ExtractionKind = ExtractionKind.TextLayer;
                record.State = EvidenceState.Ready;
                return;
            }

            if (string.Equals(extension, ".docx", StringComparison.OrdinalIgnoreCase))
            {
                var text = await ReadDocxAsync(preservedPath, cancellationToken);
                AddSegments(record, text, ExtractionKind.OfficeXml);
                record.ExtractionKind = ExtractionKind.OfficeXml;
                record.State = EvidenceState.Ready;
                return;
            }

            if (string.Equals(extension, ".pdf", StringComparison.OrdinalIgnoreCase))
            {
                record.State = EvidenceState.ReaderUnavailable;
                record.FailureCode = "PDF_RUNTIME_NOT_CONNECTED";
                record.FailureDetail = "The preserved PDF is safe, but PDF text/OCR integration has not been connected in this recovery gate.";
                return;
            }

            if (ImageExtensions.Contains(extension))
            {
                record.State = EvidenceState.ReaderUnavailable;
                record.FailureCode = "OCR_RUNTIME_NOT_CONNECTED";
                record.FailureDetail = "The preserved image is safe, but OCR integration has not been connected in this recovery gate.";
                return;
            }

            record.State = EvidenceState.Unsupported;
            record.FailureCode = "UNSUPPORTED_FORMAT";
            record.FailureDetail = $"ECO preserved the original but does not currently read {extension.ToUpperInvariant()} files.";
        }
        catch (InvalidDataException exception)
        {
            record.State = EvidenceState.Damaged;
            record.FailureCode = "DAMAGED_DOCUMENT";
            record.FailureDetail = exception.Message;
        }
        catch (IOException exception)
        {
            record.State = EvidenceState.Damaged;
            record.FailureCode = "DOCUMENT_READ_FAILED";
            record.FailureDetail = exception.Message;
        }
    }

    private static async Task<string> ReadBoundedTextAsync(
        string path,
        CancellationToken cancellationToken)
    {
        const long maximumTextBytes = 32L * 1024 * 1024;
        var info = new FileInfo(path);
        if (info.Length > maximumTextBytes)
        {
            throw new InvalidDataException("The text file is larger than the current 32 MB extraction limit. The original remains preserved.");
        }

        await using var stream = new FileStream(
            path,
            FileMode.Open,
            FileAccess.Read,
            FileShare.Read,
            64 * 1024,
            FileOptions.Asynchronous | FileOptions.SequentialScan);
        using var reader = new StreamReader(
            stream,
            Encoding.UTF8,
            detectEncodingFromByteOrderMarks: true,
            bufferSize: 64 * 1024,
            leaveOpen: false);
        return await reader.ReadToEndAsync(cancellationToken);
    }

    private static async Task<string> ReadDocxAsync(
        string path,
        CancellationToken cancellationToken)
    {
        await using var stream = new FileStream(
            path,
            FileMode.Open,
            FileAccess.Read,
            FileShare.Read,
            64 * 1024,
            FileOptions.Asynchronous | FileOptions.RandomAccess);
        using var archive = new ZipArchive(stream, ZipArchiveMode.Read, leaveOpen: false);
        var entry = archive.GetEntry("word/document.xml")
            ?? throw new InvalidDataException("The DOCX package has no word/document.xml part.");
        if (entry.Length > 64L * 1024 * 1024)
        {
            throw new InvalidDataException("The DOCX text part exceeds the current extraction limit.");
        }

        await using var documentStream = entry.Open();
        var document = await XDocument.LoadAsync(
            documentStream,
            LoadOptions.PreserveWhitespace,
            cancellationToken);
        XNamespace word = "http://schemas.openxmlformats.org/wordprocessingml/2006/main";
        var paragraphs = document
            .Descendants(word + "p")
            .Select(paragraph => string.Concat(paragraph.Descendants(word + "t").Select(node => node.Value)).Trim())
            .Where(text => text.Length > 0);
        var result = string.Join(Environment.NewLine + Environment.NewLine, paragraphs);
        if (result.Length == 0)
        {
            throw new InvalidDataException("The DOCX package contains no readable paragraph text.");
        }
        return result;
    }

    private static void AddSegments(
        EvidenceRecord record,
        string text,
        ExtractionKind kind)
    {
        var normalised = text.Replace("\r\n", "\n", StringComparison.Ordinal).Replace('\r', '\n');
        var paragraphs = normalised
            .Split("\n\n", StringSplitOptions.TrimEntries | StringSplitOptions.RemoveEmptyEntries)
            .SelectMany(paragraph => Chunk(paragraph, 1_500))
            .ToList();

        for (var index = 0; index < paragraphs.Count; index++)
        {
            var value = paragraphs[index].Trim();
            if (value.Length == 0)
            {
                continue;
            }
            record.Segments.Add(new EvidenceSegment
            {
                Id = record.Id + "-seg-" + (index + 1).ToString("D4", System.Globalization.CultureInfo.InvariantCulture),
                EvidenceId = record.Id,
                SourceSha256 = record.Sha256,
                Text = value,
                Location = $"paragraph {index + 1}",
                Page = 1,
                Searchable = true,
                ExtractionKind = kind
            });
        }

        if (record.Segments.Count == 0)
        {
            record.FailureCode = "NO_READABLE_TEXT";
            record.FailureDetail = "The preserved document contains no readable text.";
        }
    }

    private static IEnumerable<string> Chunk(string input, int maximumLength)
    {
        var remaining = input.Trim();
        while (remaining.Length > maximumLength)
        {
            var split = remaining.LastIndexOfAny(['.', '!', '?', ';', ':', ' '], maximumLength - 1, maximumLength);
            if (split < maximumLength / 2)
            {
                split = maximumLength;
            }
            yield return remaining[..split].Trim();
            remaining = remaining[split..].TrimStart();
        }
        if (remaining.Length > 0)
        {
            yield return remaining;
        }
    }

    private static string SanitizeFileName(string name)
    {
        var invalid = Path.GetInvalidFileNameChars().ToHashSet();
        var builder = new StringBuilder(name.Length);
        foreach (var character in name)
        {
            builder.Append(invalid.Contains(character) || char.IsControl(character) ? '_' : character);
        }
        var result = builder.ToString().Trim();
        return string.IsNullOrWhiteSpace(result) ? "evidence.bin" : result;
    }

    private static void TryDelete(string path)
    {
        try
        {
            if (File.Exists(path))
            {
                File.Delete(path);
            }
        }
        catch
        {
            // The original failure remains authoritative; cleanup will be retried on recovery.
        }
    }

    private static void TryDeleteEmptyDirectory(string path)
    {
        try
        {
            if (Directory.Exists(path) && !Directory.EnumerateFileSystemEntries(path).Any())
            {
                Directory.Delete(path);
            }
        }
        catch
        {
            // Best-effort cleanup only.
        }
    }
}
