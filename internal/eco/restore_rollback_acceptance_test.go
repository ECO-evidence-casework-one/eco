package eco

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// These tests recreate exact directory-rename boundaries and exercise the
// production rollback helper. They are not process-kill or power-loss tests.
type rollbackFixture struct {
	active, stage               *Vault
	root, stageRoot, checkpoint string
}

func newRollbackFixture(t *testing.T) rollbackFixture {
	t.Helper()
	parent := t.TempDir()
	f := rollbackFixture{
		root:       filepath.Join(parent, "active"),
		stageRoot:  filepath.Join(parent, "active.restore-synthetic"),
		checkpoint: filepath.Join(parent, "active.pre-restore-synthetic"),
	}
	for i, root := range []string{f.root, f.stageRoot} {
		v, err := openTestVault(root)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = v.Close() })
		label := []string{"Original synthetic matter", "Staged synthetic matter"}[i]
		if _, err := v.CreateMatter(label); err != nil {
			t.Fatal(err)
		}
		source := filepath.Join(t.TempDir(), "synthetic.txt")
		if err := os.WriteFile(source, []byte(label+"\nSynthetic preserved bytes only.\n"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := v.ImportFile(source, nil); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			f.active = v
		} else {
			f.stage = v
		}
	}
	return f
}

