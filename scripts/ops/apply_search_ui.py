from pathlib import Path

p = Path('cmd/eco/main_windows.go')
s = p.read_text(encoding='utf-8')

def replace_once(old: str, new: str, label: str) -> None:
    global s
    n = s.count(old)
    if n != 1:
        raise SystemExit(f'{label}: expected exactly one anchor, found {n}')
    s = s.replace(old, new, 1)

replace_once(
'''\tctrlQuestion      = 1001
\tctrlAsk           = 1002
\tctrlAnswer        = 1003
\tmsgImportProgress = WM_APP + 1
''',
'''\tctrlQuestion      = 1001
\tctrlAsk           = 1002
\tctrlAnswer        = 1003
\tctrlSearchEdit    = 1004
\tctrlSearchAll     = 1005
\tctrlSearchItem    = 1006
\tctrlSearchPrev    = 1007
\tctrlSearchNext    = 1008
\tctrlSearchOpen    = 1009
\tmsgImportProgress = WM_APP + 1
''', 'control ids')

replace_once(
'''\tmsgMatterDone     = WM_APP + 10
\tmsgRuntimeDone    = WM_APP + 11
\tmsgPDFPageReady   = WM_APP + 50
''',
'''\tmsgMatterDone     = WM_APP + 10
\tmsgRuntimeDone    = WM_APP + 11
\tmsgSearchDone     = WM_APP + 12
\tmsgSearchOpenReady = WM_APP + 13
\tmsgPDFPageReady   = WM_APP + 50
''', 'message ids')

replace_once(
'''\tquestionEdit, askButton, answerEdit                                       uintptr
\tbuttons                                                                   map[string]RECT
''',
'''\tquestionEdit, askButton, answerEdit                                       uintptr
\tsearchEdit, searchAllButton, searchItemButton                              uintptr
\tsearchPrevButton, searchNextButton, searchOpenButton                       uintptr
\tbuttons                                                                   map[string]RECT
''', 'application controls')

replace_once(
'''\tpendingCitationRegion                                                     *eco.NormalizedRegion
\tpendingCitationPage                                                       int
}\n''',
'''\tpendingCitationRegion                                                     *eco.NormalizedRegion
\tpendingCitationPage                                                       int
\tsearchReceipt                                                             eco.SearchReceipt
\tsearchIndex                                                               int
\tsearchErr                                                                 string
\tsearchRunning                                                             bool
\tsearchActivity                                                            string
\tsearchOpenIndex                                                           int
\tsearchOpenErr                                                             string
}\n''', 'application search state')

replace_once(
'''\t\tapp.questionEdit = createWindowEx(WS_EX_CLIENTEDGE, "EDIT", "", WS_CHILD|WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL, 0, 0, 10, 10, hwnd, ctrlQuestion, app.hInstance, nil)
\t\tapp.askButton = createWindowEx(0, "BUTTON", "Ask ECO", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON, 0, 0, 10, 10, hwnd, ctrlAsk, app.hInstance, nil)
\t\tapp.answerEdit = createWindowEx(WS_EX_CLIENTEDGE, "EDIT", "Ask a question about readable evidence. ECO will answer only from local source passages.", WS_CHILD|WS_VSCROLL|ES_LEFT|ES_MULTILINE|ES_AUTOVSCROLL|ES_READONLY, 0, 0, 10, 10, hwnd, ctrlAnswer, app.hInstance, nil)
''',
'''\t\tapp.questionEdit = createWindowEx(WS_EX_CLIENTEDGE, "EDIT", "", WS_CHILD|WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL, 0, 0, 10, 10, hwnd, ctrlQuestion, app.hInstance, nil)
\t\tapp.askButton = createWindowEx(0, "BUTTON", "Ask ECO", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON, 0, 0, 10, 10, hwnd, ctrlAsk, app.hInstance, nil)
\t\tapp.answerEdit = createWindowEx(WS_EX_CLIENTEDGE, "EDIT", "Ask a question about readable evidence. ECO will answer only from local source passages.", WS_CHILD|WS_VSCROLL|ES_LEFT|ES_MULTILINE|ES_AUTOVSCROLL|ES_READONLY, 0, 0, 10, 10, hwnd, ctrlAnswer, app.hInstance, nil)
\t\tapp.searchEdit = createWindowEx(WS_EX_CLIENTEDGE, "EDIT", "", WS_CHILD|WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL, 0, 0, 10, 10, hwnd, ctrlSearchEdit, app.hInstance, nil)
\t\tapp.searchAllButton = createWindowEx(0, "BUTTON", "Search all", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON, 0, 0, 10, 10, hwnd, ctrlSearchAll, app.hInstance, nil)
\t\tapp.searchItemButton = createWindowEx(0, "BUTTON", "This item", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON, 0, 0, 10, 10, hwnd, ctrlSearchItem, app.hInstance, nil)
\t\tapp.searchPrevButton = createWindowEx(0, "BUTTON", "Previous", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON, 0, 0, 10, 10, hwnd, ctrlSearchPrev, app.hInstance, nil)
\t\tapp.searchNextButton = createWindowEx(0, "BUTTON", "Next", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON, 0, 0, 10, 10, hwnd, ctrlSearchNext, app.hInstance, nil)
\t\tapp.searchOpenButton = createWindowEx(0, "BUTTON", "Open match", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON, 0, 0, 10, 10, hwnd, ctrlSearchOpen, app.hInstance, nil)
''', 'create search controls')

