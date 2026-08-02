package eco

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	preservationIntake     = "intake"
	preservationPreserving = "preserving"
	preservationVerifying  = "verifying"
	preservationCommitted  = "committed"
	preservationFailed     = "failed"
)

var (
	errSourceChanged       = errors.New("source changed during intake; preservation stopped without creating usable evidence")
	errPreservationStopped = errors.New("preservation was interrupted; no usable evidence was created")
)

type preservedAnalysis struct {
	detection  Detection
	text       string
	segments   []SourceSegment
	warnings   []string
	image      *ImageAssessment
	extraction ExtractionReceipt
}

func preservationUsable(item EvidenceItem) bool {
	return item.Preservation == preservationCommitted && item.SourceVerified && item.VerificationError == "" && len(item.SHA256) == sha256.Size*2 && item.ObjectFile != ""
}

func segmentBoundToPreservedSource(item EvidenceItem, segment SourceSegment) bool {
	return preservationUsable(item) && segment.SourceObject == item.ObjectFile && segment.SourceSHA256 == item.SHA256
}

func (v *Vault) beginPreservation(path string, info os.FileInfo) (PreservationRecord, error) {
	now := time.Now().UTC()
	evidenceID := NewID("EVD")
	record := PreservationRecord{
		ID:           NewID("PRS"),
		EvidenceID:   evidenceID,
		ObjectFile:   evidenceID + ".ecoobj",
		OriginalName: info.Name(),
		SafeName:     SafeDisplayName(info.Name()),
		SourcePath:   path,
		State:        preservationIntake,
		ExpectedSize: info.Size(),
		StartedAt:    now,
		UpdatedAt:    now,
	}

	v.mu.Lock()
	oldPreservations := append([]PreservationRecord(nil), v.Workspace.Preservations...)
	oldChanges := append([]ChangeRecord(nil), v.Workspace.Changes...)
	oldUpdatedAt := v.Workspace.UpdatedAt
	oldBuildID := v.Workspace.BuildID
	v.Workspace.Preservations = append([]PreservationRecord{record}, v.Workspace.Preservations...)
	v.addChangeUnlocked("system", "evidence-preservation-started", "Started immutable preservation for "+record.SafeName, map[string]any{"preservation_id": record.ID, "evidence_id": record.EvidenceID, "size": record.ExpectedSize})
	if err := v.saveUnlocked(); err != nil {
		v.Workspace.Preservations = oldPreservations
		v.Workspace.Changes = oldChanges
		v.Workspace.UpdatedAt = oldUpdatedAt
		v.Workspace.BuildID = oldBuildID
		v.mu.Unlock()
		return PreservationRecord{}, fmt.Errorf("record preservation start: %w", err)
	}
	v.mu.Unlock()
	return record, nil
}

func (v *Vault) updatePreservation(record PreservationRecord) error {
	record.UpdatedAt = time.Now().UTC()
	v.mu.Lock()
	defer v.mu.Unlock()
	for i := range v.Workspace.Preservations {
		if v.Workspace.Preservations[i].ID == record.ID {
			old := v.Workspace.Preservations[i]
			oldUpdatedAt := v.Workspace.UpdatedAt
			oldBuildID := v.Workspace.BuildID
			v.Workspace.Preservations[i] = record
			if err := v.saveUnlocked(); err != nil {
				v.Workspace.Preservations[i] = old
				v.Workspace.UpdatedAt = oldUpdatedAt
				v.Workspace.BuildID = oldBuildID
				return err
			}
			return nil
		}
	}
	return os.ErrNotExist
}

func (v *Vault) failPreservation(record PreservationRecord, cause error) error {
	if cause == nil {
		cause = errors.New("preservation failed")
	}
	record.State = preservationFailed
	record.Error = cause.Error()
	record.UpdatedAt = time.Now().UTC()
	v.mu.Lock()
	defer v.mu.Unlock()
	for i := range v.Workspace.Preservations {
		if v.Workspace.Preservations[i].ID != record.ID {
			continue
		}
		v.Workspace.Preservations[i] = record
		v.addChangeUnlocked("system", "evidence-preservation-failed", "Preservation stopped safely for "+record.SafeName, map[string]any{"preservation_id": record.ID, "evidence_id": record.EvidenceID, "state": record.State, "reason": truncate(cause.Error(), 500), "bytes_preserved": record.BytesPreserved})
		if err := v.saveUnlocked(); err != nil {
			return fmt.Errorf("%w (failure state could not be persisted: %v)", cause, err)
		}
		return cause
	}
	return cause
}

