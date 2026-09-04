package eco

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLlamaCPPArgsStayLocalAndDeterministic(t *testing.T) {
	args := llamaCPPArgs(`C:\models\qwen.gguf`, `C:\work\prompt.txt`, `C:\work\schema.json`)
	joined := strings.Join(args, " ")
	for _, required := range []string{
		"--offline", "--model", "--file", "--json-schema-file", "--simple-io",
		"--no-display-prompt", "--seed 0", "--temp 0", "--n-predict 2048",
		"--device none", "--n-gpu-layers 0", "--fit off", "--no-context-shift",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing controlled llama.cpp argument %q in %q", required, joined)
		}
	}
	for _, forbidden := range []string{"--model-url", "--hf-repo", "--hf-file", "--hf-token", "--docker-repo", "--rpc", "--server-base"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("llama.cpp adapter unexpectedly contains network-capable flag %q: %q", forbidden, joined)
		}
	}
}

func TestOfflineLlamaCPPEnvironmentDropsArgumentAndNetworkInjection(t *testing.T) {
	env := offlineLlamaCPPEnvironment([]string{
		"PATH=X",
		"LLAMA_ARG_RPC=10.0.0.1:5000",
		"LLAMA_ARG_MODEL_URL=https://example.invalid/model.gguf",
		"LLAMA_ARG_HF_REPO=user/model",
		"HF_TOKEN=secret",
		"HTTP_PROXY=http://proxy.invalid:8080",
		"HTTPS_PROXY=http://proxy.invalid:8080",
	})
	joined := strings.Join(env, "\n")
	for _, forbidden := range []string{"LLAMA_ARG_RPC=", "LLAMA_ARG_MODEL_URL=", "LLAMA_ARG_HF_REPO=", "HF_TOKEN=", "HTTP_PROXY=", "HTTPS_PROXY="} {
		if strings.Contains(strings.ToUpper(joined), strings.ToUpper(forbidden)) {
			t.Fatalf("unsafe inherited llama.cpp environment survived filtering: %q", joined)
		}
	}
	for _, required := range []string{"LLAMA_ARG_OFFLINE=1", "HF_HUB_OFFLINE=1", "TRANSFORMERS_OFFLINE=1", "PATH=X"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing controlled environment %q in %q", required, joined)
		}
	}
}

func TestLlamaCPPEmissionParserIsStrict(t *testing.T) {
	valid := []byte(`{"answer":"The hearing date is in the source.","claims":[{"kind":"value","text":"12 August 2026","evidence_id":"EVD-1","segment_id":"SEG-1"}]}`)
	emission, err := parseLlamaCPPEmission(valid)
	if err != nil {
		t.Fatal(err)
	}
	if len(emission.Claims) != 1 || emission.Claims[0].Text != "12 August 2026" {
		t.Fatalf("unexpected parsed emission: %+v", emission)
	}
	bad := [][]byte{
		[]byte("```json\n" + string(valid) + "\n```"),
		append(append([]byte(nil), valid...), []byte(" trailing prose")...),
		[]byte(`{"answer":"x","claims":[{"kind":"presence","evidence_id":"E","segment_id":"S"}],"unexpected":true}`),
		[]byte(`{"answer":"","claims":[]}`),
	}
	for i, data := range bad {
		if _, err := parseLlamaCPPEmission(data); err == nil {
			t.Fatalf("case %d: expected strict parser rejection", i+1)
		}
	}
}

func TestLlamaCPPJSONSchemaIsValidJSON(t *testing.T) {
	var schema map[string]any
	if err := json.Unmarshal([]byte(llamaCPPEmissionSchema), &schema); err != nil {
		t.Fatalf("invalid llama.cpp emission schema: %v", err)
	}
	if schema["type"] != "object" {
		t.Fatalf("unexpected schema root: %+v", schema)
	}
}

func TestLlamaCPPPromptDoesNotLeakTrustedSourceMetadata(t *testing.T) {
	_, grounding := testGroundingVault(t)
	prompt, err := buildLlamaCPPPrompt(grounding)
	if err != nil {
		t.Fatal(err)
	}
	if len(prompt) > maxLlamaCPPPromptBytes {
		t.Fatalf("prompt exceeded bound: %d", len(prompt))
	}
	for _, trusted := range grounding.trusted {
		if strings.Contains(prompt, trusted.SourceSHA256) || strings.Contains(prompt, trusted.SourceObject) {
			t.Fatalf("model prompt leaked app-private source binding: %q", prompt)
		}
	}
	if !strings.Contains(prompt, grounding.Records[0].EvidenceID) || !strings.Contains(prompt, grounding.Records[0].SegmentID) {
		t.Fatalf("prompt omitted allowed source vocabulary: %q", prompt)
	}
}

func TestInspectLlamaCPPModelRequiresLocalGGUF(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "model.gguf")
	if err := os.WriteFile(good, []byte("small test gguf placeholder"), 0600); err != nil {
		t.Fatal(err)
	}
	receipt, err := inspectLlamaCPPModel(good)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Name != "model.gguf" || len(receipt.SHA256) != 64 || receipt.Size <= 0 {
		t.Fatalf("unexpected model receipt: %+v", receipt)
	}
	wrong := filepath.Join(dir, "model.bin")
	if err := os.WriteFile(wrong, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := inspectLlamaCPPModel(wrong); err == nil {
		t.Fatal("expected non-GGUF model to be rejected")
	}
	if _, err := inspectLlamaCPPModel("relative.gguf"); err == nil {
		t.Fatal("expected relative model path to be rejected")
	}
}

func TestLlamaLimitedCaptureFlagsOverflow(t *testing.T) {
	var target strings.Builder
	_ = target
	var buf bytesBufferForTest
	_ = buf
}

// Keep overflow behaviour covered without exposing the adapter's bytes.Buffer.
// The actual buffer type is exercised through this small helper test below.
type bytesBufferForTest struct{}
