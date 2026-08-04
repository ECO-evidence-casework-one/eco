//go:build windows

package eco

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	genericRead               = 0x80000000
	genericWrite              = 0x40000000
	fileShareRead             = 0x00000001
	fileShareWrite            = 0x00000002
	fileShareDelete           = 0x00000004
	openAlways                = 4
	openExisting              = 3
	fileAttributeHidden       = 0x00000002
	fileFlagBackupSemantics   = 0x02000000
	lockfileFailImmediately   = 0x00000001
	lockfileExclusiveLock     = 0x00000002
	errorLockViolation        syscall.Errno = 33
	errorSharingViolation     syscall.Errno = 32
)

var (
	kernel32Workspace                    = syscall.NewLazyDLL("kernel32.dll")
	procLockFileExWorkspace              = kernel32Workspace.NewProc("LockFileEx")
	procUnlockFileExWorkspace            = kernel32Workspace.NewProc("UnlockFileEx")
	procGetFileInformationByHandleWorkspace = kernel32Workspace.NewProc("GetFileInformationByHandle")
)

type windowsWorkspaceLock struct {
	handle     syscall.Handle
	overlapped syscall.Overlapped
}

type byHandleFileInformation struct {
	FileAttributes     uint32
	CreationTime       syscall.Filetime
	LastAccessTime     syscall.Filetime
	LastWriteTime      syscall.Filetime
	VolumeSerialNumber uint32
	FileSizeHigh       uint32
	FileSizeLow        uint32
	NumberOfLinks      uint32
	FileIndexHigh      uint32
	FileIndexLow       uint32
}

func platformAcquireWorkspaceLock(path string) (platformWorkspaceLock, error) {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("encode workspace owner lock path: %w", err)
	}
	handle, err := syscall.CreateFile(ptr, genericRead|genericWrite, fileShareRead|fileShareWrite|fileShareDelete, nil, openAlways, fileAttributeHidden, 0)
	if err != nil {
		return nil, fmt.Errorf("open workspace owner lock: %w", err)
	}
	lock := &windowsWorkspaceLock{handle: handle}
	ok, _, callErr := procLockFileExWorkspace.Call(uintptr(handle), lockfileExclusiveLock|lockfileFailImmediately, 0, 1, 0, uintptr(unsafe.Pointer(&lock.overlapped)))
	if ok == 0 {
		_ = syscall.CloseHandle(handle)
		errno, _ := callErr.(syscall.Errno)
		if errors.Is(errno, errorLockViolation) || errors.Is(errno, errorSharingViolation) {
			return nil, ErrWorkspaceInUse
		}
		return nil, fmt.Errorf("lock workspace owner file: %w", callErr)
	}
	return lock, nil
}

func (l *windowsWorkspaceLock) Close() error {
	if l == nil || l.handle == syscall.InvalidHandle {
		return nil
	}
	ok, _, unlockErr := procUnlockFileExWorkspace.Call(uintptr(l.handle), 0, 1, 0, uintptr(unsafe.Pointer(&l.overlapped)))
	closeErr := syscall.CloseHandle(l.handle)
	l.handle = syscall.InvalidHandle
	if ok == 0 {
		return fmt.Errorf("unlock workspace owner file: %w", unlockErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close workspace owner file: %w", closeErr)
	}
	return nil
}

func platformWorkspaceObjectIdentity(path string) (workspaceObjectIdentity, error) {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return workspaceObjectIdentity{}, err
	}
	handle, err := syscall.CreateFile(ptr, 0, fileShareRead|fileShareWrite|fileShareDelete, nil, openExisting, fileFlagBackupSemantics, 0)
	if err != nil {
		return workspaceObjectIdentity{}, err
	}
	defer syscall.CloseHandle(handle)
	var info byHandleFileInformation
	ok, _, callErr := procGetFileInformationByHandleWorkspace.Call(uintptr(handle), uintptr(unsafe.Pointer(&info)))
	if ok == 0 {
		return workspaceObjectIdentity{}, callErr
	}
	file := uint64(info.FileIndexHigh)<<32 | uint64(info.FileIndexLow)
	return workspaceObjectIdentity{Volume: uint64(info.VolumeSerialNumber), File: file}, nil
}
