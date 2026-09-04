package eco

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	maxPDFCPUDiagnosticBytes = 32 * 1024
	maxPDFCPUInfoBytes       = 2 * 1024 * 1024
	maxPDFCPUTextRunes       = 4096
)

type PDFAssessment struct {
	EngineVersion          string    `json:"engine_version"`
	SourceObject           string    `json:"source_object"`
	SourceSHA256           string    `json:"source_sha256"`
	CreatedAt              time.Time `json:"created_at"`
	RelaxedValidationPassed bool      `json:"relaxed_validation_passed"`
	RelaxedValidationError string    `json:"relaxed_validation_error,omitempty"`
	StrictValidationPassed bool      `json:"strict_validation_passed"`
	StrictValidationError  string    `json:"strict_validation_error,omitempty"`
	Version                string    `json:"pdf_version,omitempty"`
	PageCount              int       `json:"page_count,omitempty"`
	Title                  string    `json:"title,omitempty"`
	Author                 string    `json:"author,omitempty"`
	Subject                string    `json:"subject,omitempty"`
	Producer               string    `json:"producer,omitempty"`
	Creator                string    `json:"creator,omitempty"`
	CreationDate           string    `json:"creation_date,omitempty"`
	ModificationDate       string    `json:"modification_date,omitempty"`
	Tagged                 bool      `json:"tagged"`
	Hybrid                 bool      `json:"hybrid"`
	Linearized             bool      `json:"linearized"`
	UsingXRefStreams       bool      `json:"using_xref_streams"`
	UsingObjectStreams     bool      `json:"using_object_streams"`
	Watermarked            bool      `json:"watermarked"`
	Thumbnails             bool      `json:"thumbnails"`
	Form                   bool      `json:"form"`
	Signatures             bool      `json:"signatures"`
	AppendOnly             bool      `json:"append_only"`
	Bookmarks              bool      `json:"bookmarks"`
	Names                  bool      `json:"names"`
	Encrypted              bool      `json:"encrypted"`
	Permissions            int       `json:"permissions"`
	AttachmentCount        int       `json:"attachment_count"`
	Warnings               []string  `json:"warnings,omitempty"`
}

type pdfcpuInfoEnvelope struct {
	Infos []pdfcpuInfo `json:"infos"`
}

type pdfcpuInfo struct {
	Source             string            `json:"source"`
	Version            string            `json:"version"`
	PageCount          int               `json:"pageCount"`
	Title              string            `json:"title"`
	Author             string            `json:"author"`
	Subject            string            `json:"subject"`
	Producer           string            `json:"producer"`
	Creator            string            `json:"creator"`
	CreationDate       string            `json:"creationDate"`
	ModificationDate   string            `json:"modificationDate"`
	Tagged             bool              `json:"tagged"`
	Hybrid             bool              `json:"hybrid"`
	Linearized         bool              `json:"linearized"`
	UsingXRefStreams   bool              `json:"usingXRefStreams"`
	UsingObjectStreams bool              `json:"usingObjectStreams"`
	Watermarked        bool              `json:"watermarked"`
	Thumbnails         bool              `json:"thumbnails"`
	Form               bool              `json:"form"`
	Signatures         bool              `json:"signatures"`
	AppendOnly         bool              `json:"appendOnly"`
	Bookmarks          bool              `json:"bookmarks"`
	Names              bool              `json:"names"`
	Encrypted          bool              `json:"encrypted"`
	Permissions        int               `json:"permissions"`
	Attachments        []json.RawMessage `json:"attachments"`
}

