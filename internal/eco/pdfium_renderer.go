package eco

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	qualifiedPDFiumCLIVersion       = "v0.11.2"
	qualifiedPDFiumCLISHA256        = "b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f"
	qualifiedPDFiumCLIBytes   int64 = 16988160
	maxPDFiumDiagnosticBytes        = 64 * 1024
	maxPDFiumInfoBytes              = 8 * 1024 * 1024
	maxPDFiumRenderPNGBytes   int64 = 32 * 1024 * 1024
	maxPDFiumRenderInputBytes int64 = 512 * 1024 * 1024
	maxPDFiumRenderPage             = 100000
	minPDFiumRenderWidth            = 320
	maxPDFiumRenderWidth            = 2000
	maxPDFiumRenderHeight           = 3000
	maxPDFiumRenderPixels           = 6_000_000
)

type PDFPageRender struct {
	EngineVersion string    `json:"engine_version"`
	SourceObject  string    `json:"source_object"`
	SourceSHA256  string    `json:"source_sha256"`
	Page          int       `json:"page"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	PNG           []byte    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
}

type PDFDocumentInfo struct {
	EngineVersion string    `json:"engine_version"`
	SourceObject  string    `json:"source_object"`
	SourceSHA256  string    `json:"source_sha256"`
	PageCount     int       `json:"page_count"`
	CreatedAt     time.Time `json:"created_at"`
}

type pdfiumCLIInfoJSON struct {
	PageCount int `json:"PageCount"`
}

type pdfiumLimitedCapture struct {
	buf      *bytes.Buffer
	max      int
	overflow bool
}

func (w *pdfiumLimitedCapture) Write(p []byte) (int, error) {
	if w.max <= 0 {
		w.overflow = true
		return len(p), nil
	}
	remaining := w.max - w.buf.Len()
	if remaining <= 0 {
		w.overflow = true
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = w.buf.Write(p[:remaining])
		w.overflow = true
		return len(p), nil
	}
	_, _ = w.buf.Write(p)
	return len(p), nil
}

func pdfiumCLIVersion(ctx context.Context, executable string) (string, error) {
	if ctx == nil {
		return "", errors.New("pdfium-cli context is required")
	}
	path, err := requireAbsoluteRegularFile(executable, "pdfium-cli executable")
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() != qualifiedPDFiumCLIBytes {
		return "", fmt.Errorf("pdfium-cli executable size %d does not match ECO's qualified v0.11.2 WebAssembly runtime", info.Size())
	}
	hash, err := hashFile(path)
	if err != nil {
		return "", fmt.Errorf("fingerprint pdfium-cli executable: %w", err)
	}
	if hash != qualifiedPDFiumCLISHA256 {
		return "", errors.New("pdfium-cli SHA-256 does not match ECO's qualified v0.11.2 WebAssembly runtime")
	}
	stdout, stderr, exitErr, runErr := runPDFiumCLICommand(ctx, path, []string{"--help"}, maxPDFiumDiagnosticBytes)
	if runErr != nil {
		return "", runErr
	}
	if exitErr != nil {
		message := strings.TrimSpace(string(stderr))
		if message == "" {
			message = strings.TrimSpace(string(stdout))
		}
		if message == "" {
			message = exitErr.Error()
		}
		return "", fmt.Errorf("pdfium-cli self-check failed: %s", boundPDFiumDiagnostic(message))
	}
	help := strings.ToLower(string(stdout) + "\n" + string(stderr))
	if !strings.Contains(help, "render") || !strings.Contains(help, "info") {
		return "", errors.New("pdfium-cli self-check did not expose the expected render/info commands")
	}
	return qualifiedPDFiumCLIVersion, nil
}

func pdfiumRenderArgs(inputPath, outputPath string, page, maxWidth int) []string {
	return []string{
		"render",
		"--pages", strconv.Itoa(page),
		"--file-type", "png",
		"--max-width", strconv.Itoa(maxWidth),
		"--max-height", strconv.Itoa(maxPDFiumRenderHeight),
		inputPath,
		outputPath,
	}
}

func pdfiumInfoArgs(inputPath string) []string {
	return []string{"info", "--output-type", "json", inputPath, "-"}
}

func parsePDFiumDocumentInfo(data []byte, source SourceReceipt, engineVersion string) (PDFDocumentInfo, error) {
	result := PDFDocumentInfo{
		EngineVersion: engineVersion,
		SourceObject:  source.ObjectFile,
		SourceSHA256:  source.SHA256,
		CreatedAt:     time.Now().UTC(),
	}
	if len(data) == 0 {
		return result, errors.New("pdfium-cli info returned empty JSON")
	}
	if len(data) > maxPDFiumInfoBytes {
		return result, errors.New("pdfium-cli info JSON exceeds ECO's safe size limit")
	}
	var payload pdfiumCLIInfoJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return result, fmt.Errorf("parse pdfium-cli info JSON: %w", err)
	}
	result.PageCount = payload.PageCount
	if err := validatePDFDocumentInfo(result); err != nil {
		return result, err
	}
	return result, nil
}

func RunPDFiumDocumentInfo(ctx context.Context, executable, inputPath string, source SourceReceipt) (PDFDocumentInfo, error) {
	result := PDFDocumentInfo{SourceObject: source.ObjectFile, SourceSHA256: source.SHA256, CreatedAt: time.Now().UTC()}
	if ctx == nil {
		return result, errors.New("pdfium-cli info context is required")
	}
	if strings.TrimSpace(source.ObjectFile) == "" || !sha256TextPattern.MatchString(source.SHA256) || source.VerifiedAt.IsZero() {
		return result, errors.New("pdfium-cli info requires a verified preserved-object source receipt")
	}
	exePath, err := requireAbsoluteRegularFile(executable, "pdfium-cli executable")
	if err != nil {
		return result, err
	}
	verifiedInput, err := requireAbsoluteRegularFile(inputPath, "PDF info input")
	if err != nil {
		return result, err
	}
	inputInfo, err := os.Stat(verifiedInput)
	if err != nil {
		return result, err
	}
	if inputInfo.Size() <= 0 || inputInfo.Size() > maxPDFiumRenderInputBytes {
		return result, fmt.Errorf("PDF info input size %d is outside ECO's safe automatic-preview limit", inputInfo.Size())
	}
	version, err := pdfiumCLIVersion(ctx, exePath)
	if err != nil {
		return result, err
	}
	stdout, stderr, exitErr, runErr := runPDFiumCLICommand(ctx, exePath, pdfiumInfoArgs(verifiedInput), maxPDFiumInfoBytes)
	if runErr != nil {
		return result, runErr
	}
	if exitErr != nil {
		message := strings.TrimSpace(string(stderr))
		if message == "" {
			message = strings.TrimSpace(string(stdout))
		}
		if message == "" {
			message = exitErr.Error()
		}
		return result, fmt.Errorf("pdfium-cli could not inspect PDF page count: %s", boundPDFiumDiagnostic(message))
	}
	return parsePDFiumDocumentInfo(stdout, source, version)
}

func validatePDFDocumentInfo(result PDFDocumentInfo) error {
	if result.EngineVersion != qualifiedPDFiumCLIVersion {
		return errors.New("PDF document info does not identify ECO's qualified pdfium-cli runtime")
	}
	if strings.TrimSpace(result.SourceObject) == "" || !sha256TextPattern.MatchString(result.SourceSHA256) {
		return errors.New("PDF document info is not bound to a preserved source object")
	}
	if result.PageCount < 1 || result.PageCount > maxPDFiumRenderPage {
		return fmt.Errorf("PDF page count %d is outside ECO's safe preview range", result.PageCount)
	}
	if result.CreatedAt.IsZero() {
		return errors.New("PDF document info timestamp is missing")
	}
	return nil
}

func RunPDFiumPageRender(ctx context.Context, executable, inputPath string, page, maxWidth int, source SourceReceipt) (PDFPageRender, error) {
	result := PDFPageRender{SourceObject: source.ObjectFile, SourceSHA256: source.SHA256, Page: page, CreatedAt: time.Now().UTC()}
	if ctx == nil {
		return result, errors.New("pdfium-cli render context is required")
	}
	if strings.TrimSpace(source.ObjectFile) == "" || !sha256TextPattern.MatchString(source.SHA256) || source.VerifiedAt.IsZero() {
		return result, errors.New("pdfium-cli rendering requires a verified preserved-object source receipt")
	}
	if page < 1 || page > maxPDFiumRenderPage {
		return result, fmt.Errorf("PDF page %d is outside ECO's safe automatic-render range", page)
	}
	if maxWidth < minPDFiumRenderWidth || maxWidth > maxPDFiumRenderWidth {
		return result, fmt.Errorf("PDF render width %d is outside ECO's safe range %d..%d", maxWidth, minPDFiumRenderWidth, maxPDFiumRenderWidth)
	}
	exePath, err := requireAbsoluteRegularFile(executable, "pdfium-cli executable")
	if err != nil {
		return result, err
	}
	verifiedInput, err := requireAbsoluteRegularFile(inputPath, "PDF render input")
	if err != nil {
		return result, err
	}
	inputInfo, err := os.Stat(verifiedInput)
	if err != nil {
		return result, err
	}
	if inputInfo.Size() <= 0 || inputInfo.Size() > maxPDFiumRenderInputBytes {
		return result, fmt.Errorf("PDF render input size %d is outside ECO's safe automatic-render limit", inputInfo.Size())
	}
	version, err := pdfiumCLIVersion(ctx, exePath)
	if err != nil {
		return result, err
	}
	result.EngineVersion = version

	tempDir, err := os.MkdirTemp("", "eco-pdfium-render-")
	if err != nil {
		return result, fmt.Errorf("create bounded PDF render workspace: %w", err)
	}
	defer os.RemoveAll(tempDir)
	outputPath := filepath.Join(tempDir, "page.png")

	runCtx := ctx
	cancel := func() {}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		runCtx, cancel = context.WithTimeout(ctx, 45*time.Second)
	}
	defer cancel()
	stdout, stderr, exitErr, runErr := runPDFiumCLICommand(runCtx, exePath, pdfiumRenderArgs(verifiedInput, outputPath, page, maxWidth), maxPDFiumDiagnosticBytes)
	if runErr != nil {
		return result, runErr
	}
	if exitErr != nil {
		message := strings.TrimSpace(string(stderr))
		if message == "" {
			message = strings.TrimSpace(string(stdout))
		}
		if message == "" {
			message = exitErr.Error()
		}
		return result, fmt.Errorf("pdfium-cli could not render page %d: %s", page, boundPDFiumDiagnostic(message))
	}
	if err := runCtx.Err(); err != nil {
		return result, fmt.Errorf("pdfium-cli page render stopped: %w", err)
	}
	outInfo, err := os.Lstat(outputPath)
	if err != nil {
		return result, fmt.Errorf("pdfium-cli did not produce a rendered PNG: %w", err)
	}
	if !outInfo.Mode().IsRegular() || outInfo.Mode()&os.ModeSymlink != 0 || outInfo.Size() <= 0 || outInfo.Size() > maxPDFiumRenderPNGBytes {
		return result, errors.New("pdfium-cli rendered output is missing, non-regular or exceeds ECO's safe PNG size limit")
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return result, fmt.Errorf("read rendered PDF page: %w", err)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return result, fmt.Errorf("validate rendered PDF PNG: %w", err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > maxWidth || cfg.Height > maxPDFiumRenderHeight || int64(cfg.Width)*int64(cfg.Height) > maxPDFiumRenderPixels {
		return result, fmt.Errorf("rendered PDF page dimensions %dx%d exceed ECO's safe preview bounds", cfg.Width, cfg.Height)
	}
	result.Width = cfg.Width
	result.Height = cfg.Height
	result.PNG = append([]byte(nil), data...)
	if err := validatePDFPageRender(result); err != nil {
		return result, err
	}
	return result, nil
}

func runPDFiumCLICommand(ctx context.Context, executable string, args []string, maxStdout int) ([]byte, []byte, error, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = offlinePDFiumEnvironment(os.Environ())
	cmd.Stdin = nil
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	outCapture := &pdfiumLimitedCapture{buf: &stdout, max: maxStdout}
	errCapture := &pdfiumLimitedCapture{buf: &stderr, max: maxPDFiumDiagnosticBytes}
	cmd.Stdout = outCapture
	cmd.Stderr = errCapture
	exitErr := cmd.Run()
	if outCapture.overflow {
		return nil, nil, exitErr, errors.New("pdfium-cli stdout exceeded ECO's safe size limit")
	}
	if errCapture.overflow {
		return nil, nil, exitErr, errors.New("pdfium-cli diagnostics exceeded ECO's safe size limit")
	}
	if ctx.Err() != nil {
		return nil, nil, exitErr, ctx.Err()
	}
	return stdout.Bytes(), stderr.Bytes(), exitErr, nil
}

func offlinePDFiumEnvironment(base []string) []string {
	out := make([]string, 0, len(base)+1)
	for _, item := range base {
		key := item
		if i := strings.IndexByte(item, '='); i >= 0 {
			key = item[:i]
		}
		switch strings.ToUpper(strings.TrimSpace(key)) {
		case "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "FTP_PROXY":
			continue
		}
		out = append(out, item)
	}
	return append(out, "NO_PROXY=*")
}

func boundPDFiumDiagnostic(text string) string {
	text = strings.TrimSpace(text)
	runes := []rune(text)
	if len(runes) > 1024 {
		return string(runes[:1024]) + "…"
	}
	return text
}

func validatePDFPageRender(result PDFPageRender) error {
	if result.EngineVersion != qualifiedPDFiumCLIVersion {
		return errors.New("PDF page render does not identify ECO's qualified pdfium-cli runtime")
	}
	if strings.TrimSpace(result.SourceObject) == "" || !sha256TextPattern.MatchString(result.SourceSHA256) {
		return errors.New("PDF page render is not bound to a preserved source object")
	}
	if result.Page < 1 || result.Page > maxPDFiumRenderPage {
		return errors.New("PDF page render has an invalid page number")
	}
	if result.Width <= 0 || result.Height <= 0 || result.Width > maxPDFiumRenderWidth || result.Height > maxPDFiumRenderHeight || int64(result.Width)*int64(result.Height) > maxPDFiumRenderPixels {
		return errors.New("PDF page render has invalid dimensions")
	}
	if len(result.PNG) == 0 || int64(len(result.PNG)) > maxPDFiumRenderPNGBytes {
		return errors.New("PDF page render has missing or unbounded PNG data")
	}
	if result.CreatedAt.IsZero() {
		return errors.New("PDF page render timestamp is missing")
	}
	return nil
}

type pdfiumInfoRunner func(context.Context, string, string, SourceReceipt) (PDFDocumentInfo, error)

type pdfiumRenderRunner func(context.Context, string, string, int, int, SourceReceipt) (PDFPageRender, error)

func (v *Vault) PDFEvidenceInfoWithRegisteredPDFium(evidenceID string) (PDFDocumentInfo, error) {
	return v.PDFEvidenceInfoWithRegisteredPDFiumContext(context.Background(), evidenceID)
}

func (v *Vault) PDFEvidenceInfoWithRegisteredPDFiumContext(ctx context.Context, evidenceID string) (PDFDocumentInfo, error) {
	tool, err := v.VerifyRegisteredLocalToolContext(ctx, "pdfium-cli")
	if err != nil {
		return PDFDocumentInfo{}, err
	}
	return v.pdfEvidenceInfoWithRunner(ctx, evidenceID, tool.Executable, RunPDFiumDocumentInfo)
}

func (v *Vault) pdfEvidenceInfoWithRunner(ctx context.Context, evidenceID, executable string, runner pdfiumInfoRunner) (PDFDocumentInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(evidenceID) == "" {
		return PDFDocumentInfo{}, errors.New("PDF info evidence ID is required")
	}
	if runner == nil {
		return PDFDocumentInfo{}, errors.New("PDF info runner is required")
	}
	item, record, err := v.pdfiumRenderSource(evidenceID)
	if err != nil {
		return PDFDocumentInfo{}, err
	}
	var result PDFDocumentInfo
	err = v.withVerifiedPreservedFile(record, item.SHA256, func(path string, source SourceReceipt) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		info, runErr := runner(ctx, executable, path, source)
		if runErr != nil {
			return runErr
		}
		result = info
		return nil
	})
	if err != nil {
		return result, err
	}
	if result.SourceObject != item.ObjectFile || result.SourceSHA256 != item.SHA256 {
		return result, errors.New("PDF document info diverges from its verified preserved source")
	}
	if err := validatePDFDocumentInfo(result); err != nil {
		return result, err
	}
	return result, nil
}

func (v *Vault) RenderEvidencePDFPageWithRegisteredPDFium(evidenceID string, page, maxWidth int) (PDFPageRender, error) {
	return v.RenderEvidencePDFPageWithRegisteredPDFiumContext(context.Background(), evidenceID, page, maxWidth)
}

func (v *Vault) RenderEvidencePDFPageWithRegisteredPDFiumContext(ctx context.Context, evidenceID string, page, maxWidth int) (PDFPageRender, error) {
	tool, err := v.VerifyRegisteredLocalToolContext(ctx, "pdfium-cli")
	if err != nil {
		return PDFPageRender{}, err
	}
	return v.renderEvidencePDFPageWithRunner(ctx, evidenceID, tool.Executable, page, maxWidth, RunPDFiumPageRender)
}

func (v *Vault) renderEvidencePDFPageWithRunner(ctx context.Context, evidenceID, executable string, page, maxWidth int, runner pdfiumRenderRunner) (PDFPageRender, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(evidenceID) == "" {
		return PDFPageRender{}, errors.New("PDF render evidence ID is required")
	}
	if runner == nil {
		return PDFPageRender{}, errors.New("PDF render runner is required")
	}
	item, record, err := v.pdfiumRenderSource(evidenceID)
	if err != nil {
		return PDFPageRender{}, err
	}
	var result PDFPageRender
	err = v.withVerifiedPreservedFile(record, item.SHA256, func(path string, source SourceReceipt) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		rendered, runErr := runner(ctx, executable, path, page, maxWidth, source)
		if runErr != nil {
			return runErr
		}
		result = rendered
		return nil
	})
	if err != nil {
		return result, err
	}
	if result.SourceObject != item.ObjectFile || result.SourceSHA256 != item.SHA256 || result.Page != page {
		return result, errors.New("PDF page render diverges from its verified preserved source/page request")
	}
	if err := validatePDFPageRender(result); err != nil {
		return result, err
	}
	return result, nil
}

func (v *Vault) pdfiumRenderSource(evidenceID string) (EvidenceItem, PreservationRecord, error) {
	ws := v.Snapshot()
	var item EvidenceItem
	found := false
	for _, candidate := range ws.Evidence {
		if candidate.ID == evidenceID {
			item = cloneEvidenceItem(candidate)
			found = true
			break
		}
	}
	if !found {
		return EvidenceItem{}, PreservationRecord{}, os.ErrNotExist
	}
	if !preservationUsable(item) {
		return EvidenceItem{}, PreservationRecord{}, errors.New("PDF rendering is blocked because the preserved source is not verified")
	}
	if item.DetectedType != "pdf" {
		return EvidenceItem{}, PreservationRecord{}, fmt.Errorf("PDF rendering requires detected PDF evidence, got %q", item.DetectedType)
	}
	for _, candidate := range ws.Preservations {
		if candidate.EvidenceID == evidenceID && candidate.State == preservationCommitted && candidate.ObjectFile == item.ObjectFile && candidate.PreservedSHA256 == item.SHA256 && candidate.ExpectedSize == item.Size {
			return item, candidate, nil
		}
	}
	return EvidenceItem{}, PreservationRecord{}, errors.New("PDF rendering is blocked because the committed preservation record is missing or inconsistent")
}
