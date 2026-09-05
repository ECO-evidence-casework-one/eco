from pathlib import Path

ROOT = Path('.')


def replace_once(text: str, old: str, new: str, label: str) -> str:
    count = text.count(old)
    if count != 1:
        raise SystemExit(f'{label}: expected exactly one anchor, found {count}')
    return text.replace(old, new, 1)


# Backend: reuse pdfium-cli info JSON for bounded, source-bound page count.
p = ROOT / 'internal/eco/pdfium_renderer.go'
s = p.read_text(encoding='utf-8')
s = replace_once(s, 'import (\n\t"bytes"\n\t"context"', 'import (\n\t"bytes"\n\t"context"\n\t"encoding/json"', 'renderer json import')
s = replace_once(
    s,
    '\tmaxPDFiumDiagnosticBytes        = 64 * 1024\n\tmaxPDFiumRenderPNGBytes   int64 = 32 * 1024 * 1024\n',
    '\tmaxPDFiumDiagnosticBytes        = 64 * 1024\n\tmaxPDFiumInfoBytes              = 8 * 1024 * 1024\n\tmaxPDFiumRenderPNGBytes   int64 = 32 * 1024 * 1024\n',
    'info size constant',
)
page_render_type = '''type PDFPageRender struct {\n\tEngineVersion string    `json:"engine_version"`\n\tSourceObject  string    `json:"source_object"`\n\tSourceSHA256  string    `json:"source_sha256"`\n\tPage          int       `json:"page"`\n\tWidth         int       `json:"width"`\n\tHeight        int       `json:"height"`\n\tPNG           []byte    `json:"-"`\n\tCreatedAt     time.Time `json:"created_at"`\n}\n\n'''
doc_info_type = page_render_type + '''type PDFDocumentInfo struct {\n\tEngineVersion string    `json:"engine_version"`\n\tSourceObject  string    `json:"source_object"`\n\tSourceSHA256  string    `json:"source_sha256"`\n\tPageCount     int       `json:"page_count"`\n\tCreatedAt     time.Time `json:"created_at"`\n}\n\ntype pdfiumCLIInfoJSON struct {\n\tPageCount int `json:"PageCount"`\n}\n\n'''
s = replace_once(s, page_render_type, doc_info_type, 'document info types')
render_args = '''func pdfiumRenderArgs(inputPath, outputPath string, page, maxWidth int) []string {\n\treturn []string{\n\t\t"render",\n\t\t"--pages", strconv.Itoa(page),\n\t\t"--file-type", "png",\n\t\t"--max-width", strconv.Itoa(maxWidth),\n\t\t"--max-height", strconv.Itoa(maxPDFiumRenderHeight),\n\t\tinputPath,\n\t\toutputPath,\n\t}\n}\n\n'''
info_impl = render_args + r'''func pdfiumInfoArgs(inputPath string) []string {
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

'''
s = replace_once(s, render_args, info_impl, 'document info implementation')
workflow_anchor = '''type pdfiumRenderRunner func(context.Context, string, string, int, int, SourceReceipt) (PDFPageRender, error)\n\n'''
workflow_impl = r'''type pdfiumInfoRunner func(context.Context, string, string, SourceReceipt) (PDFDocumentInfo, error)

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

'''
s = replace_once(s, workflow_anchor, workflow_impl, 'vault document info workflow')
p.write_text(s, encoding='utf-8')

