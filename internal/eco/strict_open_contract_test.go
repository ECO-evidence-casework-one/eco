package eco

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStrictOpenNeverCreates(t *testing.T) {
	for _, name := range []string{"missing", "empty"} {
		t.Run(name, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, "selected")
			if name == "empty" {
				if err := os.Mkdir(root, 0700); err != nil {
					t.Fatal(err)
				}
			}
			before := workspaceFormatTree(t, parent)
			v, err := OpenVault(root)
			if v != nil {
				_ = v.Close()
			}
			if err == nil {
				t.Error("ordinary open initialized missing committed state")
			}
			if !reflect.DeepEqual(before, workspaceFormatTree(t, parent)) {
				t.Error("ordinary open mutated a missing/empty selection")
			}
		})
	}
}
