//go:build !windows && (!linux || !amd64)

package eco

import "errors"

var errObjectBoundFilesystemUnavailable = errors.New("this platform does not provide the object-bound filesystem operations required for destructive workspace changes")

func openBoundObjectDirectory(string, string) (boundObjectDirectory, error) {
	return nil, errObjectBoundFilesystemUnavailable
}

func objectBoundRemoveFile(string, string, func([]byte) error, func(string) error) error {
	return errObjectBoundFilesystemUnavailable
}

func objectBoundRemoveTree(string, func() error, func(string) error) error {
	return errObjectBoundFilesystemUnavailable
}

func objectBoundRename(string, string, func() error, func(string, string) error) error {
	return errObjectBoundFilesystemUnavailable
}
