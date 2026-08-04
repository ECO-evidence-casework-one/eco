using System.IO.Compression;
using System.Security.Cryptography;
using System.Text;
using Eco.Recovery.Models;
using Eco.Recovery.Services;

namespace Eco.Recovery.Tests;

internal static class Program
{
    private static readonly List<(string Name, Func<Task> Test)> Tests =
    [
        ("Matter repository rejects stale writers", MatterRepositoryRejectsStaleWriters),
        ("Text evidence is preserved, hashed and segmented", TextEvidenceIsPreservedHashedAndSegmented),
        ("Duplicate imports preserve separate occurrences", DuplicateImportsPreserveOccurrences),
        ("DOCX text is extracted from preserved copy", DocxTextIsExtracted),
        ("Unsupported evidence remains preserved but unsearchable", UnsupportedEvidenceRemainsPreserved),
        ("Damaged DOCX remains preserved but unsearchable", DamagedDocxRemainsPreserved),
        ("Cancelled import leaves no evidence or preserved copy", CancelledImportLeavesNoEvidence),
        ("Matter roots isolate same-named evidence", MatterRootsAreIsolated),
        ("Save conflict rolls preserved evidence back", SaveConflictRollsPreservedEvidenceBack)
    ];

    private static async Task<int> Main()
    {
        var failures = new List<string>();
        foreach (var (name, test) in Tests)
        {
            try
            {
                await test();
                Console.WriteLine($"PASS\t{name}");
            }
            catch (Exception exception)
            {
                failures.Add(name);
                Console.Error.WriteLine($"FAIL\t{name}\n{exception}");
            }
        }

        Console.WriteLine($"RESULT\t{Tests.Count - failures.Count}/{Tests.Count} passed");
        return failures.Count == 0 ? 0 : 1;
    }

    private static async Task MatterRepositoryRejectsStaleWriters()
    {
        using var scope = new TempScope();
        var paths = new WorkspacePaths(scope.Root);
        var repository = new MatterRepository(paths);
        var created = await repository.CreateAsync("matter-a", "Matter A");
        Equal(1L, created.Revision, "created revision");

        var writerA = await RequiredLoad(repository, "matter-a");
        var writerB = await RequiredLoad(repository, "matter-a");
        writerA.MatterName = "Writer A saved";
        await repository.SaveAsync(writerA, expectedRevision: 1);
        Equal(2L, writerA.Revision, "writer A revision");

        writerB.MatterName = "Writer B stale overwrite";
        await ThrowsAsync<WorkspaceConflictException>(() => repository.SaveAsync(writerB, expectedRevision: 1));
        var final = await RequiredLoad(repository, "matter-a");
        Equal("Writer A saved", final.MatterName, "stale writer did not replace state");
        Equal(2L, final.Revision, "stored revision after conflict");
    }

    private static async Task TextEvidenceIsPreservedHashedAndSegmented()
    {
        using var scope = new TempScope();
        var fixture = scope.WriteText("funding-email.txt", "First grant instalment: £620.\n\nFinal certificate due by 22 March 2026.");
        var context = await CreateMatterContext(scope.Root, "matter-a");

        var record = await context.Coordinator.ImportAsync("matter-a", fixture);
        Equal(EvidenceState.Ready, record.State, "text state");
        Equal(ExtractionKind.TextLayer, record.ExtractionKind, "text extraction kind");
        True(record.Segments.Count == 2, "two text paragraphs should be segmented");
        True(record.IsSearchable, "ready text evidence should be searchable");
        Equal(await Sha256(fixture), record.Sha256, "preserved SHA-256");

        var preserved = context.Paths.ResolveMatterRelativePath("matter-a", record.PreservedRelativePath);
        True(File.Exists(preserved), "preserved original exists");
        Equal(File.ReadAllBytes(fixture), File.ReadAllBytes(preserved), "preserved bytes match source");
        var stored = await RequiredLoad(context.Repository, "matter-a");
        Equal(1, stored.Evidence.Count, "stored evidence count");
        Equal(2L, stored.Revision, "revision advanced with import");
    }

