//go:build windows

package eco

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func createTestJunction(t *testing.T, link, target string) {
	t.Helper()
	command := exec.Command("cmd.exe", "/c", "mklink", "/J", link, target)
	if output, err := command.CombinedOutput(); err != nil {
		t.Skipf("Windows junction creation is unavailable: %v (%s)", err, output)
	}
}

func TestWorkspaceOpenAndResetRejectObjectsJunction(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	runtime := runtimeFor("candidate-object-junction", "ECO-OBJECT-JUNCTION", Schema)
	vault, err := createVault(root, "Object junction workspace", runtime)
	if err != nil {
		t.Fatal(err)
	}
	item := importSynthetic(t, vault, t.TempDir(), "junction.txt", "Synthetic Windows junction protection data.")
	external := t.TempDir()
	externalObject := filepath.Join(external, item.ObjectFile)
	externalBytes := []byte("unrelated junction target")
	if err = os.WriteFile(externalObject, externalBytes, 0600); err != nil {
		t.Fatal(err)
	}
	realObjects := filepath.Join(root, "objects-real")
	if err = os.Rename(filepath.Join(root, "objects"), realObjects); err != nil {
		t.Fatal(err)
	}
	createTestJunction(t, filepath.Join(root, "objects"), external)

	if receipt, resetErr := resetVault(vault); resetErr == nil || receipt.ObjectsRemoved != 0 {
		t.Fatalf("reset did not reject an objects junction before deletion: receipt=%+v err=%v", receipt, resetErr)
	}
	vault.Close()
	if _, err = openVault(root, runtime); err == nil {
		t.Fatal("workspace opened through an objects junction")
	}
	got, err := os.ReadFile(externalObject)
	if err != nil || !bytes.Equal(got, externalBytes) {
		t.Fatalf("junction checks changed an unrelated external file: bytes=%q err=%v", got, err)
	}
	if _, err = os.Stat(filepath.Join(realObjects, item.ObjectFile)); err != nil {
		t.Fatalf("junction checks changed the real workspace object: %v", err)
	}
}

func TestResetRevalidatesObjectsDirectoryAfterJunctionSubstitution(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	runtime := runtimeFor("candidate-junction-race", "ECO-JUNCTION-RACE", Schema)
	vault, err := createVault(root, "Junction race workspace", runtime)
	if err != nil {
		t.Fatal(err)
	}
	item := importSynthetic(t, vault, t.TempDir(), "junction-race.txt", "Synthetic junction race data.")
	external := t.TempDir()
	externalObject := filepath.Join(external, item.ObjectFile)
	externalBytes := []byte("junction race target must survive")
	if err = os.WriteFile(externalObject, externalBytes, 0600); err != nil {
		t.Fatal(err)
	}
	realObjects := filepath.Join(root, "objects-before-junction-race")
	receipt, resetErr := resetVaultWithHook(vault, func(phase ResetPhase) error {
		if renameErr := os.Rename(filepath.Join(root, "objects"), realObjects); renameErr != nil {
			return renameErr
		}
		createTestJunction(t, filepath.Join(root, "objects"), external)
		return nil
	})
	if resetErr == nil || receipt.ObjectsRemoved != 0 {
		t.Fatalf("reset cleanup followed a junction substituted after inspection: receipt=%+v err=%v", receipt, resetErr)
	}
	got, err := os.ReadFile(externalObject)
	if err != nil || !bytes.Equal(got, externalBytes) {
		t.Fatalf("junction race changed an unrelated file: bytes=%q err=%v", got, err)
	}
	if _, err = os.Stat(filepath.Join(realObjects, item.ObjectFile)); err != nil {
		t.Fatalf("junction race lost the original managed object: %v", err)
	}
}
