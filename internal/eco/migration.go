package eco

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MigrationPhase string

const (
	migrationPrepared      MigrationPhase = "prepared"
	migrationOriginalMoved MigrationPhase = "original-moved"
	migrationStageReady    MigrationPhase = "stage-ready"
	migrationActivated     MigrationPhase = "activated"
)

var errMigrationInterrupted = errors.New("simulated migration interruption")

type migrationState struct {
	Format           string         `json:"format"`
	Root             string         `json:"root"`
	Checkpoint       string         `json:"checkpoint"`
	Stage            string         `json:"stage"`
	WorkspaceID      string         `json:"workspace_id,omitempty"`
	FromSchema       int            `json:"from_schema"`
	ToSchema         int            `json:"to_schema"`
	SourceBuild      string         `json:"source_build"`
	DestinationBuild string         `json:"destination_build"`
	Phase            MigrationPhase `json:"phase"`
	StartedAt        time.Time      `json:"started_at"`
}

type MigrationReceipt struct {
	WorkspaceID string
	Path        string
	Checkpoint  string
	FromSchema  int
	ToSchema    int
	SourceBuild string
	BuildID     string
}

type RecoveryReceipt struct {
	Path             string
	Checkpoint       string
	OriginalRestored bool
	MigrationKept    bool
	WorkspaceOpened  bool
	Message          string
	Compatibility    WorkspaceCompatibility
}

func migrationStatePath(root string) string {
	return root + ".eco-migration.json"
}

func MigrateWorkspace(root string, runtime RuntimeIdentity) (WorkspaceSession, MigrationReceipt, error) {
	return migrateWorkspace(root, runtime, nil)
}

func migrateWorkspace(root string, runtime RuntimeIdentity, hook func(MigrationPhase) error) (WorkspaceSession, MigrationReceipt, error) {
	inspected, err := inspectWorkspace(root, runtime)
	if err != nil {
		return WorkspaceSession{}, MigrationReceipt{}, err
	}
	defer zeroBytes(inspected.key)
	if !inspected.Compatibility.CanMigrate {
		if inspected.Compatibility.CanOpen {
			return WorkspaceSession{}, MigrationReceipt{}, errors.New("this workspace is already compatible and does not need migration")
		}
		return WorkspaceSession{}, MigrationReceipt{}, &CompatibilityError{Report: inspected.Compatibility}
	}

	parent := filepath.Dir(inspected.Path)
	base := filepath.Base(inspected.Path)
	suffix := time.Now().UTC().Format("20060102T150405.000000000Z") + "-" + NewID("MIG")
	state := migrationState{
		Format:           "ECO-MIGRATION-1",
		Root:             inspected.Path,
		Checkpoint:       filepath.Join(parent, base+".migration-checkpoint-"+suffix),
		Stage:            filepath.Join(parent, base+".migration-stage-"+suffix),
		WorkspaceID:      inspected.Identity.ID,
		FromSchema:       inspected.Workspace.Schema,
		ToSchema:         runtime.Schema,
		SourceBuild:      inspected.Workspace.BuildID,
		DestinationBuild: runtime.BuildID,
		Phase:            migrationPrepared,
		StartedAt:        time.Now().UTC(),
	}
	if err = validateMigrationState(state, inspected.Path); err != nil {
		return WorkspaceSession{}, MigrationReceipt{}, err
	}
	if err = writeMigrationState(state); err != nil {
		return WorkspaceSession{}, MigrationReceipt{}, err
	}
	if err = runMigrationHook(hook, state.Phase); err != nil {
		return WorkspaceSession{}, MigrationReceipt{}, err
	}

	failed := func(cause error) (WorkspaceSession, MigrationReceipt, error) {
		restored, rollbackErr := rollbackMigrationState(state)
		if rollbackErr != nil {
			return WorkspaceSession{}, MigrationReceipt{}, fmt.Errorf("%v; automatic rollback also needs attention: %w", cause, rollbackErr)
		}
		if restored {
			return WorkspaceSession{}, MigrationReceipt{}, fmt.Errorf("%v. The original workspace was restored and no old data was presented as migrated", cause)
		}
		return WorkspaceSession{}, MigrationReceipt{}, cause
	}

	if err = os.Rename(state.Root, state.Checkpoint); err != nil {
		return failed(fmt.Errorf("the original workspace could not be checkpointed: %w", err))
	}
	state.Phase = migrationOriginalMoved
	if err = writeMigrationState(state); err != nil {
		return failed(fmt.Errorf("the migration checkpoint could not be recorded: %w", err))
	}
	if err = runMigrationHook(hook, state.Phase); err != nil {
		if errors.Is(err, errMigrationInterrupted) {
			return WorkspaceSession{}, MigrationReceipt{}, err
		}
		return failed(err)
	}

	if err = copyWorkspaceTree(state.Checkpoint, state.Stage); err != nil {
		return failed(fmt.Errorf("the migration staging copy could not be created: %w", err))
	}
	if err = migrateStagedWorkspace(state, runtime); err != nil {
		return failed(err)
	}
	state.Phase = migrationStageReady
	if err = writeMigrationState(state); err != nil {
		return failed(fmt.Errorf("the validated migration could not be recorded: %w", err))
	}
	if err = runMigrationHook(hook, state.Phase); err != nil {
		if errors.Is(err, errMigrationInterrupted) {
			return WorkspaceSession{}, MigrationReceipt{}, err
		}
		return failed(err)
	}

	if err = os.Rename(state.Stage, state.Root); err != nil {
		return failed(fmt.Errorf("the migrated workspace could not be activated: %w", err))
	}
	state.Phase = migrationActivated
	if err = writeMigrationState(state); err != nil {
		return failed(fmt.Errorf("the activated migration could not be recorded: %w", err))
	}
	if err = runMigrationHook(hook, state.Phase); err != nil {
		if errors.Is(err, errMigrationInterrupted) {
			return WorkspaceSession{}, MigrationReceipt{}, err
		}
		return failed(err)
	}

	v, err := openVaultIgnoringRecovery(state.Root, runtime)
	if err != nil {
		return failed(fmt.Errorf("the migrated workspace did not pass its final reopening check: %w", err))
	}
	if err = os.Remove(migrationStatePath(state.Root)); err != nil {
		v.Close()
		return failed(fmt.Errorf("the completed migration marker could not be cleared: %w", err))
	}
	receipt := MigrationReceipt{
		WorkspaceID: v.Identity.ID,
		Path:        v.Root,
		Checkpoint:  state.Checkpoint,
		FromSchema:  state.FromSchema,
		ToSchema:    state.ToSchema,
		SourceBuild: state.SourceBuild,
		BuildID:     runtime.BuildID,
	}
	return sessionFor(v, DispositionMigrated, true, state.Checkpoint), receipt, nil
}