    private static async Task DuplicateImportsPreserveOccurrences()
    {
        using var scope = new TempScope();
        var fixture = scope.WriteText("quotation.txt", "Quotation total £1,240 including labour.");
        var context = await CreateMatterContext(scope.Root, "matter-a");

        var first = await context.Coordinator.ImportAsync("matter-a", fixture);
        var second = await context.Coordinator.ImportAsync("matter-a", fixture);
        Equal(first.Id, second.DuplicateOfEvidenceId, "duplicate link");
        Equal(first.Sha256, second.Sha256, "duplicate hash");
        True(!string.Equals(first.PreservedRelativePath, second.PreservedRelativePath, StringComparison.Ordinal),
            "duplicate occurrence has a separate preserved path");
        True(File.Exists(context.Paths.ResolveMatterRelativePath("matter-a", first.PreservedRelativePath)),
            "first preserved copy exists");
        True(File.Exists(context.Paths.ResolveMatterRelativePath("matter-a", second.PreservedRelativePath)),
            "second preserved copy exists");
        var stored = await RequiredLoad(context.Repository, "matter-a");
        Equal(2, stored.Evidence.Count, "both occurrences are stored");
    }

    private static async Task DocxTextIsExtracted()
    {
        using var scope = new TempScope();
        var fixture = Path.Combine(scope.Root, "caretaker-note.docx");
        CreateMinimalDocx(fixture, "Engineers returned on 17 March 2026.", "Certificate not yet received.");
        var context = await CreateMatterContext(scope.Root, "matter-a");

        var record = await context.Coordinator.ImportAsync("matter-a", fixture);
        Equal(EvidenceState.Ready, record.State, "DOCX state");
        Equal(ExtractionKind.OfficeXml, record.ExtractionKind, "DOCX extraction kind");
        True(record.Segments.Any(segment => segment.Text.Contains("17 March 2026", StringComparison.Ordinal)),
            "DOCX date text extracted");
        True(record.Segments.All(segment => segment.SourceSha256 == record.Sha256),
            "DOCX provenance hash retained on every segment");
    }

    private static async Task UnsupportedEvidenceRemainsPreserved()
    {
        using var scope = new TempScope();
        var fixture = scope.WriteText("unknown.xyz", "synthetic unsupported content");
        var context = await CreateMatterContext(scope.Root, "matter-a");

        var record = await context.Coordinator.ImportAsync("matter-a", fixture);
        Equal(EvidenceState.Unsupported, record.State, "unsupported state");
        True(!record.IsSearchable, "unsupported evidence is not searchable");
        Equal(0, record.Segments.Count, "unsupported evidence has no extracted segments");
        True(File.Exists(context.Paths.ResolveMatterRelativePath("matter-a", record.PreservedRelativePath)),
            "unsupported original remains preserved");
    }

    private static async Task DamagedDocxRemainsPreserved()
    {
        using var scope = new TempScope();
        var fixture = Path.Combine(scope.Root, "damaged.docx");
        await File.WriteAllBytesAsync(fixture, Encoding.UTF8.GetBytes("not a zip package"));
        var context = await CreateMatterContext(scope.Root, "matter-a");

        var record = await context.Coordinator.ImportAsync("matter-a", fixture);
        Equal(EvidenceState.Damaged, record.State, "damaged DOCX state");
        True(!record.IsSearchable, "damaged evidence is not searchable");
        True(File.Exists(context.Paths.ResolveMatterRelativePath("matter-a", record.PreservedRelativePath)),
            "damaged original remains preserved");
    }