func (v *Vault) commitPreservation(record PreservationRecord, item EvidenceItem) error {
	now := time.Now().UTC()
	record.State = preservationCommitted
	record.Error = ""
	record.UpdatedAt = now
	if record.VerifiedAt.IsZero() {
		record.VerifiedAt = now
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	recordIndex := -1
	for i := range v.Workspace.Preservations {
		if v.Workspace.Preservations[i].ID == record.ID {
			recordIndex = i
			break
		}
	}
	if recordIndex < 0 {
		return os.ErrNotExist
	}
	for _, existing := range v.Workspace.Evidence {
		if existing.ID == item.ID {
			v.Workspace.Preservations[recordIndex] = record
			return v.saveUnlocked()
		}
	}

	oldRecord := v.Workspace.Preservations[recordIndex]
	oldEvidence := append([]EvidenceItem(nil), v.Workspace.Evidence...)
	oldSelectedID := v.Workspace.SelectedID
	oldChanges := append([]ChangeRecord(nil), v.Workspace.Changes...)
	oldUpdatedAt := v.Workspace.UpdatedAt
	oldBuildID := v.Workspace.BuildID
	v.Workspace.Preservations[recordIndex] = record
	v.Workspace.Evidence = append([]EvidenceItem{item}, v.Workspace.Evidence...)
	v.Workspace.SelectedID = item.ID
	v.addChangeUnlocked("system", "evidence-preserved", "Preserved and verified "+item.SafeName, map[string]any{"preservation_id": record.ID, "id": item.ID, "object_file": item.ObjectFile, "source_sha256": item.SHA256, "type": item.DetectedType, "size": item.Size})
	if err := v.saveUnlocked(); err != nil {
		v.Workspace.Preservations[recordIndex] = oldRecord
		v.Workspace.Evidence = oldEvidence
		v.Workspace.SelectedID = oldSelectedID
		v.Workspace.Changes = oldChanges
		v.Workspace.UpdatedAt = oldUpdatedAt
		v.Workspace.BuildID = oldBuildID
		return fmt.Errorf("commit verified evidence record: %w", err)
	}
	return nil
}

func fingerprintSource(ctx context.Context, path string, initial os.FileInfo, progress func(ImportProgress)) (string, int64, os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, nil, err
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil {
		return "", 0, nil, err
	}
	if !sameStableFile(initial, before) {
		return "", 0, nil, errSourceChanged
	}

	h := sha256.New()
	buf := make([]byte, chunkSize)
	var done int64
	for {
		if err := ctx.Err(); err != nil {
			return "", done, before, fmt.Errorf("%w: %v", errPreservationStopped, err)
		}
		n, readErr := f.Read(buf)
		if n > 0 {
			_, _ = h.Write(buf[:n])
			done += int64(n)
			if progress != nil {
				progress(ImportProgress{Path: path, Name: initial.Name(), Stage: "Fingerprinting intake", Current: done, Total: initial.Size()})
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", done, before, readErr
		}
	}
	after, statErr := f.Stat()
	current, pathErr := os.Stat(path)
	if statErr != nil || pathErr != nil || done != initial.Size() || !sameStableFile(before, after) || !sameStableFile(after, current) {
		return "", done, before, errSourceChanged
	}
	return hex.EncodeToString(h.Sum(nil)), done, before, nil
}

func sameStableFile(a, b os.FileInfo) bool {
	if a == nil || b == nil || !a.Mode().IsRegular() || !b.Mode().IsRegular() {
		return false
	}
	return os.SameFile(a, b) && a.Size() == b.Size() && a.ModTime().Equal(b.ModTime())
}

func (v *Vault) preserveSource(ctx context.Context, path string, sourceInfo os.FileInfo, record PreservationRecord, progress func(ImportProgress)) (int64, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, "", err
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil {
		return 0, "", err
	}
	if !sameStableFile(sourceInfo, before) {
		return 0, "", errSourceChanged
	}
	objectPath, err := v.objectPath(record.EvidenceID, record.ObjectFile)
	if err != nil {
		return 0, "", err
	}
	done, copiedHash, err := encryptStreamContext(ctx, v.key, record.EvidenceID, f, objectPath, record.ExpectedSize, progress, path, record.OriginalName)
	if err != nil {
		return done, copiedHash, err
	}
	after, statErr := f.Stat()
	current, pathErr := os.Stat(path)
	if statErr != nil || pathErr != nil || done != record.ExpectedSize || !sameStableFile(before, after) || !sameStableFile(after, current) || copiedHash != record.IntakeSHA256 {
		return done, copiedHash, errSourceChanged
	}
	return done, copiedHash, nil
}

func (v *Vault) objectPath(evidenceID, objectFile string) (string, error) {
	if evidenceID == "" || objectFile != evidenceID+".ecoobj" || filepath.Base(objectFile) != objectFile {
		return "", errors.New("preserved object identity is invalid")
	}
	return filepath.Join(v.Objects, objectFile), nil
}

type countingHashWriter struct {
	w io.Writer
	h *sha256State
	n int64
}

// sha256State keeps the concrete hash implementation behind the small writer
// used by both object verification and materialisation.
type sha256State struct {
	h   io.Writer
	sum func() string
}

func newSHA256State() *sha256State {
	h := sha256.New()
	return &sha256State{h: h, sum: func() string { return hex.EncodeToString(h.Sum(nil)) }}
}

func (w *countingHashWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if n > 0 {
		_, _ = w.h.h.Write(p[:n])
		w.n += int64(n)
	}
	return n, err
}

func (v *Vault) verifyPreservedObject(evidenceID, objectFile, expectedHash string, expectedSize int64) (SourceReceipt, error) {
	objectPath, err := v.objectPath(evidenceID, objectFile)
	if err != nil {
		return SourceReceipt{}, err
	}
	f, err := os.Open(objectPath)
	if err != nil {
		return SourceReceipt{}, err
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil || !before.Mode().IsRegular() {
		return SourceReceipt{}, errors.New("preserved object is not a regular immutable file")
	}
	h := newSHA256State()
	w := &countingHashWriter{w: io.Discard, h: h}
	if err := decryptObjectToWriter(v.key, evidenceID, f, expectedSize, w); err != nil {
		return SourceReceipt{}, err
	}
	after, statErr := f.Stat()
	current, pathErr := os.Stat(objectPath)
	if statErr != nil || pathErr != nil || !sameStableFile(before, after) || !sameStableFile(after, current) {
		return SourceReceipt{}, errors.New("preserved object was replaced or mutated during verification")
	}
	actualHash := h.sum()
	if w.n != expectedSize {
		return SourceReceipt{}, fmt.Errorf("preserved object size mismatch: expected %d bytes, verified %d", expectedSize, w.n)
	}
	if expectedHash == "" || actualHash != expectedHash {
		return SourceReceipt{}, fmt.Errorf("preserved object SHA-256 mismatch: expected %s, verified %s", expectedHash, actualHash)
	}
	if err := os.Chmod(objectPath, 0400); err != nil {
		return SourceReceipt{}, fmt.Errorf("make preserved object immutable: %w", err)
	}
	return SourceReceipt{EvidenceID: evidenceID, ObjectFile: objectFile, SHA256: actualHash, Size: w.n, VerifiedAt: time.Now().UTC()}, nil
}

func (v *Vault) withVerifiedPreservedFile(record PreservationRecord, expectedHash string, fn func(string, SourceReceipt) error) error {
	objectPath, err := v.objectPath(record.EvidenceID, record.ObjectFile)
	if err != nil {
		return err
	}
	f, err := os.Open(objectPath)
	if err != nil {
		return err
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil || !before.Mode().IsRegular() {
		return errors.New("preserved object is not a regular immutable file")
	}

	workDir := filepath.Join(v.Root, "derived", ".work")
	if err := os.MkdirAll(workDir, 0700); err != nil {
		return err
	}
	ext := filepath.Ext(record.OriginalName)
	if len(ext) > 16 {
		ext = ""
	}
	tmp, err := os.CreateTemp(workDir, "verified-reading-*"+ext)
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = os.Chmod(tmpPath, 0600)
		_ = os.Remove(tmpPath)
	}()
	h := newSHA256State()
	w := &countingHashWriter{w: tmp, h: h}
	if err := decryptObjectToWriter(v.key, record.EvidenceID, f, record.ExpectedSize, w); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	actualHash := h.sum()
	if w.n != record.ExpectedSize || expectedHash == "" || actualHash != expectedHash {
		return errors.New("verified reading copy does not match the preserved object receipt")
	}
	if err := os.Chmod(tmpPath, 0400); err != nil {
		return err
	}
	receipt := SourceReceipt{EvidenceID: record.EvidenceID, ObjectFile: record.ObjectFile, SHA256: actualHash, Size: w.n, VerifiedAt: time.Now().UTC()}
	if err := fn(tmpPath, receipt); err != nil {
		return err
	}
	derivedHash, err := hashFile(tmpPath)
	if err != nil || derivedHash != actualHash {
		return errors.New("verified reading copy was unexpectedly mutated")
	}
	after, statErr := f.Stat()
	current, pathErr := os.Stat(objectPath)
	if statErr != nil || pathErr != nil || !sameStableFile(before, after) || !sameStableFile(after, current) {
		return errors.New("preserved object was replaced or mutated during downstream processing")
	}
	return nil
}

func (v *Vault) analyzePreserved(ctx context.Context, record PreservationRecord, expectedHash string, progress func(ImportProgress)) (preservedAnalysis, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var analysis preservedAnalysis
	err := v.withVerifiedPreservedFile(record, expectedHash, func(path string, source SourceReceipt) error {
		if progress != nil {
			progress(ImportProgress{Path: record.SourcePath, Name: record.OriginalName, Stage: "Reading verified preserved object", Total: record.ExpectedSize})
		}
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("%w: %v", errPreservationStopped, err)
		}
		detection, err := DetectFile(path)
		if err != nil {
			return fmt.Errorf("detect preserved object type: %w", err)
		}
		analysis.detection = detection
		if detection.Warning != "" {
			analysis.warnings = append(analysis.warnings, detection.Warning)
		}
		analysis.extraction = ExtractionReceipt{SourceObject: source.ObjectFile, SourceSHA256: source.SHA256, DetectedType: detection.Type, Status: "not-applicable", CreatedAt: time.Now().UTC()}
		if detection.Dangerous {
			analysis.extraction.Status = "blocked"
			return nil
		}

		text, segments, warnings := ExtractReadable(path, detection.Type)
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("%w: %v", errPreservationStopped, err)
		}
		analysis.text = text
		analysis.warnings = append(analysis.warnings, warnings...)
		analysis.segments = segments
		for i := range analysis.segments {
			if analysis.segments[i].Origin == "" {
				analysis.segments[i].Origin = "extraction"
			}
			analysis.segments[i].SourceObject = source.ObjectFile
			analysis.segments[i].SourceSHA256 = source.SHA256
		}
		analysis.extraction.Segments = len(segments)
		if strings.TrimSpace(text) != "" {
			analysis.extraction.Status = "ready"
		} else if extractionSupported(detection.Type) {
			analysis.extraction.Status = "no-text"
		}

		if isImageType(detection.Type) {
			data, readErr := readFileBounded(path, 120*1024*1024)
			if readErr == nil {
				if img, _, decodeErr := DecodeSupportedImage(data); decodeErr == nil {
					assessment := AssessImage(img)
					assessment.SourceObject = source.ObjectFile
					assessment.SourceSHA256 = source.SHA256
					analysis.image = &assessment
					analysis.warnings = append(analysis.warnings, assessment.Warnings...)
				} else {
					analysis.warnings = append(analysis.warnings, "Image preserved, but this native preview could not decode it for visual assessment.")
				}
			}
		}
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("%w: %v", errPreservationStopped, err)
		}
		return nil
	})
	return analysis, err
}

