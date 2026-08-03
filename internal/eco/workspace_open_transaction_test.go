package eco

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWorkspaceOpenTransactionRejectsRootReplacementBeforeFirstWrite(t *testing.T) {
	runtimeID := runtimeFor("open-root-binding", "ECO-OPEN-ROOT-BINDING", Schema)
	parent := t.TempDir()
	root := filepath.Join(parent, "workspace")
	replacementPath := filepath.Join(parent, "replacement")
	authentic, err := createVault(root, "Authentic workspace", runtimeID)
	if err != nil {
		t.Fatal(err)
	}
	authentic.Close()
	replacement, err := createVault(replacementPath, "Unrelated replacement workspace", runtimeID)
	if err != nil {
		t.Fatal(err)
	}
	replacement.Close()
	authenticBefore, _ := os.ReadFile(filepath.Join(root, "workspace.ecodb"))
	replacementBefore, _ := os.ReadFile(filepath.Join(replacementPath, "workspace.ecodb"))

	inspected, err := inspectWorkspace(root, runtimeID)
	if err != nil {
		t.Fatal(err)
	}
	retained := root + ".retained"
	replaced := false
	injected := errors.New("the retained Windows handles blocked root replacement")
	_, err = openInspectedVaultWithHook(inspected, runtimeID, func(WorkspaceOpenPhase) error {
		if renameErr := os.Rename(root, retained); renameErr != nil {
			if runtime.GOOS != "windows" {
				return renameErr
			}
			return injected
		}
		if renameErr := os.Rename(replacementPath, root); renameErr != nil {
			return renameErr
		}
		replaced = true
		return nil
	})
	if err == nil {
		t.Fatal("workspace opening accepted a root replacement after authentication")
	}
	if runtime.GOOS == "windows" && !errors.Is(err, injected) {
		t.Fatalf("Windows did not retain the authenticated root through the seam: %v", err)
	}
	if replaced {
		assertBytesAtPath(t, filepath.Join(retained, "workspace.ecodb"), authenticBefore)
		assertBytesAtPath(t, filepath.Join(root, "workspace.ecodb"), replacementBefore)
	} else {
		assertBytesAtPath(t, filepath.Join(root, "workspace.ecodb"), authenticBefore)
		assertBytesAtPath(t, filepath.Join(replacementPath, "workspace.ecodb"), replacementBefore)
	}
}

func TestWorkspaceOpenTransactionRejectsControlAndObjectsReplacement(t *testing.T) {
	for _, name := range []string{"vault.key", "workspace.ecodb", workspaceIdentityFile, "objects"} {
		t.Run(name, func(t *testing.T) {
			runtimeID := runtimeFor("open-child-binding-"+name, "ECO-OPEN-CHILD-BINDING", Schema)
			parent := t.TempDir()
			root := filepath.Join(parent, "workspace")
			otherRoot := filepath.Join(parent, "other")
			workspace, err := createVault(root, "Authentic workspace", runtimeID)
			if err != nil {
				t.Fatal(err)
			}
			workspace.Close()
			other, err := createVault(otherRoot, "Unrelated workspace", runtimeID)
			if err != nil {
				t.Fatal(err)
			}
			other.Close()
			authenticTarget := filepath.Join(root, name)
			otherTarget := filepath.Join(otherRoot, name)
			var authenticTargetBefore []byte
			if name != "objects" {
				authenticTargetBefore, _ = os.ReadFile(authenticTarget)
			}
			authenticDatabase, _ := os.ReadFile(filepath.Join(root, "workspace.ecodb"))
			otherDatabase, _ := os.ReadFile(filepath.Join(otherRoot, "workspace.ecodb"))
			sentinel := filepath.Join(otherRoot, "objects", "unrelated.keep")
			if err = os.WriteFile(sentinel, []byte("unrelated object tree"), 0600); err != nil {
				t.Fatal(err)
			}

			inspected, err := inspectWorkspace(root, runtimeID)
			if err != nil {
				t.Fatal(err)
			}
			retained := authenticTarget + ".retained"
			replaced := false
			injected := errors.New("the retained Windows handles blocked child replacement")
			_, err = openInspectedVaultWithHook(inspected, runtimeID, func(WorkspaceOpenPhase) error {
				if renameErr := os.Rename(authenticTarget, retained); renameErr != nil {
					if runtime.GOOS != "windows" {
						return renameErr
					}
					return injected
				}
				if renameErr := os.Rename(otherTarget, authenticTarget); renameErr != nil {
					return renameErr
				}
				replaced = true
				return nil
			})
			if err == nil {
				t.Fatalf("workspace opening accepted replacement of %s", name)
			}
			if runtime.GOOS == "windows" && !errors.Is(err, injected) {
				t.Fatalf("Windows did not retain %s through the seam: %v", name, err)
			}
			if replaced {
				expectedDatabase := authenticDatabase
				if name == "workspace.ecodb" {
					expectedDatabase = otherDatabase
				}
				assertBytesAtPath(t, filepath.Join(root, "workspace.ecodb"), expectedDatabase)
				if name == "objects" {
					assertBytesAtPath(t, filepath.Join(root, "objects", "unrelated.keep"), []byte("unrelated object tree"))
				} else {
					assertBytesAtPath(t, retained, authenticTargetBefore)
					assertBytesAtPath(t, sentinel, []byte("unrelated object tree"))
				}
			} else {
				assertBytesAtPath(t, filepath.Join(root, "workspace.ecodb"), authenticDatabase)
				assertBytesAtPath(t, filepath.Join(otherRoot, "workspace.ecodb"), otherDatabase)
				assertBytesAtPath(t, sentinel, []byte("unrelated object tree"))
			}
		})
	}
}

func assertBytesAtPath(t *testing.T, path string, expected []byte) {
	t.Helper()
	actual, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(actual, expected) {
		t.Fatalf("content changed at %s: bytes=%d err=%v", path, len(actual), err)
	}
}

func TestWorkspaceLifecycleLockProcessHelper(t *testing.T) {
	root := os.Getenv("ECO_TEST_LOCKED_WORKSPACE")
	if root == "" {
		return
	}
	runtimeID := runtimeFor("cross-process-lock", "ECO-CROSS-PROCESS-LOCK", Schema)
	if _, err := openVault(root, runtimeID); err == nil {
		t.Fatal("a second process entered the locked workspace lifecycle transaction")
	}
}

func TestWorkspaceLifecycleLockExcludesAnotherProcess(t *testing.T) {
	runtimeID := runtimeFor("cross-process-lock", "ECO-CROSS-PROCESS-LOCK", Schema)
	root := filepath.Join(t.TempDir(), "workspace")
	vault, err := createVault(root, "Cross-process lock workspace", runtimeID)
	if err != nil {
		t.Fatal(err)
	}
	vault.Close()
	lease, err := acquireWorkspaceLifecycleLease(root)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestWorkspaceLifecycleLockProcessHelper$")
	cmd.Env = append(os.Environ(), "ECO_TEST_LOCKED_WORKSPACE="+root)
	output, runErr := cmd.CombinedOutput()
	if closeErr := lease.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	if runErr != nil {
		t.Fatalf("cross-process lifecycle-lock helper failed: %v\n%s", runErr, output)
	}
	reopened, err := openVault(root, runtimeID)
	if err != nil {
		t.Fatalf("workspace remained locked after the lifecycle transaction ended: %v", err)
	}
	reopened.Close()
}