func runMigrationHook(hook func(MigrationPhase) error, phase MigrationPhase) error {
	if hook == nil {
		return nil
	}
	return hook(phase)
}

func migrateStagedWorkspace(state migrationState, runtime RuntimeIdentity) error {
	key, err := loadExistingMasterKey(filepath.Join(state.Stage, "vault.key"))
	if err != nil {
		return fmt.Errorf("the staged workspace key could not be opened: %w", err)
	}
	defer zeroBytes(key)
	ws, err := readEncryptedWorkspace(state.Stage, key)
	if err != nil {
		return err
	}
	if ws.Schema != state.FromSchema || state.FromSchema != 1 || state.ToSchema != 2 {
		return errors.New("this workspace format does not have an approved recoverable migration path")
	}
	id := strings.TrimSpace(ws.WorkspaceID)
	if id == "" {
		id = NewID("WS")
	}
	name := strings.TrimSpace(ws.WorkspaceName)
	if !validWorkspaceName(name) {
		name = "Migrated ECO workspace"
	}
	createdBy := strings.TrimSpace(ws.CreatedByBuild)
	if createdBy == "" {
		createdBy = layBuild(ws.BuildID)
	}
	createdAt := ws.CreatedAt
	if createdAt.IsZero() {
		createdAt = state.StartedAt
	}
	ws.Schema = runtime.Schema
	ws.BuildID = runtime.BuildID
	ws.WorkspaceID = id
	ws.WorkspaceName = name
	ws.CreatedByBuild = createdBy
	ws.CreatedAt = createdAt
	if ws.Evidence == nil {
		ws.Evidence = []EvidenceItem{}
	}
	if ws.Preservations == nil {
		ws.Preservations = []PreservationRecord{}
	}
	if ws.Matters == nil {
		ws.Matters = []Matter{}
	}
	if ws.Changes == nil {
		ws.Changes = []ChangeRecord{}
	}
	if ws.Questions == nil {
		ws.Questions = []QuestionRecord{}
	}
	identity := WorkspaceIdentity{
		Format:         workspaceIdentityFormat,
		ID:             id,
		Name:           name,
		Kind:           "development",
		Schema:         runtime.Schema,
		CreatedAt:      createdAt,
		CreatedByBuild: createdBy,
	}
	v := &Vault{
		Root:      state.Stage,
		Objects:   filepath.Join(state.Stage, "objects"),
		Identity:  identity,
		key:       key,
		runtime:   runtime,
		Workspace: ws,
	}
	v.addChangeUnlocked("system", "workspace-migrated", "Migrated an older ECO workspace through a recoverable checkpoint", map[string]any{
		"from_schema":  state.FromSchema,
		"to_schema":    state.ToSchema,
		"source_build": state.SourceBuild,
		"build":        runtime.BuildID,
		"checkpoint":   filepath.Base(state.Checkpoint),
	})
	if err = v.Save(); err != nil {
		return fmt.Errorf("the migrated workspace record could not be saved: %w", err)
	}
	if err = writeWorkspaceIdentity(state.Stage, identity); err != nil {
		return err
	}
	if err = verifyStagedVault(v); err != nil {
		return fmt.Errorf("the migrated workspace failed evidence verification: %w", err)
	}
	return nil
}

