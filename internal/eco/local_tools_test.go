package eco

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testLocalToolProbe(version string) localToolVersionProbe {
	return func(_ context.Context, kind, executable string) (string, error) {
		if strings.TrimSpace(kind) == "" || !filepath.IsAbs(executable) {
			return "", errors.New("bad test probe input")
		}
		return version, nil
	}
}

func writeLocalToolFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0700); err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func TestLocalToolRegistrationPersistsEncryptedAndVerifies(t *testing.T) {
	d := t.TempDir()
	vaultRoot := filepath.Join(d, "vault")
	v, err := OpenVault(vaultRoot)
	if err != nil {
		t.Fatal(err)
	}
	toolPath := writeLocalToolFixture(t, d, "tesseract-test.bin", "stable tesseract fixture bytes")

	registered, err := v.registerLocalToolWithProbe(context.Background(), "tesseract-ocr", toolPath, testLocalToolProbe("5.5.1-test"))
	if err != nil {
		t.Fatal(err)
	}
	if registered.Kind != "tesseract" || registered.Version != "5.5.1-test" || registered.SHA256 == "" || registered.AuditChangeID == "" {
		t.Fatalf("unexpected registration: %+v", registered)
	}
	if registered.Upstream != "tesseract-ocr/tesseract" || registered.License != "Apache-2.0" {
		t.Fatalf("donor identity was not pinned: %+v", registered)
	}

	verified, err := v.verifyRegisteredLocalToolWithProbe(context.Background(), "tesseract", testLocalToolProbe("5.5.1-test"))
	if err != nil {
		t.Fatal(err)
	}
	if verified.SHA256 != registered.SHA256 || verified.Executable != registered.Executable {
		t.Fatalf("verified registration changed identity: %+v vs %+v", verified, registered)
	}

	raw, err := os.ReadFile(filepath.Join(vaultRoot, "workspace.ecodb"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), toolPath) || strings.Contains(string(raw), "5.5.1-test") || strings.Contains(string(raw), registered.SHA256) {
		t.Fatal("encrypted workspace leaks local tool registration plaintext")
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenVault(vaultRoot)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := reopened.RegisteredLocalTool("tesseract")
	if err != nil {
		t.Fatal(err)
	}
	if persisted.SHA256 != registered.SHA256 || persisted.Version != registered.Version || persisted.AuditChangeID != registered.AuditChangeID {
		t.Fatalf("registration did not survive reopen: %+v", persisted)
	}
}

func TestLocalToolVerificationRejectsChangedExecutable(t *testing.T) {
	d := t.TempDir()
	v, err := OpenVault(filepath.Join(d, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	toolPath := writeLocalToolFixture(t, d, "pdfcpu-test.bin", "pdfcpu fixture version one")
	if _, err := v.registerLocalToolWithProbe(context.Background(), "pdfcpu", toolPath, testLocalToolProbe("pdfcpu v1")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(toolPath, []byte("pdfcpu fixture version two with changed bytes"), 0700); err != nil {
		t.Fatal(err)
	}
	_, err = v.verifyRegisteredLocalToolWithProbe(context.Background(), "pdfcpu", testLocalToolProbe("pdfcpu v1"))
	if err == nil || (!strings.Contains(err.Error(), "size changed") && !strings.Contains(err.Error(), "SHA-256 changed")) {
		t.Fatalf("expected changed executable rejection, got %v", err)
	}
}

func TestLocalToolVerificationRejectsChangedVersion(t *testing.T) {
	d := t.TempDir()
	v, err := OpenVault(filepath.Join(d, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	toolPath := writeLocalToolFixture(t, d, "docling-test.bin", "stable docling fixture")
	if _, err := v.registerLocalToolWithProbe(context.Background(), "docling", toolPath, testLocalToolProbe("docling 2.0")); err != nil {
		t.Fatal(err)
	}
	_, err = v.verifyRegisteredLocalToolWithProbe(context.Background(), "docling", testLocalToolProbe("docling 2.1"))
	if err == nil || !strings.Contains(err.Error(), "version changed") {
		t.Fatalf("expected changed version rejection, got %v", err)
	}
}

func TestLocalToolNewRegistrationSupersedesOld(t *testing.T) {
	d := t.TempDir()
	v, err := OpenVault(filepath.Join(d, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	first := writeLocalToolFixture(t, d, "llama-one.bin", "llama build one")
	second := writeLocalToolFixture(t, d, "llama-two.bin", "llama build two")
	old, err := v.registerLocalToolWithProbe(context.Background(), "llamacpp", first, testLocalToolProbe("llama build 1"))
	if err != nil {
		t.Fatal(err)
	}
	newer, err := v.registerLocalToolWithProbe(context.Background(), "llama-cli", second, testLocalToolProbe("llama build 2"))
	if err != nil {
		t.Fatal(err)
	}
	if newer.AuditChangeID == old.AuditChangeID || newer.SHA256 == old.SHA256 {
		t.Fatalf("replacement registration did not get a new identity: old=%+v new=%+v", old, newer)
	}
	active, err := v.RegisteredLocalTool("llama.cpp")
	if err != nil {
		t.Fatal(err)
	}
	if active.AuditChangeID != newer.AuditChangeID || active.Executable != newer.Executable || active.Version != "llama build 2" {
		t.Fatalf("newest registration did not supersede old one: %+v", active)
	}
}

func TestLocalToolUnknownKindAndMissingRegistrationFailClosed(t *testing.T) {
	v, err := OpenVault(filepath.Join(t.TempDir(), "vault"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.registerLocalToolWithProbe(context.Background(), "made-up-engine", filepath.Join(t.TempDir(), "x"), testLocalToolProbe("x")); err == nil {
		t.Fatal("unknown local tool kind was accepted")
	}
	if _, err := v.RegisteredLocalTool("tesseract"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing registration should return os.ErrNotExist, got %v", err)
	}
	if err := v.OCRImageWithRegisteredTesseract("EVD-does-not-matter", "eng"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("registered wrapper should fail before evidence processing when tool is absent, got %v", err)
	}
}

func TestLocalToolMalformedNewestRegistrationFailsClosed(t *testing.T) {
	d := t.TempDir()
	v, err := OpenVault(filepath.Join(d, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	toolPath := writeLocalToolFixture(t, d, "ocrmypdf-test.bin", "ocrmypdf stable fixture")
	if _, err := v.registerLocalToolWithProbe(context.Background(), "ocrmypdf", toolPath, testLocalToolProbe("17.11.0")); err != nil {
		t.Fatal(err)
	}

	v.mu.Lock()
	v.addChangeUnlocked("test", localToolRegistrationChangeType, "malformed newer registration", map[string]any{
		"kind":          "ocrmypdf",
		"executable":    toolPath,
		"sha256":        "not-a-sha",
		"size":          "1",
		"version":       "17.11.0",
		"upstream":      "ocrmypdf/OCRmyPDF",
		"license":       "MPL-2.0",
		"registered_at": time.Now().UTC().Format(time.RFC3339Nano),
		"verified":      true,
	})
	if err := v.saveUnlocked(); err != nil {
		v.mu.Unlock()
		t.Fatal(err)
	}
	v.mu.Unlock()

	if _, err := v.RegisteredLocalTool("ocrmypdf"); err == nil || !strings.Contains(err.Error(), "invalid") && !strings.Contains(err.Error(), "inconsistent") {
		t.Fatalf("malformed newest registration should fail closed, got %v", err)
	}
}

func TestLocalToolRegistrationRejectsMutationDuringProbe(t *testing.T) {
	d := t.TempDir()
	v, err := OpenVault(filepath.Join(d, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	toolPath := writeLocalToolFixture(t, d, "mutating-tool.bin", "before")
	probe := func(_ context.Context, _ string, executable string) (string, error) {
		if err := os.WriteFile(executable, []byte("after mutation with different bytes"), 0700); err != nil {
			return "", err
		}
		return "test-version", nil
	}
	if _, err := v.registerLocalToolWithProbe(context.Background(), "tesseract", toolPath, probe); err == nil || !strings.Contains(err.Error(), "changed while") {
		t.Fatalf("expected mutation-during-registration rejection, got %v", err)
	}
}
