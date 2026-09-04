from pathlib import Path

ROOT = Path('.')

renderer_go = r'''package eco

import (
	"bytes"
	"context"
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
	qualifiedPDFiumCLIVersion      = "v0.11.2"
	qualifiedPDFiumCLISHA256       = "b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f"
	qualifiedPDFiumCLIBytes  int64 = 16988160
	maxPDFiumDiagnosticBytes       = 64 * 1024
	maxPDFiumRenderPNGBytes  int64 = 32 * 1024 * 1024
	maxPDFiumRenderInputBytes int64 = 512 * 1024 * 1024
	maxPDFiumRenderPage            = 100000
	minPDFiumRenderWidth           = 320
	maxPDFiumRenderWidth           = 2000
	maxPDFiumRenderHeight          = 3000
	maxPDFiumRenderPixels          = 6_000_000
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

type pdfiumRenderRunner func(context.Context, string, string, int, int, SourceReceipt) (PDFPageRender, error)

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
'''
(ROOT / 'internal/eco/pdfium_renderer.go').write_text(renderer_go, encoding='utf-8')

renderer_test = r'''package eco

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPDFiumRenderArgsAreBoundedAndPageSpecific(t *testing.T) {
	got := pdfiumRenderArgs(`C:\source.pdf`, `C:\page.png`, 7, 1600)
	want := []string{"render", "--pages", "7", "--file-type", "png", "--max-width", "1600", "--max-height", "3000", `C:\source.pdf`, `C:\page.png`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected pdfium-cli render args: %#v", got)
	}
}

func TestPDFiumVersionRejectsUnqualifiedExecutableBeforeRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pdfium.exe")
	if err := os.WriteFile(path, []byte("not the qualified renderer"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := pdfiumCLIVersion(context.Background(), path)
	if err == nil || !strings.Contains(err.Error(), "does not match ECO's qualified") {
		t.Fatalf("unexpected qualification result: %v", err)
	}
}

func TestValidatePDFPageRenderRequiresSourceAndBounds(t *testing.T) {
	base := PDFPageRender{
		EngineVersion: qualifiedPDFiumCLIVersion,
		SourceObject:  "object-123.eco",
		SourceSHA256:  strings.Repeat("a", 64),
		Page:          2,
		Width:         1200,
		Height:        1553,
		PNG:           []byte{1, 2, 3},
		CreatedAt:     time.Now().UTC(),
	}
	if err := validatePDFPageRender(base); err != nil {
		t.Fatal(err)
	}
	bad := base
	bad.Page = 0
	if err := validatePDFPageRender(bad); err == nil {
		t.Fatal("invalid page number was accepted")
	}
	bad = base
	bad.Width = maxPDFiumRenderWidth + 1
	if err := validatePDFPageRender(bad); err == nil {
		t.Fatal("unbounded render width was accepted")
	}
}
'''
(ROOT / 'internal/eco/pdfium_renderer_test.go').write_text(renderer_test, encoding='utf-8')

# Extend the existing verified local-runtime registry.
p = ROOT / 'internal/eco/local_tools.go'
s = p.read_text(encoding='utf-8')
old = '\t"pdfcpu":    {Kind: "pdfcpu", Upstream: "pdfcpu/pdfcpu", License: "Apache-2.0"},\n'
if old not in s:
    raise SystemExit('local tool spec anchor not found')
s = s.replace(old, old + '\t"pdfium-cli": {Kind: "pdfium-cli", Upstream: "klippa-app/pdfium-cli", License: "MIT"},\n', 1)
s = s.replace('var localToolOrder = []string{"tesseract", "docling", "ocrmypdf", "llama.cpp", "pdfcpu"}', 'var localToolOrder = []string{"tesseract", "docling", "ocrmypdf", "llama.cpp", "pdfcpu", "pdfium-cli"}', 1)
anchor = '\tcase "pdfcpu":\n\t\tkind = "pdfcpu"\n\tdefault:\n'
if anchor not in s:
    raise SystemExit('canonical local tool anchor not found')
