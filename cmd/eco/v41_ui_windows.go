//go:build windows

package main

import "unsafe"

const (
	ctrlV41NavBase   = 1200
	v41BSOwnerDraw   = 0x0000000B
	WM_DRAWITEM      = 0x002B
	v41ODSSelected   = 0x0001
	v41ODSFocus      = 0x0010
)

var v41NavPages = [7]string{"home", "evidence", "ask", "matters", "review", "changes", "trust"}
var v41NavLabels = [7]string{"Home", "Evidence", "Ask ECO", "Matters", "Review", "Changes", "Trust & settings"}

type v41DrawItemStruct struct {
	CtlType    uint32
	CtlID      uint32
	ItemID     uint32
	ItemAction uint32
	ItemState  uint32
	HwndItem   uintptr
	HDC        uintptr
	RcItem     RECT
	ItemData   uintptr
}

func (a *application) createV41Navigation(hwnd uintptr) {
	for i, label := range v41NavLabels {
		a.navButtons[i] = createWindowEx(
			0,
			"BUTTON",
			label,
			WS_CHILD|WS_VISIBLE|WS_TABSTOP|v41BSOwnerDraw,
			0, 0, 10, 10,
			hwnd,
			uintptr(ctrlV41NavBase+i),
			a.hInstance,
			nil,
		)
	}
}

func (a *application) layoutV41Navigation() {
	y := int32(118)
	for _, h := range a.navButtons {
		if h == 0 {
			y += 53
			continue
		}
		procMoveWindow.Call(h, 18, uintptr(y), 239, 48, 1)
		procSendMessageW.Call(h, WM_SETFONT, a.fontNav, 1)
		showWindow(h, SW_SHOW)
		invalidate(h)
		y += 53
	}
}

func (a *application) handleV41NavCommand(id, code int) bool {
	if code != BN_CLICKED || id < ctrlV41NavBase || id >= ctrlV41NavBase+len(v41NavPages) {
		return false
	}
	a.setPage(v41NavPages[id-ctrlV41NavBase])
	return true
}

func (a *application) drawV41OwnerButton(lparam uintptr) bool {
	if lparam == 0 {
		return false
	}
	var di v41DrawItemStruct
	copyWindowsMemoryToGo(unsafe.Pointer(&di), lparam, unsafe.Sizeof(di))
	id := int(di.CtlID)
	if id < ctrlV41NavBase || id >= ctrlV41NavBase+len(v41NavPages) {
		return false
	}
	idx := id - ctrlV41NavBase
	r := di.RcItem
	bg := rgb(5, 61, 70)
	border := rgb(5, 61, 70)
	if a.page == v41NavPages[idx] {
		bg = rgb(24, 111, 112)
		border = rgb(44, 145, 143)
	}
	if di.ItemState&v41ODSSelected != 0 {
		bg = rgb(16, 90, 94)
		border = rgb(74, 171, 164)
	}
	if di.ItemState&v41ODSFocus != 0 {
		border = rgb(177, 240, 231)
	}
	roundRect(di.HDC, r, 12, bg, border)
	drawTextFont(di.HDC, navIcon(idx), RECT{12, 0, 42, r.Bottom}, a.fontNav, rgb(225, 248, 245), DT_CENTER|DT_VCENTER|DT_SINGLELINE)
	drawTextFont(di.HDC, v41NavLabels[idx], RECT{50, 0, r.Right - 39, r.Bottom}, a.fontNav, rgb(245, 253, 252), DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)

	count := 0
	switch v41NavPages[idx] {
	case "evidence":
		count = len(a.view.Evidence)
	case "matters":
		count = len(a.view.Matters)
	case "review":
		count = a.reviewCount()
	case "changes":
		count = len(a.view.Changes)
	}
	if count > 0 {
		drawTextFont(di.HDC, fmtInt(count), RECT{r.Right - 38, 0, r.Right - 9, r.Bottom}, a.fontSmall, rgb(216, 242, 238), DT_CENTER|DT_VCENTER|DT_SINGLELINE)
	}
	return true
}

func fmtInt(v int) string {
	if v == 0 {
		return "0"
	}
	buf := [24]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
