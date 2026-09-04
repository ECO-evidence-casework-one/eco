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
)

const maxDoclingDiagnosticBytes = 16 * 1024

type DoclingExtractionResult struct {
	Text          string
	Segments      []SourceSegment
	EngineVersion string
	Source        SourceReceipt
	Warnings      []string
}

// RunDoclingExtraction invokes a caller-selected local Docling CLI against one
// freshly verified ECO reading copy. It requires an explicit local artifacts
// directory and forces the common Hugging Face/Transformers clients offline.
// ECO never enables Docling remote services in this adapter.
func RunDoclingExtraction(ctx context.Context, executable, inputPath, artifactsPath string, source SourceReceipt) (DoclingExtractionResult, error) {
	if ctx == nil {
		return DoclingExtractionResult{}, errors.New("Docling context is required")
	}
	if strings.TrimSpace(source.ObjectFile) == "" || !sha256TextPattern.MatchString(source.SHA256) || source.VerifiedAt.IsZero() {
		return DoclingExtractionResult{}, errors.New("Docling requires a verified preserved-object source receipt")
	}

	exePath, err := requireAbsoluteRegularFile(executable, "Docling executable")
	if err != nil {
		return DoclingExtractionResult{}, err
	}
	verifiedInput, err := requireAbsoluteRegularFile(inputPath, "Docling input")
	if err != nil {
		return DoclingExtractionResult{}, err
	}
	artifactsDir, err := requireAbsoluteDirectory(artifactsPath, "Docling artifacts")
	if err != nil {
		return DoclingExtractionResult{}, err
	}
	version, err := doclingVersion(ctx, exePath)
	if err != nil {
		return DoclingExtractionResult{}, err
	}

	outputDir, err := os.MkdirTemp("", "eco-docling-")
	if err != nil {
		return DoclingExtractionResult{}, fmt.Errorf("create bounded Docling workspace: %w", err)
	}
	defer os.RemoveAll(outputDir)

	args := doclingExtractionArgs(verifiedInput, outputDir, artifactsDir)
	cmd := exec.CommandContext(ctx, exePath, args...)
	cmd.Env = offlineDoclingEnvironment(os.Environ(), artifactsDir)
	var stderr bytes.Buffer
	cmd.Stderr = &boundedBuffer{buf: &stderr, max: maxDoclingDiagnosticBytes}
	cmd.Stdout = nil
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return DoclingExtractionResult{}, fmt.Errorf("Docling extraction failed: %s", message)
	}

	stem := strings.TrimSuffix(filepath.Base(verifiedInput), filepath.Ext(verifiedInput))
	markdownPath := filepath.Join(outputDir, stem+".md")
	info, err := os.Stat(markdownPath)
	if err != nil {
		return DoclingExtractionResult{}, fmt.Errorf("Docling did not produce expected Markdown output: %w", err)
	}
	if !info.Mode().IsRegular() {
		return DoclingExtractionResult{}, errors.New("Docling Markdown output is not a regular file")
	}
	if info.Size() > maxExtractBytes {
		return DoclingExtractionResult{}, errors.New("Docling Markdown output exceeds ECO's extraction size limit")
	}
	data, err := os.ReadFile(markdownPath)
	if err != nil {
		return DoclingExtractionResult{}, fmt.Errorf("read Docling Markdown output: %w", err)
	}
	text := normalizeText(string(data))
	segments := segmentText(text)
	for i := range segments {
		segments[i].Origin = "docling"
		segments[i].SourceObject = source.ObjectFile
		segments[i].SourceSHA256 = source.SHA256
	}
	warnings := []string{}
	if strings.TrimSpace(text) == "" {
		warnings = append(warnings, "Docling completed but returned no readable text; the original preserved evidence remains authoritative.")
	}
	return DoclingExtractionResult{Text: text, Segments: segments, EngineVersion: version, Source: source, Warnings: warnings}, nil
}

func doclingExtractionArgs(inputPath, outputDir, artifactsDir string) []string {
	return []string{
		inputPath,
		"--to", "md",
		"--output", outputDir,
		"--artifacts-path", artifactsDir,
		"--pipeline", "standard",
		"--device", "cpu",
	}
}

func doclingVersion(ctx context.Context, executable string) (string, error) {
	cmd := exec.CommandContext(ctx, executable, "--version")
	cmd.Env = offlineDoclingEnvironment(os.Environ(), "")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Docling version check failed: %w", err)
	}
	if len(output) > maxDoclingDiagnosticBytes {
		return "", errors.New("Docling version output is unbounded")
	}
	version := strings.TrimSpace(strings.SplitN(string(output), "\n", 2)[0])
	if version == "" {
		return "", errors.New("Docling version could not be identified")
	}
	if len([]rune(version)) > maxOCRIdentityText {
		return "", errors.New("Docling version identity is unbounded")
	}
	return version, nil
}

func requireAbsoluteDirectory(path, label string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("%s path is required", label)
	}
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("%s must use an absolute path", label)
	}
	resolved, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", label, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", label, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", label)
	}
	return resolved, nil
}

func offlineDoclingEnvironment(base []string, artifactsDir string) []string {
	env := append([]string(nil), base...)
	env = setEnvironmentValue(env, "HF_HUB_OFFLINE", "1")
	env = setEnvironmentValue(env, "TRANSFORMERS_OFFLINE", "1")
	env = setEnvironmentValue(env, "DOCLING_ENABLE_REMOTE_SERVICES", "false")
	if artifactsDir != "" {
		env = setEnvironmentValue(env, "DOCLING_ARTIFACTS_PATH", artifactsDir)
	}
	return env
}

func setEnvironmentValue(env []string, key, value string) []string {
	prefix := strings.ToUpper(key) + "="
	out := make([]string, 0, len(env)+1)
	for _, item := range env {
		if strings.HasPrefix(strings.ToUpper(item), prefix) {
			continue
		}
		out = append(out, item)
	}
	return append(out, key+"="+value)
}

type boundedBuffer struct {
	buf *bytes.Buffer
	max int
}

func (w *boundedBuffer) Write(p []byte) (int, error) {
	if w.max <= 0 {
		return len(p), nil
	}
	remaining := w.max - w.buf.Len()
	if remaining > 0 {
		if len(p) < remaining {
			remaining = len(p)
		}
		_, _ = w.buf.Write(p[:remaining])
	}
	return len(p), nil
}
