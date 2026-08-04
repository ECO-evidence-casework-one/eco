using System.Text.Json;
using Eco.Recovery.Models;

namespace Eco.Recovery.Services;

public sealed class LocalStore
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        WriteIndented = true,
        PropertyNameCaseInsensitive = true
    };

    private readonly string _root;
    private readonly string _statePath;

    public LocalStore()
    {
        var overrideRoot = Environment.GetEnvironmentVariable("ECO_RECOVERY_STATE_ROOT");
        _root = string.IsNullOrWhiteSpace(overrideRoot)
            ? Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData), "ECO", "RecoveryCandidate")
            : Path.GetFullPath(overrideRoot);
        _statePath = Path.Combine(_root, "workspace-state.json");
    }

    public bool HasSavedState => File.Exists(_statePath);

    public async Task<RecoveryState?> LoadAsync(CancellationToken cancellationToken = default)
    {
        if (!File.Exists(_statePath))
        {
            return null;
        }

        await using var stream = new FileStream(
            _statePath,
            FileMode.Open,
            FileAccess.Read,
            FileShare.Read,
            bufferSize: 32 * 1024,
            useAsync: true);
        var state = await JsonSerializer.DeserializeAsync<RecoveryState>(stream, JsonOptions, cancellationToken);
        return state?.Schema == "eco-recovery-state-v1" ? state : null;
    }

    public async Task SaveAsync(RecoveryState state, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(state);
        Directory.CreateDirectory(_root);
        state.UpdatedAt = DateTimeOffset.UtcNow;

        var temporaryPath = _statePath + ".new";
        await using (var stream = new FileStream(
            temporaryPath,
            FileMode.Create,
            FileAccess.Write,
            FileShare.None,
            bufferSize: 32 * 1024,
            useAsync: true))
        {
            await JsonSerializer.SerializeAsync(stream, state, JsonOptions, cancellationToken);
            await stream.FlushAsync(cancellationToken);
        }

        File.Move(temporaryPath, _statePath, overwrite: true);
    }
}
