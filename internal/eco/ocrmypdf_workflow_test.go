package eco

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fakeOCRmyPDFResult(t *testing.T, source SourceReceipt, sidecar string) OCRmyPDFResult {
	t.Helper()
	text, segments, pageCount, ocrPages, warnings, err := parseOCRmyPDFSidecar([]byte(sidecar), source)
	if err != nil {
		t.Fatal(err)
	}
	return OCRmyPDFResult{
		Text:          text,
		Segments:      segments,
		EngineVersion: "OCRmyPDF 17.11.0 test",
		Source:        source,
		PageCount:     pageCount,
		OCRPages:      ocrPages,
		Resources: ResourceAssessment{
			Level: "normal",
			Snapshot: ResourceSnapshot{
				MemorySampled:        true,
				MemoryAvailableBytes: 2 * resourceGiB,
				DiskSampled:          true,
				DiskFreeBytes:        10 * resourceGiB,
			},
		},
		Warnings: warnings,
	}
}

func TestOCRmyPDFWorkflowUsesVerifiedReadingCopyAndPreservesExistingReading(t *testing.T) {
	v, item := testPDFCPUVault(t)
	v.mu.Lock()
	for i := range v.Workspace.Evidence {
		if v.Workspace.Evidence[i].ID != item.ID {
			continue
		}
		v.Workspace.Evidence[i].ExtractedText = "Existing Docling/native reading must survive."
		v.Workspace.Evidence[i].Readable = true
		v.Workspace.Evidence[i].Status = "Ready — existing reading"
		v.Workspace.Evidence[i].Segments = append(v.Workspace.Evidence[i].Segments, SourceSegment{
			ID: "SEG-EXISTING-1", Ordinal: 1, Text: "Existing source wording", Origin: "docling",
			SourceObject: item.ObjectFile, SourceSHA256: item.SHA256,
		})
		break
	}
	if err := v.saveUnlocked(); err != nil {
		v.mu.Unlock()
		t.Fatal(err)
	}
	v.mu.Unlock()

	seenPath := ""
	seenSource := SourceReceipt{}
	fake := func(_ context.Context, _ string, path, language string, source SourceReceipt) (OCRmyPDFResult, error) {
		seenPath = path
		seenSource = source
		if language != "eng" {
			t.Fatalf("unexpected OCR language %q", language)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return OCRmyPDFResult{}, err
		}
		if !strings.HasPrefix(string(data), "%PDF") {
			t.Fatalf("verified reading copy does not contain PDF bytes: %q", data)
		}
		return fakeOCRmyPDFResult(t, source, "Scanned page one text.\f\fScanned page three text."), nil
	}

	result, err := v.extractEvidenceWithOCRmyPDFRunner(context.Background(), item.ID, "ignored", "eng", fake)
	if err != nil {
		t.Fatal(err)
	}
	if seenPath == "" || seenPath == item.SourcePath || !strings.Contains(filepath.Base(seenPath), "verified-reading-") {
		t.Fatalf("OCRmyPDF runner did not receive ECO's verified reading copy: %q", seenPath)
	}
	if seenSource.ObjectFile != item.ObjectFile || seenSource.SHA256 != item.SHA256 {
		t.Fatalf("OCRmyPDF runner did not receive preserved source receipt: %+v", seenSource)
	}
	if result.PageCount != 3 || result.OCRPages != 2 || len(result.Segments) != 2 {
		t.Fatalf("unexpected OCRmyPDF result: %+v", result)
	}

	ws := v.Snapshot()
	var got EvidenceItem
	for _, candidate := range ws.Evidence {
		if candidate.ID == item.ID {
			got = candidate
			break
		}
	}
	if got.ExtractedText != "Existing Docling/native reading must survive." || got.Status != "Ready — existing reading" {
		t.Fatalf("OCRmyPDF overwrote stronger existing reading/status: %+v", got)
	}
	if len(got.Segments) != 3 {
		t.Fatalf("expected existing + two OCRmyPDF segments, got %+v", got.Segments)
	}
	if got.Segments[0].Origin != "docling" || got.Segments[1].Origin != "ocrmypdf" || got.Segments[2].Page != 3 {
		t.Fatalf("supplemental segment ordering/binding is wrong: %+v", got.Segments)
	}
	foundAudit := false
	for _, change := range ws.Changes {
		if change.Type != "ocrmypdf-sidecar-added" {
			continue
		}
		foundAudit = true
		if change.Details["source_sha256"] != item.SHA256 || change.Details["page_count"] != 3 || change.Details["ocr_pages"] != 2 {
			t.Fatalf("OCRmyPDF audit lost source/page binding: %+v", change.Details)
		}
		if retained, _ := change.Details["output_pdf_retained"].(bool); retained {
			t.Fatalf("OCRmyPDF audit incorrectly claims a derived PDF was retained: %+v", change.Details)
		}
		if replaces, _ := change.Details["replaces_existing_reading"].(bool); replaces {
			t.Fatalf("OCRmyPDF audit incorrectly claims replacement semantics: %+v", change.Details)
		}
		break
	}
	if !foundAudit {
		t.Fatal("missing authenticated OCRmyPDF sidecar audit event")
	}
}

