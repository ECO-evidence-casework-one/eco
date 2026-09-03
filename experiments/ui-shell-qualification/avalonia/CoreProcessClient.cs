using System;
using System.Buffers.Binary;
using System.Diagnostics;
using System.IO;
using System.Security.Cryptography;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading;
using System.Threading.Tasks;

namespace ECO.UIQualification;

public sealed class CoreProcessClient : IDisposable
{
    public const int ProtocolVersion = 1;
    public const int MaxFrameBytes = 8 * 1024 * 1024;

    private readonly Process _process;
    private readonly SemaphoreSlim _requestGate = new(1, 1);
    private bool _disposed;

    public int ProcessId => _process.Id;

    private CoreProcessClient(Process process)
    {
        _process = process;
    }

    public static CoreProcessClient Start(string executablePath, string expectedSha256, string? pidFile)
    {
        if (string.IsNullOrWhiteSpace(executablePath))
            throw new InvalidOperationException("Core executable path is missing.");

        var fullPath = Path.GetFullPath(executablePath);
        if (!File.Exists(fullPath))
            throw new FileNotFoundException("Core executable was not found.", fullPath);

        var expected = (expectedSha256 ?? string.Empty).Trim().ToLowerInvariant();
        if (expected.Length != 64)
            throw new InvalidOperationException("Expected core SHA-256 is missing or invalid.");

        using (var stream = File.OpenRead(fullPath))
        {
            var actual = Convert.ToHexString(SHA256.HashData(stream)).ToLowerInvariant();
            if (!CryptographicOperations.FixedTimeEquals(Convert.FromHexString(actual), Convert.FromHexString(expected)))
                throw new InvalidOperationException("Core executable SHA-256 mismatch.");
        }

        var start = new ProcessStartInfo
        {
            FileName = fullPath,
            UseShellExecute = false,
            RedirectStandardInput = true,
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            CreateNoWindow = true,
            WorkingDirectory = Path.GetDirectoryName(fullPath) ?? Environment.CurrentDirectory
        };
        var process = new Process { StartInfo = start, EnableRaisingEvents = true };
        if (!process.Start())
            throw new InvalidOperationException("Core process did not start.");

        if (!string.IsNullOrWhiteSpace(pidFile))
            File.WriteAllText(Path.GetFullPath(pidFile), process.Id.ToString(System.Globalization.CultureInfo.InvariantCulture));

        _ = Task.Run(async () =>
        {
            try { await process.StandardError.ReadToEndAsync().ConfigureAwait(false); }
            catch { }
        });

        return new CoreProcessClient(process);
    }

    public async Task<CoreResponse> PingAsync(CancellationToken cancellationToken = default)
    {
        return await SendAsync(new CoreRequest
        {
            ProtocolVersion = ProtocolVersion,
            RequestId = NewRequestId(),
            Kind = "ping"
        }, cancellationToken).ConfigureAwait(false);
    }

    public async Task<CoreResponse> ProjectMatterAsync(string matterId, CancellationToken cancellationToken = default)
    {
        return await SendAsync(new CoreRequest
        {
            ProtocolVersion = ProtocolVersion,
            RequestId = NewRequestId(),
            Kind = "project_matter",
            MatterId = matterId
        }, cancellationToken).ConfigureAwait(false);
    }

