package eco

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestWorkspaceSnapshotStable(t *testing.T) {
	f := newRollbackFixture(t)
	for _, root := range []string{f.root, f.stageRoot} {
		before := workspaceFormatTree(t, root)
		for i := 0; i < 10; i++ {
			if after := workspaceFormatTree(t, root); !reflect.DeepEqual(before, after) {
				t.Fatalf("snapshot changed during read-only repetition %d", i)
			}
		}
	}
}

// Negative controls: the measurement repair must not hide genuine mutations.
// The content case restores both size and mtime so only the hash can expose it.
func TestWorkspaceSnapshotDetectsChanges(t *testing.T) {
	for _, name := range []string{"content_same_size_and_mtime", "file_mtime", "directory_mtime", "file_mode", "new_entry", "missing_entry"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "directory")
			path := filepath.Join(dir, "synthetic.txt")
			if err := os.Mkdir(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("AAAA"), 0600); err != nil {
				t.Fatal(err)
			}
			stamp := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
			for _, p := range []string{path, dir} {
				if err := os.Chtimes(p, stamp, stamp); err != nil {
					t.Fatal(err)
				}
			}
			before := workspaceFormatTree(t, root)
			entry := filepath.Join("directory", "synthetic.txt")
			var err error
			switch name {
			case "content_same_size_and_mtime":
				if err = os.WriteFile(path, []byte("BBBB"), 0600); err == nil {
					err = os.Chtimes(path, stamp, stamp)
				}
			case "file_mtime":
				err = os.Chtimes(path, stamp, stamp.Add(time.Hour))
			case "directory_mtime":
				entry = "directory"
				err = os.Chtimes(dir, stamp, stamp.Add(time.Hour))
			case "file_mode":
				err = os.Chmod(path, 0400)
				defer os.Chmod(path, 0600)
			case "new_entry":
				entry = filepath.Join("directory", "added.txt")
				err = os.WriteFile(filepath.Join(root, entry), []byte("added"), 0600)
			case "missing_entry":
				err = os.Remove(path)
			}
			if err != nil {
				t.Fatal(err)
			}
			after := workspaceFormatTree(t, root)
			old, hadOld := before[entry]
			now, hasNow := after[entry]
			if hadOld == hasNow && old == now {
				t.Fatal("snapshot did not detect the deliberate mutation to the selected entry")
			}
		})
	}
}
