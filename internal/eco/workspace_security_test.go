package eco

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAutomaticRecoveryRejectsDifferentCandidateWithSameBuild(t *testing.T) {
	base := t.TempDir()
	build := "ECO-SAME-PUBLIC-BUILD"
	current := runtimeFor("candidate-current-exact", build, Schema)
	other := runtimeFor("candidate-other-exact", build, Schema)
	stateRoot := filepath.Join(base, "candidates", candidateDirectoryName(current.CandidateID))
	defaultWorkspace := filepath.Join(stateRoot, "workspaces", "development")
	foreign, preservation := createRecoverablePreservation(t, defaultWorkspace, other, []byte("Synthetic pending records owned by another exact candidate."))
	before := foreign.Snapshot()
	foreign.Close()

	if _, err := StartCandidate(base, current); err == nil || !strings.Contains(strings.ToLower(err.Error()), "another development candidate") {
		t.Fatalf("automatic startup did not reject a different exact candidate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(stateRoot, "app-state.json")); !os.IsNotExist(err) {
		t.Fatalf("blocked automatic recovery created candidate state: %v", err)
	}
	unchanged, err := inspectWorkspace(defaultWorkspace, other)
	if err != nil {
		t.Fatal(err)
	}
	after := unchanged.Workspace
	zeroBytes(unchanged.key)
	if len(after.Evidence) != 0 || len(after.Preservations) != 1 || after.Preservations[0].ID != preservation.ID || after.Preservations[0].State != preservationRecoverable || len(after.Changes) != len(before.Changes) {
		t.Fatalf("blocked automatic recovery changed or presented the foreign workspace: before=%+v after=%+v", before, after)
	}

	deliberateApp, err := StartCandidate(t.TempDir(), current)
	if err != nil {
		t.Fatal(err)
	}
	session, err := deliberateApp.OpenWorkspace(defaultWorkspace)
	if err != nil {
		t.Fatal(err)
	}
	if !session.Explicit || session.Compatibility.Status != CompatibilityCompatibleBuild || session.Identity.CreatedByCandidate != other.CandidateID || len(session.Vault.Snapshot().Evidence) != 1 {
		t.Fatalf("foreign workspace was not clearly attributed as an explicit external reopen: %+v", session)
	}
}

func TestWorkspaceCandidateIdentityFailsClosedWhenAbsentDamagedOrInconsistent(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, *Vault)
	}{
		{
			name: "routing identity absent",
			mutate: func(t *testing.T, vault *Vault) {
				identity := vault.Identity
				identity.CreatedByCandidate = ""
				if err := writeWorkspaceIdentity(vault.Root, identity); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "routing identity inconsistent",
			mutate: func(t *testing.T, vault *Vault) {
				identity := vault.Identity
				identity.CreatedByCandidate = "candidate-inconsistent"
				if err := writeWorkspaceIdentity(vault.Root, identity); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "routing identity damaged",
			mutate: func(t *testing.T, vault *Vault) {
				if err := os.WriteFile(filepath.Join(vault.Root, workspaceIdentityFile), []byte("{damaged"), 0600); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "authenticated identity absent",
			mutate: func(t *testing.T, vault *Vault) {
				vault.Workspace.CreatedByCandidate = ""
				if err := vault.Save(); err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runtime := runtimeFor("candidate-exact-binding", "ECO-EXACT-BINDING", Schema)
			root := filepath.Join(t.TempDir(), "workspace")
			vault, err := createVault(root, "Exact candidate binding", runtime)
			if err != nil {
				t.Fatal(err)
			}
			test.mutate(t, vault)
			vault.Close()

			if _, err = openVault(root, runtime); err == nil {
				t.Fatal("workspace opened without consistent exact candidate identity")
			}
		})
	}
}

func TestPlaintextRoutingFilesContainOnlyControlledIdentity(t *testing.T) {
	base := t.TempDir()
	runtime := runtimeFor("candidate-routing-minimum", "ECO-ROUTING-MINIMUM", Schema)
	app, err := StartCandidate(base, runtime)
	if err != nil {
		t.Fatal(err)
	}
	identityBytes, err := os.ReadFile(filepath.Join(app.Current.Path, workspaceIdentityFile))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{app.Current.Identity.Name, "created_at", "created_by_build", "workspace_name"} {
		if bytes.Contains(identityBytes, []byte(forbidden)) {
			t.Fatalf("plaintext workspace identity exposed %q: %s", forbidden, identityBytes)
		}
	}
	for _, required := range []string{app.Current.Identity.ID, runtime.CandidateID, `"schema"`, `"kind"`} {
		if !bytes.Contains(identityBytes, []byte(required)) {
			t.Fatalf("plaintext workspace identity omitted required routing value %q: %s", required, identityBytes)
		}
	}

	stateBytes, err := os.ReadFile(filepath.Join(app.State.StateRoot, "app-state.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{app.Current.Path, app.State.StateRoot, `"state_root"`, `"default_workspace"`, `"selected_workspace"`, `"path"`} {
		if bytes.Contains(stateBytes, []byte(forbidden)) {
			t.Fatalf("candidate app-state exposed path routing value %q: %s", forbidden, stateBytes)
		}
	}

	blockedPath := filepath.Join(base, "private", "selected", "workspace")
	if _, err = app.OpenWorkspace(blockedPath); err == nil {
		t.Fatal("opening a missing workspace unexpectedly succeeded")
	}
	stateBytes, err = os.ReadFile(filepath.Join(app.State.StateRoot, "app-state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(stateBytes, []byte(blockedPath)) {
		t.Fatalf("blocked candidate audit exposed the selected full path: %s", stateBytes)
	}
}

func TestWorkspaceOpenAndResetRejectObjectsSymbolicLink(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	runtime := runtimeFor("candidate-object-link", "ECO-OBJECT-LINK", Schema)
	vault, err := createVault(root, "Object link workspace", runtime)
	if err != nil {
		t.Fatal(err)
	}
	item := importSynthetic(t, vault, t.TempDir(), "link.txt", "Synthetic object-link protection data.")
	vault.Close()

	external := t.TempDir()
	probe := filepath.Join(t.TempDir(), "symlink-probe")
	if err = os.Symlink(external, probe); err != nil {
		t.Skipf("symbolic links are unavailable on this platform: %v", err)
	}
	if err = os.Remove(probe); err != nil {
		t.Fatal(err)
	}
	externalObject := filepath.Join(external, item.ObjectFile)
	externalBytes := []byte("unrelated external content")
	if err = os.WriteFile(externalObject, externalBytes, 0600); err != nil {
		t.Fatal(err)
	}
	realObjects := filepath.Join(root, "objects-real")
	if err = os.Rename(filepath.Join(root, "objects"), realObjects); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(external, filepath.Join(root, "objects")); err != nil {
		t.Skipf("symbolic links are unavailable on this platform: %v", err)
	}
	if _, err = openVault(root, runtime); err == nil {
		t.Fatal("workspace opened through an objects symbolic link")
	}
	got, err := os.ReadFile(externalObject)
	if err != nil || !bytes.Equal(got, externalBytes) {
		t.Fatalf("workspace open changed an external file: bytes=%q err=%v", got, err)
	}
}

func TestResetBindsObjectsDirectoryThroughFinalManagedRemovalSeam(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the Windows handle-and-junction variant or fail-closed platform path covers this seam")
	}
	root := filepath.Join(t.TempDir(), "workspace")
	runtime := runtimeFor("candidate-object-race", "ECO-OBJECT-RACE", Schema)
	vault, err := createVault(root, "Object race workspace", runtime)
	if err != nil {
		t.Fatal(err)
	}
	item := importSynthetic(t, vault, t.TempDir(), "race.txt", "Synthetic reset race protection data.")
	external := t.TempDir()
	probe := filepath.Join(t.TempDir(), "symlink-race-probe")
	if err = os.Symlink(external, probe); err != nil {
		t.Skipf("symbolic links are unavailable on this platform: %v", err)
	}
	if err = os.Remove(probe); err != nil {
		t.Fatal(err)
	}
	externalObject := filepath.Join(external, item.ObjectFile)
	externalBytes := []byte("unrelated file must survive")
	if err = os.WriteFile(externalObject, externalBytes, 0600); err != nil {
		t.Fatal(err)
	}
	realObjects := filepath.Join(root, "objects-before-race")
	receipt, resetErr := resetVaultWithHook(vault, func(phase ResetPhase) error {
		if phase == resetBeforeObjectCleanup {
			return nil
		}
		if phase != resetBeforeManagedObjectRemoval {
			return errors.New("unexpected reset test phase")
		}
		if renameErr := os.Rename(filepath.Join(root, "objects"), realObjects); renameErr != nil {
			return renameErr
		}
		if linkErr := os.Symlink(external, filepath.Join(root, "objects")); linkErr != nil {
			return linkErr
		}
		return nil
	})
	if resetErr == nil {
		t.Fatal("reset cleanup followed a substituted objects path")
	}
	if receipt.ObjectsRemoved != 0 {
		t.Fatalf("reset deleted objects after containment changed: %+v", receipt)
	}
	audit := vault.Snapshot()
	if !hasChange(audit, "workspace-reset-cleanup-blocked") || hasChange(audit, "workspace-reset-complete") {
		t.Fatalf("a partial reset was not recorded truthfully: %+v", audit.Changes)
	}
	got, err := os.ReadFile(externalObject)
	if err != nil || !bytes.Equal(got, externalBytes) {
		t.Fatalf("reset changed an unrelated external file: bytes=%q err=%v", got, err)
	}
	if _, err = os.Stat(filepath.Join(realObjects, item.ObjectFile)); err != nil {
		t.Fatalf("reset lost the original managed object after substitution: %v", err)
	}
}
