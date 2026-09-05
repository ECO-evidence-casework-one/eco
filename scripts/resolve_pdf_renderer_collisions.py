from pathlib import Path
import subprocess
import sys

PDF_BRANCH = "origin/integration/pdfium-page-renderer-20260904"
EXPECTED_CONFLICTS = ["THIRD_PARTY_NOTICES.md", "cmd/eco/main_windows.go"]


def run(*args, check=True):
    return subprocess.run(args, text=True, capture_output=True, check=check)


def stage(n: int, path: str) -> str:
    return subprocess.check_output(["git", "show", f":{n}:{path}"], text=True)


def replace_once(text: str, old: str, new: str, label: str) -> str:
    count = text.count(old)
    if count != 1:
        raise RuntimeError(f"{label}: expected exactly one anchor, found {count}")
    return text.replace(old, new, 1)


def merge_pdf_branch() -> None:
    run("git", "fetch", "origin", "integration/pdfium-page-renderer-20260904")
    result = run("git", "merge", "--no-ff", "--no-commit", PDF_BRANCH, check=False)
    if result.returncode == 0:
        raise RuntimeError("expected controlled conflicts did not occur; refusing unreviewed merge shape")
    conflicts = subprocess.check_output(
        ["git", "diff", "--name-only", "--diff-filter=U"], text=True
    ).splitlines()
    conflicts = sorted(x.strip() for x in conflicts if x.strip())
    if conflicts != EXPECTED_CONFLICTS:
        raise RuntimeError(f"unexpected conflict set: {conflicts!r}")


def resolve_notices() -> None:
    ours = stage(2, "THIRD_PARTY_NOTICES.md").rstrip() + "\n"
    theirs = stage(3, "THIRD_PARTY_NOTICES.md")
    marker = "## Qualified optional PDF page-rendering runtime (not bundled)"
    if marker not in theirs:
        raise RuntimeError("PDFium notice marker missing from PDF branch")
    if marker in ours:
        raise RuntimeError("PDFium notice unexpectedly already present on combined stack")
    pdf_notice = marker + theirs.split(marker, 1)[1]
    Path("THIRD_PARTY_NOTICES.md").write_text(
        ours + "\n" + pdf_notice.strip() + "\n", encoding="utf-8"
    )


