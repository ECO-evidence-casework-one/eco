//go:build windows

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/ECO-evidence-casework-one/eco/internal/eco"
	"github.com/shirou/gopsutil/v4/process"
)

const (
	qualifierVersion             = "ECO-PDF-LOW-SPEC-20260905.1"
	qualifiedRuntimeName         = "pdfium-webassembly-windows-amd64.exe"
	qualifiedRuntimeBytes  int64 = 16988160
	qualifiedRuntimeSHA256       = "b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f"
	adapterWidth                 = 1600
	functionalTimeout            = 45 * time.Second
	acceptMedianMillis     int64 = 10000
	acceptWorstMillis      int64 = 15000
	acceptPeakMiB                = 768.0
)

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procMessageBox = user32.NewProc("MessageBoxW")
)

type hardwareInfo struct {
	OS                 string  `json:"os"`
	Architecture       string  `json:"architecture"`
	Processor          string  `json:"processor"`
	LogicalCPUs        int     `json:"logical_cpus"`
	MemoryTotalBytes   uint64  `json:"memory_total_bytes,omitempty"`
	MemoryAvailBytes   uint64  `json:"memory_available_bytes,omitempty"`
	MemoryUsedPercent  float64 `json:"memory_used_percent,omitempty"`
	MemorySampled      bool    `json:"memory_sampled"`
	ResourceAssessment string  `json:"resource_assessment,omitempty"`
}

