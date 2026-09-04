package eco

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

const (
	resourceMiB = uint64(1024 * 1024)
	resourceGiB = uint64(1024) * resourceMiB

	defaultResourceSampleInterval         = 150 * time.Millisecond
	maxResourceSampleInterval             = 2 * time.Second
	defaultCriticalAvailableMemoryBytes   = 256 * resourceMiB
	defaultWarningAvailableMemoryBytes    = 1 * resourceGiB
	defaultCriticalDiskFreeBytes          = 256 * resourceMiB
	defaultWarningDiskFreeBytes           = 1 * resourceGiB
	defaultCriticalMemoryUsedPercent      = 98.0
	defaultWarningMemoryUsedPercent       = 92.0
	defaultWarningCPUUsedPercent          = 95.0
	localModelAdvisoryMemoryHeadroomBytes = 512 * resourceMiB
)

type ResourceSnapshot struct {
	CapturedAt           time.Time `json:"captured_at"`
	SampleIntervalMillis int64     `json:"sample_interval_millis"`
	LogicalCPUs          int       `json:"logical_cpus"`
	CPUUsedPercent       float64   `json:"cpu_used_percent,omitempty"`
	MemoryTotalBytes     uint64    `json:"memory_total_bytes,omitempty"`
	MemoryAvailableBytes uint64    `json:"memory_available_bytes,omitempty"`
	MemoryUsedPercent    float64   `json:"memory_used_percent,omitempty"`
	DiskPath             string    `json:"disk_path,omitempty"`
	DiskTotalBytes       uint64    `json:"disk_total_bytes,omitempty"`
	DiskFreeBytes        uint64    `json:"disk_free_bytes,omitempty"`
	DiskUsedPercent      float64   `json:"disk_used_percent,omitempty"`
	CPUSampled           bool      `json:"cpu_sampled"`
	MemorySampled        bool      `json:"memory_sampled"`
	DiskSampled          bool      `json:"disk_sampled"`
	Unavailable          []string  `json:"unavailable,omitempty"`
}

type EngineResourcePolicy struct {
	Engine                       string        `json:"engine"`
	DiskPath                     string        `json:"disk_path,omitempty"`
	SampleInterval               time.Duration `json:"-"`
	CriticalAvailableMemoryBytes uint64        `json:"critical_available_memory_bytes"`
	WarningAvailableMemoryBytes  uint64        `json:"warning_available_memory_bytes"`
	CriticalDiskFreeBytes        uint64        `json:"critical_disk_free_bytes"`
	WarningDiskFreeBytes         uint64        `json:"warning_disk_free_bytes"`
	CriticalMemoryUsedPercent    float64       `json:"critical_memory_used_percent"`
	WarningMemoryUsedPercent     float64       `json:"warning_memory_used_percent"`
	WarningCPUUsedPercent        float64       `json:"warning_cpu_used_percent"`
	AdvisoryMemoryBytes          uint64        `json:"advisory_memory_bytes,omitempty"`
	AdvisoryReason               string        `json:"advisory_reason,omitempty"`
}

type ResourceAssessment struct {
	Snapshot ResourceSnapshot `json:"snapshot"`
	Level    string           `json:"level"`
	Blocked  bool             `json:"blocked"`
	Reasons  []string         `json:"reasons,omitempty"`
	Warnings []string         `json:"warnings,omitempty"`
}

type ResourcePressureError struct {
	Engine  string
	Reasons []string
}

func (e *ResourcePressureError) Error() string {
	engine := strings.TrimSpace(e.Engine)
	if engine == "" {
		engine = "local engine"
	}
	if len(e.Reasons) == 0 {
		return engine + " blocked by critical local resource pressure"
	}
	return engine + " blocked by critical local resource pressure: " + strings.Join(e.Reasons, "; ")
}

// DefaultEngineResourcePolicy blocks only genuinely critical host pressure.
// The advisory memory threshold is warning-only because a model/file size is
// not the same thing as a runtime working set.
func DefaultEngineResourcePolicy(engine, diskPath string) EngineResourcePolicy {
	return EngineResourcePolicy{
		Engine:                       strings.TrimSpace(engine),
		DiskPath:                     strings.TrimSpace(diskPath),
		SampleInterval:               defaultResourceSampleInterval,
		CriticalAvailableMemoryBytes: defaultCriticalAvailableMemoryBytes,
		WarningAvailableMemoryBytes:  defaultWarningAvailableMemoryBytes,
		CriticalDiskFreeBytes:        defaultCriticalDiskFreeBytes,
		WarningDiskFreeBytes:         defaultWarningDiskFreeBytes,
		CriticalMemoryUsedPercent:    defaultCriticalMemoryUsedPercent,
		WarningMemoryUsedPercent:     defaultWarningMemoryUsedPercent,
		WarningCPUUsedPercent:        defaultWarningCPUUsedPercent,
	}
}

