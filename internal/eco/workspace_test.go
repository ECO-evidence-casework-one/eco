package eco

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runtimeFor(candidate, build string, schema int) RuntimeIdentity {
	return RuntimeIdentity{CandidateID: candidate, BuildID: build, Schema: schema}
}

func importSynthetic(t *testing.T, vault *Vault, folder, name, wording string) EvidenceItem {
	t.Helper()
	path := filepath.Join(folder, name)
	if err := os.WriteFile(path, []byte(wording), 0600); err != nil {
		t.Fatal(err)
	}
	item, duplicate, err := vault.ImportFile(path, nil)
	if err != nil || duplicate {
		t.Fatalf("import synthetic evidence: duplicate=%v err=%v", duplicate, err)
	}
	return item
}

func hasChange(workspace Workspace, changeType string) bool {
	for _, change := range workspace.Changes {
		if change.Type == changeType {
			return true
		}
	}
	return false
}

func TestCandidateIdentityIncludesSourceRevision(t *testing.T) {
	first := candidateIDForSource("ECO-TEST", "aaaaaaaa", false)
	second := candidateIDForSource("ECO-TEST", "bbbbbbbb", false)
	modified := candidateIDForSource("ECO-TEST", "aaaaaaaa", true)
	if first == second || first == modified || !strings.Contains(modified, "modified") {
		t.Fatalf("source-distinct candidates did not receive distinct identities: %q %q %q", first, second, modified)
	}
	if unrecorded := candidateIDForSource("ECO-TEST", "", false); !strings.Contains(unrecorded, "source-unrecorded") {
		t.Fatalf("unrecorded source identity is ambiguous: %q", unrecorded)
	}
}

func createPriorSchemaOneWorkspace(t *testing.T, root string, runtime RuntimeIdentity) *Vault {
	t.Helper()
	vault, err := createVault(root, "Temporary legacy fixture", runtime)
	if err != nil {
		t.Fatal(err)
	}
	vault.mu.Lock()
	vault.Workspace.WorkspaceID = ""
	vault.Workspace.WorkspaceName = ""
	vault.Workspace.CreatedByBuild = ""
	vault.Workspace.Changes = []ChangeRecord{}
	err = vault.saveUnlocked()
	vault.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(filepath.Join(root, workspaceIdentityFile)); err != nil {
		t.Fatal(err)
	}
	vault.Identity = WorkspaceIdentity{}
	return vault
}

func TestFirstLaunchOnCleanMachineCreatesGenuinelyEmptyWorkspace(t *testing.T) {
	base := t.TempDir()
	runtime := runtimeFor("candidate-clean", "ECO-CANDIDATE-CLEAN", Schema)

	app, err := StartCandidate(base, runtime)
	if err != nil {
		t.Fatal(err)
	}
	session := app.Current
	workspace := session.Vault.Snapshot()
	if session.Disposition != DispositionNew || session.Explicit {
		t.Fatalf("unexpected first-launch status: %+v", session)
	}
	if len(workspace.Evidence) != 0 || len(workspace.Preservations) != 0 || len(workspace.Matters) != 0 || len(workspace.Questions) != 0 {
		t.Fatalf("clean launch inherited records: %+v", workspace)
	}
	if workspace.Settings != (Settings{}) || workspace.WorkspaceID == "" || workspace.WorkspaceID != session.Identity.ID {
		t.Fatalf("new workspace identity or settings are not clean: %+v", workspace)
	}
	if !hasChange(workspace, "workspace-created") || len(workspace.Changes) != 1 {
		t.Fatalf("creation audit was not truthful and minimal: %+v", workspace.Changes)
	}
	if session.Path == base || !strings.Contains(session.Path, "candidates") {
		t.Fatalf("workspace was not placed in candidate-specific application state: %s", session.Path)
	}
	if _, err = os.Stat(filepath.Join(session.Path, workspaceIdentityFile)); err != nil {
		t.Fatalf("workspace identity is not visible on disk: %v", err)
	}
}

