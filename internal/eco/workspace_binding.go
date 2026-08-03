package eco

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type rawWorkspaceLifecycleLock interface {
	Close() error
}

type boundWorkspaceObjects interface {
	ReadFile(string) ([]byte, error)
	WriteFileAtomic(string, []byte, uint32) error
	Verify() error
	Close() error
}

type lifecycleRegistryEntry struct {
	lock rawWorkspaceLifecycleLock
	refs int
}

type workspaceLifecycleLease struct {
	key    string
	closed bool
}

var workspaceLifecycleRegistry = struct {
	sync.Mutex
	entries map[string]*lifecycleRegistryEntry
}{entries: make(map[string]*lifecycleRegistryEntry)}

func lifecycleRegistryKey(root string) (string, error) {
	absolute, err := normaliseWorkspaceRoot(root)
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		absolute = strings.ToLower(absolute)
	}
	return filepath.Clean(absolute), nil
}

func workspaceLifecycleLockName(root string) string {
	base := filepath.Base(filepath.Clean(root))
	if runtime.GOOS == "windows" {
		base = strings.ToLower(base)
	}
	digest := sha256.Sum256([]byte(base))
	return ".eco-workspace-lifecycle-" + hex.EncodeToString(digest[:12]) + ".lock"
}

func acquireWorkspaceLifecycleLease(root string) (*workspaceLifecycleLease, error) {
	key, err := lifecycleRegistryKey(root)
	if err != nil {
		return nil, err
	}
	workspaceLifecycleRegistry.Lock()
	defer workspaceLifecycleRegistry.Unlock()
	if existing := workspaceLifecycleRegistry.entries[key]; existing != nil {
		existing.refs++
		return &workspaceLifecycleLease{key: key}, nil
	}
	lock, err := acquireRawWorkspaceLifecycleLock(key, workspaceLifecycleLockName(key))
	if err != nil {
		return nil, fmt.Errorf("this workspace is already open or changing in another ECO process; close it there before continuing: %w", err)
	}
	workspaceLifecycleRegistry.entries[key] = &lifecycleRegistryEntry{lock: lock, refs: 1}
	return &workspaceLifecycleLease{key: key}, nil
}

func (lease *workspaceLifecycleLease) Close() error {
	if lease == nil || lease.closed {
		return nil
	}
	workspaceLifecycleRegistry.Lock()
	defer workspaceLifecycleRegistry.Unlock()
	entry := workspaceLifecycleRegistry.entries[lease.key]
	if entry == nil || entry.refs < 1 {
		lease.closed = true
		return errors.New("the workspace lifecycle lock was lost")
	}
	entry.refs--
	lease.closed = true
	if entry.refs > 0 {
		return nil
	}
	delete(workspaceLifecycleRegistry.entries, lease.key)
	return entry.lock.Close()
}

func attachWorkspaceGuards(v *Vault) error {
	if v == nil {
		return errors.New("no workspace is available to retain")
	}
	lease, err := acquireWorkspaceLifecycleLease(v.Root)
	if err != nil {
		return err
	}
	binding, err := openBoundWorkspaceObjects(v.Root)
	if err != nil {
		_ = lease.Close()
		return err
	}
	v.lifecycle = lease
	v.binding = binding
	return nil
}

func (v *Vault) releaseWorkspaceBinding() error {
	if v == nil || v.binding == nil {
		return nil
	}
	err := v.binding.Close()
	v.binding = nil
	return err
}

func (v *Vault) rebindWorkspaceObjects() error {
	if v == nil {
		return errors.New("no workspace is available to retain")
	}
	if err := v.releaseWorkspaceBinding(); err != nil {
		return err
	}
	binding, err := openBoundWorkspaceObjects(v.Root)
	if err != nil {
		return err
	}
	v.binding = binding
	return nil
}
