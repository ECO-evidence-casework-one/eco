package eco

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	maxOCRmyPDFDiagnosticBytes = 32 * 1024
	ocrmyPDFTesseractTimeout   = 30
	ocrmyPDFMaxOCRMPixels      = 25
)

type OCRmyPDFResult struct {
	Text          string             `json:"text,omitempty"`
	Segments      []SourceSegment    `json:"segments,omitempty"`
	EngineVersion string             `json:"engine_version"`
	Source        SourceReceipt      `json:"source"`
	PageCount     int                `json:"page_count"`
	OCRPages      int                `json:"ocr_pages"`
	Resources     ResourceAssessment `json:"resources"`
	Warnings      []string           `json:"warnings,omitempty"`
}

// RunOCRmyPDF invokes a caller-selected local OCRmyPDF executable against a
// freshly verified ECO reading copy. This slice is deliberately sidecar-only:
// OCRmyPDF may rasterize temporary page data internally, but ECO requests
// --output-type none and retains no derived PDF.
func RunOCRmyPDF(ctx context.Context, executable, inputPath, language string, source SourceReceipt) (OCRmyPDFResult, error) {
	result := OCRmyPDFResult{Source: source}
	if ctx == nil {
		return result, errors.New("OCRmyPDF context is required")
	}
	if strings.TrimSpace(source.ObjectFile) == "" || !sha256TextPattern.MatchString(source.SHA256) || source.VerifiedAt.IsZero() {
		return result, errors.New("OCRmyPDF requires a verified preserved-object source receipt")
	}
	if !tesseractLanguagePattern.MatchString(language) {
		return result, errors.New("OCRmyPDF Tesseract language selection is missing or invalid")
	}

	exePath, err := requireAbsoluteRegularFile(executable, "OCRmyPDF executable")
	if err != nil {
		return result, err
	}
	verifiedInput, err := requireAbsoluteRegularFile(inputPath, "OCRmyPDF input")
	if err != nil {
		return result, err
	}

	resources, err := CheckLocalEngineResources(ctx, DefaultEngineResourcePolicy("OCRmyPDF", os.TempDir()))
	result.Resources = resources
	if err != nil {
		return result, err
	}
	version, err := ocrmyPDFVersion(ctx, exePath)
	if err != nil {
		return result, err
	}
	result.EngineVersion = version

	workDir, err := os.MkdirTemp("", "eco-ocrmypdf-")
	if err != nil {
		return result, fmt.Errorf("create bounded OCRmyPDF workspace: %w", err)
	}
	defer os.RemoveAll(workDir)

	sidecarPath := filepath.Join(workDir, "sidecar.txt")
	outputPDFPath := filepath.Join(workDir, "must-not-exist.pdf")
	args := ocrmyPDFArgs(verifiedInput, outputPDFPath, sidecarPath, language)
	cmd := exec.CommandContext(ctx, exePath, args...)
	cmd.Env = offlineOCRmyPDFEnvironment(os.Environ())
	cmd.Stdin = nil
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	outCapture := &ocrmyPDFLimitedCapture{buf: &stdout, max: maxOCRmyPDFDiagnosticBytes}
	errCapture := &ocrmyPDFLimitedCapture{buf: &stderr, max: maxOCRmyPDFDiagnosticBytes}
	cmd.Stdout = outCapture
	cmd.Stderr = errCapture
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message == "" {
			message = err.Error()
		}
		return result, fmt.Errorf("OCRmyPDF sidecar extraction failed: %s", message)
	}
	if outCapture.overflow || errCapture.overflow {
		return result, errors.New("OCRmyPDF diagnostic output exceeded ECO's safe size limit")
	}
	if _, err := os.Stat(outputPDFPath); err == nil {
		_ = os.Remove(outputPDFPath)
		return result, errors.New("OCRmyPDF unexpectedly created a derived PDF despite sidecar-only mode")
	} else if !os.IsNotExist(err) {
		return result, fmt.Errorf("check OCRmyPDF no-output contract: %w", err)
	}

	info, err := os.Stat(sidecarPath)
	if err != nil {
		return result, fmt.Errorf("OCRmyPDF did not produce the expected sidecar: %w", err)
	}
	if !info.Mode().IsRegular() {
		return result, errors.New("OCRmyPDF sidecar is not a regular file")
	}
	if info.Size() > maxExtractBytes {
		return result, errors.New("OCRmyPDF sidecar exceeds ECO's extraction size limit")
	}
	data, err := os.ReadFile(sidecarPath)
	if err != nil {
		return result, fmt.Errorf("read OCRmyPDF sidecar: %w", err)
	}
	text, segments, pages, ocrPages, warnings, err := parseOCRmyPDFSidecar(data, source)
	if err != nil {
		return result, err
	}
	result.Text = text
	result.Segments = segments
	result.PageCount = pages
	result.OCRPages = ocrPages
	result.Warnings = warnings
	return result, nil
}

