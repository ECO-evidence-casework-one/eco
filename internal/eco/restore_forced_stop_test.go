package eco

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const forcedRestorePassphrase = "synthetic forced restore passphrase"

// The parent kills a real child executing RestorePortableBackup. No test
// manually recreates the directory state, calls rollback, or gracefully closes
// the child Vault at the stop boundary. The observation callback is per-instance
// and unexported; the normal application leaves it nil.
func TestRestoreForcedStopRestart(t *testing.T) {
	for _, phase := range []string{"stage_verified", "original_renamed", "original_retargeted", "staged_renamed", "activated"} {
		t.Run(phase, func(t *testing.T) {
			f := newRollbackFixture(t)
			original := f.active.Snapshot()
			replacement := f.stage.Snapshot()
			backup := filepath.Join(t.TempDir(), "synthetic.ecobackup")
			if _, err := f.stage.CreatePortableBackup(backup, forcedRestorePassphrase, nil); err != nil {
				t.Fatal(err)
			}
			if err := f.stage.Close(); err != nil {
				t.Fatal(err)
			}
			if err := f.active.Close(); err != nil {
				t.Fatal(err)
			}
			originalTree := workspaceFormatTree(t, f.root)
			closedMarker := filepath.Join(t.TempDir(), "graceful-close.txt")
			killRestoreAtBoundary(t, phase, f.root, backup, closedMarker)

			// The default startup path must not turn the missing-root interval
			// into fresh empty state. Its refusal must not mutate any checkpoint.
			gap := phase == "original_renamed" || phase == "original_retargeted"
			beforeReopen := workspaceFormatTree(t, filepath.Dir(f.root))
			v, err := OpenVault(f.root)
			if gap {
				if v != nil {
					_ = v.Close()
				}
				if err == nil {
					t.Error("restart silently created fresh state during interrupted restore")
				}
				if !reflect.DeepEqual(beforeReopen, workspaceFormatTree(t, filepath.Dir(f.root))) {
					t.Error("refusing interrupted-restore startup changed the retained workspace tree")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				want := replacement
				if phase == "stage_verified" {
					want = original
				}
				verifyForcedRestoreView(t, v, want)
				if err := v.Close(); err != nil {
					t.Fatal(err)
				}
			}

			// Whenever activation began, the real pre-restore checkpoint must
			// still retain the original bytes and be explicitly reopenable.
			if phase != "stage_verified" {
				entries, err := os.ReadDir(filepath.Dir(f.root))
				if err != nil {
					t.Fatal(err)
				}
				checkpoint := ""
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), filepath.Base(f.root)+".pre-restore-") {
						if checkpoint != "" {
							t.Fatal("ambiguous test checkpoint")
						}
						checkpoint = filepath.Join(filepath.Dir(f.root), entry.Name())
					}
				}
				if checkpoint == "" {
					t.Fatal("forced restore lost the original checkpoint")
				}
				if !reflect.DeepEqual(originalTree, workspaceFormatTree(t, checkpoint)) {
					t.Error("original checkpoint changed across forced termination")
				}
				old, err := OpenVault(checkpoint)
				if err != nil {
					t.Fatal(err)
				}
				verifyForcedRestoreView(t, old, original)
				if err := old.Close(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func verifyForcedRestoreView(t *testing.T, v *Vault, want Workspace) {
	t.Helper()
	defer v.Close() // Also release ownership if an assertion stops this subtest.
	got := v.Snapshot()
	if len(got.Matters) != 1 || len(want.Matters) != 1 || got.Matters[0].Title != want.Matters[0].Title {
		t.Fatal("restart selected fresh state or the wrong matter")
	}
	if len(got.Evidence) != 1 || got.Evidence[0].SHA256 != want.Evidence[0].SHA256 {
		t.Fatal("restart lost or substituted the preserved source")
	}
	if err := verifyStagedVault(v); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatalf("ownership/CAS did not recover after process exit: %v", err)
	}
}

func killRestoreAtBoundary(t *testing.T, phase, root, backup, marker string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRestoreForcedStopChild$")
	cmd.Env = append(os.Environ(), "ECO_TEST_FORCED_RESTORE=1", "ECO_TEST_RESTORE_PHASE="+phase,
		"ECO_TEST_RESTORE_ROOT="+root, "ECO_TEST_RESTORE_BACKUP="+backup, "ECO_TEST_RESTORE_CLOSED="+marker)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	defer func() {
		if !waited {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	type reply struct{ Phase string }
	var reached reply
	decoded := make(chan error, 1)
	go func() { decoded <- json.NewDecoder(stdout).Decode(&reached) }()
	select {
	case err := <-decoded:
		if err != nil {
			t.Fatalf("child did not reach %s: %v", phase, err)
		}
	case <-ctx.Done():
		t.Fatalf("child timed out before controlled stop at %s", phase)
	}
	if reached.Phase != phase {
		t.Fatalf("wrong child boundary: %q", reached.Phase)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	err = cmd.Wait() // Kill is asynchronous; wait before trying to reopen.
	waited = true
	var exited *exec.ExitError
	if !errors.As(err, &exited) || ctx.Err() != nil {
		t.Fatalf("child was not terminated at the acknowledged boundary: %v; %s", err, stderr.String())
	}
	if _, err := os.Lstat(marker); !os.IsNotExist(err) {
		t.Fatal("child ran graceful cleanup; this was not a forced-stop test")
	}
}

func TestRestoreForcedStopChild(t *testing.T) {
	if os.Getenv("ECO_TEST_FORCED_RESTORE") != "1" {
		return
	}
	v, err := OpenVault(os.Getenv("ECO_TEST_RESTORE_ROOT"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = v.Close()
		_ = os.WriteFile(os.Getenv("ECO_TEST_RESTORE_CLOSED"), []byte("graceful cleanup ran"), 0600)
	}()
	v.restoreBoundary = func(phase string) {
		if phase != os.Getenv("ECO_TEST_RESTORE_PHASE") {
			return
		}
		if err := json.NewEncoder(os.Stdout).Encode(struct{ Phase string }{phase}); err != nil {
			t.Fatal(err)
		}
		var one [1]byte
		_, _ = io.ReadFull(os.Stdin, one[:])
		t.Fatal("parent released boundary instead of terminating child")
	}
	if _, err := v.RestorePortableBackup(os.Getenv("ECO_TEST_RESTORE_BACKUP"), forcedRestorePassphrase, nil); err != nil {
		t.Fatal(err)
	}
	t.Fatal("restore completed without reaching requested stop boundary")
}
