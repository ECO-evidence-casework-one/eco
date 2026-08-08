//go:build windows

package main

import (
    "fmt"
    "os"
    "runtime"
    "strconv"
    "syscall"
    "unsafe"
)

const (
    csHRedraw = 0x0002
    csVRedraw = 0x0001

    wsOverlappedWindow = 0x00CF0000
    wsVisible          = 0x10000000
    wsChild            = 0x40000000
    wsTabStop          = 0x00010000
    wsBorder           = 0x00800000
    wsVScroll          = 0x00200000

    esAutoHScroll = 0x0080

    bsPushButton = 0x00000000
    bsFlat       = 0x00008000

    lbsNotify     = 0x0001
    lbsNoIntegral = 0x0100

    wmDestroy        = 0x0002
    wmSize           = 0x0005
    wmCommand        = 0x0111
    wmSetFont        = 0x0030
    wmCtlColorEdit   = 0x0133
    wmCtlColorBtn    = 0x0135
    wmCtlColorStatic = 0x0138

    lbAddString = 0x0180

    swShowDefault = 10
    transparent   = 1

    idNavWorkspace = 1001
    idNavDocuments = 1002
    idReview       = 1101
    idAsk          = 1102
    idSearch       = 1103
    idDocuments    = 1104
    idStatus       = 1105
)

var (
    user32   = syscall.NewLazyDLL("user32.dll")
    gdi32    = syscall.NewLazyDLL("gdi32.dll")
    kernel32 = syscall.NewLazyDLL("kernel32.dll")

    procRegisterClassExW = user32.NewProc("RegisterClassExW")
    procCreateWindowExW  = user32.NewProc("CreateWindowExW")
    procDefWindowProcW   = user32.NewProc("DefWindowProcW")
    procGetMessageW      = user32.NewProc("GetMessageW")
    procTranslateMessage = user32.NewProc("TranslateMessage")
    procDispatchMessageW = user32.NewProc("DispatchMessageW")
    procPostQuitMessage  = user32.NewProc("PostQuitMessage")
    procShowWindow       = user32.NewProc("ShowWindow")
    procUpdateWindow     = user32.NewProc("UpdateWindow")
    procMoveWindow       = user32.NewProc("MoveWindow")
    procSendMessageW     = user32.NewProc("SendMessageW")
    procLoadCursorW      = user32.NewProc("LoadCursorW")
    procSetWindowTextW   = user32.NewProc("SetWindowTextW")

    procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
    procCreateFontW      = gdi32.NewProc("CreateFontW")
    procSetTextColor     = gdi32.NewProc("SetTextColor")
    procSetBkColor       = gdi32.NewProc("SetBkColor")
    procSetBkMode        = gdi32.NewProc("SetBkMode")
    procDeleteObject     = gdi32.NewProc("DeleteObject")

    procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")

    backgroundBrush uintptr
    sidebarBrush    uintptr
    cardBrush       uintptr
    inputBrush      uintptr
    uiFont          uintptr

    sidebarTitle    uintptr
    offlineLabel    uintptr
    navWorkspace    uintptr
    navDocuments    uintptr
    pageTitle       uintptr
    pageSubtitle    uintptr
    evidenceHeading uintptr
    evidenceBody    uintptr
    searchLabel     uintptr
    searchEdit      uintptr
    reviewButton    uintptr
    askButton       uintptr
    documentList    uintptr
    statusLabel     uintptr
)

type point struct {
    X int32
    Y int32
}

type msg struct {
    Hwnd     uintptr
    Message  uint32
    WParam   uintptr
    LParam   uintptr
    Time     uint32
    Pt       point
    LPrivate uint32
}

type wndClassEx struct {
    CbSize        uint32
    Style         uint32
    LpfnWndProc   uintptr
    CbClsExtra    int32
    CbWndExtra    int32
    HInstance     uintptr
    HIcon         uintptr
    HCursor       uintptr
    HbrBackground uintptr
    LpszMenuName  *uint16
    LpszClassName *uint16
    HIconSm       uintptr
}

func rgb(r, g, b byte) uintptr {
    return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16
}

func utf16(s string) *uint16 {
    return syscall.StringToUTF16Ptr(s)
}

func createChild(className, text string, style uintptr, id int, parent uintptr) uintptr {
    h, _, _ := procCreateWindowExW.Call(
        0,
        uintptr(unsafe.Pointer(utf16(className))),
        uintptr(unsafe.Pointer(utf16(text))),
        style|wsChild|wsVisible,
        0, 0, 100, 30,
        parent,
        uintptr(id),
        0,
        0,
    )
    if h == 0 {
        panic("CreateWindowExW failed for " + className + ": " + text)
    }
    if uiFont != 0 {
        procSendMessageW.Call(h, wmSetFont, uiFont, 1)
    }
    return h
}

