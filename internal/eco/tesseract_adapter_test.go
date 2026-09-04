package eco

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestTesseractOCRArgs(t *testing.T) {
	got := tesseractOCRArgs(`C:\evidence\page.png`, `C:\temp\ocr`, `eng+deu`)
	want := []string{`C:\evidence\page.png`, `C:\temp\ocr`, "-l", "eng+deu", "tsv"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestTesseractLanguagePattern(t *testing.T) {
	valid := []string{"eng", "eng+deu", "script_Latin", "osd"}
	for _, v := range valid {
		if !tesseractLanguagePattern.MatchString(v) {
			t.Fatalf("expected valid language %q", v)
		}
	}
	invalid := []string{"", "eng deu", "eng;whoami", "../eng"}
	for _, v := range invalid {
		if tesseractLanguagePattern.MatchString(v) {
			t.Fatalf("expected invalid language %q", v)
		}
	}
}

func TestRequireAbsoluteRegularFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tool.bin")
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := requireAbsoluteRegularFile(path, "test tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("resolved path is not absolute: %q", got)
	}

	if _, err := requireAbsoluteRegularFile("relative-tool", "test tool"); err == nil {
		t.Fatal("expected relative path to be rejected")
	}
}