replace_once(
'''\tcase WM_COMMAND:
\t\tid := int(loword(wparam))
\t\tcode := int(hiword(wparam))
\t\tif id == ctrlAsk && code == BN_CLICKED {
\t\t\tapp.askEvidence()
\t\t}
\t\treturn 0
''',
'''\tcase WM_COMMAND:
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
''', 'command routing')

replace_once(
'''\tcase msgAskDone:
\t\tapp.mu.Lock()
\t\trec := app.lastQuestion
\t\tapp.mu.Unlock()
\t\tsetWindowText(app.answerEdit, rec.Answer)
\t\tapp.refreshView()
\t\tinvalidate(hwnd)
\t\treturn 0
\tcase msgVerifyDone:
''',
'''\tcase msgAskDone:
\t\tapp.mu.Lock()
\t\trec := app.lastQuestion
\t\tapp.mu.Unlock()
\t\tsetWindowText(app.answerEdit, rec.Answer)
\t\tapp.refreshView()
\t\tinvalidate(hwnd)
\t\treturn 0
\tcase msgSearchDone:
\t\tapp.mu.Lock()
\t\terrText := app.searchErr
\t\tapp.searchRunning = false
\t\tapp.searchActivity = ""
\t\tapp.mu.Unlock()
\t\tapp.selectSearchMatch(0)
\t\tapp.updateSearchControlState()
\t\tif errText != "" {
\t\t\tmessageBox(hwnd, "Search did not complete", errText, MB_OK|MB_ICONERROR)
\t\t}
\t\tinvalidate(hwnd)
\t\treturn 0
\tcase msgSearchOpenReady:
\t\tapp.mu.Lock()
\t\terrText := app.searchOpenErr
\t\tindex := app.searchOpenIndex
\t\tapp.searchOpenErr = ""
\t\tapp.searchRunning = false
\t\tapp.searchActivity = ""
\t\tif errText != "" {
\t\t\tapp.searchErr = errText
\t\t}
\t\tapp.mu.Unlock()
\t\tapp.updateSearchControlState()
\t\tif errText != "" {
\t\t\tmessageBox(hwnd, "Search result changed", "ECO refused to open the old match because its verified source or readable extraction changed. Run the search again.\r\n\r\n"+errText, MB_OK|MB_ICONERROR)
\t\t\tinvalidate(hwnd)
\t\t\treturn 0
\t\t}
\t\tapp.activateSearchMatch(index, true)
\t\treturn 0
\tcase msgVerifyDone:
''', 'search completion messages')

