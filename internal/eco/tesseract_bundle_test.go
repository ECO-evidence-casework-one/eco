package eco

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestSafeTesseractRuntimeRelativePath(t *testing.T) {
	good, err := safeTesseractRuntimeRelativePath("share/tessdata/eng.traineddata")
	if err != nil || good != "share/tessdata/eng.traineddata" {
		t.Fatalf("safe path rejected: %q %v", good, err)
	}
	for _, bad := range []string{"", "../evil.dll", "bin/../evil.dll", "/absolute.dll", `bin\\..\\evil.dll`} {
		if _, err := safeTesseractRuntimeRelativePath(bad); err == nil {
			t.Fatalf("unsafe path accepted: %q", bad)
		}
	}
}

func TestTesseractRuntimeRegistrationRequiresWave4Receipts(t *testing.T) {
	root := filepath.Join(t.TempDir(), "bundle")
	base := TesseractRuntimeRegistration{
		Root:                   root,
		Executable:             filepath.Join(root, "runtime", "bin", "tesseract.exe"),
		TessdataDir:            filepath.Join(root, "runtime", "share", "tessdata"),
		Version:                "5.5.0",
		BuildManifestSHA256:    wave4BuildManifestSHA256,
		RuntimeInventorySHA256: wave4RuntimeInventorySHA256,
		OCRSmokeSHA256:         wave4OCRSmokeSHA256,
		RegisteredAt:           time.Now().UTC(),
	}
	if err := validateTesseractRuntimeRegistration(base); err != nil {
		t.Fatal(err)
	}
	bad := base
	bad.RuntimeInventorySHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if err := validateTesseractRuntimeRegistration(bad); err == nil {
		t.Fatal("wrong Wave 4 inventory receipt was accepted")
	}
}

func TestTesseractOCRArgsWithTessdata(t *testing.T) {
	got := tesseractOCRArgsWithTessdata(`C:\\in.png`, `C:\\out`, "eng", `C:\\bundle\\runtime\\share\\tessdata`)
	want := []string{`C:\\in.png`, `C:\\out`, "--tessdata-dir", `C:\\bundle\\runtime\\share\\tessdata`, "-l", "eng", "tsv"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected Tesseract args: %#v", got)
	}
	if fallback := tesseractOCRArgs(`C:\\in.png`, `C:\\out`, "eng"); reflect.DeepEqual(fallback, want) {
		t.Fatal("legacy path unexpectedly injected a tessdata directory")
	}
}
