//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	CS_HREDRAW           = 0x0002
	CS_VREDRAW           = 0x0001
	CW_USEDEFAULT        = int32(-2147483648)
	WS_OVERLAPPEDWINDOW  = 0x00CF0000
	WS_CAPTION           = 0x00C00000
	WS_SYSMENU           = 0x00080000
	WS_CLIPCHILDREN      = 0x02000000
	WS_CHILD             = 0x40000000
	WS_VISIBLE           = 0x10000000
	WS_BORDER            = 0x00800000
	WS_TABSTOP           = 0x00010000
	WS_VSCROLL           = 0x00200000
	WS_EX_CLIENTEDGE     = 0x00000200
	ES_LEFT              = 0x0000
	ES_MULTILINE         = 0x0004
	ES_AUTOVSCROLL       = 0x0040
	ES_AUTOHSCROLL       = 0x0080
	ES_READONLY          = 0x0800
	ES_PASSWORD          = 0x0020
	BS_PUSHBUTTON        = 0x00000000
	BS_DEFPUSHBUTTON     = 0x00000001
	SW_SHOW              = 5
	SW_HIDE              = 0
	SW_SHOWNORMAL        = 1
	WM_CREATE            = 0x0001
	WM_DESTROY           = 0x0002
	WM_SIZE              = 0x0005
	WM_PAINT             = 0x000F
	WM_CLOSE             = 0x0010
	WM_QUIT              = 0x0012
	WM_ERASEBKGND        = 0x0014
	WM_SETCURSOR         = 0x0020
	WM_GETMINMAXINFO     = 0x0024
	WM_COMMAND           = 0x0111
	WM_KEYDOWN           = 0x0100
	WM_CHAR              = 0x0102
	WM_LBUTTONDOWN       = 0x0201
	WM_LBUTTONDBLCLK     = 0x0203
	WM_MOUSEWHEEL        = 0x020A
	WM_DROPFILES         = 0x0233
	WM_CTLCOLORSTATIC    = 0x0138
	WM_CTLCOLOREDIT      = 0x0133
	WM_DPICHANGED        = 0x02E0
	WM_SETFONT           = 0x0030
	WM_GETTEXT           = 0x000D
	WM_GETTEXTLENGTH     = 0x000E
	WM_SETTEXT           = 0x000C
	WM_APP               = 0x8000
	BN_CLICKED           = 0
	MB_OK                = 0x00000000
	MB_ICONINFORMATION   = 0x00000040
	MB_ICONERROR         = 0x00000010
	MB_ICONWARNING       = 0x00000030
	MB_YESNO             = 0x00000004
	MB_ICONQUESTION      = 0x00000020
	IDYES                = 6
	COLOR_WINDOW         = 5
	IDC_ARROW            = 32512
	DT_LEFT              = 0x00000000
	DT_CENTER            = 0x00000001
	DT_RIGHT             = 0x00000002
	DT_VCENTER           = 0x00000004
	DT_WORDBREAK         = 0x00000010
	DT_SINGLELINE        = 0x00000020
	DT_END_ELLIPSIS      = 0x00008000
	DT_NOPREFIX          = 0x00000800
	TRANSPARENT          = 1
	OPAQUE               = 2
	FW_NORMAL            = 400
	FW_SEMIBOLD          = 600
	FW_BOLD              = 700
	DEFAULT_CHARSET      = 1
	OUT_DEFAULT_PRECIS   = 0
	CLIP_DEFAULT_PRECIS  = 0
	CLEARTYPE_QUALITY    = 5
	DEFAULT_PITCH        = 0
	FF_DONTCARE          = 0
	OFN_EXPLORER         = 0x00080000
	OFN_FILEMUSTEXIST    = 0x00001000
	OFN_PATHMUSTEXIST    = 0x00000800
	OFN_ALLOWMULTISELECT = 0x00000200
	OFN_HIDEREADONLY     = 0x00000004
	OFN_OVERWRITEPROMPT  = 0x00000002
	BI_RGB               = 0
	DIB_RGB_COLORS       = 0
	CF_DIB               = 8
	BIF_RETURNONLYFSDIRS = 0x0001
	BIF_NEWDIALOGSTYLE   = 0x0040
	BIF_EDITBOX          = 0x0010
	SRCCOPY              = 0x00CC0020
	VK_ESCAPE            = 0x1B
	VK_RETURN            = 0x0D
	VK_ADD               = 0x6B
	VK_SUBTRACT          = 0x6D
	VK_OEM_PLUS          = 0xBB
	VK_OEM_MINUS         = 0xBD
	VK_F1                = 0x70
	VK_CONTROL           = 0x11
	VK_SHIFT             = 0x10
	VK_UP                = 0x26
	VK_DOWN              = 0x28
	VK_HOME              = 0x24
	VK_END               = 0x23
	MK_CONTROL           = 0x0008
	LOGPIXELSY           = 90
	SWP_NOZORDER         = 0x0004
	SWP_NOACTIVATE       = 0x0010
)

