//go:build windows

package eco

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	windowsDeleteAccess           = 0x00010000
	windowsGenericRead            = 0x80000000
	windowsSynchronizeAccess      = 0x00100000
	windowsFileListDirectory      = 0x00000001
	windowsFileTraverse           = 0x00000020
	windowsFileReadAttributes     = 0x00000080
	windowsFileShareRead          = 0x00000001
	windowsFileShareWrite         = 0x00000002
	windowsOpenExisting           = 3
	windowsFileFlagOpenReparse    = 0x00200000
	windowsFileFlagBackupSemantic = 0x02000000
	windowsFileAttributeDirectory = 0x00000010
	windowsFileDispositionInfo    = 4
	windowsFileDispositionInfoEx  = 21
	windowsFileDispositionDelete  = 0x00000001
	windowsFileDispositionIgnore  = 0x00000010
	windowsNTFileRenameInfo       = 10
	windowsErrorInvalidFunction   = syscall.Errno(1)
	windowsErrorNotSupported      = syscall.Errno(50)
	windowsErrorInvalidParameter  = syscall.Errno(87)
)

var (
	windowsKernel32                       = syscall.NewLazyDLL("kernel32.dll")
	windowsSetFileInformationByHandleProc = windowsKernel32.NewProc("SetFileInformationByHandle")
	windowsNTDLL                          = syscall.NewLazyDLL("ntdll.dll")
	windowsNTSetInformationFileProc       = windowsNTDLL.NewProc("NtSetInformationFile")
	windowsRtlNTStatusToDosErrorProc      = windowsNTDLL.NewProc("RtlNtStatusToDosError")
)

type windowsObjectIdentity struct {
	volume uint32
	high   uint32
	low    uint32
}

type windowsBoundHandle struct {
	path     string
	handle   syscall.Handle
	file     *os.File
	identity windowsObjectIdentity
	dir      bool
}

func openWindowsBoundHandle(path string, access uint32, wantDirectory bool) (*windowsBoundHandle, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	pointer, err := syscall.UTF16PtrFromString(filepath.Clean(absolute))
	if err != nil {
		return nil, err
	}
	handle, err := syscall.CreateFile(
		pointer,
		access,
		// Omitting FILE_SHARE_DELETE keeps this exact object pinned until Close.
		windowsFileShareRead|windowsFileShareWrite,
		nil,
		windowsOpenExisting,
		windowsFileFlagOpenReparse|windowsFileFlagBackupSemantic,
		0,
	)
	if err != nil {
		return nil, err
	}
	bound := &windowsBoundHandle{path: filepath.Clean(absolute), handle: handle}
	var information syscall.ByHandleFileInformation
	if err = syscall.GetFileInformationByHandle(handle, &information); err != nil {
		_ = syscall.CloseHandle(handle)
		return nil, err
	}
	if information.FileAttributes&windowsFileAttributeReparsePoint != 0 {
		_ = syscall.CloseHandle(handle)
		return nil, errors.New("the selected filesystem object is a symbolic link, junction or reparse point")
	}
	bound.dir = information.FileAttributes&windowsFileAttributeDirectory != 0
	if bound.dir != wantDirectory {
		_ = syscall.CloseHandle(handle)
		if wantDirectory {
			return nil, errors.New("the selected filesystem object is not a folder")
		}
		return nil, errors.New("the selected filesystem object is not a regular file")
	}
	bound.identity = windowsObjectIdentity{
		volume: information.VolumeSerialNumber,
		high:   information.FileIndexHigh,
		low:    information.FileIndexLow,
	}
	bound.file = os.NewFile(uintptr(handle), bound.path)
	if bound.file == nil {
		_ = syscall.CloseHandle(handle)
		return nil, errors.New("Windows could not retain the authenticated filesystem handle")
	}
	return bound, nil
}

func (handle *windowsBoundHandle) Close() error {
	if handle == nil || handle.handle == syscall.InvalidHandle {
		return nil
	}
	err := handle.file.Close()
	handle.handle = syscall.InvalidHandle
	handle.file = nil
	return err
}

func (handle *windowsBoundHandle) verifyPath() error {
	current, err := openWindowsBoundHandle(handle.path, windowsFileReadAttributes|windowsSynchronizeAccess, handle.dir)
	if err != nil {
		return fmt.Errorf("the retained filesystem object is no longer at its authenticated path: %w", err)
	}
	defer current.Close()
	if current.identity != handle.identity {
		return errors.New("the retained filesystem object was replaced after authentication")
	}
	return nil
}

