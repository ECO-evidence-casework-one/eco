package eco

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEvidenceOccurrencesRecordDistinctDuplicateSourcesWithoutExtraObject(t *testing.T) {
	root := t.TempDir()
	firstDir := filepath.Join(root, "first-source")
	secondDir := filepath.Join(root, "second-source")
	if err := os.MkdirAll(firstDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(secondDir, 0700); err != nil {
		t.Fatal(err)
	}
	content := []byte("same evidential bytes supplied through two locations")
	firstPath := filepath.Join(firstDir, "original-name.txt")
	secondPath := filepath.Join(secondDir, "renamed-copy.txt")
	if err := os.WriteFile(firstPath, content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondPath, content, 0600); err != nil {
		t.Fatal(err)
	}
	v, err := OpenVault(filepath.Join(root, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	item, duplicate, err := v.ImportFile(firstPath, nil)
	if err != nil || duplicate {
		t.Fatalf("first import err=%v duplicate=%v", err, duplicate)
	}
	returned, duplicate, err := v.ImportFile(secondPath, nil)
	if err != nil || !duplicate {
		t.Fatalf("duplicate import err=%v duplicate=%v", err, duplicate)
	}
	if returned.ID != item.ID || returned.ObjectFile != item.ObjectFile {
		t.Fatalf("duplicate import did not reuse preserved evidence: first=%+v returned=%+v", item, returned)
	}

	occurrences, err := v.EvidenceOccurrences(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(occurrences) != 2 {
		t.Fatalf("expected initial + duplicate occurrence, got %+v", occurrences)
	}
	if occurrences[0].Kind != occurrenceKindInitial || occurrences[0].OriginalName != "original-name.txt" {
		t.Fatalf("initial occurrence is wrong: %+v", occurrences[0])
	}
	if occurrences[1].Kind != occurrenceKindDuplicate || occurrences[1].OriginalName != "renamed-copy.txt" || filepath.Clean(occurrences[1].SourcePath) != filepath.Clean(secondPath) {
		t.Fatalf("duplicate occurrence lost source provenance: %+v", occurrences[1])
	}
	for _, occ := range occurrences {
		if occ.EvidenceID != item.ID || occ.SHA256 != item.SHA256 || occ.Size != item.Size || occ.ObjectFile != item.ObjectFile {
			t.Fatalf("occurrence does not identify exact retained bytes: %+v", occ)
		}
		if occ.ObservedAt.IsZero() || occ.SourceVerifiedAt.IsZero() || occ.PreservedObjectVerifiedAt.IsZero() {
			t.Fatalf("occurrence is missing verification timestamps: %+v", occ)
		}
	}
	if occurrences[1].AuditChangeID == "" {
		t.Fatalf("duplicate occurrence is not linked to authenticated audit change: %+v", occurrences[1])
	}

	entries, err := os.ReadDir(v.Objects)
	if err != nil {
		t.Fatal(err)
	}
	objectCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".ecoobj") {
			objectCount++
		}
	}
	if objectCount != 1 {
		t.Fatalf("exact duplicate created extra encrypted object(s): %d", objectCount)
	}
}

func TestEvidenceOccurrencesPersistAcrossVaultReopen(t *testing.T) {
	root := t.TempDir()
	pathA := filepath.Join(root, "a.txt")
	pathB := filepath.Join(root, "b.txt")
	content := []byte("persistent duplicate occurrence")
	if err := os.WriteFile(pathA, content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pathB, content, 0600); err != nil {
		t.Fatal(err)
	}
	vaultRoot := filepath.Join(root, "vault")
	v, err := OpenVault(vaultRoot)
	if err != nil {
		t.Fatal(err)
	}
	item, _, err := v.ImportFile(pathA, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, duplicate, err := v.ImportFile(pathB, nil); err != nil || !duplicate {
		t.Fatalf("duplicate import err=%v duplicate=%v", err, duplicate)
	}

	reopened, err := OpenVault(vaultRoot)
	if err != nil {
		t.Fatal(err)
	}
	occurrences, err := reopened.EvidenceOccurrences(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(occurrences) != 2 || occurrences[1].OriginalName != "b.txt" || occurrences[1].AuditChangeID == "" {
		t.Fatalf("duplicate occurrence did not survive encrypted workspace reload: %+v", occurrences)
	}
}

func TestEvidenceOccurrencesCountRepeatedVerifiedImports(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "same.txt")
	if err := os.WriteFile(path, []byte("same path supplied repeatedly"), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := OpenVault(filepath.Join(root, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	item, _, err := v.ImportFile(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, duplicate, err := v.ImportFile(path, nil); err != nil || !duplicate {
			t.Fatalf("repeat %d err=%v duplicate=%v", i+1, err, duplicate)
		}
	}
	occurrences, err := v.EvidenceOccurrences(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(occurrences) != 3 {
		t.Fatalf("expected one initial plus two verified repeat occurrences, got %+v", occurrences)
	}
	if occurrences[1].ID == occurrences[2].ID {
		t.Fatalf("repeat occurrences reused an identifier: %+v", occurrences)
	}
}

func TestEvidenceOccurrencesSynthesiseLegacyInitialWithoutMigration(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "legacy.txt")
	if err := os.WriteFile(path, []byte("legacy initial occurrence"), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := OpenVault(filepath.Join(root, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	item, _, err := v.ImportFile(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	beforeChanges := len(v.Snapshot().Changes)
	occurrences, err := v.EvidenceOccurrences(item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(occurrences) != 1 || occurrences[0].Kind != occurrenceKindInitial || occurrences[0].AuditChangeID != "" {
		t.Fatalf("legacy initial occurrence synthesis is wrong: %+v", occurrences)
	}
	if len(v.Snapshot().Changes) != beforeChanges {
		t.Fatal("read-only occurrence query rewrote the workspace")
	}
}

func TestEvidenceOccurrencesRejectMalformedOccurrenceAudit(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "source.txt")
	if err := os.WriteFile(path, []byte("malformed occurrence audit test"), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := OpenVault(filepath.Join(root, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	item, _, err := v.ImportFile(path, nil)
	if err != nil {
		t.Fatal(err)
	}

	v.mu.Lock()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	v.addChangeUnlocked("system", occurrenceChangeType, "malformed test occurrence", map[string]any{
		"occurrence_id":                NewID("OCC"),
		"evidence_id":                  item.ID,
		"kind":                         occurrenceKindDuplicate,
		"source_path":                  path,
		"original_name":                filepath.Base(path),
		"source_size":                  "999999",
		"source_sha256":                item.SHA256,
		"object_file":                  item.ObjectFile,
		"observed_at":                  now,
		"source_verified_at":           now,
		"preserved_object_verified_at": now,
		"duplicate_reused_object":      true,
	})
	if err := v.saveUnlocked(); err != nil {
		v.mu.Unlock()
		t.Fatal(err)
	}
	v.mu.Unlock()

	if _, err := v.EvidenceOccurrences(item.ID); err == nil || !strings.Contains(err.Error(), "inconsistent with evidence") {
		t.Fatalf("expected malformed occurrence ledger rejection, got %v", err)
	}
}

func TestEvidenceOccurrencesUnknownEvidenceFails(t *testing.T) {
	v, err := OpenVault(filepath.Join(t.TempDir(), "vault"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.EvidenceOccurrences("EVD-does-not-exist"); !os.IsNotExist(err) {
		t.Fatalf("expected missing evidence error, got %v", err)
	}
}
