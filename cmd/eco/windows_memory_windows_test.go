//go:build windows

package main

import (
	"bytes"
	"runtime"
	"testing"
	"unsafe"
)

func TestWindowsMemoryCopyHelpers(t *testing.T) {
	source := [16]byte{0x45, 0x43, 0x4f, 0x2d, 0x73, 0x61, 0x66, 0x65, 1, 2, 3, 4, 5, 6, 7, 8}
	var copiedIntoGo [16]byte

	copyWindowsMemoryToGo(
		unsafe.Pointer(&copiedIntoGo[0]),
		uintptr(unsafe.Pointer(&source[0])),
		unsafe.Sizeof(source),
	)
	runtime.KeepAlive(&source)
	if !bytes.Equal(copiedIntoGo[:], source[:]) {
		t.Fatalf("Windows-to-Go copy mismatch: got %x want %x", copiedIntoGo, source)
	}

	var copiedIntoWindows [16]byte
	copyGoMemoryToWindows(
		uintptr(unsafe.Pointer(&copiedIntoWindows[0])),
		unsafe.Pointer(&source[0]),
		unsafe.Sizeof(source),
	)
	runtime.KeepAlive(&copiedIntoWindows)
	if !bytes.Equal(copiedIntoWindows[:], source[:]) {
		t.Fatalf("Go-to-Windows copy mismatch: got %x want %x", copiedIntoWindows, source)
	}
}

func TestWindowsMemoryCopyHelpersRejectEmptyArguments(t *testing.T) {
	var target [4]byte
	copyWindowsMemoryToGo(unsafe.Pointer(&target[0]), 0, unsafe.Sizeof(target))
	copyGoMemoryToWindows(0, unsafe.Pointer(&target[0]), unsafe.Sizeof(target))
	copyWindowsMemoryToGo(nil, uintptr(unsafe.Pointer(&target[0])), unsafe.Sizeof(target))
	copyGoMemoryToWindows(uintptr(unsafe.Pointer(&target[0])), nil, unsafe.Sizeof(target))
}
