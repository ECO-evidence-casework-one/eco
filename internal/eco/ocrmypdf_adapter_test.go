package eco

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func testOCRmyPDFSourceReceipt() SourceReceipt {
	return SourceReceipt{
		EvidenceID: "EVD-1",
		ObjectFile: "EVD-1.ecoobj",
		SHA256:     strings.Repeat("a", 64),
		Size:       123,
		VerifiedAt: time.Now().UTC(),
	}
}

func TestOCRmyPDFArgsAreSidecarOnlyAndBounded(t *testing.T) {
	args := ocrmyPDFArgs(`C:\work\verified.pdf`, `C:\work\must-not-exist.pdf`, `C:\work\sidecar.txt`, "eng")
	joined := strings.Join(args, " ")
	for _, required := range []string{
		"--output-type none",
		"--sidecar C:\\work\\sidecar.txt",
		"--mode skip",
		"--ocr-engine tesseract",
		"--rasterizer pypdfium",
		"--jobs 1",
		"--no-progress-bar",
		"--tesseract-timeout 30",
		"--max-ocr-image-mpixels 25",
		"--language eng",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing controlled OCRmyPDF argument %q in %q", required, joined)
		}
	}
	for _, forbidden := range []string{
		"--force-ocr", "--redo-ocr", "--deskew", "--clean", "--clean-final",
		"--rotate-pages", "--remove-background", "--plugin", "--keep-temporary-files",
		"--output-type pdf ", "--output-type pdfa",
	} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("sidecar-only adapter unexpectedly contains %q: %q", forbidden, joined)
		}
	}
}

func TestOfflineOCRmyPDFEnvironmentRemovesProxyRoutes(t *testing.T) {
	env := offlineOCRmyPDFEnvironment([]string{
		"PATH=X",
		"HTTP_PROXY=http://proxy.invalid:8080",
		"https_proxy=http://proxy.invalid:8080",
		"ALL_PROXY=socks5://proxy.invalid:1080",
	})
	joined := strings.Join(env, "\n")
	upper := strings.ToUpper(joined)
	for _, forbidden := range []string{"HTTP_PROXY=", "HTTPS_PROXY=", "ALL_PROXY=", "FTP_PROXY="} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("proxy route survived OCRmyPDF environment isolation: %q", joined)
		}
	}
	if !strings.Contains(joined, "PATH=X") || !strings.Contains(joined, "NO_PROXY=*") {
		t.Fatalf("expected OCRmyPDF environment controls are missing: %q", joined)
	}
}

func TestParseOCRmyPDFSidecarPreservesPagePositionsAndUniqueIDs(t *testing.T) {
	source := testOCRmyPDFSourceReceipt()
	data := []byte("First page OCR line.\nSecond sentence.\f\fThird page position has OCR text.")
	text, segments, pageCount, ocrPages, warnings, err := parseOCRmyPDFSidecar(data, source)
	if err != nil {
		t.Fatal(err)
	}
	if pageCount != 3 || ocrPages != 2 {
		t.Fatalf("wrong sidecar page accounting: pages=%d ocrPages=%d", pageCount, ocrPages)
	}
	if !strings.Contains(text, "\f\f") {
		t.Fatalf("sidecar page positions were not preserved: %q", text)
	}
	if len(segments) != 2 {
		t.Fatalf("expected one compact segment on each OCR-bearing page, got %+v", segments)
	}
	if segments[0].Page != 1 || segments[1].Page != 3 {
		t.Fatalf("OCRmyPDF page positions were collapsed: %+v", segments)
	}
	seen := map[string]bool{}
	for i, segment := range segments {
		if segment.ID != "SEG-OCRPDF-000"+string(rune('1'+i)) {
			// Avoid depending on fmt in this regression test; exact first two IDs are checked below.
		}
		if seen[segment.ID] {
			t.Fatalf("duplicate OCRmyPDF segment ID %q", segment.ID)
		}
		seen[segment.ID] = true
		if segment.Origin != "ocrmypdf" || segment.SourceObject != source.ObjectFile || segment.SourceSHA256 != source.SHA256 {
			t.Fatalf("segment lost source binding: %+v", segment)
		}
		if segment.Region != nil || segment.Confidence != 0 {
			t.Fatalf("sidecar invented unavailable OCR geometry/confidence: %+v", segment)
		}
	}
	if segments[0].ID != "SEG-OCRPDF-0001" || segments[1].ID != "SEG-OCRPDF-0002" {
		t.Fatalf("unexpected OCRmyPDF segment IDs: %+v", segments)
	}
	if len(warnings) < 2 {
		t.Fatalf("expected OCR uncertainty/completeness warnings, got %v", warnings)
	}
}

func TestParseOCRmyPDFSidecarRepairsInvalidUTF8WithoutInventingConfidence(t *testing.T) {
	source := testOCRmyPDFSourceReceipt()
	_, segments, pages, ocrPages, _, err := parseOCRmyPDFSidecar([]byte{'A', 0xff, 'B'}, source)
	if err != nil {
		t.Fatal(err)
	}
	if pages != 1 || ocrPages != 1 || len(segments) != 1 {
		t.Fatalf("unexpected repaired sidecar result: pages=%d ocrPages=%d segments=%+v", pages, ocrPages, segments)
	}
	if !strings.Contains(segments[0].Text, "�") || segments[0].Confidence != 0 {
		t.Fatalf("invalid UTF-8 handling or confidence semantics are wrong: %+v", segments[0])
	}
}

func TestOCRmyPDFLimitedCaptureFlagsOverflow(t *testing.T) {
	var buf bytes.Buffer
	capture := &ocrmyPDFLimitedCapture{buf: &buf, max: 5}
	if n, err := capture.Write([]byte("123456789")); err != nil || n != 9 {
		t.Fatalf("unexpected capture write: n=%d err=%v", n, err)
	}
	if !capture.overflow || buf.String() != "12345" {
		t.Fatalf("bounded OCRmyPDF capture failed: overflow=%v data=%q", capture.overflow, buf.String())
	}
}