func setWindowsDisposition(handle syscall.Handle, remove bool) error {
	if remove {
		flags := uint32(windowsFileDispositionDelete | windowsFileDispositionIgnore)
		result, _, callErr := windowsSetFileInformationByHandleProc.Call(
			uintptr(handle),
			windowsFileDispositionInfoEx,
			uintptr(unsafe.Pointer(&flags)),
			unsafe.Sizeof(flags),
		)
		runtime.KeepAlive(&flags)
		if result != 0 {
			return nil
		}
		if callErr != windowsErrorInvalidParameter && callErr != windowsErrorInvalidFunction && callErr != windowsErrorNotSupported {
			if callErr != syscall.Errno(0) {
				return callErr
			}
			return syscall.EINVAL
		}
	}
	var disposition byte
	if remove {
		disposition = 1
	}
	result, _, callErr := windowsSetFileInformationByHandleProc.Call(
		uintptr(handle),
		windowsFileDispositionInfo,
		uintptr(unsafe.Pointer(&disposition)),
		1,
	)
	runtime.KeepAlive(&disposition)
	if result == 0 {
		if callErr != syscall.Errno(0) {
			return callErr
		}
		return syscall.EINVAL
	}
	return nil
}

type windowsFileRenameInformation struct {
	replaceIfExists byte
	rootDirectory   syscall.Handle
	fileNameLength  uint32
	fileName        [syscall.MAX_PATH]uint16
}

type windowsIOStatusBlock struct {
	status      uintptr
	information uintptr
}

func renameWindowsHandleReplacing(handle, parent syscall.Handle, destinationPath string, replace bool) error {
	// NtSetInformationFile resolves this simple child name from the retained
	// parent handle, rather than traversing an independently checked pathname.
	name, err := syscall.UTF16FromString(filepath.Base(destinationPath))
	if err != nil {
		return err
	}
	if len(name) > len(windowsFileRenameInformation{}.fileName) {
		return errors.New("the migration destination name is too long for a handle-bound Windows rename")
	}
	information := windowsFileRenameInformation{
		rootDirectory:  parent,
		fileNameLength: uint32((len(name) - 1) * 2),
	}
	if replace {
		information.replaceIfExists = 1
	}
	copy(information.fileName[:], name)
	var ioStatus windowsIOStatusBlock
	status, _, _ := windowsNTSetInformationFileProc.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&ioStatus)),
		uintptr(unsafe.Pointer(&information)),
		unsafe.Sizeof(information),
		windowsNTFileRenameInfo,
	)
	runtime.KeepAlive(&information)
	runtime.KeepAlive(&ioStatus)
	if status != 0 {
		dosError, _, _ := windowsRtlNTStatusToDosErrorProc.Call(status)
		if dosError == 0 {
			return fmt.Errorf("Windows rename failed with NT status 0x%x", status)
		}
		return syscall.Errno(dosError)
	}
	return nil
}

func renameWindowsHandle(handle, parent syscall.Handle, destinationPath string) error {
	return renameWindowsHandleReplacing(handle, parent, destinationPath, false)
}

type windowsBoundObjectDirectory struct {
	directory *windowsBoundHandle
}

func openBoundObjectDirectory(path, description string) (boundObjectDirectory, error) {
	handle, err := openWindowsBoundHandle(
		path,
		windowsFileListDirectory|windowsFileReadAttributes|windowsSynchronizeAccess,
		true,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot retain %s safely: %w", description, err)
	}
	return &windowsBoundObjectDirectory{directory: handle}, nil
}

func (directory *windowsBoundObjectDirectory) Close() error { return directory.directory.Close() }

func (directory *windowsBoundObjectDirectory) VerifyPath() error {
	return directory.directory.verifyPath()
}

func (directory *windowsBoundObjectDirectory) SameFilesystem(other boundObjectDirectory) bool {
	compared, ok := other.(*windowsBoundObjectDirectory)
	return ok && directory.directory.identity.volume == compared.directory.identity.volume
}

type windowsBoundRegularChild struct {
	name   string
	handle *windowsBoundHandle
}

type windowsBoundRegularChildren struct {
	directory *windowsBoundObjectDirectory
	children  []windowsBoundRegularChild
}