func extractionSupported(typ string) bool {
	switch typ {
	case "text", "markdown", "csv", "json", "xml", "html", "rtf", "docx", "xlsx", "pptx", "odt", "ods", "odp", "eml", "zip":
		return true
	default:
		return false
	}
}

func evidenceItemFromPreservation(record PreservationRecord, analysis preservedAnalysis) EvidenceItem {
	status := "Preserved - contents not read"
	readable := strings.TrimSpace(analysis.text) != ""
	if analysis.detection.Dangerous {
		status = "Quarantined"
	} else if readable {
		status = "Ready"
	} else if isImageType(analysis.detection.Type) {
		status = "Image ready"
	}
	extraction := analysis.extraction
	return EvidenceItem{
		ID:               record.EvidenceID,
		OriginalName:     record.OriginalName,
		SafeName:         record.SafeName,
		SourcePath:       record.SourcePath,
		Size:             record.ExpectedSize,
		SHA256:           record.PreservedSHA256,
		DetectedType:     analysis.detection.Type,
		ExtensionType:    analysis.detection.ExtensionType,
		TypeMismatch:     analysis.detection.Mismatch,
		Readable:         readable,
		Status:           status,
		ImportedAt:       record.VerifiedAt,
		ObjectFile:       record.ObjectFile,
		Preservation:     preservationCommitted,
		SourceVerified:   true,
		SourceVerifiedAt: record.VerifiedAt,
		ExtractedText:    analysis.text,
		Extraction:       &extraction,
		Segments:         analysis.segments,
		Image:            analysis.image,
		Warnings:         analysis.warnings,
	}
}

