package eco

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenVaultExcludesSecondWriterUntilClose(t *testing.T) {
	root := filepath.Join(t.TempDir(), "vault")
	first, err := openTestVault(root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := openTestVault(root)
	if second != nil {
		_ = second.Close()
	}
	if !errors.Is(err, ErrWorkspaceInUse) {
		_ = first.Close()
		t.Fatalf("second OpenVault error=%v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := openTestVault(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := reopened.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestVaultCloseMakesSaveFailClosed(t *testing.T) {
	root := filepath.Join(t.TempDir(), "vault")
	v, err := openTestVault(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); !errors.Is(err, ErrVaultClosed) {
		t.Fatalf("Save after Close error=%v", err)
	}
	if len(v.key) != 0 {
		t.Fatal("closed Vault retained its encryption key")
	}
}

func TestRestoreTransfersActiveWorkspaceOwnership(t *testing.T) {
	root := t.TempDir()
	sourceFile := filepath.Join(root, "source.txt")
	if err := os.WriteFile(sourceFile, []byte("restored source bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	sourceRoot := filepath.Join(root, "source-vault")
	source, err := openTestVault(sourceRoot)
	if err != nil {
		t.Fatal(err)
	}
	item, _, err := source.ImportFile(sourceFile, nil)
	if err != nil {
		_ = source.Close()
		t.Fatal(err)
	}
	backupPath := filepath.Join(root, "portable.ecobackup")
	if _, err := source.CreatePortableBackup(backupPath, "workspace-owner-test-passphrase", nil); err != nil {
		_ = source.Close()
		t.Fatal(err)
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}

	activeRoot := filepath.Join(root, "active-vault")
	active, err := openTestVault(activeRoot)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := active.RestorePortableBackup(backupPath, "workspace-owner-test-passphrase", nil)
	if err != nil {
		_ = active.Close()
		t.Fatal(err)
	}
	if receipt.PreRestoreVault == "" {
		_ = active.Close()
		t.Fatal("restore did not retain a pre-restore checkpoint path")
	}
	snapshot := active.Snapshot()
	if len(snapshot.Evidence) != 1 || snapshot.Evidence[0].ID != item.ID {
		_ = active.Close()
		t.Fatalf("active Vault did not receive restored workspace: %+v", snapshot.Evidence)
	}
	other, err := openTestVault(activeRoot)
	if other != nil {
		_ = other.Close()
	}
	if !errors.Is(err, ErrWorkspaceInUse) {
		_ = active.Close()
		t.Fatalf("restored active root lost exclusive ownership: %v", err)
	}
	if err := active.Save(); err != nil {
		_ = active.Close()
		t.Fatalf("restored active Vault could not persist through transferred owner: %v", err)
	}
	if err := active.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := openTestVault(activeRoot)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	reopenedSnapshot := reopened.Snapshot()
	if len(reopenedSnapshot.Evidence) != 1 || reopenedSnapshot.Evidence[0].ID != item.ID {
		t.Fatalf("restored workspace did not survive close/reopen: %+v", reopenedSnapshot.Evidence)
	}
}
