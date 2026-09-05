package eco

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExplicitCreateThenStrictOpen(t *testing.T) {
	root := filepath.Join(t.TempDir(), "new")
	v, err := CreateVault(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.CreateMatter("explicit creation retained"); err != nil {
		_ = v.Close()
		t.Fatal(err)
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	before := workspaceFormatTree(t, root)
	reopened, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if len(reopened.Snapshot().Matters) != 1 {
		t.Fatal("strict reopen lost explicit state")
	}
	if !reflect.DeepEqual(before, workspaceFormatTree(t, root)) {
		t.Fatal("strict reopen changed committed state")
	}
}
func TestExplicitCreateRefusesExistingRoutes(t *testing.T) {
	for _, name := range []string{"empty", "unrelated", "workspace"} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, "existing")
			if name == "workspace" {
				v, err := CreateVault(root)
				if err != nil {
					t.Fatal(err)
				}
				if err := v.Close(); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Mkdir(root, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if name == "unrelated" {
				if err := os.WriteFile(filepath.Join(root, "keep.txt"), []byte("synthetic unrelated file"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before := workspaceFormatTree(t, parent)
			v, err := CreateVault(root)
			if v != nil {
				_ = v.Close()
			}
			if err == nil {
				t.Error("explicit create adopted existing route")
			}
			if !reflect.DeepEqual(before, workspaceFormatTree(t, parent)) {
				t.Error("refused creation changed existing route")
			}
		})
	}
}
