//go:build windows

package main

import (
	"runtime"
	"unsafe"
)

var procRtlMoveMemory = kernel32w.NewProc("RtlMoveMemory")

// copyWindowsMemoryToGo copies bytes from memory owned by Windows into a
// caller-provided Go object while the Windows API contract says the source
// address is valid. It avoids converting a Windows-supplied uintptr into a Go
// pointer, which would lose the source lifetime and provenance information.
func copyWindowsMemoryToGo(dst unsafe.Pointer, src, size uintptr) {
	if dst == nil || src == 0 || size == 0 {
		return
	}
	procRtlMoveMemory.Call(uintptr(dst), src, size)
	runtime.KeepAlive(dst)
}

// copyGoMemoryToWindows copies bytes from a live Go object into memory owned
// by Windows. The caller must use this only during the Windows callback or API
// call that guarantees the destination address remains valid.
func copyGoMemoryToWindows(dst uintptr, src unsafe.Pointer, size uintptr) {
	if dst == 0 || src == nil || size == 0 {
		return
	}
	procRtlMoveMemory.Call(dst, uintptr(src), size)
	runtime.KeepAlive(src)
}
