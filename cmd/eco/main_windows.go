//go:build windows

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/ECO-evidence-casework-one/eco/internal/eco"
)

const (
	className         = "ECO_V25_NATIVE_MAIN"
	previewClass      = "ECO_V25_IMAGE_PREVIEW"
	inputClass        = "ECO_V25_INPUT_DIALOG"
	ctrlQuestion      = 1001
	ctrlAsk           = 1002
	ctrlAnswer        = 1003
	msgImportProgress = WM_APP + 1
	msgImportDone     = WM_APP + 2
	msgImportError    = WM_APP + 3
	msgRefresh        = WM_APP + 4
	msgBackupDone     = WM_APP + 5
	msgRestoreDone    = WM_APP + 6
	msgAskDone        = WM_APP + 7
	msgVerifyDone     = WM_APP + 8
	msgPreviewReady   = WM_APP + 9
	msgMatterDone     = WM_APP + 10
)

type hitRegion struct {
	Name  string
	Rect  RECT
	Index int
}

type application struct {
	hInstance                                                                 uintptr
	hwnd                                                                      uintptr
	candidate                                                                 *eco.CandidateApplication
	workspace                                                                 eco.WorkspaceSession
	vault                                                                     *eco.Vault
	view                                                                      eco.Workspace
	page                                                                      string
	selected                                                                  int
	dpi                                                                       int32
	fontBody, fontSmall, fontLabel, fontHeading, fontHero, fontNav, fontBrand uintptr
	controlBrush                                                              uintptr
	questionEdit, askButton, answerEdit                                       uintptr
	buttons                                                                   map[string]RECT
	nav                                                                       []hitRegion
	evidenceRows                                                              []hitRegion
	citationRows                                                              []hitRegion
	mu                                                                        sync.Mutex
	progress                                                                  eco.ImportProgress
	importing                                                                 bool
	lastErr                                                                   string
	backupReceipt                                                             eco.BackupReceipt
	restoreReceipt                                                            eco.RestoreReceipt
	lastQuestion                                                              eco.QuestionRecord
	verifyAlerts                                                              []string
	pendingPreview                                                            *previewState
	previewErr                                                                string
	matterCreated                                                             string
	matterErr                                                                 string
	pendingCitationRegion                                                     *eco.NormalizedRegion
}

var app *application
var previewMu sync.Mutex
var previews = map[uintptr]*previewState{}

type previewState struct {
	itemID        string
	title         string
	original      image.Image
	image         image.Image
	rotation      int
	mode          string
	zoom          float64
	cropEnabled   bool
	deskewEnabled bool
	cropRect      image.Rectangle
	assessment    eco.ImageAssessment
	highlight     *eco.NormalizedRegion
	pixels        []byte
	width, height int
}

type inputDialogState struct {
	hwnd                                 uintptr
	owner                                uintptr
	prompt, edit, okButton, cancelButton uintptr
	done                                 bool
	accepted                             bool
	value                                string
	secret                               bool
}

var dialogMu sync.Mutex
var dialogs = map[uintptr]*inputDialogState{}

func main() {
	defer func() {
		if r := recover(); r != nil {
			messageBox(0, "ECO startup failure", fmt.Sprintf("ECO encountered an unexpected startup error.\r\n\r\n%v\r\n\r\n%s", r, string(debug.Stack())), MB_OK|MB_ICONERROR)
		}
	}()
	procSetProcessDpiAwarenessContext.Call(^uintptr(3)) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 (-4)
	procOleInitialize.Call(0)
	defer procOleUninitialize.Call()
	hInst, _, _ := procGetModuleHandleW.Call(0)
	baseRoot := os.Getenv("LOCALAPPDATA")
	if baseRoot == "" {
		baseRoot, _ = os.UserConfigDir()
	}
	if baseRoot == "" {
		messageBox(0, "ECO could not start safely", "Windows did not provide a private application-state folder. ECO did not open or create any workspace.", MB_OK|MB_ICONERROR)
		return
	}
	candidate, err := eco.StartCandidate(filepath.Join(baseRoot, "EvidenceCaseworkOne"), eco.CurrentRuntime())
	if err != nil {
		messageBox(0, "ECO workspace could not open safely", err.Error()+"\r\n\r\nNo older records were opened as though they belonged to this build.", MB_OK|MB_ICONERROR)
		return
	}
	session := candidate.Current
	v := session.Vault
	initialView := v.Snapshot()
	page := initialView.SelectedPage
	if page == "" {
		page = "home"
	}
	app = &application{hInstance: hInst, candidate: candidate, workspace: session, vault: v, view: initialView, page: page, selected: 0, dpi: 96, buttons: map[string]RECT{}}
	registerClasses()
	app.hwnd = createWindowEx(0, className, eco.BuildName, WS_OVERLAPPEDWINDOW|WS_CLIPCHILDREN, CW_USEDEFAULT, CW_USEDEFAULT, 1260, 820, 0, 0, hInst, nil)
	if app.hwnd == 0 {
		messageBox(0, "ECO could not start", "Windows could not create the native ECO window.", MB_OK|MB_ICONERROR)
		return
	}
	if d, _, _ := procGetDpiForWindow.Call(app.hwnd); d > 0 {
		app.dpi = int32(d)
	}
	app.createFonts()
	var initial RECT
	procGetClientRect.Call(app.hwnd, uintptr(unsafe.Pointer(&initial)))
	app.layoutControls(initial.Right, initial.Bottom)
	showWindow(app.hwnd, SW_SHOW)
	procUpdateWindow.Call(app.hwnd)
	app.showWorkspaceStatus("Workspace opened")
	whatsMarker := filepath.Join(candidate.State.StateRoot, "whats-seen-N2-P1")
	if _, err := os.Stat(whatsMarker); os.IsNotExist(err) {
		messageBox(app.hwnd, "What’s new — Evidence & Casework One Version 25 N2", "DOCUMENT VISION FOUNDATION PREVIEW 1\r\n\r\n• Added conservative photographed-page boundary detection and non-destructive auto-crop preview.\r\n• Added bounded skew estimation and non-destructive deskew preview.\r\n• Added adaptive black-and-white reading enhancement.\r\n• Added glare, uneven-lighting, edge-cutoff and probable double-page assessment.\r\n• Added perspective-correction foundations for a later four-corner correction studio.\r\n• Added coordinate-bearing OCR receipt and exact image-region source models.\r\n• Added a vault integration gate that refuses OCR results whose source hash or coordinates do not match the preserved original.\r\n• Preserved the standalone native window, encrypted live vault, streaming intake, signature checks, duplicate detection, Matters, source-backed search and transactional backups.\r\n\r\nA bundled OCR engine and generative local language model are not yet included in N2 P1. This source milestone builds and tests the native document-vision and OCR provenance foundation before those components are approved and bundled.", MB_OK|MB_ICONINFORMATION)
		_ = os.WriteFile(whatsMarker, []byte(eco.BuildID), 0600)
	}
	var msg MSG
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		if msg.Message == WM_KEYDOWN && app.handleGlobalShortcut(uint32(msg.WParam)) {
			continue
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func registerClasses() {
	cursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
	wc := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), Style: CS_HREDRAW | CS_VREDRAW, LpfnWndProc: syscall.NewCallback(mainWndProc), HInstance: app.hInstance, HCursor: cursor, LpszClassName: utf16Ptr(className)}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	pwc := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), Style: CS_HREDRAW | CS_VREDRAW, LpfnWndProc: syscall.NewCallback(previewWndProc), HInstance: app.hInstance, HCursor: cursor, LpszClassName: utf16Ptr(previewClass)}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&pwc)))
	iwc := WNDCLASSEX{CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})), Style: CS_HREDRAW | CS_VREDRAW, LpfnWndProc: syscall.NewCallback(inputWndProc), HInstance: app.hInstance, HCursor: cursor, LpszClassName: utf16Ptr(inputClass)}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&iwc)))
}

func (a *application) createFonts() {
	a.deleteFonts()
	fontDPI := a.dpi
	// P2 scaled fonts to the full monitor DPI while all layout rectangles were
	// still fixed pixels. That caused widespread clipping at 125–150% Windows
	// display scaling. P3 keeps text comfortably readable while the native
	// layout is progressively converted to responsive measurements.
	if fontDPI > 108 {
		fontDPI = 108
	}
	mk := func(pt int, weight int32) uintptr {
		height := -int32(pt) * fontDPI / 72
		r, _, _ := procCreateFontW.Call(uintptr(height), 0, 0, 0, uintptr(weight), 0, 0, 0, DEFAULT_CHARSET, OUT_DEFAULT_PRECIS, CLIP_DEFAULT_PRECIS, CLEARTYPE_QUALITY, DEFAULT_PITCH|FF_DONTCARE, uintptr(unsafe.Pointer(utf16Ptr("Segoe UI"))))
		return r
	}
	a.fontSmall = mk(9, FW_NORMAL)
	a.fontBody = mk(10, FW_NORMAL)
	a.fontLabel = mk(9, FW_SEMIBOLD)
	a.fontHeading = mk(17, FW_BOLD)
	a.fontHero = mk(26, FW_BOLD)
	a.fontNav = mk(10, FW_SEMIBOLD)
	a.fontBrand = mk(13, FW_BOLD)
	a.controlBrush, _, _ = procCreateSolidBrush.Call(rgb(255, 255, 255))
}

func (a *application) deleteFonts() {
	for _, f := range []uintptr{a.fontBody, a.fontSmall, a.fontLabel, a.fontHeading, a.fontHero, a.fontNav, a.fontBrand, a.controlBrush} {
		if f != 0 {
			procDeleteObject.Call(f)
		}
	}
	a.fontBody, a.fontSmall, a.fontLabel, a.fontHeading, a.fontHero, a.fontNav, a.fontBrand, a.controlBrush = 0, 0, 0, 0, 0, 0, 0, 0
}

