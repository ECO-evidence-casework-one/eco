package eco

import (
	"errors"
	"os"
)

// removeUnactivatedRestoreStage must run before Close releases the stage owner.
// On failed rollback the old pathname may be absent or occupied by an unrelated
// directory. Only remove the exact staged object still owned by this restore;
// otherwise preserve it/the replacement for explicit recovery.
func removeUnactivatedRestoreStage(stage *Vault, stageRoot string) error {
	if stage == nil {
		return errors.New("staged workspace is missing")
	}
	if err := validateRollbackOwnedSource(stageRoot, stage.owner); err != nil {
		return err
	}
	return os.RemoveAll(stageRoot)
}
