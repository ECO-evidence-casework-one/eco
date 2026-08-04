//go:build linux

package eco

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

type linuxWorkspaceLock struct {
	file *os.File
}

func platformAcquireWorkspaceLock(path string) (platformWorkspaceLock, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open workspace owner lock: %w", err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, ErrWorkspaceInUse
		}
		return nil, fmt.Errorf("lock workspace owner file: %w", err)
	}
	return &linuxWorkspaceLock{file: file}, nil
}

func (l *linuxWorkspaceLock) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	unlockErr := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	closeErr := l.file.Close()
	l.file = nil
	if unlockErr != nil {
		return fmt.Errorf("unlock workspace owner file: %w", unlockErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close workspace owner file: %w", closeErr)
	}
	return nil
}

func platformWorkspaceObjectIdentity(path string) (workspaceObjectIdentity, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(path, &stat); err != nil {
		return workspaceObjectIdentity{}, err
	}
	return workspaceObjectIdentity{Volume: uint64(stat.Dev), File: stat.Ino}, nil
}