# Tests: parser, command surface, bounds/source binding.
p = ROOT / 'internal/eco/pdfium_renderer_test.go'
s = p.read_text(encoding='utf-8')
append_tests = r'''

func TestPDFiumInfoArgsRequestJSONToStdout(t *testing.T) {
	got := pdfiumInfoArgs(`C:\source.pdf`)
	want := []string{"info", "--output-type", "json", `C:\source.pdf`, "-"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected pdfium-cli info args: %#v", got)
	}
}

func TestParsePDFiumDocumentInfoBindsPageCountToSource(t *testing.T) {
	source := SourceReceipt{ObjectFile: "object-abc.eco", SHA256: strings.Repeat("b", 64), VerifiedAt: time.Now().UTC()}
	info, err := parsePDFiumDocumentInfo([]byte(`{"PageCount": 23}`), source, qualifiedPDFiumCLIVersion)
	if err != nil {
		t.Fatal(err)
	}
	if info.PageCount != 23 || info.SourceObject != source.ObjectFile || info.SourceSHA256 != source.SHA256 {
		t.Fatalf("unexpected document info: %+v", info)
	}
}

func TestParsePDFiumDocumentInfoRejectsUnsafePageCount(t *testing.T) {
	source := SourceReceipt{ObjectFile: "object-abc.eco", SHA256: strings.Repeat("b", 64), VerifiedAt: time.Now().UTC()}
	if _, err := parsePDFiumDocumentInfo([]byte(`{"PageCount": 0}`), source, qualifiedPDFiumCLIVersion); err == nil {
		t.Fatal("zero page count was accepted")
	}
	if _, err := parsePDFiumDocumentInfo([]byte(`{"PageCount": 100001}`), source, qualifiedPDFiumCLIVersion); err == nil {
		t.Fatal("unbounded page count was accepted")
	}
}
'''
if 'TestPDFiumInfoArgsRequestJSONToStdout' in s:
    raise SystemExit('navigation tests already present')
s = s.rstrip() + append_tests + '\n'
p.write_text(s, encoding='utf-8')

# Windows constants for navigation keys.
p = ROOT / 'cmd/eco/winapi_windows.go'
s = p.read_text(encoding='utf-8')
s = replace_once(
    s,
    '\tVK_UP                = 0x26\n\tVK_DOWN              = 0x28\n\tVK_HOME              = 0x24\n',
    '\tVK_UP                = 0x26\n\tVK_DOWN              = 0x28\n\tVK_LEFT              = 0x25\n\tVK_RIGHT             = 0x27\n\tVK_PRIOR             = 0x21\n\tVK_NEXT              = 0x22\n\tVK_HOME              = 0x24\n',
    'Windows PDF navigation keys',
)
p.write_text(s, encoding='utf-8')