func LlamaCPPResourcePolicy(modelSize int64) EngineResourcePolicy {
	policy := DefaultEngineResourcePolicy("llama.cpp", os.TempDir())
	if modelSize > 0 {
		modelBytes := uint64(modelSize)
		maxUint64 := ^uint64(0)
		if modelBytes <= maxUint64-localModelAdvisoryMemoryHeadroomBytes {
			policy.AdvisoryMemoryBytes = modelBytes + localModelAdvisoryMemoryHeadroomBytes
		} else {
			policy.AdvisoryMemoryBytes = maxUint64
		}
		policy.AdvisoryReason = "available RAM is below the GGUF file size plus conservative headroom; this is an advisory only because runtime memory use depends on model/layout settings"
	}
	return policy
}

// SampleLocalResources reads only local machine state. gopsutil is used as a
// direct library and no network package or remote endpoint is involved.
// Individual metric failures are reported in Unavailable rather than turning a
// telemetry problem into a blanket application outage. Context cancellation is
// always propagated.
func SampleLocalResources(ctx context.Context, diskPath string, interval time.Duration) (ResourceSnapshot, error) {
	if ctx == nil {
		return ResourceSnapshot{}, errors.New("resource sampling context is required")
	}
	if interval <= 0 {
		interval = defaultResourceSampleInterval
	}
	if interval > maxResourceSampleInterval {
		return ResourceSnapshot{}, fmt.Errorf("resource sampling interval exceeds %s", maxResourceSampleInterval)
	}

	snapshot := ResourceSnapshot{
		LogicalCPUs:          runtime.NumCPU(),
		SampleIntervalMillis: interval.Milliseconds(),
	}

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		if contextError(err) {
			return ResourceSnapshot{}, err
		}
		snapshot.Unavailable = append(snapshot.Unavailable, "memory: "+truncateResourceDiagnostic(err.Error()))
	} else if vm == nil || vm.Total == 0 || !validPercent(vm.UsedPercent) {
		snapshot.Unavailable = append(snapshot.Unavailable, "memory: invalid local memory sample")
	} else {
		snapshot.MemorySampled = true
		snapshot.MemoryTotalBytes = vm.Total
		snapshot.MemoryAvailableBytes = vm.Available
		snapshot.MemoryUsedPercent = vm.UsedPercent
	}

	if strings.TrimSpace(diskPath) != "" {
		probePath, pathErr := resourceProbeDirectory(diskPath)
		if pathErr != nil {
			snapshot.Unavailable = append(snapshot.Unavailable, "disk: "+truncateResourceDiagnostic(pathErr.Error()))
		} else {
			du, diskErr := disk.UsageWithContext(ctx, probePath)
			if diskErr != nil {
				if contextError(diskErr) {
					return ResourceSnapshot{}, diskErr
				}
				snapshot.Unavailable = append(snapshot.Unavailable, "disk: "+truncateResourceDiagnostic(diskErr.Error()))
			} else if du == nil || du.Total == 0 || !validPercent(du.UsedPercent) {
				snapshot.Unavailable = append(snapshot.Unavailable, "disk: invalid local disk sample")
			} else {
				snapshot.DiskSampled = true
				snapshot.DiskPath = probePath
				snapshot.DiskTotalBytes = du.Total
				snapshot.DiskFreeBytes = du.Free
				snapshot.DiskUsedPercent = du.UsedPercent
			}
		}
	}

	cpuPercent, err := cpu.PercentWithContext(ctx, interval, false)
	if err != nil {
		if contextError(err) {
			return ResourceSnapshot{}, err
		}
		snapshot.Unavailable = append(snapshot.Unavailable, "cpu: "+truncateResourceDiagnostic(err.Error()))
	} else if len(cpuPercent) != 1 || !validPercent(cpuPercent[0]) {
		snapshot.Unavailable = append(snapshot.Unavailable, "cpu: invalid local CPU sample")
	} else {
		snapshot.CPUSampled = true
		snapshot.CPUUsedPercent = cpuPercent[0]
	}

	snapshot.CapturedAt = time.Now().UTC()
	return snapshot, nil
}

