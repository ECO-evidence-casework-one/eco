from pathlib import Path

main = Path('cmd/eco/main_windows.go')
winapi = Path('cmd/eco/winapi_windows.go')
test = Path('cmd/eco/ui_regression_test.go')
s = main.read_text(encoding='utf-8')
w = winapi.read_text(encoding='utf-8')
t = test.read_text(encoding='utf-8')

def rep(text, old, new, label):
    n = text.count(old)
    if n != 1:
        raise SystemExit(f'{label}: expected one anchor, found {n}')
    return text.replace(old, new, 1)

w = rep(w, '''\tBS_PUSHBUTTON        = 0x00000000
\tBS_DEFPUSHBUTTON     = 0x00000001
''', '''\tBS_PUSHBUTTON        = 0x00000000
\tBS_DEFPUSHBUTTON     = 0x00000001
\tBS_OWNERDRAW         = 0x0000000B
\tLBS_NOTIFY           = 0x0001
\tLBN_SELCHANGE        = 1
\tLBN_DBLCLK           = 2
\tLB_ADDSTRING         = 0x0180
\tLB_RESETCONTENT      = 0x0184
\tLB_SETCURSEL         = 0x0186
\tLB_GETCURSEL         = 0x0188
''', 'winapi control constants')
w = rep(w, '''\tWM_GETMINMAXINFO     = 0x0024
\tWM_COMMAND           = 0x0111
''', '''\tWM_GETMINMAXINFO     = 0x0024
\tWM_DRAWITEM          = 0x002B
\tWM_COMMAND           = 0x0111
''', 'WM_DRAWITEM')
w = rep(w, '''\tWM_CTLCOLORSTATIC    = 0x0138
\tWM_CTLCOLOREDIT      = 0x0133
''', '''\tWM_CTLCOLORSTATIC    = 0x0138
\tWM_CTLCOLOREDIT      = 0x0133
\tWM_CTLCOLORLISTBOX   = 0x0134
''', 'listbox color')
w = rep(w, '''\tSWP_NOACTIVATE       = 0x0010
)''', '''\tSWP_NOACTIVATE       = 0x0010
\tODS_SELECTED         = 0x0001
\tODS_DISABLED         = 0x0004
\tODS_FOCUS            = 0x0010
)''', 'owner draw states')
w = rep(w, '''type PAINTSTRUCT struct {
\tHdc                uintptr
\tErase              int32
\tRcPaint            RECT
\tRestore, IncUpdate int32
\tReserved           [32]byte
}
''', '''type PAINTSTRUCT struct {
\tHdc                uintptr
\tErase              int32
\tRcPaint            RECT
\tRestore, IncUpdate int32
\tReserved           [32]byte
}
type DRAWITEMSTRUCT struct {
\tCtlType, CtlID, ItemID, ItemAction, ItemState uint32
\tHwndItem, HDC                                uintptr
\tRcItem                                       RECT
\tItemData                                     uintptr
}
''', 'draw item struct')

s = rep(s, '''\tctrlSearchOpen     = 1009
\tmsgImportProgress  = WM_APP + 1
''', '''\tctrlSearchOpen     = 1009
\tctrlNavHome        = 1100
\tctrlNavEvidence    = 1101
\tctrlNavAsk         = 1102
\tctrlNavMatters     = 1103
\tctrlNavReview      = 1104
\tctrlNavChanges     = 1105
\tctrlNavTrust       = 1106
\tctrlAddFiles       = 1120
\tctrlAddFolder      = 1121
\tctrlPasteImage     = 1122
\tctrlOpenPreview    = 1123
\tctrlEvidenceList   = 1130
\tmsgImportProgress  = WM_APP + 1
''', 'semantic control ids')
s = rep(s, '''\tsearchEdit, searchAllButton, searchItemButton                             uintptr
\tsearchPrevButton, searchNextButton, searchOpenButton                      uintptr
\tbuttons                                                                   map[string]RECT
''', '''\tsearchEdit, searchAllButton, searchItemButton                             uintptr
\tsearchPrevButton, searchNextButton, searchOpenButton                      uintptr
\tnavButtons                                                                [7]uintptr
\taddFilesButton, addFolderButton, pasteImageButton                         uintptr
\topenPreviewButton, evidenceList                                            uintptr
\tbuttons                                                                   map[string]RECT
''', 'semantic fields')

