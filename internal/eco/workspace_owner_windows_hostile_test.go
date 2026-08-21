//go:build windows

package eco

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

// workspaceSemaphoreNameForHostileTest derives the exact legacy named
// semaphore used by the vulnerable PR #72 implementation. The repair must
// remain immune to this preserved hostile setup.
func workspaceSemaphoreNameForHostileTest(root string) (string, error) {
	identity, err := platformWorkspaceObjectIdentity(root)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"Global\\ECO.Workspace.%016x.%016x.%s",
		identity.Volume,
		identity.File,
		strings.ToLower(workspaceOwnerFilename),
	), nil
}

// TestWorkspaceOwnerRejectsHostilePrecreatedSemaphoreCount is the exact
// regression for PR #86 / workflow 32485114343. The vulnerable implementation
// allowed two owners after this semaphore was pre-created with count 2.
func TestWorkspaceOwnerRejectsHostilePrecreatedSemaphoreCount(t *testing.T) {
	root := t.TempDir()
	name, err := workspaceSemaphoreNameForHostileTest(root)
	if err != nil {
		t.Fatal(err)
	}
	ptr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		t.Fatal(err)
	}

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	createSemaphore := kernel32.NewProc("CreateSemaphoreW")
	h, _, createErr := createSemaphore.Call(
		0,
		2,
		2,
		uintptr(unsafe.Pointer(ptr)),
	)
	if h == 0 {
		t.Fatalf("pre-create hostile workspace semaphore: %v", createErr)
	}
	foreign := syscall.Handle(h)
	defer syscall.CloseHandle(foreign)

	first, err := acquireWorkspaceRootOwner(root)
	if err != nil {
		t.Fatalf("first owner should acquire synthetic workspace: %v", err)
	}
	defer first.Close()

	second, err := acquireWorkspaceRootOwner(root)
	if second != nil {
		_ = second.Close()
		t.Fatal("P0 regression: hostile legacy semaphore allowed two simultaneous workspace owners")
	}
	if !errors.Is(err, ErrWorkspaceInUse) {
		t.Fatalf("second acquisition error=%v; want ErrWorkspaceInUse", err)
	}
}