    private static async Task CancelledImportLeavesNoEvidence()
    {
        using var scope = new TempScope();
        var fixture = Path.Combine(scope.Root, "large.bin");
        await WriteLargeFixture(fixture, 8 * 1024 * 1024);
        var context = await CreateMatterContext(scope.Root, "matter-a");
        using var cancellation = new CancellationTokenSource();
        var progress = new SynchronousProgress<EvidenceImportProgress>(value =>
        {
            if (value.Stage == "Preserving original" && value.CompletedBytes > 0)
            {
                cancellation.Cancel();
            }
        });

        await ThrowsAsync<OperationCanceledException>(() =>
            context.Coordinator.ImportAsync("matter-a", fixture, progress, cancellation.Token));
        var stored = await RequiredLoad(context.Repository, "matter-a");
        Equal(0, stored.Evidence.Count, "cancelled import did not enter Matter state");
        var evidenceRoot = context.Paths.EvidenceRoot("matter-a");
        var preservedFiles = Directory.Exists(evidenceRoot)
            ? Directory.EnumerateFiles(evidenceRoot, "*", SearchOption.AllDirectories)
                .Where(path => !path.EndsWith(".matter.lock", StringComparison.OrdinalIgnoreCase))
                .ToList()
            : [];
        Equal(0, preservedFiles.Count, "cancelled import left no preserved or staging file");
    }

    private static async Task MatterRootsAreIsolated()
    {
        using var scope = new TempScope();
        var sourceA = scope.WriteText(Path.Combine("source-a", "same-name.txt"), "matter A canary");
        var sourceB = scope.WriteText(Path.Combine("source-b", "same-name.txt"), "matter B canary");
        var paths = new WorkspacePaths(scope.Root);
        var repository = new MatterRepository(paths);
        var intake = new EvidenceIntakeService(paths);
        var coordinator = new EvidenceCoordinator(repository, intake);
        await repository.CreateAsync("matter-a", "Matter A");
        await repository.CreateAsync("matter-b", "Matter B");

        var recordA = await coordinator.ImportAsync("matter-a", sourceA);
        var recordB = await coordinator.ImportAsync("matter-b", sourceB);
        var storedA = await RequiredLoad(repository, "matter-a");
        var storedB = await RequiredLoad(repository, "matter-b");
        True(storedA.Evidence.Single().Segments.Single().Text.Contains("matter A canary", StringComparison.Ordinal),
            "Matter A contains only A canary");
        True(storedB.Evidence.Single().Segments.Single().Text.Contains("matter B canary", StringComparison.Ordinal),
            "Matter B contains only B canary");
        True(!paths.ResolveMatterRelativePath("matter-a", recordA.PreservedRelativePath)
            .Equals(paths.ResolveMatterRelativePath("matter-b", recordB.PreservedRelativePath), StringComparison.OrdinalIgnoreCase),
            "same-named evidence resolves to separate Matter roots");
    }

    private static async Task SaveConflictRollsPreservedEvidenceBack()
    {
        using var scope = new TempScope();
        var fixture = scope.WriteText("conflict.txt", "preserved bytes must roll back after a stale save conflict");
        var context = await CreateMatterContext(scope.Root, "matter-a");
        var conflictInjected = false;
        var progress = new SynchronousProgress<EvidenceImportProgress>(value =>
        {
            if (conflictInjected || value.Stage != "Reading preserved copy")
            {
                return;
            }
            conflictInjected = true;
            var competing = RequiredLoad(context.Repository, "matter-a").GetAwaiter().GetResult();
            competing.MatterName = "Competing revision";
            context.Repository.SaveAsync(competing, expectedRevision: competing.Revision).GetAwaiter().GetResult();
        });

        await ThrowsAsync<WorkspaceConflictException>(() =>
            context.Coordinator.ImportAsync("matter-a", fixture, progress));
        var stored = await RequiredLoad(context.Repository, "matter-a");
        Equal("Competing revision", stored.MatterName, "competing save remained authoritative");
        Equal(0, stored.Evidence.Count, "failed import was not committed");
        var originalsRoot = Path.Combine(context.Paths.EvidenceRoot("matter-a"), "originals");
        var originals = Directory.Exists(originalsRoot)
            ? Directory.EnumerateFiles(originalsRoot, "*", SearchOption.AllDirectories).ToList()
            : [];
        Equal(0, originals.Count, "preserved copy rolled back after state conflict");
    }

    private static async Task<TestContext> CreateMatterContext(string root, string matterId)
    {
        var paths = new WorkspacePaths(Path.Combine(root, "workspace"));
        var repository = new MatterRepository(paths);
        var intake = new EvidenceIntakeService(paths);
        var coordinator = new EvidenceCoordinator(repository, intake);
        await repository.CreateAsync(matterId, "Synthetic Matter");
        return new TestContext(paths, repository, coordinator);
    }

