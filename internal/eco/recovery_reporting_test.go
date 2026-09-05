package eco

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestRecoveryReportingContext(t *testing.T) {
	cause := errors.New("synthetic read failure")
	err := withRecoveryContext("Restore", cause, "synthetic-active", "synthetic-checkpoint", "synthetic-stage")
	var report *WorkspaceRecoveryError
	if !errors.As(err, &report) || !errors.Is(err, cause) { t.Fatal("typed report lost the original cause") }
	for _, text := range []string{"synthetic-active", "synthetic-checkpoint", "synthetic-stage", "Do not delete", "Open existing", "not proof"} {
		if !strings.Contains(err.Error(), text) { t.Fatalf("recovery report omits %q", text) }
	}
	if withRecoveryContext("Restore", nil, "", "", "") != nil { t.Fatal("nil error became failure") }
}

func TestRecoveryReportingFinalizationStates(t *testing.T) {
	for _, name := range []string{"no_failure", "failed_operation", "committed_with_warning"} {
		t.Run(name, func(t *testing.T) {
			primary := errors.New("synthetic primary problem")
			secondary := errors.New("synthetic release problem")
			receipt := RestoreReceipt{PreRestoreVault: "synthetic-original-checkpoint", EvidenceItems: 1}
			var resultErr error
			activated := name == "committed_with_warning"
			extra := secondary
			if name == "failed_operation" { resultErr = primary }
			if name == "no_failure" { extra = nil }
			recordRestoreFinalization(&receipt, &resultErr, activated, extra)
			if name == "failed_operation" {
				if !errors.Is(resultErr, primary) || !errors.Is(resultErr, secondary) { t.Fatal("finalization replaced the initiating failure") }
				if len(receipt.RecoveryWarnings) != 0 { t.Fatal("failed operation labelled committed") }
			} else if resultErr != nil { t.Fatal("successful activation labelled failed") }
			title, body := RestoreCompletionNotice(receipt)
			if !strings.Contains(body, receipt.PreRestoreVault) { t.Fatal("completion notice lost previous-workspace route") }
			if activated {
				if len(receipt.RecoveryWarnings) != 1 || !strings.Contains(title, "attention required") || !strings.Contains(body, secondary.Error()) || !strings.Contains(body, "workspace is active") { t.Fatal("committed warning hidden or mislabelled") }
			} else if len(receipt.RecoveryWarnings) != 0 { t.Fatal("spurious warning") }
		})
	}
}

func TestRecoveryReportingArchiveFailurePaths(t *testing.T) {
	root := filepath.Join(t.TempDir(), "synthetic-active")
	v, err := CreateVault(root)
	if err != nil { t.Fatal(err) }
	if _, err := v.CreateMatter("Synthetic retained original"); err != nil { _ = v.Close(); t.Fatal(err) }
	if err := v.Close(); err != nil { t.Fatal(err) }
	var attempted string
	var before map[string]string
	archive, err := archiveDevelopmentWorkspace(root, func(path string) {
		attempted = path
		if err := os.Rename(path, path+".retained"); err != nil { t.Fatal(err) }
		if err := os.Mkdir(path, 0700); err != nil { t.Fatal(err) }
		if err := os.WriteFile(filepath.Join(path, "unrelated.txt"), []byte("retain synthetic replacement"), 0600); err != nil { t.Fatal(err) }
		before = workspaceFormatTree(t, filepath.Dir(root))
	})
	if err == nil || archive != attempted { t.Fatalf("archive failure omitted attempted recovery route: %q %v", archive, err) }
	var report *WorkspaceRecoveryError
	if !errors.As(err, &report) || report.CheckpointRoot != attempted || report.ActiveRoot != root { t.Fatal("archive error did not identify both possible routes") }
	if !strings.Contains(err.Error(), strconv.Quote(attempted)) || !strings.Contains(err.Error(), "Open existing") { t.Fatal("archive report omitted local recovery guidance") }
	if !reflect.DeepEqual(before, workspaceFormatTree(t, filepath.Dir(root))) { t.Fatal("rollback moved substituted directory or changed retained files") }
	old, err := OpenVault(attempted+".retained")
	if err != nil { t.Fatal(err) }
	defer old.Close()
	if len(old.Snapshot().Matters) != 1 { t.Fatal("original archive no longer recoverable") }
}

func TestRecoveryReportingNormalRestore(t *testing.T) {
	f := newRollbackFixture(t)
	backup := filepath.Join(t.TempDir(), "synthetic.ecobackup")
	if _, err := f.stage.CreatePortableBackup(backup, forcedRestorePassphrase, nil); err != nil { t.Fatal(err) }
	if err := f.stage.Close(); err != nil { t.Fatal(err) }
	r, err := f.active.RestorePortableBackup(backup, forcedRestorePassphrase, nil)
	if err != nil { t.Fatal(err) }
	if len(r.RecoveryWarnings) != 0 { t.Fatal("normal restore received a spurious finalization warning") }
	if err := verifyStagedVault(f.active); err != nil { t.Fatal(err) }
	title, body := RestoreCompletionNotice(r)
	if title != "Encrypted backup restored" || !strings.Contains(body, r.PreRestoreVault) { t.Fatal("normal completion notice no longer describes the retained checkpoint") }
}

func TestRecoveryReportingArchiveRollbackSuccess(t *testing.T) {
	root := filepath.Join(t.TempDir(), "synthetic-active")
	v, err := CreateVault(root)
	if err != nil { t.Fatal(err) }
	defer v.Close()
	before := workspaceFormatTree(t, root)
	archive := root+".previous-synthetic"
	if err := os.Rename(root, archive); err != nil { t.Fatal(err) }
	if err := rollbackArchivedWorkspace(archive, root, v.owner); err != nil { t.Fatal(err) }
	if !reflect.DeepEqual(before, workspaceFormatTree(t, root)) { t.Fatal("rollback changed the owned workspace") }
	if err := v.Save(); err != nil { t.Fatal("rollback left ownership unusable", err) }
}
