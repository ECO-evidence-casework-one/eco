//go:build windows

package main

import (
	"syscall"
	"time"
	"unsafe"
)

const (
	whGetMessage = 3
	wmNull       = 0x0000
	vkTab        = 0x09
)

var (
	procSetWindowsHookExW        = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx      = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx           = user32.NewProc("CallNextHookEx")
	procGetWindowThreadProcessID = user32.NewProc("GetWindowThreadProcessId")
	procIsDialogMessageW         = user32.NewProc("IsDialogMessageW")
	tabNavigationCallback        = syscall.NewCallback(tabNavigationHookProc)
)

// ECO's top-level window is custom rather than a dialog resource. Several real
// child controls already carry WS_TABSTOP, but the ordinary message loop does
// not call IsDialogMessage. Windows therefore does not get a chance to apply
// its standard Tab/Shift+Tab traversal. Install a thread-local GETMESSAGE hook
// for only the Tab key so the existing message loop and global shortcuts remain
// otherwise unchanged.
func init() {
	go installTabNavigationHook()
}

func installTabNavigationHook() {
	var a *application
	for i := 0; i < 600; i++ {
		if app != nil && app.hwnd != 0 {
			a = app
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if a == nil {
		return
	}

	threadID, _, _ := procGetWindowThreadProcessID.Call(a.hwnd, 0)
	if threadID == 0 {
		return
	}
	hook, _, _ := procSetWindowsHookExW.Call(whGetMessage, tabNavigationCallback, 0, threadID)
	if hook == 0 {
		return
	}
	defer procUnhookWindowsHookEx.Call(hook)

	for app == a && a.hwnd != 0 {
		time.Sleep(250 * time.Millisecond)
	}
}

func tabNavigationHookProc(code, wparam, lparam uintptr) uintptr {
	if int32(code) >= 0 && lparam >= minimumWindowsUserAddress && app != nil && app.hwnd != 0 {
		var msg MSG
		copyWindowsMemoryToGo(unsafe.Pointer(&msg), lparam, unsafe.Sizeof(msg))
		if msg.Message == WM_KEYDOWN && msg.WParam == vkTab {
			handled, _, _ := procIsDialogMessageW.Call(app.hwnd, lparam)
			if handled != 0 {
				// IsDialogMessage already translated/dispatched the keyboard action.
				// Replace the returned queue message with WM_NULL so ECO's ordinary
				// loop cannot translate or dispatch the same key a second time.
				msg.Message = wmNull
				msg.WParam = 0
				msg.LParam = 0
				copyGoMemoryToWindows(lparam, unsafe.Pointer(&msg), unsafe.Sizeof(msg))
			}
		}
	}
	next, _, _ := procCallNextHookEx.Call(0, code, wparam, lparam)
	return next
}
