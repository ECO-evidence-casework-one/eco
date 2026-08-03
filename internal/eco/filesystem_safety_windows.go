//go:build windows

package eco

import (
	"os"
	"syscall"
)

const windowsFileAttributeReparsePoint = 0x400

func fileInfoHasReparsePoint(info os.FileInfo) bool {
	data, ok := info.Sys().(*syscall.Win32FileAttributeData)
	return ok && data.FileAttributes&windowsFileAttributeReparsePoint != 0
}
