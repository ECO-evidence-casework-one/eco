package eco

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMutationDuringIntakeCannotCreateEvidence(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "changing.bin")
	data := bytes.Repeat([]byte("A"), 3*chunkSize)
	if err := os.WriteFile(sourcePath, data, 0600); err != nil {
		t.Fatal(err)
	}
	vault, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	mutated := false
	_, _, err = vault.ImportFile(sourcePath, func(progress ImportProgress) {
		if progress.Stage != "Fingerprinting intake" || progress.Current < chunkSize || mutated {
			return
		}
		mutated = true
		f, openErr := os.OpenFile(sourcePath, os.O_WRONLY, 0600)
		if openErr != nil {
			t.Fatalf("open mutation target: %v", openErr)
		}
		if _, writeErr := f.WriteAt([]byte("changed-during-intake"), 2*chunkSize); writeErr != nil {
			_ = f.Close()
			t.Fatalf("mutate intake: %v", writeErr)
		}
		if closeErr := f.Close(); closeErr != nil {
			t.Fatalf("close mutation target: %v", closeErr)
		}
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "source changed") {
		t.Fatalf("changed intake was not stopped clearly: %v", err)
	}
	if !mutated {
		t.Fatal("test did not mutate the synthetic source")
	}
	if got := vault.Snapshot(); len(got.Evidence) != 0 {
		t.Fatalf("changed intake created usable evidence: %+v", got.Evidence)
	}
}

func TestInterruptedPreservationLeavesFailedNonUsableState(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "cancelled.bin")
	if err := os.WriteFile(sourcePath, bytes.Repeat([]byte("B"), 4*chunkSize), 0600); err != nil {
		t.Fatal(err)
	}
	vault, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancelled := false
	_, _, err = vault.ImportFileContext(ctx, sourcePath, func(progress ImportProgress) {
		if progress.Stage == "Preserving immutable original" && progress.Current >= chunkSize && !cancelled {
			cancelled = true
			cancel()
		}
	})
	if err == nil {
		t.Fatal("cancelled preservation was accepted")
	}
	state := vault.Snapshot()
	if len(state.Evidence) != 0 {
		t.Fatalf("cancelled preservation created usable evidence: %+v", state.Evidence)
	}
	if len(state.Preservations) != 1 || state.Preservations[0].State != preservationFailed || state.Preservations[0].Error == "" {
		t.Fatalf("cancelled preservation state is not truthful: %+v", state.Preservations)
	}
	objectPath, pathErr := vault.objectPath(state.Preservations[0].EvidenceID, state.Preservations[0].ObjectFile)
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	if _, statErr := os.Stat(objectPath); !os.IsNotExist(statErr) {
		t.Fatalf("cancelled preservation left a completed object: %v", statErr)
	}
}

func TestMutationDuringPreservationCannotCreateEvidence(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "changing-while-preserving.bin")
	data := bytes.Repeat([]byte("C"), 4*chunkSize)
	if err := os.WriteFile(sourcePath, data, 0600); err != nil {
		t.Fatal(err)
	}
	vault, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	mutated := false
	_, _, err = vault.ImportFile(sourcePath, func(progress ImportProgress) {
		if progress.Stage != "Preserving immutable original" || progress.Current < chunkSize || mutated {
			return
		}
		mutated = true
		f, openErr := os.OpenFile(sourcePath, os.O_WRONLY, 0600)
		if openErr != nil {
			t.Fatalf("open preservation mutation target: %v", openErr)
		}
		if _, writeErr := f.WriteAt([]byte("changed-while-preserving"), 3*chunkSize); writeErr != nil {
			_ = f.Close()
			t.Fatalf("mutate during preservation: %v", writeErr)
		}
		if closeErr := f.Close(); closeErr != nil {
			t.Fatalf("close preservation mutation target: %v", closeErr)
		}
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "source changed") {
		t.Fatalf("source mutation during preservation was not stopped: %v", err)
	}
	state := vault.Snapshot()
	if !mutated || len(state.Evidence) != 0 || len(state.Preservations) != 1 || state.Preservations[0].State != preservationFailed {
		t.Fatalf("mutation left an untruthful or usable state: mutated=%v state=%+v", mutated, state)
	}
}

func TestCancellationAfterPreservationDoesNotCreateUsableEvidence(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "cancel-after-preserve.txt")
	if err := os.WriteFile(sourcePath, []byte("synthetic verified bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	vault, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	_, _, err = vault.ImportFileContext(ctx, sourcePath, func(progress ImportProgress) {
		if progress.Stage == "Preserved object verified" {
			cancel()
		}
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "interrupted") {
		t.Fatalf("post-preservation cancellation was not stopped clearly: %v", err)
	}
	state := vault.Snapshot()
	if len(state.Evidence) != 0 || len(state.Preservations) != 1 || state.Preservations[0].State != preservationFailed {
		t.Fatalf("post-preservation cancellation created usable evidence: %+v", state)
	}
	objectPath := filepath.Join(vault.Objects, state.Preservations[0].ObjectFile)
	if _, err := os.Stat(objectPath); err != nil {
		t.Fatalf("verified original was not retained in recoverable failed state: %v", err)
	}
}

