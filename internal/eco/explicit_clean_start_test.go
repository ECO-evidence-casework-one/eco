package eco

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestExplicitCleanStartWithRestoreHistory(t *testing.T) {
	f := newRollbackFixture(t)
	before := f.active.Snapshot()
	if err := f.active.Close(); err != nil {
		t.Fatal(err)
	}
	prior := f.root + ".pre-restore-historical"
	if err := os.Mkdir(prior, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prior, "retain.txt"), []byte("synthetic retained history"), 0600); err != nil {
		t.Fatal(err)
	}
	retained := workspaceFormatTree(t, prior)
	archive, err := StartCleanDevelopmentWorkspace(f.root)
	if err != nil {
		t.Fatal(err)
	}
	fresh, err := OpenVault(f.root)
	if err != nil {
		t.Fatal(err)
	}
	got := fresh.Snapshot()
	_ = fresh.Close()
	if len(got.Matters) != 0 || len(got.Evidence) != 0 || len(got.Questions) != 0 {
		t.Fatal("explicit new workspace inherited old state")
	}
	if !reflect.DeepEqual(retained, workspaceFormatTree(t, prior)) {
		t.Fatal("clean start changed earlier restore history")
	}
	old, err := OpenVault(archive)
	if err != nil {
		t.Fatal(err)
	}
	verifyForcedRestoreView(t, old, before)
}

func TestExplicitCleanStartRefusesLiveOwner(t *testing.T) {
	f := newRollbackFixture(t)
	before := workspaceFormatTree(t, filepath.Dir(f.root))
	if _, err := StartCleanDevelopmentWorkspace(f.root); !errors.Is(err, ErrWorkspaceInUse) {
		t.Fatalf("live owner was not respected: %v", err)
	}
	if !reflect.DeepEqual(before, workspaceFormatTree(t, filepath.Dir(f.root))) {
		t.Fatal("refused reset changed existing state")
	}
}

func TestResetForcedStopRestart(t *testing.T) {
	for _, phase := range []string{"reset_archived", "reset_created"} {
		t.Run(phase, func(t *testing.T) {
			f := newRollbackFixture(t)
			original := f.active.Snapshot()
			if err := f.active.Close(); err != nil {
				t.Fatal(err)
			}
			if err := f.stage.Close(); err != nil {
				t.Fatal(err)
			}
			originalTree := workspaceFormatTree(t, f.root)
			marker := filepath.Join(t.TempDir(), "graceful-close.txt")
			killRestoreAtBoundary(t, phase, f.root, "", marker)
			parentBefore := workspaceFormatTree(t, filepath.Dir(f.root))
			fresh, err := OpenVault(f.root)
			if phase == "reset_archived" {
				if fresh != nil {
					_ = fresh.Close()
				}
				if !errors.Is(err, ErrWorkspaceRecoveryRequired) {
					t.Fatalf("interrupted reset silently continued: %v", err)
				}
				if !reflect.DeepEqual(parentBefore, workspaceFormatTree(t, filepath.Dir(f.root))) {
					t.Fatal("reset recovery refusal changed retained state")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				got := fresh.Snapshot()
				if len(got.Matters) != 0 || len(got.Evidence) != 0 || len(got.Questions) != 0 {
					_ = fresh.Close()
					t.Fatal("completed fresh creation inherited old state")
				}
				if err := fresh.Save(); err != nil {
					_ = fresh.Close()
					t.Fatal(err)
				}
				if err := fresh.Close(); err != nil {
					t.Fatal(err)
				}
			}
			entries, err := os.ReadDir(filepath.Dir(f.root))
			if err != nil {
				t.Fatal(err)
			}
			archive := ""
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), filepath.Base(f.root)+".previous-") {
					if archive != "" {
						t.Fatal("ambiguous test reset archive")
					}
					archive = filepath.Join(filepath.Dir(f.root), entry.Name())
				}
			}
			if archive == "" {
				t.Fatal("original reset archive missing")
			}
			if !reflect.DeepEqual(originalTree, workspaceFormatTree(t, archive)) {
				t.Fatal("forced reset changed original archived bytes")
			}
			old, err := OpenVault(archive)
			if err != nil {
				t.Fatal(err)
			}
			verifyForcedRestoreView(t, old, original)
		})
	}
}

func TestResetForcedStopChild(t *testing.T) {
	if os.Getenv("ECO_TEST_FORCED_RESTORE") != "1" {
		return
	}
	defer func() { _ = os.WriteFile(os.Getenv("ECO_TEST_RESTORE_CLOSED"), []byte("graceful cleanup ran"), 0600) }()
	boundary := func(phase string) {
		if phase != os.Getenv("ECO_TEST_RESTORE_PHASE") {
			return
		}
		if err := json.NewEncoder(os.Stdout).Encode(struct{ Phase string }{phase}); err != nil {
			t.Fatal(err)
		}
		var one [1]byte
		_, _ = io.ReadFull(os.Stdin, one[:])
		t.Fatal("parent released reset boundary instead of killing child")
	}
	if _, err := startCleanDevelopmentWorkspace(os.Getenv("ECO_TEST_RESTORE_ROOT"), boundary); err != nil {
		t.Fatal(err)
	}
	t.Fatal("reset completed without reaching requested boundary")
}