func (v *Vault) recoverPreservations() error {
	// First migrate pre-control records by re-verifying the preserved bytes and
	// rebuilding their extraction/image derivatives from the object itself.
	if err := v.recoverLegacyEvidence(); err != nil {
		return err
	}

	ws := v.Snapshot()
	for _, record := range ws.Preservations {
		if record.State == preservationCommitted || record.State == preservationFailed {
			continue
		}
		if evidenceExists(ws.Evidence, record.EvidenceID) {
			record.State = preservationCommitted
			record.Error = ""
			if err := v.updatePreservation(record); err != nil {
				return err
			}
			continue
		}
		if record.IntakeSHA256 == "" || record.ExpectedSize < 0 {
			_ = v.failPreservation(record, errors.New("preservation was interrupted before a complete source fingerprint was recorded; no usable evidence was created"))
			continue
		}
		objectPath, err := v.objectPath(record.EvidenceID, record.ObjectFile)
		if err != nil {
			_ = v.failPreservation(record, err)
			continue
		}
		if _, err = os.Stat(objectPath); os.IsNotExist(err) {
			tmpPath := objectPath + ".tmp"
			if receipt, verifyErr := v.verifyObjectAtPath(record.EvidenceID, record.ObjectFile, tmpPath, record.IntakeSHA256, record.ExpectedSize); verifyErr == nil {
				if renameErr := os.Rename(tmpPath, objectPath); renameErr != nil {
					_ = v.failPreservation(record, fmt.Errorf("resume completed preservation: %w", renameErr))
					continue
				}
				_ = os.Chmod(objectPath, 0400)
				record.PreservedSHA256 = receipt.SHA256
			} else {
				_ = v.failPreservation(record, errors.New("preservation was interrupted with an incomplete object; no usable evidence was created"))
				continue
			}
		} else if err != nil {
			_ = v.failPreservation(record, err)
			continue
		}
		receipt, err := v.verifyPreservedObject(record.EvidenceID, record.ObjectFile, record.IntakeSHA256, record.ExpectedSize)
		if err != nil {
			_ = v.failPreservation(record, fmt.Errorf("recovered preservation verification failed: %w", err))
			continue
		}
		record.State = preservationVerifying
		record.PreservedSHA256 = receipt.SHA256
		record.BytesPreserved = receipt.Size
		record.VerifiedAt = receipt.VerifiedAt
		if err = v.updatePreservation(record); err != nil {
			return err
		}
		analysis, err := v.analyzePreserved(context.Background(), record, receipt.SHA256, nil)
		if err != nil {
			_ = v.failPreservation(record, fmt.Errorf("recovered preserved object could not be read safely: %w", err))
			continue
		}
		if err = v.commitPreservation(record, evidenceItemFromPreservation(record, analysis)); err != nil {
			return err
		}
	}
	for _, item := range v.Snapshot().Evidence {
		if !preservationUsable(item) {
			continue
		}
		if _, err := v.verifyPreservedObject(item.ID, item.ObjectFile, item.SHA256, item.Size); err != nil {
			v.markEvidenceVerificationFailure(item.ID, fmt.Errorf("restart source verification failed: %w", err))
		}
	}
	return nil
}

