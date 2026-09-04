package eco

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeQualificationPDF(path string) error {
	stream1 := []byte("BT /F1 24 Tf 72 720 Td (ECO PDFIUM PAGE ONE) Tj ET\n")
	stream2 := []byte("BT /F1 24 Tf 36 540 Td (ECO PDFIUM PAGE TWO) Tj ET\n")
	objects := [][]byte{
		[]byte("<< /Type /Catalog /Pages 2 0 R >>"),
		[]byte("<< /Type /Pages /Kids [3 0 R 4 0 R] /Count 2 >>"),
		[]byte("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 7 0 R >> >> /Contents 5 0 R >>"),
		[]byte("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 600] /Resources << /Font << /F1 7 0 R >> >> /Contents 6 0 R >>"),
		[]byte(fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(stream1), stream1)),
		[]byte(fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(stream2), stream2)),
		[]byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"),
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	offsets := make([]int, len(objects)+1)
	for i, obj := range objects {
		offsets[i+1] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n", i+1)
		out.Write(obj)
		out.WriteString("\nendobj\n")
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n", len(objects)+1)
	out.WriteString("0000000000 65535 f \n")
	for _, off := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return os.WriteFile(path, out.Bytes(), 0600)
}

func TestActualQualifiedPDFiumRuntimeThroughVault(t *testing.T) {
	if os.PathSeparator != '\\' {
		t.Skip("actual qualified pdfium-cli runtime test is Windows-only")
	}
	exe := strings.TrimSpace(os.Getenv("ECO_PDFIUM_CLI"))
	if exe == "" {
		t.Skip("ECO_PDFIUM_CLI is not set")
	}
	exe, err := filepath.Abs(exe)
	if err != nil {
		t.Fatal(err)
	}

	root := filepath.Join(t.TempDir(), "vault")
	v, err := OpenVault(root)
	if err != nil {
		t.Fatalf("open qualification vault: %v", err)
	}

	pdf := filepath.Join(t.TempDir(), "two-pages.pdf")
	if err := writeQualificationPDF(pdf); err != nil {
		t.Fatal(err)
	}
	item, duplicate, err := v.ImportFile(pdf, nil)
	if err != nil {
		t.Fatalf("import qualification PDF: %v", err)
	}
	if duplicate {
		t.Fatal("first qualification import unexpectedly reported a duplicate")
	}
	if item.DetectedType != "pdf" || !item.SourceVerified || item.ObjectFile == "" || item.SHA256 == "" {
		t.Fatalf("qualification PDF was not committed as verified PDF evidence: %+v", item)
	}

	reg, err := v.RegisterLocalTool("pdfium-cli", exe)
	if err != nil {
		t.Fatalf("register exact qualified renderer: %v", err)
	}
	if reg.Version != qualifiedPDFiumCLIVersion || reg.SHA256 != qualifiedPDFiumCLISHA256 || reg.Size != qualifiedPDFiumCLIBytes {
		t.Fatalf("registered renderer identity diverged from qualified runtime: %+v", reg)
	}

	page1, err := v.RenderEvidencePDFPageWithRegisteredPDFium(item.ID, 1, 1600)
	if err != nil {
		t.Fatalf("render page 1 through ECO vault: %v", err)
	}
	page2, err := v.RenderEvidencePDFPageWithRegisteredPDFium(item.ID, 2, 1600)
	if err != nil {
		t.Fatalf("render page 2 through ECO vault: %v", err)
	}
	for _, result := range []PDFPageRender{page1, page2} {
		if result.SourceObject != item.ObjectFile || result.SourceSHA256 != item.SHA256 {
			t.Fatalf("render is not bound to preserved source: %+v", result)
		}
		if len(result.PNG) < 8 || !bytes.Equal(result.PNG[:8], []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}) {
			t.Fatal("rendered output is not PNG")
		}
	}
	if page1.Page != 1 || page2.Page != 2 {
		t.Fatalf("page request identity was not preserved: page1=%d page2=%d", page1.Page, page2.Page)
	}
	if page1.Width == page2.Width && page1.Height == page2.Height {
		t.Fatalf("different-size PDF pages rendered with identical dimensions: %dx%d and %dx%d", page1.Width, page1.Height, page2.Width, page2.Height)
	}

	f, err := os.OpenFile(exe, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open renderer for deliberate tamper: %v", err)
	}
	if _, err := f.Write([]byte{0}); err != nil {
		f.Close()
		t.Fatalf("tamper renderer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := v.RenderEvidencePDFPageWithRegisteredPDFium(item.ID, 1, 1600); err == nil {
		t.Fatal("tampered renderer was accepted after registration")
	}
}
