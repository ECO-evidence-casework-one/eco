package eco

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func prepareSyntheticRestore(t *testing.T) (*Vault, RuntimeIdentity, string, string, string) {
	t.Helper()
	runtimeID := runtimeFor("portable-restore-transaction", "ECO-PORTABLE-RESTORE-TRANSACTION", Schema)
	parent := t.TempDir()
	sourceRoot := filepath.Join(parent, "source")
	targetRoot := filepath.Join(parent, "target")
	backupPath := filepath.Join(parent, "synthetic.ecobak")
	source, err := createVault(sourceRoot, "Synthetic restore source", runtimeID)
	if err != nil {
		t.Fatal(err)
	}
	source.Ask("Synthetic restored conversation marker.", nil)
	sourceID := source.Identity.ID
	if _, err = source.CreatePortableBackup(backupPath, "synthetic restore passphrase", nil); err != nil {
		t.Fatal(err)
	}
	source.Close()
	target, err := createVault(targetRoot, "Synthetic active target", runtimeID)
	if err != nil {
		t.Fatal(err)
	}
	target.Ask("Synthetic original conversation marker.", nil)
	return target, runtimeID, backupPath, sourceID, target.Identity.ID
}

func interruptedRestoreState(t *testing.T, phase RestorePhase) (restoreState, RuntimeIdentity) {
	t.Helper()
	target, runtimeID, backupPath, _, _ := prepareSyntheticRestore(t)
	_, err := target.restorePortableBackupWithOps(backupPath, "synthetic restore passphrase", nil, func(current RestorePhase, _ *Vault) error {
		if current == phase {
			return errRestoreInterrupted
		}
		return nil
	}, operatingFilesystem)
	if !errors.Is(err, errRestoreInterrupted) {
		t.Fatalf("portable restore did not stop at %s: %v", phase, err)
	}
	state, key, err := readRestoreState(target.Root)
	if err != nil {
		t.Fatal(err)
	}
	zeroBytes(key)
	target.Close()
	return state, runtimeID
}

func assertRestoreRemainsRecoverable(t *testing.T, state restoreState, runtimeID RuntimeIdentity) {
	t.Helper()
	remaining, key, err := readRestoreState(state.Root)
	if err != nil {
		t.Fatalf("authenticated portable restore record was lost: %v", err)
	}
	zeroBytes(key)
	if !sameRestoreState(remaining, state) {
		t.Fatalf("portable restore record stopped truthfully describing the partial operation: got=%+v want=%+v", remaining, state)
	}
	session, _, err := RecoverPortableRestore(state.Root, runtimeID)
	if err != nil {
		t.Fatalf("portable restore was not recoverable: %v", err)
	}
	session.Vault.Close()
}

func TestPortableRestoreRecoversEveryActivationBoundary(t *testing.T) {
	for _, phase := range []RestorePhase{restoreStagedHook, restoreOriginalMovedHook, restoreActivatedHook, restoreRecoveredHook} {
		t.Run(string(phase), func(t *testing.T) {
			target, runtimeID, backupPath, restoredID, originalID := prepareSyntheticRestore(t)
			_, err := target.restorePortableBackupWithOps(backupPath, "synthetic restore passphrase", nil, func(current RestorePhase, _ *Vault) error {
				if current == phase {
					return errRestoreInterrupted
				}
				return nil
			}, operatingFilesystem)
			if !errors.Is(err, errRestoreInterrupted) {
				t.Fatalf("restore was not interrupted after %s: %v", phase, err)
			}
			state, key, stateErr := readRestoreState(target.Root)
			if stateErr != nil {
				t.Fatalf("authenticated restore record was unavailable after %s: %v", phase, stateErr)
			}
			zeroBytes(key)
			if state.Phase != restoreTransactionPhase(phase) {
				t.Fatalf("restore record was not truthful after %s: %+v", phase, state)
			}
			target.Close()
			session, receipt, recoverErr := RecoverPortableRestore(state.Root, runtimeID)
			if recoverErr != nil {
				t.Fatalf("restore recovery failed after %s: %v", phase, recoverErr)
			}
			defer session.Vault.Close()
			expectedID := originalID
			if phase == restoreActivatedHook || phase == restoreRecoveredHook {
				expectedID = restoredID
				if !receipt.MigrationKept {
					t.Fatalf("activated restore recovery did not keep the original checkpoint: %+v", receipt)
				}
			}
			if session.Identity.ID != expectedID {
				t.Fatalf("recovery selected the wrong workspace after %s: got=%s want=%s", phase, session.Identity.ID, expectedID)
			}
			if !hasChange(session.Vault.Snapshot(), "workspace-recovered") {
				t.Fatalf("recovery after %s was not recorded truthfully", phase)
			}
			if _, markerErr := os.Lstat(restoreStatePath(state.Root)); !os.IsNotExist(markerErr) {
				t.Fatalf("completed recovery left a misleading restore marker: %v", markerErr)
			}
		})
	}
}

func TestPortableRestoreStageCleanupUsesAuthenticatedObjectBoundTree(t *testing.T) {
	target, _, backupPath, _, originalID := prepareSyntheticRestore(t)
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep.txt")
	content := []byte("unrelated restore cleanup content")
	if err := os.WriteFile(sentinel, content, 0600); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("stop at the final restore cleanup seam")
	attempted := false
	ops := operatingFilesystem
	ops.beforeRemove = func(path string) error {
		if !strings.Contains(filepath.Base(path), ".restore-stage-") {
			return nil
		}
		attempted = true
		return injected
	}
	_, err := target.restorePortableBackupWithOps(backupPath, "wrong synthetic passphrase", nil, nil, ops)
	if err == nil || !attempted {
		t.Fatalf("failed restore did not reach authenticated object-bound staging cleanup: attempted=%v err=%v", attempted, err)
	}
	assertFileExact(t, sentinel, content)
	state, key, stateErr := readRestoreState(target.Root)
	if stateErr != nil {
		t.Fatalf("cleanup failure removed the truthful recovery record: %v", stateErr)
	}
	zeroBytes(key)
	if state.Phase != restorePrepared || target.Identity.ID != originalID {
		t.Fatalf("partial cleanup was presented as complete: state=%+v target=%s", state, target.Identity.ID)
	}
}
