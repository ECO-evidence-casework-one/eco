//go:build linux && amd64

package eco

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationCleanupBindsParentThroughFinalRemovalSeam(t *testing.T) {
	_, oldRuntime, current, state := interruptedMigrationState(t, migrationStageReady)
	originalParent := filepath.Dir(state.Root)
	retainedParent := originalParent + ".retained"
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep.txt")
	content := []byte("portable cleanup parent substitution must remain exact")
	if err := os.WriteFile(sentinel, content, 0600); err != nil {
		t.Fatal(err)
	}
	ops := operatingFilesystem
	attempted := false
	ops.beforeRemove = func(path string) error {
		if !sameFilesystemPath(path, state.Stage) {
			return nil
		}
		attempted = true
		if err := os.Rename(originalParent, retainedParent); err != nil {
			return err
		}
		return os.Symlink(external, originalParent)
	}
	if err := removeMigrationStageWithOps(state, ops); err == nil || !attempted {
		t.Fatalf("cleanup did not reject its substituted retained parent: attempted=%v err=%v", attempted, err)
	}
	assertFileExact(t, sentinel, content)
	retainedStage := filepath.Join(retainedParent, filepath.Base(state.Stage))
	if _, err := os.Stat(filepath.Join(retainedStage, "workspace.ecodb")); err != nil {
		t.Fatalf("the authentic stage was removed through a substituted parent: %v", err)
	}
	if err := os.Remove(originalParent); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(retainedParent, originalParent); err != nil {
		t.Fatal(err)
	}
	assertStageReadyMigrationRemainsRecoverable(t, state, oldRuntime, current)
}

func TestMigrationRenameBindsSourceThroughFinalRenameSeam(t *testing.T) {
	_, oldRuntime, current, state := interruptedMigrationState(t, migrationStageReady)
	retainedStage := state.Stage + ".retained"
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep.txt")
	content := []byte("portable rename source substitution must remain exact")
	if err := os.WriteFile(sentinel, content, 0600); err != nil {
		t.Fatal(err)
	}
	ops := operatingFilesystem
	attempted := false
	ops.beforeRename = func(source, destination string) error {
		if !sameFilesystemPath(source, state.Stage) || !sameFilesystemPath(destination, state.Root) {
			return nil
		}
		attempted = true
		if err := os.Rename(state.Stage, retainedStage); err != nil {
			return err
		}
		return os.Symlink(external, state.Stage)
	}
	if err := secureMigrationRename(state, state.Stage, state.Root, true, "", ops); err == nil || !attempted {
		t.Fatalf("rename did not reject its substituted source object: attempted=%v err=%v", attempted, err)
	}
	assertFileExact(t, sentinel, content)
	if _, err := os.Lstat(state.Root); !os.IsNotExist(err) {
		t.Fatalf("a substituted source was reported or installed as the active workspace: %v", err)
	}
	if _, err := os.Stat(filepath.Join(retainedStage, "workspace.ecodb")); err != nil {
		t.Fatalf("the authentic migration stage was lost: %v", err)
	}
	if err := os.Remove(state.Stage); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(retainedStage, state.Stage); err != nil {
		t.Fatal(err)
	}
	assertStageReadyMigrationRemainsRecoverable(t, state, oldRuntime, current)
}

func TestMigrationRenameUsesNoReplaceAtFinalDestinationSeam(t *testing.T) {
	_, oldRuntime, current, state := interruptedMigrationState(t, migrationStageReady)
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep.txt")
	content := []byte("portable rename destination substitution must remain exact")
	if err := os.WriteFile(sentinel, content, 0600); err != nil {
		t.Fatal(err)
	}
	ops := operatingFilesystem
	attempted := false
	ops.beforeRename = func(source, destination string) error {
		if !sameFilesystemPath(source, state.Stage) || !sameFilesystemPath(destination, state.Root) {
			return nil
		}
		attempted = true
		return os.Symlink(external, state.Root)
	}
	if err := secureMigrationRename(state, state.Stage, state.Root, true, "", ops); err == nil || !attempted {
		t.Fatalf("rename replaced a destination introduced at the final seam: attempted=%v err=%v", attempted, err)
	}
	assertFileExact(t, sentinel, content)
	if _, err := os.Stat(filepath.Join(state.Root, "workspace.ecodb")); !os.IsNotExist(err) {
		t.Fatalf("the substituted destination was activated or populated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(state.Stage, "workspace.ecodb")); err != nil {
		t.Fatalf("the authentic migration stage was renamed: %v", err)
	}
	if err := os.Remove(state.Root); err != nil {
		t.Fatal(err)
	}
	assertStageReadyMigrationRemainsRecoverable(t, state, oldRuntime, current)
}

func TestAuthenticatedControlRemovalBindsInspectedFile(t *testing.T) {
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
	external := filepath.Join(t.TempDir(), "outside-role.json")
	content := []byte("unrelated control content must remain exact")
	if err = os.WriteFile(external, content, 0600); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("stop after authenticated control substitution")
	ops := operatingFilesystem
	ops.beforeRemove = func(path string) error {
		if !sameFilesystemPath(path, rolePath) {
			return nil
		}
		if renameErr := os.Rename(rolePath, retainedRole); renameErr != nil {
			return renameErr
		}
		if linkErr := os.Symlink(external, rolePath); linkErr != nil {
			return linkErr
		}
		return injected
	}
	if err = removeMigrationRole(state, "stage", key, ops); !errors.Is(err, injected) {
		t.Fatalf("control removal did not stop at the adversarial seam: %v", err)
	}
	assertFileExact(t, external, content)
	if err = os.Remove(rolePath); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(retainedRole, rolePath); err != nil {
		t.Fatal(err)
	}
	assertStageReadyMigrationRemainsRecoverable(t, state, oldRuntime, current)
}