func cleanupInterruptedReadingCopies(root string) error {
	workDir := filepath.Join(root, "derived", ".work")
	entries, err := os.ReadDir(workDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "verified-reading-") {
			continue
		}
		path := filepath.Join(workDir, entry.Name())
		if filepath.Dir(path) != workDir {
			return errors.New("unsafe interrupted reading-copy path")
		}
		_ = os.Chmod(path, 0600)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove interrupted derived reading copy: %w", err)
		}
	}
	return nil
}

func (v *Vault) verifyObjectAtPath(evidenceID, objectFile, path, expectedHash string, expectedSize int64) (SourceReceipt, error) {
	f, err := os.Open(path)
	if err != nil {
		return SourceReceipt{}, err
	}
	defer f.Close()
	h := newSHA256State()
	w := &countingHashWriter{w: io.Discard, h: h}
	if err := decryptObjectToWriter(v.key, evidenceID, f, expectedSize, w); err != nil {
		return SourceReceipt{}, err
	}
	if w.n != expectedSize || h.sum() != expectedHash {
		return SourceReceipt{}, errors.New("incomplete preserved object")
	}
	return SourceReceipt{EvidenceID: evidenceID, ObjectFile: objectFile, SHA256: h.sum(), Size: w.n, VerifiedAt: time.Now().UTC()}, nil
}

