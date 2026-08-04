//go:build !linux && !windows

package eco

import "errors"

func platformAcquireWorkspaceLock(string) (platformWorkspaceLock, error) {
	return nil, errors.New("workspace ownership is unsupported on this platform")
}

func platformWorkspaceObjectIdentity(string) (workspaceObjectIdentity, error) {
	return workspaceObjectIdentity{}, errors.New("workspace object identity is unsupported on this platform")
}
