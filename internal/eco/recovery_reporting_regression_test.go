package eco

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// These tests use only pre-existing APIs so they can run on the unmodified
// application baseline. All files and messages are synthetic.
func TestRecoveryReportingPreservesRollbackCause(t *testing.T) {
	primary := errors.New("synthetic activation failure")
	rollback := errors.New("synthetic rollback obstruction")
	err := restoreActivationFailure(primary, rollback)
	if !errors.Is(err, primary) || !errors.Is(err, rollback) {
		t.Fatal("activation report discarded one of the two error identities")
	}
}

func TestRecoveryReportingReturnsCleanupFailure(t *testing.T) {
	f := newRollbackFixture(t)
	backup := filepath.Join(t.TempDir(), "synthetic.ecobackup")
	if _, err := f.stage.CreatePortableBackup(backup, forcedRestorePassphrase, nil); err != nil { t.Fatal(err) }
	if err := f.stage.Close(); err != nil { t.Fatal(err) }
	before := workspaceFormatTree(t, f.root)
	var stageRoot string
	f.active.restoreBoundary = func(phase string) {
		if phase != "stage_verified" { return }
		entries, err := os.ReadDir(filepath.Dir(f.root))
		if err != nil { t.Fatal(err) }
		for _, entry := range entries {
			path := filepath.Join(filepath.Dir(f.root), entry.Name())
			if path != f.stageRoot && strings.HasPrefix(entry.Name(), filepath.Base(f.root)+".restore-") {
				if stageRoot != "" { t.Fatal("ambiguous synthetic restore stage") }
				stageRoot = path
			}
		}
		if stageRoot == "" { t.Fatal("restore did not create a stage") }
		// Abort before activation and make cleanup encounter an unowned route.
		// The original active workspace is not modified by the test hook.
		if err := f.active.Close(); err != nil { t.Fatal(err) }
		if err := os.Rename(stageRoot, stageRoot+".retained"); err != nil { t.Fatal(err) }
		if err := os.Mkdir(stageRoot, 0700); err != nil { t.Fatal(err) }
		if err := os.WriteFile(filepath.Join(stageRoot, "unrelated.txt"), []byte("retain synthetic replacement"), 0600); err != nil { t.Fatal(err) }
	}
	_, err := f.active.RestorePortableBackup(backup, forcedRestorePassphrase, nil)
	if !errors.Is(err, ErrVaultClosed) { t.Fatalf("report lost initiating failure: %v", err) }
	for _, text := range []string{"cleanup", stageRoot, "Open existing"} {
		if !strings.Contains(err.Error(), text) { t.Errorf("returned failure omits %q", text) }
	}
	if !reflect.DeepEqual(before, workspaceFormatTree(t, f.root)) { t.Error("reporting path changed the original workspace") }
	if _, err := os.Stat(filepath.Join(stageRoot, "unrelated.txt")); err != nil { t.Fatal("cleanup removed unrelated replacement") }
	if err := ValidateExistingWorkspaceRoot(stageRoot+".retained"); err != nil { t.Fatal("displaced synthetic stage was not retained") }
}