func AssessLocalResources(snapshot ResourceSnapshot, policy EngineResourcePolicy) ResourceAssessment {
	assessment := ResourceAssessment{Snapshot: snapshot, Level: "normal"}

	if snapshot.MemorySampled {
		if policy.CriticalAvailableMemoryBytes > 0 && snapshot.MemoryAvailableBytes < policy.CriticalAvailableMemoryBytes {
			assessment.Blocked = true
			assessment.Reasons = append(assessment.Reasons, fmt.Sprintf("only %s RAM is available", formatResourceBytes(snapshot.MemoryAvailableBytes)))
		}
		if policy.CriticalMemoryUsedPercent > 0 && snapshot.MemoryUsedPercent >= policy.CriticalMemoryUsedPercent {
			assessment.Blocked = true
			assessment.Reasons = append(assessment.Reasons, fmt.Sprintf("RAM usage is %.1f%%", snapshot.MemoryUsedPercent))
		}
		if !assessment.Blocked {
			if policy.WarningAvailableMemoryBytes > 0 && snapshot.MemoryAvailableBytes < policy.WarningAvailableMemoryBytes {
				assessment.Warnings = append(assessment.Warnings, fmt.Sprintf("available RAM is low (%s)", formatResourceBytes(snapshot.MemoryAvailableBytes)))
			}
			if policy.WarningMemoryUsedPercent > 0 && snapshot.MemoryUsedPercent >= policy.WarningMemoryUsedPercent {
				assessment.Warnings = append(assessment.Warnings, fmt.Sprintf("RAM usage is high (%.1f%%)", snapshot.MemoryUsedPercent))
			}
			if policy.AdvisoryMemoryBytes > 0 && snapshot.MemoryAvailableBytes < policy.AdvisoryMemoryBytes {
				reason := strings.TrimSpace(policy.AdvisoryReason)
				if reason == "" {
					reason = "available RAM is below this engine's advisory headroom"
				}
				assessment.Warnings = append(assessment.Warnings, reason)
			}
		}
	}

	if snapshot.DiskSampled {
		if policy.CriticalDiskFreeBytes > 0 && snapshot.DiskFreeBytes < policy.CriticalDiskFreeBytes {
			assessment.Blocked = true
			assessment.Reasons = append(assessment.Reasons, fmt.Sprintf("only %s is free on the working disk", formatResourceBytes(snapshot.DiskFreeBytes)))
		} else if policy.WarningDiskFreeBytes > 0 && snapshot.DiskFreeBytes < policy.WarningDiskFreeBytes {
			assessment.Warnings = append(assessment.Warnings, fmt.Sprintf("working-disk free space is low (%s)", formatResourceBytes(snapshot.DiskFreeBytes)))
		}
	}

	if snapshot.CPUSampled && policy.WarningCPUUsedPercent > 0 && snapshot.CPUUsedPercent >= policy.WarningCPUUsedPercent {
		assessment.Warnings = append(assessment.Warnings, fmt.Sprintf("CPU usage is already high (%.1f%%); ECO will not block solely for CPU load", snapshot.CPUUsedPercent))
	}
	for _, unavailable := range snapshot.Unavailable {
		assessment.Warnings = append(assessment.Warnings, "resource telemetry unavailable: "+unavailable)
	}

	switch {
	case assessment.Blocked:
		assessment.Level = "critical"
	case len(assessment.Warnings) > 0:
		assessment.Level = "elevated"
	default:
		assessment.Level = "normal"
	}
	return assessment
}

func CheckLocalEngineResources(ctx context.Context, policy EngineResourcePolicy) (ResourceAssessment, error) {
	snapshot, err := SampleLocalResources(ctx, policy.DiskPath, policy.SampleInterval)
	if err != nil {
		return ResourceAssessment{}, err
	}
	assessment := AssessLocalResources(snapshot, policy)
	if assessment.Blocked {
		return assessment, &ResourcePressureError{Engine: policy.Engine, Reasons: append([]string(nil), assessment.Reasons...)}
	}
	return assessment, nil
}

func resourceProbeDirectory(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("resource disk path is empty")
	}
	if !filepath.IsAbs(path) {
		return "", errors.New("resource disk path must be absolute")
	}
	resolved, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("resolve resource disk path: %w", err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("open resource disk path: %w", err)
	}
	if info.Mode().IsRegular() {
		resolved = filepath.Dir(resolved)
		info, err = os.Stat(resolved)
		if err != nil {
			return "", fmt.Errorf("open resource disk directory: %w", err)
		}
	}
	if !info.IsDir() {
		return "", errors.New("resource disk path is not a file or directory")
	}
	return resolved, nil
}

func validPercent(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 100
}

func contextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func truncateResourceDiagnostic(text string) string {
	text = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(text, "\r", " "), "\n", " "))
	if len([]rune(text)) > 300 {
		return string([]rune(text)[:300]) + "…"
	}
	return text
}

func formatResourceBytes(value uint64) string {
	switch {
	case value >= resourceGiB:
		return fmt.Sprintf("%.2f GiB", float64(value)/float64(resourceGiB))
	case value >= resourceMiB:
		return fmt.Sprintf("%.0f MiB", float64(value)/float64(resourceMiB))
	default:
		return fmt.Sprintf("%d bytes", value)
	}
}