func evidenceExists(items []EvidenceItem, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func (v *Vault) recoverLegacyEvidence() error {
	ws := v.Snapshot()
	for _, legacy := range ws.Evidence {
		if legacy.Preservation != "" {
			continue
		}
		record := PreservationRecord{ID: NewID("PRS"), EvidenceID: legacy.ID, ObjectFile: legacy.ObjectFile, OriginalName: legacy.OriginalName, SafeName: legacy.SafeName, SourcePath: legacy.SourcePath, State: preservationVerifying, ExpectedSize: legacy.Size, IntakeSHA256: legacy.SHA256, StartedAt: legacy.ImportedAt, UpdatedAt: time.Now().UTC()}
		receipt, err := v.verifyPreservedObject(legacy.ID, legacy.ObjectFile, legacy.SHA256, legacy.Size)
		if err != nil {
			v.markEvidenceVerificationFailure(legacy.ID, fmt.Errorf("legacy preserved object verification failed: %w", err))
			continue
		}
		record.PreservedSHA256 = receipt.SHA256
		record.VerifiedAt = receipt.VerifiedAt
		analysis, err := v.analyzePreserved(context.Background(), record, receipt.SHA256, nil)
		if err != nil {
			v.markEvidenceVerificationFailure(legacy.ID, err)
			continue
		}
		v.mu.Lock()
		for i := range v.Workspace.Evidence {
			if v.Workspace.Evidence[i].ID != legacy.ID {
				continue
			}
			migrated := evidenceItemFromPreservation(record, analysis)
			migrated.MatterIDs = append([]string(nil), legacy.MatterIDs...)
			migrated.Rotation = legacy.Rotation
			migrated.DuplicateOf = legacy.DuplicateOf
			migrated.NearDuplicateOf = legacy.NearDuplicateOf
			if legacy.OCR != nil && legacy.OCR.SourceSHA256 == legacy.SHA256 {
				ocr := cloneOCRReceipt(*legacy.OCR)
				ocr.SourceObject = legacy.ObjectFile
				migrated.OCR = &ocr
				for _, segment := range legacy.Segments {
					if segment.Origin == "ocr" {
						segment.SourceObject = legacy.ObjectFile
						segment.SourceSHA256 = legacy.SHA256
						migrated.Segments = append(migrated.Segments, segment)
					}
				}
				migrated.Readable = len(migrated.Segments) > 0
			}
			v.Workspace.Evidence[i] = migrated
			v.Workspace.Preservations = append(v.Workspace.Preservations, PreservationRecord{ID: record.ID, EvidenceID: record.EvidenceID, ObjectFile: record.ObjectFile, OriginalName: record.OriginalName, SafeName: record.SafeName, SourcePath: record.SourcePath, State: preservationCommitted, ExpectedSize: record.ExpectedSize, BytesPreserved: receipt.Size, IntakeSHA256: legacy.SHA256, PreservedSHA256: receipt.SHA256, StartedAt: record.StartedAt, UpdatedAt: receipt.VerifiedAt, VerifiedAt: receipt.VerifiedAt})
			v.addChangeUnlocked("system", "evidence-source-migrated", "Re-verified preserved source and rebuilt derived readings for "+migrated.SafeName, map[string]any{"id": migrated.ID, "object_file": migrated.ObjectFile, "source_sha256": migrated.SHA256})
			break
		}
		err = v.saveUnlocked()
		v.mu.Unlock()
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *Vault) markEvidenceVerificationFailure(evidenceID string, cause error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	for i := range v.Workspace.Evidence {
		item := &v.Workspace.Evidence[i]
		if item.ID != evidenceID {
			continue
		}
		item.SourceVerified = false
		item.SourceVerifiedAt = time.Time{}
		item.VerificationError = cause.Error()
		item.Status = "Source verification failed - indexing, citation and retrieval blocked"
		v.addChangeUnlocked("system", "evidence-source-verification-failed", "Blocked downstream use of "+item.SafeName, map[string]any{"id": item.ID, "object_file": item.ObjectFile, "expected_sha256": item.SHA256, "reason": truncate(cause.Error(), 500)})
		_ = v.saveUnlocked()
		return
	}
}

func (v *Vault) markEvidenceVerificationSuccess(evidenceID string, receipt SourceReceipt) {
	v.mu.Lock()
	defer v.mu.Unlock()
	for i := range v.Workspace.Evidence {
		item := &v.Workspace.Evidence[i]
		if item.ID != evidenceID {
			continue
		}
		item.Preservation = preservationCommitted
		item.SourceVerified = true
		item.SourceVerifiedAt = receipt.VerifiedAt
		item.VerificationError = ""
		if strings.HasPrefix(item.Status, "Source verification failed") {
			switch {
			case item.DetectedType == "windows-executable" || item.DetectedType == "script" || item.DetectedType == "shortcut":
				item.Status = "Quarantined"
			case len(item.Segments) > 0:
				item.Status = "Ready"
			case isImageType(item.DetectedType):
				item.Status = "Image ready"
			default:
				item.Status = "Preserved - contents not read"
			}
		}
		_ = v.saveUnlocked()
		return
	}
}

func (v *Vault) verifyEvidenceForUse(scopeIDs []string) int {
	allowed := make(map[string]bool, len(scopeIDs))
	for _, id := range scopeIDs {
		allowed[id] = true
	}
	useScope := len(scopeIDs) > 0
	ws := v.Snapshot()
	failures := 0
	for _, item := range ws.Evidence {
		if useScope && !allowed[item.ID] {
			continue
		}
		if !preservationUsable(item) {
			failures++
			continue
		}
		if _, err := v.verifyPreservedObject(item.ID, item.ObjectFile, item.SHA256, item.Size); err != nil {
			v.markEvidenceVerificationFailure(item.ID, err)
			failures++
		}
	}
	return failures
}
