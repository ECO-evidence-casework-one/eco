package eco

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRestoreRestartGuard(t *testing.T) {
	for _, name := range []string{"missing_active", "empty_active", "checkpoint_is_file", "literal_brackets", "large_parent"} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			base := "candidate"
			if name == "literal_brackets" {
				base = "candidate[1]"
			}
			root := filepath.Join(parent, base)
			checkpoint := root + ".pre-restore-synthetic"
			if name == "checkpoint_is_file" {
				if err := os.WriteFile(checkpoint, []byte("not automatically trusted"), 0600); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Mkdir(checkpoint, 0700); err != nil {
				t.Fatal(err)
			}
			if name == "empty_active" {
				if err := os.Mkdir(root, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if name == "large_parent" {
				for i := 0; i < 260; i++ {
					if err := os.WriteFile(filepath.Join(parent, fmt.Sprintf("unrelated-%03d", i)), []byte("x"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			before := workspaceFormatTree(t, parent)
			if err := CheckWorkspaceRecoveryState(root); !errors.Is(err, ErrWorkspaceRecoveryRequired) {
				t.Fatalf("recovery check did not require an explicit choice: %v", err)
			}
			v, err := OpenVault(root)
			if v != nil {
				_ = v.Close()
			}
			if !errors.Is(err, ErrWorkspaceRecoveryRequired) {
				t.Fatalf("open bypassed recovery guard: %v", err)
			}
			if !reflect.DeepEqual(before, workspaceFormatTree(t, parent)) {
				t.Fatal("recovery refusal mutated parent/checkpoint/active state")
			}
		})
	}
}

func TestRestoreRestartPositiveControls(t *testing.T) {
	for _, name := range []string{"new_hierarchy", "unrelated_checkpoint", "healthy_existing"} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, "candidate")
			if name == "new_hierarchy" {
				root = filepath.Join(parent, "new-parent", "candidate")
			}
			if name == "unrelated_checkpoint" {
				if err := os.Mkdir(filepath.Join(parent, "other.pre-restore-synthetic"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			if name == "healthy_existing" {
				v, err := CreateVault(root)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := v.CreateMatter("retained healthy state"); err != nil {
					_ = v.Close()
					t.Fatal(err)
				}
				if err := v.Close(); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(root+".pre-restore-synthetic", 0700); err != nil {
					t.Fatal(err)
				}
			}
			if err := CheckWorkspaceRecoveryState(root); err != nil {
				t.Fatal(err)
			}
			var v *Vault
			var err error
			if name == "healthy_existing" {
				v, err = OpenVault(root)
			} else {
				v, err = CreateVault(root)
			}
			if err != nil {
				t.Fatal(err)
			}
			defer v.Close()
			if name == "healthy_existing" && len(v.Snapshot().Matters) != 1 {
				t.Fatal("existing state was replaced")
			}
		})
	}
}
