namespace Eco.Recovery.Models;

public enum EvidenceState
{
    Preserving,
    Preserved,
    Extracting,
    Ready,
    ReaderUnavailable,
    Unsupported,
    Damaged,
    Cancelled
}

public enum ExtractionKind
{
    None,
    TextLayer,
    OfficeXml,
    Ocr,
    Manual
}

public sealed class EvidenceRecord
{
    public required string Id { get; init; }
    public required string OriginalName { get; init; }
    public required string PreservedRelativePath { get; init; }
    public required string Sha256 { get; init; }
    public long SizeBytes { get; init; }
    public DateTimeOffset ImportedAt { get; init; } = DateTimeOffset.UtcNow;
    public EvidenceState State { get; set; } = EvidenceState.Preserved;
    public ExtractionKind ExtractionKind { get; set; }
    public string? DuplicateOfEvidenceId { get; init; }
    public string? FailureCode { get; set; }
    public string? FailureDetail { get; set; }
    public List<EvidenceSegment> Segments { get; init; } = [];

    public bool IsSearchable => State == EvidenceState.Ready && Segments.Any(segment => segment.Searchable);
}

public sealed class EvidenceSegment
{
    public required string Id { get; init; }
    public required string EvidenceId { get; init; }
    public required string SourceSha256 { get; init; }
    public required string Text { get; init; }
    public required string Location { get; init; }
    public int? Page { get; init; }
    public string? Region { get; init; }
    public bool Searchable { get; init; }
    public ExtractionKind ExtractionKind { get; init; }
}

public sealed class EvidenceImportProgress
{
    public required string Stage { get; init; }
    public long CompletedBytes { get; init; }
    public long TotalBytes { get; init; }
    public int Percent => TotalBytes <= 0 ? 0 : (int)Math.Clamp(CompletedBytes * 100L / TotalBytes, 0, 100);
}
