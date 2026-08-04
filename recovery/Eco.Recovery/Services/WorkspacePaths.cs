using System.Text;

namespace Eco.Recovery.Services;

public sealed class WorkspacePaths
{
    public WorkspacePaths(string? rootOverride = null)
    {
        var environmentOverride = Environment.GetEnvironmentVariable("ECO_RECOVERY_STATE_ROOT");
        Root = Path.GetFullPath(
            rootOverride
            ?? (string.IsNullOrWhiteSpace(environmentOverride) ? null : environmentOverride)
            ?? Path.Combine(
                Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData),
                "ECO",
                "RecoveryCandidate"));
    }

    public string Root { get; }

    public string MatterRoot(string matterId) =>
        EnsureInsideRoot(Path.Combine(Root, "matters", SanitizeId(matterId)));

    public string MatterStatePath(string matterId) =>
        EnsureInsideRoot(Path.Combine(MatterRoot(matterId), "matter.json"));

    public string MatterLockPath(string matterId) =>
        EnsureInsideRoot(Path.Combine(MatterRoot(matterId), ".matter.lock"));

    public string EvidenceRoot(string matterId) =>
        EnsureInsideRoot(Path.Combine(MatterRoot(matterId), "evidence"));

    public string OriginalEvidenceRoot(string matterId, string evidenceId) =>
        EnsureInsideRoot(Path.Combine(EvidenceRoot(matterId), "originals", SanitizeId(evidenceId)));

    public string ImportStagingRoot(string matterId) =>
        EnsureInsideRoot(Path.Combine(EvidenceRoot(matterId), ".staging"));

    public string ResolveMatterRelativePath(string matterId, string relativePath)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(relativePath);
        var matterRoot = MatterRoot(matterId);
        var resolved = Path.GetFullPath(Path.Combine(matterRoot, relativePath.Replace('/', Path.DirectorySeparatorChar)));
        if (!IsInside(resolved, matterRoot))
        {
            throw new InvalidOperationException("The requested Matter path escapes the managed workspace.");
        }
        return resolved;
    }

    public string ToMatterRelativePath(string matterId, string absolutePath)
    {
        var matterRoot = MatterRoot(matterId);
        var resolved = Path.GetFullPath(absolutePath);
        if (!IsInside(resolved, matterRoot))
        {
            throw new InvalidOperationException("The path is outside the managed Matter workspace.");
        }
        return Path.GetRelativePath(matterRoot, resolved).Replace(Path.DirectorySeparatorChar, '/');
    }

    private string EnsureInsideRoot(string path)
    {
        var resolved = Path.GetFullPath(path);
        if (!IsInside(resolved, Root) && !string.Equals(resolved, Root, PathComparison))
        {
            throw new InvalidOperationException("The managed path escapes the ECO recovery root.");
        }
        return resolved;
    }

    private static bool IsInside(string candidate, string parent)
    {
        var parentWithSeparator = parent.TrimEnd(Path.DirectorySeparatorChar, Path.AltDirectorySeparatorChar)
            + Path.DirectorySeparatorChar;
        return candidate.StartsWith(parentWithSeparator, PathComparison);
    }

    private static string SanitizeId(string value)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(value);
        var builder = new StringBuilder(value.Length);
        foreach (var character in value.Trim())
        {
            if (char.IsAsciiLetterOrDigit(character) || character is '-' or '_')
            {
                builder.Append(char.ToLowerInvariant(character));
            }
            else
            {
                builder.Append('-');
            }
        }

        var result = builder.ToString().Trim('-');
        if (result.Length is < 1 or > 96)
        {
            throw new ArgumentOutOfRangeException(nameof(value), "A Matter or evidence ID must contain 1 to 96 safe characters.");
        }
        return result;
    }

    private static StringComparison PathComparison =>
        OperatingSystem.IsWindows() ? StringComparison.OrdinalIgnoreCase : StringComparison.Ordinal;
}
