package eco

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// OCRImageWithTesseract runs local Tesseract against a freshly verified
// materialisation of one already-preserved ECO image and commits the
// coordinate-bearing result through ApplyOCRResult. The preserved encrypted
// object remains authoritative and is never passed to Tesseract directly.
func (v *Vault) OCRImageWithTesseract(evidenceID, executable, language string) error {
	return v.OCRImageWithTesseractContext(context.Background(), evidenceID, executable, language)
}

// OCRImageWithTesseractContext is the cancellable form of
// OCRImageWithTesseract. It performs no download or network operation.
func (v *Vault) OCRImageWithTesseractContext(ctx context.Context, evidenceID, executable, language string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if evidenceID == "" {
		return errors.New("OCR evidence ID is required")
	}

	item, record, err := v.tesseractOCRSource(evidenceID)
	if err != nil {
		return err
	}
	if item.Image == nil || item.Image.Width <= 0 || item.Image.Height <= 0 {
		return errors.New("Tesseract OCR requires a preserved image with verified dimensions")
	}

	var receipt OCRReceipt
	var segments []SourceSegment
	err = v.withVerifiedPreservedFile(record, item.SHA256, func(path string, source SourceReceipt) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		result, resultSegments, runErr := RunTesseractOCR(
			ctx,
			executable,
			path,
			language,
			source,
			item.Image.Width,
			item.Image.Height,
		)
		if runErr != nil {
			return runErr
		}
		receipt = result
		segments = resultSegments
		return nil
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("Tesseract OCR stopped: %w", err)
		}
		return err
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("Tesseract OCR stopped before result commit: %w", err)
	}
	return v.ApplyOCRResult(evidenceID, receipt, segments)
}

func (v *Vault) tesseractOCRSource(evidenceID string) (EvidenceItem, PreservationRecord, error) {
	ws := v.Snapshot()
	var item EvidenceItem
	foundItem := false
	for _, candidate := range ws.Evidence {
		if candidate.ID == evidenceID {
			item = cloneEvidenceItem(candidate)
			foundItem = true
			break
		}
	}
	if !foundItem {
		return EvidenceItem{}, PreservationRecord{}, os.ErrNotExist
	}
	if !preservationUsable(item) {
		return EvidenceItem{}, PreservationRecord{}, errors.New("OCR is blocked because the preserved source is not verified")
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
	return EvidenceItem{}, PreservationRecord{}, errors.New("OCR is blocked because the committed preservation record is missing or inconsistent")
}