s = s.replace(anchor, '\tcase "pdfcpu":\n\t\tkind = "pdfcpu"\n\tcase "pdfium-cli", "pdfium", "pdfium-renderer":\n\t\tkind = "pdfium-cli"\n\tdefault:\n', 1)
anchor = '\tcase "pdfcpu":\n\t\treturn pdfcpuVersion(ctx, executable)\n\tdefault:\n'
if anchor not in s:
    raise SystemExit('local tool version anchor not found')
s = s.replace(anchor, '\tcase "pdfcpu":\n\t\treturn pdfcpuVersion(ctx, executable)\n\tcase "pdfium-cli":\n\t\treturn pdfiumCLIVersion(ctx, executable)\n\tdefault:\n', 1)
p.write_text(s, encoding='utf-8')

# Add Windows UI registration and PDF preview plumbing.
p = ROOT / 'cmd/eco/main_windows.go'
s = p.read_text(encoding='utf-8')
if '\t"context"\n' not in s:
    s = s.replace('import (\n\t"encoding/binary"', 'import (\n\t"context"\n\t"encoding/binary"', 1)
s = s.replace('\tpendingCitationRegion                                                     *eco.NormalizedRegion\n}', '\tpendingCitationRegion                                                     *eco.NormalizedRegion\n\tpendingCitationPage                                                       int\n}', 1)
s = s.replace('type previewState struct {\n\titemID        string', 'type previewState struct {\n\titemID        string\n\tisPDF         bool\n\tpdfPage       int', 1)

s = s.replace('\tbuttonTop := rc.Bottom - 108\n', '\tbuttonTop := rc.Bottom - 160\n', 1)
trust_anchor = '\ta.drawButton(hdc, "restore", "Restore encrypted backup", RECT{x + w2 + gapB, buttonTop + 52, right, buttonTop + 94}, false)\n'
if trust_anchor not in s:
    raise SystemExit('trust button anchor not found')
trust_insert = trust_anchor + '''\tpdfRendererLabel := "Locate verified PDF page renderer"\n\tif reg, err := a.vault.RegisteredLocalTool("pdfium-cli"); err == nil {\n\t\tpdfRendererLabel = "PDF renderer " + reg.Version + " — verified runtime registered"\n\t}\n\ta.drawButton(hdc, "pdfRenderer", pdfRendererLabel, RECT{x, buttonTop + 104, right, buttonTop + 146}, false)\n'''
s = s.replace(trust_anchor, trust_insert, 1)

click_anchor = '\t\t\tcase "openVault":\n\t\t\t\texec.Command("explorer.exe", a.vault.Root).Start()\n\t\t\tcase "lowSensory":\n'
if click_anchor not in s:
    raise SystemExit('click handler anchor not found')
s = s.replace(click_anchor, '\t\t\tcase "openVault":\n\t\t\t\texec.Command("explorer.exe", a.vault.Root).Start()\n\t\t\tcase "pdfRenderer":\n\t\t\t\ta.locatePDFRenderer()\n\t\t\tcase "lowSensory":\n', 1)

choose_anchor = '''func (a *application) chooseFiles() {\n\tpaths, err := openFileDialog(a.hwnd)\n\tif err != nil {\n\t\tmessageBox(a.hwnd, "Could not open file picker", err.Error(), MB_OK|MB_ICONERROR)\n\t\treturn\n\t}\n\tif len(paths) > 0 {\n\t\ta.beginImport(paths)\n\t}\n}\n\n'''
if choose_anchor not in s:
    raise SystemExit('chooseFiles anchor not found')
