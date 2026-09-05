from pathlib import Path
import subprocess


def run(*args, check=True):
    return subprocess.run(args, text=True, capture_output=True, check=check)


def replace_once(text: str, old: str, new: str, label: str) -> str:
    count = text.count(old)
    if count != 1:
        raise RuntimeError(f'{label}: expected exactly one anchor, found {count}')
    return text.replace(old, new, 1)


run('git', 'fetch', 'origin', 'integration/pdfium-page-navigation-20260905')
result = run('git', 'merge', '--no-ff', '--no-commit', 'origin/integration/pdfium-page-navigation-20260905', check=False)
if result.returncode == 0:
    raise RuntimeError('expected one controlled Windows UI conflict but merge was clean')
conflicts = sorted(x.strip() for x in subprocess.check_output(['git','diff','--name-only','--diff-filter=U'], text=True).splitlines() if x.strip())
if conflicts != ['cmd/eco/main_windows.go']:
    raise RuntimeError(f'unexpected conflict set: {conflicts!r}')

# Preserve #120's already-green combined Tesseract/PDF renderer UI and add only
# the page-navigation changes from #121.
subprocess.run(['git','checkout','--ours','--','cmd/eco/main_windows.go'], check=True)
p = Path('cmd/eco/main_windows.go')
s = p.read_text(encoding='utf-8')
s = replace_once(s, '\tmsgRuntimeDone    = WM_APP + 11\n)', '\tmsgRuntimeDone    = WM_APP + 11\n\tmsgPDFPageReady   = WM_APP + 50\n)', 'page-ready message')
s = replace_once(
    s,
    '\tisPDF         bool\n\tpdfPage       int\n\ttitle         string\n',
    '\tisPDF         bool\n\tpdfPage       int\n\tpdfPages      int\n\tpdfTargetPage int\n\tpdfLoading    bool\n\tpdfErr        string\n\tsourceTitle   string\n\ttitle         string\n',
    'preview navigation state',
)
old_pdf = r'''		go func(item eco.EvidenceItem, page int, highlight *eco.NormalizedRegion) {
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
new_pdf = r'''		go func(item eco.EvidenceItem, page int, highlight *eco.NormalizedRegion) {
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
s = replace_once(s, old_pdf, new_pdf, 'initial PDF navigation preparation')
s = replace_once(
    s,
    '\tcase WM_KEYDOWN:\n\t\tpreviewMu.Lock()\n\t\tp := previews[hwnd]\n',
    '\tcase WM_KEYDOWN:\n\t\tpdfTarget := 0\n\t\tpreviewMu.Lock()\n\t\tp := previews[hwnd]\n',
    'preview target state',
)
s = replace_once(
    s,
    '\t\t\tcase VK_ESCAPE:\n\t\t\t\tpreviewMu.Unlock()\n\t\t\t\tprocDestroyWindow.Call(hwnd)\n\t\t\t\treturn 0\n\t\t\tcase \'R\':\n',
    '\t\t\tcase VK_ESCAPE:\n\t\t\t\tpreviewMu.Unlock()\n\t\t\t\tprocDestroyWindow.Call(hwnd)\n\t\t\t\treturn 0\n\t\t\tcase VK_LEFT, VK_PRIOR:\n\t\t\t\tif p.isPDF && !p.pdfLoading && p.pdfPages > 0 && p.pdfPage > 1 {\n\t\t\t\t\tpdfTarget = p.pdfPage - 1\n\t\t\t\t}\n\t\t\tcase VK_RIGHT, VK_NEXT:\n\t\t\t\tif p.isPDF && !p.pdfLoading && p.pdfPages > 0 && p.pdfPage < p.pdfPages {\n\t\t\t\t\tpdfTarget = p.pdfPage + 1\n\t\t\t\t}\n\t\t\tcase \'R\':\n',
    'navigation key cases',
)
s = replace_once(
    s,
    '\t\tpreviewMu.Unlock()\n\t\tinvalidate(hwnd)\n\t\treturn 0\n\tcase WM_MOUSEWHEEL:\n',
    '\t\tpreviewMu.Unlock()\n\t\tif pdfTarget > 0 {\n\t\t\trequestPDFPreviewPage(hwnd, pdfTarget)\n\t\t}\n\t\tinvalidate(hwnd)\n\t\treturn 0\n\tcase msgPDFPageReady:\n\t\tpreviewMu.Lock()\n\t\tp := previews[hwnd]\n\t\terrText := ""\n\t\tif p != nil {\n\t\t\terrText = p.pdfErr\n\t\t\tp.pdfErr = ""\n\t\t}\n\t\tpreviewMu.Unlock()\n\t\tif errText != "" {\n\t\t\tmessageBox(hwnd, "PDF page unavailable", errText, MB_OK|MB_ICONERROR)\n\t\t}\n\t\tinvalidate(hwnd)\n\t\treturn 0\n\tcase WM_MOUSEWHEEL:\n',
    'page ready message handling',
)
helper = r'''func requestPDFPreviewPage(hwnd uintptr, target int) {
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
s = replace_once(s, 'func (p *previewState) rebuild() {\n', helper + 'func (p *previewState) rebuild() {\n', 'navigation render helper')
s = replace_once(
    s,
    '\tif p.isPDF {\n\t\tcontrols = fmt.Sprintf("PDF page %d · R rotate view · O original · G greyscale · H fixed contrast · A adaptive · +/− zoom · Esc close", p.pdfPage)\n\t}\n',
    '\tif p.isPDF {\n\t\tif p.pdfLoading {\n\t\t\tcontrols = fmt.Sprintf("Rendering PDF page %d of %d locally…", p.pdfTargetPage, p.pdfPages)\n\t\t} else if p.pdfPages > 0 {\n\t\t\tcontrols = fmt.Sprintf("PDF page %d of %d · ←/→ or PageUp/PageDown change page · R rotate view · +/− zoom · Esc close", p.pdfPage, p.pdfPages)\n\t\t} else {\n\t\t\tcontrols = fmt.Sprintf("PDF page %d · page count unavailable · R rotate view · +/− zoom · Esc close", p.pdfPage)\n\t\t}\n\t}\n',
    'navigation controls text',
)
p.write_text(s, encoding='utf-8')
subprocess.run(['git','add','cmd/eco/main_windows.go'], check=True)
remaining = subprocess.check_output(['git','diff','--name-only','--diff-filter=U'], text=True).strip()
if remaining:
    raise RuntimeError(f'unresolved conflicts remain: {remaining}')
