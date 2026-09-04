package eco

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testPDFCPUVault(t *testing.T) (*Vault, EvidenceItem) {
	t.Helper()
	d := t.TempDir()
	src := filepath.Join(d, "fixture.pdf")
	// The process runner is faked in workflow tests; this fixture only needs a
	// genuine PDF signature so ECO's own content detector classifies it as PDF.
	if err := os.WriteFile(src, []byte("%PDF-1.7\n% ECO workflow fixture\n%%EOF\n"), 0600); err != nil {
		t.Fatal(err)
	}
	v, err := OpenVault(filepath.Join(d, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	item, duplicate, err := v.ImportFile(src, nil)
	if err != nil {
		t.Fatal(err)
	}
	if duplicate || item.DetectedType != "pdf" || !preservationUsable(item) {
		t.Fatalf("unexpected PDF fixture import: item=%+v duplicate=%v", item, duplicate)
	}
	return v, item
}

func TestPDFCPUWorkflowUsesVerifiedReadingCopyAndAuditsAssessment(t *testing.T) {
	v, item := testPDFCPUVault(t)
	seenPath := ""
	seenSource := SourceReceipt{}
	fake := func(_ context.Context, _ string, path string, source SourceReceipt) (PDFAssessment, error) {
		seenPath = path
		seenSource = source
		data, err := os.ReadFile(path)
		if err != nil {
			return PDFAssessment{}, err
		}
		if !strings.HasPrefix(string(data), "%PDF") {
			t.Fatalf("verified reading copy does not contain the PDF source: %q", data)
		}
		return PDFAssessment{
			EngineVersion:           "pdfcpu test-build",
			SourceObject:            source.ObjectFile,
			SourceSHA256:            source.SHA256,
			CreatedAt:               time.Now().UTC(),
			RelaxedValidationPassed: true,
			StrictValidationPassed:  true,
			Version:                 "1.7",
			PageCount:               3,
			Signatures:              true,
			AttachmentCount:         2,
		}, nil
	}
	assessment, err := v.inspectEvidencePDFWithRunner(context.Background(), item.ID, "ignored", fake)
	if err != nil {
		t.Fatal(err)
	}
	if seenPath == "" || seenPath == item.SourcePath || !strings.Contains(filepath.Base(seenPath), "verified-reading-") {
		t.Fatalf("pdfcpu runner did not receive ECO's verified temporary reading copy: %q", seenPath)
	}
	if seenSource.ObjectFile != item.ObjectFile || seenSource.SHA256 != item.SHA256 {
		t.Fatalf("pdfcpu runner did not receive the preserved source receipt: %+v", seenSource)
	}
	if assessment.PageCount != 3 || assessment.Version != "1.7" || !assessment.Signatures || assessment.AttachmentCount != 2 {
		t.Fatalf("unexpected assessment: %+v", assessment)
	}
	ws := v.Snapshot()
	found := false
	for _, change := range ws.Changes {
		if change.Type != "pdf-structure-inspected" {
			continue
		}
		found = true
		if change.Details["source_sha256"] != item.SHA256 || change.Details["page_count"] != 3 {
			t.Fatalf("PDF audit event lost source/summary binding: %+v", change)
		}
		if truth, ok := change.Details["inspection_is_content_truth"].(bool); !ok || truth {
			t.Fatalf("PDF audit event overstated structural inspection: %+v", change)
		}
		break
	}
	if !found {
		t.Fatal("missing authenticated PDF inspection audit event")
	}
}

func TestPDFCPUWorkflowRecordsNonPassingRelaxedValidationWithoutCallingItCorrupt(t *testing.T) {
	v, item := testPDFCPUVault(t)
	fake := func(_ context.Context, _ string, _ string, source SourceReceipt) (PDFAssessment, error) {
		return PDFAssessment{
			EngineVersion:            "pdfcpu test-build",
			SourceObject:             source.ObjectFile,
			SourceSHA256:             source.SHA256,
			CreatedAt:                time.Now().UTC(),
			RelaxedValidationPassed:  false,
			RelaxedValidationError:   "validation failed",
			StrictValidationPassed:   false,
			Warnings: []string{"The original remains preserved; this may also represent unsupported/password-protected input."},
		}, nil
	}
	assessment, err := v.inspectEvidencePDFWithRunner(context.Background(), item.ID, "ignored", fake)
	if err != nil {
		t.Fatal(err)
	}
	if assessment.RelaxedValidationPassed || !strings.Contains(assessment.RelaxedValidationError, "validation failed") {
		t.Fatalf("non-passing validation was not preserved accurately: %+v", assessment)
	}
	for _, warning := range assessment.Warnings {
		if strings.Contains(strings.ToLower(warning), "corrupt") {
			t.Fatalf("assessment improperly converted a validation diagnostic into a corruption claim: %q", warning)
		}
	}
}

func TestPDFCPUWorkflowRejectsNonPDFEvidenceBeforeRunner(t *testing.T) {
	v, _ := testGroundingVault(t)
	item := v.Snapshot().Evidence[0]
	called := false
	fake := func(_ context.Context, _, _ string, _ SourceReceipt) (PDFAssessment, error) {
		called = true
		return PDFAssessment{}, nil
	}
	_, err := v.inspectEvidencePDFWithRunner(context.Background(), item.ID, "ignored", fake)
	if err == nil || !strings.Contains(err.Error(), "requires detected PDF evidence") {
		t.Fatalf("expected non-PDF rejection, got %v", err)
	}
	if called {
		t.Fatal("pdfcpu runner was called for non-PDF evidence")
	}
}

func TestPDFCPUWorkflowRejectsAssessmentWithWrongSourceBinding(t *testing.T) {
	v, item := testPDFCPUVault(t)
	fake := func(_ context.Context, _ string, _ string, source SourceReceipt) (PDFAssessment, error) {
		return PDFAssessment{
			EngineVersion:            "pdfcpu test-build",
			SourceObject:             source.ObjectFile,
			SourceSHA256:             strings.Repeat("0", 64),
			CreatedAt:                time.Now().UTC(),
			RelaxedValidationPassed:  true,
			StrictValidationPassed:   true,
			Version:                  "1.7",
			PageCount:                1,
		}, nil
	}
	_, err := v.inspectEvidencePDFWithRunner(context.Background(), item.ID, "ignored", fake)
	if err == nil || !strings.Contains(err.Error(), "not bound") {
		t.Fatalf("expected forged source binding rejection, got %v", err)
	}
}

func TestPDFCPUWorkflowRejectsPassingAssessmentWithoutPageIdentity(t *testing.T) {
	v, item := testPDFCPUVault(t)
	fake := func(_ context.Context, _ string, _ string, source SourceReceipt) (PDFAssessment, error) {
		return PDFAssessment{
			EngineVersion:            "pdfcpu test-build",
			SourceObject:             source.ObjectFile,
			SourceSHA256:             source.SHA256,
			CreatedAt:                time.Now().UTC(),
			RelaxedValidationPassed:  true,
			StrictValidationPassed:   true,
		}, nil
	}
	_, err := v.inspectEvidencePDFWithRunner(context.Background(), item.ID, "ignored", fake)
	if err == nil || !strings.Contains(err.Error(), "lacks PDF version/page count") {
		t.Fatalf("expected incomplete passing assessment rejection, got %v", err)
	}
}