locate_func = choose_anchor + r'''func (a *application) locatePDFRenderer() {
	path := openExecutableDialog(a.hwnd, "Locate the verified pdfium-cli v0.11.2 WebAssembly renderer")
	if path == "" {
		return
	}
	a.mu.Lock()
	if a.importing {
		a.mu.Unlock()
		messageBox(a.hwnd, "ECO is already processing", "Wait for the current local task to finish.", MB_OK|MB_ICONINFORMATION)
		return
	}
	a.importing = true
	a.progress = eco.ImportProgress{Name: filepath.Base(path), Stage: "Verifying PDF renderer SHA-256 and self-check"}
	a.mu.Unlock()
	invalidate(a.hwnd)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_, err := a.vault.RegisterLocalToolContext(ctx, "pdfium-cli", path)
		if err != nil {
			a.mu.Lock()
			a.lastErr = "PDF renderer registration failed: " + err.Error()
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportError, 0, 0)
			return
		}
		a.mu.Lock()
		a.importing = false
		a.progress = eco.ImportProgress{}
		a.mu.Unlock()
		procPostMessageW.Call(a.hwnd, msgRefresh, 0, 0)
	}()
}

'''
s = s.replace(choose_anchor, locate_func, 1)

citation_anchor = '''\t\t\ta.mu.Lock()\n\t\t\tif c.Region != nil {\n\t\t\t\tregion := *c.Region\n\t\t\t\ta.pendingCitationRegion = &region\n\t\t\t} else {\n\t\t\t\ta.pendingCitationRegion = nil\n\t\t\t}\n\t\t\ta.mu.Unlock()\n'''
if citation_anchor not in s:
    raise SystemExit('citation pending-region anchor not found')
citation_repl = '''\t\t\ta.mu.Lock()\n\t\t\tif c.Region != nil {\n\t\t\t\tregion := *c.Region\n\t\t\t\ta.pendingCitationRegion = &region\n\t\t\t} else {\n\t\t\t\ta.pendingCitationRegion = nil\n\t\t\t}\n\t\t\ta.pendingCitationPage = c.Page\n\t\t\ta.mu.Unlock()\n'''
s = s.replace(citation_anchor, citation_repl, 1)

preview_anchor = '\te := a.view.Evidence[a.selected]\n\tif e.Image != nil {\n'
if preview_anchor not in s:
    raise SystemExit('preview PDF insertion anchor not found')
pdf_preview = r'''	e := a.view.Evidence[a.selected]
	if e.DetectedType == "pdf" {
		a.mu.Lock()
		if a.importing {
			a.mu.Unlock()
			messageBox(a.hwnd, "ECO is already processing", "Wait for the current local task to finish before rendering a PDF page.", MB_OK|MB_ICONINFORMATION)
			return
		}
		page := a.pendingCitationPage
		if page < 1 {
			page = 1
		}
		a.pendingCitationPage = 0
		var highlight *eco.NormalizedRegion
		if a.pendingCitationRegion != nil {
			region := *a.pendingCitationRegion
			highlight = &region
			a.pendingCitationRegion = nil
		}
		a.mu.Unlock()
		if _, err := a.vault.RegisteredLocalTool("pdfium-cli"); err != nil {
			if e.ExtractedText != "" {
				text := e.ExtractedText
				if len([]rune(text)) > 1400 {
					text = string([]rune(text)[:1400]) + "\r\n\r\n[Text preview bounded here.]"
				}
				messageBox(a.hwnd, "PDF visual renderer not registered — readable text available", "To enable visual PDF pages, open Trust & Settings and choose ‘Locate verified PDF page renderer’.\r\n\r\n"+text, MB_OK|MB_ICONINFORMATION)
			} else {
				messageBox(a.hwnd, "PDF visual renderer not registered", "Open Trust & Settings and choose ‘Locate verified PDF page renderer’, then select the exact qualified pdfium-cli v0.11.2 WebAssembly executable.", MB_OK|MB_ICONINFORMATION)
			}
			return
		}
		a.mu.Lock()
		a.importing = true
		a.progress = eco.ImportProgress{Name: e.SafeName, Stage: fmt.Sprintf("Rendering verified PDF page %d locally", page)}
		a.mu.Unlock()
		invalidate(a.hwnd)
		go func(item eco.EvidenceItem, page int, highlight *eco.NormalizedRegion) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			rendered, err := a.vault.RenderEvidencePDFPageWithRegisteredPDFiumContext(ctx, item.ID, page, 1600)
			if err == nil {
				var img image.Image
				img, _, err = eco.DecodeSupportedImage(rendered.PNG)
				if err == nil {
					assessment := eco.AssessImage(img)
					previewImage := eco.BoundedPreviewImage(img, 8_000_000)
					state := &previewState{itemID: item.ID, isPDF: true, pdfPage: page, title: fmt.Sprintf("%s — page %d", item.SafeName, page), original: previewImage, rotation: 0, mode: "original", zoom: 1, assessment: assessment, cropRect: previewImage.Bounds(), highlight: highlight}
					state.rebuild()
					a.mu.Lock()
					a.pendingPreview = state
					a.mu.Unlock()
				}
			}
			if err != nil {
				a.mu.Lock()
				a.previewErr = err.Error()
				a.mu.Unlock()
			}
			procPostMessageW.Call(a.hwnd, msgPreviewReady, 0, 0)
		}(e, page, highlight)
		return
	}
	if e.Image != nil {
'''
s = s.replace(preview_anchor, pdf_preview, 1)

