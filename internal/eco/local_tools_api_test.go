package eco

import (
	"context"
	"path/filepath"
	"testing"
)

func TestRegisteredLocalToolsListsCanonicalActiveRegistrations(t *testing.T) {
	d := t.TempDir()
	v, err := openTestVault(filepath.Join(d, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	tesseractPath := writeLocalToolFixture(t, d, "tesseract.bin", "tesseract registry fixture")
	pdfcpuPath := writeLocalToolFixture(t, d, "pdfcpu.bin", "pdfcpu registry fixture")
	if _, err := v.registerLocalToolWithProbe(context.Background(), "tesseract-ocr", tesseractPath, testLocalToolProbe("5.5-test")); err != nil {
		t.Fatal(err)
	}
	if _, err := v.registerLocalToolWithProbe(context.Background(), "pdfcpu", pdfcpuPath, testLocalToolProbe("pdfcpu-test")); err != nil {
		t.Fatal(err)
	}
	registered, err := v.RegisteredLocalTools()
	if err != nil {
		t.Fatal(err)
	}
	if len(registered) != 2 {
		t.Fatalf("expected two active registrations, got %+v", registered)
	}
	if registered[0].Kind != "tesseract" || registered[1].Kind != "pdfcpu" {
		t.Fatalf("registered tool list did not preserve canonical order: %+v", registered)
	}
}