create_anchor = '''\tcase WM_CREATE:
\t\tapp.hwnd = hwnd
\t\tapp.questionEdit = createWindowEx'''
create_repl = '''\tcase WM_CREATE:
\t\tapp.hwnd = hwnd
\t\tfor i, spec := range []struct{ id int; label string }{{ctrlNavHome, "Home"}, {ctrlNavEvidence, "Evidence"}, {ctrlNavAsk, "Ask ECO"}, {ctrlNavMatters, "Matters"}, {ctrlNavReview, "Review"}, {ctrlNavChanges, "Changes"}, {ctrlNavTrust, "Trust & settings"}} {
\t\t\tapp.navButtons[i] = createWindowEx(0, "BUTTON", spec.label, WS_CHILD|WS_TABSTOP|BS_OWNERDRAW, 0, 0, 10, 10, hwnd, spec.id, app.hInstance, nil)
\t\t}
\t\tapp.addFilesButton = createWindowEx(0, "BUTTON", "+  Add files", WS_CHILD|WS_TABSTOP|BS_OWNERDRAW, 0, 0, 10, 10, hwnd, ctrlAddFiles, app.hInstance, nil)
\t\tapp.addFolderButton = createWindowEx(0, "BUTTON", "+  Add folder", WS_CHILD|WS_TABSTOP|BS_OWNERDRAW, 0, 0, 10, 10, hwnd, ctrlAddFolder, app.hInstance, nil)
\t\tapp.pasteImageButton = createWindowEx(0, "BUTTON", "Paste image", WS_CHILD|WS_TABSTOP|BS_OWNERDRAW, 0, 0, 10, 10, hwnd, ctrlPasteImage, app.hInstance, nil)
\t\tapp.openPreviewButton = createWindowEx(0, "BUTTON", "Open native preview", WS_CHILD|WS_TABSTOP|BS_OWNERDRAW, 0, 0, 10, 10, hwnd, ctrlOpenPreview, app.hInstance, nil)
\t\tapp.evidenceList = createWindowEx(WS_EX_CLIENTEDGE, "LISTBOX", "", WS_CHILD|WS_TABSTOP|WS_VSCROLL|LBS_NOTIFY, 0, 0, 10, 10, hwnd, ctrlEvidenceList, app.hInstance, nil)
\t\tapp.questionEdit = createWindowEx'''
s = rep(s, create_anchor, create_repl, 'WM_CREATE semantic controls')

s = rep(s, '''\tcase WM_PAINT:
\t\tapp.paint(hwnd)
\t\treturn 0
\tcase WM_ERASEBKGND:
''', '''\tcase WM_PAINT:
\t\tapp.paint(hwnd)
\t\treturn 0
\tcase WM_DRAWITEM:
\t\tif app.drawSemanticButton(lparam) {
\t\t\treturn 1
\t\t}
\tcase WM_ERASEBKGND:
''', 'WM_DRAWITEM route')

