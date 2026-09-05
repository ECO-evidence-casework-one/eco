package eco

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

type pdfcpuInspectionRunner func(context.Context, string, string, SourceReceipt) (PDFAssessment, error)

// InspectEvidencePDFWithPDFCPU performs a source-verified, read-only structural
// inspection of one already-preserved ECO PDF. The inspection result is
// returned to the caller and a compact audit event is persisted; no modified
// PDF and no plaintext derivative database are created.
func (v *Vault) InspectEvidencePDFWithPDFCPU(evidenceID, executable string) (PDFAssessment, error) {
	return v.InspectEvidencePDFWithPDFCPUContext(context.Background(), evidenceID, executable)
}

func (v *Vault) InspectEvidencePDFWithPDFCPUContext(ctx context.Context, evidenceID, executable string) (PDFAssessment, error) {
	return v.inspectEvidencePDFWithRunner(ctx, evidenceID, executable, RunPDFCPUInspection)
}

func (v *Vault) inspectEvidencePDFWithRunner(ctx context.Context, evidenceID, executable string, runner pdfcpuInspectionRunner) (PDFAssessment, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(evidenceID) == "" {
		return PDFAssessment{}, errors.New("pdfcpu evidence ID is required")
	}
	if runner == nil {
		return PDFAssessment{}, errors.New("pdfcpu inspection runner is required")
	}
	item, record, err := v.pdfcpuSource(evidenceID)
	if err != nil {
		return PDFAssessment{}, err
	}

	var assessment PDFAssessment
	err = v.withVerifiedPreservedFile(record, item.SHA256, func(path string, source SourceReceipt) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		result, runErr := runner(ctx, executable, path, source)
		if runErr != nil {
			return runErr
		}
		assessment = result
		return nil
	})
	if err != nil {
		return assessment, err
	}
	if assessment.SourceObject != item.ObjectFile || assessment.SourceSHA256 != item.SHA256 || assessment.CreatedAt.IsZero() {
		return assessment, errors.New("pdfcpu assessment is not bound to the verified preserved PDF")
	}
	if assessment.RelaxedValidationPassed && (strings.TrimSpace(assessment.Version) == "" || assessment.PageCount < 1) {
		return assessment, errors.New("pdfcpu assessment passed relaxed validation but lacks PDF version/page count")
	}
	if err := v.recordPDFCPUAssessment(item, assessment); err != nil {
		return assessment, err
	}
	return assessment, nil
}

func (v *Vault) pdfcpuSource(evidenceID string) (EvidenceItem, PreservationRecord, error) {
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
		return EvidenceItem{}, PreservationRecord{}, errors.New("pdfcpu inspection is blocked because the preserved source is not verified")
	}
	if item.DetectedType != "pdf" {
		return EvidenceItem{}, PreservationRecord{}, fmt.Errorf("pdfcpu inspection requires detected PDF evidence, got %q", item.DetectedType)
	}
	for _, candidate := range ws.Preservations {
		if candidate.EvidenceID == evidenceID && candidate.State == preservationCommitted && candidate.ObjectFile == item.ObjectFile && candidate.PreservedSHA256 == item.SHA256 && candidate.ExpectedSize == item.Size {
			return item, candidate, nil
		}
	}
	return EvidenceItem{}, PreservationRecord{}, errors.New("pdfcpu inspection is blocked because the committed preservation record is missing or inconsistent")
}

func (v *Vault) recordPDFCPUAssessment(item EvidenceItem, assessment PDFAssessment) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	oldChanges := append([]ChangeRecord(nil), v.Workspace.Changes...)
	oldUpdatedAt := v.Workspace.UpdatedAt
	oldBuildID := v.Workspace.BuildID
	v.addChangeUnlocked("pdfcpu-worker", "pdf-structure-inspected", "Inspected preserved PDF structure for "+item.SafeName, map[string]any{
		"id":                          item.ID,
		"object_file":                 item.ObjectFile,
		"source_sha256":               item.SHA256,
		"engine":                      "pdfcpu",
		"engine_version":              truncate(assessment.EngineVersion, maxOCRIdentityText),
		"relaxed_validation_passed":   assessment.RelaxedValidationPassed,
		"strict_validation_passed":    assessment.StrictValidationPassed,
		"pdf_version":                 truncate(assessment.Version, maxOCRIdentityText),
		"page_count":                  assessment.PageCount,
		"encrypted":                   assessment.Encrypted,
		"signatures":                  assessment.Signatures,
		"form":                        assessment.Form,
		"attachments":                 assessment.AttachmentCount,
		"inspection_is_content_truth": false,
	})
	if err := v.saveUnlocked(); err != nil {
		v.Workspace.Changes = oldChanges
		v.Workspace.UpdatedAt = oldUpdatedAt
		v.Workspace.BuildID = oldBuildID
		return fmt.Errorf("persist pdfcpu inspection audit: %w", err)
	}
	return nil
}