func TestRestartReopensOnlyCandidateOwnedWorkspace(t *testing.T) {
	base := t.TempDir()
	synthetic := t.TempDir()
	runtime := runtimeFor("candidate-restart", "ECO-CANDIDATE-RESTART", Schema)
	first, err := StartCandidate(base, runtime)
	if err != nil {
		t.Fatal(err)
	}
	item := importSynthetic(t, first.Current.Vault, synthetic, "restart.txt", "Synthetic restart wording 412.")
	first.Current.Vault.Close()

	restarted, err := StartCandidate(base, runtime)
	if err != nil {
		t.Fatal(err)
	}
	workspace := restarted.Current.Vault.Snapshot()
	if restarted.Current.Disposition != DispositionReopened || restarted.Current.Explicit {
		t.Fatalf("restart status is unclear: %+v", restarted.Current)
	}
	if len(workspace.Evidence) != 1 || workspace.Evidence[0].ID != item.ID {
		t.Fatalf("candidate-owned workspace did not reopen truthfully: %+v", workspace.Evidence)
	}
	if !hasChange(workspace, "workspace-reopened") {
		t.Fatal("workspace reopen was not audited")
	}
}

func TestNewCandidateDoesNotInheritPriorCandidateState(t *testing.T) {
	base := t.TempDir()
	synthetic := t.TempDir()
	oldRuntime := runtimeFor("candidate-old", "ECO-CANDIDATE-OLD", Schema)
	oldApp, err := StartCandidate(base, oldRuntime)
	if err != nil {
		t.Fatal(err)
	}
	importSynthetic(t, oldApp.Current.Vault, synthetic, "old.txt", "Synthetic evidence belonging only to the old candidate.")
	if _, err = oldApp.Current.Vault.CreateMatter("Synthetic old matter"); err != nil {
		t.Fatal(err)
	}
	oldApp.Current.Vault.Ask("What does the synthetic evidence say?", nil)
	if _, err = oldApp.Current.Vault.ToggleLowSensory(); err != nil {
		t.Fatal(err)
	}

	newRuntime := runtimeFor("candidate-new", "ECO-CANDIDATE-NEW", Schema)
	newApp, err := StartCandidate(base, newRuntime)
	if err != nil {
		t.Fatal(err)
	}
	fresh := newApp.Current.Vault.Snapshot()
	if newApp.Current.Path == oldApp.Current.Path {
		t.Fatal("different candidates shared an application-state or workspace path")
	}
	if len(fresh.Evidence) != 0 || len(fresh.Matters) != 0 || len(fresh.Questions) != 0 || fresh.Settings != (Settings{}) {
		t.Fatalf("new candidate silently inherited old candidate data: %+v", fresh)
	}
	if fresh.BuildID != newRuntime.BuildID || fresh.CreatedByBuild != newRuntime.BuildID {
		t.Fatalf("fresh workspace was attributed to the wrong build: %+v", fresh)
	}
}