type POINT struct{ X, Y int32 }
type RECT struct{ Left, Top, Right, Bottom int32 }
type MINMAXINFO struct {
	PtReserved, PtMaxSize, PtMaxPosition, PtMinTrackSize, PtMaxTrackSize POINT
}
type MSG struct {
	Hwnd           uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             POINT
	LPrivate       uint32
}
type WNDCLASSEX struct {
	CbSize                                   uint32
	Style                                    uint32
	LpfnWndProc                              uintptr
	CbClsExtra, CbWndExtra                   int32
	HInstance, HIcon, HCursor, HbrBackground uintptr
	LpszMenuName, LpszClassName              *uint16
	HIconSm                                  uintptr
}
type PAINTSTRUCT struct {
	Hdc                uintptr
	Erase              int32
	RcPaint            RECT
	Restore, IncUpdate int32
	Reserved           [32]byte
}
type OPENFILENAME struct {
	LStructSize                    uint32
	HwndOwner, HInstance           uintptr
	LpstrFilter, LpstrCustomFilter *uint16
	NMaxCustFilter, NFilterIndex   uint32
	LpstrFile                      *uint16
	NMaxFile                       uint32
	LpstrFileTitle                 *uint16
	NMaxFileTitle                  uint32
	LpstrInitialDir, LpstrTitle    *uint16
	Flags                          uint32
	NFileOffset, NFileExtension    uint16
	LpstrDefExt                    *uint16
	LCustData                      uintptr
	LpfnHook                       uintptr
	LpTemplateName                 *uint16
	PvReserved                     uintptr
	DwReserved, FlagsEx            uint32
}
type BROWSEINFO struct {
	HwndOwner      uintptr
	PidlRoot       uintptr
	PszDisplayName *uint16
	LpszTitle      *uint16
	UlFlags        uint32
	Lpfn           uintptr
	LParam         uintptr
	IImage         int32
}
type BITMAPINFOHEADER struct {
	BiSize                           uint32
	BiWidth, BiHeight                int32
	BiPlanes, BiBitCount             uint16
	BiCompression                    uint32
	BiSizeImage                      uint32
	BiXPelsPerMeter, BiYPelsPerMeter int32
	BiClrUsed, BiClrImportant        uint32
}
type RGBQUAD struct{ Blue, Green, Red, Reserved byte }
type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]RGBQUAD
}