func RecoverWorkspace(root string, runtime RuntimeIdentity) (WorkspaceSession, RecoveryReceipt, error) {
	absolute, err := normaliseWorkspaceRoot(root)
	if err != nil {
		return WorkspaceSession{}, RecoveryReceipt{}, err
	}
	state, err := readMigrationState(absolute)
	if err != nil {
		return WorkspaceSession{}, RecoveryReceipt{}, err
	}
	if err = validateMigrationState(state, absolute); err != nil {
		return WorkspaceSession{}, RecoveryReceipt{}, err
	}
	receipt := RecoveryReceipt{Path: absolute, Checkpoint: state.Checkpoint}

	if state.Phase == migrationActivated {
		v, openErr := openVaultIgnoringRecovery(absolute, runtime)
		if openErr == nil {
			if err = v.recordLifecycle("system", "workspace-recovered", "Recovered and verified an interrupted workspace migration", map[string]any{
				"checkpoint": filepath.Base(state.Checkpoint),
				"phase":      state.Phase,
			}); err != nil {
				v.Close()
				return WorkspaceSession{}, receipt, err
			}
			_ = removeMigrationStage(state)
			if err = os.Remove(migrationStatePath(absolute)); err != nil {
				v.Close()
				return WorkspaceSession{}, receipt, fmt.Errorf("the recovered migration marker could not be cleared: %w", err)
			}
			receipt.MigrationKept = true
			receipt.WorkspaceOpened = true
			receipt.Message = "The migrated workspace was verified and recovered. Its original checkpoint was kept."
			receipt.Compatibility = compatibilityFor(v.Workspace, runtime)
			return sessionFor(v, DispositionRecovered, true, state.Checkpoint), receipt, nil
		}
	}

	restored, err := rollbackMigrationState(state)
	if err != nil {
		return WorkspaceSession{}, receipt, err
	}
	receipt.OriginalRestored = restored
	receipt.Message = "The interrupted migration was rolled back. The original workspace was restored without presenting it as new-build data."
	inspected, inspectErr := inspectWorkspaceAt(absolute, runtime, false)
	if inspectErr != nil {
		return WorkspaceSession{}, receipt, inspectErr
	}
	receipt.Compatibility = inspected.Compatibility
	zeroBytes(inspected.key)
	if !inspected.Compatibility.CanOpen {
		return WorkspaceSession{}, receipt, &CompatibilityError{Report: inspected.Compatibility}
	}
	v, err := openVaultIgnoringRecovery(absolute, runtime)
	if err != nil {
		return WorkspaceSession{}, receipt, err
	}
	if err = v.recordLifecycle("system", "workspace-recovered", "Recovered the original workspace after an interrupted migration", map[string]any{
		"phase": state.Phase,
	}); err != nil {
		v.Close()
		return WorkspaceSession{}, receipt, err
	}
	receipt.WorkspaceOpened = true
	return sessionFor(v, DispositionRecovered, true, ""), receipt, nil
}