func setBounds(h uintptr, x, y, width, height int32) {
    if h == 0 {
        return
    }
    procMoveWindow.Call(h, uintptr(x), uintptr(y), uintptr(width), uintptr(height), 1)
}

func loword(v uintptr) uint16 {
    return uint16(v & 0xffff)
}

func mainWndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
    switch message {
    case wmSize:
        width := int32(loword(lParam))
        height := int32(uint16((lParam >> 16) & 0xffff))
        layoutControls(width, height)
        return 0

    case wmCommand:
        switch int(loword(wParam)) {
        case idReview:
            procSetWindowTextW.Call(statusLabel, uintptr(unsafe.Pointer(utf16("Review action invoked — synthetic test only"))))
        case idAsk:
            procSetWindowTextW.Call(statusLabel, uintptr(unsafe.Pointer(utf16("Ask ECO action invoked — synthetic test only"))))
        case idNavWorkspace:
            procSetWindowTextW.Call(statusLabel, uintptr(unsafe.Pointer(utf16("Workspace selected"))))
        case idNavDocuments:
            procSetWindowTextW.Call(statusLabel, uintptr(unsafe.Pointer(utf16("Documents selected"))))
        }
        return 0

    case wmCtlColorStatic:
        hdc := wParam
        child := lParam
        if child == sidebarTitle || child == offlineLabel {
            procSetTextColor.Call(hdc, rgb(238, 249, 247))
            procSetBkMode.Call(hdc, transparent)
            return sidebarBrush
        }
        procSetTextColor.Call(hdc, rgb(9, 48, 55))
        procSetBkMode.Call(hdc, transparent)
        return backgroundBrush

    case wmCtlColorEdit:
        hdc := wParam
        procSetTextColor.Call(hdc, rgb(9, 48, 55))
        procSetBkColor.Call(hdc, rgb(255, 255, 255))
        return inputBrush

    case wmCtlColorBtn:
        return cardBrush

    case wmDestroy:
        procPostQuitMessage.Call(0)
        return 0
    }

    ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
    return ret
}

func layoutControls(width, height int32) {
    if width < 900 {
        width = 900
    }
    if height < 600 {
        height = 600
    }

    const side int32 = 210
    const pad int32 = 28
    contentX := side + pad
    contentW := width - contentX - pad
    if contentW < 500 {
        contentW = 500
    }

    setBounds(sidebarTitle, 24, 28, side-48, 54)
    setBounds(offlineLabel, 24, 82, side-48, 32)
    setBounds(navWorkspace, 18, 150, side-36, 42)
    setBounds(navDocuments, 18, 200, side-36, 42)

    setBounds(pageTitle, contentX, 28, contentW, 42)
    setBounds(pageSubtitle, contentX, 72, contentW, 30)

    leftW := (contentW * 58) / 100
    rightX := contentX + leftW + 20
    rightW := contentW - leftW - 20

    setBounds(evidenceHeading, contentX, 126, leftW, 34)
    setBounds(evidenceBody, contentX, 164, leftW, 54)
    setBounds(searchLabel, contentX, 238, leftW, 28)
    setBounds(searchEdit, contentX, 270, leftW, 38)
    setBounds(reviewButton, contentX, 326, 170, 42)
    setBounds(askButton, contentX+184, 326, 150, 42)

    setBounds(documentList, rightX, 126, rightW, height-226)
    setBounds(statusLabel, contentX, height-62, contentW, 28)
}

func addListItem(list uintptr, text string) {
    procSendMessageW.Call(list, lbAddString, 0, uintptr(unsafe.Pointer(utf16(text))))
}

func findHandleFile() string {
    for i := 1; i < len(os.Args); i++ {
        if os.Args[i] == "--handle-file" && i+1 < len(os.Args) {
            return os.Args[i+1]
        }
        const prefix = "--handle-file="
        if len(os.Args[i]) > len(prefix) && os.Args[i][:len(prefix)] == prefix {
            return os.Args[i][len(prefix):]
        }
    }
    return ""
}

