package eco

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func createRecoverablePreservation(t *testing.T, root string, runtime RuntimeIdentity, content []byte) (*Vault, PreservationRecord) {
	t.Helper()
	vault, err := createVault(root, "Recoverable preservation workspace", runtime)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "recoverable-source.txt")
	if err = os.WriteFile(source, content, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	_, _, err = vault.ImportFileContext(ctx, source, func(progress ImportProgress) {
		if progress.Stage == "Preserved object verified" {
			cancel()
		}
	})
	if err == nil {
		t.Fatal("synthetic import did not stop in verified-awaiting-recovery state")
	}
	state := vault.Snapshot()
	if len(state.Evidence) != 0 || len(state.Preservations) != 1 || state.Preservations[0].State != preservationRecoverable {
		t.Fatalf("fixture is not a verified recoverable preservation: %+v", state)
	}
	return vault, state.Preservations[0]
}

func assertPendingPreservationBlocksDownstreamUse(t *testing.T, vault *Vault, record PreservationRecord) {
	t.Helper()
	if _, err := vault.ReadEvidence(record.EvidenceID, 1<<20); err == nil {
		t.Fatal("preview or extraction read a pending preservation")
	}
	ocr := OCRReceipt{
		Engine:       "synthetic-test-engine",
		Status:       "no-text",
		SourceObject: record.ObjectFile,
		SourceSHA256: record.PreservedSHA256,
		CreatedAt:    time.Now().UTC(),
	}
	if err := vault.ApplyOCRResult(record.EvidenceID, ocr, nil); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("OCR was not blocked before a usable evidence record existed: %v", err)
	}
	answer := vault.Ask("What does the synthetic preservation say?", []string{record.EvidenceID})
	if answer.EvidenceConsidered != 0 || answer.RetrievedSegments != 0 || len(answer.Citations) != 0 {
		t.Fatalf("Ask ECO or citation used pending preservation state: %+v", answer)
	}
}

func workspaceTreeDigest(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(data)
		result[relative] = hex.EncodeToString(digest[:])
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertRecoveredEvidenceReceipt(t *testing.T, vault *Vault, record PreservationRecord, content []byte) EvidenceItem {
	t.Helper()
	state := vault.Snapshot()
	if len(state.Evidence) != 1 || len(state.Preservations) != 1 || state.Preservations[0].State != preservationCommitted {
		t.Fatalf("preservation was not committed through the approved recovery path: %+v", state)
	}
	item := state.Evidence[0]
	preservation := state.Preservations[0]
	if !preservationUsable(item) || item.ID != record.EvidenceID || item.ObjectFile != record.ObjectFile || item.ObjectFile != preservation.ObjectFile || item.SHA256 != record.PreservedSHA256 || item.SHA256 != preservation.PreservedSHA256 {
		t.Fatalf("usable evidence is not bound to its preservation receipt: item=%+v preservation=%+v", item, preservation)
	}
	data, receipt, err := vault.ReadEvidenceSource(item.ID, 1<<20)
	if err != nil || !bytes.Equal(data, content) || receipt.ObjectFile != item.ObjectFile || receipt.SHA256 != item.SHA256 || receipt.EvidenceID != item.ID {
		t.Fatalf("fresh preserved-object verification did not match migrated evidence: receipt=%+v err=%v", receipt, err)
	}
	return item
}

func TestMigrationRecoversVerifiedPendingPreservationThroughIssueThreePath(t *testing.T) {
	root := filepath.Join(t.TempDir(), "schema-one")
	oldRuntime := runtimeFor("legacy-preservation-candidate", "ECO-LEGACY-PRESERVATION", 1)
	content := []byte("Synthetic migration preservation wording with a distinctive amber deadline.")
	legacy, record := createRecoverablePreservation(t, root, oldRuntime, content)
	assertPendingPreservationBlocksDownstreamUse(t, legacy, record)
	legacy.Close()
	before := workspaceTreeDigest(t, root)

	current := runtimeFor("current-preservation-candidate", "ECO-CURRENT-PRESERVATION", Schema)
	session, receipt, err := MigrateWorkspace(root, current)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, workspaceTreeDigest(t, receipt.Checkpoint)) {
		t.Fatal("successful migration changed the original checkpoint")
	}
	item := assertRecoveredEvidenceReceipt(t, session.Vault, record, content)
	if item.Extraction == nil || item.Extraction.SourceObject != item.ObjectFile || item.Extraction.SourceSHA256 != item.SHA256 {
		t.Fatalf("migrated extraction is not bound to the freshly verified source: %+v", item.Extraction)
	}
	for _, segment := range item.Segments {
		if segment.SourceObject != item.ObjectFile || segment.SourceSHA256 != item.SHA256 {
			t.Fatalf("migrated source segment lost its preservation binding: %+v", segment)
		}
	}
	answer := session.Vault.Ask("Summarise the synthetic migration preservation wording.", []string{item.ID})
	if answer.EvidenceConsidered != 1 || len(answer.Citations) == 0 || answer.Citations[0].SourceObject != item.ObjectFile || answer.Citations[0].SourceSHA256 != item.SHA256 {
		t.Fatalf("Ask ECO did not use the freshly verified migrated receipt: %+v", answer)
	}
}

