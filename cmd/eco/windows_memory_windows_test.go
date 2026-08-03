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

func TestWindowsMinMaxInfoFieldWrite(t *testing.T) {
	if got, want := unsafe.Sizeof(MINMAXINFO{}), uintptr(40); got != want {
		t.Fatalf("MINMAXINFO size changed: got %d want %d", got, want)
	}
	if got, want := unsafe.Offsetof(MINMAXINFO{}.PtMinTrackSize), uintptr(24); got != want {
		t.Fatalf("PtMinTrackSize offset changed: got %d want %d", got, want)
	}

	var info MINMAXINFO
	minTrack := POINT{X: 1000, Y: 730}
	copyGoMemoryToWindows(
		uintptr(unsafe.Pointer(&info))+unsafe.Offsetof(MINMAXINFO{}.PtMinTrackSize),
		unsafe.Pointer(&minTrack),
		unsafe.Sizeof(minTrack),
	)
	runtime.KeepAlive(&info)
	if info.PtMinTrackSize != minTrack {
		t.Fatalf("minimum-track write mismatch: got %+v want %+v", info.PtMinTrackSize, minTrack)
	}
}

func TestWindowsMemoryCopyHelpersRejectInvalidAddresses(t *testing.T) {
	var source = [4]byte{1, 2, 3, 4}
	var target [4]byte

	copyWindowsMemoryToGo(unsafe.Pointer(&target[0]), 0, unsafe.Sizeof(target))
	copyWindowsMemoryToGo(unsafe.Pointer(&target[0]), minimumWindowsUserAddress-1, unsafe.Sizeof(target))
	copyGoMemoryToWindows(0, unsafe.Pointer(&source[0]), unsafe.Sizeof(source))
	copyGoMemoryToWindows(minimumWindowsUserAddress-1, unsafe.Pointer(&source[0]), unsafe.Sizeof(source))
	copyWindowsMemoryToGo(nil, uintptr(unsafe.Pointer(&source[0])), unsafe.Sizeof(source))
	copyGoMemoryToWindows(uintptr(unsafe.Pointer(&target[0])), nil, unsafe.Sizeof(target))

	if target != [4]byte{} {
		t.Fatalf("invalid-address copy changed target: got %x", target)
	}
}

func TestMinMaxInfoMessageRejectsNilAddress(t *testing.T) {
	if got := mainWndProc(0, WM_GETMINMAXINFO, 0, 0); got != 0 {
		t.Fatalf("WM_GETMINMAXINFO nil-address result = %d, want 0", got)
	}
}
