package eco

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const maxHostileFuzzInput = 1 << 20

func makeFuzzZIPSeed(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	w, err := zw.Create("evidence/readme.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("ECO fuzz seed")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func FuzzZIPPreflightAndInspection(f *testing.F) {
	f.Add([]byte("not a zip"))
	f.Add([]byte{'P', 'K', 0x05, 0x06, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxHostileFuzzInput {
			t.Skip()
		}
		path := filepath.Join(t.TempDir(), "hostile.zip")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		if err := zipPreflight(path, 10000); err != nil {
			return
		}
		text, err := inspectZIP(path)
		if err != nil {
			return
		}
		if len(text) > 8*maxHostileFuzzInput {
			t.Fatalf("ZIP inspection produced unexpectedly large output: %d bytes", len(text))
		}
	})
}

func TestZIPFuzzValidSeed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "seed.zip")
	if err := os.WriteFile(path, makeFuzzZIPSeed(t), 0600); err != nil {
		t.Fatal(err)
	}
	if err := zipPreflight(path, 10000); err != nil {
		t.Fatalf("valid ZIP seed failed preflight: %v", err)
	}
	text, err := inspectZIP(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "evidence/readme.txt") {
		t.Fatalf("valid ZIP seed was not inspected: %q", text)
	}
}

func FuzzEMLReader(f *testing.F) {
	f.Add([]byte("From: alice@example.test\r\nTo: bob@example.test\r\nSubject: ECO seed\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nHello Bob.\r\n"))
	f.Add([]byte("Content-Type: multipart/mixed; boundary=broken\r\n\r\n--broken"))
	f.Add([]byte("not mail"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxHostileFuzzInput {
			t.Skip()
		}
		path := filepath.Join(t.TempDir(), "hostile.eml")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		text, _ := extractEML(path)
		if len(text) > int(maxExtractBytes)+2*maxHostileFuzzInput {
			t.Fatalf("EML reader produced unexpectedly large output: %d bytes", len(text))
		}
	})
}

func FuzzXMLReadableText(f *testing.F) {
	f.Add([]byte("<?xml version=\"1.0\"?><document><p>ECO XML seed</p></document>"))
	f.Add([]byte("<p><p><p>unterminated"))
	f.Add([]byte{0, '<', 'p', '>'})
	paragraphs := map[string]bool{"p": true, "h": true, "tr": true, "table-row": true}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxHostileFuzzInput {
			t.Skip()
		}
		text := xmlText(data, paragraphs)
		if len(text) > 4*len(data)+4096 {
			t.Fatalf("XML text extraction expanded unexpectedly: input=%d output=%d", len(data), len(text))
		}
	})
}

func FuzzFileTypeSniffer(f *testing.F) {
	f.Add([]byte("%PDF-1.7\n"))
	f.Add([]byte{'P', 'K', 0x03, 0x04})
	f.Add([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a})
	f.Add([]byte("     "))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxHostileFuzzInput {
			t.Skip()
		}
		detected := detectBytes(data)
		if len(detected) > 64 {
			t.Fatalf("file-type detector returned unbounded type string: %q", detected)
		}
	})
}