start = s.index('func (a *application) layoutControls(')
end = s.index('func (a *application) paint(', start)
new_layout = '''func (a *application) layoutControls(w, h int32) {
\taskShow := a.page == "ask"
\tsearchShow := a.page == "evidence"

\tif askShow {
\t\tleft := int32(305)
\t\ttop := int32(168)
\t\tright := w - 45
\t\tif right-left < 720 {
\t\t\tright = left + 720
\t\t}
\t\tprocMoveWindow.Call(a.questionEdit, uintptr(left), uintptr(top), uintptr(right-left-135), 44, 1)
\t\tprocMoveWindow.Call(a.askButton, uintptr(right-120), uintptr(top), 120, 44, 1)
\t\tanswerRight := right - 350
\t\tif answerRight-left < 360 {
\t\t\tanswerRight = right
\t\t}
\t\tprocMoveWindow.Call(a.answerEdit, uintptr(left), uintptr(top+70), uintptr(answerRight-left), uintptr(max32(260, h-top-120)), 1)
\t\tfor _, c := range []uintptr{a.questionEdit, a.askButton, a.answerEdit} {
\t\t\tprocSendMessageW.Call(c, WM_SETFONT, a.fontBody, 1)
\t\t\tshowWindow(c, SW_SHOW)
\t\t}
\t} else {
\t\tfor _, c := range []uintptr{a.questionEdit, a.askButton, a.answerEdit} {
\t\t\tshowWindow(c, SW_HIDE)
\t\t}
\t}

\tsearchControls := []uintptr{a.searchEdit, a.searchAllButton, a.searchItemButton, a.searchPrevButton, a.searchNextButton, a.searchOpenButton}
\tif !searchShow {
\t\tfor _, c := range searchControls {
\t\t\tshowWindow(c, SW_HIDE)
\t\t}
\t\treturn
\t}
\tleft := int32(300)
\tright := w - 25
\tlistRight := right - 365
\tif listRight-left < 300 {
\t\tlistRight = left + 300
\t}
\twidth := listRight - left
\tgap := int32(5)
\tallW := int32(72)
\titemW := int32(82)
\teditW := width - allW - itemW - 2*gap
\tif editW < 100 {
\t\teditW = 100
\t}
\ty := int32(300)
\tprocMoveWindow.Call(a.searchEdit, uintptr(left), uintptr(y), uintptr(editW), 34, 1)
\tprocMoveWindow.Call(a.searchAllButton, uintptr(left+editW+gap), uintptr(y), uintptr(allW), 34, 1)
\tprocMoveWindow.Call(a.searchItemButton, uintptr(left+editW+gap+allW+gap), uintptr(y), uintptr(itemW), 34, 1)
\tprevW := int32(78)
\tnextW := int32(68)
\topenW := width - prevW - nextW - 2*gap
\tprocMoveWindow.Call(a.searchPrevButton, uintptr(left), uintptr(y+40), uintptr(prevW), 34, 1)
\tprocMoveWindow.Call(a.searchNextButton, uintptr(left+prevW+gap), uintptr(y+40), uintptr(nextW), 34, 1)
\tprocMoveWindow.Call(a.searchOpenButton, uintptr(left+prevW+gap+nextW+gap), uintptr(y+40), uintptr(openW), 34, 1)
\tfor _, c := range searchControls {
\t\tprocSendMessageW.Call(c, WM_SETFONT, a.fontBody, 1)
\t\tshowWindow(c, SW_SHOW)
\t}
\ta.updateSearchControlState()
}

'''
s = s[:start] + new_layout + s[end:]

replace_once(
'''\ty := int32(292)
\tif !a.processing() {
\t\ty = 218
\t}
\tlistRight := right - 365
\tdrawTextFont(hdc, "PRESERVED ITEMS", RECT{x, y, listRight, y + 26}, a.fontLabel, rgb(65, 91, 93), DT_LEFT|DT_SINGLELINE)
\ty += 34
''',
'''\tlistRight := right - 365
\tdrawTextFont(hdc, "SEARCH VERIFIED READINGS", RECT{x, 280, listRight, 298}, a.fontLabel, rgb(65, 91, 93), DT_LEFT|DT_SINGLELINE)
\tdrawTextFont(hdc, a.searchStatusText(), RECT{x, 379, listRight, 400}, a.fontSmall, rgb(82, 102, 105), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
\ty := int32(406)
\tdrawTextFont(hdc, "PRESERVED ITEMS", RECT{x, y, listRight, y + 26}, a.fontLabel, rgb(65, 91, 93), DT_LEFT|DT_SINGLELINE)
\ty += 34
''', 'evidence search layout')

replace_once(
'''\t\tcase 'R':
\t\t\ta.restoreBackup()
\t\t\treturn true
\t\t}
''',
'''\t\tcase 'R':
\t\t\ta.restoreBackup()
\t\t\treturn true
\t\tcase 'F':
\t\t\tif a.page != "evidence" {
\t\t\t\ta.setPage("evidence")
\t\t\t}
\t\t\tprocSetFocus.Call(a.searchEdit)
\t\t\treturn true
\t\t}
''', 'Ctrl+F shortcut')

replace_once(
'''Ctrl+R: restore encrypted backup safely\\r\\nAlt+1..7: change workspace\\r\\nUp/Down: select evidence\\r\\nEnter: open selected preview\\r\\n\\r\\nImage preview:''',
'''Ctrl+R: restore encrypted backup safely\\r\\nCtrl+F: focus verified evidence search\\r\\nAlt+1..7: change workspace\\r\\nUp/Down: select evidence\\r\\nEnter: open selected preview\\r\\n\\r\\nImage preview:''', 'help shortcut')