old_cmd = '''\tcase WM_COMMAND:
\t\tid := int(loword(wparam))
\t\tcode := int(hiword(wparam))
\t\tif code == BN_CLICKED {
\t\t\tswitch id {
\t\t\tcase ctrlAsk:
\t\t\t\tapp.askEvidence()
\t\t\tcase ctrlSearchAll:
\t\t\t\tapp.searchEvidence(false)
\t\t\tcase ctrlSearchItem:
\t\t\t\tapp.searchEvidence(true)
\t\t\tcase ctrlSearchPrev:
\t\t\t\tapp.moveSearchMatch(-1)
\t\t\tcase ctrlSearchNext:
\t\t\t\tapp.moveSearchMatch(1)
\t\t\tcase ctrlSearchOpen:
\t\t\t\tapp.openSearchMatch()
\t\t\t}
\t\t}
\t\treturn 0
'''
new_cmd = '''\tcase WM_COMMAND:
\t\tid := int(loword(wparam))
\t\tcode := int(hiword(wparam))
\t\tif id == ctrlEvidenceList && (code == LBN_SELCHANGE || code == LBN_DBLCLK) {
\t\t\tindex, _, _ := procSendMessageW.Call(app.evidenceList, LB_GETCURSEL, 0, 0)
\t\t\tif int(index) >= 0 && int(index) < len(app.view.Evidence) {
\t\t\t\tapp.selected = int(index)
\t\t\t\tinvalidate(hwnd)
\t\t\t\tif code == LBN_DBLCLK {
\t\t\t\t\tapp.openSelectedPreview()
\t\t\t\t}
\t\t\t}
\t\t\treturn 0
\t\t}
\t\tif code == BN_CLICKED {
\t\t\tswitch id {
\t\t\tcase ctrlNavHome:
\t\t\t\tapp.setPage("home")
\t\t\tcase ctrlNavEvidence:
\t\t\t\tapp.setPage("evidence")
\t\t\tcase ctrlNavAsk:
\t\t\t\tapp.setPage("ask")
\t\t\tcase ctrlNavMatters:
\t\t\t\tapp.setPage("matters")
\t\t\tcase ctrlNavReview:
\t\t\t\tapp.setPage("review")
\t\t\tcase ctrlNavChanges:
\t\t\t\tapp.setPage("changes")
\t\t\tcase ctrlNavTrust:
\t\t\t\tapp.setPage("trust")
\t\t\tcase ctrlAddFiles:
\t\t\t\tapp.chooseFiles()
\t\t\tcase ctrlAddFolder:
\t\t\t\tapp.chooseFolder()
\t\t\tcase ctrlPasteImage:
\t\t\t\tapp.pasteClipboardImage()
\t\t\tcase ctrlOpenPreview:
\t\t\t\tapp.openSelectedPreview()
\t\t\tcase ctrlAsk:
\t\t\t\tapp.askEvidence()
\t\t\tcase ctrlSearchAll:
\t\t\t\tapp.searchEvidence(false)
\t\t\tcase ctrlSearchItem:
\t\t\t\tapp.searchEvidence(true)
\t\t\tcase ctrlSearchPrev:
\t\t\t\tapp.moveSearchMatch(-1)
\t\t\tcase ctrlSearchNext:
\t\t\t\tapp.moveSearchMatch(1)
\t\t\tcase ctrlSearchOpen:
\t\t\t\tapp.openSearchMatch()
\t\t\t}
\t\t}
\t\treturn 0
'''
s = rep(s, old_cmd, new_cmd, 'WM_COMMAND')
s = rep(s, '''\tcase WM_CTLCOLORSTATIC, WM_CTLCOLOREDIT:
''', '''\tcase WM_CTLCOLORSTATIC, WM_CTLCOLOREDIT, WM_CTLCOLORLISTBOX:
''', 'listbox color route')