def resolve_windows_ui() -> None:
    s = stage(2, "cmd/eco/main_windows.go")
    s = replace_once(
        s,
        'import (\n\t"encoding/binary"',
        'import (\n\t"context"\n\t"encoding/binary"',
        "context import",
    )
    s = replace_once(
        s,
        '\truntimeNotice                                                             string\n\tpendingCitationRegion                                                     *eco.NormalizedRegion\n}',
        '\truntimeNotice                                                             string\n\tpendingCitationRegion                                                     *eco.NormalizedRegion\n\tpendingCitationPage                                                       int\n}',
        "application citation page field",
    )
    s = replace_once(
        s,
        'type previewState struct {\n\titemID        string\n\ttitle         string',
        'type previewState struct {\n\titemID        string\n\tisPDF         bool\n\tpdfPage       int\n\ttitle         string',
        "preview PDF fields",
    )
    s = replace_once(
        s,
        '\tbuttonTop := rc.Bottom - 108\n',
        '\tbuttonTop := rc.Bottom - 160\n',
        "Trust page third-row space",
    )

    old_runtime_row = '''\ttesseractLabel := "Locate verified Tesseract runtime"\n\tif reg, err := a.vault.RegisteredTesseractRuntimeBundle(); err == nil {\n\t\ttesseractLabel = "Tesseract " + reg.Version + " — verified runtime registered"\n\t}\n\ta.drawButton(hdc, "tesseractRuntime", tesseractLabel, RECT{x, buttonTop + 104, right, buttonTop + 146}, false)\n'''
    new_runtime_row = '''\ttesseractLabel := "Locate verified Tesseract runtime"\n\tif reg, err := a.vault.RegisteredTesseractRuntimeBundle(); err == nil {\n\t\ttesseractLabel = "Tesseract " + reg.Version + " — verified runtime registered"\n\t}\n\ta.drawButton(hdc, "tesseractRuntime", tesseractLabel, RECT{x, buttonTop + 104, x + w2, buttonTop + 146}, false)\n\tpdfRendererLabel := "Locate verified PDF page renderer"\n\tif reg, err := a.vault.RegisteredLocalTool("pdfium-cli"); err == nil {\n\t\tpdfRendererLabel = "PDF renderer " + reg.Version + " — verified runtime registered"\n\t}\n\ta.drawButton(hdc, "pdfRenderer", pdfRendererLabel, RECT{x + w2 + gapB, buttonTop + 104, right, buttonTop + 146}, false)\n'''
    s = replace_once(s, old_runtime_row, new_runtime_row, "combined runtime button row")

    s = replace_once(
        s,
        '\t\t\tcase "tesseractRuntime":\n\t\t\t\ta.locateTesseractRuntime()\n\t\t\tcase "lowSensory":\n',
        '\t\t\tcase "tesseractRuntime":\n\t\t\t\ta.locateTesseractRuntime()\n\t\t\tcase "pdfRenderer":\n\t\t\t\ta.locatePDFRenderer()\n\t\t\tcase "lowSensory":\n',
        "PDF renderer click handler",
    )

    choose_files = '''func (a *application) chooseFiles() {\n\tpaths, err := openFileDialog(a.hwnd)\n\tif err != nil {\n\t\tmessageBox(a.hwnd, "Could not open file picker", err.Error(), MB_OK|MB_ICONERROR)\n\t\treturn\n\t}\n\tif len(paths) > 0 {\n\t\ta.beginImport(paths)\n\t}\n}\n\n'''
    locate_pdf = choose_files + r'''func (a *application) locatePDFRenderer() {
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
    s = replace_once(s, choose_files, locate_pdf, "locate PDF renderer function")

    s = replace_once(
        s,
        '\t\t\t} else {\n\t\t\t\ta.pendingCitationRegion = nil\n\t\t\t}\n\t\t\ta.mu.Unlock()\n\t\t\tnote := "Select OK to open the preserved source preview next."',
        '\t\t\t} else {\n\t\t\t\ta.pendingCitationRegion = nil\n\t\t\t}\n\t\t\ta.pendingCitationPage = c.Page\n\t\t\ta.mu.Unlock()\n\t\t\tnote := "Select OK to open the preserved source preview next."',
        "citation page carry-through",
    )

    preview_anchor = '\te := a.view.Evidence[a.selected]\n\tif e.Image != nil {\n'
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
    s = replace_once(s, preview_anchor, pdf_preview, "PDF preview insertion")

    s = replace_once(
        s,
        '\t\t\tcase \'R\':\n\t\t\t\tp.rotation = (p.rotation + 90) % 360\n\t\t\t\tp.rebuild()\n\t\t\t\t_ = app.vault.SetRotation(p.itemID, p.rotation)\n',
        '\t\t\tcase \'R\':\n\t\t\t\tp.rotation = (p.rotation + 90) % 360\n\t\t\t\tp.rebuild()\n\t\t\t\tif !p.isPDF {\n\t\t\t\t\t_ = app.vault.SetRotation(p.itemID, p.rotation)\n\t\t\t\t}\n',
        "PDF view-only rotation",
    )

    guard = '\t\t\t\tif p.highlight != nil {\n'
    if s.count(guard) != 2:
        raise RuntimeError(f"crop/deskew highlight guards: expected 2, found {s.count(guard)}")
    s = s.replace(guard, '\t\t\t\tif p.isPDF || p.highlight != nil {\n', 2)

    controls_old = '\tdrawTextFont(hdc, "R rotate · C auto-crop · D deskew · O original · G greyscale · H fixed contrast · A adaptive · Q quality · +/− zoom · Esc close", RECT{24, 63, rc.Right - 24, 90}, app.fontSmall, rgb(218, 242, 238), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)\n'
    controls_new = '''\tcontrols := "R rotate · C auto-crop · D deskew · O original · G greyscale · H fixed contrast · A adaptive · Q quality · +/− zoom · Esc close"\n\tif p.isPDF {\n\t\tcontrols = fmt.Sprintf("PDF page %d · R rotate view · O original · G greyscale · H fixed contrast · A adaptive · +/− zoom · Esc close", p.pdfPage)\n\t}\n\tdrawTextFont(hdc, controls, RECT{24, 63, rc.Right - 24, 90}, app.fontSmall, rgb(218, 242, 238), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)\n'''
    s = replace_once(s, controls_old, controls_new, "PDF preview controls")

    multi_anchor = 'func multiString(parts []string) []uint16 {\n'
    executable_dialog = r'''func openExecutableDialog(owner uintptr, title string) string {
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
    s = replace_once(s, multi_anchor, executable_dialog + multi_anchor, "executable picker")
    Path("cmd/eco/main_windows.go").write_text(s, encoding="utf-8")


def main() -> int:
    merge_pdf_branch()
    resolve_notices()
    resolve_windows_ui()
    run("git", "add", "THIRD_PARTY_NOTICES.md", "cmd/eco/main_windows.go")
    remaining = subprocess.check_output(
        ["git", "diff", "--name-only", "--diff-filter=U"], text=True
    ).strip()
    if remaining:
        raise RuntimeError(f"unresolved conflicts remain: {remaining}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"controlled PDF renderer conflict resolution failed: {exc}", file=sys.stderr)
        raise
