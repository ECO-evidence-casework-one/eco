namespace Eco.Recovery.Models;

public sealed class MatterWorkspace
{
    public string Schema { get; set; } = "eco-recovery-matter-v1";
    public required string MatterId { get; init; }
    public required string MatterName { get; set; }
    public long Revision { get; set; }
    public DateTimeOffset CreatedAt { get; init; } = DateTimeOffset.UtcNow;
    public DateTimeOffset UpdatedAt { get; set; } = DateTimeOffset.UtcNow;
    public List<EvidenceRecord> Evidence { get; init; } = [];
    public List<ConversationThread> Conversations { get; init; } = [];
    public List<PositionFinding> PositionFindings { get; init; } = [];
    public UserWorkspacePreferences Preferences { get; set; } = new();
}

public sealed class ConversationThread
{
    public required string Id { get; init; }
    public required string Title { get; set; }
    public DateTimeOffset CreatedAt { get; init; } = DateTimeOffset.UtcNow;
    public DateTimeOffset UpdatedAt { get; set; } = DateTimeOffset.UtcNow;
    public string Draft { get; set; } = string.Empty;
    public List<ConversationMessage> Messages { get; init; } = [];
    public List<string> SelectedEvidenceIds { get; init; } = [];
}

public enum ConversationSpeaker
{
    User,
    Eco,
    System
}

public sealed class ConversationMessage
{
    public required string Id { get; init; }
    public ConversationSpeaker Speaker { get; init; }
    public required string Text { get; init; }
    public DateTimeOffset CreatedAt { get; init; } = DateTimeOffset.UtcNow;
    public bool Verified { get; init; }
    public List<SourceCitation> Citations { get; init; } = [];
}

public sealed class SourceCitation
{
    public required string EvidenceId { get; init; }
    public required string EvidenceName { get; init; }
    public required string SegmentId { get; init; }
    public required string Location { get; init; }
    public required string SourceSha256 { get; init; }
    public string? QuotedPassage { get; init; }
}

public enum PositionFindingKind
{
    DueSoon,
    MissingEvidence,
    DifferentAccounts,
    SupportedPosition,
    UserConfirmed
}

public sealed class PositionFinding
{
    public required string Id { get; init; }
    public PositionFindingKind Kind { get; init; }
    public required string Title { get; set; }
    public required string Detail { get; set; }
    public List<SourceCitation> Citations { get; init; } = [];
    public bool UserConfirmed { get; set; }
}

public sealed class UserWorkspacePreferences
{
    public string ViewMode { get; set; } = "Guided";
    public string ActivePage { get; set; } = "CurrentPosition";
    public string? ActiveConversationId { get; set; }
    public double InterfaceScale { get; set; } = 1.0;
    public bool LowSensory { get; set; }
}