layout_start = s.index('func (a *application) layoutControls(')
layout_end = s.index('func (a *application) paint(', layout_start)
layout = r'''func (a *application) layoutControls(w, h int32) {
	// Sidebar navigation is always real keyboard-focusable Win32 buttons.
	y := int32(118)
	for _, c := range a.navButtons {
		procMoveWindow.Call(c, 18, uintptr(y), 239, 48, 1)
		procSendMessageW.Call(c, WM_SETFONT, a.fontNav, 1)
		showWindow(c, SW_SHOW)
		y += 53
	}

	askShow := a.page == "ask"
	if askShow {
		left := int32(305)
		top := int32(168)
		right := w - 45
		if right-left < 720 { right = left + 720 }
		procMoveWindow.Call(a.questionEdit, uintptr(left), uintptr(top), uintptr(right-left-135), 44, 1)
		procMoveWindow.Call(a.askButton, uintptr(right-120), uintptr(top), 120, 44, 1)
		answerRight := right - 350
		if answerRight-left < 360 { answerRight = right }
		procMoveWindow.Call(a.answerEdit, uintptr(left), uintptr(top+70), uintptr(answerRight-left), uintptr(max32(260, h-top-120)), 1)
		for _, c := range []uintptr{a.questionEdit, a.askButton, a.answerEdit} { procSendMessageW.Call(c, WM_SETFONT, a.fontBody, 1); showWindow(c, SW_SHOW) }
	} else {
		for _, c := range []uintptr{a.questionEdit, a.askButton, a.answerEdit} { showWindow(c, SW_HIDE) }
	}

	evidenceShow := a.page == "evidence"
	searchControls := []uintptr{a.searchEdit, a.searchAllButton, a.searchItemButton, a.searchPrevButton, a.searchNextButton, a.searchOpenButton}
	evidenceControls := []uintptr{a.addFilesButton, a.addFolderButton, a.pasteImageButton, a.evidenceList, a.openPreviewButton}
	if !evidenceShow {
		for _, c := range searchControls { showWindow(c, SW_HIDE) }
		for _, c := range evidenceControls { showWindow(c, SW_HIDE) }
		return
	}

	left := int32(300)
	right := w - 25
	listRight := right - 365
	if listRight-left < 300 { listRight = left + 300 }
	width := listRight - left
	gap := int32(5)
	allW, itemW := int32(72), int32(82)
	editW := width - allW - itemW - 2*gap
	if editW < 100 { editW = 100 }
	searchY := int32(300)
	procMoveWindow.Call(a.searchEdit, uintptr(left), uintptr(searchY), uintptr(editW), 34, 1)
	procMoveWindow.Call(a.searchAllButton, uintptr(left+editW+gap), uintptr(searchY), uintptr(allW), 34, 1)
	procMoveWindow.Call(a.searchItemButton, uintptr(left+editW+gap+allW+gap), uintptr(searchY), uintptr(itemW), 34, 1)
	prevW, nextW := int32(78), int32(68)
	openW := width - prevW - nextW - 2*gap
	procMoveWindow.Call(a.searchPrevButton, uintptr(left), uintptr(searchY+40), uintptr(prevW), 34, 1)
	procMoveWindow.Call(a.searchNextButton, uintptr(left+prevW+gap), uintptr(searchY+40), uintptr(nextW), 34, 1)
	procMoveWindow.Call(a.searchOpenButton, uintptr(left+prevW+gap+nextW+gap), uintptr(searchY+40), uintptr(openW), 34, 1)
	for _, c := range searchControls { procSendMessageW.Call(c, WM_SETFONT, a.fontBody, 1); showWindow(c, SW_SHOW) }

	procMoveWindow.Call(a.addFilesButton, uintptr(right-180), 118, 180, 46, 1)
	procMoveWindow.Call(a.addFolderButton, uintptr(right-375), 118, 180, 46, 1)
	procMoveWindow.Call(a.pasteImageButton, uintptr(right-545), 118, 155, 46, 1)
	listTop := int32(440)
	listH := h - listTop - 25
	if listH < 120 { listH = 120 }
	procMoveWindow.Call(a.evidenceList, uintptr(left), uintptr(listTop), uintptr(listRight-left), uintptr(listH), 1)
	detailLeft := listRight + 20
	previewTop := h - 89
	if previewTop < 520 { previewTop = 520 }
	procMoveWindow.Call(a.openPreviewButton, uintptr(detailLeft+22), uintptr(previewTop), 176, 42, 1)
	for _, c := range []uintptr{a.addFilesButton, a.addFolderButton, a.pasteImageButton, a.openPreviewButton} { procSendMessageW.Call(c, WM_SETFONT, a.fontLabel, 1); showWindow(c, SW_SHOW) }
	procSendMessageW.Call(a.evidenceList, WM_SETFONT, a.fontBody, 1)
	showWindow(a.evidenceList, SW_SHOW)
	a.updateSemanticControlState()
	a.updateSearchControlState()
}

'''
s = s[:layout_start] + layout + s[layout_end:]