func mainWndProc(hwnd uintptr, msg uint32, wparam uintptr, lparam unsafe.Pointer) uintptr {
	switch msg {
	case WM_CREATE:
		app.hwnd = hwnd
		app.questionEdit = createWindowEx(WS_EX_CLIENTEDGE, "EDIT", "", WS_CHILD|WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL, 0, 0, 10, 10, hwnd, ctrlQuestion, app.hInstance, nil)
		app.askButton = createWindowEx(0, "BUTTON", "Ask ECO", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON, 0, 0, 10, 10, hwnd, ctrlAsk, app.hInstance, nil)
		app.answerEdit = createWindowEx(WS_EX_CLIENTEDGE, "EDIT", "Ask a question about readable evidence. ECO will answer only from local source passages.", WS_CHILD|WS_VSCROLL|ES_LEFT|ES_MULTILINE|ES_AUTOVSCROLL|ES_READONLY, 0, 0, 10, 10, hwnd, ctrlAnswer, app.hInstance, nil)
		procDragAcceptFiles.Call(hwnd, 1)
		return 0
	case WM_GETMINMAXINFO:
		mmi := (*MINMAXINFO)(lparam)
		mmi.PtMinTrackSize = POINT{X: 1000, Y: 730}
		return 0
	case WM_DPICHANGED:
		newDPI := int32(loword(wparam))
		if newDPI > 0 && newDPI != app.dpi {
			app.dpi = newDPI
			app.createFonts()
			if suggested := (*RECT)(lparam); suggested != nil {
				procSetWindowPos.Call(hwnd, 0, uintptr(suggested.Left), uintptr(suggested.Top), uintptr(suggested.Right-suggested.Left), uintptr(suggested.Bottom-suggested.Top), SWP_NOZORDER|SWP_NOACTIVATE)
			}
			var dr RECT
			procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&dr)))
			app.layoutControls(dr.Right, dr.Bottom)
			invalidate(hwnd)
		}
		return 0
	case WM_SIZE:
		app.layoutControls(loword(uintptr(lparam)), hiword(uintptr(lparam)))
		invalidate(hwnd)
		return 0
	case WM_PAINT:
		app.paint(hwnd)
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_COMMAND:
		id := int(loword(wparam))
		code := int(hiword(wparam))
		if id == ctrlAsk && code == BN_CLICKED {
			app.askEvidence()
		}
		return 0
	case WM_LBUTTONDOWN:
		app.handleClick(signedLoword(uintptr(lparam)), signedHiword(uintptr(lparam)), false)
		return 0
	case WM_LBUTTONDBLCLK:
		app.handleClick(signedLoword(uintptr(lparam)), signedHiword(uintptr(lparam)), true)
		return 0
	case WM_KEYDOWN:
		app.handleKey(uint32(wparam))
		return 0
	case WM_DROPFILES:
		paths := droppedPaths(wparam)
		if len(paths) > 0 {
			app.beginImport(paths)
		}
		return 0
	case WM_CTLCOLORSTATIC, WM_CTLCOLOREDIT:
		procSetBkColor.Call(wparam, rgb(255, 255, 255))
		procSetTextColor.Call(wparam, rgb(23, 38, 41))
		return app.controlBrush
	case msgImportProgress:
		invalidate(hwnd)
		return 0
	case msgRefresh:
		app.refreshView()
		invalidate(hwnd)
		return 0
	case msgImportDone:
		app.mu.Lock()
		app.importing = false
		app.progress = eco.ImportProgress{}
		app.mu.Unlock()
		app.refreshView()
		app.selected = 0
		invalidate(hwnd)
		return 0
	case msgAskDone:
		app.mu.Lock()
		rec := app.lastQuestion
		app.mu.Unlock()
		setWindowText(app.answerEdit, rec.Answer)
		app.refreshView()
		invalidate(hwnd)
		return 0
	case msgVerifyDone:
		app.mu.Lock()
		alerts := append([]string(nil), app.verifyAlerts...)
		app.verifyAlerts = nil
		app.mu.Unlock()
		if len(alerts) == 0 {
			messageBox(hwnd, "Integrity verification passed", fmt.Sprintf("All %d encrypted evidence objects authenticated and matched their recorded SHA-256 values.", len(app.view.Evidence)), MB_OK|MB_ICONINFORMATION)
		} else {
			messageBox(hwnd, "Integrity alerts", strings.Join(alerts, "\r\n"), MB_OK|MB_ICONERROR)
		}
		app.refreshView()
		invalidate(hwnd)
		return 0
	case msgPreviewReady:
		app.mu.Lock()
		state := app.pendingPreview
		errText := app.previewErr
		app.pendingPreview = nil
		app.previewErr = ""
		app.importing = false
		app.progress = eco.ImportProgress{}
		app.mu.Unlock()
		if errText != "" {
			messageBox(hwnd, "Preview unavailable", errText, MB_OK|MB_ICONERROR)
			invalidate(hwnd)
			return 0
		}
		if state != nil {
			hw := createWindowEx(0, previewClass, "ECO source preview — "+state.title, WS_OVERLAPPEDWINDOW, CW_USEDEFAULT, CW_USEDEFAULT, 1050, 780, app.hwnd, 0, app.hInstance, nil)
			if hw != 0 {
				previewMu.Lock()
				previews[hw] = state
				previewMu.Unlock()
				showWindow(hw, SW_SHOW)
				procUpdateWindow.Call(hw)
				invalidate(hw)
			}
		}
		invalidate(hwnd)
		return 0
	case msgMatterDone:
		app.mu.Lock()
		title := app.matterCreated
		errText := app.matterErr
		app.matterCreated = ""
		app.matterErr = ""
		app.importing = false
		app.progress = eco.ImportProgress{}
		app.mu.Unlock()
		app.refreshView()
		if errText != "" {
			messageBox(hwnd, "Matter could not be created", errText, MB_OK|MB_ICONERROR)
		} else if title != "" {
			messageBox(hwnd, "Matter created", title+" was created. A fuller matter editor is scheduled for the next native milestone.", MB_OK|MB_ICONINFORMATION)
		}
		invalidate(hwnd)
		return 0
	case msgBackupDone:
		app.mu.Lock()
		r := app.backupReceipt
		app.importing = false
		app.progress = eco.ImportProgress{}
		app.mu.Unlock()
		app.refreshView()
		messageBox(hwnd, "Encrypted backup created", fmt.Sprintf("Backup: %s\r\nItems: %d\r\nSize: %s\r\nSHA-256: %s", r.Path, r.EvidenceItems, eco.HumanBytes(r.BackupBytes), r.SHA256), MB_OK|MB_ICONINFORMATION)
		invalidate(hwnd)
		return 0
	case msgRestoreDone:
		app.mu.Lock()
		r := app.restoreReceipt
		app.importing = false
		app.progress = eco.ImportProgress{}
		app.mu.Unlock()
		app.refreshView()
		app.selected = 0
		if session, sessionErr := app.candidate.RefreshCurrentAfterRestore(); sessionErr == nil {
			app.workspace = session
		} else if session.Vault != nil {
			app.workspace = session
			messageBox(hwnd, "Restore completed with an application-state warning", sessionErr.Error(), MB_OK|MB_ICONWARNING)
		} else {
			messageBox(hwnd, "Restored workspace needs attention", "The encrypted restore completed, but ECO could not update its candidate-specific selection record.\r\n\r\n"+sessionErr.Error(), MB_OK|MB_ICONWARNING)
		}
		messageBox(hwnd, "Encrypted backup restored safely", fmt.Sprintf("Restored items: %d\r\nRestored bytes: %s\r\nSource build: %s\r\nSource SHA-256: %s\r\n\r\nYour previous vault was retained at:\r\n%s", r.EvidenceItems, eco.HumanBytes(r.RestoredBytes), r.SourceBuildID, r.SourceSHA256, r.PreRestoreVault), MB_OK|MB_ICONINFORMATION)
		invalidate(hwnd)
		return 0
	case msgImportError:
		app.mu.Lock()
		e := app.lastErr
		app.importing = false
		app.mu.Unlock()
		messageBox(hwnd, "Local task did not complete", e, MB_OK|MB_ICONERROR)
		invalidate(hwnd)
		return 0
	case WM_DESTROY:
		app.deleteFonts()
		procPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wparam, uintptr(lparam))
	return r
}

func (a *application) layoutControls(w, h int32) {
	show := a.page == "ask"
	if !show {
		showWindow(a.questionEdit, SW_HIDE)
		showWindow(a.askButton, SW_HIDE)
		showWindow(a.answerEdit, SW_HIDE)
		return
	}
	left := int32(305)
	top := int32(168)
	right := w - 45
	if right-left < 720 {
		right = left + 720
	}
	procMoveWindow.Call(a.questionEdit, uintptr(left), uintptr(top), uintptr(right-left-135), 44, 1)
	procMoveWindow.Call(a.askButton, uintptr(right-120), uintptr(top), 120, 44, 1)
	answerRight := right - 350
	if answerRight-left < 360 {
		answerRight = right
	}
	procMoveWindow.Call(a.answerEdit, uintptr(left), uintptr(top+70), uintptr(answerRight-left), uintptr(max32(260, h-top-120)), 1)
	for _, c := range []uintptr{a.questionEdit, a.askButton, a.answerEdit} {
		procSendMessageW.Call(c, WM_SETFONT, a.fontBody, 1)
		showWindow(c, SW_SHOW)
	}
}

func (a *application) paint(hwnd uintptr) {
	var ps PAINTSTRUCT
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var rc RECT
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	a.buttons = map[string]RECT{}
	a.nav = nil
	a.evidenceRows = nil
	a.citationRows = nil
	fillRect(hdc, rc, rgb(237, 242, 239))
	a.drawSidebar(hdc, rc)
	a.drawTopbar(hdc, rc)
	switch a.page {
	case "home":
		a.drawHome(hdc, rc)
	case "evidence":
		a.drawEvidence(hdc, rc)
	case "ask":
		a.drawAsk(hdc, rc)
	case "matters":
		a.drawMatters(hdc, rc)
	case "review":
		a.drawReview(hdc, rc)
	case "changes":
		a.drawChanges(hdc, rc)
	case "trust":
		a.drawTrust(hdc, rc)
	}
}

func (a *application) refreshView() {
	a.view = a.vault.Snapshot()
	if len(a.view.Evidence) == 0 {
		a.selected = 0
		return
	}
	if a.selected < 0 {
		a.selected = 0
	}
	if a.selected >= len(a.view.Evidence) {
		a.selected = len(a.view.Evidence) - 1
	}
}