func TestRestartMarksIncompletePreservationFailed(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "vault")
	vault, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	record := PreservationRecord{
		ID:           NewID("PRS"),
		EvidenceID:   "EVD-INTERRUPTED",
		ObjectFile:   "EVD-INTERRUPTED.ecoobj",
		OriginalName: "interrupted.bin",
		SafeName:     "interrupted.bin",
		State:        preservationPreserving,
		ExpectedSize: int64(2 * chunkSize),
		IntakeSHA256: strings.Repeat("a", 64),
		StartedAt:    now,
		UpdatedAt:    now,
	}
	vault.mu.Lock()
	vault.Workspace.Preservations = append(vault.Workspace.Preservations, record)
	err = vault.saveUnlocked()
	vault.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	tmpObject := filepath.Join(vault.Objects, record.ObjectFile+".tmp")
	if err := os.WriteFile(tmpObject, []byte(objectMagic+"incomplete"), 0600); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	state := reopened.Snapshot()
	if len(state.Evidence) != 0 {
		t.Fatalf("restart exposed incomplete preservation as evidence: %+v", state.Evidence)
	}
	if len(state.Preservations) != 1 || state.Preservations[0].State != preservationFailed || !strings.Contains(strings.ToLower(state.Preservations[0].Error), "incomplete") {
		t.Fatalf("restart did not retain a truthful incomplete state: %+v", state.Preservations)
	}
	if _, err := os.Stat(tmpObject); err != nil {
		t.Fatalf("recoverable incomplete encrypted object was discarded: %v", err)
	}
}

func TestRestartRecoversCompleteUncommittedPreservation(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "vault")
	vault, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("A fully written synthetic preservation can be recovered after restart.")
	h := sha256.Sum256(content)
	digest := hex.EncodeToString(h[:])
	now := time.Now().UTC()
	record := PreservationRecord{
		ID:             NewID("PRS"),
		EvidenceID:     "EVD-RECOVERABLE",
		ObjectFile:     "EVD-RECOVERABLE.ecoobj",
		OriginalName:   "recoverable.txt",
		SafeName:       "recoverable.txt",
		State:          preservationPreserving,
		ExpectedSize:   int64(len(content)),
		BytesPreserved: int64(len(content)),
		IntakeSHA256:   digest,
		StartedAt:      now,
		UpdatedAt:      now,
	}
	vault.mu.Lock()
	vault.Workspace.Preservations = append(vault.Workspace.Preservations, record)
	err = vault.saveUnlocked()
	vault.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	objectPath := filepath.Join(vault.Objects, record.ObjectFile)
	if _, _, err := encryptStreamContext(context.Background(), vault.key, record.EvidenceID, bytes.NewReader(content), objectPath, int64(len(content)), nil, "", record.OriginalName); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(objectPath, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(objectPath, objectPath+".tmp"); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	state := reopened.Snapshot()
	if len(state.Evidence) != 1 || state.Evidence[0].SHA256 != digest || !state.Evidence[0].SourceVerified || state.Preservations[0].State != preservationCommitted {
		t.Fatalf("complete interrupted preservation was not recovered truthfully: %+v", state)
	}
	preserved, _, err := reopened.ReadEvidenceSource(record.EvidenceID, 1<<20)
	if err != nil || !bytes.Equal(preserved, content) {
		t.Fatalf("recovered preservation bytes differ: %v", err)
	}
}

