package eco

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalAIStatusReportsUnconfiguredSourceOnlyMode(t *testing.T) {
	clearLocalAIEnv(t)
	status := (&Vault{}).LocalAIStatusContext(context.Background())
	if status.State != "unavailable" || status.Configured || status.Ready {
		t.Fatalf("unexpected unconfigured status: %#v", status)
	}
	if !strings.Contains(status.Detail, "deterministic source-backed") {
		t.Fatalf("unconfigured detail must explain the source-backed fallback: %q", status.Detail)
	}
}

func TestLocalAIStatusReportsPartialConfigurationAsInvalid(t *testing.T) {
	clearLocalAIEnv(t)
	t.Setenv(envLlamaCPPExecutable, filepath.Join(t.TempDir(), "llama-cli.exe"))
	status := (&Vault{}).LocalAIStatusContext(context.Background())
	if status.State != "invalid" || !status.Configured || status.Ready {
		t.Fatalf("unexpected partial-configuration status: %#v", status)
	}
	if !strings.Contains(status.Detail, "could not be verified") {
		t.Fatalf("invalid detail must explain verification failure: %q", status.Detail)
	}
}

func TestQuestionUsedLocalAIDistinguishesGroundedModelFromFallback(t *testing.T) {
	model := QuestionRecord{Support: "Local llama.cpp selected the passages; ECO released only deterministically grounded source wording. Source truth remains unverified."}
	if !QuestionUsedLocalAI(model) {
		t.Fatal("grounded llama.cpp answer must be identified as local-AI used")
	}
	fallback := QuestionRecord{Support: "Directly supported by cited source passages"}
	if QuestionUsedLocalAI(fallback) {
		t.Fatal("deterministic fallback must not be identified as local-AI used")
	}
}

func TestBaseNameHandlesWindowsAndSlashPaths(t *testing.T) {
	for input, want := range map[string]string{
		`E:\ECO\AI\qwen.gguf`: "qwen.gguf",
		`/opt/eco/qwen.gguf`:     "qwen.gguf",
		`qwen.gguf`:              "qwen.gguf",
	} {
		if got := baseName(input); got != want {
			t.Fatalf("baseName(%q) = %q, want %q", input, got, want)
		}
	}
}