# Native preview: query page count once, show X of Y, and render subsequent
# pages asynchronously in the same preview window.
p = ROOT / 'cmd/eco/main_windows.go'
s = p.read_text(encoding='utf-8')
s = replace_once(
    s,
    '\tmsgMatterDone     = WM_APP + 10\n)',
    '\tmsgMatterDone     = WM_APP + 10\n\tmsgPDFPageReady   = WM_APP + 50\n)',
    'preview page-ready message',
)
s = replace_once(
    s,
    '\tisPDF         bool\n\tpdfPage       int\n\ttitle         string\n',
    '\tisPDF         bool\n\tpdfPage       int\n\tpdfPages      int\n\tpdfTargetPage int\n\tpdfLoading    bool\n\tpdfErr        string\n\tsourceTitle   string\n\ttitle         string\n',
    'preview navigation state',
)
old_pdf_goroutine = r'''		go func(item eco.EvidenceItem, page int, highlight *eco.NormalizedRegion) {
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
'''
new_pdf_goroutine = r'''		go func(item eco.EvidenceItem, page int, highlight *eco.NormalizedRegion) {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			pageCount := 0
			if info, infoErr := a.vault.PDFEvidenceInfoWithRegisteredPDFiumContext(ctx, item.ID); infoErr == nil {
				pageCount = info.PageCount
				if page > pageCount {
					err := fmt.Errorf("PDF page %d is outside this document's %d pages", page, pageCount)
					a.mu.Lock()
					a.previewErr = err.Error()
					a.mu.Unlock()
					procPostMessageW.Call(a.hwnd, msgPreviewReady, 0, 0)
					return
				}
			}
			rendered, err := a.vault.RenderEvidencePDFPageWithRegisteredPDFiumContext(ctx, item.ID, page, 1600)
			if err == nil {
				var img image.Image
				img, _, err = eco.DecodeSupportedImage(rendered.PNG)
				if err == nil {
					assessment := eco.AssessImage(img)
					previewImage := eco.BoundedPreviewImage(img, 8_000_000)
					title := fmt.Sprintf("%s — page %d", item.SafeName, page)
					if pageCount > 0 {
						title = fmt.Sprintf("%s — page %d of %d", item.SafeName, page, pageCount)
					}
					state := &previewState{itemID: item.ID, isPDF: true, pdfPage: page, pdfPages: pageCount, sourceTitle: item.SafeName, title: title, original: previewImage, rotation: 0, mode: "original", zoom: 1, assessment: assessment, cropRect: previewImage.Bounds(), highlight: highlight}
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
'''
s = replace_once(s, old_pdf_goroutine, new_pdf_goroutine, 'initial PDF page-count query')
key_anchor = '''\tcase WM_KEYDOWN:\n\t\tpreviewMu.Lock()\n\t\tp := previews[hwnd]\n\t\tif p != nil {\n\t\t\tswitch uint32(wparam) {\n'''
key_repl = '''\tcase WM_KEYDOWN:\n\t\tpdfTarget := 0\n\t\tpreviewMu.Lock()\n\t\tp := previews[hwnd]\n\t\tif p != nil {\n\t\t\tswitch uint32(wparam) {\n'''
s = replace_once(s, key_anchor, key_repl, 'preview key target state')
escape_anchor = '''\t\t\tcase VK_ESCAPE:\n\t\t\t\tpreviewMu.Unlock()\n\t\t\t\tprocDestroyWindow.Call(hwnd)\n\t\t\t\treturn 0\n\t\t\tcase 'R':\n'''
escape_repl = '''\t\t\tcase VK_ESCAPE:\n\t\t\t\tpreviewMu.Unlock()\n\t\t\t\tprocDestroyWindow.Call(hwnd)\n\t\t\t\treturn 0\n\t\t\tcase VK_LEFT, VK_PRIOR:\n\t\t\t\tif p.isPDF && !p.pdfLoading && p.pdfPages > 0 && p.pdfPage > 1 {\n\t\t\t\t\tpdfTarget = p.pdfPage - 1\n\t\t\t\t}\n\t\t\tcase VK_RIGHT, VK_NEXT:\n\t\t\t\tif p.isPDF && !p.pdfLoading && p.pdfPages > 0 && p.pdfPage < p.pdfPages {\n\t\t\t\t\tpdfTarget = p.pdfPage + 1\n\t\t\t\t}\n\t\t\tcase 'R':\n'''
s = replace_once(s, escape_anchor, escape_repl, 'preview navigation keys')
unlock_anchor = '''\t\tpreviewMu.Unlock()\n\t\tinvalidate(hwnd)\n\t\treturn 0\n\tcase WM_MOUSEWHEEL:\n'''
unlock_repl = '''\t\tpreviewMu.Unlock()\n\t\tif pdfTarget > 0 {\n\t\t\trequestPDFPreviewPage(hwnd, pdfTarget)\n\t\t}\n\t\tinvalidate(hwnd)\n\t\treturn 0\n\tcase msgPDFPageReady:\n\t\tpreviewMu.Lock()\n\t\tp := previews[hwnd]\n\t\terrText := ""\n\t\tif p != nil {\n\t\t\terrText = p.pdfErr\n\t\t\tp.pdfErr = ""\n\t\t}\n\t\tpreviewMu.Unlock()\n\t\tif errText != "" {\n\t\t\tmessageBox(hwnd, "PDF page unavailable", errText, MB_OK|MB_ICONERROR)\n\t\t}\n\t\tinvalidate(hwnd)\n\t\treturn 0\n\tcase WM_MOUSEWHEEL:\n'''
s = replace_once(s, unlock_anchor, unlock_repl, 'preview page-ready handling')
rebuild_anchor = '''func (p *previewState) rebuild() {\n'''
request_helper = r'''func requestPDFPreviewPage(hwnd uintptr, target int) {
	previewMu.Lock()
	p := previews[hwnd]
	if p == nil || !p.isPDF || p.pdfLoading || p.pdfPages < 1 || target < 1 || target > p.pdfPages || target == p.pdfPage {
		previewMu.Unlock()
		return
	}
	itemID := p.itemID
	p.pdfLoading = true
	p.pdfTargetPage = target
	p.pdfErr = ""
	previewMu.Unlock()
	invalidate(hwnd)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		rendered, err := app.vault.RenderEvidencePDFPageWithRegisteredPDFiumContext(ctx, itemID, target, 1600)
		var previewImage image.Image
		var assessment eco.ImageAssessment
		if err == nil {
			var img image.Image
			img, _, err = eco.DecodeSupportedImage(rendered.PNG)
			if err == nil {
				assessment = eco.AssessImage(img)
				previewImage = eco.BoundedPreviewImage(img, 8_000_000)
			}
		}

		previewMu.Lock()
		current := previews[hwnd]
		if current == nil || current.itemID != itemID {
			previewMu.Unlock()
			return
		}
		current.pdfLoading = false
		current.pdfTargetPage = 0
		if err != nil {
			current.pdfErr = err.Error()
			previewMu.Unlock()
			procPostMessageW.Call(hwnd, msgPDFPageReady, 0, 0)
			return
		}
		current.pdfPage = target
		current.title = fmt.Sprintf("%s — page %d of %d", current.sourceTitle, target, current.pdfPages)
		current.original = previewImage
		current.assessment = assessment
		current.cropRect = previewImage.Bounds()
		current.highlight = nil
		current.rebuild()
		previewMu.Unlock()
		procPostMessageW.Call(hwnd, msgPDFPageReady, 0, 0)
	}()
}

'''
s = replace_once(s, rebuild_anchor, request_helper + rebuild_anchor, 'async PDF navigation helper')
controls_old = '''\tif p.isPDF {\n\t\tcontrols = fmt.Sprintf("PDF page %d · R rotate view · O original · G greyscale · H fixed contrast · A adaptive · +/− zoom · Esc close", p.pdfPage)\n\t}\n'''
controls_new = '''\tif p.isPDF {\n\t\tif p.pdfLoading {\n\t\t\tcontrols = fmt.Sprintf("Rendering PDF page %d of %d locally…", p.pdfTargetPage, p.pdfPages)\n\t\t} else if p.pdfPages > 0 {\n\t\t\tcontrols = fmt.Sprintf("PDF page %d of %d · ←/→ or PageUp/PageDown change page · R rotate view · +/− zoom · Esc close", p.pdfPage, p.pdfPages)\n\t\t} else {\n\t\t\tcontrols = fmt.Sprintf("PDF page %d · page count unavailable · R rotate view · +/− zoom · Esc close", p.pdfPage)\n\t\t}\n\t}\n'''
s = replace_once(s, controls_old, controls_new, 'PDF navigation control text')
p.write_text(s, encoding='utf-8')

