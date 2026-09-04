package eco

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// ExtractEvidenceWithDocling performs a source-verified local Docling reading
// and commits only the derived text/segments to ECO. The preserved evidence
// object remains unchanged and authoritative.
func (v *Vault) ExtractEvidenceWithDocling(evidenceID, executable, artifactsPath string) error {
	return v.ExtractEvidenceWithDoclingContext(context.Background(), evidenceID, executable, artifactsPath)
}

func (v *Vault) ExtractEvidenceWithDoclingContext(ctx context.Context, evidenceID, executable, artifactsPath string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(evidenceID) == "" {
		return errors.New("Docling evidence ID is required")
	}
	item, record, err := v.doclingSource(evidenceID)
	if err != nil {
		return err
	}
	if item.DetectedType == "executable" || item.Status == "Quarantined" {
		return errors.New("Docling extraction is blocked for quarantined or executable evidence")
	}

	var result DoclingExtractionResult
	err = v.withVerifiedPreservedFile(record, item.SHA256, func(path string, source SourceReceipt) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		run, runErr := RunDoclingExtraction(ctx, executable, path, artifactsPath, source)
		if runErr != nil {
			return runErr
		}
		result = run
		return nil
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("Docling extraction stopped: %w", err)
		}
		return err
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("Docling extraction stopped before result commit: %w", err)
	}

	status := "ready"
	if strings.TrimSpace(result.Text) == "" {
		status = "no-text"
	}
	receipt := ExtractionReceipt{
		SourceObject: item.ObjectFile,
		SourceSHA256: item.SHA256,
		DetectedType: item.DetectedType,
		Status:       status,
		CreatedAt:    time.Now().UTC(),
		Segments:     len(result.Segments),
	}
	return v.applyDoclingExtraction(evidenceID, result.EngineVersion, receipt, result.Text, result.Segments, result.Warnings)
}

func (v *Vault) doclingSource(evidenceID string) (EvidenceItem, PreservationRecord, error) {
	ws := v.Snapshot()
	var item EvidenceItem
	found := false
	for _, candidate := range ws.Evidence {
		if candidate.ID == evidenceID {
			item = cloneEvidenceItem(candidate)
			found = true
			break
		}
	}
	if !found {
		return EvidenceItem{}, PreservationRecord{}, os.ErrNotExist
	}
	if !preservationUsable(item) {
		return EvidenceItem{}, PreservationRecord{}, errors.New("Docling extraction is blocked because the preserved source is not verified")
	}
	for _, candidate := range ws.Preservations {
		if candidate.EvidenceID != evidenceID {
			continue
		}
		if candidate.State != preservationCommitted || candidate.ObjectFile != item.ObjectFile || candidate.PreservedSHA256 != item.SHA256 || candidate.ExpectedSize != item.Size {
			continue
		}
		return item, candidate, nil
	}
	return EvidenceItem{}, PreservationRecord{}, errors.New("Docling extraction is blocked because the committed preservation record is missing or inconsistent")
}

func (v *Vault) applyDoclingExtraction(evidenceID, engineVersion string, receipt ExtractionReceipt, text string, segments []SourceSegment, warnings []string) error {
	if receipt.Status != "ready" && receipt.Status != "no-text" {
		return errors.New("Docling extraction receipt status is invalid")
	}
	if receipt.SourceObject == "" || !sha256TextPattern.MatchString(receipt.SourceSHA256) || receipt.CreatedAt.IsZero() {
		return errors.New("Docling extraction receipt is incomplete")
	}
	if len(text) > int(maxExtractBytes) {
		return errors.New("Docling extracted text exceeds ECO's safe size limit")
	}
	for _, segment := range segments {
		if segment.Origin != "docling" || segment.SourceObject != receipt.SourceObject || segment.SourceSHA256 != receipt.SourceSHA256 {
			return errors.New("Docling source segment is not bound to the verified preserved object")
		}
	}

	v.opMu.RLock()
	defer v.opMu.RUnlock()
	v.mu.Lock()
	var source EvidenceItem
	found := false
	for i := range v.Workspace.Evidence {
		if v.Workspace.Evidence[i].ID == evidenceID {
			source = cloneEvidenceItem(v.Workspace.Evidence[i])
			found = true
			break
		}
	}
	v.mu.Unlock()
	if !found {
		return os.ErrNotExist
	}
	if !preservationUsable(source) || receipt.SourceObject != source.ObjectFile || receipt.SourceSHA256 != source.SHA256 || receipt.DetectedType != source.DetectedType {
		return errors.New("Docling extraction receipt does not match the verified preserved evidence")
	}
	if _, err := v.verifyPreservedObject(source.ID, source.ObjectFile, source.SHA256, source.Size); err != nil {
		v.markEvidenceVerificationFailure(source.ID, err)
		return fmt.Errorf("Docling extraction is blocked because preserved source verification failed: %w", err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	for i := range v.Workspace.Evidence {
		item := &v.Workspace.Evidence[i]
		if item.ID != evidenceID {
			continue
		}
		if !preservationUsable(*item) || receipt.SourceObject != item.ObjectFile || receipt.SourceSHA256 != item.SHA256 || receipt.DetectedType != item.DetectedType {
			return errors.New("Docling extraction source changed before commit")
		}

		oldItem := cloneEvidenceItem(*item)
		oldChangeLen := len(v.Workspace.Changes)
		oldUpdatedAt := v.Workspace.UpdatedAt
		oldBuildID := v.Workspace.BuildID

		kept := make([]SourceSegment, 0, len(item.Segments)+len(segments))
		for _, seg := range item.Segments {
			if seg.Origin != "extraction" && seg.Origin != "docling" {
				kept = append(kept, seg)
			}
		}
		for _, seg := range segments {
			copySeg := seg
			if seg.Region != nil {
				region := *seg.Region
				copySeg.Region = &region
			}
			kept = append(kept, copySeg)
		}
		item.Segments = kept
		item.ExtractedText = text
		copyReceipt := receipt
		item.Extraction = &copyReceipt
		item.Readable = strings.TrimSpace(text) != "" || len(kept) > len(segments)
		if receipt.Status == "ready" {
			item.Status = "Ready — Docling extracted"
		} else {
			item.Status = "Docling found no text — original preserved"
		}
		item.Warnings = append(item.Warnings, warnings...)
		v.addChangeUnlocked("docling-worker", "docling-extraction-added", "Added source-bound Docling reading for "+item.SafeName, map[string]any{
			"id": item.ID, "object_file": item.ObjectFile, "source_sha256": item.SHA256,
			"engine": "docling", "engine_version": engineVersion, "status": receipt.Status, "segments": len(segments),
		})
		if err := v.saveUnlocked(); err != nil {
			*item = oldItem
			v.Workspace.Changes = v.Workspace.Changes[:oldChangeLen]
			v.Workspace.UpdatedAt = oldUpdatedAt
			v.Workspace.BuildID = oldBuildID
			return fmt.Errorf("persist Docling extraction: %w", err)
		}
		return nil
	}
	return os.ErrNotExist
}
