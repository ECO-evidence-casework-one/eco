package eco

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const workspaceOwnerFilename = ".eco-owner-v2.lock"

var ErrWorkspaceInUse = errors.New("workspace is already open for writing")

const workspaceCreationLockPrefix = ".eco-create-v2-"

type workspaceCreationLease struct {
	parent   string
	root     string
	missing  []string
	identity workspaceObjectIdentity
	lock     platformWorkspaceLock
	once     sync.Once
	closeErr error
}

type createdWorkspaceDirectory struct {
	path     string
	identity workspaceObjectIdentity
}

func acquireOrCreateWorkspaceRootOwner(root string) (*workspaceOwnerLease, error) {
	if root == "" {
		return nil, errors.New("empty workspace root")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}
	if info, statErr := os.Stat(absolute); statErr == nil {
		if !info.IsDir() {
			return nil, errors.New("workspace root is not a directory")
		}
		return acquireWorkspaceRootOwner(absolute)
	} else if !os.IsNotExist(statErr) {
		return nil, fmt.Errorf("inspect workspace root: %w", statErr)
	}

	creation, err := acquireWorkspaceCreationOwner(absolute)
	if err != nil {
		return nil, err
	}
	defer creation.Close()
	created, err := creation.createMissingHierarchy()
	if err != nil {
		cleanupCreatedWorkspaceDirectories(created)
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			cleanupCreatedWorkspaceDirectories(created)
		}
	}()
	owner, err := acquireWorkspaceRootOwner(absolute)
	if err != nil {
		return nil, err
	}
	if err := creation.revalidate(); err != nil {
		_ = owner.Close()
		return nil, err
	}
	committed = true
	return owner, nil
}

func acquireWorkspaceCreationOwner(root string) (*workspaceCreationLease, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}
	parent, missing, err := workspaceCreationAnchor(absolute)
	if err != nil {
		return nil, err
	}
	if len(missing) == 0 {
		return nil, errors.New("workspace root already exists")
	}
	identity, err := platformWorkspaceObjectIdentity(parent)
	if err != nil {
		return nil, fmt.Errorf("identify workspace creation anchor: %w", err)
	}
	relative := filepath.Join(missing...)
	sum := sha256.Sum256([]byte(platformWorkspaceCreationKey(relative)))
	lockName := workspaceCreationLockPrefix + hex.EncodeToString(sum[:16]) + ".lock"
	lock, err := platformAcquireWorkspaceLock(filepath.Join(parent, lockName))
	if err != nil {
		return nil, err
	}
	after, err := platformWorkspaceObjectIdentity(parent)
	if err != nil || !identity.equal(after) {
		_ = lock.Close()
		if err != nil {
			return nil, fmt.Errorf("revalidate workspace creation anchor: %w", err)
		}
		return nil, errors.New("workspace creation anchor changed while ownership was acquired")
	}
	return &workspaceCreationLease{parent: parent, root: absolute, missing: missing, identity: identity, lock: lock}, nil
}

func workspaceCreationAnchor(root string) (string, []string, error) {
	current := root
	missingReverse := make([]string, 0, 4)
	for {
		info, err := os.Stat(current)
		if err == nil {
			if !info.IsDir() {
				return "", nil, errors.New("workspace creation anchor is not a directory")
			}
			missing := make([]string, len(missingReverse))
			for i := range missingReverse {
				missing[len(missingReverse)-1-i] = missingReverse[i]
			}
			return current, missing, nil
		}
		if !os.IsNotExist(err) {
			return "", nil, fmt.Errorf("inspect workspace creation route: %w", err)
		}
		parent := filepath.Dir(current)
		leaf := filepath.Base(current)
		if parent == current || leaf == "." || leaf == string(filepath.Separator) || leaf == "" {
			return "", nil, errors.New("workspace creation route has no existing parent")
		}
		missingReverse = append(missingReverse, leaf)
		current = parent
	}
}

func (l *workspaceCreationLease) createMissingHierarchy() ([]createdWorkspaceDirectory, error) {
	if err := l.revalidate(); err != nil {
		return nil, err
	}
	created := make([]createdWorkspaceDirectory, 0, len(l.missing))
	current := l.parent
	for _, leaf := range l.missing {
		if leaf == "" || leaf == "." || leaf == ".." {
			return created, errors.New("workspace creation route contains an invalid component")
		}
		if err := l.revalidate(); err != nil {
			return created, err
		}
		current = filepath.Join(current, leaf)
		info, err := os.Lstat(current)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return created, errors.New("workspace creation route contains a non-directory or symbolic link")
			}
			continue
		}
		if !os.IsNotExist(err) {
			return created, fmt.Errorf("inspect workspace creation component: %w", err)
		}
		if err := os.Mkdir(current, 0700); err != nil {
			return created, fmt.Errorf("create workspace directory: %w", err)
		}
		identity, err := platformWorkspaceObjectIdentity(current)
		if err != nil {
			return created, fmt.Errorf("identify created workspace directory: %w", err)
		}
		created = append(created, createdWorkspaceDirectory{path: current, identity: identity})
	}
	if err := l.revalidate(); err != nil {
		return created, err
	}
	return created, nil
}

