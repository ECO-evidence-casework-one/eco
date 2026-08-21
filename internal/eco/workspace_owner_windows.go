//go:build windows

package eco

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
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
	genericRead             = 0x80000000
	genericWrite            = 0x40000000
	openExisting            = 3
	openAlways              = 4
	fileAttributeNormal     = 0x00000080
	fileFlagBackupSemantics = 0x02000000
)

var (
	kernel32Workspace                       = syscall.NewLazyDLL("kernel32.dll")
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

// platformAcquireWorkspaceLock binds cross-process ownership to an exclusive
// handle on a deterministic user-local lock file keyed by the underlying
// filesystem object identity and the logical lock role.
//
// The lock file deliberately lives outside the workspace object. Holding its
// handle therefore does not block workspace directory rename/retarget, while
// share mode 0 prevents a second process from opening, replacing or deleting
// that exact lock file until the owning handle closes or the process exits.
func platformAcquireWorkspaceLock(path string) (platformWorkspaceLock, error) {
	identity, err := platformWorkspaceObjectIdentity(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("identify workspace lock object: %w", err)
	}
	lockPath, err := platformWorkspaceLockPath(identity, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	ptr, err := syscall.UTF16PtrFromString(lockPath)
	if err != nil {
		return nil, fmt.Errorf("encode workspace lock path: %w", err)
	}
	handle, err := syscall.CreateFile(
		ptr,
		genericRead|genericWrite,
		0, // exclusive sharing: no competing open/delete/replace while held
		nil,
		openAlways,
		fileAttributeNormal,
		0,
	)
	if err != nil {
		if errors.Is(err, syscall.Errno(32)) { // ERROR_SHARING_VIOLATION
			return nil, ErrWorkspaceInUse
		}
		return nil, fmt.Errorf("open exclusive workspace lock file: %w", err)
	}
	return &windowsWorkspaceLock{handle: handle}, nil
}

func platformWorkspaceLockPath(identity workspaceObjectIdentity, role string) (string, error) {
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache for workspace lock: %w", err)
	}
	lockDir := filepath.Join(cacheRoot, "ECO", "workspace-locks")
	if err := os.MkdirAll(lockDir, 0700); err != nil {
		return "", fmt.Errorf("create workspace lock directory: %w", err)
	}
	sum := sha256.Sum256([]byte(strings.ToLower(role)))
	name := fmt.Sprintf(
		"%016x-%016x-%s.lck",
		identity.Volume,
		identity.File,
		hex.EncodeToString(sum[:8]),
	)
	return filepath.Join(lockDir, name), nil
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
	closeErr := syscall.CloseHandle(l.handle)
	l.handle = syscall.InvalidHandle
	if closeErr != nil {
		return fmt.Errorf("close workspace lock file: %w", closeErr)
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
