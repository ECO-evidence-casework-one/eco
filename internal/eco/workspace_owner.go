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
	identity workspaceObjectIdentity
	lock     platformWorkspaceLock
	once     sync.Once
	closeErr error
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
	if err := creation.revalidate(); err != nil {
		return nil, err
	}
	if info, statErr := os.Stat(absolute); statErr == nil {
		if !info.IsDir() {
			return nil, errors.New("workspace root is not a directory")
		}
	} else if os.IsNotExist(statErr) {
		if err := os.Mkdir(absolute, 0700); err != nil {
			return nil, fmt.Errorf("create workspace root: %w", err)
		}
	} else {
		return nil, fmt.Errorf("inspect workspace root after creation claim: %w", statErr)
	}
	owner, err := acquireWorkspaceRootOwner(absolute)
	if err != nil {
		return nil, err
	}
	if err := creation.revalidate(); err != nil {
		_ = owner.Close()
		return nil, err
	}
	return owner, nil
}

func acquireWorkspaceCreationOwner(root string) (*workspaceCreationLease, error) {
	parent := filepath.Dir(root)
	leaf := filepath.Base(root)
	if leaf == "." || leaf == string(filepath.Separator) || leaf == "" {
		return nil, errors.New("workspace root name is invalid")
	}
	info, err := os.Stat(parent)
	if err != nil {
		return nil, fmt.Errorf("inspect workspace parent: %w", err)
	}
	if !info.IsDir() {
		return nil, errors.New("workspace parent is not a directory")
	}
	identity, err := platformWorkspaceObjectIdentity(parent)
	if err != nil {
		return nil, fmt.Errorf("identify workspace parent: %w", err)
	}
	sum := sha256.Sum256([]byte(platformWorkspaceCreationKey(leaf)))
	lockName := workspaceCreationLockPrefix + hex.EncodeToString(sum[:16]) + ".lock"
	lock, err := platformAcquireWorkspaceLock(filepath.Join(parent, lockName))
	if err != nil {
		return nil, err
	}
	after, err := platformWorkspaceObjectIdentity(parent)
	if err != nil || !identity.equal(after) {
		_ = lock.Close()
		if err != nil {
			return nil, fmt.Errorf("revalidate workspace parent: %w", err)
		}
		return nil, errors.New("workspace parent changed while creation ownership was acquired")
	}
	return &workspaceCreationLease{parent: parent, root: root, identity: identity, lock: lock}, nil
}

func (l *workspaceCreationLease) Close() error {
	if l == nil {
		return nil
	}
	l.once.Do(func() {
		if l.lock != nil {
			l.closeErr = l.lock.Close()
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
		return fmt.Errorf("revalidate workspace parent: %w", err)
	}
	if !l.identity.equal(current) {
		return errors.New("workspace parent identity changed")
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

// retarget moves the routing path of a still-held lease after the exact owned
// directory object has been renamed. It never changes the locked object.
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
