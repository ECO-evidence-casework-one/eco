package eco

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRecoveryStageCleanupOwnership(t *testing.T) {
	for _, scenario := range []string{"owned_stage", "displaced_stage", "substituted_stage"} {
		t.Run(scenario, func(t *testing.T) {
			f := newRollbackFixture(t)
			original := workspaceFormatTree(t, f.root)
			if scenario != "owned_stage" {
				if err := os.Rename(f.stageRoot, f.stageRoot+".displaced"); err != nil {
					t.Fatal(err)
				}
				if scenario == "substituted_stage" {
					if err := os.Mkdir(f.stageRoot, 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(f.stageRoot, "unrelated.txt"), []byte("retain unrelated synthetic bytes"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			before := workspaceFormatTree(t, filepath.Dir(f.root))
			err := removeUnactivatedRestoreStage(f.stage, f.stageRoot)
			if scenario == "owned_stage" {
				if err != nil {
					t.Fatal(err)
				}
				if _, err := os.Lstat(f.stageRoot); !os.IsNotExist(err) {
					t.Fatal("owned disposable stage was not removed")
				}
			} else {
				if err == nil {
					t.Error("unowned/missing staged pathname accepted for removal")
				}
				if !reflect.DeepEqual(before, workspaceFormatTree(t, filepath.Dir(f.root))) {
					t.Error("cleanup changed a displaced stage or unrelated replacement")
				}
			}
			if !reflect.DeepEqual(original, workspaceFormatTree(t, f.root)) {
				t.Error("stage cleanup changed the original workspace")
			}
			if err := f.stage.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