func TestCorruptedPendingPreservationBlocksMigrationActivationAndRollsBack(t *testing.T) {
	root := filepath.Join(t.TempDir(), "schema-one-corrupt")
	oldRuntime := runtimeFor("legacy-corrupt-candidate", "ECO-LEGACY-CORRUPT", 1)
	content := []byte("Synthetic pending preservation that will be corrupted only in staging.")
	legacy, record := createRecoverablePreservation(t, root, oldRuntime, content)
	legacy.Close()
	before := workspaceTreeDigest(t, root)
	current := runtimeFor("current-corrupt-candidate", "ECO-CURRENT-CORRUPT", Schema)

	_, _, err := migrateWorkspace(root, current, func(phase MigrationPhase) error {
		if phase != migrationStageUnverified {
			return nil
		}
		state, key, readErr := readMigrationState(root)
		if readErr != nil {
			return readErr
		}
		zeroBytes(key)
		objectPath := filepath.Join(state.Stage, "objects", record.ObjectFile)
		if chmodErr := os.Chmod(objectPath, 0600); chmodErr != nil {
			return chmodErr
		}
		data, readErr := os.ReadFile(objectPath)
		if readErr != nil {
			return readErr
		}
		data[len(data)-1] ^= 0x3c
		return os.WriteFile(objectPath, data, 0600)
	})
	if err == nil {
		t.Fatal("corrupted pending preservation was activated")
	}
	if !reflect.DeepEqual(before, workspaceTreeDigest(t, root)) {
		t.Fatal("failed migration did not restore the original checkpoint exactly")
	}
	if _, statErr := os.Stat(migrationStatePath(root)); !os.IsNotExist(statErr) {
		t.Fatalf("completed rollback left a migration marker: %v", statErr)
	}
	restored, err := openVaultIgnoringRecovery(root, oldRuntime)
	if err != nil {
		t.Fatalf("restored original checkpoint is not usable: %v", err)
	}
	assertRecoveredEvidenceReceipt(t, restored, record, content)
}

func TestPortableBackupRestoresVerifiedPendingPreservationThroughIssueThreePath(t *testing.T) {
	runtime := runtimeFor("backup-preservation-candidate", "ECO-BACKUP-PRESERVATION", Schema)
	content := []byte("Synthetic portable backup pending preservation wording.")
	source, record := createRecoverablePreservation(t, filepath.Join(t.TempDir(), "source"), runtime, content)
	assertPendingPreservationBlocksDownstreamUse(t, source, record)
	backup := filepath.Join(t.TempDir(), "pending.ecobackup")
	if _, err := source.CreatePortableBackup(backup, "synthetic-backup-passphrase", nil); err != nil {
		t.Fatal(err)
	}

	target, err := createVault(filepath.Join(t.TempDir(), "target"), "Restore target", runtime)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := target.RestorePortableBackup(backup, "synthetic-backup-passphrase", nil)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.EvidenceItems != 1 {
		t.Fatalf("restore did not report the pending object recovered into usable evidence: %+v", receipt)
	}
	assertRecoveredEvidenceReceipt(t, target, record, content)
}

func TestPortableBackupCorruptedPendingStageCannotReplaceActiveWorkspace(t *testing.T) {
	runtime := runtimeFor("backup-corruption-candidate", "ECO-BACKUP-CORRUPTION", Schema)
	content := []byte("Synthetic portable backup object corrupted only after staging.")
	source, record := createRecoverablePreservation(t, filepath.Join(t.TempDir(), "source"), runtime, content)
	backup := filepath.Join(t.TempDir(), "pending.ecobackup")
	if _, err := source.CreatePortableBackup(backup, "synthetic-backup-passphrase", nil); err != nil {
		t.Fatal(err)
	}
	target, err := createVault(filepath.Join(t.TempDir(), "target"), "Unchanged restore target", runtime)
	if err != nil {
		t.Fatal(err)
	}
	existing := importSynthetic(t, target, t.TempDir(), "existing.txt", "Existing active workspace data must survive.")
	before := target.Snapshot()
	_, err = target.restorePortableBackup(backup, "synthetic-backup-passphrase", nil, func(phase RestorePhase, stage *Vault) error {
		if phase != restoreStageReady {
			return errors.New("unexpected restore test phase")
		}
		path := filepath.Join(stage.Objects, record.ObjectFile)
		if chmodErr := os.Chmod(path, 0600); chmodErr != nil {
			return chmodErr
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		data[len(data)-1] ^= 0x55
		return os.WriteFile(path, data, 0600)
	})
	if err == nil {
		t.Fatal("restore activated a corrupted pending preserved object")
	}
	after := target.Snapshot()
	if len(after.Evidence) != 1 || after.Evidence[0].ID != existing.ID || after.WorkspaceID != before.WorkspaceID {
		t.Fatalf("failed restore changed the active workspace: before=%+v after=%+v", before, after)
	}
	data, _, readErr := target.ReadEvidenceSource(existing.ID, 1<<20)
	if readErr != nil || string(data) != "Existing active workspace data must survive." {
		t.Fatalf("failed restore damaged active preserved evidence: data=%q err=%v", data, readErr)
	}
}
