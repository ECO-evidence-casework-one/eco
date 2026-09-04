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

func healthyResourceSnapshot() ResourceSnapshot {
	return ResourceSnapshot{
		CapturedAt:           time.Now().UTC(),
		LogicalCPUs:          8,
		CPUSampled:           true,
		CPUUsedPercent:       25,
		MemorySampled:        true,
		MemoryTotalBytes:     8 * resourceGiB,
		MemoryAvailableBytes: 4 * resourceGiB,
		MemoryUsedPercent:    50,
		DiskSampled:          true,
		DiskTotalBytes:       100 * resourceGiB,
		DiskFreeBytes:        50 * resourceGiB,
		DiskUsedPercent:      50,
	}
}

func TestAssessLocalResourcesBlocksCriticalMemoryPressure(t *testing.T) {
	snapshot := healthyResourceSnapshot()
	snapshot.MemoryAvailableBytes = 128 * resourceMiB
	snapshot.MemoryUsedPercent = 99
	assessment := AssessLocalResources(snapshot, DefaultEngineResourcePolicy("test-engine", t.TempDir()))
	if !assessment.Blocked || assessment.Level != "critical" {
		t.Fatalf("critical RAM pressure was not blocked: %+v", assessment)
	}
	if len(assessment.Reasons) < 1 || !strings.Contains(strings.Join(assessment.Reasons, " "), "RAM") {
		t.Fatalf("critical RAM reason is missing: %+v", assessment)
	}
}

func TestAssessLocalResourcesBlocksCriticalDiskPressure(t *testing.T) {
	snapshot := healthyResourceSnapshot()
	snapshot.DiskFreeBytes = 128 * resourceMiB
	assessment := AssessLocalResources(snapshot, DefaultEngineResourcePolicy("test-engine", t.TempDir()))
	if !assessment.Blocked || assessment.Level != "critical" {
		t.Fatalf("critical disk pressure was not blocked: %+v", assessment)
	}
	if !strings.Contains(strings.Join(assessment.Reasons, " "), "working disk") {
		t.Fatalf("critical disk reason is missing: %+v", assessment)
	}
}

func TestAssessLocalResourcesHighCPUWarnsButNeverBlocksAlone(t *testing.T) {
	snapshot := healthyResourceSnapshot()
	snapshot.CPUUsedPercent = 99.5
	assessment := AssessLocalResources(snapshot, DefaultEngineResourcePolicy("test-engine", t.TempDir()))
	if assessment.Blocked {
		t.Fatalf("high CPU load alone must not block a local engine: %+v", assessment)
	}
	if assessment.Level != "elevated" || !strings.Contains(strings.Join(assessment.Warnings, " "), "CPU") {
		t.Fatalf("high CPU load was not surfaced as an advisory: %+v", assessment)
	}
}

func TestLlamaCPPModelHeadroomIsAdvisoryOnly(t *testing.T) {
	snapshot := healthyResourceSnapshot()
	snapshot.MemoryAvailableBytes = 2 * resourceGiB
	policy := LlamaCPPResourcePolicy(int64(3 * resourceGiB))
	assessment := AssessLocalResources(snapshot, policy)
	if assessment.Blocked {
		t.Fatalf("GGUF size-derived headroom must remain advisory: %+v", assessment)
	}
	if assessment.Level != "elevated" || !strings.Contains(strings.Join(assessment.Warnings, " "), "GGUF") {
		t.Fatalf("GGUF advisory was not surfaced: %+v", assessment)
	}
	if policy.CriticalAvailableMemoryBytes != defaultCriticalAvailableMemoryBytes {
		t.Fatalf("model size unexpectedly changed hard memory threshold: %+v", policy)
	}
}

func TestUnavailableResourceTelemetryWarnsWithoutBlocking(t *testing.T) {
	snapshot := ResourceSnapshot{
		CapturedAt:  time.Now().UTC(),
		LogicalCPUs: 4,
		Unavailable: []string{"memory: unavailable", "cpu: unavailable"},
	}
	assessment := AssessLocalResources(snapshot, DefaultEngineResourcePolicy("test-engine", ""))
	if assessment.Blocked || assessment.Level != "elevated" {
		t.Fatalf("telemetry failure should degrade to an advisory, not a blanket block: %+v", assessment)
	}
	if len(assessment.Warnings) != 2 {
		t.Fatalf("telemetry warnings were not preserved: %+v", assessment)
	}
}

func TestResourcePressureErrorNamesEngineAndReasons(t *testing.T) {
	err := &ResourcePressureError{Engine: "llama.cpp", Reasons: []string{"only 128 MiB RAM is available"}}
	if !strings.Contains(err.Error(), "llama.cpp") || !strings.Contains(err.Error(), "128 MiB") {
		t.Fatalf("resource pressure error lost useful context: %q", err.Error())
	}
}

func TestResourceProbeDirectoryUsesParentForRegularFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "model.gguf")
	if err := os.WriteFile(path, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := resourceProbeDirectory(path)
	if err != nil {
		t.Fatal(err)
	}
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Clean(got) != filepath.Clean(resolvedDir) {
		t.Fatalf("resource probe used wrong disk directory: got %q want %q", got, resolvedDir)
	}
	if _, err := resourceProbeDirectory("relative/path"); err == nil {
		t.Fatal("expected relative resource path rejection")
	}
}

func TestSampleLocalResourcesSmoke(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	snapshot, err := SampleLocalResources(ctx, t.TempDir(), 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.CapturedAt.IsZero() || snapshot.LogicalCPUs < 1 {
		t.Fatalf("resource sampler returned an incomplete base snapshot: %+v", snapshot)
	}
	if !snapshot.MemorySampled {
		t.Fatalf("local RAM sampling is required on qualified Linux/Windows platforms: %+v", snapshot)
	}
	if !snapshot.DiskSampled {
		t.Fatalf("local disk sampling is required on qualified Linux/Windows platforms: %+v", snapshot)
	}
	if !snapshot.CPUSampled {
		t.Fatalf("local CPU sampling is required on qualified Linux/Windows platforms: %+v", snapshot)
	}
}

func TestSampleLocalResourcesRejectsUnsafeControlInputs(t *testing.T) {
	if _, err := SampleLocalResources(nil, "", 10*time.Millisecond); err == nil {
		t.Fatal("expected nil context rejection")
	}
	ctx := context.Background()
	if _, err := SampleLocalResources(ctx, "", maxResourceSampleInterval+time.Millisecond); err == nil {
		t.Fatal("expected unbounded sampling interval rejection")
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := SampleLocalResources(cancelled, "", 50*time.Millisecond)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation to propagate, got %v", err)
	}
}