func rollbackMigrationState(state migrationState) (bool, error) {
	if err := validateMigrationState(state, state.Root); err != nil {
		return false, err
	}
	stageCleanupErr := removeMigrationStage(state)
	rootInfo, rootErr := os.Stat(state.Root)
	checkpointInfo, checkpointErr := os.Stat(state.Checkpoint)
	rootExists := rootErr == nil && rootInfo.IsDir()
	checkpointExists := checkpointErr == nil && checkpointInfo.IsDir()
	if rootErr != nil && !os.IsNotExist(rootErr) {
		return false, fmt.Errorf("inspect migration target: %w", rootErr)
	}
	if checkpointErr != nil && !os.IsNotExist(checkpointErr) {
		return false, fmt.Errorf("inspect migration checkpoint: %w", checkpointErr)
	}

	restored := false
	if checkpointExists {
		if rootExists {
			failed := state.Stage + ".failed-" + NewID("ROLLBACK")
			if err := os.Rename(state.Root, failed); err != nil {
				return false, fmt.Errorf("move failed migrated workspace aside: %w", err)
			}
		}
		if err := os.Rename(state.Checkpoint, state.Root); err != nil {
			return false, fmt.Errorf("restore original workspace checkpoint: %w", err)
		}
		restored = true
	} else if !rootExists {
		return false, errors.New("neither the workspace nor its migration checkpoint is available; automatic recovery was blocked")
	}
	if err := os.Remove(migrationStatePath(state.Root)); err != nil && !os.IsNotExist(err) {
		return restored, fmt.Errorf("clear migration state: %w", err)
	}
	if stageCleanupErr != nil {
		return restored, fmt.Errorf("the original workspace was restored, but the separate migration staging folder could not be removed: %w", stageCleanupErr)
	}
	return restored, nil
}

func writeMigrationState(state migrationState) error {
	if err := validateMigrationState(state, state.Root); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	path := migrationStatePath(state.Root)
	tmp := path + ".tmp"
	if err = os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write migration recovery state: %w", err)
	}
	if err = os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("activate migration recovery state: %w", err)
	}
	return nil
}

func readMigrationState(root string) (migrationState, error) {
	data, err := os.ReadFile(migrationStatePath(root))
	if err != nil {
		if os.IsNotExist(err) {
			return migrationState{}, errors.New("this workspace has no unfinished migration to recover")
		}
		return migrationState{}, err
	}
	var state migrationState
	if err = json.Unmarshal(data, &state); err != nil {
		return migrationState{}, errors.New("the migration recovery record is damaged; automatic recovery was blocked")
	}
	return state, nil
}

func validateMigrationState(state migrationState, expectedRoot string) error {
	root, err := normaliseWorkspaceRoot(expectedRoot)
	if err != nil {
		return err
	}
	if state.Format != "ECO-MIGRATION-1" || state.Root != root {
		return errors.New("the migration recovery record does not match the selected workspace")
	}
	parent := filepath.Dir(root)
	base := filepath.Base(root)
	for path, prefix := range map[string]string{
		state.Checkpoint: base + ".migration-checkpoint-",
		state.Stage:      base + ".migration-stage-",
	} {
		clean := filepath.Clean(path)
		if filepath.Dir(clean) != parent || !strings.HasPrefix(filepath.Base(clean), prefix) {
			return errors.New("the migration recovery paths are unsafe; automatic changes were blocked")
		}
	}
	if state.FromSchema < 1 || state.ToSchema <= state.FromSchema {
		return errors.New("the migration format transition is invalid")
	}
	switch state.Phase {
	case migrationPrepared, migrationOriginalMoved, migrationStageReady, migrationActivated:
	default:
		return errors.New("the migration recovery phase is invalid")
	}
	return nil
}

func removeMigrationStage(state migrationState) error {
	if err := validateMigrationState(state, state.Root); err != nil {
		return err
	}
	if err := os.RemoveAll(state.Stage); err != nil {
		return fmt.Errorf("remove migration staging folder: %w", err)
	}
	return nil
}

func copyWorkspaceTree(source, destination string) error {
	sourceInfo, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if !sourceInfo.IsDir() || sourceInfo.Mode()&os.ModeSymlink != 0 {
		return errors.New("the workspace checkpoint is not a normal folder")
	}
	if _, err = os.Stat(destination); err == nil {
		return errors.New("the migration staging folder already exists")
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = os.Mkdir(destination, 0700); err != nil {
		return err
	}
	ok := false
	defer func() {
		if !ok {
			_ = os.RemoveAll(destination)
		}
	}()
	err = filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == source {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("workspace migration will not follow symbolic links")
		}
		relative, err := filepath.Rel(source, path)
		if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return errors.New("workspace migration found an unsafe path")
		}
		target := filepath.Join(destination, relative)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.Mkdir(target, info.Mode().Perm())
		}
		if !info.Mode().IsRegular() {
			return errors.New("workspace migration only accepts normal files and folders")
		}
		return copyRegularFile(path, target, info.Mode().Perm())
	})
	if err != nil {
		return err
	}
	ok = true
	return nil
}

func copyRegularFile(source, destination string, mode os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = output.Close()
		if !ok {
			_ = os.Remove(destination)
		}
	}()
	if _, err = io.Copy(output, input); err != nil {
		return err
	}
	if err = output.Sync(); err != nil {
		return err
	}
	if err = output.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}