# Keep view and native list in sync.
s = rep(s, '''func (a *application) refreshView() {
\ta.view = a.vault.Snapshot()
\tif len(a.view.Evidence) == 0 {
\t\ta.selected = 0
\t\treturn
\t}
''', '''func (a *application) refreshView() {
\ta.view = a.vault.Snapshot()
\tif len(a.view.Evidence) == 0 {
\t\ta.selected = 0
\t\ta.syncEvidenceList()
\t\treturn
\t}
''', 'refresh empty list')
s = rep(s, '''\tif a.selected >= len(a.view.Evidence) {
\t\ta.selected = len(a.view.Evidence) - 1
\t}
}

func (a *application) drawSidebar''', '''\tif a.selected >= len(a.view.Evidence) {
\t\ta.selected = len(a.view.Evidence) - 1
\t}
\ta.syncEvidenceList()
}

func (a *application) syncEvidenceList() {
\tif a.evidenceList == 0 { return }
\tprocSendMessageW.Call(a.evidenceList, LB_RESETCONTENT, 0, 0)
\tfor _, e := range a.view.Evidence {
\t\tname := utf16Ptr(e.SafeName)
\t\tprocSendMessageW.Call(a.evidenceList, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(name)))
\t}
\tif len(a.view.Evidence) > 0 && a.selected >= 0 && a.selected < len(a.view.Evidence) {
\t\tprocSendMessageW.Call(a.evidenceList, LB_SETCURSEL, uintptr(a.selected), 0)
\t}
\ta.updateSemanticControlState()
}

func (a *application) updateSemanticControlState() {
\tenable := func(hwnd uintptr, on bool) { if hwnd == 0 { return }; v := uintptr(0); if on { v = 1 }; procEnableWindow.Call(hwnd, v) }
\thasEvidence := len(a.view.Evidence) > 0
\tenable(a.openPreviewButton, hasEvidence && a.selected >= 0 && a.selected < len(a.view.Evidence))
\tenable(a.evidenceList, hasEvidence)
}

func semanticNavPage(id uint32) string {
\tswitch id { case ctrlNavHome: return "home"; case ctrlNavEvidence: return "evidence"; case ctrlNavAsk: return "ask"; case ctrlNavMatters: return "matters"; case ctrlNavReview: return "review"; case ctrlNavChanges: return "changes"; case ctrlNavTrust: return "trust" }
\treturn ""
}

func (a *application) drawSemanticButton(lparam uintptr) bool {
\tif lparam == 0 { return false }
\tvar dis DRAWITEMSTRUCT
\tcopyWindowsMemoryToGo(unsafe.Pointer(&dis), lparam, unsafe.Sizeof(dis))
\tpage := semanticNavPage(dis.CtlID)
\tisAction := dis.CtlID == ctrlAddFiles || dis.CtlID == ctrlAddFolder || dis.CtlID == ctrlPasteImage || dis.CtlID == ctrlOpenPreview
\tif page == "" && !isAction { return false }
\tlabel := getWindowText(dis.HwndItem)
\tfocused := dis.ItemState&ODS_FOCUS != 0
\tdisabled := dis.ItemState&ODS_DISABLED != 0
\tif page != "" {
\t\tbg, fg, border := rgb(5, 61, 70), rgb(238, 251, 249), rgb(5, 61, 70)
\t\tif a.page == page || dis.ItemState&ODS_SELECTED != 0 { bg, border = rgb(24, 111, 112), rgb(24, 111, 112) }
\t\tif focused { border = rgb(255, 255, 255) }
\t\tif disabled { fg = rgb(150, 176, 176) }
\t\troundRect(dis.HDC, dis.RcItem, 10, bg, border)
\t\ttextRect := dis.RcItem; textRect.Left += 16; textRect.Right -= 8
\t\tdrawTextFont(dis.HDC, label, textRect, a.fontNav, fg, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
\t\treturn true
\t}
\tprimary := dis.CtlID == ctrlAddFiles || dis.CtlID == ctrlOpenPreview
\tbg, fg, border := rgb(255,255,255), rgb(10,83,82), rgb(205,219,214)
\tif primary { bg, fg, border = rgb(10,105,102), rgb(255,255,255), rgb(10,105,102) }
\tif focused { border = rgb(28, 72, 150) }
\tif disabled { bg, fg, border = rgb(239,242,241), rgb(135,145,146), rgb(211,218,216) }
\troundRect(dis.HDC, dis.RcItem, 10, bg, border)
\tdrawTextFont(dis.HDC, label, dis.RcItem, a.fontLabel, fg, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
\treturn true
}

func (a *application) drawSidebar''', 'semantic helpers')

