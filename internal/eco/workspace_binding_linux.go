//go:build linux && amd64

package eco

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

type linuxWorkspaceLifecycleLock struct {
	parent int
	lock   int
}

func acquireRawWorkspaceLifecycleLock(root, lockName string) (rawWorkspaceLifecycleLock, error) {
	parent, err := syscall.Open(filepath.Dir(root), syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	lock, err := syscall.Openat(parent, lockName, syscall.O_RDWR|syscall.O_CREAT|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		_ = syscall.Close(parent)
		return nil, err
	}
	identity, err := linuxIdentity(lock)
	if err != nil || !identity.isRegular() {
		_ = syscall.Close(lock)
		_ = syscall.Close(parent)
		if err != nil {
			return nil, err
		}
		return nil, errors.New("the workspace lifecycle lock is not a normal file")
	}
	if err = syscall.Flock(lock, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = syscall.Close(lock)
		_ = syscall.Close(parent)
		return nil, err
	}
	return &linuxWorkspaceLifecycleLock{parent: parent, lock: lock}, nil
}

func (lock *linuxWorkspaceLifecycleLock) Close() error {
	if lock == nil {
		return nil
	}
	var result error
	if lock.lock >= 0 {
		result = errors.Join(result, syscall.Flock(lock.lock, syscall.LOCK_UN), syscall.Close(lock.lock))
		lock.lock = -1
	}
	if lock.parent >= 0 {
		result = errors.Join(result, syscall.Close(lock.parent))
		lock.parent = -1
	}
	return result
}

type linuxWorkspaceFile struct {
	fd       int
	identity linuxObjectIdentity
}

type linuxWorkspaceObjects struct {
	path            string
	parent          int
	parentIdentity  linuxObjectIdentity
	root            int
	rootIdentity    linuxObjectIdentity
	objects         int
	objectsIdentity linuxObjectIdentity
	files           map[string]linuxWorkspaceFile
	missing         map[string]bool
}

func openBoundWorkspaceObjects(root string) (boundWorkspaceObjects, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	bound := &linuxWorkspaceObjects{path: filepath.Clean(absolute), parent: -1, root: -1, objects: -1, files: make(map[string]linuxWorkspaceFile), missing: make(map[string]bool)}
	bound.parent, err = syscall.Open(filepath.Dir(bound.path), syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	bound.parentIdentity, err = linuxIdentity(bound.parent)
	if err != nil {
		bound.Close()
		return nil, err
	}
	bound.root, bound.rootIdentity, err = openLinuxDirectoryAt(bound.parent, filepath.Base(bound.path))
	if err != nil {
		bound.Close()
		return nil, fmt.Errorf("cannot retain the workspace folder safely: %w", err)
	}
	bound.objects, bound.objectsIdentity, err = openLinuxDirectoryAt(bound.root, "objects")
	if err != nil {
		bound.Close()
		return nil, fmt.Errorf("cannot retain the encrypted object folder safely: %w", err)
	}
	if bound.objectsIdentity.device != bound.rootIdentity.device || bound.objectsIdentity.mount != bound.rootIdentity.mount {
		bound.Close()
		return nil, errors.New("the encrypted object folder is on another filesystem mount")
	}
	for _, name := range []string{"vault.key", "workspace.ecodb", workspaceIdentityFile} {
		fd, openErr := syscall.Openat(bound.root, name, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
		if openErr != nil {
			if errors.Is(openErr, syscall.ENOENT) && name == workspaceIdentityFile {
				bound.missing[name] = true
				continue
			}
			bound.Close()
			return nil, fmt.Errorf("cannot retain workspace control file %q safely: %w", name, openErr)
		}
		identity, identityErr := linuxIdentity(fd)
		if identityErr != nil {
			_ = syscall.Close(fd)
			bound.Close()
			return nil, fmt.Errorf("cannot identify workspace control file %q safely: %w", name, identityErr)
		}
		if !identity.isRegular() || identity.device != bound.rootIdentity.device || identity.mount != bound.rootIdentity.mount {
			_ = syscall.Close(fd)
			bound.Close()
			return nil, fmt.Errorf("workspace control file %q is not a normal file on the workspace mount", name)
		}
		bound.files[name] = linuxWorkspaceFile{fd: fd, identity: identity}
	}
	return bound, nil
}

func readLinuxWorkspaceFile(fd int) ([]byte, error) {
	var stat syscall.Stat_t
	if err := syscall.Fstat(fd, &stat); err != nil {
		return nil, err
	}
	if stat.Size < 0 || uint64(stat.Size) > uint64(^uint(0)>>1) {
		return nil, errors.New("the workspace control file is too large to read safely")
	}
	data := make([]byte, int(stat.Size))
	offset := 0
	for offset < len(data) {
		read, err := syscall.Pread(fd, data[offset:], int64(offset))
		offset += read
		if err != nil {
			return nil, err
		}
		if read == 0 {
			return nil, io.ErrUnexpectedEOF
		}
	}
	return data, nil
}

func (bound *linuxWorkspaceObjects) ReadFile(name string) ([]byte, error) {
	if bound.missing[name] {
		return nil, os.ErrNotExist
	}
	file, ok := bound.files[name]
	if !ok {
		return nil, errors.New("the requested file is not part of the retained workspace")
	}
	return readLinuxWorkspaceFile(file.fd)
}

func (bound *linuxWorkspaceObjects) Verify() error {
	if err := verifyLinuxNameAt(bound.parent, filepath.Base(bound.path), bound.rootIdentity); err != nil {
		return fmt.Errorf("workspace folder changed after authentication: %w", err)
	}
	if err := verifyLinuxNameAt(bound.root, "objects", bound.objectsIdentity); err != nil {
		return fmt.Errorf("encrypted object folder changed after authentication: %w", err)
	}
	for name, file := range bound.files {
		if err := verifyLinuxNameAt(bound.root, name, file.identity); err != nil {
			return fmt.Errorf("workspace control file %q changed after authentication: %w", name, err)
		}
	}
	for name := range bound.missing {
		fd, _, err := openLinuxPathAt(bound.root, name)
		if err == nil {
			_ = syscall.Close(fd)
			return fmt.Errorf("workspace control file %q appeared after authentication", name)
		}
		if !errors.Is(err, syscall.ENOENT) {
			return err
		}
	}
	return nil
}

func writeAllLinux(fd int, data []byte) error {
	for len(data) > 0 {
		written, err := syscall.Write(fd, data)
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
		data = data[written:]
	}
	return nil
}

func (bound *linuxWorkspaceObjects) WriteFileAtomic(name string, data []byte, mode uint32) error {
	if name != "workspace.ecodb" && name != workspaceIdentityFile {
		return errors.New("the requested workspace file cannot be replaced through this transaction")
	}
	if err := bound.Verify(); err != nil {
		return err
	}
	tmpName := ".eco-write-" + NewID("FS") + ".tmp"
	fd, err := syscall.Openat(bound.root, tmpName, syscall.O_RDWR|syscall.O_CREAT|syscall.O_EXCL|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, mode)
	if err != nil {
		return err
	}
	cleanup := func() {
		_ = syscall.Close(fd)
		_ = unlinkLinuxAt(bound.root, tmpName, false)
	}
	if err = writeAllLinux(fd, data); err != nil {
		cleanup()
		return err
	}
	if err = syscall.Fsync(fd); err != nil {
		cleanup()
		return err
	}
	identity, err := linuxIdentity(fd)
	if err != nil {
		cleanup()
		return err
	}
	if err = syscall.Renameat(bound.root, tmpName, bound.root, name); err != nil {
		cleanup()
		return err
	}
	if err = syscall.Fsync(bound.root); err != nil {
		_ = syscall.Close(fd)
		return err
	}
	if previous, ok := bound.files[name]; ok {
		_ = syscall.Close(previous.fd)
	}
	bound.files[name] = linuxWorkspaceFile{fd: fd, identity: identity}
	delete(bound.missing, name)
	return verifyLinuxNameAt(bound.root, name, identity)
}

func (bound *linuxWorkspaceObjects) Close() error {
	if bound == nil {
		return nil
	}
	var result error
	for name, file := range bound.files {
		if file.fd >= 0 {
			result = errors.Join(result, syscall.Close(file.fd))
		}
		delete(bound.files, name)
	}
	if bound.objects >= 0 {
		result = errors.Join(result, syscall.Close(bound.objects))
		bound.objects = -1
	}
	if bound.root >= 0 {
		result = errors.Join(result, syscall.Close(bound.root))
		bound.root = -1
	}
	if bound.parent >= 0 {
		result = errors.Join(result, syscall.Close(bound.parent))
		bound.parent = -1
	}
	return result
}