    public async Task<CoreResponse> SendAsync(CoreRequest request, CancellationToken cancellationToken = default)
    {
        if (_disposed)
            throw new ObjectDisposedException(nameof(CoreProcessClient));
        await _requestGate.WaitAsync(cancellationToken).ConfigureAwait(false);
        try
        {
            if (_process.HasExited)
                throw new InvalidOperationException("Core process exited before the request completed.");

            var payload = JsonSerializer.SerializeToUtf8Bytes(request);
            if (payload.Length <= 0 || payload.Length > MaxFrameBytes)
                throw new InvalidOperationException($"Core request frame size {payload.Length} outside allowed range.");

            var header = new byte[4];
            BinaryPrimitives.WriteUInt32BigEndian(header, (uint)payload.Length);
            await _process.StandardInput.BaseStream.WriteAsync(header, cancellationToken).ConfigureAwait(false);
            await _process.StandardInput.BaseStream.WriteAsync(payload, cancellationToken).ConfigureAwait(false);
            await _process.StandardInput.BaseStream.FlushAsync(cancellationToken).ConfigureAwait(false);

            await ReadExactlyAsync(_process.StandardOutput.BaseStream, header, cancellationToken).ConfigureAwait(false);
            var responseBytes = checked((int)BinaryPrimitives.ReadUInt32BigEndian(header));
            if (responseBytes <= 0 || responseBytes > MaxFrameBytes)
                throw new InvalidOperationException($"Core response frame size {responseBytes} outside allowed range.");

            var responsePayload = new byte[responseBytes];
            await ReadExactlyAsync(_process.StandardOutput.BaseStream, responsePayload, cancellationToken).ConfigureAwait(false);
            var response = JsonSerializer.Deserialize<CoreResponse>(responsePayload)
                ?? throw new InvalidOperationException("Core response could not be decoded.");

            if (response.ProtocolVersion != ProtocolVersion)
                throw new InvalidOperationException("Core response protocol version mismatch.");
            if (!string.Equals(response.RequestId, request.RequestId, StringComparison.Ordinal))
                throw new InvalidOperationException("Core response request ID mismatch.");
            return response;
        }
        finally
        {
            _requestGate.Release();
        }
    }

    public void Dispose()
    {
        if (_disposed)
            return;
        _disposed = true;
        try
        {
            try { _process.StandardInput.Close(); } catch { }
            if (!_process.WaitForExit(2000))
            {
                _process.Kill(entireProcessTree: true);
                if (!_process.WaitForExit(5000))
                    throw new InvalidOperationException("Core process termination was not confirmed by WaitForExit.");
            }
        }
        finally
        {
            _requestGate.Dispose();
            _process.Dispose();
        }
    }

    private static string NewRequestId() => "REQ-UI-" + Guid.NewGuid().ToString("N");

    private static async Task ReadExactlyAsync(Stream stream, byte[] buffer, CancellationToken cancellationToken)
    {
        var offset = 0;
        while (offset < buffer.Length)
        {
            var read = await stream.ReadAsync(buffer.AsMemory(offset), cancellationToken).ConfigureAwait(false);
            if (read == 0)
                throw new EndOfStreamException("Core stream ended before the complete frame arrived.");
            offset += read;
        }
    }
}

public sealed class CoreRequest
{
    [JsonPropertyName("protocol_version")]
    public int ProtocolVersion { get; init; }

    [JsonPropertyName("request_id")]
    public string RequestId { get; init; } = string.Empty;

    [JsonPropertyName("kind")]
    public string Kind { get; init; } = string.Empty;

    [JsonPropertyName("matter_id")]
    [JsonIgnore(Condition = JsonIgnoreCondition.WhenWritingNull)]
    public string? MatterId { get; init; }
}

public sealed class CoreResponse
{
    [JsonPropertyName("protocol_version")]
    public int ProtocolVersion { get; init; }

    [JsonPropertyName("request_id")]
    public string RequestId { get; init; } = string.Empty;

    [JsonPropertyName("status")]
    public string Status { get; init; } = string.Empty;

    [JsonPropertyName("error_code")]
    public string? ErrorCode { get; init; }

    [JsonPropertyName("user_message")]
    public string? UserMessage { get; init; }

    [JsonPropertyName("projection")]
    public CoreProjection? Projection { get; init; }
}

public sealed class CoreProjection
{
    [JsonPropertyName("matter_id")]
    public string MatterId { get; init; } = string.Empty;

    [JsonPropertyName("revision")]
    public string Revision { get; init; } = string.Empty;

    [JsonPropertyName("identity")]
    public CoreMatterIdentity Identity { get; init; } = new();

    [JsonPropertyName("evidence")]
    public CoreEvidence Evidence { get; init; } = new();
}

public sealed class CoreMatterIdentity
{
    [JsonPropertyName("title")]
    public string Title { get; init; } = string.Empty;

    [JsonPropertyName("status")]
    public string Status { get; init; } = string.Empty;
}

public sealed class CoreEvidence
{
    [JsonPropertyName("records")]
    public int Records { get; init; }

    [JsonPropertyName("readable")]
    public int Readable { get; init; }

    [JsonPropertyName("unresolved")]
    public int Unresolved { get; init; }
}