insert_at = s.index('func (a *application) askEvidence()')
search_methods = r'''func (a *application) searchEvidence(currentOnly bool) {
	q := strings.TrimSpace(getWindowText(a.searchEdit))
	if q == "" {
		a.mu.Lock()
		a.searchReceipt = eco.SearchReceipt{}
		a.searchIndex = 0
		a.searchErr = ""
		a.searchActivity = ""
		a.mu.Unlock()
		a.updateSearchControlState()
		invalidate(a.hwnd)
		return
	}
	var scope []string
	if currentOnly {
		if len(a.view.Evidence) == 0 || a.selected < 0 || a.selected >= len(a.view.Evidence) {
			return
		}
		scope = []string{a.view.Evidence[a.selected].ID}
	}
	a.mu.Lock()
	if a.searchRunning {
		a.mu.Unlock()
		return
	}
	a.searchRunning = true
	a.searchActivity = "Searching verified readable segments…"
	a.searchErr = ""
	a.searchReceipt = eco.SearchReceipt{}
	a.searchIndex = 0
	a.mu.Unlock()
	a.updateSearchControlState()
	invalidate(a.hwnd)
	go func(query string, scopeIDs []string) {
		receipt, err := a.vault.SearchWorkspace(query, scopeIDs)
		a.mu.Lock()
		if err != nil {
			a.searchErr = err.Error()
			a.searchReceipt = eco.SearchReceipt{}
		} else {
			a.searchReceipt = receipt
			a.searchErr = ""
		}
		a.searchIndex = 0
		a.mu.Unlock()
		procPostMessageW.Call(a.hwnd, msgSearchDone, 0, 0)
	}(q, append([]string(nil), scope...))
}

func (a *application) searchStatusText() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.searchRunning {
		if a.searchActivity != "" {
			return a.searchActivity
		}
		return "Working locally…"
	}
	if a.searchErr != "" {
		return "Search needs to be run again — " + truncateUI(a.searchErr, 90)
	}
	if a.searchReceipt.ID == "" {
		return "Search all readable evidence or limit the search to the selected item."
	}
	if len(a.searchReceipt.Matches) == 0 {
		return "No matches in the selected search scope."
	}
	idx := a.searchIndex
	if idx < 0 || idx >= len(a.searchReceipt.Matches) {
		idx = 0
	}
	m := a.searchReceipt.Matches[idx]
	where := m.SafeName
	if m.Page > 0 {
		where += fmt.Sprintf(" · page %d", m.Page)
	}
	if m.Origin == "ocr" {
		where += fmt.Sprintf(" · OCR %.0f%%", m.Confidence*100)
	}
	count := fmt.Sprintf("%d of %d", idx+1, len(a.searchReceipt.Matches))
	if a.searchReceipt.Truncated {
		count += " · result limit reached"
	}
	return count + " · " + where + " · " + truncateUI(m.Snippet, 90)
}

func (a *application) updateSearchControlState() {
	a.mu.Lock()
	running := a.searchRunning
	n := len(a.searchReceipt.Matches)
	idx := a.searchIndex
	a.mu.Unlock()
	enable := func(hwnd uintptr, on bool) {
		if hwnd == 0 {
			return
		}
		value := uintptr(0)
		if on {
			value = 1
		}
		procEnableWindow.Call(hwnd, value)
	}
	enable(a.searchAllButton, !running)
	enable(a.searchItemButton, !running && len(a.view.Evidence) > 0)
	enable(a.searchPrevButton, !running && n > 0 && idx > 0)
	enable(a.searchNextButton, !running && n > 0 && idx+1 < n)
	enable(a.searchOpenButton, !running && n > 0 && idx >= 0 && idx < n)
}

func (a *application) selectSearchMatch(index int) {
	a.mu.Lock()
	if len(a.searchReceipt.Matches) == 0 {
		a.searchIndex = 0
		a.mu.Unlock()
		a.updateSearchControlState()
		invalidate(a.hwnd)
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(a.searchReceipt.Matches) {
		index = len(a.searchReceipt.Matches) - 1
	}
	a.searchIndex = index
	match := a.searchReceipt.Matches[index]
	a.mu.Unlock()
	for i, item := range a.view.Evidence {
		if item.ID == match.EvidenceID {
			a.selected = i
			break
		}
	}
	a.updateSearchControlState()
	invalidate(a.hwnd)
}

func (a *application) moveSearchMatch(delta int) {
	a.mu.Lock()
	idx := a.searchIndex
	n := len(a.searchReceipt.Matches)
	running := a.searchRunning
	a.mu.Unlock()
	if running || n == 0 {
		return
	}
	next := idx + delta
	if next < 0 || next >= n {
		return
	}
	a.selectSearchMatch(next)
}

func (a *application) openSearchMatch() {
	a.mu.Lock()
	if a.searchRunning || len(a.searchReceipt.Matches) == 0 || a.searchIndex < 0 || a.searchIndex >= len(a.searchReceipt.Matches) {
		a.mu.Unlock()
		return
	}
	receipt := a.searchReceipt
	index := a.searchIndex
	a.searchRunning = true
	a.searchActivity = "Rechecking the selected match against preserved source bytes…"
	a.searchOpenErr = ""
	a.searchOpenIndex = index
	a.mu.Unlock()
	a.updateSearchControlState()
	invalidate(a.hwnd)
	go func(r eco.SearchReceipt, idx int) {
		err := a.vault.ValidateSearchReceipt(r)
		a.mu.Lock()
		a.searchOpenIndex = idx
		if err != nil {
			a.searchOpenErr = err.Error()
		} else {
			a.searchOpenErr = ""
		}
		a.mu.Unlock()
		procPostMessageW.Call(a.hwnd, msgSearchOpenReady, 0, 0)
	}(receipt, index)
}

func (a *application) activateSearchMatch(index int, open bool) {
	a.mu.Lock()
	if index < 0 || index >= len(a.searchReceipt.Matches) {
		a.mu.Unlock()
		return
	}
	a.searchIndex = index
	match := a.searchReceipt.Matches[index]
	a.mu.Unlock()
	for i, item := range a.view.Evidence {
		if item.ID != match.EvidenceID {
			continue
		}
		a.selected = i
		if !open {
			break
		}
		a.mu.Lock()
		if match.Region != nil {
			region := *match.Region
			a.pendingCitationRegion = &region
		} else {
			a.pendingCitationRegion = nil
		}
		a.pendingCitationPage = match.Page
		a.mu.Unlock()
		where := "No page number is recorded for this reading."
		if match.Page > 0 {
			where = fmt.Sprintf("Recorded page: %d.", match.Page)
		}
		provenance := "Native/extracted readable text."
		if match.Origin == "ocr" {
			provenance = fmt.Sprintf("OCR-derived reading · %.0f%% recorded confidence.", match.Confidence*100)
		}
		highlightNote := "ECO can open the recorded page/source, but this reading does not contain an exact visual text box."
		if match.Region != nil {
			highlightNote = "The preview will highlight the recorded source region for this reading."
		}
		messageBox(a.hwnd, "Verified search match — "+match.SafeName, match.Snippet+"\r\n\r\n"+where+"\r\n"+provenance+"\r\n\r\n"+highlightNote+"\r\n\r\nSelect OK to open the preserved source preview.", MB_OK|MB_ICONINFORMATION)
		a.openSelectedPreview()
		break
	}
	a.updateSearchControlState()
	invalidate(a.hwnd)
}

'''
s = s[:insert_at] + search_methods + s[insert_at:]

