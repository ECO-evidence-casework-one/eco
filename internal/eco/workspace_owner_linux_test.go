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

func TestWorkspaceCreationParentAliasSharesClaimDomain(t *testing.T) {
	base := t.TempDir()
	parent := filepath.Join(base, "parent")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "parent-alias")
	if err := os.Symlink(parent, alias); err != nil {
		t.Fatal(err)
	}
	first, err := acquireWorkspaceCreationOwner(filepath.Join(parent, "workspace"))
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := acquireWorkspaceCreationOwner(filepath.Join(alias, "workspace"))
	if second != nil {
		_ = second.Close()
	}
	if !errors.Is(err, ErrWorkspaceInUse) {
		t.Fatalf("aliased parent acquisition error=%v", err)
	}
}

func TestWorkspaceCreationDetectsParentSubstitution(t *testing.T) {
	base := t.TempDir()
	parent := filepath.Join(base, "parent")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	lease, err := acquireWorkspaceCreationOwner(filepath.Join(parent, "workspace"))
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	moved := filepath.Join(base, "moved-parent")
	if err := os.Rename(parent, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(parent, "sentinel.txt")
	if err := os.WriteFile(sentinel, []byte("unrelated"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := lease.revalidate(); err == nil {
		t.Fatal("parent substitution was not detected")
	}
	data, err := os.ReadFile(sentinel)
	if err != nil || string(data) != "unrelated" {
		t.Fatalf("replacement parent was altered: %q err=%v", data, err)
	}
}

func TestWorkspaceOwnerDetectsRootSubstitution(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "workspace")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	lease, err := acquireWorkspaceRootOwner(root)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	moved := filepath.Join(base, "moved-workspace")
	if err := os.Rename(root, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(root, "sentinel.txt")
	if err := os.WriteFile(sentinel, []byte("replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := lease.revalidate(); err == nil {
		t.Fatal("root substitution was not detected")
	}
	data, err := os.ReadFile(sentinel)
	if err != nil || string(data) != "replacement" {
		t.Fatalf("replacement root was altered: %q err=%v", data, err)
	}
}
