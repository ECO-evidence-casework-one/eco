//go:build windows

package eco

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

const (
	windowsGenericWrite        = 0x40000000
	windowsOpenAlways          = 4
	windowsCreateNew           = 1
	windowsFileAttributeNormal = 0x00000080
)

type windowsWorkspaceLifecycleLock struct {
	parent *windowsBoundHandle
	lock   syscall.Handle
}

func acquireRawWorkspaceLifecycleLock(root, lockName string) (rawWorkspaceLifecycleLock, error) {
	parent, err := openWindowsBoundHandle(filepath.Dir(root), windowsFileReadAttributes|windowsSynchronizeAccess, true)
	if err != nil {
		return nil, err
	}
	lockPath := filepath.Join(parent.path, lockName)
	pointer, err := syscall.UTF16PtrFromString(lockPath)
	if err != nil {
		parent.Close()
		return nil, err
	}
	handle, err := syscall.CreateFile(
		pointer,
		windowsGenericRead|windowsGenericWrite|windowsFileReadAttributes|windowsSynchronizeAccess,
		0,
		nil,
		windowsOpenAlways,
		windowsFileFlagOpenReparse|windowsFileAttributeNormal,
		0,
	)
	if err != nil {
		parent.Close()
		return nil, err
	}
	var information syscall.ByHandleFileInformation
	if err = syscall.GetFileInformationByHandle(handle, &information); err != nil {
		_ = syscall.CloseHandle(handle)
		parent.Close()
		return nil, err
	}
	if information.FileAttributes&(windowsFileAttributeDirectory|windowsFileAttributeReparsePoint) != 0 {
		_ = syscall.CloseHandle(handle)
		parent.Close()
		return nil, errors.New("the workspace lifecycle lock is not a normal file")
	}
	return &windowsWorkspaceLifecycleLock{parent: parent, lock: handle}, nil
}

func (lock *windowsWorkspaceLifecycleLock) Close() error {
	if lock == nil {
		return nil
	}
	var result error
	if lock.lock != syscall.InvalidHandle {
		result = syscall.CloseHandle(lock.lock)
		lock.lock = syscall.InvalidHandle
	}
	if lock.parent != nil {
		result = errors.Join(result, lock.parent.Close())
		lock.parent = nil
	}
	return result
}

type windowsWorkspaceObjects struct {
	root    *windowsBoundHandle
	objects *windowsBoundHandle
	files   map[string]*windowsBoundHandle
	missing map[string]bool
}

func openBoundWorkspaceObjects(root string) (boundWorkspaceObjects, error) {
	bound := &windowsWorkspaceObjects{files: make(map[string]*windowsBoundHandle), missing: make(map[string]bool)}
	var err error
	bound.root, err = openWindowsBoundHandle(root, windowsFileListDirectory|windowsFileReadAttributes|windowsSynchronizeAccess, true)
	if err != nil {
		return nil, fmt.Errorf("cannot retain the workspace folder safely: %w", err)
	}
	bound.objects, err = openWindowsBoundHandle(filepath.Join(bound.root.path, "objects"), windowsFileListDirectory|windowsFileReadAttributes|windowsSynchronizeAccess, true)
	if err != nil {
		bound.Close()
		return nil, fmt.Errorf("cannot retain the encrypted object folder safely: %w", err)
	}
	if bound.objects.identity.volume != bound.root.identity.volume {
		bound.Close()
		return nil, errors.New("the encrypted object folder is on another filesystem volume")
	}
	for _, name := range []string{"vault.key", "workspace.ecodb", workspaceIdentityFile} {
		handle, openErr := openWindowsBoundHandle(filepath.Join(bound.root.path, name), windowsGenericRead|windowsFileReadAttributes|windowsSynchronizeAccess, false)
		if openErr != nil {
			if os.IsNotExist(openErr) && name == workspaceIdentityFile {
				bound.missing[name] = true
				continue
			}
			bound.Close()
			return nil, fmt.Errorf("cannot retain workspace control file %q safely: %w", name, openErr)
		}
		if handle.identity.volume != bound.root.identity.volume {
			handle.Close()
			bound.Close()
			return nil, fmt.Errorf("workspace control file %q is on another filesystem volume", name)
		}
		bound.files[name] = handle
	}
	return bound, nil
}