# Integration record.
doc = '''# PDFium multi-page preview navigation\n\nDate: 2026-09-05\n\n## Decision\n\nExtend the already-qualified optional `klippa-app/pdfium-cli` v0.11.2 renderer; add no new donor. ECO uses the renderer's existing `info --output-type json` surface to obtain a bounded page count from the same verified preserved PDF source.\n\n## Behaviour\n\n- Initial PDF preview requests page metadata once, bound to the preserved object SHA-256.\n- The current page is shown as `page X of Y` when metadata is available.\n- Left/Right Arrow and PageUp/PageDown request the previous/next page without closing the preview window.\n- Page changes are rendered asynchronously so the native preview window remains responsive.\n- While a page is rendering, duplicate navigation requests are ignored and the preview reports the target page.\n- Each new page still goes through `RenderEvidencePDFPageWithRegisteredPDFiumContext`, so the registered runtime is re-verified and the preserved source is freshly verified before rendering.\n- Page navigation clears any citation-region highlight when moving away from the cited page.\n- If page-count metadata is unavailable, the existing page-1 / exact-citation-page preview remains usable; navigation is simply disabled.\n\n## Safety\n\nThe existing renderer limits remain unchanged: no runtime download, exact v0.11.2 SHA/size identity, 45-second render deadline, source <=512 MiB, bounded diagnostics, width/height/pixel/PNG limits and temporary derivative cleanup. Metadata stdout is additionally capped at 8 MiB and page count is constrained to ECO's existing safe render range.\n\nThis source slice still requires the normal Linux/Windows/source-policy/Gitleaks/Syft/Cosign pipeline and an exact-runtime Windows navigation qualification. The 8 GB Acer remains the controlling low-spec hardware gate for performance claims.\n'''
(ROOT / 'docs/foss/PDFIUM_PAGE_NAVIGATION_INTEGRATION.md').write_text(doc, encoding='utf-8')