func (f rollbackFixture) advance(t *testing.T, phase int) {
	t.Helper()
	if err := os.Rename(f.root, f.checkpoint); err != nil {
		t.Fatal(err)
	}
	if phase >= 2 {
		if err := f.active.owner.retarget(f.checkpoint); err != nil {
			t.Fatal(err)
		}
	}
	if phase >= 3 {
		if err := os.Rename(f.stageRoot, f.root); err != nil {
			t.Fatal(err)
		}
	}
	if phase >= 4 {
		if err := f.stage.owner.retarget(f.root); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRecoveryRollbackRenameBoundaries(t *testing.T) {
	for phase := 1; phase <= 4; phase++ {
		t.Run(fmt.Sprintf("boundary_%d", phase), func(t *testing.T) {
			f := newRollbackFixture(t)
			original := workspaceFormatTree(t, f.root)
			staged := workspaceFormatTree(t, f.stageRoot)
			f.advance(t, phase)
			if err := rollbackRestoreActivation(f.root, f.stageRoot, f.checkpoint, f.active.owner, f.stage.owner, phase >= 3); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(original, workspaceFormatTree(t, f.root)) {
				t.Error("original workspace bytes or entries changed")
			}
			if !reflect.DeepEqual(staged, workspaceFormatTree(t, f.stageRoot)) {
				t.Error("staged workspace bytes or entries changed")
			}
			for _, v := range []*Vault{f.active, f.stage} {
				if err := v.owner.revalidate(); err != nil {
					t.Fatal(err)
				}
				if err := verifyStagedVault(v); err != nil {
					t.Fatal(err)
				}
				if err := v.Save(); err != nil {
					t.Fatalf("owner/CAS unusable after rollback: %v", err)
				}
				root := v.Root
				want := v.Snapshot().Matters[0].Title
				if err := v.Close(); err != nil {
					t.Fatal(err)
				}
				reopened, err := openTestVault(root)
				if err != nil {
					t.Fatal(err)
				}
				if reopened.Snapshot().Matters[0].Title != want {
					t.Error("reopen selected the wrong workspace")
				}
				if err := verifyStagedVault(reopened); err != nil {
					_ = reopened.Close()
					t.Fatal(err)
				}
				if err := reopened.Close(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestRecoveryRollbackRefusesUnsafeState(t *testing.T) {
	for _, scenario := range []string{
		"missing_checkpoint", "substituted_checkpoint", "occupied_active_empty",
		"occupied_active_nonempty", "occupied_stage_empty", "occupied_stage_nonempty",
		"substituted_stage_source",
	} {
		t.Run(scenario, func(t *testing.T) {
			f := newRollbackFixture(t)
			stageMoved := strings.HasPrefix(scenario, "occupied_stage") || scenario == "substituted_stage_source"
			phase := 2
			if stageMoved {
				phase = 4
			}
			f.advance(t, phase)
			switch scenario {
			case "missing_checkpoint", "substituted_checkpoint":
				if err := os.Rename(f.checkpoint, f.checkpoint+".displaced"); err != nil {
					t.Fatal(err)
				}
				if scenario == "substituted_checkpoint" {
					if err := os.Mkdir(f.checkpoint, 0700); err != nil {
						t.Fatal(err)
					}
				}
			case "occupied_active_empty", "occupied_active_nonempty":
				if err := os.Mkdir(f.root, 0700); err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(scenario, "nonempty") {
					if err := os.WriteFile(filepath.Join(f.root, "unrelated.txt"), []byte("retain unrelated synthetic file"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			case "occupied_stage_empty", "occupied_stage_nonempty":
				if err := os.Mkdir(f.stageRoot, 0700); err != nil {
					t.Fatal(err)
				}
				if strings.HasSuffix(scenario, "nonempty") {
					if err := os.WriteFile(filepath.Join(f.stageRoot, "unrelated.txt"), []byte("retain unrelated synthetic file"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			case "substituted_stage_source":
				if err := os.Rename(f.root, f.root+".displaced-stage"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(f.root, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(f.root, "unrelated.txt"), []byte("not the staged workspace"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before := workspaceFormatTree(t, filepath.Dir(f.root))
			err := rollbackRestoreActivation(f.root, f.stageRoot, f.checkpoint, f.active.owner, f.stage.owner, stageMoved)
			if err == nil {
				t.Error("unsafe or missing recovery state was reported as successfully rolled back")
			}
			if !reflect.DeepEqual(before, workspaceFormatTree(t, filepath.Dir(f.root))) {
				t.Error("refusal moved/replaced unrelated state or recovery checkpoints")
			}
		})
	}
}

func TestRecoveryMetadataWriteInterruption(t *testing.T) {
	for _, scenario := range []string{"temporary_write_refused", "uncommitted_temporary_metadata"} {
		t.Run(scenario, func(t *testing.T) {
			root, key, document := workspaceFormatFixture(t)
			v, err := openTestVault(root)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = v.Close() })
			before := v.Snapshot()
			path := filepath.Join(root, "workspace.ecodb")
			committed, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			tmp := path + ".tmp"
			if scenario == "temporary_write_refused" {
				if err := os.Mkdir(tmp, 0700); err != nil {
					t.Fatal(err)
				}
				if err := v.Save(); err == nil {
					t.Fatal("save into a directory unexpectedly succeeded")
				}
				if !reflect.DeepEqual(before, v.Snapshot()) {
					t.Error("failed save changed revision/transaction or workspace state")
				}
				if err := os.Remove(tmp); err != nil {
					t.Fatal(err)
				}
			} else {
				document["selected_page"] = json.RawMessage(`"UNCOMMITTED-SYNTHETIC-STATE"`)
				plain := workspaceFormatJSON(t, document)
				enc, err := encryptBlob(key, metaMagic, plain)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(tmp, enc, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := v.Close(); err != nil {
				t.Fatal(err)
			}
			reopened, err := openTestVault(root)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.Close()
			if !reflect.DeepEqual(before, reopened.Snapshot()) {
				t.Error("reopen did not retain the last committed workspace")
			}
			actual, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(committed, actual) {
				t.Error("last committed metadata bytes were replaced")
			}
		})
	}
}

func TestRecoveryRollbackFailureRetainsCause(t *testing.T) {
	cause := errors.New("synthetic restore interruption")
	rollback := errors.New("synthetic rollback obstruction")
	got := restoreActivationFailure(cause, rollback)
	if !errors.Is(got, cause) || !strings.Contains(got.Error(), rollback.Error()) {
		t.Fatal("rollback failure hid the initiating failure or recovery failure")
	}
}