func TestCandidateIgnoresLegacyImplicitStateLocation(t *testing.T) {
	base := t.TempDir()
	synthetic := t.TempDir()
	runtime := runtimeFor("candidate-after-legacy", "ECO-CANDIDATE-AFTER-LEGACY", Schema)
	legacyPath := filepath.Join(base, "V25N2")
	legacy, err := createVault(legacyPath, "Earlier implicit test workspace", runtime)
	if err != nil {
		t.Fatal(err)
	}
	importSynthetic(t, legacy, synthetic, "legacy-location.txt", "Synthetic data in the earlier implicit application-state location.")
	legacy.Close()

	app, err := StartCandidate(base, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if app.Current.Path == legacyPath {
		t.Fatal("new candidate silently reused the earlier fixed application-state location")
	}
	workspace := app.Current.Vault.Snapshot()
	if len(workspace.Evidence) != 0 || len(workspace.Questions) != 0 || len(workspace.Matters) != 0 {
		t.Fatalf("new candidate inherited data from the earlier implicit location: %+v", workspace)
	}
	old, err := openVault(legacyPath, runtime)
	if err != nil || len(old.Snapshot().Evidence) != 1 {
		t.Fatalf("earlier test workspace was changed instead of ignored: workspace=%+v err=%v", old, err)
	}
}

func TestOpenExistingDoesNotCreateWorkspaceInUnrelatedFolder(t *testing.T) {
	root := t.TempDir()
	unrelated := filepath.Join(root, "unrelated.keep")
	if err := os.WriteFile(unrelated, []byte("synthetic unrelated file"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenVault(root); err == nil {
		t.Fatal("an unrelated folder was silently turned into an ECO workspace")
	}
	for _, managed := range []string{"vault.key", "workspace.ecodb", workspaceIdentityFile, "objects"} {
		if _, err := os.Stat(filepath.Join(root, managed)); !os.IsNotExist(err) {
			t.Fatalf("deliberate reopen created managed state %s in an unrelated folder", managed)
		}
	}
	data, err := os.ReadFile(unrelated)
	if err != nil || string(data) != "synthetic unrelated file" {
		t.Fatalf("deliberate reopen changed the unrelated file: %q err=%v", data, err)
	}
}

func TestExistingSelectedWorkspaceRequiresExplicitOpen(t *testing.T) {
	base := t.TempDir()
	synthetic := t.TempDir()
	externalParent := t.TempDir()
	external := filepath.Join(externalParent, "selected-workspace")
	runtime := runtimeFor("candidate-explicit", "ECO-CANDIDATE-EXPLICIT", Schema)
	app, err := StartCandidate(base, runtime)
	if err != nil {
		t.Fatal(err)
	}
	defaultPath := app.Current.Path
	selected, err := app.CreateWorkspace(external, "Explicit synthetic workspace")
	if err != nil {
		t.Fatal(err)
	}
	item := importSynthetic(t, selected.Vault, synthetic, "selected.txt", "Synthetic explicitly selected wording.")
	selected.Vault.Close()

	restarted, err := StartCandidate(base, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if restarted.Current.Path != defaultPath || len(restarted.Current.Vault.Snapshot().Evidence) != 0 {
		t.Fatal("restart silently reopened a user-selected external workspace")
	}
	reopened, err := restarted.OpenWorkspace(external)
	if err != nil {
		t.Fatal(err)
	}
	if !reopened.Explicit || reopened.Disposition != DispositionReopened {
		t.Fatalf("explicit reopen was not reported: %+v", reopened)
	}
	got := reopened.Vault.Snapshot()
	if len(got.Evidence) != 1 || got.Evidence[0].ID != item.ID {
		t.Fatalf("explicitly selected workspace did not reopen its own data: %+v", got.Evidence)
	}
}

func TestFreshWorkspaceCannotDisplayOldDataWithoutExplicitAction(t *testing.T) {
	base := t.TempDir()
	synthetic := t.TempDir()
	runtime := runtimeFor("candidate-fresh", "ECO-CANDIDATE-FRESH", Schema)
	app, err := StartCandidate(base, runtime)
	if err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(t.TempDir(), "old-workspace")
	oldSession, err := app.CreateWorkspace(oldPath, "Old synthetic workspace")
	if err != nil {
		t.Fatal(err)
	}
	importSynthetic(t, oldSession.Vault, synthetic, "old-record.txt", "Old synthetic record wording.")
	oldSession.Vault.Ask("Record this old synthetic conversation.", nil)
	freshPath := filepath.Join(t.TempDir(), "fresh-workspace")
	fresh, err := app.CreateWorkspace(freshPath, "Fresh synthetic workspace")
	if err != nil {
		t.Fatal(err)
	}
	workspace := fresh.Vault.Snapshot()
	if len(workspace.Evidence) != 0 || len(workspace.Questions) != 0 || len(workspace.Matters) != 0 {
		t.Fatalf("fresh workspace displayed old evidence, conversations or records: %+v", workspace)
	}
	if fresh.Identity.ID == oldSession.Identity.ID || fresh.Path == oldSession.Path {
		t.Fatal("fresh workspace reused the old workspace identity")
	}
}

func TestCompatibleBuildUpgradeRequiresExplicitReopenAndIsAudited(t *testing.T) {
	root := filepath.Join(t.TempDir(), "compatible")
	synthetic := t.TempDir()
	oldRuntime := runtimeFor("compatible-old", "ECO-COMPATIBLE-OLD", Schema)
	oldVault, err := createVault(root, "Compatible synthetic workspace", oldRuntime)
	if err != nil {
		t.Fatal(err)
	}
	item := importSynthetic(t, oldVault, synthetic, "compatible.txt", "Compatible synthetic upgrade data.")
	oldVault.Close()
	newRuntime := runtimeFor("compatible-new", "ECO-COMPATIBLE-NEW", Schema)
	_, report, err := InspectWorkspace(root, newRuntime)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != CompatibilityCompatibleBuild || !report.CanOpen || report.CanMigrate {
		t.Fatalf("same-schema workspace was not reported as compatible: %+v", report)
	}
	app, err := StartCandidate(t.TempDir(), newRuntime)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := app.OpenWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	got := reopened.Vault.Snapshot()
	if got.BuildID != newRuntime.BuildID || len(got.Evidence) != 1 || got.Evidence[0].ID != item.ID || !hasChange(got, "workspace-reopened") {
		t.Fatalf("compatible reopen was not truthful: %+v", got)
	}
}

func TestUnsupportedOlderWorkspaceIsBlockedWithoutChange(t *testing.T) {
	root := filepath.Join(t.TempDir(), "unsupported")
	oldRuntime := runtimeFor("format-three", "ECO-FORMAT-THREE", 3)
	vault, err := createVault(root, "Unsupported synthetic workspace", oldRuntime)
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, "workspace.ecodb"))
	if err != nil {
		t.Fatal(err)
	}
	vault.Close()
	newRuntime := runtimeFor("format-four", "ECO-FORMAT-FOUR", 4)
	_, report, err := InspectWorkspace(root, newRuntime)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != CompatibilityTooOld || report.CanOpen || report.CanMigrate {
		t.Fatalf("unsupported older format was not blocked: %+v", report)
	}
	if _, err = openVault(root, newRuntime); err == nil {
		t.Fatal("unsupported older format opened")
	}
	after, err := os.ReadFile(filepath.Join(root, "workspace.ecodb"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("blocked incompatible upgrade changed the workspace")
	}
}

func TestDowngradeAttemptIsBlockedWithoutChange(t *testing.T) {
	root := filepath.Join(t.TempDir(), "newer")
	newRuntime := runtimeFor("format-three", "ECO-FORMAT-THREE", 3)
	vault, err := createVault(root, "Newer synthetic workspace", newRuntime)
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, "workspace.ecodb"))
	if err != nil {
		t.Fatal(err)
	}
	vault.Close()
	olderRuntime := runtimeFor("format-two", "ECO-FORMAT-TWO", 2)
	_, report, err := InspectWorkspace(root, olderRuntime)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != CompatibilityDowngradeBlocked || report.CanOpen || report.CanMigrate {
		t.Fatalf("downgrade was not blocked clearly: %+v", report)
	}
	if _, err = openVault(root, olderRuntime); err == nil {
		t.Fatal("newer workspace opened in an older build")
	}
	after, err := os.ReadFile(filepath.Join(root, "workspace.ecodb"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("blocked downgrade changed the workspace")
	}
}

func TestOlderWorkspaceMigratesThroughPreservedCheckpoint(t *testing.T) {
	root := filepath.Join(t.TempDir(), "legacy")
	synthetic := t.TempDir()
	oldRuntime := runtimeFor("legacy", "ECO-LEGACY-SCHEMA-1", 1)
	legacy := createPriorSchemaOneWorkspace(t, root, oldRuntime)
	item := importSynthetic(t, legacy, synthetic, "legacy.txt", "Synthetic data preserved through migration.")
	legacy.Close()
	current := runtimeFor("current", "ECO-CURRENT-SCHEMA-2", 2)

	session, receipt, err := MigrateWorkspace(root, current)
	if err != nil {
		t.Fatal(err)
	}
	got := session.Vault.Snapshot()
	if session.Disposition != DispositionMigrated || receipt.FromSchema != 1 || receipt.ToSchema != 2 {
		t.Fatalf("migration status is incomplete: session=%+v receipt=%+v", session, receipt)
	}
	if got.Schema != 2 || got.BuildID != current.BuildID || len(got.Evidence) != 1 || got.Evidence[0].ID != item.ID || !hasChange(got, "workspace-migrated") {
		t.Fatalf("migrated workspace is not truthful: %+v", got)
	}
	if _, err = os.Stat(receipt.Checkpoint); err != nil {
		t.Fatalf("original migration checkpoint was not retained: %v", err)
	}
	original, err := openVault(receipt.Checkpoint, oldRuntime)
	if err != nil {
		t.Fatalf("preserved original checkpoint is not readable by its source runtime: %v", err)
	}
	if original.Snapshot().Schema != 1 || len(original.Snapshot().Evidence) != 1 {
		t.Fatal("migration checkpoint did not preserve the original state")
	}
}

func TestInterruptedMigrationRollsBackToOriginalWorkspace(t *testing.T) {
	root := filepath.Join(t.TempDir(), "interrupted")
	synthetic := t.TempDir()
	oldRuntime := runtimeFor("legacy-interrupted", "ECO-LEGACY-INTERRUPTED", 1)
	legacy := createPriorSchemaOneWorkspace(t, root, oldRuntime)
	item := importSynthetic(t, legacy, synthetic, "interrupted.txt", "Synthetic interrupted migration data.")
	legacy.Close()
	current := runtimeFor("current-interrupted", "ECO-CURRENT-INTERRUPTED", 2)
	_, _, err := migrateWorkspace(root, current, func(phase MigrationPhase) error {
		if phase == migrationStageReady {
			return errMigrationInterrupted
		}
		return nil
	})
	if !errors.Is(err, errMigrationInterrupted) {
		t.Fatalf("migration was not interrupted at the requested recovery point: %v", err)
	}
	if _, statErr := os.Stat(root); !os.IsNotExist(statErr) {
		t.Fatalf("interrupted migration unexpectedly left an active root: %v", statErr)
	}
	_, receipt, recoverErr := RecoverWorkspace(root, current)
	var compatibilityErr *CompatibilityError
	if !errors.As(recoverErr, &compatibilityErr) {
		t.Fatalf("recovery should restore the old format and require a deliberate migration: receipt=%+v err=%v", receipt, recoverErr)
	}
	if !receipt.OriginalRestored || receipt.Compatibility.Status != CompatibilityMigrationNeeded {
		t.Fatalf("interrupted migration did not report safe rollback: %+v", receipt)
	}
	restored, err := openVault(root, oldRuntime)
	if err != nil {
		t.Fatal(err)
	}
	got := restored.Snapshot()
	if got.Schema != 1 || len(got.Evidence) != 1 || got.Evidence[0].ID != item.ID {
		t.Fatalf("rollback did not restore the exact old workspace: %+v", got)
	}
	if _, err = os.Stat(migrationStatePath(root)); !os.IsNotExist(err) {
		t.Fatal("recovery marker remained after rollback")
	}
}

func TestActivatedMigrationCanRecoverAfterInterruption(t *testing.T) {
	root := filepath.Join(t.TempDir(), "activated")
	oldRuntime := runtimeFor("legacy-activated", "ECO-LEGACY-ACTIVATED", 1)
	legacy := createPriorSchemaOneWorkspace(t, root, oldRuntime)
	legacy.Close()
	current := runtimeFor("current-activated", "ECO-CURRENT-ACTIVATED", 2)
	_, _, err := migrateWorkspace(root, current, func(phase MigrationPhase) error {
		if phase == migrationActivated {
			return errMigrationInterrupted
		}
		return nil
	})
	if !errors.Is(err, errMigrationInterrupted) {
		t.Fatalf("migration was not interrupted after activation: %v", err)
	}
	session, receipt, err := RecoverWorkspace(root, current)
	if err != nil {
		t.Fatal(err)
	}
	if session.Disposition != DispositionRecovered || !receipt.MigrationKept || !receipt.WorkspaceOpened || !hasChange(session.Vault.Snapshot(), "workspace-recovered") {
		t.Fatalf("activated migration was not recovered truthfully: session=%+v receipt=%+v", session, receipt)
	}
	if _, err = os.Stat(receipt.Checkpoint); err != nil {
		t.Fatalf("recovery discarded the original checkpoint: %v", err)
	}
}

func TestResetAffectsOnlySelectedWorkspaceAndManagedObjects(t *testing.T) {
	base := t.TempDir()
	synthetic := t.TempDir()
	runtime := runtimeFor("candidate-reset", "ECO-CANDIDATE-RESET", Schema)
	app, err := StartCandidate(base, runtime)
	if err != nil {
		t.Fatal(err)
	}
	firstPath := filepath.Join(t.TempDir(), "selected-reset")
	secondPath := filepath.Join(t.TempDir(), "other-workspace")
	first, err := createVault(firstPath, "Selected reset workspace", runtime)
	if err != nil {
		t.Fatal(err)
	}
	firstItem := importSynthetic(t, first, synthetic, "selected-source.txt", "Synthetic selected reset evidence.")
	first.Ask("Keep this only until the selected reset.", nil)
	if _, err = first.CreateMatter("Synthetic reset matter"); err != nil {
		t.Fatal(err)
	}
	if _, err = first.ToggleLowSensory(); err != nil {
		t.Fatal(err)
	}
	unrelatedInside := filepath.Join(firstPath, "unrelated.keep")
	if err = os.WriteFile(unrelatedInside, []byte("unrelated local file"), 0600); err != nil {
		t.Fatal(err)
	}
	second, err := createVault(secondPath, "Other synthetic workspace", runtime)
	if err != nil {
		t.Fatal(err)
	}
	secondItem := importSynthetic(t, second, synthetic, "other-source.txt", "Synthetic other workspace evidence.")
	second.Close()
	sourcePath := filepath.Join(synthetic, "selected-source.txt")
	first.Close()
	if _, err = app.OpenWorkspace(firstPath); err != nil {
		t.Fatal(err)
	}

	session, receipt, err := app.ResetCurrentWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	reset := session.Vault.Snapshot()
	if session.Disposition != DispositionReset || receipt.WorkspaceID != first.Identity.ID {
		t.Fatalf("reset target or status is unclear: session=%+v receipt=%+v", session, receipt)
	}
	if len(reset.Evidence) != 0 || len(reset.Preservations) != 0 || len(reset.Matters) != 0 || len(reset.Questions) != 0 || reset.Settings != (Settings{}) {
		t.Fatalf("selected workspace was not genuinely reset: %+v", reset)
	}
	if !hasChange(reset, "workspace-reset") || reset.Changes[0].PrevHash == "" {
		t.Fatalf("reset audit did not preserve truthful continuity: %+v", reset.Changes)
	}
	if _, err = os.Stat(filepath.Join(firstPath, "objects", firstItem.ObjectFile)); !os.IsNotExist(err) {
		t.Fatal("selected workspace's managed encrypted object was not retired")
	}
	for _, path := range []string{unrelatedInside, sourcePath} {
		if _, err = os.Stat(path); err != nil {
			t.Fatalf("reset removed unrelated or source evidence file %s: %v", path, err)
		}
	}
	other, err := openVault(secondPath, runtime)
	if err != nil {
		t.Fatal(err)
	}
	otherState := other.Snapshot()
	if len(otherState.Evidence) != 1 || otherState.Evidence[0].ID != secondItem.ID {
		t.Fatalf("reset changed another workspace: %+v", otherState.Evidence)
	}
}
