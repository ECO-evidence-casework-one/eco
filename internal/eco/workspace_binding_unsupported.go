//go:build !windows && (!linux || !amd64)

package eco

func acquireRawWorkspaceLifecycleLock(string, string) (rawWorkspaceLifecycleLock, error) {
	return nil, errObjectBoundFilesystemUnavailable
}

func openBoundWorkspaceObjects(string) (boundWorkspaceObjects, error) {
	return nil, errObjectBoundFilesystemUnavailable
}
