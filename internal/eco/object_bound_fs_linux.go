//go:build linux && amd64

package eco

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	linuxOpenPath        = 0x200000
	linuxRenameNoReplace = 1
	linuxRemoveDirectory = 0x200
	linuxSysRenameat2    = 316
)

type linuxObjectIdentity struct {
	device uint64
	inode  uint64
	mode   uint32
	mount  uint64
}

func linuxMountID(fd int) (uint64, error) {
	// fdinfo supplies the mount identity of the retained descriptor itself;
	// comparing only st_dev would not detect a same-volume bind mount.
	data, err := os.ReadFile(fmt.Sprintf("/proc/self/fdinfo/%d", fd))
	if err != nil {
		return 0, fmt.Errorf("cannot read the retained filesystem handle's mount identity: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "mnt_id:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, "mnt_id:"))
		mount, parseErr := strconv.ParseUint(value, 10, 64)
		if parseErr != nil || mount == 0 {
			return 0, errors.New("the retained filesystem handle has an invalid mount identity")
		}
		return mount, nil
	}
	return 0, errors.New("the retained filesystem handle has no mount identity")
}

func linuxIdentity(fd int) (linuxObjectIdentity, error) {
	var information syscall.Stat_t
	if err := syscall.Fstat(fd, &information); err != nil {
		return linuxObjectIdentity{}, err
	}
	mount, err := linuxMountID(fd)
	if err != nil {
		return linuxObjectIdentity{}, err
	}
	return linuxObjectIdentity{
		device: uint64(information.Dev),
		inode:  information.Ino,
		mode:   information.Mode,
		mount:  mount,
	}, nil
}

func sameLinuxIdentity(left, right linuxObjectIdentity) bool { return left == right }

func (identity linuxObjectIdentity) isDirectory() bool {
	return identity.mode&syscall.S_IFMT == syscall.S_IFDIR
}

func (identity linuxObjectIdentity) isRegular() bool {
	return identity.mode&syscall.S_IFMT == syscall.S_IFREG
}

func openLinuxPathAt(parent int, name string) (int, linuxObjectIdentity, error) {
	fd, err := syscall.Openat(parent, name, linuxOpenPath|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return -1, linuxObjectIdentity{}, err
	}
	identity, err := linuxIdentity(fd)
	if err != nil {
		_ = syscall.Close(fd)
		return -1, linuxObjectIdentity{}, err
	}
	if identity.mode&syscall.S_IFMT == syscall.S_IFLNK {
		_ = syscall.Close(fd)
		return -1, linuxObjectIdentity{}, errors.New("the selected filesystem object is a symbolic link")
	}
	return fd, identity, nil
}

func openLinuxDirectoryAt(parent int, name string) (int, linuxObjectIdentity, error) {
	fd, err := syscall.Openat(parent, name, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return -1, linuxObjectIdentity{}, err
	}
	identity, err := linuxIdentity(fd)
	if err != nil {
		_ = syscall.Close(fd)
		return -1, linuxObjectIdentity{}, err
	}
	if !identity.isDirectory() {
		_ = syscall.Close(fd)
		return -1, linuxObjectIdentity{}, errors.New("the selected filesystem object is not a folder")
	}
	return fd, identity, nil
}

func verifyLinuxNameAt(parent int, name string, expected linuxObjectIdentity) error {
	fd, identity, err := openLinuxPathAt(parent, name)
	if err != nil {
		return err
	}
	_ = syscall.Close(fd)
	if !sameLinuxIdentity(identity, expected) {
		return errors.New("the retained filesystem object was replaced after authentication")
	}
	return nil
}