func (directory *windowsBoundObjectDirectory) BindRegularChildren(names []string) (boundRegularChildren, error) {
	children := &windowsBoundRegularChildren{directory: directory}
	for _, name := range names {
		if !safeManagedObjectName(name) {
			children.Close()
			return nil, errors.New("the managed encrypted object name is invalid")
		}
		path := filepath.Join(directory.directory.path, name)
		handle, err := openWindowsBoundHandle(
			path,
			windowsDeleteAccess|windowsFileReadAttributes|windowsSynchronizeAccess,
			false,
		)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			children.Close()
			return nil, fmt.Errorf("cannot retain managed encrypted object %q safely: %w", name, err)
		}
		if handle.identity.volume != directory.directory.identity.volume {
			handle.Close()
			children.Close()
			return nil, fmt.Errorf("managed encrypted object %q is on a different filesystem volume", name)
		}
		children.children = append(children.children, windowsBoundRegularChild{name: name, handle: handle})
	}
	return children, nil
}

func (children *windowsBoundRegularChildren) Close() error {
	var result error
	for index := range children.children {
		result = errors.Join(result, children.children[index].handle.Close())
	}
	children.children = nil
	return result
}

func (children *windowsBoundRegularChildren) RemoveAll(before func(string) error) (int, error) {
	for index := range children.children {
		child := &children.children[index]
		if before != nil {
			if err := before(child.handle.path); err != nil {
				return 0, err
			}
		}
		if err := children.directory.VerifyPath(); err != nil {
			return 0, err
		}
		if err := child.handle.verifyPath(); err != nil {
			return 0, err
		}
	}
	marked := make([]*windowsBoundHandle, 0, len(children.children))
	for index := range children.children {
		entry := &children.children[index]
		child := entry.handle
		if err := setWindowsDisposition(child.handle, true); err != nil {
			for _, prior := range marked {
				_ = setWindowsDisposition(prior.handle, false)
			}
			return 0, fmt.Errorf("cannot remove retained managed object %q: %w", entry.name, err)
		}
		marked = append(marked, child)
	}
	return len(marked), nil
}

func readWindowsBoundFile(handle syscall.Handle) ([]byte, error) {
	const maximumControlFileSize = 1 << 20
	result := make([]byte, 0, 4096)
	buffer := make([]byte, 4096)
	for {
		var read uint32
		err := syscall.ReadFile(handle, buffer, &read, nil)
		if read > 0 {
			if len(result)+int(read) > maximumControlFileSize {
				return nil, errors.New("the authenticated control file is unexpectedly large")
			}
			result = append(result, buffer[:read]...)
		}
		if err != nil {
			if errors.Is(err, syscall.ERROR_HANDLE_EOF) {
				return result, nil
			}
			return nil, err
		}
		if read == 0 {
			return result, nil
		}
	}
}

func objectBoundRemoveFile(path, description string, validate func([]byte) error, before func(string) error) error {
	parent, err := openWindowsBoundHandle(
		filepath.Dir(path),
		windowsFileListDirectory|windowsFileReadAttributes|windowsSynchronizeAccess,
		true,
	)
	if err != nil {
		return fmt.Errorf("cannot retain the parent of %s safely: %w", description, err)
	}
	defer parent.Close()
	target, err := openWindowsBoundHandle(
		path,
		windowsGenericRead|windowsDeleteAccess|windowsFileReadAttributes|windowsSynchronizeAccess,
		false,
	)
	if err != nil {
		return fmt.Errorf("cannot retain %s safely: %w", description, err)
	}
	defer target.Close()
	if target.identity.volume != parent.identity.volume {
		return errors.New("the authenticated control file is on a different filesystem volume from its retained parent")
	}
	data, err := readWindowsBoundFile(target.handle)
	if err != nil {
		return fmt.Errorf("cannot read the retained %s: %w", description, err)
	}
	if validate != nil {
		if err = validate(data); err != nil {
			return err
		}
	}
	if before != nil {
		if err = before(path); err != nil {
			return err
		}
	}
	if err = parent.verifyPath(); err != nil {
		return err
	}
	if err = target.verifyPath(); err != nil {
		return err
	}
	return setWindowsDisposition(target.handle, true)
}

type windowsTreeNode struct {
	handle *windowsBoundHandle
}

func bindWindowsTree(handle *windowsBoundHandle, nodes *[]windowsTreeNode) error {
	entries, err := handle.file.ReadDir(-1)
	if err != nil {
		handle.Close()
		return err
	}
	for _, entry := range entries {
		childPath := filepath.Join(handle.path, entry.Name())
		child, childErr := openWindowsBoundHandle(
			childPath,
			windowsDeleteAccess|windowsFileListDirectory|windowsFileReadAttributes|windowsSynchronizeAccess,
			entry.IsDir(),
		)
		if childErr != nil {
			handle.Close()
			return childErr
		}
		if child.identity.volume != handle.identity.volume {
			child.Close()
			handle.Close()
			return errors.New("migration cleanup found a filesystem object on a different volume")
		}
		if child.dir {
			if childErr = bindWindowsTree(child, nodes); childErr != nil {
				handle.Close()
				return childErr
			}
		} else {
			*nodes = append(*nodes, windowsTreeNode{handle: child})
		}
	}
	*nodes = append(*nodes, windowsTreeNode{handle: handle})
	return nil
}

