package eco

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDoclingExtractionArgsAreLocalAndBounded(t *testing.T) {
	got := doclingExtractionArgs(`C:\evidence\report.pdf`, `C:\temp\out`, `C:\models\docling`)
	want := []string{
		`C:\evidence\report.pdf`,
		"--to", "md",
		"--output", `C:\temp\out`,
		"--artifacts-path", `C:\models\docling`,
		"--pipeline", "standard",
		"--device", "cpu",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args mismatch\n got: %#v\nwant: %#v", got, want)
	}
	for _, arg := range got {
		if strings.Contains(strings.ToLower(arg), "remote") || strings.HasPrefix(strings.ToLower(arg), "http://") || strings.HasPrefix(strings.ToLower(arg), "https://") {
			t.Fatalf("Docling local adapter unexpectedly enables a remote path: %q", arg)
		}
	}
}

func TestOfflineDoclingEnvironmentOverridesNetworkModelClients(t *testing.T) {
	env := offlineDoclingEnvironment([]string{
		"PATH=X",
		"HF_HUB_OFFLINE=0",
		"TRANSFORMERS_OFFLINE=0",
		"DOCLING_ENABLE_REMOTE_SERVICES=true",
	}, `C:\models\docling`)
	joined := strings.Join(env, "\n")
	for _, want := range []string{
		"HF_HUB_OFFLINE=1",
		"TRANSFORMERS_OFFLINE=1",
		"DOCLING_ENABLE_REMOTE_SERVICES=false",
		`DOCLING_ARTIFACTS_PATH=C:\models\docling`,
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing offline environment setting %q in %q", want, joined)
		}
	}
	if strings.Contains(joined, "HF_HUB_OFFLINE=0") || strings.Contains(joined, "DOCLING_ENABLE_REMOTE_SERVICES=true") {
		t.Fatalf("unsafe inherited Docling environment survived override: %q", joined)
	}
}

func TestRequireAbsoluteDirectory(t *testing.T) {
	dir := t.TempDir()
	resolved, err := requireAbsoluteDirectory(dir, "models")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(resolved) {
		t.Fatalf("resolved path is not absolute: %q", resolved)
	}
	if _, err := requireAbsoluteDirectory("relative-models", "models"); err == nil {
		t.Fatal("expected relative model directory to be rejected")
	}
}

func TestBoundedBufferDoesNotGrowPastLimit(t *testing.T) {
	var target bytes.Buffer
	writer := &boundedBuffer{buf: &target, max: 5}
	payload := []byte("0123456789")
	n, err := writer.Write(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(payload) {
		t.Fatalf("writer must report the input consumed; got %d want %d", n, len(payload))
	}
	if target.String() != "01234" {
		t.Fatalf("diagnostic buffer exceeded bound or stored wrong prefix: %q", target.String())
	}
}
