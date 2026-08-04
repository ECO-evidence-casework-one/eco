using System.Text.Json;
using Eco.Recovery.Models;

namespace Eco.Recovery.Services;

public sealed class MatterRepository
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        WriteIndented = true,
        PropertyNameCaseInsensitive = true
    };

    private readonly WorkspacePaths _paths;

    public MatterRepository(WorkspacePaths paths)
    {
        _paths = paths ?? throw new ArgumentNullException(nameof(paths));
    }

    public async Task<MatterWorkspace?> LoadAsync(string matterId, CancellationToken cancellationToken = default)
    {
        var statePath = _paths.MatterStatePath(matterId);
        if (!File.Exists(statePath))
        {
            return null;
        }

        await using var stream = new FileStream(
            statePath,
            FileMode.Open,
            FileAccess.Read,
            FileShare.Read,
            64 * 1024,
            FileOptions.Asynchronous | FileOptions.SequentialScan);
        var workspace = await JsonSerializer.DeserializeAsync<MatterWorkspace>(stream, JsonOptions, cancellationToken);
        if (workspace is null || !string.Equals(workspace.Schema, "eco-recovery-matter-v1", StringComparison.Ordinal))
        {
            throw new InvalidDataException("The Matter state is missing or uses an unsupported schema.");
        }
        if (!string.Equals(workspace.MatterId, matterId, StringComparison.Ordinal))
        {
            throw new InvalidDataException("The Matter state identity does not match its managed location.");
        }
        return workspace;
    }

    public async Task<MatterWorkspace> CreateAsync(
        string matterId,
        string matterName,
        CancellationToken cancellationToken = default)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(matterName);
        var workspace = new MatterWorkspace
        {
            MatterId = matterId,
            MatterName = matterName.Trim(),
            Revision = 0
        };
        await SaveAsync(workspace, expectedRevision: 0, cancellationToken);
        return workspace;
    }

    public async Task SaveAsync(
        MatterWorkspace workspace,
        long expectedRevision,
        CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(workspace);
        if (workspace.Revision != expectedRevision)
        {
            throw new WorkspaceConflictException(
                workspace.MatterId,
                expectedRevision,
                workspace.Revision,
                "The caller's Matter revision does not match its own expected revision.");
        }

        var matterRoot = _paths.MatterRoot(workspace.MatterId);
        Directory.CreateDirectory(matterRoot);
        var statePath = _paths.MatterStatePath(workspace.MatterId);
        var lockPath = _paths.MatterLockPath(workspace.MatterId);

        await using var ownership = await AcquireLockAsync(lockPath, cancellationToken);
        var persistedRevision = await ReadPersistedRevisionAsync(statePath, cancellationToken);
        if (persistedRevision != expectedRevision)
        {
            throw new WorkspaceConflictException(
                workspace.MatterId,
                expectedRevision,
                persistedRevision,
                "A newer Matter revision is already stored. The stale state was not written.");
        }

        var nextRevision = checked(expectedRevision + 1);
        var originalRevision = workspace.Revision;
        var originalUpdatedAt = workspace.UpdatedAt;
        workspace.Revision = nextRevision;
        workspace.UpdatedAt = DateTimeOffset.UtcNow;

        var temporaryPath = statePath + "." + Guid.NewGuid().ToString("N") + ".new";
        try
        {
            await using (var stream = new FileStream(
                temporaryPath,
                FileMode.CreateNew,
                FileAccess.Write,
                FileShare.None,
                64 * 1024,
                FileOptions.Asynchronous | FileOptions.WriteThrough))
            {
                await JsonSerializer.SerializeAsync(stream, workspace, JsonOptions, cancellationToken);
                await stream.FlushAsync(cancellationToken);
                stream.Flush(flushToDisk: true);
            }

            File.Move(temporaryPath, statePath, overwrite: true);
        }
        catch
        {
            workspace.Revision = originalRevision;
            workspace.UpdatedAt = originalUpdatedAt;
            TryDelete(temporaryPath);
            throw;
        }
    }

    private static async Task<long> ReadPersistedRevisionAsync(
        string statePath,
        CancellationToken cancellationToken)
    {
        if (!File.Exists(statePath))
        {
            return 0;
        }

        await using var stream = new FileStream(
            statePath,
            FileMode.Open,
            FileAccess.Read,
            FileShare.Read,
            16 * 1024,
            FileOptions.Asynchronous | FileOptions.SequentialScan);
        using var document = await JsonDocument.ParseAsync(stream, cancellationToken: cancellationToken);
        if (!document.RootElement.TryGetProperty("Revision", out var revisionElement) ||
            !revisionElement.TryGetInt64(out var revision))
        {
            throw new InvalidDataException("The stored Matter revision is missing or invalid.");
        }
        return revision;
    }

    private static async Task<FileStream> AcquireLockAsync(
        string lockPath,
        CancellationToken cancellationToken)
    {
        Directory.CreateDirectory(Path.GetDirectoryName(lockPath)!);
        Exception? lastFailure = null;
        for (var attempt = 0; attempt < 80; attempt++)
        {
            cancellationToken.ThrowIfCancellationRequested();
            try
            {
                return new FileStream(
                    lockPath,
                    FileMode.OpenOrCreate,
                    FileAccess.ReadWrite,
                    FileShare.None,
                    1,
                    FileOptions.Asynchronous | FileOptions.WriteThrough);
            }
            catch (IOException exception)
            {
                lastFailure = exception;
                await Task.Delay(25, cancellationToken);
            }
        }

        throw new IOException("The Matter is currently owned by another writer.", lastFailure);
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
            // The original exception remains authoritative. Recovery can remove a stale .new file later.
        }
    }
}

public sealed class WorkspaceConflictException : IOException
{
    public WorkspaceConflictException(
        string matterId,
        long expectedRevision,
        long actualRevision,
        string message)
        : base(message)
    {
        MatterId = matterId;
        ExpectedRevision = expectedRevision;
        ActualRevision = actualRevision;
    }

    public string MatterId { get; }
    public long ExpectedRevision { get; }
    public long ActualRevision { get; }
}
