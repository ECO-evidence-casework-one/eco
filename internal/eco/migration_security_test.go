package eco

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func interruptedMigrationState(t *testing.T, phase MigrationPhase) (string, RuntimeIdentity, RuntimeIdentity, migrationState) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "workspace")
	oldRuntime := runtimeFor("legacy-migration-security", "ECO-LEGACY-MIGRATION-SECURITY", 1)
	legacy := createPriorSchemaOneWorkspace(t, root, oldRuntime)
	importSynthetic(t, legacy, t.TempDir(), "security.txt", "Synthetic migration security data.")
	legacy.Close()
	current := runtimeFor("current-migration-security", "ECO-CURRENT-MIGRATION-SECURITY", Schema)
	_, _, err := migrateWorkspace(root, current, func(currentPhase MigrationPhase) error {
		if currentPhase == phase {
			return errMigrationInterrupted
		}
		return nil
	})
	if !errors.Is(err, errMigrationInterrupted) {
		t.Fatalf("migration did not stop at %s: %v", phase, err)
	}
	state, key, err := readMigrationState(root)
	if err != nil {
		t.Fatal(err)
	}
	zeroBytes(key)
	return root, oldRuntime, current, state
}

func writeTamperedMigrationState(t *testing.T, root string, state migrationState) {
	t.Helper()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(migrationStatePath(root), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func assertFileExact(t *testing.T, path string, expected []byte) {
	t.Helper()
	actual, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(actual, expected) {
		t.Fatalf("unrelated file changed: path=%s bytes=%q err=%v", path, actual, err)
	}
}

func assertStageReadyMigrationRemainsRecoverable(t *testing.T, state migrationState, oldRuntime, current RuntimeIdentity) {
	t.Helper()
	remaining, key, err := readMigrationState(state.Root)
	if err != nil {
		t.Fatalf("the authenticated migration record is no longer recoverable: %v", err)
	}
	zeroBytes(key)
	if remaining.Phase != migrationStageReady || remaining.Root != state.Root || remaining.Stage != state.Stage || remaining.Checkpoint != state.Checkpoint {
		t.Fatalf("the migration record no longer truthfully reports the interrupted operation: %+v", remaining)
	}
	stage, err := openVaultIgnoringRecovery(state.Stage, current)
	if err != nil {
		t.Fatalf("the authentic staged workspace is unavailable: %v", err)
	}
	stage.Close()
	checkpoint, err := openVaultIgnoringRecovery(state.Checkpoint, oldRuntime)
	if err != nil {
		t.Fatalf("the authentic checkpoint is unavailable: %v", err)
	}
	checkpoint.Close()
}

func TestTamperedMigrationPathsCannotAuthoriseUnrelatedSiblingChanges(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(root, unrelated string, state *migrationState)
	}{
		{
			name: "stage points to matching-prefix sibling",
			mutate: func(root, unrelated string, state *migrationState) {
				state.Stage = filepath.Join(filepath.Dir(root), filepath.Base(root)+".migration-stage-unrelated")
			},
		},
		{
			name: "checkpoint points to matching-prefix sibling",
			mutate: func(root, unrelated string, state *migrationState) {
				state.Checkpoint = filepath.Join(filepath.Dir(root), filepath.Base(root)+".migration-checkpoint-unrelated")
			},
		},
		{
			name:   "root altered",
			mutate: func(root, unrelated string, state *migrationState) { state.Root = unrelated },
		},
		{
			name:   "nonce altered",
			mutate: func(root, unrelated string, state *migrationState) { state.Nonce = strings.Repeat("a", 32) },
		},
		{
			name:   "workspace identity altered",
			mutate: func(root, unrelated string, state *migrationState) { state.WorkspaceID = "WS-TAMPERED" },
		},
		{
			name:   "phase altered",
			mutate: func(root, unrelated string, state *migrationState) { state.Phase = migrationActivated },
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, _, current, state := interruptedMigrationState(t, migrationStageReady)
			unrelated := filepath.Join(filepath.Dir(root), filepath.Base(root)+".migration-unrelated")
			if err := os.Mkdir(unrelated, 0700); err != nil {
				t.Fatal(err)
			}
			sentinel := filepath.Join(unrelated, "keep.txt")
			content := []byte("unrelated sibling must remain exact")
			if err := os.WriteFile(sentinel, content, 0600); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(test.name, "stage points") {
				state.Stage = filepath.Join(filepath.Dir(root), filepath.Base(root)+".migration-stage-unrelated")
				if err := os.Rename(unrelated, state.Stage); err != nil {
					t.Fatal(err)
				}
				unrelated = state.Stage
				sentinel = filepath.Join(unrelated, "keep.txt")
			} else if strings.Contains(test.name, "checkpoint points") {
				state.Checkpoint = filepath.Join(filepath.Dir(root), filepath.Base(root)+".migration-checkpoint-unrelated")
				if err := os.Rename(unrelated, state.Checkpoint); err != nil {
					t.Fatal(err)
				}
				unrelated = state.Checkpoint
				sentinel = filepath.Join(unrelated, "keep.txt")
			} else {
				test.mutate(root, unrelated, &state)
			}
			writeTamperedMigrationState(t, root, state)
			if _, _, err := RecoverWorkspace(root, current); err == nil {
				t.Fatal("tampered migration recovery record was accepted")
			}
			assertFileExact(t, sentinel, content)
		})
	}
}