// RunPDFCPUInspection invokes a caller-selected local pdfcpu binary against a
// freshly verified ECO PDF reading copy. It is read-only: only `version`,
// `validate`, and `info --json` are used. `--offline` and `--conf disable`
// are passed on every invocation.
func RunPDFCPUInspection(ctx context.Context, executable, inputPath string, source SourceReceipt) (PDFAssessment, error) {
	assessment := PDFAssessment{SourceObject: source.ObjectFile, SourceSHA256: source.SHA256, CreatedAt: time.Now().UTC()}
	if ctx == nil {
		return assessment, errors.New("pdfcpu context is required")
	}
	if strings.TrimSpace(source.ObjectFile) == "" || !sha256TextPattern.MatchString(source.SHA256) || source.VerifiedAt.IsZero() {
		return assessment, errors.New("pdfcpu requires a verified preserved-object source receipt")
	}
	exePath, err := requireAbsoluteRegularFile(executable, "pdfcpu executable")
	if err != nil {
		return assessment, err
	}
	verifiedInput, err := requireAbsoluteRegularFile(inputPath, "pdfcpu input")
	if err != nil {
		return assessment, err
	}
	version, err := pdfcpuVersion(ctx, exePath)
	if err != nil {
		return assessment, err
	}
	assessment.EngineVersion = version

	relaxedErr, err := runPDFCPUValidation(ctx, exePath, verifiedInput, "relaxed")
	if err != nil {
		return assessment, err
	}
	assessment.RelaxedValidationPassed = relaxedErr == ""
	assessment.RelaxedValidationError = relaxedErr
	if !assessment.RelaxedValidationPassed {
		assessment.Warnings = append(assessment.Warnings, "pdfcpu relaxed validation did not pass. The original remains preserved; this result alone does not distinguish structural damage from unsupported/password-protected input.")
		return assessment, nil
	}

	strictErr, err := runPDFCPUValidation(ctx, exePath, verifiedInput, "strict")
	if err != nil {
		return assessment, err
	}
	assessment.StrictValidationPassed = strictErr == ""
	assessment.StrictValidationError = strictErr
	if !assessment.StrictValidationPassed {
		assessment.Warnings = append(assessment.Warnings, "PDF passed pdfcpu relaxed validation but not strict validation. Treat the strict diagnostics as a compatibility/structure warning, not proof that the document's contents are false or unsafe.")
	}

	info, err := runPDFCPUInfo(ctx, exePath, verifiedInput)
	if err != nil {
		return assessment, err
	}
	assessment.applyInfo(info)
	return assessment, nil
}

func pdfcpuBaseArgs() []string {
	return []string{"--offline", "--conf", "disable"}
}

func pdfcpuValidationArgs(inputPath, mode string) []string {
	args := append([]string{}, pdfcpuBaseArgs()...)
	return append(args, "validate", "--mode", mode, inputPath)
}

func pdfcpuInfoArgs(inputPath string) []string {
	args := append([]string{}, pdfcpuBaseArgs()...)
	return append(args, "info", "--json", inputPath)
}

func pdfcpuVersionArgs() []string {
	args := append([]string{}, pdfcpuBaseArgs()...)
	return append(args, "version")
}

func pdfcpuVersion(ctx context.Context, executable string) (string, error) {
	stdout, stderr, exitErr, err := runPDFCPUCommand(ctx, executable, pdfcpuVersionArgs(), maxPDFCPUDiagnosticBytes)
	if err != nil {
		return "", err
	}
	if exitErr != nil {
		message := strings.TrimSpace(string(stderr))
		if message == "" {
			message = exitErr.Error()
		}
		return "", fmt.Errorf("pdfcpu version check failed: %s", boundPDFCPUText(message))
	}
	text := strings.TrimSpace(string(stdout))
	if text == "" {
		text = strings.TrimSpace(string(stderr))
	}
	if text == "" {
		return "", errors.New("pdfcpu version could not be identified")
	}
	line := strings.TrimSpace(strings.SplitN(text, "\n", 2)[0])
	if len([]rune(line)) > maxOCRIdentityText {
		return "", errors.New("pdfcpu version identity is unbounded")
	}
	return line, nil
}

func runPDFCPUValidation(ctx context.Context, executable, inputPath, mode string) (string, error) {
	if mode != "relaxed" && mode != "strict" {
		return "", errors.New("unsupported pdfcpu validation mode")
	}
	stdout, stderr, exitErr, err := runPDFCPUCommand(ctx, executable, pdfcpuValidationArgs(inputPath, mode), maxPDFCPUDiagnosticBytes)
	if err != nil {
		return "", err
	}
	if exitErr == nil {
		return "", nil
	}
	message := strings.TrimSpace(string(stderr))
	if message == "" {
		message = strings.TrimSpace(string(stdout))
	}
	if message == "" {
		message = exitErr.Error()
	}
	return boundPDFCPUText(message), nil
}

