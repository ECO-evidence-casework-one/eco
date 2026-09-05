package eco

import (
	"errors"
	"fmt"
	"strings"
)

// WorkspaceRecoveryError is for local recovery UI, not automatic telemetry or
// diagnostics export. Locations are leads to verify, never trusted instructions
// to rename/delete/activate files. The underlying error tree is preserved.
type WorkspaceRecoveryError struct {
	Operation      string
	Cause          error
	ActiveRoot     string
	CheckpointRoot string
	StageRoot      string
}

func (e *WorkspaceRecoveryError) Error() string {
	return fmt.Sprintf("%s did not complete: %v\n\n%s\nStop this operation. Do not delete or rename retained copies. Close ECO, then use Open existing to select and verify the original workspace or checkpoint. Listed locations are recovery leads, not proof that their contents are complete.", e.Operation, e.Cause, recoveryLocations(e.ActiveRoot, e.CheckpointRoot, e.StageRoot))
}

func (e *WorkspaceRecoveryError) Unwrap() error { return e.Cause }

func recoveryLocations(active, checkpoint, stage string) string {
	lines := []string{}
	for _, pair := range [][2]string{{"Requested workspace", active}, {"Checkpoint location", checkpoint}, {"Staging location", stage}} {
		if pair[1] != "" {
			lines = append(lines, fmt.Sprintf("%s: %q", pair[0], pair[1]))
		}
	}
	return strings.Join(lines, "\n")
}

func withRecoveryContext(operation string, cause error, active, checkpoint, stage string) error {
	if cause == nil { return nil }
	return &WorkspaceRecoveryError{Operation: operation, Cause: cause, ActiveRoot: active, CheckpointRoot: checkpoint, StageRoot: stage}
}

func finalizeRestoreStage(stage *Vault, stageRoot string, activated bool) error {
	var cleanupErr error
	if !activated {
		if err := removeUnactivatedRestoreStage(stage, stageRoot); err != nil {
			cleanupErr = fmt.Errorf("staged-workspace cleanup did not complete at %q: %w", stageRoot, err)
		}
	}
	var closeErr error
	if err := stage.Close(); err != nil {
		closeErr = fmt.Errorf("staged-workspace ownership release did not complete at %q: %w", stageRoot, err)
	}
	return errors.Join(cleanupErr, closeErr)
}

// Once activation committed, a finalization failure is a warning on a successful
// receipt. Returning a failed-restore error at that point would misstate which
// workspace is active and could encourage an unsafe retry.
func recordRestoreFinalization(receipt *RestoreReceipt, resultErr *error, activated bool, finalErr error) {
	if finalErr == nil { return }
	if activated {
		receipt.RecoveryWarnings = append(receipt.RecoveryWarnings,
			"The restored workspace is active, but final cleanup or ownership release needs attention: "+finalErr.Error()+". Close ECO before another restore and retain the previous checkpoint.")
		return
	}
	*resultErr = errors.Join(*resultErr, finalErr)
}

// RestoreCompletionNotice is shared by Windows UI and its wording tests. It
// never turns a completed activation warning into 'restore failed' or suppresses
// the previous-workspace route.
func RestoreCompletionNotice(r RestoreReceipt) (string, string) {
	title := "Encrypted backup restored"
	body := fmt.Sprintf("Restored items: %d\r\nRestored bytes: %s\r\nSource build: %s\r\nSource SHA-256: %s\r\n\r\nYour previous vault was retained at:\r\n%s", r.EvidenceItems, HumanBytes(r.RestoredBytes), r.SourceBuildID, r.SourceSHA256, r.PreRestoreVault)
	if len(r.RecoveryWarnings) != 0 {
		title = "Backup restored - attention required"
		body += "\r\n\r\n" + strings.Join(r.RecoveryWarnings, "\r\n\r\n")
	}
	return title, body
}

func rollbackArchivedWorkspace(archive, original string, owner *workspaceOwnerLease) error {
	if err := validateRollbackOwnedSource(archive, owner); err != nil { return err }
	if err := requireVacantRollbackDestination(original); err != nil { return err }
	return renameArchivedWorkspaceBack(archive, original, owner)
}
