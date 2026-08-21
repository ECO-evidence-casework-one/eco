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

// workspaceSemaphoreNameForHostileTest derives the exact named semaphore used
// by platformAcquireWorkspaceLock for a workspace root. This test helper is
// intentionally kept outside production code.
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

// TestWorkspaceOwnerRejectsHostilePrecreatedSemaphoreCount proves the Windows
// ownership primitive remains exclusive even when a same-user local process
// creates the predictable named semaphore before ECO does.
//
// Windows documents that CreateSemaphore opens an existing named semaphore and
// ignores the caller's requested initial/maximum counts in that case. A hostile
// pre-creator therefore supplies count=2,max=2 here. ECO must still fail closed
// rather than allowing two simultaneous owners for one workspace object.
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

	h, _, createErr := procCreateSemaphoreWWorkspace.Call(
		0,
		2, // hostile pre-created current count
		2, // hostile pre-created maximum count
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
		t.Fatal("P0: hostile pre-created semaphore allowed two simultaneous workspace owners")
	}
	if !errors.Is(err, ErrWorkspaceInUse) {
		t.Fatalf("second acquisition error=%v; want ErrWorkspaceInUse", err)
	}
}