func cleanupCreatedWorkspaceDirectories(created []createdWorkspaceDirectory) {
	for i := len(created) - 1; i >= 0; i-- {
		entry := created[i]
		info, err := os.Lstat(entry.path)
		if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			continue
		}
		identity, err := platformWorkspaceObjectIdentity(entry.path)
		if err != nil || !entry.identity.equal(identity) {
			continue
		}
		_ = os.Remove(entry.path)
	}
}

func (l *workspaceCreationLease) Close() error {
	if l == nil {
		return nil
	}
	l.once.Do(func() {
		if l.lock != nil {
			l.closeErr = l.lock.Close()
			l.lock = nil
		}
	})
	return l.closeErr
}

func (l *workspaceCreationLease) revalidate() error {
	if l == nil || l.lock == nil {
		return errors.New("workspace creation ownership is not held")
	}
	current, err := platformWorkspaceObjectIdentity(l.parent)
	if err != nil {
		return fmt.Errorf("revalidate workspace creation anchor: %w", err)
	}
	if !l.identity.equal(current) {
		return errors.New("workspace creation anchor identity changed")
	}
	return nil
}

type workspaceObjectIdentity struct {
	Volume uint64
	File   uint64
}

func (i workspaceObjectIdentity) valid() bool { return i.Volume != 0 || i.File != 0 }
func (i workspaceObjectIdentity) equal(other workspaceObjectIdentity) bool {
	return i.valid() && other.valid() && i == other
}

type workspaceOwnerLease struct {
	routeMu  sync.Mutex
	root     string
	identity workspaceObjectIdentity
	lock     platformWorkspaceLock
	once     sync.Once
	closeErr error
}

func acquireWorkspaceRootOwner(root string) (*workspaceOwnerLease, error) {
	if root == "" {
		return nil, errors.New("empty workspace root")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return nil, fmt.Errorf("inspect workspace root: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("workspace root is not a directory")
	}
	identity, err := platformWorkspaceObjectIdentity(absolute)
	if err != nil {
		return nil, fmt.Errorf("identify workspace root: %w", err)
	}
	lock, err := platformAcquireWorkspaceLock(filepath.Join(absolute, workspaceOwnerFilename))
	if err != nil {
		return nil, err
	}
	after, err := platformWorkspaceObjectIdentity(absolute)
	if err != nil || !identity.equal(after) {
		_ = lock.Close()
		if err != nil {
			return nil, fmt.Errorf("revalidate workspace root: %w", err)
		}
		return nil, errors.New("workspace root changed while ownership was acquired")
	}
	return &workspaceOwnerLease{root: absolute, identity: identity, lock: lock}, nil
}

func (l *workspaceOwnerLease) Close() error {
	if l == nil {
		return nil
	}
	l.routeMu.Lock()
	defer l.routeMu.Unlock()
	l.once.Do(func() {
		if l.lock != nil {
			l.closeErr = l.lock.Close()
			l.lock = nil
		}
	})
	return l.closeErr
}

func (l *workspaceOwnerLease) revalidate() error {
	if l == nil {
		return errors.New("workspace ownership is not held")
	}
	l.routeMu.Lock()
	defer l.routeMu.Unlock()
	return l.revalidateLocked()
}

func (l *workspaceOwnerLease) revalidateLocked() error {
	if l.lock == nil {
		return errors.New("workspace ownership is not held")
	}
	current, err := platformWorkspaceObjectIdentity(l.root)
	if err != nil {
		return fmt.Errorf("revalidate workspace root: %w", err)
	}
	if !l.identity.equal(current) {
		return errors.New("workspace root identity changed")
	}
	return nil
}

func (l *workspaceOwnerLease) retarget(root string) error {
	if l == nil {
		return errors.New("workspace ownership is not held")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve retargeted workspace root: %w", err)
	}
	l.routeMu.Lock()
	defer l.routeMu.Unlock()
	if l.lock == nil {
		return errors.New("workspace ownership is not held")
	}
	current, err := platformWorkspaceObjectIdentity(absolute)
	if err != nil {
		return fmt.Errorf("identify retargeted workspace root: %w", err)
	}
	if !l.identity.equal(current) {
		return errors.New("retargeted workspace root is not the owned object")
	}
	l.root = absolute
	return l.revalidateLocked()
}

type platformWorkspaceLock interface{ Close() error }
