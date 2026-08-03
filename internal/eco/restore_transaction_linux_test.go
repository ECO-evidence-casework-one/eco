//go:build linux && amd64

package eco

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPortableRestoreCleanupBindsParentAtFinalLinuxSeam(t *testing.T) {
	state, runtimeID := interruptedRestoreState(t, restoreStagedHook)
	parent := filepath.Dir(state.Root)
	retainedParent := parent + ".retained"
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep.txt")
	content := []byte("Linux restore cleanup parent target")
	if err := os.WriteFile(sentinel, content, 0600); err != nil {
		t.Fatal(err)
	}
	ops := operatingFilesystem
	ops.beforeRemove = func(path string) error {
		if !sameFilesystemPath(path, state.Stage) {
			return nil
		}
		if err := os.Rename(parent, retainedParent); err != nil {
			return err
		}
		return os.Symlink(external, parent)
	}
	key, err := reloadAuthenticatedRestoreState(state)
	if err != nil {
		t.Fatal(err)
	}
	err = removeRestoreStage(state, key, ops)
	zeroBytes(key)
	if err == nil {
		t.Fatal("restore cleanup accepted a substituted parent at the final seam")
	}
	assertFileExact(t, sentinel, content)
	if _, err = os.Stat(filepath.Join(retainedParent, filepath.Base(state.Stage), "workspace.ecodb")); err != nil {
		t.Fatalf("authentic restore stage was removed through the substituted parent: %v", err)
	}
	if err = os.Remove(parent); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(retainedParent, parent); err != nil {
		t.Fatal(err)
	}
	assertRestoreRemainsRecoverable(t, state, runtimeID)
}

func TestPortableRestoreRenameRejectsFinalLinuxSourceAndDestinationSubstitution(t *testing.T) {
	state, runtimeID := interruptedRestoreState(t, restoreOriginalMovedHook)
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep.txt")
	content := []byte("Linux restore rename substitution target")
	if err := os.WriteFile(sentinel, content, 0600); err != nil {
		t.Fatal(err)
	}
	retained := state.Stage + ".retained"
	ops := operatingFilesystem
	ops.beforeRename = func(source, destination string) error {
		if err := os.Rename(state.Stage, retained); err != nil {
			return err
		}
		return os.Symlink(external, state.Stage)
	}
	if err := secureRestoreRename(state, state.Stage, state.Root, "stage", "root", ops); err == nil {
		t.Fatal("restore rename accepted a substituted source at the final seam")
	}
	assertFileExact(t, sentinel, content)
	if _, err := os.Stat(filepath.Join(retained, "workspace.ecodb")); err != nil {
		t.Fatalf("authentic restore stage was lost: %v", err)
	}
	if err := os.Remove(state.Stage); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(retained, state.Stage); err != nil {
		t.Fatal(err)
	}

	ops.beforeRename = func(source, destination string) error { return os.Symlink(external, state.Root) }
	if err := secureRestoreRename(state, state.Stage, state.Root, "stage", "root", ops); err == nil {
		t.Fatal("restore rename replaced a destination introduced at the final seam")
	}
	assertFileExact(t, sentinel, content)
	if err := os.Remove(state.Root); err != nil {
		t.Fatal(err)
	}
	assertRestoreRemainsRecoverable(t, state, runtimeID)
}
