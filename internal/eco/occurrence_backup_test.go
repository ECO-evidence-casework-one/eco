package eco

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEvidenceOccurrencesSurviveEncryptedBackupRestore(t *testing.T) {
	root := t.TempDir()
	firstPath := filepath.Join(root, "first.txt")
	secondPath := filepath.Join(root, "second.txt")
	content := []byte("occurrence history must survive portable backup")
	if err := os.WriteFile(firstPath, content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondPath, content, 0600); err != nil {
		t.Fatal(err)
	}

	sourceVault, err := openTestVault(filepath.Join(root, "source-vault"))
	if err != nil {
		t.Fatal(err)
	}
	item, _, err := sourceVault.ImportFile(firstPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, duplicate, err := sourceVault.ImportFile(secondPath, nil); err != nil || !duplicate {
		t.Fatalf("duplicate import err=%v duplicate=%v", err, duplicate)
	}
	backupPath := filepath.Join(root, "occurrences.ecobackup")
	if _, err := sourceVault.CreatePortableBackup(backupPath, "correct horse battery staple", nil); err != nil {
		t.Fatal(err)
	}

	activeVault, err := openTestVault(filepath.Join(root, "active-vault"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := activeVault.RestorePortableBackup(backupPath, "correct horse battery staple", nil); err != nil {
		t.Fatal(err)
	}
	occurrences, err := activeVault.EvidenceOccurrences(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(occurrences) != 2 {
		t.Fatalf("restored occurrence history is incomplete: %+v", occurrences)
	}
	if occurrences[0].SourcePath != "" {
		t.Fatalf("restored initial live source locator should remain cleared, got %q", occurrences[0].SourcePath)
	}
	if filepath.Clean(occurrences[1].SourcePath) != filepath.Clean(secondPath) || occurrences[1].AuditChangeID == "" {
		t.Fatalf("historical duplicate provenance did not survive encrypted backup: %+v", occurrences[1])
	}
}