func runPDFCPUInfo(ctx context.Context, executable, inputPath string) (pdfcpuInfo, error) {
	stdout, stderr, exitErr, err := runPDFCPUCommand(ctx, executable, pdfcpuInfoArgs(inputPath), maxPDFCPUInfoBytes)
	if err != nil {
		return pdfcpuInfo{}, err
	}
	if exitErr != nil {
		message := strings.TrimSpace(string(stderr))
		if message == "" {
			message = exitErr.Error()
		}
		return pdfcpuInfo{}, fmt.Errorf("pdfcpu info failed after relaxed validation: %s", boundPDFCPUText(message))
	}
	return parsePDFCPUInfo(stdout)
}

func parsePDFCPUInfo(data []byte) (pdfcpuInfo, error) {
	if len(data) == 0 {
		return pdfcpuInfo{}, errors.New("pdfcpu info returned empty JSON")
	}
	if len(data) > maxPDFCPUInfoBytes {
		return pdfcpuInfo{}, errors.New("pdfcpu info JSON exceeds ECO's safe size limit")
	}
	var envelope pdfcpuInfoEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return pdfcpuInfo{}, fmt.Errorf("parse pdfcpu info JSON: %w", err)
	}
	if len(envelope.Infos) != 1 {
		return pdfcpuInfo{}, fmt.Errorf("pdfcpu info expected one PDF record, got %d", len(envelope.Infos))
	}
	info := envelope.Infos[0]
	if strings.TrimSpace(info.Version) == "" || info.PageCount < 0 {
		return pdfcpuInfo{}, errors.New("pdfcpu info JSON is missing required PDF identity fields")
	}
	return info, nil
}

func runPDFCPUCommand(ctx context.Context, executable string, args []string, maxStdout int) ([]byte, []byte, error, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = offlinePDFCPUEnvironment(os.Environ())
	cmd.Stdin = nil
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	outCapture := &pdfcpuLimitedCapture{buf: &stdout, max: maxStdout}
	errCapture := &pdfcpuLimitedCapture{buf: &stderr, max: maxPDFCPUDiagnosticBytes}
	cmd.Stdout = outCapture
	cmd.Stderr = errCapture
	exitErr := cmd.Run()
	if outCapture.overflow {
		return nil, nil, exitErr, errors.New("pdfcpu stdout exceeded ECO's safe size limit")
	}
	if errCapture.overflow {
		return nil, nil, exitErr, errors.New("pdfcpu diagnostics exceeded ECO's safe size limit")
	}
	if ctx.Err() != nil {
		return nil, nil, exitErr, ctx.Err()
	}
	return stdout.Bytes(), stderr.Bytes(), exitErr, nil
}

func offlinePDFCPUEnvironment(base []string) []string {
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

func (a *PDFAssessment) applyInfo(info pdfcpuInfo) {
	a.Version = boundPDFCPUText(info.Version)
	a.PageCount = info.PageCount
	a.Title = boundPDFCPUText(info.Title)
	a.Author = boundPDFCPUText(info.Author)
	a.Subject = boundPDFCPUText(info.Subject)
	a.Producer = boundPDFCPUText(info.Producer)
	a.Creator = boundPDFCPUText(info.Creator)
	a.CreationDate = boundPDFCPUText(info.CreationDate)
	a.ModificationDate = boundPDFCPUText(info.ModificationDate)
	a.Tagged = info.Tagged
	a.Hybrid = info.Hybrid
	a.Linearized = info.Linearized
	a.UsingXRefStreams = info.UsingXRefStreams
	a.UsingObjectStreams = info.UsingObjectStreams
	a.Watermarked = info.Watermarked
	a.Thumbnails = info.Thumbnails
	a.Form = info.Form
	a.Signatures = info.Signatures
	a.AppendOnly = info.AppendOnly
	a.Bookmarks = info.Bookmarks
	a.Names = info.Names
	a.Encrypted = info.Encrypted
	a.Permissions = info.Permissions
	a.AttachmentCount = len(info.Attachments)
}

func boundPDFCPUText(s string) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= maxPDFCPUTextRunes {
		return s
	}
	return strings.TrimSpace(string(r[:maxPDFCPUTextRunes]))
}

type pdfcpuLimitedCapture struct {
	buf      *bytes.Buffer
	max      int
	overflow bool
}

func (w *pdfcpuLimitedCapture) Write(p []byte) (int, error) {
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
