package eco

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const maxTesseractTSVBytes = int64(64 * 1024 * 1024)

var tesseractLanguagePattern = regexp.MustCompile(`^[A-Za-z0-9_+-]{1,100}$`)

// RunTesseractOCR is the controlled glue between ECO's preserved-evidence
// model and a caller-selected local Tesseract executable. It performs no
// download and no network access. The caller must provide the exact local
// executable and a verified source receipt.
func RunTesseractOCR(ctx context.Context, executable, imagePath, language string, source SourceReceipt, imageWidth, imageHeight int) (OCRReceipt, []SourceSegment, error) {
	return runTesseractOCR(ctx, executable, imagePath, language, "", source, imageWidth, imageHeight)
}

func RunTesseractOCRWithTessdata(ctx context.Context, executable, imagePath, language, tessdataDir string, source SourceReceipt, imageWidth, imageHeight int) (OCRReceipt, []SourceSegment, error) {
	return runTesseractOCR(ctx, executable, imagePath, language, tessdataDir, source, imageWidth, imageHeight)
}

func runTesseractOCR(ctx context.Context, executable, imagePath, language, tessdataDir string, source SourceReceipt, imageWidth, imageHeight int) (OCRReceipt, []SourceSegment, error) {
	if ctx == nil {
		return OCRReceipt{}, nil, errors.New("OCR context is required")
	}
	if imageWidth <= 0 || imageHeight <= 0 {
		return OCRReceipt{}, nil, errors.New("OCR image dimensions must be positive")
	}
	if strings.TrimSpace(source.ObjectFile) == "" || !sha256TextPattern.MatchString(source.SHA256) || source.VerifiedAt.IsZero() {
		return OCRReceipt{}, nil, errors.New("OCR requires a verified preserved-object source receipt")
	}
	if !tesseractLanguagePattern.MatchString(language) {
		return OCRReceipt{}, nil, errors.New("Tesseract language selection is missing or invalid")
	}

	exePath, err := requireAbsoluteRegularFile(executable, "Tesseract executable")
	if err != nil {
		return OCRReceipt{}, nil, err
	}
	inputPath, err := requireAbsoluteRegularFile(imagePath, "OCR input image")
	if err != nil {
		return OCRReceipt{}, nil, err
	}

	version, err := tesseractVersion(ctx, exePath)
	if err != nil {
		return OCRReceipt{}, nil, err
	}

	tempDir, err := os.MkdirTemp("", "eco-tesseract-")
	if err != nil {
		return OCRReceipt{}, nil, fmt.Errorf("create bounded OCR workspace: %w", err)
	}
	defer os.RemoveAll(tempDir)

	outputBase := filepath.Join(tempDir, "ocr")
	args := tesseractOCRArgsWithTessdata(inputPath, outputBase, language, tessdataDir)
	cmd := exec.CommandContext(ctx, exePath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = nil
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if len(message) > 4096 {
			message = message[:4096]
		}
		if message == "" {
			message = err.Error()
		}
		return OCRReceipt{}, nil, fmt.Errorf("Tesseract OCR failed: %s", message)
	}

	tsvPath := outputBase + ".tsv"
	info, err := os.Stat(tsvPath)
	if err != nil {
		return OCRReceipt{}, nil, fmt.Errorf("Tesseract did not produce TSV output: %w", err)
	}
	if !info.Mode().IsRegular() {
		return OCRReceipt{}, nil, errors.New("Tesseract TSV output is not a regular file")
	}
	if info.Size() > maxTesseractTSVBytes {
		return OCRReceipt{}, nil, errors.New("Tesseract TSV output exceeds ECO's safe size limit")
	}
	data, err := os.ReadFile(tsvPath)
	if err != nil {
		return OCRReceipt{}, nil, fmt.Errorf("read Tesseract TSV output: %w", err)
	}

	return ParseOCRTSV(string(data), "tesseract", version, language, source, imageWidth, imageHeight)
}

func tesseractOCRArgs(inputPath, outputBase, language string) []string {
	return tesseractOCRArgsWithTessdata(inputPath, outputBase, language, "")
}

func tesseractOCRArgsWithTessdata(inputPath, outputBase, language, tessdataDir string) []string {
	args := []string{inputPath, outputBase}
	if strings.TrimSpace(tessdataDir) != "" {
		args = append(args, "--tessdata-dir", tessdataDir)
	}
	return append(args, "-l", language, "tsv")
}

func tesseractVersion(ctx context.Context, executable string) (string, error) {
	cmd := exec.CommandContext(ctx, executable, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Tesseract version check failed: %w", err)
	}
	if len(output) > 16*1024 {
		return "", errors.New("Tesseract version output is unbounded")
	}
	first := strings.TrimSpace(strings.SplitN(string(output), "\n", 2)[0])
	first = strings.TrimSpace(strings.TrimPrefix(first, "tesseract"))
	if first == "" {
		return "", errors.New("Tesseract version could not be identified")
	}
	if len([]rune(first)) > maxOCRIdentityText {
		return "", errors.New("Tesseract version identity is unbounded")
	}
	return first, nil
}

func requireAbsoluteRegularFile(path, label string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("%s path is required", label)
	}
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("%s must use an absolute path", label)
	}
	clean := filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(clean)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", label, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", label, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", label)
	}
	return resolved, nil
}