# Prevent PDF-derived rotations from being persisted as source-image rotations.
s = s.replace('\t\t\t\t_ = app.vault.SetRotation(p.itemID, p.rotation)\n', '\t\t\t\tif !p.isPDF {\n\t\t\t\t\t_ = app.vault.SetRotation(p.itemID, p.rotation)\n\t\t\t\t}\n', 1)
s = s.replace("\t\t\tcase 'C':\n\t\t\t\tif p.highlight != nil {", "\t\t\tcase 'C':\n\t\t\t\tif p.isPDF || p.highlight != nil {", 1)
s = s.replace("\t\t\tcase 'D':\n\t\t\t\tif p.highlight != nil {", "\t\t\tcase 'D':\n\t\t\t\tif p.isPDF || p.highlight != nil {", 1)

controls = '\tdrawTextFont(hdc, "R rotate · C auto-crop · D deskew · O original · G greyscale · H fixed contrast · A adaptive · Q quality · +/− zoom · Esc close", RECT{24, 63, rc.Right - 24, 90}, app.fontSmall, rgb(218, 242, 238), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)\n'
if controls not in s:
    raise SystemExit('preview controls anchor not found')
controls_repl = '''\tcontrols := "R rotate · C auto-crop · D deskew · O original · G greyscale · H fixed contrast · A adaptive · Q quality · +/− zoom · Esc close"\n\tif p.isPDF {\n\t\tcontrols = fmt.Sprintf("PDF page %d · R rotate view · O original · G greyscale · H fixed contrast · A adaptive · +/− zoom · Esc close", p.pdfPage)\n\t}\n\tdrawTextFont(hdc, controls, RECT{24, 63, rc.Right - 24, 90}, app.fontSmall, rgb(218, 242, 238), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)\n'''
s = s.replace(controls, controls_repl, 1)

# Add a dedicated single-EXE picker for the optional renderer.
dialog_anchor = 'func multiString(parts []string) []uint16 {\n'
if dialog_anchor not in s:
    raise SystemExit('dialog helper anchor not found')
exe_dialog = r'''func openExecutableDialog(owner uintptr, title string) string {
	buf := make([]uint16, 32768)
	filter := multiString([]string{"Windows executable", "*.exe", "All files", "*.*"})
	ofn := OPENFILENAME{LStructSize: uint32(unsafe.Sizeof(OPENFILENAME{})), HwndOwner: owner, LpstrFilter: &filter[0], NFilterIndex: 1, LpstrFile: &buf[0], NMaxFile: uint32(len(buf)), LpstrTitle: utf16Ptr(title), Flags: OFN_EXPLORER | OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST | OFN_HIDEREADONLY}
	r, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

'''
s = s.replace(dialog_anchor, exe_dialog + dialog_anchor, 1)
p.write_text(s, encoding='utf-8')

