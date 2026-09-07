package eco

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func clearLocalAIEnv(t *testing.T) {
	t.Helper()
	t.Setenv(envLlamaCPPExecutable, "")
	t.Setenv(envLlamaCPPExecutableSHA256, "")
	t.Setenv(envLlamaCPPModel, "")
	t.Setenv(envLlamaCPPModelSHA256, "")
}

func TestLoadConfiguredLocalAINotConfigured(t *testing.T) {
	clearLocalAIEnv(t)
	_, configured, err := loadConfiguredLocalAI()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if configured {
		t.Fatal("empty environment must not enable local AI")
	}
}

func TestLoadConfiguredLocalAIRejectsPartialConfiguration(t *testing.T) {
	clearLocalAIEnv(t)
	t.Setenv(envLlamaCPPExecutable, filepath.Join(t.TempDir(), "llama-cli.exe"))
	_, configured, err := loadConfiguredLocalAI()
	if !configured {
		t.Fatal("partial local AI environment must count as an attempted configuration")
	}
	if err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("expected incomplete-configuration error, got %v", err)
	}
}

func TestLoadConfiguredLocalAIAcceptsHashPinnedLocalFiles(t *testing.T) {
	clearLocalAIEnv(t)
	root := t.TempDir()
	executable := filepath.Join(root, "llama-cli.exe")
	model := filepath.Join(root, "qwen2.5-1.5b-instruct-q4_k_m.gguf")
	if err := os.WriteFile(executable, []byte("synthetic llama runtime"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(model, []byte("synthetic gguf model"), 0600); err != nil {
		t.Fatal(err)
	}
	executableHash, err := hashFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	modelHash, err := hashFile(model)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(envLlamaCPPExecutable, executable)
	t.Setenv(envLlamaCPPExecutableSHA256, strings.ToUpper(executableHash))
	t.Setenv(envLlamaCPPModel, model)
	t.Setenv(envLlamaCPPModelSHA256, strings.ToUpper(modelHash))

	cfg, configured, err := loadConfiguredLocalAI()
	if err != nil {
		t.Fatalf("expected valid local AI configuration, got %v", err)
	}
	if !configured {
		t.Fatal("complete hash-pinned environment must enable local AI")
	}
	// Windows may expand an 8.3 short temp path to its long form. Compare the
	// actual file identities rather than spelling; product code separately
	// requires absolute regular files and validates their exact bytes by SHA-256.
	wantExeInfo, err := os.Stat(executable)
	if err != nil {
		t.Fatal(err)
	}
	gotExeInfo, err := os.Stat(cfg.Executable)
	if err != nil {
		t.Fatal(err)
	}
	wantModelInfo, err := os.Stat(model)
	if err != nil {
		t.Fatal(err)
	}
	gotModelInfo, err := os.Stat(cfg.Model)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(wantExeInfo, gotExeInfo) || !os.SameFile(wantModelInfo, gotModelInfo) {
		t.Fatalf("resolved paths do not identify the configured files: %#v", cfg)
	}
	if cfg.ExecutableSHA256 != executableHash || cfg.ModelSHA256 != modelHash {
		t.Fatalf("hashes were not normalized/preserved: %#v", cfg)
	}
}

func TestLoadConfiguredLocalAIRejectsHashMismatch(t *testing.T) {
	clearLocalAIEnv(t)
	root := t.TempDir()
	executable := filepath.Join(root, "llama-cli.exe")
	model := filepath.Join(root, "model.gguf")
	if err := os.WriteFile(executable, []byte("synthetic llama runtime"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(model, []byte("synthetic gguf model"), 0600); err != nil {
		t.Fatal(err)
	}
	modelHash, err := hashFile(model)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(envLlamaCPPExecutable, executable)
	t.Setenv(envLlamaCPPExecutableSHA256, strings.Repeat("0", 64))
	t.Setenv(envLlamaCPPModel, model)
	t.Setenv(envLlamaCPPModelSHA256, modelHash)

	_, configured, err := loadConfiguredLocalAI()
	if !configured {
		t.Fatal("mismatched hash is still an attempted configuration")
	}
	if err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("expected SHA-256 mismatch error, got %v", err)
	}
}