func run() error {
    runtime.LockOSThread()

    hInstance, _, _ := procGetModuleHandleW.Call(0)
    if hInstance == 0 {
        return fmt.Errorf("GetModuleHandleW failed")
    }

    backgroundBrush, _, _ = procCreateSolidBrush.Call(rgb(241, 247, 246))
    sidebarBrush, _, _ = procCreateSolidBrush.Call(rgb(9, 48, 55))
    cardBrush, _, _ = procCreateSolidBrush.Call(rgb(255, 255, 255))
    inputBrush, _, _ = procCreateSolidBrush.Call(rgb(255, 255, 255))

    fontHeight := int32(-18)
    uiFont, _, _ = procCreateFontW.Call(
        uintptr(fontHeight), 0, 0, 0, 400, 0, 0, 0,
        1, 0, 0, 5, 0,
        uintptr(unsafe.Pointer(utf16("Segoe UI"))),
    )

    className := utf16("ECO_NATIVE_A11Y_SPIKE")
    cursor, _, _ := procLoadCursorW.Call(0, 32512)
    wc := wndClassEx{
        CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
        Style:         csHRedraw | csVRedraw,
        LpfnWndProc:   syscall.NewCallback(mainWndProc),
        HInstance:     hInstance,
        HCursor:       cursor,
        HbrBackground: backgroundBrush,
        LpszClassName: className,
    }
    atom, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
    if atom == 0 {
        return fmt.Errorf("RegisterClassExW failed")
    }

    hwnd, _, _ := procCreateWindowExW.Call(
        0,
        uintptr(unsafe.Pointer(className)),
        uintptr(unsafe.Pointer(utf16("Evidence & Casework One — Native Accessibility Spike"))),
        wsOverlappedWindow|wsVisible,
        60, 60, 1180, 760,
        0, 0, hInstance, 0,
    )
    if hwnd == 0 {
        return fmt.Errorf("CreateWindowExW main window failed")
    }

    sidebarTitle = createChild("STATIC", "Evidence & Casework One", 0, 0, hwnd)
    offlineLabel = createChild("STATIC", "OFFLINE • LOCAL", 0, 0, hwnd)
    navWorkspace = createChild("BUTTON", "Workspace", wsTabStop|bsPushButton|bsFlat, idNavWorkspace, hwnd)
    navDocuments = createChild("BUTTON", "Documents", wsTabStop|bsPushButton|bsFlat, idNavDocuments, hwnd)

    pageTitle = createChild("STATIC", "Matter Workspace", 0, 0, hwnd)
    pageSubtitle = createChild("STATIC", "Local evidence, source-backed casework and review", 0, 0, hwnd)
    evidenceHeading = createChild("STATIC", "Evidence overview", 0, 0, hwnd)
    evidenceBody = createChild("STATIC", "12 documents • 4 verified • 2 need review", 0, 0, hwnd)
    searchLabel = createChild("STATIC", "Search this Matter", 0, 0, hwnd)
    searchEdit = createChild("EDIT", "", wsTabStop|wsBorder|esAutoHScroll, idSearch, hwnd)
    reviewButton = createChild("BUTTON", "Review evidence", wsTabStop|bsPushButton, idReview, hwnd)
    askButton = createChild("BUTTON", "Ask ECO", wsTabStop|bsPushButton, idAsk, hwnd)
    documentList = createChild("LISTBOX", "", wsTabStop|wsBorder|wsVScroll|lbsNotify|lbsNoIntegral, idDocuments, hwnd)
    statusLabel = createChild("STATIC", "Ready — synthetic accessibility architecture spike", 0, idStatus, hwnd)

    addListItem(documentList, "2026-08-02 — Council letter.pdf — verified")
    addListItem(documentList, "2026-08-03 — Bank evidence.zip — needs review")
    addListItem(documentList, "2026-08-04 — Case chronology.txt — verified")
    addListItem(documentList, "2026-08-05 — Supporting image.png — indexed")

    layoutControls(1160, 720)
    procShowWindow.Call(hwnd, swShowDefault)
    procUpdateWindow.Call(hwnd)

    if handleFile := findHandleFile(); handleFile != "" {
        if err := os.WriteFile(handleFile, []byte(strconv.FormatUint(uint64(hwnd), 10)), 0600); err != nil {
            return fmt.Errorf("write handle file: %w", err)
        }
    }

    var m msg
    for {
        result, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
        if int32(result) == -1 {
            return fmt.Errorf("GetMessageW failed")
        }
        if result == 0 {
            break
        }
        procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
        procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
    }

    if uiFont != 0 {
        procDeleteObject.Call(uiFont)
    }
    if backgroundBrush != 0 {
        procDeleteObject.Call(backgroundBrush)
    }
    if sidebarBrush != 0 {
        procDeleteObject.Call(sidebarBrush)
    }
    if cardBrush != 0 {
        procDeleteObject.Call(cardBrush)
    }
    if inputBrush != 0 {
        procDeleteObject.Call(inputBrush)
    }
    return nil
}

func main() {
    if err := run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
