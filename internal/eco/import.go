package eco

import (
	"context"
	"errors"
	"fmt"
	"os"
)

func (v *Vault) ImportFile(path string, progress func(ImportProgress)) (EvidenceItem, bool, error) {
	return v.ImportFileContext(context.Background(), path, progress)
}

func (v *Vault) ImportFileContext(ctx context.Context, path string, progress func(ImportProgress)) (EvidenceItem, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	v.opMu.Lock()
	defer v.opMu.Unlock()
	info, err := os.Stat(path)
	if err != nil {
		return EvidenceItem{}, false, err
	}
	if !info.Mode().IsRegular() {
		return EvidenceItem{}, false, errors.New("only regular files can be imported")
	}
	if progress != nil {
		progress(ImportProgress{Path: path, Name: info.Name(), Stage: "Preparing immutable preservation", Total: info.Size()})
	}
	intakeHash, _, sourceInfo, err := fingerprintSource(ctx, path, info, progress)
	if err != nil {
		return EvidenceItem{}, false, err
	}

	for _, existing := range v.Snapshot().Evidence {
		if existing.SHA256 != intakeHash || !preservationUsable(existing) {
			continue
		}
		if _, verifyErr := v.verifyPreservedObject(existing.ID, existing.ObjectFile, existing.SHA256, existing.Size); verifyErr != nil {
			v.markEvidenceVerificationFailure(existing.ID, verifyErr)
			continue
		}
		if progress != nil {
			progress(ImportProgress{Path: path, Name: info.Name(), Stage: "Revalidating duplicate source", Total: info.Size()})
		}
		duplicateHash, _, _, sourceErr := fingerprintSource(ctx, path, sourceInfo, nil)
		if sourceErr != nil {
			return EvidenceItem{}, false, sourceErr
		}
		if duplicateHash != intakeHash {
			return EvidenceItem{}, false, errSourceChanged
		}
		return existing, true, nil
	}

	record, err := v.beginPreservation(path, info)
	if err != nil {
		return EvidenceItem{}, false, err
	}
	record.IntakeSHA256 = intakeHash
	record.State = preservationPreserving
	if err = v.updatePreservation(record); err != nil {
		return EvidenceItem{}, false, v.failPreservation(record, fmt.Errorf("record preservation fingerprint: %w", err))
	}
	if progress != nil {
		progress(ImportProgress{Path: path, Name: info.Name(), Stage: "Preserving immutable original", Total: info.Size()})
	}
	done, copiedHash, err := v.preserveSource(ctx, path, sourceInfo, record, progress)
	record.BytesPreserved = done
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			err = fmt.Errorf("%w: %v", errPreservationStopped, err)
		}
		return EvidenceItem{}, false, v.failPreservation(record, err)
	}
	if copiedHash != record.IntakeSHA256 {
		return EvidenceItem{}, false, v.failPreservation(record, errSourceChanged)
	}
	if err = ctx.Err(); err != nil {
		return EvidenceItem{}, false, v.failPreservation(record, fmt.Errorf("%w: %v", errPreservationStopped, err))
	}
	record.State = preservationVerifying
	if err = v.updatePreservation(record); err != nil {
		return EvidenceItem{}, false, fmt.Errorf("preserved object awaits restart recovery because verification state could not be saved: %w", err)
	}
	receipt, err := v.verifyPreservedObject(record.EvidenceID, record.ObjectFile, record.IntakeSHA256, record.ExpectedSize)
	if err != nil {
		return EvidenceItem{}, false, v.failPreservation(record, fmt.Errorf("preserved-byte verification failed: %w", err))
	}
	record.State = preservationRecoverable
	record.PreservedSHA256 = receipt.SHA256
	record.BytesPreserved = receipt.Size
	record.VerifiedAt = receipt.VerifiedAt
	if err = v.persistRecoverablePreservation(record, nil); err != nil {
		return EvidenceItem{}, false, fmt.Errorf("verified object awaits restart recovery because its receipt could not be saved: %w", err)
	}
	if progress != nil {
		progress(ImportProgress{Path: path, Name: info.Name(), Stage: "Preserved object verified", Current: receipt.Size, Total: receipt.Size})
	}
	if err = ctx.Err(); err != nil {
		return EvidenceItem{}, false, v.stopPreservationForRecovery(record, fmt.Errorf("%w: %v", errPreservationStopped, err))
	}
	analysis, err := v.analyzePreserved(ctx, record, receipt.SHA256, progress)
	if err != nil {
		if errors.Is(err, errPreservationStopped) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return EvidenceItem{}, false, v.stopPreservationForRecovery(record, err)
		}
		return EvidenceItem{}, false, v.failPreservation(record, fmt.Errorf("verified preserved object could not be processed safely: %w", err))
	}
	item := evidenceItemFromPreservation(record, analysis)
	if item.Image != nil && item.Image.PerceptualHash != "" {
		for _, existing := range v.Snapshot().Evidence {
			if existing.Image != nil && existing.Image.PerceptualHash != "" && preservationUsable(existing) && HashDistance(item.Image.PerceptualHash, existing.Image.PerceptualHash) <= 6 {
				item.NearDuplicateOf = existing.ID
				item.Warnings = append(item.Warnings, "This image appears visually similar to "+existing.SafeName+". Review both before excluding either one.")
				break
			}
		}
	}
	if err = v.commitPreservation(record, item); err != nil {
		return EvidenceItem{}, false, err
	}
	return item, false, nil
}
