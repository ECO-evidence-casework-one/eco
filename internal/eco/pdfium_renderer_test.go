package eco

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPDFiumRenderArgsAreBoundedAndPageSpecific(t *testing.T) {
	got := pdfiumRenderArgs(`C:\source.pdf`, `C:\page.png`, 7, 1600)
	want := []string{"render", "--pages", "7", "--file-type", "png", "--max-width", "1600", "--max-height", "3000", `C:\source.pdf`, `C:\page.png`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected pdfium-cli render args: %#v", got)
	}
}

func TestPDFiumVersionRejectsUnqualifiedExecutableBeforeRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pdfium.exe")
	if err := os.WriteFile(path, []byte("not the qualified renderer"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := pdfiumCLIVersion(context.Background(), path)
	if err == nil || !strings.Contains(err.Error(), "does not match ECO's qualified") {
		t.Fatalf("unexpected qualification result: %v", err)
	}
}

func TestValidatePDFPageRenderRequiresSourceAndBounds(t *testing.T) {
	base := PDFPageRender{
		EngineVersion: qualifiedPDFiumCLIVersion,
		SourceObject:  "object-123.eco",
		SourceSHA256:  strings.Repeat("a", 64),
		Page:          2,
		Width:         1200,
		Height:        1553,
		PNG:           []byte{1, 2, 3},
		CreatedAt:     time.Now().UTC(),
	}
	if err := validatePDFPageRender(base); err != nil {
		t.Fatal(err)
	}
	bad := base
	bad.Page = 0
	if err := validatePDFPageRender(bad); err == nil {
		t.Fatal("invalid page number was accepted")
	}
	bad = base
	bad.Width = maxPDFiumRenderWidth + 1
	if err := validatePDFPageRender(bad); err == nil {
		t.Fatal("unbounded render width was accepted")
	}
}