func ocrmyPDFArgs(inputPath, outputPDFPath, sidecarPath, language string) []string {
	return []string{
		"--output-type", "none",
		"--sidecar", sidecarPath,
		"--mode", "skip",
		"--ocr-engine", "tesseract",
		"--rasterizer", "pypdfium",
		"--jobs", "1",
		"--no-progress-bar",
		"--tesseract-timeout", fmt.Sprintf("%d", ocrmyPDFTesseractTimeout),
		"--max-ocr-image-mpixels", fmt.Sprintf("%d", ocrmyPDFMaxOCRMPixels),
		"--language", language,
		inputPath,
		outputPDFPath,
	}
}

func ocrmyPDFVersion(ctx context.Context, executable string) (string, error) {
	cmd := exec.CommandContext(ctx, executable, "--version")
	cmd.Env = offlineOCRmyPDFEnvironment(os.Environ())
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	outCapture := &ocrmyPDFLimitedCapture{buf: &stdout, max: maxOCRmyPDFDiagnosticBytes}
	errCapture := &ocrmyPDFLimitedCapture{buf: &stderr, max: maxOCRmyPDFDiagnosticBytes}
	cmd.Stdout = outCapture
	cmd.Stderr = errCapture
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("OCRmyPDF version check failed: %w", err)
	}
	if outCapture.overflow || errCapture.overflow {
		return "", errors.New("OCRmyPDF version output is unbounded")
	}
	text := strings.TrimSpace(stdout.String())
	if text == "" {
		text = strings.TrimSpace(stderr.String())
	}
	if text == "" {
		return "", errors.New("OCRmyPDF version could not be identified")
	}
	version := strings.TrimSpace(strings.SplitN(text, "\n", 2)[0])
	if len([]rune(version)) > maxOCRIdentityText {
		return "", errors.New("OCRmyPDF version identity is unbounded")
	}
	return version, nil
}

func offlineOCRmyPDFEnvironment(base []string) []string {
	blocked := map[string]bool{
		"HTTP_PROXY": true, "HTTPS_PROXY": true, "ALL_PROXY": true, "FTP_PROXY": true,
		"http_proxy": true, "https_proxy": true, "all_proxy": true, "ftp_proxy": true,
	}
	out := make([]string, 0, len(base)+1)
	for _, item := range base {
		key := item
		if i := strings.IndexByte(item, '='); i >= 0 {
			key = item[:i]
		}
		if blocked[key] || blocked[strings.ToUpper(key)] {
			continue
		}
		out = append(out, item)
	}
	return append(out, "NO_PROXY=*")
}

func parseOCRmyPDFSidecar(data []byte, source SourceReceipt) (string, []SourceSegment, int, int, []string, error) {
	if int64(len(data)) > maxExtractBytes {
		return "", nil, 0, 0, nil, errors.New("OCRmyPDF sidecar exceeds ECO's extraction size limit")
	}
	if !utf8.Valid(data) {
		data = []byte(strings.ToValidUTF8(string(data), "�"))
	}
	pages := strings.Split(string(data), "\f")
	if len(pages) == 0 {
		pages = []string{""}
	}
	segments := []SourceSegment{}
	ocrPages := 0
	ordinal := 1
	pageTexts := make([]string, len(pages))
	for pageIndex, rawPage := range pages {
		pageText := normalizeText(rawPage)
		pageTexts[pageIndex] = pageText
		if pageText == "" {
			continue
		}
		ocrPages++
		pageSegments := segmentText(pageText)
		for _, segment := range pageSegments {
			segment.ID = fmt.Sprintf("SEG-OCRPDF-%04d", ordinal)
			segment.Ordinal = ordinal
			segment.Page = pageIndex + 1
			segment.PageHint = fmt.Sprintf("Page %d", pageIndex+1)
			segment.Origin = "ocrmypdf"
			segment.SourceObject = source.ObjectFile
			segment.SourceSHA256 = source.SHA256
			segments = append(segments, segment)
			ordinal++
		}
	}
	warnings := []string{}
	if len(segments) == 0 {
		warnings = append(warnings, "OCRmyPDF completed but its sidecar contained no OCR text; the preserved PDF remains authoritative.")
	} else {
		warnings = append(warnings, "OCRmyPDF sidecar text is an OCR suggestion and may contain recognition errors; verify wording against the preserved PDF page.")
	}
	if ocrPages < len(pages) {
		warnings = append(warnings, fmt.Sprintf("OCRmyPDF sidecar contained OCR text for %d of %d page positions; pages with existing text, blank content, timeouts or other skips may intentionally have no OCR sidecar text.", ocrPages, len(pages)))
	}
	return strings.Join(pageTexts, "\f"), segments, len(pages), ocrPages, warnings, nil
}

type ocrmyPDFLimitedCapture struct {
	buf      *bytes.Buffer
	max      int
	overflow bool
}

func (w *ocrmyPDFLimitedCapture) Write(p []byte) (int, error) {
	if w.max <= 0 {
		w.overflow = w.overflow || len(p) > 0
		return len(p), nil
	}
	remaining := w.max - w.buf.Len()
	if remaining <= 0 {
		w.overflow = w.overflow || len(p) > 0
		return len(p), nil
	}
	writeN := len(p)
	if writeN > remaining {
		writeN = remaining
		w.overflow = true
	}
	_, _ = w.buf.Write(p[:writeN])
	return len(p), nil
}
