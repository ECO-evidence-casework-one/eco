//go:build windows

package main

import (
	"runtime"
	"unsafe"
)

const minimumWindowsUserAddress = uintptr(0x10000)

var procRtlMoveMemory = kernel32w.NewProc("RtlMoveMemory")

// copyWindowsMemoryToGo copies bytes from memory owned by Windows into a
// caller-provided Go object while the Windows API contract says the source
// address is valid. It avoids converting a Windows-supplied uintptr into a Go
// pointer, which would lose the source lifetime and provenance information.
func copyWindowsMemoryToGo(dst unsafe.Pointer, src, size uintptr) {
	if dst == nil || src < minimumWindowsUserAddress || size == 0 {
		return
	}
	procRtlMoveMemory.Call(uintptr(dst), src, size)
	runtime.KeepAlive(dst)
}

// copyGoMemoryToWindows copies bytes from a live Go object into memory owned
// by Windows. The caller must use this only during the Windows callback or API
// call that guarantees the destination address remains valid. Windows treats
// addresses below 0x10000 as the null-pointer range, so null-plus-offset values
// are rejected before calling RtlMoveMemory.
func copyGoMemoryToWindows(dst uintptr, src unsafe.Pointer, size uintptr) {
	if dst < minimumWindowsUserAddress || src == nil || size == 0 {
		return
	}
	procRtlMoveMemory.Call(dst, uintptr(src), size)
	runtime.KeepAlive(src)
}