type pageSample struct {
	Ordinal      int     `json:"ordinal"`
	Page         int     `json:"page"`
	Milliseconds int64   `json:"milliseconds"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	PNGBytes     int     `json:"png_bytes"`
	MiBPerSecond float64 `json:"-"`
}

type acceptanceCriteria struct {
	FunctionalMustPass            bool    `json:"functional_must_pass"`
	MedianRenderMillisAtMost      int64   `json:"median_render_millis_at_most"`
	WorstRenderMillisAtMost       int64   `json:"worst_render_millis_at_most"`
	PeakRendererWorkingSetMiBMax  float64 `json:"peak_renderer_working_set_mib_max"`
	ProductRenderTimeoutMillis    int64   `json:"product_render_timeout_millis"`
	CriticalResourcePressureBlock bool    `json:"critical_resource_pressure_block"`
}

type qualificationReport struct {
	QualifierVersion        string             `json:"qualifier_version"`
	CombinedSourceCommit    string             `json:"combined_source_commit"`
	StartedAt               time.Time          `json:"started_at"`
	FinishedAt              time.Time          `json:"finished_at"`
	Status                  string             `json:"status"`
	Summary                 string             `json:"summary"`
	RuntimeFile             string             `json:"runtime_file"`
	RuntimeBytes            int64              `json:"runtime_bytes"`
	RuntimeSHA256           string             `json:"runtime_sha256"`
	RuntimeRegisteredSHA256 string             `json:"runtime_registered_sha256,omitempty"`
	Hardware                hardwareInfo       `json:"hardware"`
	PageCount               int                `json:"page_count,omitempty"`
	InfoMilliseconds        int64              `json:"page_info_milliseconds,omitempty"`
	NavigationSequence      []int              `json:"navigation_sequence"`
	PageSamples             []pageSample       `json:"page_samples,omitempty"`
	MedianRenderMillis      int64              `json:"median_render_millis,omitempty"`
	WorstRenderMillis       int64              `json:"worst_render_millis,omitempty"`
	DirectBenchmarkMillis   int64              `json:"direct_benchmark_millis,omitempty"`
	PeakRendererRSSMiB      float64            `json:"peak_renderer_rss_mib,omitempty"`
	Criteria                acceptanceCriteria `json:"acceptance_criteria"`
	Warnings                []string           `json:"warnings,omitempty"`
	Errors                  []string           `json:"errors,omitempty"`
	ReportPath              string             `json:"report_path,omitempty"`
}

func main() {
	report := qualificationReport{
		QualifierVersion:     qualifierVersion,
		CombinedSourceCommit: "85c670e2b3a5fc5a3d87d814e896d749d8cd732b",
		StartedAt:            time.Now().UTC(),
		Status:               "FAIL",
		RuntimeFile:          qualifiedRuntimeName,
		RuntimeBytes:         qualifiedRuntimeBytes,
		RuntimeSHA256:        qualifiedRuntimeSHA256,
		NavigationSequence:   []int{1, 2, 3, 2, 1},
		Criteria: acceptanceCriteria{
			FunctionalMustPass:            true,
			MedianRenderMillisAtMost:      acceptMedianMillis,
			WorstRenderMillisAtMost:       acceptWorstMillis,
			PeakRendererWorkingSetMiBMax:  acceptPeakMiB,
			ProductRenderTimeoutMillis:    functionalTimeout.Milliseconds(),
			CriticalResourcePressureBlock: true,
		},
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("unexpected qualifier panic: %v", recovered))
			report.Status = "FAIL"
			report.Summary = "The qualification program encountered an unexpected error. ECO low-spec PDF approval is withheld."
		}
		report.FinishedAt = time.Now().UTC()
		finalizeReport(&report)
	}()

	if err := runQualification(&report); err != nil {
		report.Errors = append(report.Errors, err.Error())
		report.Status = "FAIL"
		report.Summary = "Functional qualification failed. Do not approve PDF rendering/navigation for this machine yet."
		return
	}

	applyVerdict(&report)
}

func runQualification(report *qualificationReport) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	packageDir := filepath.Dir(self)
	runtimePath := filepath.Join(packageDir, qualifiedRuntimeName)
	if err := verifyRuntime(runtimePath); err != nil {
		return err
	}

	resourceCtx, cancelResources := context.WithTimeout(context.Background(), 5*time.Second)
	baseline, resourceErr := eco.SampleLocalResources(resourceCtx, packageDir, 250*time.Millisecond)
	cancelResources()
	if resourceErr != nil {
		report.Warnings = append(report.Warnings, "baseline resource sampling failed: "+resourceErr.Error())
	} else {
		report.Hardware.LogicalCPUs = baseline.LogicalCPUs
		report.Hardware.MemorySampled = baseline.MemorySampled
		report.Hardware.MemoryTotalBytes = baseline.MemoryTotalBytes
		report.Hardware.MemoryAvailBytes = baseline.MemoryAvailableBytes
		report.Hardware.MemoryUsedPercent = baseline.MemoryUsedPercent
		assessment := eco.AssessLocalResources(baseline, eco.DefaultEngineResourcePolicy("pdfium-cli", packageDir))
		report.Hardware.ResourceAssessment = assessment.Level
		if assessment.Blocked {
			return fmt.Errorf("critical local resource pressure blocked qualification: %s", strings.Join(assessment.Reasons, "; "))
		}
		for _, warning := range assessment.Warnings {
			report.Warnings = append(report.Warnings, "host resource warning: "+warning)
		}
	}
	report.Hardware.OS = windowsVersion()
	report.Hardware.Architecture = runtime.GOARCH
	report.Hardware.Processor = strings.TrimSpace(os.Getenv("PROCESSOR_IDENTIFIER"))
	if report.Hardware.Processor == "" {
		report.Hardware.Processor = "Windows processor identifier unavailable"
	}
	if report.Hardware.LogicalCPUs == 0 {
		report.Hardware.LogicalCPUs = runtime.NumCPU()
	}

	workRoot, err := os.MkdirTemp("", "ECO-PDF-LOW-SPEC-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workRoot)
	pdfPath := filepath.Join(workRoot, "three-page-qualification.pdf")
	if err := writeQualificationPDF(pdfPath); err != nil {
		return err
	}

	vault, err := eco.OpenVault(filepath.Join(workRoot, "vault"))
	if err != nil {
		return fmt.Errorf("open disposable ECO vault: %w", err)
	}
	item, duplicate, err := vault.ImportFile(pdfPath, nil)
	if err != nil {
		return fmt.Errorf("preserve qualification PDF: %w", err)
	}
	if duplicate {
		return errors.New("fresh qualification PDF was unexpectedly treated as a duplicate")
	}
	if item.DetectedType != "pdf" || !item.SourceVerified {
		return fmt.Errorf("qualification PDF did not enter a verified preserved state: type=%s verified=%v", item.DetectedType, item.SourceVerified)
	}

	registerCtx, cancelRegister := context.WithTimeout(context.Background(), 30*time.Second)
	registration, err := vault.RegisterLocalToolContext(registerCtx, "pdfium-cli", runtimePath)
	cancelRegister()
	if err != nil {
		return fmt.Errorf("register exact PDF renderer: %w", err)
	}
	if registration.SHA256 != qualifiedRuntimeSHA256 || registration.Size != qualifiedRuntimeBytes {
		return errors.New("ECO runtime registration did not retain the qualified renderer identity")
	}
	report.RuntimeRegisteredSHA256 = registration.SHA256

	infoCtx, cancelInfo := context.WithTimeout(context.Background(), functionalTimeout)
	infoStarted := time.Now()
	info, err := vault.PDFEvidenceInfoWithRegisteredPDFiumContext(infoCtx, item.ID)
	report.InfoMilliseconds = time.Since(infoStarted).Milliseconds()
	cancelInfo()
	if err != nil {
		return fmt.Errorf("read source-bound PDF page count: %w", err)
	}
	if info.PageCount != 3 || info.SourceObject != item.ObjectFile || info.SourceSHA256 != item.SHA256 {
		return fmt.Errorf("unexpected source-bound page information: pages=%d", info.PageCount)
	}
	report.PageCount = info.PageCount

	renderMillis := make([]int64, 0, len(report.NavigationSequence))
	for index, page := range report.NavigationSequence {
		renderCtx, cancelRender := context.WithTimeout(context.Background(), functionalTimeout)
		started := time.Now()
		rendered, renderErr := vault.RenderEvidencePDFPageWithRegisteredPDFiumContext(renderCtx, item.ID, page, adapterWidth)
		elapsed := time.Since(started).Milliseconds()
		cancelRender()
		if renderErr != nil {
			return fmt.Errorf("ECO navigation render %d (page %d) failed after %d ms: %w", index+1, page, elapsed, renderErr)
		}
		if rendered.Page != page || rendered.SourceObject != item.ObjectFile || rendered.SourceSHA256 != item.SHA256 || len(rendered.PNG) == 0 {
			return fmt.Errorf("ECO navigation render %d was not bound to the requested preserved page", index+1)
		}
		report.PageSamples = append(report.PageSamples, pageSample{
			Ordinal:      index + 1,
			Page:         page,
			Milliseconds: elapsed,
			Width:        rendered.Width,
			Height:       rendered.Height,
			PNGBytes:     len(rendered.PNG),
		})
		renderMillis = append(renderMillis, elapsed)
	}
	report.MedianRenderMillis = medianInt64(renderMillis)
	report.WorstRenderMillis = maxInt64(renderMillis)

	badCtx, cancelBad := context.WithTimeout(context.Background(), functionalTimeout)
	_, badErr := vault.RenderEvidencePDFPageWithRegisteredPDFiumContext(badCtx, item.ID, 4, adapterWidth)
	cancelBad()
	if badErr == nil {
		return errors.New("out-of-range page 4 unexpectedly rendered")
	}

	directOutput := filepath.Join(workRoot, "direct-page.png")
	benchmarkCtx, cancelBenchmark := context.WithTimeout(context.Background(), functionalTimeout)
	directMillis, peakBytes, benchmarkErr := benchmarkRendererProcess(benchmarkCtx, runtimePath, pdfPath, directOutput)
	cancelBenchmark()
	if benchmarkErr != nil {
		return fmt.Errorf("measure renderer working set: %w", benchmarkErr)
	}
	report.DirectBenchmarkMillis = directMillis
	report.PeakRendererRSSMiB = float64(peakBytes) / (1024 * 1024)

	return nil
}

func applyVerdict(report *qualificationReport) {
	if len(report.Errors) > 0 {
		report.Status = "FAIL"
		return
	}
	cautions := []string{}
	if report.MedianRenderMillis > acceptMedianMillis {
		cautions = append(cautions, fmt.Sprintf("median ECO page render was %d ms (target <= %d ms)", report.MedianRenderMillis, acceptMedianMillis))
	}
	if report.WorstRenderMillis > acceptWorstMillis {
		cautions = append(cautions, fmt.Sprintf("slowest ECO page render was %d ms (target <= %d ms)", report.WorstRenderMillis, acceptWorstMillis))
	}
	if report.PeakRendererRSSMiB > acceptPeakMiB {
		cautions = append(cautions, fmt.Sprintf("renderer peak working set was %.1f MiB (target <= %.0f MiB)", report.PeakRendererRSSMiB, acceptPeakMiB))
	}
	if report.Hardware.MemorySampled && report.Hardware.MemoryTotalBytes < 7*1024*1024*1024 {
		cautions = append(cautions, fmt.Sprintf("Windows reported only %.2f GiB total RAM", float64(report.Hardware.MemoryTotalBytes)/(1024*1024*1024)))
	}
	if len(cautions) > 0 {
		report.Warnings = append(report.Warnings, cautions...)
		report.Status = "CAUTION"
		report.Summary = "All source-integrity and navigation tests passed, but one or more low-spec performance targets were missed. Functional use is proven; low-spec approval remains withheld pending review."
		return
	}
	report.Status = "PASS"
	report.Summary = "All functional, source-integrity, navigation, timing and renderer-memory targets passed on this machine."
}

func benchmarkRendererProcess(ctx context.Context, executable, inputPDF, outputPNG string) (int64, uint64, error) {
	args := []string{"render", "--pages", "2", "--file-type", "png", "--max-width", fmt.Sprint(adapterWidth), "--max-height", "3000", inputPDF, outputPNG}
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = offlineEnvironment(os.Environ())
	cmd.Stdin = nil
	cmd.Stdout = io.Discard
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	started := time.Now()
	if err := cmd.Start(); err != nil {
		return 0, 0, err
	}
	p, procErr := process.NewProcess(int32(cmd.Process.Pid))
	if procErr != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return 0, 0, procErr
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	var peak uint64
	for {
		select {
		case waitErr := <-done:
			elapsed := time.Since(started).Milliseconds()
			if waitErr != nil {
				return elapsed, peak, fmt.Errorf("pdfium-cli benchmark failed: %s", boundText(stderr.String(), 1000))
			}
			if info, err := os.Stat(outputPNG); err != nil || info.Size() <= 0 {
				return elapsed, peak, errors.New("direct renderer benchmark produced no PNG")
			}
			return elapsed, peak, nil
		case <-ticker.C:
			if memory, err := p.MemoryInfo(); err == nil && memory != nil && memory.RSS > peak {
				peak = memory.RSS
			}
		case <-ctx.Done():
			_ = cmd.Process.Kill()
			<-done
			return time.Since(started).Milliseconds(), peak, ctx.Err()
		}
	}
}

func verifyRuntime(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("required bundled renderer is missing: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() != qualifiedRuntimeBytes {
		return fmt.Errorf("bundled renderer size is %d, expected %d", info.Size(), qualifiedRuntimeBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != qualifiedRuntimeSHA256 {
		return fmt.Errorf("bundled renderer SHA-256 mismatch: %s", actual)
	}
	return nil
}

func writeQualificationPDF(path string) error {
	stream := func(text string) string {
		return fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(text), text)
	}
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R 5 0 R 7 0 R] /Count 3 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 9 0 R >> >> /Contents 4 0 R >>",
		stream("BT /F1 24 Tf 72 720 Td (ECO PAGE ONE) Tj ET"),
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 792 612] /Resources << /Font << /F1 9 0 R >> >> /Contents 6 0 R >>",
		stream("BT /F1 24 Tf 72 540 Td (ECO PAGE TWO) Tj ET"),
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 400 800] /Resources << /Font << /F1 9 0 R >> >> /Contents 8 0 R >>",
		stream("BT /F1 24 Tf 50 720 Td (ECO PAGE THREE) Tj ET"),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	out.Write([]byte{'%', 0xe2, 0xe3, 0xcf, 0xd3, '\n'})
	offsets := make([]int, len(objects)+1)
	for index, object := range objects {
		offsets[index+1] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", index+1, object)
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n", len(objects)+1)
	out.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return os.WriteFile(path, out.Bytes(), 0600)
}

func finalizeReport(report *qualificationReport) {
	report.FinishedAt = time.Now().UTC()
	self, _ := os.Executable()
	dir := filepath.Dir(self)
	txtPath := filepath.Join(dir, "ECO_PDF_LOW_SPEC_RESULT.txt")
	jsonPath := filepath.Join(dir, "ECO_PDF_LOW_SPEC_RESULT.json")
	if err := os.WriteFile(txtPath, []byte(textReport(*report)), 0600); err != nil {
		dir = os.TempDir()
		txtPath = filepath.Join(dir, "ECO_PDF_LOW_SPEC_RESULT.txt")
		jsonPath = filepath.Join(dir, "ECO_PDF_LOW_SPEC_RESULT.json")
		_ = os.WriteFile(txtPath, []byte(textReport(*report)), 0600)
	}
	report.ReportPath = txtPath
	if data, err := json.MarshalIndent(report, "", "  "); err == nil {
		_ = os.WriteFile(jsonPath, data, 0600)
	}
	if os.Getenv("ECO_QUALIFIER_CI") == "1" {
		fmt.Printf("ECO_QUALIFIER_STATUS=%s\n", report.Status)
		return
	}
	_ = exec.Command("notepad.exe", txtPath).Start()
	icon := uintptr(0x40) // MB_ICONINFORMATION
	if report.Status == "CAUTION" {
		icon = 0x30 // warning
	}
	if report.Status == "FAIL" {
		icon = 0x10 // error
	}
	showMessageBox("ECO PDF low-spec qualification — "+report.Status, report.Summary+"\r\n\r\nThe plain-English report has been opened in Notepad.\r\n\r\nPlease keep ECO_PDF_LOW_SPEC_RESULT.txt so it can be reviewed.", icon)
}

func textReport(r qualificationReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "ECO PDF LOW-SPEC WINDOWS QUALIFICATION\r\n")
	fmt.Fprintf(&b, "======================================\r\n\r\n")
	fmt.Fprintf(&b, "RESULT: %s\r\n\r\n%s\r\n\r\n", r.Status, r.Summary)
	fmt.Fprintf(&b, "WHAT WAS TESTED\r\n")
	fmt.Fprintf(&b, "- Exact PDFium runtime SHA-256: %s\r\n", r.RuntimeSHA256)
	fmt.Fprintf(&b, "- Real ECO disposable encrypted vault + preserved PDF source\r\n")
	fmt.Fprintf(&b, "- Source-bound page count: expected 3, got %d\r\n", r.PageCount)
	fmt.Fprintf(&b, "- Navigation sequence: 1 -> 2 -> 3 -> 2 -> 1\r\n")
	fmt.Fprintf(&b, "- Out-of-range page 4 rejection\r\n")
	fmt.Fprintf(&b, "- Direct renderer working-set measurement\r\n\r\n")
	fmt.Fprintf(&b, "MACHINE\r\n")
	fmt.Fprintf(&b, "Windows: %s\r\nArchitecture: %s\r\nProcessor: %s\r\nLogical CPUs: %d\r\n", r.Hardware.OS, r.Hardware.Architecture, r.Hardware.Processor, r.Hardware.LogicalCPUs)
	if r.Hardware.MemorySampled {
		fmt.Fprintf(&b, "RAM total: %.2f GiB\r\nRAM available before test: %.2f GiB\r\nRAM used before test: %.1f%%\r\n", float64(r.Hardware.MemoryTotalBytes)/(1024*1024*1024), float64(r.Hardware.MemoryAvailBytes)/(1024*1024*1024), r.Hardware.MemoryUsedPercent)
	}
	fmt.Fprintf(&b, "\r\nMEASURED PERFORMANCE\r\n")
	fmt.Fprintf(&b, "Page-count lookup: %d ms\r\nMedian ECO page render: %d ms\r\nSlowest ECO page render: %d ms\r\nDirect renderer benchmark: %d ms\r\nPeak renderer working set: %.1f MiB\r\n", r.InfoMilliseconds, r.MedianRenderMillis, r.WorstRenderMillis, r.DirectBenchmarkMillis, r.PeakRendererRSSMiB)
	fmt.Fprintf(&b, "\r\nECO LOW-SPEC ACCEPTANCE TARGETS (SET BEFORE THIS TEST)\r\n")
	fmt.Fprintf(&b, "- All functional/source-integrity checks must pass\r\n- Median page render <= %d ms\r\n- Slowest page render <= %d ms\r\n- Peak renderer working set <= %.0f MiB\r\n- No critical ECO host-resource block\r\n", r.Criteria.MedianRenderMillisAtMost, r.Criteria.WorstRenderMillisAtMost, r.Criteria.PeakRendererWorkingSetMiBMax)
	if len(r.PageSamples) > 0 {
		fmt.Fprintf(&b, "\r\nPAGE SAMPLES\r\n")
		for _, sample := range r.PageSamples {
			fmt.Fprintf(&b, "%d. page %d: %d ms, %dx%d, %d PNG bytes\r\n", sample.Ordinal, sample.Page, sample.Milliseconds, sample.Width, sample.Height, sample.PNGBytes)
		}
	}
	if len(r.Warnings) > 0 {
		fmt.Fprintf(&b, "\r\nWARNINGS\r\n")
		for _, warning := range r.Warnings {
			fmt.Fprintf(&b, "- %s\r\n", warning)
		}
	}
	if len(r.Errors) > 0 {
		fmt.Fprintf(&b, "\r\nERRORS\r\n")
		for _, problem := range r.Errors {
			fmt.Fprintf(&b, "- %s\r\n", problem)
		}
	}
	fmt.Fprintf(&b, "\r\nRESULT MEANING\r\n")
	fmt.Fprintf(&b, "PASS = functional and all pre-set low-spec targets passed.\r\n")
	fmt.Fprintf(&b, "CAUTION = functional tests passed, but one or more low-spec performance targets were missed; approval stays withheld pending review.\r\n")
	fmt.Fprintf(&b, "FAIL = a functional/source-integrity test failed or the workload could not complete safely.\r\n")
	fmt.Fprintf(&b, "\r\nQualifier: %s\r\nCombined source commit: %s\r\nStarted UTC: %s\r\nFinished UTC: %s\r\n", r.QualifierVersion, r.CombinedSourceCommit, r.StartedAt.Format(time.RFC3339), r.FinishedAt.Format(time.RFC3339))
	return b.String()
}

func offlineEnvironment(base []string) []string {
	out := make([]string, 0, len(base)+1)
	for _, item := range base {
		key := item
		if index := strings.IndexByte(item, '='); index >= 0 {
			key = item[:index]
		}
		switch strings.ToUpper(strings.TrimSpace(key)) {
		case "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "FTP_PROXY":
			continue
		}
		out = append(out, item)
	}
	return append(out, "NO_PROXY=*")
}

func medianInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	copyValues := append([]int64(nil), values...)
	sort.Slice(copyValues, func(i, j int) bool { return copyValues[i] < copyValues[j] })
	middle := len(copyValues) / 2
	if len(copyValues)%2 == 1 {
		return copyValues[middle]
	}
	return int64(math.Round(float64(copyValues[middle-1]+copyValues[middle]) / 2))
}

func maxInt64(values []int64) int64 {
	var maximum int64
	for _, value := range values {
		if value > maximum {
			maximum = value
		}
	}
	return maximum
}

func windowsVersion() string {
	output, err := exec.Command("cmd.exe", "/c", "ver").Output()
	if err != nil {
		return "Windows version unavailable"
	}
	return strings.TrimSpace(string(output))
}

func boundText(text string, limit int) string {
	text = strings.TrimSpace(text)
	if len(text) <= limit {
		return text
	}
	return text[:limit] + "..."
}

func showMessageBox(title, message string, flags uintptr) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)
	procMessageBox.Call(0, uintptr(unsafe.Pointer(messagePtr)), uintptr(unsafe.Pointer(titlePtr)), flags)
}
