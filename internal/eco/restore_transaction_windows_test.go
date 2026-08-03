//go:build windows

package eco

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPortableRestoreCleanupPinsStageThroughFinalWindowsSeam(t *testing.T) {
	state, runtimeID := interruptedRestoreState(t, restoreStagedHook)
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep.txt")
	content := []byte("Windows restore cleanup junction target")
	if err := os.WriteFile(sentinel, content, 0600); err != nil {
		t.Fatal(err)
	}
	retained := state.Stage + ".retained"
	injected := errors.New("stop after retained stage substitution attempt")
	attempted := false
	ops := operatingFilesystem
	ops.beforeRemove = func(path string) error {
		if !sameFilesystemPath(path, state.Stage) {
			return nil
		}
		attempted = true
		if err := os.Rename(state.Stage, retained); err == nil {
			createTestJunction(t, state.Stage, external)
			return errors.New("the retained restore-stage handles allowed a junction substitution")
		}
		return injected
	}
	key, err := reloadAuthenticatedRestoreState(state)
	if err != nil {
		t.Fatal(err)
	}
	err = removeRestoreStage(state, key, ops)
	zeroBytes(key)
	if !errors.Is(err, injected) || !attempted {
		t.Fatalf("restore cleanup did not retain the authenticated stage: attempted=%v err=%v", attempted, err)
	}
	assertFileExact(t, sentinel, content)
	if _, err = os.Stat(filepath.Join(state.Stage, "workspace.ecodb")); err != nil {
		t.Fatalf("authentic restore stage was deleted or renamed: %v", err)
	}
	assertRestoreRemainsRecoverable(t, state, runtimeID)
}

func TestPortableRestoreRenamePinsSourcesThroughFinalWindowsSeam(t *testing.T) {
	for _, phase := range []RestorePhase{restoreStagedHook, restoreOriginalMovedHook} {
		t.Run(string(phase), func(t *testing.T) {
			state, runtimeID := interruptedRestoreState(t, phase)
			source, destination, sourceRole, destinationRole := state.Root, state.Checkpoint, "root-original", "checkpoint"
			if phase == restoreOriginalMovedHook {
				source, destination, sourceRole, destinationRole = state.Stage, state.Root, "stage", "root"
			}
			external := t.TempDir()
			sentinel := filepath.Join(external, "keep.txt")
			content := []byte("Windows restore rename source junction target")
			if err := os.WriteFile(sentinel, content, 0600); err != nil {
				t.Fatal(err)
			}
			retained := source + ".retained"
			injected := errors.New("stop after retained restore source substitution attempt")
			attempted := false
			ops := operatingFilesystem
			ops.beforeRename = func(actualSource, actualDestination string) error {
				if !sameFilesystemPath(actualSource, source) || !sameFilesystemPath(actualDestination, destination) {
					return nil
				}
				attempted = true
				if err := os.Rename(source, retained); err == nil {
					createTestJunction(t, source, external)
					return errors.New("the retained restore source handle allowed a junction substitution")
				}
				return injected
			}
			err := secureRestoreRename(state, source, destination, sourceRole, destinationRole, ops)
			if !errors.Is(err, injected) || !attempted {
				t.Fatalf("restore rename did not retain its authenticated source: attempted=%v err=%v", attempted, err)
			}
			assertFileExact(t, sentinel, content)
			if _, err = os.Stat(filepath.Join(source, "workspace.ecodb")); err != nil {
				t.Fatalf("authentic restore source was renamed: %v", err)
			}
			assertRestoreRemainsRecoverable(t, state, runtimeID)
		})
	}
}

func TestPortableRestoreRenameRejectsFinalDestinationJunctionOnWindows(t *testing.T) {
	state, runtimeID := interruptedRestoreState(t, restoreOriginalMovedHook)
	external := t.TempDir()
	sentinel := filepath.Join(external, "keep.txt")
	content := []byte("Windows restore destination junction target")
	if err := os.WriteFile(sentinel, content, 0600); err != nil {
		t.Fatal(err)
	}
	attempted := false
	ops := operatingFilesystem
	ops.beforeRename = func(source, destination string) error {
		if sameFilesystemPath(source, state.Stage) && sameFilesystemPath(destination, state.Root) {
			attempted = true
			createTestJunction(t, state.Root, external)
		}
		return nil
	}
	err := secureRestoreRename(state, state.Stage, state.Root, "stage", "root", ops)
	if err == nil || !attempted {
		t.Fatalf("restore rename replaced a destination junction inserted at its final seam: attempted=%v err=%v", attempted, err)
	}
	assertFileExact(t, sentinel, content)
	if _, err = os.Stat(filepath.Join(state.Stage, "workspace.ecodb")); err != nil {
		t.Fatalf("authentic restore stage was renamed: %v", err)
	}
	if _, err = os.Stat(filepath.Join(state.Root, "workspace.ecodb")); !os.IsNotExist(err) {
		t.Fatalf("the substituted junction tree was activated or populated: %v", err)
	}
	if err = os.Remove(state.Root); err != nil {
		t.Fatal(err)
	}
	assertRestoreRemainsRecoverable(t, state, runtimeID)
}
