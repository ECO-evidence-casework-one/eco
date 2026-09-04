package eco

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractReadablePDFNativeText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "native.pdf")
	if err := os.WriteFile(path, makeNativeTextPDF("ECO PDF TEXT 123"), 0600); err != nil {
		t.Fatal(err)
	}
	text, segments, warnings := ExtractReadable(path, "pdf")
	if !strings.Contains(text, "ECO PDF TEXT 123") {
		t.Fatalf("missing known PDF text: %q", text)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(segments) == 0 {
		t.Fatal("expected page-aware PDF segments")
	}
	if segments[0].Page != 1 || segments[0].PageHint != "Page 1" || segments[0].Origin != "pdf-native" {
		t.Fatalf("unexpected PDF segment provenance: %+v", segments[0])
	}
}

func TestExtractReadableMalformedPDFFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog"), 0600); err != nil {
		t.Fatal(err)
	}
	_, _, warnings := ExtractReadable(path, "pdf")
	if len(warnings) == 0 {
		t.Fatal("malformed PDF should produce a bounded warning")
	}
}

func makeNativeTextPDF(text string) []byte {
	text = strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)").Replace(text)
	stream := []byte("BT /F1 24 Tf 72 700 Td (" + text + ") Tj ET\n")
	objects := [][]byte{
		[]byte("<< /Type /Catalog /Pages 2 0 R >>"),
		[]byte("<< /Type /Pages /Kids [3 0 R] /Count 1 >>"),
		[]byte("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>"),
		[]byte(fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(stream), stream)),
		[]byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"),
	}
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, object := range objects {
		offsets[i+1] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n", i+1)
		b.Write(object)
		b.WriteString("\nendobj\n")
	}
	xref := b.Len()
	b.WriteString("xref\n0 6\n0000000000 65535 f \n")
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xref)
	return b.Bytes()
}
