package eco

import (
	"errors"
	"fmt"
	"os"
)

// checkedRollbackRestoreActivation refuses missing/substituted checkpoints and
// occupied destinations before moving anything. A successful return means the
// original owned directory is back at activeRoot, not merely that no rename
// happened to fail. This is handled-error recovery, not a crash journal.
func checkedRollbackRestoreActivation(activeRoot, stageRoot, preRestore string, activeOwner, stageOwner *workspaceOwnerLease, stageMoved bool) error {
	if err := validateRollbackOwnedSource(preRestore, activeOwner); err != nil {
		return fmt.Errorf("validate original recovery checkpoint: %w", err)
	}
	var stageErr error
	if stageMoved {
		if err := validateRollbackOwnedSource(activeRoot, stageOwner); err != nil {
			return fmt.Errorf("validate activated staged workspace before rollback: %w", err)
		}
		if err := requireVacantRollbackDestination(stageRoot); err != nil {
			return err
		}
		if err := os.Rename(activeRoot, stageRoot); err != nil {
			return fmt.Errorf("move failed staged vault back: %w", err)
		}
		if err := stageOwner.retarget(stageRoot); err != nil {
			stageErr = fmt.Errorf("retarget staged owner during rollback: %w", err)
		}
	}
	if err := requireVacantRollbackDestination(activeRoot); err != nil {
		return errors.Join(stageErr, err)
	}
	// Revalidate after moving the staged directory; never rename an unrelated
	// replacement merely because it appeared at the checkpoint pathname.
	if err := validateRollbackOwnedSource(preRestore, activeOwner); err != nil {
		return errors.Join(stageErr, fmt.Errorf("revalidate original recovery checkpoint: %w", err))
	}
	if err := os.Rename(preRestore, activeRoot); err != nil {
		return errors.Join(stageErr, fmt.Errorf("restore original active vault: %w", err))
	}
	if err := activeOwner.retarget(activeRoot); err != nil {
		return errors.Join(stageErr, fmt.Errorf("retarget active owner during rollback: %w", err))
	}
	return stageErr
}

// validateRollbackOwnedSource checks the intended filesystem object without
// changing the lease route. Immediately after a rename the recorded route may
// still be old, so owner.revalidate() alone cannot validate this boundary.
func validateRollbackOwnedSource(path string, owner *workspaceOwnerLease) error {
	if owner == nil {
		return errors.New("workspace recovery ownership is missing")
	}
	owner.routeMu.Lock()
	defer owner.routeMu.Unlock()
	if owner.lock == nil {
		return errors.New("workspace recovery ownership is closed")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect recovery source: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("recovery source is not a real directory")
	}
	identity, err := platformWorkspaceObjectIdentity(path)
	if err != nil {
		return fmt.Errorf("identify recovery source: %w", err)
	}
	if !owner.identity.equal(identity) {
		return errors.New("recovery source is not the owned workspace object")
	}
	return nil
}

func requireVacantRollbackDestination(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return errors.New("recovery destination is occupied; preserving existing entries")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect recovery destination: %w", err)
	}
	return nil
}
