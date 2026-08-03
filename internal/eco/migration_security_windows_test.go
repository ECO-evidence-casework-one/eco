//go:build windows

package eco

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationRecoveryRejectsJunctionParticipants(t *testing.T) {
	for _, role := range []string{"stage", "checkpoint"} {
		t.Run(role, func(t *testing.T) {
			root, _, current, state := interruptedMigrationState(t, migrationStageReady)
			participant := state.Stage
			if role == "checkpoint" {
				participant = state.Checkpoint
			}
			kept := participant + ".kept"
			if err := os.Rename(participant, kept); err != nil {
				t.Fatal(err)
			}
			external := t.TempDir()
			sentinel := filepath.Join(external, "keep.txt")
			content := []byte("junction target must remain exact")
			if err := os.WriteFile(sentinel, content, 0600); err != nil {
				t.Fatal(err)
			}
			createTestJunction(t, participant, external)
			if _, _, err := RecoverWorkspace(root, current); err == nil {
				t.Fatal("migration recovery accepted a junction participant")
			}
			assertFileExact(t, sentinel, content)
			if _, err := os.Stat(filepath.Join(kept, "workspace.ecodb")); err != nil {
				t.Fatalf("authentic migration participant was lost: %v", err)
			}
		})
	}
}

func TestMigrationCleanupPinsStageThroughFinalRemovalSeam(t *testing.T) {
	_, oldRuntime, current, state := interruptedMigrationState(t, migrationStageReady)
	retainedStage := state.Stage + ".retained"
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep.txt")
	content := []byte("Windows cleanup substitution must remain exact")
	if err := os.WriteFile(sentinel, content, 0600); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("stop after the pinned cleanup substitution attempt")
	attempted := false
	ops := operatingFilesystem
	ops.beforeRemove = func(path string) error {
		if !sameFilesystemPath(path, state.Stage) {
			return nil
		}
		attempted = true
		if err := os.Rename(state.Stage, retainedStage); err == nil {
			createTestJunction(t, state.Stage, external)
			return errors.New("the retained stage handle allowed a junction substitution")
		}
		return injected
	}
	if err := removeMigrationStageWithOps(state, ops); !errors.Is(err, injected) || !attempted {
		t.Fatalf("cleanup did not retain the authenticated stage: attempted=%v err=%v", attempted, err)
	}
	assertFileExact(t, sentinel, content)
	if _, err := os.Stat(filepath.Join(state.Stage, "workspace.ecodb")); err != nil {
		t.Fatalf("the authentic stage was deleted or renamed: %v", err)
	}
	if _, err := os.Lstat(retainedStage); !os.IsNotExist(err) {
		t.Fatalf("the retained stage handle allowed pathname replacement: %v", err)
	}
	assertStageReadyMigrationRemainsRecoverable(t, state, oldRuntime, current)
}

func TestMigrationRenamePinsSourceThroughFinalRenameSeam(t *testing.T) {
	_, oldRuntime, current, state := interruptedMigrationState(t, migrationStageReady)
	retainedStage := state.Stage + ".retained"
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep.txt")
	content := []byte("Windows rename source substitution must remain exact")
	if err := os.WriteFile(sentinel, content, 0600); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("stop after the pinned source substitution attempt")
	attempted := false
	ops := operatingFilesystem
	ops.beforeRename = func(source, destination string) error {
		if !sameFilesystemPath(source, state.Stage) || !sameFilesystemPath(destination, state.Root) {
			return nil
		}
		attempted = true
		if err := os.Rename(state.Stage, retainedStage); err == nil {
			createTestJunction(t, state.Stage, external)
			return errors.New("the retained source handle allowed a junction substitution")
		}
		return injected
	}
	if err := secureMigrationRename(state, state.Stage, state.Root, true, "", ops); !errors.Is(err, injected) || !attempted {
		t.Fatalf("rename did not retain its authenticated source: attempted=%v err=%v", attempted, err)
	}
	assertFileExact(t, sentinel, content)
	if _, err := os.Lstat(state.Root); !os.IsNotExist(err) {
		t.Fatalf("a substituted tree was activated at the workspace path: %v", err)
	}
	if _, err := os.Stat(filepath.Join(state.Stage, "workspace.ecodb")); err != nil {
		t.Fatalf("the authentic migration stage was renamed: %v", err)
	}
	assertStageReadyMigrationRemainsRecoverable(t, state, oldRuntime, current)
}

func TestMigrationRenameRejectsJunctionInsertedAtFinalDestinationSeam(t *testing.T) {
	_, oldRuntime, current, state := interruptedMigrationState(t, migrationStageReady)
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep.txt")
	content := []byte("Windows rename destination junction must remain exact")
	if err := os.WriteFile(sentinel, content, 0600); err != nil {
		t.Fatal(err)
	}
	attempted := false
	ops := operatingFilesystem
	ops.beforeRename = func(source, destination string) error {
		if !sameFilesystemPath(source, state.Stage) || !sameFilesystemPath(destination, state.Root) {
			return nil
		}
		attempted = true
		createTestJunction(t, state.Root, external)
		return nil
	}
	if err := secureMigrationRename(state, state.Stage, state.Root, true, "", ops); err == nil || !attempted {
		t.Fatalf("rename replaced a junction inserted after the destination absence check: attempted=%v err=%v", attempted, err)
	}
	assertFileExact(t, sentinel, content)
	if _, err := os.Stat(filepath.Join(state.Root, "workspace.ecodb")); !os.IsNotExist(err) {
		t.Fatalf("the junction target was populated or activated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(state.Stage, "workspace.ecodb")); err != nil {
		t.Fatalf("the authentic migration stage was renamed: %v", err)
	}
	if err := os.Remove(state.Root); err != nil {
		t.Fatal(err)
	}
	assertStageReadyMigrationRemainsRecoverable(t, state, oldRuntime, current)
}

func TestAuthenticatedControlRemovalPinsInspectedFileOnWindows(t *testing.T) {
	_, oldRuntime, current, state := interruptedMigrationState(t, migrationStageReady)
	key, err := reloadAuthenticatedMigrationState(state)
	if err != nil {
		t.Fatal(err)
	}
	defer zeroBytes(key)
	_, rolePath, err := roleForMigration(state, "stage")
	if err != nil {
		t.Fatal(err)
	}
	retainedRole := rolePath + ".retained"
	injected := errors.New("stop after the pinned control-file substitution attempt")
	attempted := false
	ops := operatingFilesystem
	ops.beforeRemove = func(path string) error {
		if !sameFilesystemPath(path, rolePath) {
			return nil
		}
		attempted = true
		if renameErr := os.Rename(rolePath, retainedRole); renameErr == nil {
			return errors.New("the retained control-file handle allowed pathname substitution")
		}
		return injected
	}
	if err = removeMigrationRole(state, "stage", key, ops); !errors.Is(err, injected) || !attempted {
		t.Fatalf("control-file removal did not retain the authenticated file: attempted=%v err=%v", attempted, err)
	}
	if _, err = os.Stat(rolePath); err != nil {
		t.Fatalf("the authentic control file was removed or renamed: %v", err)
	}
	if _, err = os.Lstat(retainedRole); !os.IsNotExist(err) {
		t.Fatalf("the control file was renamed despite its retained handle: %v", err)
	}
	assertStageReadyMigrationRemainsRecoverable(t, state, oldRuntime, current)
}