p.write_text(s, encoding='utf-8')

# Add source-regression assertions to the existing cross-platform UI tests.
tp = Path('cmd/eco/ui_regression_test.go')
t = tp.read_text(encoding='utf-8')
marker = '''func TestNativeSourceContainsNoBrowserOrLocalhostRuntime(t *testing.T) {\n'''
if t.count(marker) != 1:
    raise SystemExit('UI regression insertion marker missing or duplicated')
new_tests = r'''func TestEvidenceSearchUIUsesBackgroundSourceBoundNavigation(t *testing.T) {
	src := windowsSource(t)
	for _, required := range []string{"ctrlSearchEdit", "Search all", "This item", "Previous", "Next", "Open match", "msgSearchDone", "msgSearchOpenReady", "SearchWorkspace", "ValidateSearchReceipt", "pendingCitationPage = match.Page", "pendingCitationRegion", "openSelectedPreview", "procEnableWindow", "Ctrl+F: focus verified evidence search"} {
		if !strings.Contains(src, required) {
			t.Fatalf("missing native search UI safeguard %q", required)
		}
	}
	body := functionBody(t, src, "func (a *application) searchEvidence", "func (a *application) searchStatusText")
	if !strings.Contains(body, "go func") || !strings.Contains(body, "procPostMessageW.Call(a.hwnd, msgSearchDone") {
		t.Fatal("search must run off the UI thread and return through the Windows message queue")
	}
	openBody := functionBody(t, src, "func (a *application) openSearchMatch", "func (a *application) activateSearchMatch")
	if !strings.Contains(openBody, "ValidateSearchReceipt") || !strings.Contains(openBody, "msgSearchOpenReady") {
		t.Fatal("opening a search result must revalidate its receipt asynchronously before preview")
	}
}

'''
t = t.replace(marker, new_tests + marker, 1)
tp.write_text(t, encoding='utf-8')