func readWindowsWorkspaceFile(handle *windowsBoundHandle) ([]byte, error) {
	info, err := handle.file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() < 0 || uint64(info.Size()) > uint64(^uint(0)>>1) {
		return nil, errors.New("the workspace control file is too large to read safely")
	}
	data := make([]byte, int(info.Size()))
	read, err := handle.file.ReadAt(data, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if read != len(data) {
		return nil, io.ErrUnexpectedEOF
	}
	return data, nil
}

func (bound *windowsWorkspaceObjects) ReadFile(name string) ([]byte, error) {
	if bound.missing[name] {
		return nil, os.ErrNotExist
	}
	handle := bound.files[name]
	if handle == nil {
		return nil, errors.New("the requested file is not part of the retained workspace")
	}
	return readWindowsWorkspaceFile(handle)
}

func (bound *windowsWorkspaceObjects) Verify() error {
	if err := bound.root.verifyPath(); err != nil {
		return fmt.Errorf("workspace folder changed after authentication: %w", err)
	}
	if err := bound.objects.verifyPath(); err != nil {
		return fmt.Errorf("encrypted object folder changed after authentication: %w", err)
	}
	for name, handle := range bound.files {
		if err := handle.verifyPath(); err != nil {
			return fmt.Errorf("workspace control file %q changed after authentication: %w", name, err)
		}
	}
	for name := range bound.missing {
		if _, err := os.Lstat(filepath.Join(bound.root.path, name)); err == nil {
			return fmt.Errorf("workspace control file %q appeared after authentication", name)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func createWindowsBoundFile(path string) (*windowsBoundHandle, error) {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := syscall.CreateFile(pointer, windowsGenericRead|windowsGenericWrite|windowsDeleteAccess|windowsFileReadAttributes|windowsSynchronizeAccess, windowsFileShareRead, nil, windowsCreateNew, windowsFileFlagOpenReparse|windowsFileAttributeNormal, 0)
	if err != nil {
		return nil, err
	}
	var information syscall.ByHandleFileInformation
	if err = syscall.GetFileInformationByHandle(handle, &information); err != nil {
		_ = syscall.CloseHandle(handle)
		return nil, err
	}
	if information.FileAttributes&(windowsFileAttributeDirectory|windowsFileAttributeReparsePoint) != 0 {
		_ = syscall.CloseHandle(handle)
		return nil, errors.New("the new workspace control file is not a normal file")
	}
	file := os.NewFile(uintptr(handle), path)
	if file == nil {
		_ = syscall.CloseHandle(handle)
		return nil, errors.New("Windows could not retain the new workspace control file")
	}
	return &windowsBoundHandle{
		path:   path,
		handle: handle,
		file:   file,
		identity: windowsObjectIdentity{
			volume: information.VolumeSerialNumber,
			high:   information.FileIndexHigh,
			low:    information.FileIndexLow,
		},
	}, nil
}

func (bound *windowsWorkspaceObjects) WriteFileAtomic(name string, data []byte, mode uint32) error {
	if name != "workspace.ecodb" && name != workspaceIdentityFile {
		return errors.New("the requested workspace file cannot be replaced through this transaction")
	}
	if err := bound.Verify(); err != nil {
		return err
	}
	tmpName := ".eco-write-" + NewID("FS") + ".tmp"
	tmpPath := filepath.Join(bound.root.path, tmpName)
	tmp, err := createWindowsBoundFile(tmpPath)
	if err != nil {
		return err
	}
	cleanup := func() {
		_ = setWindowsDisposition(tmp.handle, true)
		_ = tmp.Close()
	}
	if _, err = tmp.file.Write(data); err != nil {
		cleanup()
		return err
	}
	if err = tmp.file.Sync(); err != nil {
		cleanup()
		return err
	}
	old := bound.files[name]
	if old != nil {
		if err = old.Close(); err != nil {
			cleanup()
			return err
		}
		delete(bound.files, name)
	}
	if err = renameWindowsHandleReplacing(tmp.handle, bound.root.handle, name, true); err != nil {
		cleanup()
		reopened, reopenErr := openWindowsBoundHandle(filepath.Join(bound.root.path, name), windowsGenericRead|windowsFileReadAttributes|windowsSynchronizeAccess, false)
		if reopenErr == nil {
			bound.files[name] = reopened
		}
		return errors.Join(err, reopenErr)
	}
	tmp.path = filepath.Join(bound.root.path, name)
	bound.files[name] = tmp
	delete(bound.missing, name)
	if tmp.identity.volume != bound.root.identity.volume {
		return errors.New("the replacement workspace control file moved to another filesystem volume")
	}
	return tmp.verifyPath()
}

func (bound *windowsWorkspaceObjects) Close() error {
	if bound == nil {
		return nil
	}
	var result error
	for name, handle := range bound.files {
		result = errors.Join(result, handle.Close())
		delete(bound.files, name)
	}
	if bound.objects != nil {
		result = errors.Join(result, bound.objects.Close())
		bound.objects = nil
	}
	if bound.root != nil {
		result = errors.Join(result, bound.root.Close())
		bound.root = nil
	}
	return result
}
