//go:build windows

package eco

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

const (
	fileShareRead           = 0x00000001
	fileShareWrite          = 0x00000002
	fileShareDelete         = 0x00000004
	openExisting            = 3
	fileFlagBackupSemantics = 0x02000000
	waitObject0             = 0x00000000
	waitTimeout             = 0x00000102
)

var (
	kernel32Workspace                       = syscall.NewLazyDLL("kernel32.dll")
	procCreateSemaphoreWWorkspace           = kernel32Workspace.NewProc("CreateSemaphoreW")
	procWaitForSingleObjectWorkspace        = kernel32Workspace.NewProc("WaitForSingleObject")
	procReleaseSemaphoreWorkspace           = kernel32Workspace.NewProc("ReleaseSemaphore")
	procGetFileInformationByHandleWorkspace = kernel32Workspace.NewProc("GetFileInformationByHandle")
)

type windowsWorkspaceLock struct {
	mu     sync.Mutex
	handle syscall.Handle
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
	identity, err := platformWorkspaceObjectIdentity(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("identify workspace lock object: %w", err)
	}
	name := fmt.Sprintf("Global\\ECO.Workspace.%016x.%016x.%s", identity.Volume, identity.File, strings.ToLower(filepath.Base(path)))
	ptr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, fmt.Errorf("encode workspace semaphore name: %w", err)
	}
	h, _, createErr := procCreateSemaphoreWWorkspace.Call(0, 1, 1, uintptr(unsafe.Pointer(ptr)))
	if h == 0 {
		return nil, fmt.Errorf("create workspace semaphore: %w", createErr)
	}
	handle := syscall.Handle(h)
	result, _, waitErr := procWaitForSingleObjectWorkspace.Call(h, 0)
	switch uint32(result) {
	case waitObject0:
		return &windowsWorkspaceLock{handle: handle}, nil
	case waitTimeout:
		_ = syscall.CloseHandle(handle)
		return nil, ErrWorkspaceInUse
	default:
		_ = syscall.CloseHandle(handle)
		return nil, fmt.Errorf("wait for workspace semaphore: %w", waitErr)
	}
}

func (l *windowsWorkspaceLock) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.handle == 0 || l.handle == syscall.InvalidHandle {
		return nil
	}
	ok, _, releaseErr := procReleaseSemaphoreWorkspace.Call(uintptr(l.handle), 1, 0)
	closeErr := syscall.CloseHandle(l.handle)
	l.handle = syscall.InvalidHandle
	if ok == 0 {
		return fmt.Errorf("release workspace semaphore: %w", releaseErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close workspace semaphore: %w", closeErr)
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

func platformWorkspaceCreationKey(route string) string {
	parts := strings.FieldsFunc(route, func(r rune) bool { return r == '/' || r == '\\' })
	for i := range parts {
		parts[i] = strings.ToLower(strings.TrimRight(parts[i], " ."))
	}
	return strings.Join(parts, "\\")
}
