//go:build linux

package eco

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceOwnerAliasSharesLockDomain(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "workspace")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(parent, "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	lease, err := acquireWorkspaceRootOwner(root)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	other, err := acquireWorkspaceRootOwner(alias)
	if other != nil {
		_ = other.Close()
	}
	if !errors.Is(err, ErrWorkspaceInUse) {
		t.Fatalf("alias acquisition error=%v", err)
	}
}