func objectBoundRemoveTree(root string, validate func() error, before func(string) error) error {
	parent, err := openWindowsBoundHandle(
		filepath.Dir(root),
		windowsFileListDirectory|windowsFileReadAttributes|windowsSynchronizeAccess,
		true,
	)
	if err != nil {
		return err
	}
	defer parent.Close()
	nodes := []windowsTreeNode{}
	rootHandle, err := openWindowsBoundHandle(
		root,
		windowsDeleteAccess|windowsFileListDirectory|windowsFileReadAttributes|windowsSynchronizeAccess,
		true,
	)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if rootHandle.identity.volume != parent.identity.volume {
		rootHandle.Close()
		return errors.New("the migration staging folder is on a different filesystem volume from its retained parent")
	}
	if validate != nil {
		if err = validate(); err != nil {
			rootHandle.Close()
			return err
		}
	}
	if err = bindWindowsTree(rootHandle, &nodes); err != nil {
		for index := range nodes {
			_ = nodes[index].handle.Close()
		}
		return err
	}
	defer func() {
		for index := range nodes {
			_ = nodes[index].handle.Close()
		}
	}()
	if err = parent.verifyPath(); err != nil {
		return err
	}
	for index := range nodes {
		if err = nodes[index].handle.verifyPath(); err != nil {
			return err
		}
	}
	if before != nil {
		if err = before(root); err != nil {
			return err
		}
	}
	if err = parent.verifyPath(); err != nil {
		return err
	}
	for index := range nodes {
		handle := nodes[index].handle
		if err = setWindowsDisposition(handle.handle, true); err != nil {
			return fmt.Errorf("cannot remove retained migration object %q: %w", handle.path, err)
		}
		if err = handle.Close(); err != nil {
			return fmt.Errorf("cannot close removed migration object %q: %w", handle.path, err)
		}
	}
	return nil
}

func objectBoundRename(source, destination string, validate func() error, before func(string, string) error) error {
	if !sameFilesystemPath(filepath.Dir(source), filepath.Dir(destination)) {
		return errors.New("the migration rename endpoints are not direct siblings")
	}
	parent, err := openWindowsBoundHandle(
		filepath.Dir(source),
		windowsFileTraverse|windowsFileReadAttributes|windowsSynchronizeAccess,
		true,
	)
	if err != nil {
		return err
	}
	defer parent.Close()
	sourceHandle, err := openWindowsBoundHandle(
		source,
		windowsDeleteAccess|windowsFileListDirectory|windowsFileReadAttributes|windowsSynchronizeAccess,
		true,
	)
	if err != nil {
		return err
	}
	defer sourceHandle.Close()
	if sourceHandle.identity.volume != parent.identity.volume {
		return errors.New("the migration source and retained parent are not on the same filesystem volume")
	}
	if validate != nil {
		if err = validate(); err != nil {
			return err
		}
	}
	if err = parent.verifyPath(); err != nil {
		return err
	}
	if err = sourceHandle.verifyPath(); err != nil {
		return err
	}
	if _, statErr := os.Lstat(destination); statErr == nil {
		return errors.New("the migration rename destination already exists")
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	if before != nil {
		if err = before(source, destination); err != nil {
			return err
		}
	}
	if err = parent.verifyPath(); err != nil {
		return err
	}
	if err = sourceHandle.verifyPath(); err != nil {
		return err
	}
	if err = renameWindowsHandle(sourceHandle.handle, parent.handle, destination); err != nil {
		return err
	}
	destinationHandle, openErr := openWindowsBoundHandle(
		destination,
		windowsFileReadAttributes|windowsSynchronizeAccess,
		true,
	)
	if openErr != nil {
		_ = renameWindowsHandle(sourceHandle.handle, parent.handle, source)
		return fmt.Errorf("cannot verify the renamed migration folder: %w", openErr)
	}
	defer destinationHandle.Close()
	if destinationHandle.identity != sourceHandle.identity {
		_ = renameWindowsHandle(sourceHandle.handle, parent.handle, source)
		return errors.New("the migration rename did not move the authenticated filesystem object")
	}
	return nil
}