func TestStoredSHA256MatchesExactPreservedBytes(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "hash.txt")
	content := []byte("synthetic preserved byte hash\nwith a second line\n")
	if err := os.WriteFile(sourcePath, content, 0600); err != nil {
		t.Fatal(err)
	}
	vault, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	item, _, err := vault.ImportFile(sourcePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	preserved, receipt, err := vault.ReadEvidenceSource(item.ID, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(preserved)
	actual := hex.EncodeToString(h[:])
	if actual != item.SHA256 || receipt.SHA256 != item.SHA256 || receipt.ObjectFile != item.ObjectFile || !bytes.Equal(preserved, content) {
		t.Fatalf("stored receipt does not match exact preserved bytes: item=%+v receipt=%+v actual=%s", item, receipt, actual)
	}
	if item.Preservation != preservationCommitted || !item.SourceVerified || item.SourceVerifiedAt.IsZero() {
		t.Fatalf("completed item lacks verified preservation state: %+v", item)
	}
	objectInfo, err := os.Stat(filepath.Join(vault.Objects, item.ObjectFile))
	if err != nil {
		t.Fatal(err)
	}
	if objectInfo.Mode().Perm()&0222 != 0 {
		t.Fatalf("preserved object remains writable: mode=%v", objectInfo.Mode())
	}
}

func TestExtractionUsesPreservedObjectAfterOriginalChanges(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "reading.txt")
	original := []byte("The preserved deadline is 14 October 2026.")
	if err := os.WriteFile(sourcePath, original, 0600); err != nil {
		t.Fatal(err)
	}
	vault, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	item, _, err := vault.ImportFile(sourcePath, func(progress ImportProgress) {
		if progress.Stage == "Preserved object verified" && !changed {
			changed = true
			if writeErr := os.WriteFile(sourcePath, []byte("The intake path now says the wrong deadline."), 0600); writeErr != nil {
				t.Fatalf("change original after preservation: %v", writeErr)
			}
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if !changed || !strings.Contains(item.ExtractedText, "14 October 2026") || strings.Contains(item.ExtractedText, "wrong deadline") {
		t.Fatalf("extraction did not use the preserved object: %q", item.ExtractedText)
	}
	if item.Extraction == nil || item.Extraction.SourceObject != item.ObjectFile || item.Extraction.SourceSHA256 != item.SHA256 {
		t.Fatalf("extraction receipt is not bound to the preserved source: %+v", item.Extraction)
	}
	for _, segment := range item.Segments {
		if !segmentBoundToPreservedSource(item, segment) {
			t.Fatalf("extracted segment is not bound to the preserved source: %+v", segment)
		}
	}
}

func TestOCRUsesVerifiedPreservedObjectReceipt(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "scan.png")
	img := image.NewRGBA(image.Rect(0, 0, 120, 80))
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, encoded.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	vault, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	item, _, err := vault.ImportFile(sourcePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte("not the preserved image"), 0600); err != nil {
		t.Fatal(err)
	}
	preserved, source, err := vault.ReadEvidenceSource(item.ID, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(preserved, encoded.Bytes()) {
		t.Fatal("OCR source bytes came from the changed intake path")
	}
	tsv := "level\tpage_num\tblock_num\tpar_num\tline_num\tword_num\tleft\ttop\twidth\theight\tconf\ttext\n5\t1\t1\t1\t1\t1\t10\t10\t50\t15\t93\tVerified"
	receipt, segments, err := ParseOCRTSV(tsv, "Synthetic", "1", "eng", source, 120, 80)
	if err != nil {
		t.Fatal(err)
	}
	if err := vault.ApplyOCRResult(item.ID, receipt, segments); err != nil {
		t.Fatal(err)
	}
	stored := vault.Snapshot().Evidence[0]
	if stored.OCR == nil || stored.OCR.SourceObject != item.ObjectFile || stored.OCR.SourceSHA256 != item.SHA256 {
		t.Fatalf("OCR receipt is not bound to the preserved source: %+v", stored.OCR)
	}
	for _, segment := range stored.Segments {
		if segment.Origin == "ocr" && !segmentBoundToPreservedSource(stored, segment) {
			t.Fatalf("OCR segment is not bound to the preserved source: %+v", segment)
		}
	}
}

func TestSourceMismatchBlocksIndexingCitationAndRetrieval(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "citation.txt")
	if err := os.WriteFile(sourcePath, []byte("The verified hearing is on 22 November 2026."), 0600); err != nil {
		t.Fatal(err)
	}
	vault, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	item, _, err := vault.ImportFile(sourcePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	objectPath := filepath.Join(vault.Objects, item.ObjectFile)
	if err := os.Chmod(objectPath, 0600); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(objectPath)
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 0x5a
	if err := os.WriteFile(objectPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	record := vault.Ask("When is the verified hearing?", nil)
	if record.RetrievedSegments != 0 || len(record.Citations) != 0 || record.SourceVerificationFailures == 0 {
		t.Fatalf("mismatched source was retrieved or cited: %+v", record)
	}
	state := vault.Snapshot()
	if state.Evidence[0].SourceVerified || !strings.Contains(state.Evidence[0].Status, "blocked") {
		t.Fatalf("mismatched source was not blocked: %+v", state.Evidence[0])
	}
	if ranked, _, _ := rankSegments("hearing", state.Evidence, nil); len(ranked) != 0 {
		t.Fatalf("mismatched source remained indexable: %+v", ranked)
	}
	if _, err := vault.ReadEvidence(item.ID, 1<<20); err == nil {
		t.Fatal("mismatched preserved source remained previewable")
	}
}

func TestLargeFileWithDeliberatelySlowPreservation(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "large-synthetic.bin")
	content := bytes.Repeat([]byte("0123456789abcdef"), (9*chunkSize)/16)
	if err := os.WriteFile(sourcePath, content, 0600); err != nil {
		t.Fatal(err)
	}
	vault, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	chunks := 0
	item, _, err := vault.ImportFile(sourcePath, func(progress ImportProgress) {
		if progress.Stage == "Preserving immutable original" && progress.Current > 0 {
			chunks++
			time.Sleep(2 * time.Millisecond)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if chunks < 8 {
		t.Fatalf("large-file preservation did not exercise multiple slow chunks: %d", chunks)
	}
	h := sha256.Sum256(content)
	if item.Size != int64(len(content)) || item.SHA256 != hex.EncodeToString(h[:]) {
		t.Fatalf("large preserved object receipt is wrong: %+v", item)
	}
	preserved, _, err := vault.ReadEvidenceSource(item.ID, int64(len(content)))
	if err != nil || !bytes.Equal(preserved, content) {
		t.Fatalf("large preserved object does not round trip: %v", err)
	}
}