func renameLinuxAtNoReplace(oldParent int, oldName string, newParent int, newName string) error {
	oldPointer, err := syscall.BytePtrFromString(oldName)
	if err != nil {
		return err
	}
	newPointer, err := syscall.BytePtrFromString(newName)
	if err != nil {
		return err
	}
	_, _, errno := syscall.Syscall6(
		linuxSysRenameat2,
		uintptr(oldParent),
		uintptr(unsafe.Pointer(oldPointer)),
		uintptr(newParent),
		uintptr(unsafe.Pointer(newPointer)),
		linuxRenameNoReplace,
		0,
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func unlinkLinuxAt(parent int, name string, directory bool) error {
	pointer, err := syscall.BytePtrFromString(name)
	if err != nil {
		return err
	}
	flags := 0
	if directory {
		flags = linuxRemoveDirectory
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_UNLINKAT,
		uintptr(parent),
		uintptr(unsafe.Pointer(pointer)),
		uintptr(flags),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func isolateAndRemoveLinuxObject(parent int, name string, expected linuxObjectIdentity) error {
	// RENAME_NOREPLACE isolates one entry under the retained parent. The
	// isolated descriptor identity is checked before unlinkat; a mismatch is
	// compensated back to its original name.
	tombstone := ".eco-delete-" + NewID("FS")
	if err := renameLinuxAtNoReplace(parent, name, parent, tombstone); err != nil {
		return err
	}
	isolatedFD, isolatedIdentity, err := openLinuxPathAt(parent, tombstone)
	if err != nil {
		rollbackErr := renameLinuxAtNoReplace(parent, tombstone, parent, name)
		return errors.Join(err, rollbackErr)
	}
	_ = syscall.Close(isolatedFD)
	if !sameLinuxIdentity(isolatedIdentity, expected) {
		rollbackErr := renameLinuxAtNoReplace(parent, tombstone, parent, name)
		return errors.Join(errors.New("the destructive operation isolated a different filesystem object"), rollbackErr)
	}
	if err = unlinkLinuxAt(parent, tombstone, expected.isDirectory()); err != nil {
		rollbackErr := renameLinuxAtNoReplace(parent, tombstone, parent, name)
		return errors.Join(err, rollbackErr)
	}
	return nil
}

type linuxBoundObjectDirectory struct {
	path     string
	fd       int
	identity linuxObjectIdentity
}

func openBoundObjectDirectory(path, description string) (boundObjectDirectory, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	fd, err := syscall.Open(filepath.Clean(absolute), syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot retain %s safely: %w", description, err)
	}
	identity, err := linuxIdentity(fd)
	if err != nil {
		_ = syscall.Close(fd)
		return nil, err
	}
	return &linuxBoundObjectDirectory{path: filepath.Clean(absolute), fd: fd, identity: identity}, nil
}

func (directory *linuxBoundObjectDirectory) Close() error {
	if directory.fd < 0 {
		return nil
	}
	err := syscall.Close(directory.fd)
	directory.fd = -1
	return err
}

func (directory *linuxBoundObjectDirectory) VerifyPath() error {
	fd, err := syscall.Open(directory.path, linuxOpenPath|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("the retained directory is no longer at its authenticated path: %w", err)
	}
	identity, identityErr := linuxIdentity(fd)
	_ = syscall.Close(fd)
	if identityErr != nil {
		return identityErr
	}
	if !sameLinuxIdentity(identity, directory.identity) {
		return errors.New("the retained directory was replaced after authentication")
	}
	return nil
}

func (directory *linuxBoundObjectDirectory) SameFilesystem(other boundObjectDirectory) bool {
	compared, ok := other.(*linuxBoundObjectDirectory)
	return ok && directory.identity.device == compared.identity.device && directory.identity.mount == compared.identity.mount
}

type linuxBoundRegularChild struct {
	name     string
	fd       int
	identity linuxObjectIdentity
}

type linuxBoundRegularChildren struct {
	directory *linuxBoundObjectDirectory
	children  []linuxBoundRegularChild
}

func (directory *linuxBoundObjectDirectory) BindRegularChildren(names []string) (boundRegularChildren, error) {
	children := &linuxBoundRegularChildren{directory: directory}
	for _, name := range names {
		if !safeManagedObjectName(name) {
			children.Close()
			return nil, errors.New("the managed encrypted object name is invalid")
		}
		fd, identity, err := openLinuxPathAt(directory.fd, name)
		if err != nil {
			if errors.Is(err, syscall.ENOENT) {
				continue
			}
			children.Close()
			return nil, fmt.Errorf("cannot retain managed encrypted object %q safely: %w", name, err)
		}
		if !identity.isRegular() {
			_ = syscall.Close(fd)
			children.Close()
			return nil, fmt.Errorf("managed encrypted object %q is not a regular file", name)
		}
		if identity.device != directory.identity.device || identity.mount != directory.identity.mount {
			_ = syscall.Close(fd)
			children.Close()
			return nil, fmt.Errorf("managed encrypted object %q is on a different filesystem mount", name)
		}
		children.children = append(children.children, linuxBoundRegularChild{name: name, fd: fd, identity: identity})
	}
	return children, nil
}

func (children *linuxBoundRegularChildren) Close() error {
	var result error
	for index := range children.children {
		if children.children[index].fd >= 0 {
			result = errors.Join(result, syscall.Close(children.children[index].fd))
			children.children[index].fd = -1
		}
	}
	children.children = nil
	return result
}

func (children *linuxBoundRegularChildren) RemoveAll(before func(string) error) (int, error) {
	for index := range children.children {
		child := &children.children[index]
		if before != nil {
			if err := before(filepath.Join(children.directory.path, child.name)); err != nil {
				return 0, err
			}
		}
		if err := children.directory.VerifyPath(); err != nil {
			return 0, err
		}
		if err := verifyLinuxNameAt(children.directory.fd, child.name, child.identity); err != nil {
			return 0, err
		}
	}
	removed := 0
	for index := range children.children {
		child := &children.children[index]
		if err := isolateAndRemoveLinuxObject(children.directory.fd, child.name, child.identity); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func openLinuxParent(path string) (*linuxBoundObjectDirectory, error) {
	parent, err := openBoundObjectDirectory(filepath.Dir(path), "the authenticated parent folder")
	if err != nil {
		return nil, err
	}
	return parent.(*linuxBoundObjectDirectory), nil
}

func readLinuxBoundFile(fd int) ([]byte, error) {
	const maximumControlFileSize = 1 << 20
	result := make([]byte, 0, 4096)
	buffer := make([]byte, 4096)
	for offset := int64(0); ; {
		read, err := syscall.Pread(fd, buffer, offset)
		if read > 0 {
			if len(result)+read > maximumControlFileSize {
				return nil, errors.New("the authenticated control file is unexpectedly large")
			}
			result = append(result, buffer[:read]...)
			offset += int64(read)
		}
		if err != nil {
			return nil, err
		}
		if read == 0 {
			return result, nil
		}
	}
}

func objectBoundRemoveFile(path, description string, validate func([]byte) error, before func(string) error) error {
	parent, err := openLinuxParent(path)
	if err != nil {
		return err
	}
	defer parent.Close()
	inspectionFD, identity, err := openLinuxPathAt(parent.fd, filepath.Base(path))
	if err != nil {
		return fmt.Errorf("cannot retain %s safely: %w", description, err)
	}
	defer syscall.Close(inspectionFD)
	if !identity.isRegular() {
		return fmt.Errorf("%s is not a regular file", description)
	}
	if identity.device != parent.identity.device || identity.mount != parent.identity.mount {
		return fmt.Errorf("%s is on a different filesystem mount from its retained parent", description)
	}
	fd, err := syscall.Openat(parent.fd, filepath.Base(path), syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("cannot open the retained %s for authenticated reading: %w", description, err)
	}
	defer syscall.Close(fd)
	readIdentity, err := linuxIdentity(fd)
	if err != nil {
		return err
	}
	if !sameLinuxIdentity(readIdentity, identity) {
		return errors.New("the authenticated control file changed while it was being retained")
	}
	data, err := readLinuxBoundFile(fd)
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
	if err = parent.VerifyPath(); err != nil {
		return err
	}
	if err = verifyLinuxNameAt(parent.fd, filepath.Base(path), identity); err != nil {
		return err
	}
	return isolateAndRemoveLinuxObject(parent.fd, filepath.Base(path), identity)
}

type linuxTreeNode struct {
	name     string
	parent   int
	fd       int
	identity linuxObjectIdentity
}

func bindLinuxTree(parent int, name string, nodes *[]linuxTreeNode) error {
	parentIdentity, err := linuxIdentity(parent)
	if err != nil {
		return err
	}
	fd, identity, err := openLinuxDirectoryAt(parent, name)
	if err != nil {
		return err
	}
	if identity.device != parentIdentity.device || identity.mount != parentIdentity.mount {
		_ = syscall.Close(fd)
		return errors.New("migration cleanup found a nested filesystem mount")
	}
	duplicate, err := syscall.Dup(fd)
	if err != nil {
		_ = syscall.Close(fd)
		return err
	}
	reader := os.NewFile(uintptr(duplicate), name)
	entries, err := reader.ReadDir(-1)
	_ = reader.Close()
	if err != nil {
		_ = syscall.Close(fd)
		return err
	}
	for _, entry := range entries {
		childFD, childIdentity, childErr := openLinuxPathAt(fd, entry.Name())
		if childErr != nil {
			_ = syscall.Close(fd)
			return childErr
		}
		if childIdentity.device != identity.device || childIdentity.mount != identity.mount {
			_ = syscall.Close(childFD)
			_ = syscall.Close(fd)
			return errors.New("migration cleanup found a nested filesystem mount")
		}
		if childIdentity.isDirectory() {
			_ = syscall.Close(childFD)
			if childErr = bindLinuxTree(fd, entry.Name(), nodes); childErr != nil {
				_ = syscall.Close(fd)
				return childErr
			}
		} else if childIdentity.isRegular() {
			*nodes = append(*nodes, linuxTreeNode{name: entry.Name(), parent: fd, fd: childFD, identity: childIdentity})
		} else {
			_ = syscall.Close(childFD)
			_ = syscall.Close(fd)
			return errors.New("migration cleanup found a non-regular filesystem object")
		}
	}
	*nodes = append(*nodes, linuxTreeNode{name: name, parent: parent, fd: fd, identity: identity})
	return nil
}

func objectBoundRemoveTree(root string, validate func() error, before func(string) error) error {
	parent, err := openLinuxParent(root)
	if err != nil {
		return err
	}
	defer parent.Close()
	nodes := []linuxTreeNode{}
	if err = bindLinuxTree(parent.fd, filepath.Base(root), &nodes); err != nil {
		for index := range nodes {
			if nodes[index].fd >= 0 {
				_ = syscall.Close(nodes[index].fd)
			}
		}
		if errors.Is(err, syscall.ENOENT) {
			return nil
		}
		return err
	}
	defer func() {
		for index := range nodes {
			if nodes[index].fd >= 0 {
				_ = syscall.Close(nodes[index].fd)
			}
		}
	}()
	if validate != nil {
		if err = validate(); err != nil {
			return err
		}
	}
	if err = parent.VerifyPath(); err != nil {
		return err
	}
	for index := range nodes {
		if err = verifyLinuxNameAt(nodes[index].parent, nodes[index].name, nodes[index].identity); err != nil {
			return err
		}
	}
	if before != nil {
		if err = before(root); err != nil {
			return err
		}
	}
	if err = parent.VerifyPath(); err != nil {
		return err
	}
	for index := range nodes {
		if err = verifyLinuxNameAt(nodes[index].parent, nodes[index].name, nodes[index].identity); err != nil {
			return err
		}
	}
	for index := range nodes {
		node := &nodes[index]
		if err = isolateAndRemoveLinuxObject(node.parent, node.name, node.identity); err != nil {
			return err
		}
		if err = syscall.Close(node.fd); err != nil {
			return err
		}
		node.fd = -1
	}
	return nil
}

func objectBoundRename(source, destination string, validate func() error, before func(string, string) error) error {
	if !sameFilesystemPath(filepath.Dir(source), filepath.Dir(destination)) {
		return errors.New("the migration rename endpoints are not direct siblings")
	}
	parent, err := openLinuxParent(source)
	if err != nil {
		return err
	}
	defer parent.Close()
	sourceFD, sourceIdentity, err := openLinuxDirectoryAt(parent.fd, filepath.Base(source))
	if err != nil {
		return err
	}
	defer syscall.Close(sourceFD)
	if sourceIdentity.device != parent.identity.device || sourceIdentity.mount != parent.identity.mount {
		return errors.New("the migration source is on a different filesystem mount from its retained parent")
	}
	if validate != nil {
		if err = validate(); err != nil {
			return err
		}
	}
	if err = parent.VerifyPath(); err != nil {
		return err
	}
	if err = verifyLinuxNameAt(parent.fd, filepath.Base(source), sourceIdentity); err != nil {
		return err
	}
	destinationFD, _, destinationErr := openLinuxPathAt(parent.fd, filepath.Base(destination))
	if destinationErr == nil {
		_ = syscall.Close(destinationFD)
		return errors.New("the migration rename destination already exists")
	}
	if !errors.Is(destinationErr, syscall.ENOENT) {
		return destinationErr
	}
	if before != nil {
		if err = before(source, destination); err != nil {
			return err
		}
	}
	if err = parent.VerifyPath(); err != nil {
		return err
	}
	if err = verifyLinuxNameAt(parent.fd, filepath.Base(source), sourceIdentity); err != nil {
		return err
	}
	if err = renameLinuxAtNoReplace(parent.fd, filepath.Base(source), parent.fd, filepath.Base(destination)); err != nil {
		return err
	}
	destinationFD, destinationIdentity, openErr := openLinuxPathAt(parent.fd, filepath.Base(destination))
	if openErr != nil {
		rollbackErr := renameLinuxAtNoReplace(parent.fd, filepath.Base(destination), parent.fd, filepath.Base(source))
		return errors.Join(fmt.Errorf("cannot verify the renamed migration folder: %w", openErr), rollbackErr)
	}
	_ = syscall.Close(destinationFD)
	if !sameLinuxIdentity(destinationIdentity, sourceIdentity) {
		rollbackErr := renameLinuxAtNoReplace(parent.fd, filepath.Base(destination), parent.fd, filepath.Base(source))
		return errors.Join(errors.New("the migration rename did not move the authenticated filesystem object"), rollbackErr)
	}
	return nil
}
