package eco

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestWorkspaceFormatRejectsUnsafeRootWithoutMutation(t *testing.T) {
	for _, name := range []string{"empty_symlink", "empty_symlink_trailing_separator", "dangling_symlink", "regular_file"} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, "selected-root")
			if name == "regular_file" {
				if err := os.WriteFile(root, []byte("synthetic unrelated file"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				target := filepath.Join(parent, "target")
				if name != "dangling_symlink" {
					if err := os.Mkdir(target, 0700); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Symlink(target, root); err != nil {
					if runtime.GOOS == "windows" {
						t.Skipf("Windows runner cannot create this symlink fixture: %v", err)
					}
					t.Fatal(err)
				}
				if name == "empty_symlink_trailing_separator" {
					root += string(filepath.Separator)
				}
			}
			before := workspaceFormatTree(t, parent)
			v, err := OpenVault(root)
			if v != nil {
				if closeErr := v.Close(); closeErr != nil {
					t.Fatal(closeErr)
				}
			}
			if err == nil {
				t.Error("unsafe selected root was accepted for workspace initialization")
			}
			if after := workspaceFormatTree(t, parent); !reflect.DeepEqual(before, after) {
				t.Error("refusing unsafe root changed its parent, target, link or unrelated file")
			}
		})
	}
}