func (a *application) drawSidebar(hdc uintptr, rc RECT) {
	sideWidth := int32(275)
	side := RECT{0, 0, sideWidth, rc.Bottom}
	fillRect(hdc, side, rgb(5, 61, 70))

	drawTextFont(hdc, "E1", RECT{22, 18, 72, 66}, a.fontHeading, rgb(218, 249, 243), DT_CENTER|DT_VCENTER|DT_SINGLELINE)
	drawTextFont(hdc, "EVIDENCE &\r\nCASEWORK ONE", RECT{78, 15, sideWidth - 7, 74}, a.fontBrand, rgb(255, 255, 255), DT_LEFT|DT_WORDBREAK)
	drawTextFont(hdc, "Private. Local.\r\nSource-backed.", RECT{80, 75, sideWidth - 15, 108}, a.fontSmall, rgb(197, 231, 228), DT_LEFT|DT_WORDBREAK)

	items := []struct {
		p, l  string
		count int
	}{{"home", "Home", 0}, {"evidence", "Evidence", len(a.view.Evidence)}, {"ask", "Ask ECO", 0}, {"matters", "Matters", len(a.view.Matters)}, {"review", "Review", a.reviewCount()}, {"changes", "Changes", len(a.view.Changes)}, {"trust", "Trust & settings", 0}}
	y := int32(118)
	for i, it := range items {
		r := RECT{18, y, sideWidth - 18, y + 48}
		if a.page == it.p {
			roundRect(hdc, r, 12, rgb(24, 111, 112), rgb(24, 111, 112))
		}
		drawTextFont(hdc, navIcon(i), RECT{30, y, 54, y + 48}, a.fontNav, rgb(225, 248, 245), DT_CENTER|DT_VCENTER|DT_SINGLELINE)
		drawTextFont(hdc, it.l, RECT{67, y, sideWidth - 55, y + 48}, a.fontNav, rgb(238, 251, 249), DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
		if it.count > 0 {
			drawTextFont(hdc, fmt.Sprint(it.count), RECT{sideWidth - 50, y, sideWidth - 23, y + 48}, a.fontSmall, rgb(216, 242, 238), DT_CENTER|DT_VCENTER|DT_SINGLELINE)
		}
		a.nav = append(a.nav, hitRegion{Name: it.p, Rect: r})
		y += 53
	}

	if rc.Bottom >= 690 {
		card := RECT{18, rc.Bottom - 146, sideWidth - 18, rc.Bottom - 20}
		roundRect(hdc, card, 14, rgb(7, 75, 80), rgb(20, 112, 109))
		drawTextFont(hdc, "NATIVE WINDOWS EDITION", RECT{32, card.Top + 13, card.Right - 12, card.Top + 34}, a.fontLabel, rgb(255, 255, 255), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
		drawTextFont(hdc, "✓ No browser or localhost\r\n✓ AES-256-GCM vault\r\n✓ Windows-protected key\r\n✓ No cloud or telemetry", RECT{32, card.Top + 39, card.Right - 12, card.Bottom - 9}, a.fontSmall, rgb(209, 241, 236), DT_LEFT|DT_WORDBREAK)
	}
}

func navIcon(i int) string {
	icons := []string{"⌂", "▣", "✦", "▤", "✓", "↶", "◇"}
	if i < len(icons) {
		return icons[i]
	}
	return "•"
}

func (a *application) drawTopbar(hdc uintptr, rc RECT) {
	top := RECT{275, 0, rc.Right, 70}
	fillRect(hdc, top, rgb(255, 255, 255))
	workspace := "No workspace"
	if a.workspace.Identity.Name != "" {
		workspace = a.workspace.Identity.Name + " • " + string(a.workspace.Disposition)
	}
	drawTextFont(hdc, "●  LOCAL • OFFLINE • "+workspace, RECT{298, 0, rc.Right - 345, 70}, a.fontLabel, rgb(24, 107, 56), DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
	drawTextFont(hdc, eco.BuildID, RECT{rc.Right - 335, 0, rc.Right - 24, 70}, a.fontSmall, rgb(80, 101, 104), DT_RIGHT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
}

func (a *application) drawHome(hdc uintptr, rc RECT) {
	x := int32(300)
	right := rc.Right - 25
	contentW := right - x
	if contentW < 640 {
		return
	}

	heroTop := int32(92)
	heroBottom := int32(420)
	hero := RECT{x, heroTop, right, heroBottom}
	roundRect(hdc, hero, 22, rgb(10, 104, 105), rgb(10, 104, 105))

	stepsW := int32(300)
	if contentW < 900 {
		stepsW = 250
	}
	stepsLeft := right - stepsW - 36
	leftRight := stepsLeft - 30
	if leftRight < x+390 {
		leftRight = x + 390
	}

	drawTextFont(hdc, "VERSION 25 N2 • NATIVE DOCUMENT VISION FOUNDATION", RECT{x + 42, heroTop + 28, leftRight, heroTop + 52}, a.fontLabel, rgb(188, 239, 233), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
	drawTextFont(hdc, "Native. Private.\r\nSource-backed.", RECT{x + 42, heroTop + 58, leftRight, heroTop + 158}, a.fontHero, rgb(255, 255, 255), DT_LEFT|DT_WORDBREAK)
	drawTextFont(hdc, "Preserve evidence inside an encrypted local vault, detect photographed pages, preview crop and deskew corrections, and ask questions using only validated readable source passages.", RECT{x + 42, heroTop + 163, leftRight, heroTop + 226}, a.fontBody, rgb(225, 246, 242), DT_LEFT|DT_WORDBREAK)

	add := RECT{x + 42, heroTop + 248, x + 190, heroTop + 292}
	a.drawButton(hdc, "add", "+  Add evidence", add, true)
	ask := RECT{x + 202, heroTop + 248, x + 332, heroTop + 292}
	a.drawButton(hdc, "goAsk", "✦  Ask ECO", ask, false)
	mat := RECT{x + 344, heroTop + 248, x + 498, heroTop + 292}
	a.drawButton(hdc, "goMatters", "▤  Matters", mat, false)

	steps := []string{"1   Add and encrypt evidence", "2   Review page quality and vision suggestions", "3   Ask from exact validated local sources"}
	sy := heroTop + 42
	for _, text := range steps {
		r := RECT{stepsLeft, sy, right - 36, sy + 64}
		roundRect(hdc, r, 13, rgb(245, 248, 246), rgb(245, 248, 246))
		drawTextFont(hdc, text, RECT{r.Left + 15, r.Top + 10, r.Right - 12, r.Bottom - 9}, a.fontBody, rgb(26, 64, 64), DT_LEFT|DT_WORDBREAK)
		sy += 74
	}

	sectionTop := heroBottom + 26
	drawTextFont(hdc, "CURRENT WORKSPACE", RECT{x, sectionTop, right, sectionTop + 22}, a.fontLabel, rgb(46, 104, 101), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, a.workspace.Identity.Name, RECT{x, sectionTop + 27, right, sectionTop + 58}, a.fontHeading, rgb(20, 61, 59), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
	statusCard := RECT{x, sectionTop + 65, right, sectionTop + 144}
	roundRect(hdc, statusCard, 13, rgb(255, 255, 255), rgb(185, 213, 205))
	drawTextFont(hdc, string(a.workspace.Disposition)+" • "+a.workspace.Compatibility.Message, RECT{statusCard.Left + 15, statusCard.Top + 8, statusCard.Right - 15, statusCard.Top + 29}, a.fontLabel, rgb(17, 91, 86), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
	drawTextFont(hdc, "Identity: "+a.workspace.Identity.ID+" • Format: "+fmt.Sprint(a.workspace.Compatibility.WorkspaceSchema)+" • Created by: "+a.workspace.Identity.CreatedByBuild, RECT{statusCard.Left + 15, statusCard.Top + 31, statusCard.Right - 15, statusCard.Top + 51}, a.fontSmall, rgb(65, 86, 88), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
	drawTextFont(hdc, "Path: "+a.workspace.Path, RECT{statusCard.Left + 15, statusCard.Top + 53, statusCard.Right - 15, statusCard.Bottom - 7}, a.fontSmall, rgb(65, 86, 88), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)

	metrics := a.metrics()
	gap := int32(12)
	cols := 6
	if contentW < 1050 {
		cols = 3
	}
	rows := (len(metrics) + cols - 1) / cols
	cardTop := sectionTop + 157
	cardH := int32(85)
	cw := (contentW - gap*int32(cols-1)) / int32(cols)
	for i, m := range metrics {
		col := i % cols
		row := i / cols
		mx := x + int32(col)*(cw+gap)
		my := cardTop + int32(row)*(cardH+gap)
		r := RECT{mx, my, mx + cw, my + cardH}
		roundRect(hdc, r, 14, rgb(255, 255, 255), rgb(214, 223, 219))
		drawTextFont(hdc, m.label, RECT{r.Left + 13, r.Top + 12, r.Right - 10, r.Top + 31}, a.fontSmall, rgb(90, 108, 112), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
		drawTextFont(hdc, fmt.Sprint(m.value), RECT{r.Left + 13, r.Top + 30, r.Right - 10, r.Top + 59}, a.fontHeading, rgb(14, 62, 61), DT_LEFT|DT_SINGLELINE)
		drawTextFont(hdc, m.note, RECT{r.Left + 13, r.Top + 60, r.Right - 10, r.Bottom - 6}, a.fontSmall, rgb(91, 108, 111), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
	}
	if a.processing() {
		progressTop := cardTop + int32(rows)*(cardH+gap) + 6
		a.drawProgress(hdc, RECT{x, progressTop, right, progressTop + 62})
	}
}

type metric struct {
	label string
	value int
	note  string
}

func (a *application) metrics() []metric {
	readable, images, attention, quarantine := 0, 0, 0, 0
	for _, e := range a.view.Evidence {
		if e.Readable {
			readable++
		}
		if e.Image != nil {
			images++
		}
		if len(e.Warnings) > 0 {
			attention++
		}
		if e.Status == "Quarantined" {
			quarantine++
		}
	}
	return []metric{{"MATTERS", len(a.view.Matters), "separate workstreams"}, {"EVIDENCE", len(a.view.Evidence), "encrypted originals"}, {"READABLE", readable, "source-backed text"}, {"IMAGES", images, "locally assessed"}, {"ATTENTION", attention, "review warnings"}, {"QUARANTINED", quarantine, "not auto-read"}}
}

func (a *application) drawEvidence(hdc uintptr, rc RECT) {
	x := int32(300)
	right := rc.Right - 25
	drawTextFont(hdc, "EVIDENCE INTAKE STUDIO", RECT{x, 94, right, 120}, a.fontLabel, rgb(44, 104, 101), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, "Preserve, identify and inspect evidence", RECT{x, 122, right, 160}, a.fontHeading, rgb(20, 61, 59), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, "Files are streaming-hashed, encrypted, type-checked and assessed for page boundaries, skew, glare and reading quality.", RECT{x, 165, right - 240, 198}, a.fontBody, rgb(87, 106, 109), DT_LEFT|DT_SINGLELINE)
	a.drawButton(hdc, "add", "+  Add files", RECT{right - 180, 118, right, 164}, true)
	a.drawButton(hdc, "addFolder", "+  Add folder", RECT{right - 375, 118, right - 195, 164}, false)
	a.drawButton(hdc, "pasteImage", "Paste image", RECT{right - 545, 118, right - 390, 164}, false)
	if a.processing() {
		a.drawProgress(hdc, RECT{x, 210, right, 274})
	}
	y := int32(292)
	if !a.processing() {
		y = 218
	}
	listRight := right - 365
	drawTextFont(hdc, "PRESERVED ITEMS", RECT{x, y, listRight, y + 26}, a.fontLabel, rgb(65, 91, 93), DT_LEFT|DT_SINGLELINE)
	y += 34
	if len(a.view.Evidence) == 0 {
		r := RECT{x, y, listRight, y + 180}
		roundRect(hdc, r, 16, rgb(255, 255, 255), rgb(214, 223, 219))
		drawTextFont(hdc, "No evidence has been added yet.\r\n\r\nUse Add files, Ctrl+U or drag files onto this window.", RECT{r.Left + 24, r.Top + 28, r.Right - 24, r.Bottom - 20}, a.fontBody, rgb(70, 94, 96), DT_LEFT|DT_WORDBREAK)
	} else {
		rowH := int32(76)
		for i, e := range a.view.Evidence {
			if y+rowH > rc.Bottom-25 {
				break
			}
			r := RECT{x, y, listRight, y + rowH}
			bg := rgb(255, 255, 255)
			border := rgb(214, 223, 219)
			if i == a.selected {
				bg = rgb(222, 243, 238)
				border = rgb(12, 105, 102)
			}
			roundRect(hdc, r, 13, bg, border)
			drawTextFont(hdc, e.SafeName, RECT{r.Left + 18, r.Top + 12, r.Right - 20, r.Top + 34}, a.fontLabel, rgb(22, 55, 57), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
			meta := fmt.Sprintf("%s · %s · SHA-256 %s", strings.ToUpper(e.DetectedType), eco.HumanBytes(e.Size), shortHash(e.SHA256))
			drawTextFont(hdc, meta, RECT{r.Left + 18, r.Top + 38, r.Right - 20, r.Top + 58}, a.fontSmall, rgb(82, 102, 105), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
			drawTextFont(hdc, e.Status, RECT{r.Right - 160, r.Top + 8, r.Right - 14, r.Top + 30}, a.fontSmall, statusColor(e.Status), DT_RIGHT|DT_SINGLELINE)
			a.evidenceRows = append(a.evidenceRows, hitRegion{Rect: r, Index: i})
			y += rowH + 9
		}
	}
	detail := RECT{listRight + 20, 218, right, rc.Bottom - 25}
	roundRect(hdc, detail, 16, rgb(255, 255, 255), rgb(214, 223, 219))
	a.drawEvidenceDetail(hdc, detail)
}

func (a *application) drawEvidenceDetail(hdc uintptr, r RECT) {
	if len(a.view.Evidence) == 0 {
		return
	}
	if a.selected < 0 || a.selected >= len(a.view.Evidence) {
		a.selected = 0
	}
	e := a.view.Evidence[a.selected]
	drawTextFont(hdc, "SOURCE INSPECTOR", RECT{r.Left + 22, r.Top + 20, r.Right - 20, r.Top + 43}, a.fontLabel, rgb(46, 104, 101), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, e.SafeName, RECT{r.Left + 22, r.Top + 52, r.Right - 20, r.Top + 90}, a.fontHeading, rgb(19, 61, 60), DT_LEFT|DT_WORDBREAK)
	y := r.Top + 104
	lines := []string{fmt.Sprintf("Detected type: %s", e.DetectedType), fmt.Sprintf("Original size: %s", eco.HumanBytes(e.Size)), fmt.Sprintf("SHA-256: %s", e.SHA256), fmt.Sprintf("Readable source segments: %d", len(e.Segments)), fmt.Sprintf("State: %s", e.Status)}
	if e.NearDuplicateOf != "" {
		lines = append(lines, "Visual relationship: possible near duplicate — review required")
	}
	for _, s := range lines {
		drawTextFont(hdc, s, RECT{r.Left + 22, y, r.Right - 22, y + 24}, a.fontSmall, rgb(73, 95, 98), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
		y += 26
	}
	if e.Image != nil {
		imgA := e.Image
		y += 8
		drawTextFont(hdc, "LOCAL IMAGE ASSESSMENT", RECT{r.Left + 22, y, r.Right - 22, y + 22}, a.fontLabel, rgb(46, 104, 101), DT_LEFT|DT_SINGLELINE)
		y += 28
		cropText := "No confident page-boundary suggestion"
		if imgA.SuggestedCrop != nil {
			cropText = fmt.Sprintf("Page-boundary suggestion %.0f%% confidence", imgA.SuggestedCrop.Confidence*100)
		}
		summary := fmt.Sprintf("%d × %d pixels · %.1f MP · %s\r\nBrightness %.0f/255 · Contrast %.0f · Blur %.0f · Edge density %.2f\r\nSkew correction %.1f° (%.0f%% confidence) · Glare %.1f%% · Lighting imbalance %.0f\r\n%s · %s", imgA.Width, imgA.Height, imgA.Megapixels, imgA.Orientation, imgA.Brightness, imgA.Contrast, imgA.BlurVariance, imgA.EdgeDensity, imgA.SkewCorrectionDegrees, imgA.SkewConfidence*100, imgA.GlareRatio*100, imgA.ShadowImbalance, cropText, imgA.QualityLabel)
		drawTextFont(hdc, summary, RECT{r.Left + 22, y, r.Right - 22, y + 112}, a.fontSmall, rgb(71, 93, 96), DT_LEFT|DT_WORDBREAK)
		y += 118
	}
	if len(e.Warnings) > 0 {
		drawTextFont(hdc, "REVIEW WARNINGS", RECT{r.Left + 22, y, r.Right - 22, y + 22}, a.fontLabel, rgb(130, 88, 0), DT_LEFT|DT_SINGLELINE)
		y += 27
		for i, w := range e.Warnings {
			if i >= 3 {
				break
			}
			drawTextFont(hdc, "• "+w, RECT{r.Left + 22, y, r.Right - 22, y + 48}, a.fontSmall, rgb(116, 79, 0), DT_LEFT|DT_WORDBREAK)
			y += 50
		}
	}
	if y < r.Bottom-80 {
		a.drawButton(hdc, "preview", "Open native preview", RECT{r.Left + 22, r.Bottom - 64, r.Left + 198, r.Bottom - 22}, true)
		if e.Image != nil {
			a.drawButton(hdc, "rotate", "Rotate 90°", RECT{r.Left + 210, r.Bottom - 64, r.Right - 22, r.Bottom - 22}, false)
		}
	}
}

func (a *application) drawAsk(hdc uintptr, rc RECT) {
	x := int32(300)
	right := rc.Right - 25
	drawTextFont(hdc, "ASK ECO", RECT{x, 94, right, 120}, a.fontLabel, rgb(44, 104, 101), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, "Evidence conversation", RECT{x, 122, right, 158}, a.fontHeading, rgb(20, 61, 59), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, "Local deterministic retrieval • no model download • no cloud • coordinate-bearing OCR sources are accepted only through the validated OCR receipt gate", RECT{x, 160, right, 190}, a.fontBody, rgb(83, 104, 107), DT_LEFT|DT_SINGLELINE)

	inspectorLeft := right - 330
	r := RECT{inspectorLeft, 238, right, rc.Bottom - 58}
	roundRect(hdc, r, 15, rgb(246, 250, 248), rgb(204, 218, 213))
	drawTextFont(hdc, "SOURCE INSPECTOR", RECT{r.Left + 18, r.Top + 14, r.Right - 18, r.Top + 38}, a.fontLabel, rgb(34, 103, 99), DT_LEFT|DT_SINGLELINE)
	a.mu.Lock()
	last := a.lastQuestion
	a.mu.Unlock()
	if len(last.Citations) == 0 {
		drawTextFont(hdc, "Ask a supported question. ECO will list the exact source passages used in its answer here. The coordinate-bearing OCR source model is implemented, but an approved bundled OCR engine is not yet included in N2 P1.", RECT{r.Left + 18, r.Top + 50, r.Right - 18, r.Bottom - 18}, a.fontSmall, rgb(75, 97, 99), DT_LEFT|DT_WORDBREAK)
	} else {
		y := r.Top + 50
		for i, c := range last.Citations {
			if i >= 5 || y+82 > r.Bottom-10 {
				break
			}
			cr := RECT{r.Left + 12, y, r.Right - 12, y + 76}
			roundRect(hdc, cr, 11, rgb(255, 255, 255), rgb(192, 214, 207))
			drawTextFont(hdc, c.Label, RECT{cr.Left + 12, cr.Top + 9, cr.Right - 12, cr.Top + 29}, a.fontLabel, rgb(24, 84, 80), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
			drawTextFont(hdc, truncateUI(c.Quote, 150), RECT{cr.Left + 12, cr.Top + 32, cr.Right - 12, cr.Bottom - 8}, a.fontSmall, rgb(70, 91, 93), DT_LEFT|DT_WORDBREAK)
			a.citationRows = append(a.citationRows, hitRegion{Name: "citation", Rect: cr, Index: i})
			y += 86
		}
	}
	receipt := "Current mode: Built-in source-backed evidence engine"
	if last.ReceiptID != "" {
		receipt = fmt.Sprintf("Answer receipt %s · considered %d evidence items · ranked %d segments · excluded %d suspicious and %d low-confidence OCR segments", last.ReceiptID, last.EvidenceConsidered, last.RetrievedSegments, last.SuspiciousSourcesExcluded, last.LowConfidenceSourcesExcluded)
	}
	drawTextFont(hdc, receipt, RECT{x, rc.Bottom - 45, right, rc.Bottom - 20}, a.fontSmall, rgb(25, 105, 102), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
}

func (a *application) drawMatters(hdc uintptr, rc RECT) {
	x := int32(300)
	right := rc.Right - 25
	drawTextFont(hdc, "MATTERS", RECT{x, 94, right, 120}, a.fontLabel, rgb(44, 104, 101), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, "Matter command centre", RECT{x, 122, right, 158}, a.fontHeading, rgb(20, 61, 59), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, "Separate evidence workstreams without duplicating or changing the encrypted originals.", RECT{x, 164, right - 220, 194}, a.fontBody, rgb(83, 104, 107), DT_LEFT|DT_SINGLELINE)
	a.drawButton(hdc, "newMatter", "+  New matter", RECT{right - 175, 118, right, 164}, true)
	y := 220
	if len(a.view.Matters) == 0 {
		r := RECT{x, int32(y), right, int32(y + 170)}
		roundRect(hdc, r, 16, rgb(255, 255, 255), rgb(214, 223, 219))
		drawTextFont(hdc, "No Matter has been created yet.\r\n\r\nCreate a Matter to record an objective, current next action and selected evidence scope.", RECT{r.Left + 25, r.Top + 28, r.Right - 25, r.Bottom - 20}, a.fontBody, rgb(72, 94, 97), DT_LEFT|DT_WORDBREAK)
		return
	}
	for i, m := range a.view.Matters {
		if i >= 8 {
			break
		}
		r := RECT{x, int32(y), right, int32(y + 96)}
		roundRect(hdc, r, 15, rgb(255, 255, 255), rgb(214, 223, 219))
		drawTextFont(hdc, m.Title, RECT{r.Left + 20, r.Top + 14, r.Right - 180, r.Top + 40}, a.fontLabel, rgb(20, 60, 59), DT_LEFT|DT_SINGLELINE)
		drawTextFont(hdc, "Status: "+m.Status+"  ·  Evidence: "+fmt.Sprint(len(m.EvidenceIDs)), RECT{r.Left + 20, r.Top + 46, r.Right - 180, r.Top + 68}, a.fontSmall, rgb(82, 102, 105), DT_LEFT|DT_SINGLELINE)
		drawTextFont(hdc, "Next: "+emptyFallback(m.NextAction, "Not yet set"), RECT{r.Left + 20, r.Top + 70, r.Right - 20, r.Bottom - 8}, a.fontSmall, rgb(62, 89, 91), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
		y += 110
	}
}

func (a *application) drawReview(hdc uintptr, rc RECT) {
	x := int32(300)
	right := rc.Right - 25
	drawTextFont(hdc, "REVIEW", RECT{x, 94, right, 120}, a.fontLabel, rgb(44, 104, 101), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, "Items needing human attention", RECT{x, 122, right, 158}, a.fontHeading, rgb(20, 61, 59), DT_LEFT|DT_SINGLELINE)
	y := 205
	count := 0
	for _, e := range a.view.Evidence {
		for _, w := range e.Warnings {
			if y > int(rc.Bottom)-80 {
				break
			}
			r := RECT{x, int32(y), right, int32(y + 68)}
			roundRect(hdc, r, 13, rgb(255, 248, 230), rgb(232, 199, 118))
			drawTextFont(hdc, e.SafeName, RECT{r.Left + 18, r.Top + 10, r.Right - 20, r.Top + 31}, a.fontLabel, rgb(91, 62, 0), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
			drawTextFont(hdc, w, RECT{r.Left + 18, r.Top + 34, r.Right - 20, r.Bottom - 8}, a.fontSmall, rgb(113, 78, 0), DT_LEFT|DT_WORDBREAK)
			y += 78
			count++
		}
	}
	if count == 0 {
		r := RECT{x, 205, right, 365}
		roundRect(hdc, r, 16, rgb(235, 248, 239), rgb(171, 217, 184))
		drawTextFont(hdc, "Nothing currently requires review.\r\n\r\nECO will place file-type mismatches, quarantined program content and image-quality warnings here.", RECT{r.Left + 25, r.Top + 28, r.Right - 25, r.Bottom - 20}, a.fontBody, rgb(39, 101, 59), DT_LEFT|DT_WORDBREAK)
	}
}

func (a *application) drawChanges(hdc uintptr, rc RECT) {
	x := int32(300)
	right := rc.Right - 25
	drawTextFont(hdc, "CHANGES", RECT{x, 94, right, 120}, a.fontLabel, rgb(44, 104, 101), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, "Hash-chained local activity record", RECT{x, 122, right, 158}, a.fontHeading, rgb(20, 61, 59), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, "Each new record includes the preceding record hash. This provides tamper-evidence, not a claim that the device itself cannot be altered.", RECT{x, 164, right, 205}, a.fontBody, rgb(83, 104, 107), DT_LEFT|DT_WORDBREAK)
	y := 220
	for i, c := range a.view.Changes {
		if i >= 10 || y > int(rc.Bottom)-70 {
			break
		}
		r := RECT{x, int32(y), right, int32(y + 62)}
		roundRect(hdc, r, 12, rgb(255, 255, 255), rgb(214, 223, 219))
		drawTextFont(hdc, c.Summary, RECT{r.Left + 18, r.Top + 9, r.Right - 240, r.Top + 31}, a.fontLabel, rgb(24, 59, 60), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
		drawTextFont(hdc, c.At.Local().Format("02 Jan 2006 15:04:05")+" · "+c.Actor+" · "+c.Type, RECT{r.Left + 18, r.Top + 34, r.Right - 20, r.Bottom - 7}, a.fontSmall, rgb(83, 103, 106), DT_LEFT|DT_SINGLELINE)
		drawTextFont(hdc, shortHash(c.Hash), RECT{r.Right - 220, r.Top + 9, r.Right - 18, r.Top + 31}, a.fontSmall, rgb(41, 105, 102), DT_RIGHT|DT_SINGLELINE)
		y += 72
	}
}

func (a *application) drawTrust(hdc uintptr, rc RECT) {
	x := int32(300)
	right := rc.Right - 25
	drawTextFont(hdc, "TRUST & SETTINGS", RECT{x, 94, right, 120}, a.fontLabel, rgb(44, 104, 101), DT_LEFT|DT_SINGLELINE)
	drawTextFont(hdc, "Verifiable local protection", RECT{x, 122, right, 158}, a.fontHeading, rgb(20, 61, 59), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
	workspaceCard := RECT{x, 164, right, 226}
	roundRect(hdc, workspaceCard, 14, rgb(232, 247, 243), rgb(145, 205, 194))
	drawTextFont(hdc, a.workspace.Identity.Name+" • "+string(a.workspace.Disposition), RECT{x + 16, 173, right - 16, 195}, a.fontLabel, rgb(17, 91, 86), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
	drawTextFont(hdc, "Identity: "+a.workspace.Identity.ID+" • Workspace format "+fmt.Sprint(a.workspace.Compatibility.WorkspaceSchema)+" • "+a.workspace.Compatibility.Message, RECT{x + 16, 195, right - 16, 213}, a.fontSmall, rgb(49, 82, 83), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
	drawTextFont(hdc, "Path: "+a.workspace.Path, RECT{x + 16, 211, right - 16, 224}, a.fontSmall, rgb(49, 82, 83), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
	facts := []struct {
		title, body string
		good        bool
	}{{"Native desktop window", "No Brave, Chromium, WebView, localhost or Python runtime.", true}, {"Encrypted casework content", "Evidence, conversations, settings, workspace names and creation details are encrypted. Minimal routing IDs and action audit fields remain plaintext.", true}, {"Local key protection", "Windows DPAPI protects the random vault master key for this Windows account.", true}, {"Network design", "This executable imports no HTTP package and exposes no listening socket.", true}, {"Local evidence intelligence", "Local retrieval and intent classification return extractive source-backed passages.", true}, {"Transactional restore", "Backups are authenticated and restored into a separate encrypted staging vault before activation.", true}, {"Automatic OCR", "Not bundled in N1 P3. Images are preserved and assessed, but image wording is not claimed as read.", false}, {"Generative local model", "Not bundled in N1 P3. ECO does not present deterministic retrieval as a language model.", false}}
	y := int32(236)
	gap := int32(14)
	cardW := (right - x - gap) / 2
	cardH := int32(72)
	rowStep := cardH + 8
	for i, f := range facts {
		col := i % 2
		row := i / 2
		l := x + int32(col)*(cardW+gap)
		t := y + int32(row)*rowStep
		r := RECT{l, t, l + cardW, t + cardH}
		bg, border := rgb(235, 248, 239), rgb(171, 217, 184)
		if !f.good {
			bg, border = rgb(255, 247, 225), rgb(228, 196, 112)
		}
		roundRect(hdc, r, 14, bg, border)
		drawTextFont(hdc, f.title, RECT{r.Left + 15, r.Top + 10, r.Right - 15, r.Top + 31}, a.fontLabel, rgb(25, 67, 65), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
		drawTextFont(hdc, f.body, RECT{r.Left + 15, r.Top + 31, r.Right - 15, r.Bottom - 5}, a.fontSmall, rgb(66, 90, 92), DT_LEFT|DT_WORDBREAK)
	}

	buttonTop := rc.Bottom - 108
	if buttonTop < y+4*rowStep+8 {
		buttonTop = y + 4*rowStep + 8
	}
	gapB := int32(10)
	w4 := (right - x - gapB*3) / 4
	buttons := []struct {
		name, label string
		primary     bool
	}{{"verify", "Verify evidence", true}, {"openVault", "Open workspace folder", false}, {"lowSensory", toggleLabel("Low Sensory", a.view.Settings.LowSensory), false}, {"newWorkspace", "Create new workspace", false}, {"backup", "Create backup", true}, {"restore", "Restore backup", false}, {"openWorkspace", "Open / migrate", false}, {"resetWorkspace", "Reset selected", false}}
	for i, button := range buttons {
		col := i % 4
		row := i / 4
		left := x + int32(col)*(w4+gapB)
		a.drawButton(hdc, button.name, button.label, RECT{left, buttonTop + int32(row)*52, left + w4, buttonTop + int32(row)*52 + 42}, button.primary)
	}
	if a.processing() {
		a.drawProgress(hdc, RECT{x, buttonTop - 68, right, buttonTop - 10})
	}
}

func (a *application) drawProgress(hdc uintptr, r RECT) {
	a.mu.Lock()
	p := a.progress
	a.mu.Unlock()
	roundRect(hdc, r, 13, rgb(232, 247, 243), rgb(145, 205, 194))
	drawTextFont(hdc, "PROCESSING LOCALLY", RECT{r.Left + 18, r.Top + 11, r.Right - 20, r.Top + 31}, a.fontLabel, rgb(17, 104, 99), DT_LEFT|DT_SINGLELINE)
	text := p.Stage
	if p.Name != "" {
		text += " — " + p.Name
	}
	drawTextFont(hdc, text, RECT{r.Left + 18, r.Top + 34, r.Right - 20, r.Bottom - 8}, a.fontSmall, rgb(48, 83, 84), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
}

func (a *application) processing() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.importing
}

func (a *application) drawButton(hdc uintptr, name, label string, r RECT, primary bool) {
	bg, fg, border := rgb(255, 255, 255), rgb(10, 83, 82), rgb(205, 219, 214)
	if primary {
		bg, fg, border = rgb(10, 105, 102), rgb(255, 255, 255), rgb(10, 105, 102)
	}
	roundRect(hdc, r, 11, bg, border)
	drawTextFont(hdc, label, r, a.fontLabel, fg, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
	a.buttons[name] = r
}

func (a *application) handleClick(x, y int32, double bool) {
	for _, n := range a.nav {
		if pointIn(n.Rect, x, y) {
			a.setPage(n.Name)
			return
		}
	}
	for name, r := range a.buttons {
		if pointIn(r, x, y) {
			switch name {
			case "add":
				a.chooseFiles()
			case "addFolder":
				a.chooseFolder()
			case "pasteImage":
				a.pasteClipboardImage()
			case "goAsk":
				a.setPage("ask")
			case "goMatters":
				a.setPage("matters")
			case "preview":
				a.openSelectedPreview()
			case "rotate":
				a.rotateSelected()
			case "verify":
				a.verifyEvidence()
			case "openVault":
				exec.Command("explorer.exe", a.vault.Root).Start()
			case "lowSensory":
				go func() {
					_, _ = a.vault.ToggleLowSensory()
					procPostMessageW.Call(a.hwnd, msgRefresh, 0, 0)
				}()
			case "newMatter":
				a.createQuickMatter()
			case "backup":
				a.createBackup()
			case "restore":
				a.restoreBackup()
			case "newWorkspace":
				a.createWorkspace()
			case "openWorkspace":
				a.openWorkspace()
			case "resetWorkspace":
				a.resetWorkspace()
			}
			return
		}
	}
	if a.page == "ask" {
		for _, hr := range a.citationRows {
			if pointIn(hr.Rect, x, y) {
				a.openCitation(hr.Index)
				return
			}
		}
	}
	if a.page == "evidence" {
		for _, hr := range a.evidenceRows {
			if pointIn(hr.Rect, x, y) {
				a.selected = hr.Index
				invalidate(a.hwnd)
				if double {
					a.openSelectedPreview()
				}
				return
			}
		}
	}
}

func (a *application) handleGlobalShortcut(vk uint32) bool {
	ctrl, _, _ := procGetKeyState.Call(VK_CONTROL)
	shift, _, _ := procGetKeyState.Call(VK_SHIFT)
	alt, _, _ := procGetKeyState.Call(0x12) // VK_MENU
	ctrlDown := int16(ctrl&0xffff) < 0
	shiftDown := int16(shift&0xffff) < 0
	altDown := int16(alt&0xffff) < 0
	if altDown && vk >= '1' && vk <= '7' {
		pages := []string{"home", "evidence", "ask", "matters", "review", "changes", "trust"}
		a.setPage(pages[int(vk-'1')])
		return true
	}
	if ctrlDown {
		switch vk {
		case 'U':
			a.chooseFiles()
			return true
		case 'V':
			focus, _, _ := procGetFocus.Call()
			if shiftDown && focus != a.questionEdit && focus != a.answerEdit {
				a.pasteClipboardImage()
				return true
			}
			return false
		case 'B':
			a.createBackup()
			return true
		case 'R':
			a.restoreBackup()
			return true
		}
	}
	if vk == VK_F1 {
		messageBox(a.hwnd, "ECO help", "Ctrl+U: add evidence files\r\nCtrl+Shift+V: paste a clipboard image\r\nCtrl+B: create encrypted backup\r\nCtrl+R: restore encrypted backup safely\r\nAlt+1..7: change workspace\r\nUp/Down: select evidence\r\nEnter: open selected preview\r\n\r\nImage preview: R rotate 90°, C auto-crop suggestion, D deskew suggestion, O original colour, G greyscale, H fixed high contrast, A adaptive reading mode, Q quality report, +/− zoom, Esc close.", MB_OK|MB_ICONINFORMATION)
		return true
	}
	return false
}

func (a *application) handleKey(vk uint32) {
	if a.handleGlobalShortcut(vk) {
		return
	}
	if a.page == "evidence" && len(a.view.Evidence) > 0 {
		switch vk {
		case VK_UP:
			if a.selected > 0 {
				a.selected--
				invalidate(a.hwnd)
			}
		case VK_DOWN:
			if a.selected+1 < len(a.view.Evidence) {
				a.selected++
				invalidate(a.hwnd)
			}
		case VK_HOME:
			a.selected = 0
			invalidate(a.hwnd)
		case VK_END:
			a.selected = len(a.view.Evidence) - 1
			invalidate(a.hwnd)
		case VK_RETURN:
			a.openSelectedPreview()
		}
	}
}

func (a *application) setPage(p string) {
	a.page = p
	var rc RECT
	procGetClientRect.Call(a.hwnd, uintptr(unsafe.Pointer(&rc)))
	a.layoutControls(rc.Right, rc.Bottom)
	invalidate(a.hwnd)
	if p == "ask" {
		procSetFocus.Call(a.questionEdit)
	}
}

func (a *application) chooseFiles() {
	paths, err := openFileDialog(a.hwnd)
	if err != nil {
		messageBox(a.hwnd, "Could not open file picker", err.Error(), MB_OK|MB_ICONERROR)
		return
	}
	if len(paths) > 0 {
		a.beginImport(paths)
	}
}

func (a *application) chooseFolder() {
	path := openFolderDialog(a.hwnd)
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
	a.progress = eco.ImportProgress{Name: filepath.Base(path), Stage: "Listing folder safely"}
	a.mu.Unlock()
	a.setPage("evidence")
	go func() {
		paths := make([]string, 0, 256)
		_ = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if p != path && d.Type()&os.ModeSymlink != 0 {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Type().IsRegular() {
				paths = append(paths, p)
				if len(paths)%100 == 0 {
					a.mu.Lock()
					a.progress = eco.ImportProgress{Name: filepath.Base(path), Stage: fmt.Sprintf("Listing folder safely — %d files found", len(paths))}
					a.mu.Unlock()
					procPostMessageW.Call(a.hwnd, msgImportProgress, 0, 0)
				}
			}
			if len(paths) >= 10000 {
				return filepath.SkipAll
			}
			return nil
		})
		if len(paths) == 0 {
			a.mu.Lock()
			a.lastErr = "The selected folder did not contain regular files."
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportError, 0, 0)
			return
		}
		a.runImport(paths)
	}()
}

func (a *application) pasteClipboardImage() {
	a.mu.Lock()
	if a.importing {
		a.mu.Unlock()
		messageBox(a.hwnd, "ECO is already processing", "Wait for the current local task to finish.", MB_OK|MB_ICONINFORMATION)
		return
	}
	a.importing = true
	a.progress = eco.ImportProgress{Name: "Clipboard image", Stage: "Reading Windows clipboard"}
	a.mu.Unlock()
	a.setPage("evidence")
	go func() {
		data, err := clipboardBMP()
		if err != nil {
			a.mu.Lock()
			a.lastErr = err.Error()
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportError, 0, 0)
			return
		}
		tmp, err := os.CreateTemp("", "ECO_clipboard_*.bmp")
		if err != nil {
			a.mu.Lock()
			a.lastErr = err.Error()
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportError, 0, 0)
			return
		}
		name := tmp.Name()
		defer os.Remove(name)
		if _, err = tmp.Write(data); err == nil {
			err = tmp.Close()
		} else {
			_ = tmp.Close()
		}
		if err != nil {
			a.mu.Lock()
			a.lastErr = "The temporary local clipboard image could not be written: " + err.Error()
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportError, 0, 0)
			return
		}
		_, _, err = a.vault.ImportFile(name, func(pr eco.ImportProgress) {
			pr.Name = "Pasted clipboard image.bmp"
			a.mu.Lock()
			a.progress = pr
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportProgress, 0, 0)
		})
		if err != nil {
			a.mu.Lock()
			a.lastErr = err.Error()
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportError, 0, 0)
			return
		}
		procPostMessageW.Call(a.hwnd, msgImportDone, 0, 0)
	}()
}

func (a *application) beginImport(paths []string) {
	a.mu.Lock()
	if a.importing {
		a.mu.Unlock()
		messageBox(a.hwnd, "Import already running", "Wait for the current intake queue to finish.", MB_OK|MB_ICONINFORMATION)
		return
	}
	a.importing = true
	a.progress = eco.ImportProgress{Name: fmt.Sprintf("%d selected file(s)", len(paths)), Stage: "Preparing intake queue"}
	a.mu.Unlock()
	a.setPage("evidence")
	go a.runImport(paths)
}

func (a *application) runImport(paths []string) {
	for _, p := range paths {
		_, _, err := a.vault.ImportFile(p, func(pr eco.ImportProgress) {
			a.mu.Lock()
			a.progress = pr
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportProgress, 0, 0)
		})
		if err != nil {
			a.mu.Lock()
			a.lastErr = p + "\r\n\r\n" + err.Error()
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportError, 0, 0)
			return
		}
	}
	procPostMessageW.Call(a.hwnd, msgImportDone, 0, 0)
}

func (a *application) askEvidence() {
	q := strings.TrimSpace(getWindowText(a.questionEdit))
	if q == "" {
		return
	}
	setWindowText(a.answerEdit, "Searching local source passages…")
	go func() {
		rec := a.vault.Ask(q, nil)
		a.mu.Lock()
		a.lastQuestion = rec
		a.mu.Unlock()
		procPostMessageW.Call(a.hwnd, msgAskDone, 0, 0)
	}()
}

func (a *application) openCitation(index int) {
	a.mu.Lock()
	last := a.lastQuestion
	a.mu.Unlock()
	if index < 0 || index >= len(last.Citations) {
		return
	}
	c := last.Citations[index]
	for i, e := range a.view.Evidence {
		if e.ID == c.EvidenceID {
			a.selected = i
			a.mu.Lock()
			if c.Region != nil {
				region := *c.Region
				a.pendingCitationRegion = &region
			} else {
				a.pendingCitationRegion = nil
			}
			a.mu.Unlock()
			note := "Select OK to open the preserved source preview next."
			if c.Region != nil {
				note = "Select OK to open the preserved source preview with the exact OCR region highlighted."
			}
			messageBox(a.hwnd, "Exact cited passage — "+c.Label, c.Quote+"\r\n\r\nECO support classification: "+last.Support+"\r\n\r\n"+note, MB_OK|MB_ICONINFORMATION)
			a.openSelectedPreview()
			return
		}
	}
}

func (a *application) openSelectedPreview() {
	if len(a.view.Evidence) == 0 {
		return
	}
	if a.selected < 0 || a.selected >= len(a.view.Evidence) {
		a.selected = 0
	}
	e := a.view.Evidence[a.selected]
	if e.Image != nil {
		a.mu.Lock()
		if a.importing {
			a.mu.Unlock()
			messageBox(a.hwnd, "ECO is already processing", "Wait for the current local task to finish before opening a large source preview.", MB_OK|MB_ICONINFORMATION)
			return
		}
		a.importing = true
		a.progress = eco.ImportProgress{Name: e.SafeName, Stage: "Preparing encrypted image preview"}
		a.mu.Unlock()
		invalidate(a.hwnd)
		a.mu.Lock()
		var citationRegion *eco.NormalizedRegion
		if a.pendingCitationRegion != nil {
			region := *a.pendingCitationRegion
			citationRegion = &region
			a.pendingCitationRegion = nil
		}
		a.mu.Unlock()
		go func(item eco.EvidenceItem, highlight *eco.NormalizedRegion) {
			data, err := a.vault.ReadEvidence(item.ID, 120*1024*1024)
			if err == nil {
				var img image.Image
				img, _, err = eco.DecodeSupportedImage(data)
				if err == nil {
					assessment := eco.AssessImage(img)
					previewImage := eco.BoundedPreviewImage(img, 8_000_000)
					cropRect, _ := eco.SuggestDocumentBounds(previewImage)
					state := &previewState{itemID: item.ID, title: item.SafeName, original: previewImage, rotation: item.Rotation, mode: "original", zoom: 1, assessment: assessment, cropRect: cropRect, highlight: highlight}
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
		}(e, citationRegion)
		return
	}
	if e.ExtractedText != "" {
		text := e.ExtractedText
		if len([]rune(text)) > 1800 {
			text = string([]rune(text)[:1800]) + "\r\n\r\n[Preview bounded here; Ask ECO can search all extracted source segments.]"
		}
		messageBox(a.hwnd, "Readable source preview — "+e.SafeName, text, MB_OK|MB_ICONINFORMATION)
		return
	}
	messageBox(a.hwnd, "Original preserved", "ECO preserved and encrypted this original, but N2 P1 does not yet have an approved native reader for its contents.", MB_OK|MB_ICONINFORMATION)
}

func (a *application) rotateSelected() {
	if len(a.view.Evidence) == 0 {
		return
	}
	e := a.view.Evidence[a.selected]
	if e.Image == nil {
		return
	}
	go func(id string, rotation int) {
		_ = a.vault.SetRotation(id, rotation)
		procPostMessageW.Call(a.hwnd, msgRefresh, 0, 0)
	}(e.ID, e.Rotation+90)
}

func (a *application) verifyEvidence() {
	go func() {
		alerts := a.vault.VerifyAll(nil)
		a.mu.Lock()
		a.verifyAlerts = append([]string(nil), alerts...)
		a.mu.Unlock()
		procPostMessageW.Call(a.hwnd, msgVerifyDone, 0, 0)
	}()
}

func (a *application) createBackup() {
	path := saveBackupDialog(a.hwnd)
	if path == "" {
		return
	}
	pass, ok := promptText(a.hwnd, "Protect ECO backup", "Enter a backup passphrase of at least 12 characters. ECO does not store it and cannot recover it.", true)
	if !ok {
		return
	}
	confirm, ok := promptText(a.hwnd, "Confirm backup passphrase", "Enter the same passphrase again.", true)
	if !ok {
		return
	}
	if pass != confirm {
		messageBox(a.hwnd, "Passphrases do not match", "The backup was not created.", MB_OK|MB_ICONERROR)
		return
	}
	a.mu.Lock()
	if a.importing {
		a.mu.Unlock()
		messageBox(a.hwnd, "ECO is already processing", "Wait for the current local task to finish.", MB_OK|MB_ICONINFORMATION)
		return
	}
	a.importing = true
	a.progress = eco.ImportProgress{Name: "Encrypted portable backup", Stage: "Preparing"}
	a.mu.Unlock()
	invalidate(a.hwnd)
	go func() {
		r, err := a.vault.CreatePortableBackup(path, pass, func(p eco.BackupProgress) {
			a.mu.Lock()
			a.progress = eco.ImportProgress{Name: p.Name, Stage: p.Stage, Current: p.Current, Total: p.Total}
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportProgress, 0, 0)
		})
		for i := range []byte(pass) {
			_ = i
		}
		pass = ""
		confirm = ""
		if err != nil {
			a.mu.Lock()
			a.lastErr = err.Error()
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportError, 0, 0)
			return
		}
		a.mu.Lock()
		a.backupReceipt = r
		a.mu.Unlock()
		procPostMessageW.Call(a.hwnd, msgBackupDone, 0, 0)
	}()
}

func (a *application) restoreBackup() {
	path := openBackupDialog(a.hwnd)
	if path == "" {
		return
	}
	messageBox(a.hwnd, "Safe staged restore", "ECO will not overwrite the active vault directly. It will authenticate and validate the complete backup in a separate encrypted staging vault, verify every evidence SHA-256, preserve the current vault as a pre-restore checkpoint, and only then activate the restored workspace.", MB_OK|MB_ICONINFORMATION)
	pass, ok := promptText(a.hwnd, "Unlock ECO backup", "Enter the backup passphrase. ECO does not store it.", true)
	if !ok {
		return
	}
	a.mu.Lock()
	if a.importing {
		a.mu.Unlock()
		messageBox(a.hwnd, "ECO is already processing", "Wait for the current local task to finish.", MB_OK|MB_ICONINFORMATION)
		return
	}
	a.importing = true
	a.progress = eco.ImportProgress{Name: filepath.Base(path), Stage: "Authenticating encrypted backup"}
	a.mu.Unlock()
	invalidate(a.hwnd)
	go func() {
		r, err := a.vault.RestorePortableBackup(path, pass, func(p eco.BackupProgress) {
			a.mu.Lock()
			a.progress = eco.ImportProgress{Name: p.Name, Stage: p.Stage, Current: p.Current, Total: p.Total}
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportProgress, 0, 0)
		})
		pass = ""
		if err != nil {
			a.mu.Lock()
			a.lastErr = err.Error()
			a.mu.Unlock()
			procPostMessageW.Call(a.hwnd, msgImportError, 0, 0)
			return
		}
		a.mu.Lock()
		a.restoreReceipt = r
		a.mu.Unlock()
		procPostMessageW.Call(a.hwnd, msgRestoreDone, 0, 0)
	}()
}

func (a *application) createWorkspace() {
	if a.processing() {
		messageBox(a.hwnd, "ECO is already processing", "Wait for the current local task to finish before changing workspace.", MB_OK|MB_ICONINFORMATION)
		return
	}
	messageBox(a.hwnd, "Create a genuinely empty workspace", "Choose a new or empty folder. ECO will not reuse records from the current workspace and will not remove files from a non-empty folder.", MB_OK|MB_ICONINFORMATION)
	path := workspaceFolderDialog(a.hwnd, "Choose a new or empty folder for the ECO workspace")
	if path == "" {
		return
	}
	name, ok := promptText(a.hwnd, "Name the workspace", "Enter a clear name that will identify this workspace in ECO.", false)
	if !ok {
		return
	}
	session, err := a.candidate.CreateWorkspace(path, name)
	if err != nil {
		if a.handleAppliedWorkspaceWarning(session, "Workspace created with an application-state warning", err) {
			return
		}
		messageBox(a.hwnd, "New workspace was not created", err.Error(), MB_OK|MB_ICONERROR)
		return
	}
	a.switchWorkspace(session)
	a.showWorkspaceStatus("New empty workspace created")
}

func (a *application) openWorkspace() {
	if a.processing() {
		messageBox(a.hwnd, "ECO is already processing", "Wait for the current local task to finish before changing workspace.", MB_OK|MB_ICONINFORMATION)
		return
	}
	path := workspaceFolderDialog(a.hwnd, "Deliberately choose an existing ECO workspace")
	if path == "" {
		return
	}
	session, err := a.candidate.OpenWorkspace(path)
	if err == nil {
		a.switchWorkspace(session)
		a.showWorkspaceStatus("Existing workspace reopened")
		return
	}
	if a.handleAppliedWorkspaceWarning(session, "Workspace reopened with an application-state warning", err) {
		return
	}
	var recoveryRequired *eco.RecoveryRequiredError
	if errors.As(err, &recoveryRequired) {
		choice := messageBox(a.hwnd, "Unfinished workspace upgrade", err.Error()+"\r\n\r\nRecover this workspace now? ECO will verify any activated migration or roll back to the preserved original.", MB_YESNO|MB_ICONWARNING)
		if choice != IDYES {
			return
		}
		recovered, receipt, recoverErr := a.candidate.RecoverWorkspace(path)
		if recoverErr != nil {
			if a.handleAppliedWorkspaceWarning(recovered, "Workspace recovered with an application-state warning", recoverErr) {
				return
			}
			messageBox(a.hwnd, "Workspace recovery needs attention", receipt.Message+"\r\n\r\n"+recoverErr.Error(), MB_OK|MB_ICONWARNING)
			return
		}
		a.switchWorkspace(recovered)
		a.showWorkspaceStatus("Workspace recovered")
		return
	}
	var compatibility *eco.CompatibilityError
	if errors.As(err, &compatibility) && compatibility.Report.CanMigrate {
		choice := messageBox(a.hwnd, "Older workspace needs an upgrade", compatibility.Report.Message+"\r\n\r\nUpgrade this selected workspace now? Its original state will be preserved in a separate checkpoint and restored automatically if the upgrade fails.", MB_YESNO|MB_ICONWARNING)
		if choice != IDYES {
			return
		}
		migrated, receipt, migrateErr := a.candidate.MigrateWorkspace(path)
		if migrateErr != nil {
			if a.handleAppliedWorkspaceWarning(migrated, "Workspace migrated with an application-state warning", migrateErr) {
				return
			}
			messageBox(a.hwnd, "Workspace upgrade did not complete", migrateErr.Error(), MB_OK|MB_ICONERROR)
			return
		}
		a.switchWorkspace(migrated)
		messageBox(a.hwnd, "Workspace migrated safely", fmt.Sprintf("Workspace: %s\r\nIdentity: %s\r\nPath: %s\r\nOriginal checkpoint: %s\r\n\r\nThe original checkpoint was kept for rollback.", migrated.Identity.Name, migrated.Identity.ID, migrated.Path, receipt.Checkpoint), MB_OK|MB_ICONINFORMATION)
		return
	}
	messageBox(a.hwnd, "Workspace was not opened", err.Error()+"\r\n\r\nNothing in the selected folder was changed.", MB_OK|MB_ICONERROR)
}

func (a *application) resetWorkspace() {
	if a.processing() {
		messageBox(a.hwnd, "ECO is already processing", "Wait for the current local task to finish before resetting a workspace.", MB_OK|MB_ICONINFORMATION)
		return
	}
	choice := messageBox(a.hwnd, "Reset only this selected workspace", fmt.Sprintf("Workspace: %s\r\nIdentity: %s\r\nPath: %s\r\n\r\nThis will clear this workspace's evidence list, conversations, matters and settings. It will not delete source evidence outside the workspace, unrelated files, another workspace or an arbitrary folder.\r\n\r\nContinue?", a.workspace.Identity.Name, a.workspace.Identity.ID, a.workspace.Path), MB_YESNO|MB_ICONWARNING)
	if choice != IDYES {
		return
	}
	session, receipt, err := a.candidate.ResetCurrentWorkspace()
	if err != nil {
		if a.handleAppliedWorkspaceWarning(session, "Workspace reset with an application-state warning", err) {
			return
		}
		messageBox(a.hwnd, "Workspace was not reset", err.Error(), MB_OK|MB_ICONERROR)
		return
	}
	a.switchWorkspace(session)
	message := fmt.Sprintf("Reset only: %s\r\nIdentity: %s\r\nPath: %s\r\n\r\nCleared %d evidence records, %d matters and %d conversations.", session.Identity.Name, session.Identity.ID, session.Path, receipt.PreviousEvidence, receipt.PreviousMatters, receipt.PreviousQuestions)
	if len(receipt.CleanupWarnings) > 0 {
		message += "\r\n\r\nSome unreferenced encrypted object files could not be removed. They are not part of the reset workspace and will not be displayed."
	}
	messageBox(a.hwnd, "Selected workspace reset", message, MB_OK|MB_ICONINFORMATION)
}

func (a *application) switchWorkspace(session eco.WorkspaceSession) {
	a.workspace = session
	a.vault = session.Vault
	a.page = "home"
	a.selected = 0
	a.refreshView()
	var rc RECT
	procGetClientRect.Call(a.hwnd, uintptr(unsafe.Pointer(&rc)))
	a.layoutControls(rc.Right, rc.Bottom)
	invalidate(a.hwnd)
}

func (a *application) handleAppliedWorkspaceWarning(session eco.WorkspaceSession, title string, err error) bool {
	var warning *eco.CandidateStateWarning
	if session.Vault == nil || !errors.As(err, &warning) {
		return false
	}
	a.switchWorkspace(session)
	message := fmt.Sprintf("Workspace: %s\r\nIdentity: %s\r\nPath: %s\r\n\r\n%s\r\n\r\nThe workspace action completed, but the candidate-specific selection audit needs attention.", session.Identity.Name, session.Identity.ID, session.Path, warning.Error())
	if session.Checkpoint != "" {
		message += "\r\nOriginal checkpoint: " + session.Checkpoint
	}
	messageBox(a.hwnd, title, message, MB_OK|MB_ICONWARNING)
	return true
}

func (a *application) showWorkspaceStatus(title string) {
	messageBox(a.hwnd, title, fmt.Sprintf("Workspace: %s\r\nIdentity: %s\r\nPath: %s\r\nOpened as: %s\r\nCompatibility: %s", a.workspace.Identity.Name, a.workspace.Identity.ID, a.workspace.Path, a.workspace.Disposition, a.workspace.Compatibility.Message), MB_OK|MB_ICONINFORMATION)
}

func (a *application) createQuickMatter() {
	title := fmt.Sprintf("Matter %d", len(a.view.Matters)+1)
	a.mu.Lock()
	if a.importing {
		a.mu.Unlock()
		messageBox(a.hwnd, "ECO is already processing", "Wait for the current local task to finish.", MB_OK|MB_ICONINFORMATION)
		return
	}
	a.importing = true
	a.progress = eco.ImportProgress{Name: title, Stage: "Creating encrypted Matter record"}
	a.mu.Unlock()
	invalidate(a.hwnd)
	go func() {
		m, err := a.vault.CreateMatter(title)
		a.mu.Lock()
		if err != nil {
			a.matterErr = err.Error()
		} else {
			a.matterCreated = m.Title
		}
		a.mu.Unlock()
		procPostMessageW.Call(a.hwnd, msgMatterDone, 0, 0)
	}()
}

func (a *application) reviewCount() int {
	n := 0
	for _, e := range a.view.Evidence {
		n += len(e.Warnings)
	}
	return n
}

func previewWndProc(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		drawPreview(hwnd)
		return 0
	case WM_KEYDOWN:
		previewMu.Lock()
		p := previews[hwnd]
		if p != nil {
			switch uint32(wparam) {
			case VK_ESCAPE:
				previewMu.Unlock()
				procDestroyWindow.Call(hwnd)
				return 0
			case 'R':
				p.rotation = (p.rotation + 90) % 360
				p.rebuild()
				_ = app.vault.SetRotation(p.itemID, p.rotation)
			case 'C':
				if p.highlight != nil {
					break
				}
				if p.cropRect != p.original.Bounds() && !p.cropRect.Empty() {
					p.cropEnabled = !p.cropEnabled
					p.rebuild()
				}
			case 'D':
				if p.highlight != nil {
					break
				}
				if math.Abs(p.assessment.SkewCorrectionDegrees) >= 0.25 && p.assessment.SkewConfidence >= 0.08 {
					p.deskewEnabled = !p.deskewEnabled
					p.rebuild()
				}
			case 'G':
				p.mode = "greyscale"
				p.rebuild()
			case 'H':
				p.mode = "contrast"
				p.rebuild()
			case 'A':
				p.mode = "adaptive"
				p.rebuild()
			case 'O':
				p.mode = "original"
				p.rebuild()
			case 'Q':
				report := previewQualityReport(p)
				previewMu.Unlock()
				messageBox(hwnd, "Local document-vision assessment", report, MB_OK|MB_ICONINFORMATION)
				return 0
			case VK_ADD, VK_OEM_PLUS:
				p.zoom *= 1.15
				if p.zoom > 5 {
					p.zoom = 5
				}
			case VK_SUBTRACT, VK_OEM_MINUS:
				p.zoom /= 1.15
				if p.zoom < 0.2 {
					p.zoom = 0.2
				}
			}
		}
		previewMu.Unlock()
		invalidate(hwnd)
		return 0
	case WM_MOUSEWHEEL:
		delta := int16(hiword(wparam))
		previewMu.Lock()
		if p := previews[hwnd]; p != nil {
			if delta > 0 {
				p.zoom *= 1.1
			} else {
				p.zoom /= 1.1
			}
			if p.zoom < 0.2 {
				p.zoom = 0.2
			}
			if p.zoom > 5 {
				p.zoom = 5
			}
		}
		previewMu.Unlock()
		invalidate(hwnd)
		return 0
	case WM_DESTROY:
		previewMu.Lock()
		delete(previews, hwnd)
		previewMu.Unlock()
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wparam, lparam)
	return r
}

func (p *previewState) rebuild() {
	img := p.original
	if p.cropEnabled && !p.cropRect.Empty() && p.cropRect != p.original.Bounds() {
		img = eco.CropImage(img, p.cropRect)
	}
	if p.deskewEnabled {
		img = eco.RotateImageAngle(img, p.assessment.SkewCorrectionDegrees)
	}
	img = eco.RotateImage(img, p.rotation)
	p.image = eco.ApplyReadingMode(img, p.mode)
	b := p.image.Bounds()
	p.width, p.height = b.Dx(), b.Dy()
	p.pixels = make([]byte, p.width*p.height*4)
	i := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := p.image.At(x, y).RGBA()
			p.pixels[i] = byte(bl >> 8)
			p.pixels[i+1] = byte(g >> 8)
			p.pixels[i+2] = byte(r >> 8)
			p.pixels[i+3] = byte(a >> 8)
			i += 4
		}
	}
}

func previewQualityReport(p *previewState) string {
	a := p.assessment
	crop := "No confident page-boundary suggestion."
	if a.SuggestedCrop != nil {
		crop = fmt.Sprintf("Suggested page crop: %.0f%% confidence.", a.SuggestedCrop.Confidence*100)
	}
	doublePage := "No double-page pattern detected."
	if a.ProbableDoublePage {
		doublePage = "Possible two-page photograph detected; review before OCR."
	}
	warnings := "No quality warnings."
	if len(a.Warnings) > 0 {
		warnings = "• " + strings.Join(a.Warnings, "\r\n• ")
	}
	return fmt.Sprintf("Dimensions: %d × %d (%.1f MP)\r\nOrientation: %s\r\nBrightness: %.0f/255\r\nContrast: %.0f\r\nBlur variance: %.0f\r\nEdge density: %.3f\r\nGlare: %.1f%%\r\nLighting imbalance: %.0f\r\nEdge-dark-content ratio: %.1f%%\r\nSuggested deskew correction: %.1f° (%.0f%% confidence)\r\n%s\r\n%s\r\n\r\n%s", a.Width, a.Height, a.Megapixels, a.Orientation, a.Brightness, a.Contrast, a.BlurVariance, a.EdgeDensity, a.GlareRatio*100, a.ShadowImbalance, a.BorderInkRatio*100, a.SkewCorrectionDegrees, a.SkewConfidence*100, crop, doublePage, warnings)
}

func drawPreview(hwnd uintptr) {
	previewMu.Lock()
	p := previews[hwnd]
	previewMu.Unlock()
	var ps PAINTSTRUCT
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var rc RECT
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	fillRect(hdc, rc, rgb(25, 39, 42))
	fillRect(hdc, RECT{0, 0, rc.Right, 96}, rgb(6, 65, 73))
	if p == nil {
		return
	}
	drawTextFont(hdc, p.title, RECT{24, 0, rc.Right - 24, 42}, app.fontHeading, rgb(255, 255, 255), DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
	cropState := "crop off"
	if p.cropEnabled {
		cropState = "crop on"
	}
	deskewState := "deskew off"
	if p.deskewEnabled {
		deskewState = fmt.Sprintf("deskew %.1f°", p.assessment.SkewCorrectionDegrees)
	}
	highlightState := ""
	if p.highlight != nil {
		highlightState = " · exact source region highlighted"
	}
	drawTextFont(hdc, fmt.Sprintf("%d × %d · rotation %d° · %s · %s · %s · zoom %.0f%%%s", p.width, p.height, p.rotation, p.mode, cropState, deskewState, p.zoom*100, highlightState), RECT{24, 39, rc.Right - 24, 64}, app.fontSmall, rgb(196, 232, 228), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
	drawTextFont(hdc, "R rotate · C auto-crop · D deskew · O original · G greyscale · H fixed contrast · A adaptive · Q quality · +/− zoom · Esc close", RECT{24, 63, rc.Right - 24, 90}, app.fontSmall, rgb(218, 242, 238), DT_LEFT|DT_SINGLELINE|DT_END_ELLIPSIS)
	availW := float64(rc.Right - 40)
	availH := float64(rc.Bottom - 126)
	scale := minFloat(availW/float64(p.width), availH/float64(p.height)) * p.zoom
	dw := int32(float64(p.width) * scale)
	dh := int32(float64(p.height) * scale)
	dx := (rc.Right - dw) / 2
	dy := 108 + (int32(availH)-dh)/2
	info := BITMAPINFO{BmiHeader: BITMAPINFOHEADER{BiSize: uint32(unsafe.Sizeof(BITMAPINFOHEADER{})), BiWidth: int32(p.width), BiHeight: -int32(p.height), BiPlanes: 1, BiBitCount: 32, BiCompression: BI_RGB}}
	if len(p.pixels) > 0 {
		procStretchDIBits.Call(hdc, uintptr(dx), uintptr(dy), uintptr(dw), uintptr(dh), 0, 0, uintptr(p.width), uintptr(p.height), uintptr(unsafe.Pointer(&p.pixels[0])), uintptr(unsafe.Pointer(&info)), DIB_RGB_COLORS, SRCCOPY)
	}
	if p.highlight != nil {
		r := rotateNormalizedRegion(*p.highlight, p.rotation)
		hx0 := dx + int32(math.Round(r.X*float64(dw)))
		hy0 := dy + int32(math.Round(r.Y*float64(dh)))
		hx1 := dx + int32(math.Round((r.X+r.Width)*float64(dw)))
		hy1 := dy + int32(math.Round((r.Y+r.Height)*float64(dh)))
		drawHighlightRect(hdc, RECT{hx0, hy0, hx1, hy1}, rgb(255, 185, 35))
	}
}

func rotateNormalizedRegion(r eco.NormalizedRegion, rotation int) eco.NormalizedRegion {
	rotation = ((rotation % 360) + 360) % 360
	switch rotation {
	case 90:
		return eco.NormalizedRegion{X: 1 - (r.Y + r.Height), Y: r.X, Width: r.Height, Height: r.Width}
	case 180:
		return eco.NormalizedRegion{X: 1 - (r.X + r.Width), Y: 1 - (r.Y + r.Height), Width: r.Width, Height: r.Height}
	case 270:
		return eco.NormalizedRegion{X: r.Y, Y: 1 - (r.X + r.Width), Width: r.Height, Height: r.Width}
	default:
		return r
	}
}

func drawHighlightRect(hdc uintptr, r RECT, colour uintptr) {
	if r.Right <= r.Left || r.Bottom <= r.Top {
		return
	}
	thickness := int32(4)
	fillRect(hdc, RECT{r.Left, r.Top, r.Right, r.Top + thickness}, colour)
	fillRect(hdc, RECT{r.Left, r.Bottom - thickness, r.Right, r.Bottom}, colour)
	fillRect(hdc, RECT{r.Left, r.Top, r.Left + thickness, r.Bottom}, colour)
	fillRect(hdc, RECT{r.Right - thickness, r.Top, r.Right, r.Bottom}, colour)
}

func fillRect(hdc uintptr, r RECT, color uintptr) {
	b, _, _ := procCreateSolidBrush.Call(color)
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), b)
	procDeleteObject.Call(b)
}
func roundRect(hdc uintptr, r RECT, radius int32, bg, border uintptr) {
	b, _, _ := procCreateSolidBrush.Call(bg)
	p, _, _ := procCreatePen.Call(0, 1, border)
	oldB, _, _ := procSelectObject.Call(hdc, b)
	oldP, _, _ := procSelectObject.Call(hdc, p)
	procRoundRect.Call(hdc, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right), uintptr(r.Bottom), uintptr(radius), uintptr(radius))
	procSelectObject.Call(hdc, oldB)
	procSelectObject.Call(hdc, oldP)
	procDeleteObject.Call(b)
	procDeleteObject.Call(p)
}
func drawTextFont(hdc uintptr, text string, r RECT, font, color uintptr, flags uint32) {
	old, _, _ := procSelectObject.Call(hdc, font)
	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, color)
	drawText(hdc, text, &r, flags|DT_NOPREFIX)
	procSelectObject.Call(hdc, old)
}
func pointIn(r RECT, x, y int32) bool {
	return x >= r.Left && x <= r.Right && y >= r.Top && y <= r.Bottom
}
func shortHash(s string) string {
	if len(s) > 16 {
		return s[:8] + "…" + s[len(s)-8:]
	}
	return s
}
func statusColor(s string) uintptr {
	if strings.Contains(strings.ToLower(s), "quarant") || strings.Contains(strings.ToLower(s), "failed") {
		return rgb(155, 52, 52)
	}
	if strings.Contains(strings.ToLower(s), "ready") {
		return rgb(25, 107, 56)
	}
	return rgb(130, 88, 0)
}
func emptyFallback(s, f string) string {
	if strings.TrimSpace(s) == "" {
		return f
	}
	return s
}
func toggleLabel(s string, on bool) string {
	if on {
		return "✓  " + s
	}
	return s
}
func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func openFileDialog(owner uintptr) ([]string, error) {
	buf := make([]uint16, 65536)
	filter := multiString([]string{"All evidence files", "*.*", "All files", "*.*"})
	title := utf16Ptr("Add evidence to ECO — originals will be encrypted locally")
	ofn := OPENFILENAME{LStructSize: uint32(unsafe.Sizeof(OPENFILENAME{})), HwndOwner: owner, LpstrFilter: &filter[0], NFilterIndex: 1, LpstrFile: &buf[0], NMaxFile: uint32(len(buf)), LpstrTitle: title, Flags: OFN_EXPLORER | OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST | OFN_ALLOWMULTISELECT | OFN_HIDEREADONLY}
	r, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return nil, nil
	}
	parts := splitMultiString(buf)
	if len(parts) == 1 {
		return []string{parts[0]}, nil
	}
	dir := parts[0]
	out := make([]string, 0, len(parts)-1)
	for _, n := range parts[1:] {
		out = append(out, filepath.Join(dir, n))
	}
	return out, nil
}
func multiString(parts []string) []uint16 {
	out := []uint16{}
	for _, s := range parts {
		u, _ := syscall.UTF16FromString(s)
		out = append(out, u...)
	}
	out = append(out, 0)
	return out
}
func splitMultiString(buf []uint16) []string {
	parts := []string{}
	start := 0
	for i, v := range buf {
		if v == 0 {
			if i == start {
				break
			}
			parts = append(parts, syscall.UTF16ToString(buf[start:i]))
			start = i + 1
		}
	}
	return parts
}
func droppedPaths(hdrop uintptr) []string {
	count, _, _ := procDragQueryFileW.Call(hdrop, 0xFFFFFFFF, 0, 0)
	out := []string{}
	for i := uintptr(0); i < count; i++ {
		n, _, _ := procDragQueryFileW.Call(hdrop, i, 0, 0)
		buf := make([]uint16, n+1)
		procDragQueryFileW.Call(hdrop, i, uintptr(unsafe.Pointer(&buf[0])), n+1)
		out = append(out, syscall.UTF16ToString(buf))
	}
	procDragFinish.Call(hdrop)
	return out
}

func openFolderDialog(owner uintptr) string {
	return browseFolderDialog(owner, "Add an evidence folder — symbolic links will not be followed")
}

func workspaceFolderDialog(owner uintptr, title string) string {
	return browseFolderDialog(owner, title)
}

func browseFolderDialog(owner uintptr, title string) string {
	display := make([]uint16, 260)
	bi := BROWSEINFO{HwndOwner: owner, PszDisplayName: &display[0], LpszTitle: utf16Ptr(title), UlFlags: BIF_RETURNONLYFSDIRS | BIF_NEWDIALOGSTYLE | BIF_EDITBOX}
	pidl, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return ""
	}
	defer procCoTaskMemFree.Call(pidl)
	path := make([]uint16, 32768)
	r, _, _ := procSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&path[0])))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(path)
}

func clipboardBMP() ([]byte, error) {
	if r, _, _ := procIsClipboardFormatAvailable.Call(CF_DIB); r == 0 {
		return nil, fmt.Errorf("the clipboard does not currently contain a compatible image")
	}
	if r, _, _ := procOpenClipboard.Call(app.hwnd); r == 0 {
		return nil, fmt.Errorf("another application is using the clipboard")
	}
	defer procCloseClipboard.Call()
	h, _, _ := procGetClipboardData.Call(CF_DIB)
	if h == 0 {
		return nil, fmt.Errorf("Windows could not read the clipboard image")
	}
	p, _, _ := procGlobalLock.Call(h)
	if p == 0 {
		return nil, fmt.Errorf("Windows could not lock the clipboard image")
	}
	defer procGlobalUnlock.Call(h)
	sz, _, _ := procGlobalSize.Call(h)
	if sz < 40 || sz > 250*1024*1024 {
		return nil, fmt.Errorf("clipboard image size is unsafe or unsupported")
	}
	dib := make([]byte, sz)
	procRtlMoveMemory.Call(uintptr(unsafe.Pointer(&dib[0])), p, sz)
	headerSize := int(binary.LittleEndian.Uint32(dib[:4]))
	if headerSize < 40 || headerSize > len(dib) {
		return nil, fmt.Errorf("clipboard DIB header is unsupported")
	}
	bits := binary.LittleEndian.Uint16(dib[14:16])
	compression := binary.LittleEndian.Uint32(dib[16:20])
	colors := binary.LittleEndian.Uint32(dib[32:36])
	if compression != 0 || (bits != 24 && bits != 32) {
		return nil, fmt.Errorf("this clipboard image uses a Windows bitmap encoding not yet accepted by N1 P3")
	}
	palette := 0
	if bits <= 8 {
		if colors == 0 {
			colors = 1 << bits
		}
		palette = int(colors) * 4
	}
	offset := 14 + headerSize + palette
	out := make([]byte, 14+len(dib))
	copy(out[:2], []byte("BM"))
	binary.LittleEndian.PutUint32(out[2:6], uint32(len(out)))
	binary.LittleEndian.PutUint32(out[10:14], uint32(offset))
	copy(out[14:], dib)
	return out, nil
}

func promptText(owner uintptr, title, prompt string, secret bool) (string, bool) {
	hw := createWindowEx(0, inputClass, title, WS_CAPTION|WS_SYSMENU, CW_USEDEFAULT, CW_USEDEFAULT, 590, 245, owner, 0, app.hInstance, nil)
	if hw == 0 {
		return "", false
	}
	st := &inputDialogState{hwnd: hw, owner: owner, secret: secret}
	st.prompt = createWindowEx(0, "STATIC", prompt, WS_CHILD|WS_VISIBLE, 22, 22, 535, 58, hw, 0, app.hInstance, nil)
	style := uint32(WS_CHILD | WS_VISIBLE | WS_BORDER | WS_TABSTOP | ES_LEFT | ES_AUTOHSCROLL)
	if secret {
		style |= ES_PASSWORD
	}
	st.edit = createWindowEx(WS_EX_CLIENTEDGE, "EDIT", "", style, 22, 88, 535, 36, hw, 2001, app.hInstance, nil)
	st.okButton = createWindowEx(0, "BUTTON", "OK", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_DEFPUSHBUTTON, 353, 150, 96, 36, hw, 1, app.hInstance, nil)
	st.cancelButton = createWindowEx(0, "BUTTON", "Cancel", WS_CHILD|WS_VISIBLE|WS_TABSTOP|BS_PUSHBUTTON, 461, 150, 96, 36, hw, 2, app.hInstance, nil)
	for _, c := range []uintptr{st.prompt, st.edit, st.okButton, st.cancelButton} {
		procSendMessageW.Call(c, WM_SETFONT, app.fontBody, 1)
	}
	dialogMu.Lock()
	dialogs[hw] = st
	dialogMu.Unlock()
	procEnableWindow.Call(owner, 0)
	showWindow(hw, SW_SHOW)
	procUpdateWindow.Call(hw)
	procSetFocus.Call(st.edit)
	var msg MSG
	for !st.done {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
	procEnableWindow.Call(owner, 1)
	procSetFocus.Call(owner)
	dialogMu.Lock()
	delete(dialogs, hw)
	dialogMu.Unlock()
	return st.value, st.accepted
}

func inputWndProc(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	dialogMu.Lock()
	st := dialogs[hwnd]
	dialogMu.Unlock()
	switch msg {
	case WM_COMMAND:
		if st != nil {
			switch int(loword(wparam)) {
			case 1:
				st.value = getWindowText(st.edit)
				st.accepted = true
				st.done = true
				procDestroyWindow.Call(hwnd)
				return 0
			case 2:
				st.done = true
				procDestroyWindow.Call(hwnd)
				return 0
			}
		}
	case WM_CLOSE:
		if st != nil {
			st.done = true
		}
		procDestroyWindow.Call(hwnd)
		return 0
	case WM_CTLCOLORSTATIC, WM_CTLCOLOREDIT:
		procSetBkColor.Call(wparam, rgb(255, 255, 255))
		procSetTextColor.Call(wparam, rgb(23, 38, 41))
		return app.controlBrush
	case WM_DESTROY:
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wparam, lparam)
	return r
}

func saveBackupDialog(owner uintptr) string {
	buf := make([]uint16, 32768)
	defaultName := "ECO_Backup_" + time.Now().Format("2006-01-02") + ".ecobackup"
	u, _ := syscall.UTF16FromString(defaultName)
	copy(buf, u)
	filter := multiString([]string{"ECO encrypted backup", "*.ecobackup", "All files", "*.*"})
	def := utf16Ptr("ecobackup")
	ofn := OPENFILENAME{LStructSize: uint32(unsafe.Sizeof(OPENFILENAME{})), HwndOwner: owner, LpstrFilter: &filter[0], NFilterIndex: 1, LpstrFile: &buf[0], NMaxFile: uint32(len(buf)), LpstrTitle: utf16Ptr("Create encrypted ECO backup"), LpstrDefExt: def, Flags: OFN_EXPLORER | OFN_PATHMUSTEXIST | OFN_OVERWRITEPROMPT}
	r, _, _ := procGetSaveFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func openBackupDialog(owner uintptr) string {
	buf := make([]uint16, 32768)
	filter := multiString([]string{"ECO encrypted backup", "*.ecobackup", "All files", "*.*"})
	of := OPENFILENAME{LStructSize: uint32(unsafe.Sizeof(OPENFILENAME{})), HwndOwner: owner, LpstrFilter: &filter[0], NFilterIndex: 1, LpstrFile: &buf[0], NMaxFile: uint32(len(buf)), LpstrTitle: utf16Ptr("Restore encrypted ECO backup"), Flags: OFN_EXPLORER | OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST}
	r, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&of)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func truncateUI(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