var (
	user32    = syscall.NewLazyDLL("user32.dll")
	gdi32     = syscall.NewLazyDLL("gdi32.dll")
	comdlg32  = syscall.NewLazyDLL("comdlg32.dll")
	shell32   = syscall.NewLazyDLL("shell32.dll")
	kernel32w = syscall.NewLazyDLL("kernel32.dll")
	ole32     = syscall.NewLazyDLL("ole32.dll")

	procRegisterClassExW              = user32.NewProc("RegisterClassExW")
	procCreateWindowExW               = user32.NewProc("CreateWindowExW")
	procDefWindowProcW                = user32.NewProc("DefWindowProcW")
	procShowWindow                    = user32.NewProc("ShowWindow")
	procUpdateWindow                  = user32.NewProc("UpdateWindow")
	procGetMessageW                   = user32.NewProc("GetMessageW")
	procTranslateMessage              = user32.NewProc("TranslateMessage")
	procDispatchMessageW              = user32.NewProc("DispatchMessageW")
	procPostQuitMessage               = user32.NewProc("PostQuitMessage")
	procBeginPaint                    = user32.NewProc("BeginPaint")
	procEndPaint                      = user32.NewProc("EndPaint")
	procFillRect                      = user32.NewProc("FillRect")
	procGetClientRect                 = user32.NewProc("GetClientRect")
	procInvalidateRect                = user32.NewProc("InvalidateRect")
	procLoadCursorW                   = user32.NewProc("LoadCursorW")
	procMessageBoxW                   = user32.NewProc("MessageBoxW")
	procPostMessageW                  = user32.NewProc("PostMessageW")
	procMoveWindow                    = user32.NewProc("MoveWindow")
	procEnableWindow                  = user32.NewProc("EnableWindow")
	procSetFocus                      = user32.NewProc("SetFocus")
	procGetFocus                      = user32.NewProc("GetFocus")
	procSetWindowPos                  = user32.NewProc("SetWindowPos")
	procGetKeyState                   = user32.NewProc("GetKeyState")
	procOpenClipboard                 = user32.NewProc("OpenClipboard")
	procCloseClipboard                = user32.NewProc("CloseClipboard")
	procIsClipboardFormatAvailable    = user32.NewProc("IsClipboardFormatAvailable")
	procGetClipboardData              = user32.NewProc("GetClipboardData")
	procSetWindowTextW                = user32.NewProc("SetWindowTextW")
	procGetWindowTextW                = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW          = user32.NewProc("GetWindowTextLengthW")
	procDestroyWindow                 = user32.NewProc("DestroyWindow")
	procSetProcessDpiAwarenessContext = user32.NewProc("SetProcessDpiAwarenessContext")
	procGetDpiForWindow               = user32.NewProc("GetDpiForWindow")
	procSendMessageW                  = user32.NewProc("SendMessageW")
	procSetTextColor                  = gdi32.NewProc("SetTextColor")
	procSetBkMode                     = gdi32.NewProc("SetBkMode")
	procSetBkColor                    = gdi32.NewProc("SetBkColor")
	procCreateSolidBrush              = gdi32.NewProc("CreateSolidBrush")
	procDeleteObject                  = gdi32.NewProc("DeleteObject")
	procCreateFontW                   = gdi32.NewProc("CreateFontW")
	procSelectObject                  = gdi32.NewProc("SelectObject")
	procRoundRect                     = gdi32.NewProc("RoundRect")
	procRectangle                     = gdi32.NewProc("Rectangle")
	procCreatePen                     = gdi32.NewProc("CreatePen")
	procStretchDIBits                 = gdi32.NewProc("StretchDIBits")
	procDrawTextW                     = user32.NewProc("DrawTextW")
	procGetOpenFileNameW              = comdlg32.NewProc("GetOpenFileNameW")
	procGetSaveFileNameW              = comdlg32.NewProc("GetSaveFileNameW")
	procDragAcceptFiles               = shell32.NewProc("DragAcceptFiles")
	procDragQueryFileW                = shell32.NewProc("DragQueryFileW")
	procDragFinish                    = shell32.NewProc("DragFinish")
	procGetModuleHandleW              = kernel32w.NewProc("GetModuleHandleW")
	procGlobalLock                    = kernel32w.NewProc("GlobalLock")
	procGlobalUnlock                  = kernel32w.NewProc("GlobalUnlock")
	procGlobalSize                    = kernel32w.NewProc("GlobalSize")
	procSHBrowseForFolderW            = shell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW          = shell32.NewProc("SHGetPathFromIDListW")
	procCoTaskMemFree                 = ole32.NewProc("CoTaskMemFree")
	procOleInitialize                 = ole32.NewProc("OleInitialize")
	procOleUninitialize               = ole32.NewProc("OleUninitialize")
)

func utf16Ptr(s string) *uint16    { p, _ := syscall.UTF16PtrFromString(s); return p }
func rgb(r, g, b byte) uintptr     { return uintptr(uint32(r) | uint32(g)<<8 | uint32(b)<<16) }
func loword(v uintptr) int32       { return int32(uint16(v & 0xffff)) }
func hiword(v uintptr) int32       { return int32(uint16((v >> 16) & 0xffff)) }
func signedLoword(v uintptr) int32 { return int32(int16(v & 0xffff)) }
func signedHiword(v uintptr) int32 { return int32(int16((v >> 16) & 0xffff)) }

func createWindowEx(ex uint32, class, title string, style uint32, x, y, w, h int32, parent, menu, instance uintptr, param unsafe.Pointer) uintptr {
	r, _, _ := procCreateWindowExW.Call(uintptr(ex), uintptr(unsafe.Pointer(utf16Ptr(class))), uintptr(unsafe.Pointer(utf16Ptr(title))), uintptr(style), uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, menu, instance, uintptr(param))
	return r
}
func showWindow(hwnd uintptr, cmd int32) { procShowWindow.Call(hwnd, uintptr(cmd)) }
func invalidate(hwnd uintptr)            { procInvalidateRect.Call(hwnd, 0, 1) }
func setWindowText(hwnd uintptr, s string) {
	procSetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(utf16Ptr(s))))
}
func getWindowText(hwnd uintptr) string {
	n, _, _ := procGetWindowTextLengthW.Call(hwnd)
	buf := make([]uint16, n+1)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}
func messageBox(hwnd uintptr, title, text string, flags uint32) int {
	r, _, _ := procMessageBoxW.Call(hwnd, uintptr(unsafe.Pointer(utf16Ptr(text))), uintptr(unsafe.Pointer(utf16Ptr(title))), uintptr(flags))
	return int(r)
}
func drawText(hdc uintptr, text string, r *RECT, flags uint32) {
	u, _ := syscall.UTF16FromString(text)
	if len(u) == 0 {
		return
	}
	procDrawTextW.Call(hdc, uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)-1), uintptr(unsafe.Pointer(r)), uintptr(flags))
}