func TestMigrationRecoveryRejectsSymbolicLinkParticipants(t *testing.T) {
	for _, role := range []string{"stage", "checkpoint"} {
		t.Run(role, func(t *testing.T) {
			root, _, current, state := interruptedMigrationState(t, migrationStageReady)
			var participant string
			if role == "stage" {
				participant = state.Stage
			} else {
				participant = state.Checkpoint
			}
			kept := participant + ".kept"
			if err := os.Rename(participant, kept); err != nil {
				t.Fatal(err)
			}
			external := t.TempDir()
			sentinel := filepath.Join(external, "keep.txt")
			content := []byte("symbolic-link target must remain exact")
			if err := os.WriteFile(sentinel, content, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(external, participant); err != nil {
				t.Skipf("symbolic links are unavailable on this platform: %v", err)
			}
			if _, _, err := RecoverWorkspace(root, current); err == nil {
				t.Fatal("migration recovery accepted a symbolic-link participant")
			}
			assertFileExact(t, sentinel, content)
			if _, err := os.Stat(filepath.Join(kept, "workspace.ecodb")); err != nil {
				t.Fatalf("the authentic migration participant was lost: %v", err)
			}
		})
	}
}

func TestRollbackCompensatesWhenCheckpointActivationFails(t *testing.T) {
	root, oldRuntime, current, state := interruptedMigrationState(t, migrationActivated)
	injected := errors.New("injected checkpoint activation failure")
	failedOnce := false
	ops := operatingFilesystem
	ops.beforeRename = func(source, destination string) error {
		if !failedOnce && sameFilesystemPath(source, state.Checkpoint) && sameFilesystemPath(destination, state.Root) {
			failedOnce = true
			return injected
		}
		return nil
	}
	if restored, err := rollbackMigrationStateWithOps(state, ops); err == nil || restored || !errors.Is(err, injected) {
		t.Fatalf("injected rollback failure was not reported truthfully: restored=%v err=%v", restored, err)
	}
	active, err := openVaultIgnoringRecovery(root, current)
	if err != nil {
		t.Fatalf("compensation did not restore the active workspace path: %v", err)
	}
	if active.Snapshot().Schema != current.Schema {
		t.Fatalf("compensated active workspace is invalid: %+v", active.Snapshot())
	}
	active.Close()
	checkpoint, err := openVaultIgnoringRecovery(state.Checkpoint, oldRuntime)
	if err != nil {
		t.Fatalf("checkpoint copy was lost after injected failure: %v", err)
	}
	checkpoint.Close()
	remaining, key, err := readMigrationState(root)
	if err != nil {
		t.Fatalf("authenticated recovery record was not retained: %v", err)
	}
	zeroBytes(key)
	if remaining.Phase != migrationActivated || remaining.Checkpoint != state.Checkpoint || remaining.Root != state.Root {
		t.Fatalf("remaining recovery state is not truthful: %+v", remaining)
	}

	if restored, err := rollbackMigrationState(remaining); err != nil || !restored {
		t.Fatalf("retrying recovery was not deterministic: restored=%v err=%v", restored, err)
	}
	original, err := openVaultIgnoringRecovery(root, oldRuntime)
	if err != nil {
		t.Fatalf("retry did not restore the original workspace: %v", err)
	}
	original.Close()
	failedCopy, err := openVaultIgnoringRecovery(state.Failed, current)
	if err != nil {
		t.Fatalf("retry silently lost the migrated copy: %v", err)
	}
	failedCopy.Close()
	if _, err = os.Stat(migrationStatePath(root)); !os.IsNotExist(err) {
		t.Fatalf("completed deterministic retry left a migration marker: %v", err)
	}
}

func TestPreparedMigrationRecoveryClearsFalseRecoveryStateWithoutChangingSource(t *testing.T) {
	root, oldRuntime, current, state := interruptedMigrationState(t, migrationPrepared)
	before := workspaceTreeDigest(t, root)
	if _, _, err := RecoverWorkspace(root, current); err == nil {
		t.Fatal("schema-one source unexpectedly opened as a current workspace")
	}
	if !reflect.DeepEqual(before, workspaceTreeDigest(t, root)) {
		t.Fatal("prepared-state recovery changed the untouched source workspace")
	}
	original, err := openVaultIgnoringRecovery(root, oldRuntime)
	if err != nil {
		t.Fatalf("prepared-state recovery made the source workspace unusable: %v", err)
	}
	original.Close()
	for _, control := range []string{
		migrationStatePath(root),
		migrationRolePath(state.Checkpoint),
		migrationRolePath(state.Stage),
		migrationRolePath(state.Failed),
	} {
		if _, err = os.Lstat(control); !os.IsNotExist(err) {
			t.Fatalf("prepared-state recovery retained unused control %s: %v", control, err)
		}
	}
}