    private static async Task<MatterWorkspace> RequiredLoad(MatterRepository repository, string matterId) =>
        await repository.LoadAsync(matterId) ?? throw new InvalidOperationException("Expected Matter was not found.");

    private static async Task<string> Sha256(string path)
    {
        await using var stream = File.OpenRead(path);
        return Convert.ToHexString(await SHA256.HashDataAsync(stream)).ToLowerInvariant();
    }

    private static void CreateMinimalDocx(string path, params string[] paragraphs)
    {
        Directory.CreateDirectory(Path.GetDirectoryName(path)!);
        using var archive = ZipFile.Open(path, ZipArchiveMode.Create);
        var entry = archive.CreateEntry("word/document.xml", CompressionLevel.Optimal);
        using var writer = new StreamWriter(entry.Open(), new UTF8Encoding(encoderShouldEmitUTF8Identifier: false));
        writer.Write("<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>");
        writer.Write("<w:document xmlns:w=\"http://schemas.openxmlformats.org/wordprocessingml/2006/main\"><w:body>");
        foreach (var paragraph in paragraphs)
        {
            writer.Write("<w:p><w:r><w:t>");
            writer.Write(System.Security.SecurityElement.Escape(paragraph));
            writer.Write("</w:t></w:r></w:p>");
        }
        writer.Write("</w:body></w:document>");
    }

    private static async Task WriteLargeFixture(string path, int size)
    {
        var buffer = new byte[128 * 1024];
        Random.Shared.NextBytes(buffer);
        await using var stream = new FileStream(path, FileMode.CreateNew, FileAccess.Write, FileShare.None, buffer.Length, true);
        var remaining = size;
        while (remaining > 0)
        {
            var count = Math.Min(remaining, buffer.Length);
            await stream.WriteAsync(buffer.AsMemory(0, count));
            remaining -= count;
        }
    }

    private static async Task ThrowsAsync<T>(Func<Task> action) where T : Exception
    {
        try
        {
            await action();
        }
        catch (T)
        {
            return;
        }
        throw new InvalidOperationException($"Expected {typeof(T).Name} was not thrown.");
    }

    private static void Equal<T>(T expected, T actual, string label)
    {
        if (expected is byte[] expectedBytes && actual is byte[] actualBytes)
        {
            if (!expectedBytes.AsSpan().SequenceEqual(actualBytes))
            {
                throw new InvalidOperationException($"Assertion failed for {label}: byte sequences differ.");
            }
            return;
        }
        if (!EqualityComparer<T>.Default.Equals(expected, actual))
        {
            throw new InvalidOperationException($"Assertion failed for {label}: expected '{expected}', got '{actual}'.");
        }
    }

    private static void True(bool condition, string label)
    {
        if (!condition)
        {
            throw new InvalidOperationException($"Assertion failed: {label}.");
        }
    }

    private sealed record TestContext(
        WorkspacePaths Paths,
        MatterRepository Repository,
        EvidenceCoordinator Coordinator);

    private sealed class SynchronousProgress<T>(Action<T> callback) : IProgress<T>, IDisposable
    {
        public void Report(T value) => callback(value);
        public void Dispose()
        {
        }
    }

    private sealed class TempScope : IDisposable
    {
        public TempScope()
        {
            Root = Path.Combine(Path.GetTempPath(), "eco-recovery-tests", Guid.NewGuid().ToString("N"));
            Directory.CreateDirectory(Root);
        }

        public string Root { get; }

        public string WriteText(string relativePath, string content)
        {
            var path = Path.Combine(Root, relativePath);
            Directory.CreateDirectory(Path.GetDirectoryName(path)!);
            File.WriteAllText(path, content, new UTF8Encoding(encoderShouldEmitUTF8Identifier: false));
            return path;
        }

        public void Dispose()
        {
            try
            {
                if (Directory.Exists(Root))
                {
                    Directory.Delete(Root, recursive: true);
                }
            }
            catch
            {
                // The runner's temporary directory will be removed after the job.
            }
        }
    }
}
