package eco

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const workspaceOwnerFilename = ".eco-owner-v2.lock"

var ErrWorkspaceInUse = errors.New("workspace is already open for writing")

type workspaceObjectIdentity struct {
	Volume uint64
	File   uint64
}

func (i workspaceObjectIdentity) valid() bool { return i.Volume != 0 || i.File != 0 }
func (i workspaceObjectIdentity) equal(other workspaceObjectIdentity) bool {
	return i.valid() && other.valid() && i == other
}

type workspaceOwnerLease struct {
	root      string
	identity  workspaceObjectIdentity
	lock      platformWorkspaceLock
	once      sync.Once
	closeErr  error
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
	l.once.Do(func() {
		if l.lock != nil {
			l.closeErr = l.lock.Close()
		}
	})
	return l.closeErr
}

func (l *workspaceOwnerLease) revalidate() error {
	if l == nil || l.lock == nil {
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

type platformWorkspaceLock interface {
	Close() error
}