# Third-party/provenance notice for the optional runtime.
p = ROOT / 'THIRD_PARTY_NOTICES.md'
s = p.read_text(encoding='utf-8')
notice = '''\n\n## Qualified optional PDF page-rendering runtime (not bundled)\n\n- `klippa-app/pdfium-cli` release `v0.11.2`, exact source-tag commit `260c846dbbd180fdc478a2771e9dae9914164846`.\n- ECO accepts only the single-file Windows WebAssembly asset `pdfium-webassembly-windows-amd64`, 16,988,160 bytes, SHA-256 `b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f`.\n- `pdfium-cli` is MIT licensed. Upstream states its embedded Wazero and PDFium components are Apache-2.0 licensed.\n- The renderer is an optional caller-located local executable. It is not committed into this source repository and ECO does not download it at runtime.\n- ECO re-verifies the executable before each use, materialises only a freshly verified temporary reading copy of the preserved PDF, renders a bounded temporary PNG, and deletes both temporary derivatives afterwards.\n'''
if 'Qualified optional PDF page-rendering runtime' not in s:
    s += notice
p.write_text(s, encoding='utf-8')

doc = '''# pdfium-cli isolated PDF page renderer integration\n\nDate: 2026-09-04\n\n## Decision\n\nUSE/ADAPT as an isolated optional runtime. ECO does not link `go-pdfium` or PDFium/CGO into the application. The exact `klippa-app/pdfium-cli` WebAssembly Windows release `v0.11.2` is used as a caller-located executable because it is a single file and does not require a separately installed PDFium DLL stack.\n\n## Exact qualified identity\n\n- Upstream: `klippa-app/pdfium-cli`\n- Source tag commit: `260c846dbbd180fdc478a2771e9dae9914164846`\n- Wrapper licence: MIT; upstream states bundled Wazero and PDFium are Apache-2.0.\n- Asset: `pdfium-webassembly-windows-amd64`\n- Bytes: `16988160`\n- SHA-256: `b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f`\n\nGitHub-hosted Windows qualification rendered the same known one-page PDF repeatedly at a 1200-pixel width in roughly 2.7 seconds, with a maximum observed live working set of about 286.9 MiB. The native comparison executable did not render successfully on the clean runner without its external runtime stack. These measurements are development evidence only; they are not the controlling low-spec Acer qualification.\n\n## ECO safety boundary\n\nThe runtime registry checks the exact executable identity and self-checks its `render`/`info` command surface. The executable is re-verified immediately before use. ECO then materialises a freshly verified temporary reading copy of one preserved PDF, renders one explicit page with a 45-second default deadline, limits requested width to 320–2000 pixels and height to 3000 pixels, limits the source reading copy to 512 MiB, limits the PNG to 32 MiB and six million pixels, validates the PNG header/dimensions, and deletes the temporary workspace. No rendered page is persisted as evidence and no renderer download occurs inside ECO.\n\nOrdinary PDF preview renders page 1. A source-backed citation with a recorded `Citation.Page` requests that exact page. The existing image-preview window is reused for viewing the rendered derivative; PDF-specific preview rotation is view-only and is not persisted as evidence rotation.\n\n## Remaining gate\n\nBefore this renderer can be described as suitable for the controlling low-spec Windows target, the exact runtime plus ECO adapter must be exercised on that real machine. This integration does not waive accessibility, clean-machine, signing, publisher or release gates.\n'''
(ROOT / 'docs/foss/PDFIUM_CLI_PAGE_RENDERER_INTEGRATION.md').write_text(doc, encoding='utf-8')