# Update keyboard evidence selection to keep LISTBOX selected when arrow keys are used.
s = rep(s, '''\t\tcase VK_UP:
\t\t\tif a.page == "evidence" && a.selected > 0 {
\t\t\t\ta.selected--
\t\t\t\tinvalidate(a.hwnd)
\t\t\t}
''', '''\t\tcase VK_UP:
\t\t\tif a.page == "evidence" && a.selected > 0 {
\t\t\t\ta.selected--
\t\t\t\tif a.evidenceList != 0 { procSendMessageW.Call(a.evidenceList, LB_SETCURSEL, uintptr(a.selected), 0) }
\t\t\t\tinvalidate(a.hwnd)
\t\t\t}
''', 'keyboard up sync')
s = rep(s, '''\t\tcase VK_DOWN:
\t\t\tif a.page == "evidence" && a.selected+1 < len(a.view.Evidence) {
\t\t\t\ta.selected++
\t\t\t\tinvalidate(a.hwnd)
\t\t\t}
''', '''\t\tcase VK_DOWN:
\t\t\tif a.page == "evidence" && a.selected+1 < len(a.view.Evidence) {
\t\t\t\ta.selected++
\t\t\t\tif a.evidenceList != 0 { procSendMessageW.Call(a.evidenceList, LB_SETCURSEL, uintptr(a.selected), 0) }
\t\t\t\tinvalidate(a.hwnd)
\t\t\t}
''', 'keyboard down sync')

# Source-regression test for semantic core controls.
marker = 'func TestEvidenceSearchUIUsesBackgroundSourceBoundNavigation(t *testing.T) {'
if t.count(marker) != 1:
    raise SystemExit('UI test marker not found exactly once')
new_test = r'''func TestCoreJourneyUsesNativeSemanticControls(t *testing.T) {
	src := windowsSource(t)
	for _, required := range []string{"ctrlNavHome", "ctrlNavEvidence", "ctrlNavAsk", "ctrlNavMatters", "ctrlNavReview", "ctrlNavChanges", "ctrlNavTrust", "ctrlAddFiles", "ctrlAddFolder", "ctrlPasteImage", "ctrlOpenPreview", "ctrlEvidenceList", "BS_OWNERDRAW", "LBS_NOTIFY", "WM_DRAWITEM", "LBN_SELCHANGE", "LBN_DBLCLK", "syncEvidenceList", "drawSemanticButton", "WS_TABSTOP"} {
		if !strings.Contains(src, required) { t.Fatalf("missing core accessibility control %q", required) }
	}
	createBody := functionBody(t, src, "func mainWndProc", "func (a *application) layoutControls")
	for _, label := range []string{"\"Home\"", "\"Evidence\"", "\"Ask ECO\"", "\"Matters\"", "\"Review\"", "\"Changes\"", "\"Trust & settings\"", "\"+  Add files\"", "\"+  Add folder\"", "\"Paste image\"", "\"Open native preview\"", "\"LISTBOX\""} {
		if !strings.Contains(createBody, label) { t.Fatalf("native core semantic label missing %s", label) }
	}
	if !strings.Contains(src, "focused := dis.ItemState&ODS_FOCUS != 0") { t.Fatal("owner-drawn native controls must render a visible focus state") }
}

'''
t = t.replace(marker, new_test + marker, 1)

winapi.write_text(w, encoding='utf-8')
main.write_text(s, encoding='utf-8')
test.write_text(t, encoding='utf-8')
