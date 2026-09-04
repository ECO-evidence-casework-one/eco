package eco

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

type ocrmyPDFRunner func(context.Context, string, string, string, SourceReceipt) (OCRmyPDFResult, error)

// ExtractEvidenceWithOCRmyPDF adds page-aware OCR suggestions from a preserved
// PDF without replacing ECO's native/Docling reading. OCRmyPDF is invoked in
// sidecar-only mode and the temporary sidecar is consumed into the encrypted
// workspace; no derived PDF is retained.
func (v *Vault) ExtractEvidenceWithOCRmyPDF(evidenceID, executable, language string) (OCRmyPDFResult, error) {
	return v.ExtractEvidenceWithOCRmyPDFContext(context.Background(), evidenceID, executable, language)
}

func (v *Vault) ExtractEvidenceWithOCRmyPDFContext(ctx context.Context, evidenceID, executable, language string) (OCRmyPDFResult, error) {
	return v.extractEvidenceWithOCRmyPDFRunner(ctx, evidenceID, executable, language, RunOCRmyPDF)
}

func (v *Vault) extractEvidenceWithOCRmyPDFRunner(ctx context.Context, evidenceID, executable, language string, runner ocrmyPDFRunner) (OCRmyPDFResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(evidenceID) == "" {
		return OCRmyPDFResult{}, errors.New("OCRmyPDF evidence ID is required")
	}
	if runner == nil {
		return OCRmyPDFResult{}, errors.New("OCRmyPDF runner is required")
	}
	if !tesseractLanguagePattern.MatchString(language) {
		return OCRmyPDFResult{}, errors.New("OCRmyPDF Tesseract language selection is missing or invalid")
	}
	item, record, err := v.ocrmyPDFSource(evidenceID)
	if err != nil {
		return OCRmyPDFResult{}, err
	}

	var result OCRmyPDFResult
	err = v.withVerifiedPreservedFile(record, item.SHA256, func(path string, source SourceReceipt) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		run, runErr := runner(ctx, executable, path, language, source)
		result = run
		if runErr != nil {
			return runErr
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if err := validateOCRmyPDFResult(item, result); err != nil {
		return result, err
	}
	if err := v.applyOCRmyPDFSidecar(item.ID, result); err != nil {
		return result, err
	}
	return result, nil
}

func (v *Vault) ocrmyPDFSource(evidenceID string) (EvidenceItem, PreservationRecord, error) {
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
		return EvidenceItem{}, PreservationRecord{}, errors.New("OCRmyPDF extraction is blocked because the preserved source is not verified")
	}
	if item.DetectedType != "pdf" {
		return EvidenceItem{}, PreservationRecord{}, fmt.Errorf("OCRmyPDF extraction requires detected PDF evidence, got %q", item.DetectedType)
	}
	for _, candidate := range ws.Preservations {
		if candidate.EvidenceID == evidenceID && candidate.State == preservationCommitted && candidate.ObjectFile == item.ObjectFile && candidate.PreservedSHA256 == item.SHA256 && candidate.ExpectedSize == item.Size {
			return item, candidate, nil
		}
	}
	return EvidenceItem{}, PreservationRecord{}, errors.New("OCRmyPDF extraction is blocked because the committed preservation record is missing or inconsistent")
}

func validateOCRmyPDFResult(item EvidenceItem, result OCRmyPDFResult) error {
	if result.Source.ObjectFile != item.ObjectFile || result.Source.SHA256 != item.SHA256 || result.Source.VerifiedAt.IsZero() {
		return errors.New("OCRmyPDF result is not bound to the verified preserved PDF")
	}
	if strings.TrimSpace(result.EngineVersion) == "" || len([]rune(result.EngineVersion)) > maxOCRIdentityText {
		return errors.New("OCRmyPDF engine identity is missing or unbounded")
	}
	if result.PageCount < 1 || result.OCRPages < 0 || result.OCRPages > result.PageCount {
		return errors.New("OCRmyPDF result has invalid page counts")
	}
	if int64(len(result.Text)) > maxExtractBytes {
		return errors.New("OCRmyPDF result text exceeds ECO's extraction size limit")
	}
	seenIDs := map[string]bool{}
	for i, segment := range result.Segments {
		if !strings.HasPrefix(segment.ID, "SEG-OCRPDF-") || segment.Ordinal != i+1 || segment.Page < 1 || segment.Page > result.PageCount {
			return errors.New("OCRmyPDF source segment has invalid identity/page ordering")
		}
		if segment.Origin != "ocrmypdf" || segment.SourceObject != item.ObjectFile || segment.SourceSHA256 != item.SHA256 || strings.TrimSpace(segment.Text) == "" {
			return errors.New("OCRmyPDF source segment is not bound to the verified preserved PDF")
		}
		if segment.Region != nil || segment.Confidence != 0 {
			return errors.New("OCRmyPDF sidecar segment must not invent coordinates or confidence")
		}
		if seenIDs[segment.ID] {
			return errors.New("OCRmyPDF result contains duplicate segment IDs")
		}
		seenIDs[segment.ID] = true
	}
	return nil
}

func (v *Vault) applyOCRmyPDFSidecar(evidenceID string, result OCRmyPDFResult) error {
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
	if !preservationUsable(source) || source.DetectedType != "pdf" || result.Source.ObjectFile != source.ObjectFile || result.Source.SHA256 != source.SHA256 {
		return errors.New("OCRmyPDF result source changed before commit")
	}
	if _, err := v.verifyPreservedObject(source.ID, source.ObjectFile, source.SHA256, source.Size); err != nil {
		v.markEvidenceVerificationFailure(source.ID, err)
		return fmt.Errorf("OCRmyPDF commit blocked because preserved source verification failed: %w", err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	for i := range v.Workspace.Evidence {
		item := &v.Workspace.Evidence[i]
		if item.ID != evidenceID {
			continue
		}
		if !preservationUsable(*item) || item.DetectedType != "pdf" || result.Source.ObjectFile != item.ObjectFile || result.Source.SHA256 != item.SHA256 {
			return errors.New("OCRmyPDF source changed before result persistence")
		}

		oldItem := cloneEvidenceItem(*item)
		oldChangeLen := len(v.Workspace.Changes)
		oldUpdatedAt := v.Workspace.UpdatedAt
		oldBuildID := v.Workspace.BuildID

		kept := make([]SourceSegment, 0, len(item.Segments)+len(result.Segments))
		for _, segment := range item.Segments {
			if segment.Origin != "ocrmypdf" {
				kept = append(kept, segment)
			}
		}
		kept = append(kept, result.Segments...)
		item.Segments = kept
		wasReadable := item.Readable
		if len(result.Segments) > 0 {
			item.Readable = true
			if !wasReadable {
				item.Status = "Ready — OCRmyPDF OCR suggestions"
			}
		}
		for _, warning := range result.Warnings {
			item.Warnings = appendUniqueOCRmyPDFWarning(item.Warnings, warning)
		}

		details := map[string]any{
			"id":                        item.ID,
			"object_file":               item.ObjectFile,
			"source_sha256":             item.SHA256,
			"engine":                    "ocrmypdf",
			"engine_version":            truncate(result.EngineVersion, maxOCRIdentityText),
			"page_count":                result.PageCount,
			"ocr_pages":                 result.OCRPages,
			"segments":                  len(result.Segments),
			"sidecar_only":              true,
			"output_pdf_retained":       false,
			"replaces_existing_reading": false,
			"content_truth_verified":    false,
		}
		addResourceAuditDetails(details, result.Resources)
		v.addChangeUnlocked("ocrmypdf-worker", "ocrmypdf-sidecar-added", "Added page-bound OCRmyPDF sidecar suggestions for "+item.SafeName, details)
		if err := v.saveUnlocked(); err != nil {
			*item = oldItem
			v.Workspace.Changes = v.Workspace.Changes[:oldChangeLen]
			v.Workspace.UpdatedAt = oldUpdatedAt
			v.Workspace.BuildID = oldBuildID
			return fmt.Errorf("persist OCRmyPDF sidecar: %w", err)
		}
		return nil
	}
	return os.ErrNotExist
}

func appendUniqueOCRmyPDFWarning(existing []string, warning string) []string {
	warning = strings.TrimSpace(warning)
	if warning == "" {
		return existing
	}
	for _, current := range existing {
		if current == warning {
			return existing
		}
	}
	return append(existing, warning)
}
