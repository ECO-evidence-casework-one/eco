namespace Eco.Recovery.Models;

public sealed class RecoveryState
{
    public string Schema { get; set; } = "eco-recovery-state-v1";
    public string MatterId { get; set; } = "community-hall-boiler-repair";
    public string MatterName { get; set; } = "Community Hall Boiler Repair";
    public string ViewMode { get; set; } = "Guided";
    public string ActivePage { get; set; } = "CurrentPosition";
    public string ActiveConversation { get; set; } = "General casework";
    public string ConversationDraft { get; set; } = string.Empty;
    public double InterfaceScale { get; set; } = 1.0;
    public bool LowSensory { get; set; }
    public DateTimeOffset UpdatedAt { get; set; } = DateTimeOffset.UtcNow;
}