func TestOCRmyPDFWorkflowRerunReplacesOnlyPriorOCRmyPDFSegments(t *testing.T) {
	v, item := testPDFCPUVault(t)
	calls := 0
	fake := func(_ context.Context, _ string, _ string, _ string, source SourceReceipt) (OCRmyPDFResult, error) {
		calls++
		if calls == 1 {
			return fakeOCRmyPDFResult(t, source, "First OCR pass text."), nil
		}
		return fakeOCRmyPDFResult(t, source, "Second OCR pass replacement."), nil
	}
	if _, err := v.extractEvidenceWithOCRmyPDFRunner(context.Background(), item.ID, "ignored", "eng", fake); err != nil {
		t.Fatal(err)
	}
	if _, err := v.extractEvidenceWithOCRmyPDFRunner(context.Background(), item.ID, "ignored", "eng", fake); err != nil {
		t.Fatal(err)
	}
	ws := v.Snapshot()
	for _, candidate := range ws.Evidence {
		if candidate.ID != item.ID {
			continue
		}
		ocrSegments := []SourceSegment{}
		for _, segment := range candidate.Segments {
			if segment.Origin == "ocrmypdf" {
				ocrSegments = append(ocrSegments, segment)
			}
		}
		if len(ocrSegments) != 1 || !strings.Contains(ocrSegments[0].Text, "Second OCR pass") {
			t.Fatalf("OCRmyPDF rerun accumulated stale sidecar segments: %+v", candidate.Segments)
		}
		return
	}
	t.Fatal("fixture evidence disappeared")
}

func TestOCRmyPDFWorkflowRejectsNonPDFBeforeRunner(t *testing.T) {
	v, _ := testGroundingVault(t)
	item := v.Snapshot().Evidence[0]
	called := false
	fake := func(_ context.Context, _, _, _ string, _ SourceReceipt) (OCRmyPDFResult, error) {
		called = true
		return OCRmyPDFResult{}, nil
	}
	_, err := v.extractEvidenceWithOCRmyPDFRunner(context.Background(), item.ID, "ignored", "eng", fake)
	if err == nil || !strings.Contains(err.Error(), "requires detected PDF evidence") {
		t.Fatalf("expected non-PDF rejection, got %v", err)
	}
	if called {
		t.Fatal("OCRmyPDF runner was called for non-PDF evidence")
	}
}

func TestOCRmyPDFWorkflowRejectsForgedSourceBinding(t *testing.T) {
	v, item := testPDFCPUVault(t)
	fake := func(_ context.Context, _ string, _ string, _ string, source SourceReceipt) (OCRmyPDFResult, error) {
		result := fakeOCRmyPDFResult(t, source, "OCR text")
		result.Source.SHA256 = strings.Repeat("0", 64)
		return result, nil
	}
	_, err := v.extractEvidenceWithOCRmyPDFRunner(context.Background(), item.ID, "ignored", "eng", fake)
	if err == nil || !strings.Contains(err.Error(), "not bound") {
		t.Fatalf("expected forged OCRmyPDF result rejection, got %v", err)
	}
}

func TestValidateOCRmyPDFResultRejectsInventedGeometryAndConfidence(t *testing.T) {
	now := time.Now().UTC()
	item := EvidenceItem{
		ID: "EVD-1", ObjectFile: "EVD-1.ecoobj", SHA256: strings.Repeat("a", 64),
		DetectedType: "pdf", SourceVerified: true, SourceVerifiedAt: now, Preservation: preservationCommitted,
	}
	source := SourceReceipt{EvidenceID: item.ID, ObjectFile: item.ObjectFile, SHA256: item.SHA256, Size: 1, VerifiedAt: now}
	result := fakeOCRmyPDFResult(t, source, "OCR text")
	region := NormalizedRegion{X: 0, Y: 0, Width: 1, Height: 1}
	result.Segments[0].Region = &region
	result.Segments[0].Confidence = 0.99
	if err := validateOCRmyPDFResult(item, result); err == nil || !strings.Contains(err.Error(), "invent") {
		t.Fatalf("expected invented geometry/confidence rejection, got %v", err)
	}
}
