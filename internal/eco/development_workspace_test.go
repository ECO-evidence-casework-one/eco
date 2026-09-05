package eco

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateExistingWorkspaceRootIsReadOnly(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ordinary-folder")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExistingWorkspaceRoot(root); err == nil {
		t.Fatal("ordinary folder was accepted as an ECO workspace")
	}
	for _, name := range []string{"vault.key", "workspace.ecodb", "objects"} {
		if _, err := os.Lstat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatalf("validation mutated ordinary folder by creating %s: %v", name, err)
		}
	}
}

func TestValidateExistingWorkspaceRootAcceptsClosedVault(t *testing.T) {
	root := filepath.Join(t.TempDir(), "vault")
	v, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExistingWorkspaceRoot(root); err != nil {
		t.Fatal(err)
	}
}

func TestArchiveDevelopmentWorkspaceForCleanStartPreservesOldState(t *testing.T) {
	root := filepath.Join(t.TempDir(), "candidate")
	v, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.CreateMatter("Preserved prior state"); err != nil {
		t.Fatal(err)
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}

	archive, err := ArchiveDevelopmentWorkspaceForCleanStart(root)
	if err != nil {
		t.Fatal(err)
	}
	if archive == "" || archive == root {
		t.Fatalf("invalid archive route %q", archive)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("old candidate route still exists after archive: %v", err)
	}
	if err := ValidateExistingWorkspaceRoot(archive); err != nil {
		t.Fatalf("archived workspace is not reopenable: %v", err)
	}
	old, err := OpenVault(archive)
	if err != nil {
		t.Fatal(err)
	}
	oldView := old.Snapshot()
	if len(oldView.Matters) != 1 || oldView.Matters[0].Title != "Preserved prior state" {
		_ = old.Close()
		t.Fatalf("archived workspace lost prior state: %+v", oldView.Matters)
	}
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}

	refused, refusal := OpenVault(root)
	if refused != nil {
		_ = refused.Close()
	}
	if !errors.Is(refusal, ErrWorkspaceRecoveryRequired) {
		t.Fatalf("archived route did not require an explicit new-state choice: %v", refusal)
	}
	if _, err := StartCleanDevelopmentWorkspace(root); err != nil {
		t.Fatal(err)
	}
	fresh, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	if got := fresh.Snapshot(); len(got.Evidence) != 0 || len(got.Matters) != 0 || len(got.Questions) != 0 {
		t.Fatalf("fresh candidate inherited archived records: evidence=%d matters=%d questions=%d", len(got.Evidence), len(got.Matters), len(got.Questions))
	}
}

func TestArchiveDevelopmentWorkspaceRefusesLiveOwner(t *testing.T) {
	root := filepath.Join(t.TempDir(), "candidate")
	v, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	archive, err := ArchiveDevelopmentWorkspaceForCleanStart(root)
	if archive != "" {
		t.Fatalf("archive unexpectedly created while workspace was live: %q", archive)
	}
	if !errors.Is(err, ErrWorkspaceInUse) {
		t.Fatalf("live workspace archive error=%v", err)
	}
}

func TestArchiveMissingDevelopmentWorkspaceIsNoOp(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	archive, err := ArchiveDevelopmentWorkspaceForCleanStart(root)
	if err != nil || archive != "" {
		t.Fatalf("missing workspace archive=%q err=%v", archive, err)
	}
}
