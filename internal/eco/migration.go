package eco

import (
	"bytes"
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
	migrationPrepared        MigrationPhase = "prepared"
	migrationOriginalMoved   MigrationPhase = "original-moved"
	migrationStageUnverified MigrationPhase = "stage-unverified"
	migrationStageReady      MigrationPhase = "stage-ready"
	migrationActivated       MigrationPhase = "activated"
)

var errMigrationInterrupted = errors.New("simulated migration interruption")

type migrationState struct {
	Format               string         `json:"format"`
	Root                 string         `json:"root"`
	Checkpoint           string         `json:"checkpoint"`
	Stage                string         `json:"stage"`
	Failed               string         `json:"failed"`
	Nonce                string         `json:"nonce"`
	WorkspaceID          string         `json:"workspace_id"`
	SourceWorkspaceID    string         `json:"source_workspace_id,omitempty"`
	CheckpointID         string         `json:"checkpoint_id"`
	StageID              string         `json:"stage_id"`
	FailedID             string         `json:"failed_id"`
	FromSchema           int            `json:"from_schema"`
	ToSchema             int            `json:"to_schema"`
	SourceBuild          string         `json:"source_build"`
	DestinationBuild     string         `json:"destination_build"`
	SourceCandidate      string         `json:"source_candidate,omitempty"`
	DestinationCandidate string         `json:"destination_candidate"`
	Phase                MigrationPhase `json:"phase"`
	StartedAt            time.Time      `json:"started_at"`
	AuthTag              string         `json:"auth_tag"`
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

	nonce, err := newMigrationNonce()
	if err != nil {
		return WorkspaceSession{}, MigrationReceipt{}, fmt.Errorf("create migration identity: %w", err)
	}
	checkpoint, stage, failedPath := migrationSiblingPaths(inspected.Path, nonce)
	workspaceID := inspected.Workspace.WorkspaceID
	if !safeRecordID(workspaceID) {
		workspaceID = NewID("WS")
	}
	state := migrationState{
		Format:               migrationFormat,
		Root:                 inspected.Path,
		Checkpoint:           checkpoint,
		Stage:                stage,
		Failed:               failedPath,
		Nonce:                nonce,
		WorkspaceID:          workspaceID,
		SourceWorkspaceID:    inspected.Workspace.WorkspaceID,
		CheckpointID:         NewID("MCP"),
		StageID:              NewID("MST"),
		FailedID:             NewID("MFL"),
		FromSchema:           inspected.Workspace.Schema,
		ToSchema:             runtime.Schema,
		SourceBuild:          inspected.Workspace.BuildID,
		DestinationBuild:     runtime.BuildID,
		SourceCandidate:      inspected.Workspace.CreatedByCandidate,
		DestinationCandidate: runtime.CandidateID,
		Phase:                migrationPrepared,
		StartedAt:            time.Now().UTC(),
	}
	if err = validateMigrationStateStructure(state, inspected.Path); err != nil {
		return WorkspaceSession{}, MigrationReceipt{}, err
	}
	for _, role := range []string{"checkpoint", "stage", "failed"} {
		if err = writeMigrationRole(state, role, inspected.key); err != nil {
			return WorkspaceSession{}, MigrationReceipt{}, fmt.Errorf("create authenticated migration folder identity: %w", err)
		}
	}
	if err = writeMigrationState(&state, inspected.key); err != nil {
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

	if err = secureMigrationRename(state, state.Root, state.Checkpoint, false, "checkpoint", operatingFilesystem); err != nil {
		return failed(fmt.Errorf("the original workspace could not be checkpointed: %w", err))
	}
	nextState := state
	nextState.Phase = migrationOriginalMoved
	if err = writeMigrationState(&nextState, inspected.key); err != nil {
		return failed(fmt.Errorf("the migration checkpoint could not be recorded: %w", err))
	}
	state = nextState
	if err = runMigrationHook(hook, state.Phase); err != nil {
		if errors.Is(err, errMigrationInterrupted) {
			return WorkspaceSession{}, MigrationReceipt{}, err
		}
		return failed(err)
	}

	if err = copyWorkspaceTree(state); err != nil {
		return failed(fmt.Errorf("the migration staging copy could not be created: %w", err))
	}
	if err = migrateStagedWorkspace(state, runtime, hook); err != nil {
		return failed(err)
	}
	nextState = state
	nextState.Phase = migrationStageReady
	if err = writeMigrationState(&nextState, inspected.key); err != nil {
		return failed(fmt.Errorf("the validated migration could not be recorded: %w", err))
	}
	state = nextState
	if err = runMigrationHook(hook, state.Phase); err != nil {
		if errors.Is(err, errMigrationInterrupted) {
			return WorkspaceSession{}, MigrationReceipt{}, err
		}
		return failed(err)
	}

	if err = secureMigrationRename(state, state.Stage, state.Root, true, "", operatingFilesystem); err != nil {
		return failed(fmt.Errorf("the migrated workspace could not be activated: %w", err))
	}
	nextState = state
	nextState.Phase = migrationActivated
	if err = writeMigrationState(&nextState, inspected.key); err != nil {
		return failed(fmt.Errorf("the activated migration could not be recorded: %w", err))
	}
	state = nextState
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
	if err = removeActivatedStageIdentity(state, inspected.key, operatingFilesystem); err != nil {
		v.Close()
		return failed(fmt.Errorf("the activated migration stage identity could not be cleared: %w", err))
	}
	if err = removeMigrationRole(state, "stage", inspected.key, operatingFilesystem); err != nil {
		v.Close()
		return failed(fmt.Errorf("the completed migration stage identity could not be cleared: %w", err))
	}
	if err = removeMigrationRole(state, "failed", inspected.key, operatingFilesystem); err != nil {
		v.Close()
		return failed(fmt.Errorf("the unused migration rollback identity could not be cleared: %w", err))
	}
	if err = removeMigrationMarker(state, inspected.key, operatingFilesystem); err != nil {
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

func migrateStagedWorkspace(state migrationState, runtime RuntimeIdentity, hook func(MigrationPhase) error) error {
	key, err := reloadAuthenticatedMigrationState(state)
	if err != nil {
		return fmt.Errorf("the staged migration state could not be authenticated: %w", err)
	}
	defer zeroBytes(key)
	if err = validateMigrationDirectory(state, state.Stage, true, key); err != nil {
		return fmt.Errorf("the staged migration folder could not be authenticated: %w", err)
	}
	if err = verifyMigrationWorkspace(state.Stage, state, false, key); err != nil {
		return fmt.Errorf("the staged source workspace does not match the checkpoint: %w", err)
	}
	ws, err := readEncryptedWorkspace(state.Stage, key)
	if err != nil {
		return err
	}
	if ws.Schema != state.FromSchema || state.FromSchema != 1 || state.ToSchema != 2 {
		return errors.New("this workspace format does not have an approved recoverable migration path")
	}
	id := state.WorkspaceID
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
	ws.CreatedByCandidate = runtime.CandidateID
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
		Format:             workspaceIdentityFormat,
		ID:                 id,
		Kind:               "development",
		Schema:             runtime.Schema,
		CreatedByCandidate: runtime.CandidateID,
		Name:               name,
		CreatedAt:          createdAt,
		CreatedByBuild:     createdBy,
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
	if err = runMigrationHook(hook, migrationStageUnverified); err != nil {
		return err
	}
	if err = verifyStagedVault(v); err != nil {
		return fmt.Errorf("the migrated workspace failed evidence verification: %w", err)
	}
	pending := map[string]bool{}
	for _, record := range v.Snapshot().Preservations {
		if record.State != preservationCommitted && record.State != preservationFailed {
			pending[record.ID] = true
		}
	}
	if err = v.recoverPreservations(); err != nil {
		return fmt.Errorf("the migrated workspace could not recover verified preservation state: %w", err)
	}
	if len(pending) > 0 {
		recovered := v.Snapshot()
		for _, record := range recovered.Preservations {
			if !pending[record.ID] {
				continue
			}
			if record.State != preservationCommitted {
				return errors.New("the migrated workspace did not commit its freshly verified pending preservation")
			}
			found := false
			for _, item := range recovered.Evidence {
				if item.ID == record.EvidenceID && preservationUsable(item) && item.ObjectFile == record.ObjectFile && item.SHA256 == record.PreservedSHA256 {
					found = true
					break
				}
			}
			if !found {
				return errors.New("the migrated workspace did not bind recovered evidence to its preservation receipt")
			}
		}
	}
	return nil
}

func RecoverWorkspace(root string, runtime RuntimeIdentity) (WorkspaceSession, RecoveryReceipt, error) {
	absolute, err := canonicalWorkspacePath(root)
	if err != nil {
		return WorkspaceSession{}, RecoveryReceipt{}, err
	}
	state, key, err := readMigrationState(absolute)
	if err != nil {
		return WorkspaceSession{}, RecoveryReceipt{}, err
	}
	defer zeroBytes(key)
	if state.DestinationBuild != runtime.BuildID || state.DestinationCandidate != runtime.CandidateID || state.ToSchema != runtime.Schema {
		return WorkspaceSession{}, RecoveryReceipt{}, errors.New("this unfinished migration belongs to another development candidate or build; recovery was blocked")
	}
	receipt := RecoveryReceipt{Path: absolute, Checkpoint: state.Checkpoint}

	if state.Phase == migrationActivated {
		if verifyErr := verifyMigrationWorkspace(absolute, state, true, key); verifyErr == nil {
			v, openErr := openVaultIgnoringRecovery(absolute, runtime)
			if openErr == nil {
				if err = v.recordLifecycle("system", "workspace-recovered", "Recovered and verified an interrupted workspace migration", map[string]any{
					"checkpoint": filepath.Base(state.Checkpoint),
					"phase":      state.Phase,
				}); err != nil {
					v.Close()
					return WorkspaceSession{}, receipt, err
				}
				if err = removeActivatedStageIdentity(state, key, operatingFilesystem); err != nil {
					v.Close()
					return WorkspaceSession{}, receipt, fmt.Errorf("the recovered stage identity could not be cleared: %w", err)
				}
				if err = removeMigrationRole(state, "stage", key, operatingFilesystem); err != nil {
					v.Close()
					return WorkspaceSession{}, receipt, fmt.Errorf("the recovered stage control could not be cleared: %w", err)
				}
				if err = removeMigrationRole(state, "failed", key, operatingFilesystem); err != nil {
					v.Close()
					return WorkspaceSession{}, receipt, fmt.Errorf("the unused rollback control could not be cleared: %w", err)
				}
				if err = removeMigrationMarker(state, key, operatingFilesystem); err != nil {
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
	return rollbackMigrationStateWithOps(state, operatingFilesystem)
}

func rollbackMigrationStateWithOps(state migrationState, ops filesystemOps) (bool, error) {
	key, err := reloadAuthenticatedMigrationState(state)
	if err != nil {
		return false, err
	}
	defer zeroBytes(key)

	rootExists := false
	if _, statErr := os.Lstat(state.Root); statErr == nil {
		if _, err = validateDirectSibling(state.Root, state.Root, filepath.Base(state.Root), true); err != nil {
			return false, err
		}
		rootExists = true
	} else if !os.IsNotExist(statErr) {
		return false, fmt.Errorf("inspect migration target: %w", statErr)
	}
	checkpointExists := false
	if _, statErr := os.Lstat(state.Checkpoint); statErr == nil {
		if err = validateMigrationDirectory(state, state.Checkpoint, true, key); err != nil {
			return false, err
		}
		if err = verifyMigrationWorkspace(state.Checkpoint, state, false, key); err != nil {
			return false, err
		}
		checkpointExists = true
	} else if !os.IsNotExist(statErr) {
		return false, fmt.Errorf("inspect migration checkpoint: %w", statErr)
	}
	if !checkpointExists {
		if !rootExists {
			return false, errors.New("neither the workspace nor its authenticated migration checkpoint is available; recovery was blocked")
		}
		if state.Phase != migrationPrepared {
			return false, errors.New("the authenticated migration checkpoint is unavailable; the recovery record was retained and no workspace was changed")
		}
		if err = verifyMigrationWorkspace(state.Root, state, false, key); err != nil {
			return false, fmt.Errorf("verify unchanged source workspace: %w", err)
		}
		for _, role := range []string{"checkpoint", "stage", "failed"} {
			if err = removeMigrationRole(state, role, key, ops); err != nil {
				return false, fmt.Errorf("clear unused %s migration control: %w", role, err)
			}
		}
		if err = removeMigrationMarker(state, key, ops); err != nil {
			return false, fmt.Errorf("clear unused migration recovery record: %w", err)
		}
		return false, nil
	}

	activeMoved := false
	if rootExists {
		if err = secureMigrationRename(state, state.Root, state.Failed, true, "failed", ops); err != nil {
			return false, fmt.Errorf("move failed migrated workspace aside: %w", err)
		}
		activeMoved = true
	}
	if err = secureMigrationRename(state, state.Checkpoint, state.Root, false, "", ops); err != nil {
		if activeMoved {
			compensateErr := secureMigrationRename(state, state.Failed, state.Root, true, "", ops)
			if compensateErr != nil {
				return false, fmt.Errorf("restore original workspace checkpoint: %v; compensating restoration of the active workspace also failed: %w", err, compensateErr)
			}
		}
		return false, fmt.Errorf("restore original workspace checkpoint: %w; the active workspace was restored to its expected path and the authenticated recovery record was retained for retry", err)
	}

	if err = removeMigrationStageWithOps(state, ops); err != nil {
		return true, fmt.Errorf("the original workspace was restored, but authenticated migration staging cleanup needs attention: %w", err)
	}
	if err = removeMigrationRole(state, "checkpoint", key, ops); err != nil {
		return true, fmt.Errorf("the original workspace was restored, but its obsolete checkpoint control could not be cleared: %w", err)
	}
	if !activeMoved {
		if err = removeMigrationRole(state, "failed", key, ops); err != nil {
			return true, fmt.Errorf("the original workspace was restored, but its unused rollback control could not be cleared: %w", err)
		}
	}
	if err = removeMigrationMarker(state, key, ops); err != nil {
		return true, fmt.Errorf("the original workspace was restored, but its migration recovery record could not be cleared: %w", err)
	}
	return true, nil
}

func writeMigrationState(state *migrationState, key []byte) error {
	if err := validateMigrationStateStructure(*state, state.Root); err != nil {
		return err
	}
	signed := *state
	if err := signMigrationState(&signed, key); err != nil {
		return err
	}
	data, err := json.MarshalIndent(signed, "", "  ")
	if err != nil {
		return err
	}
	path := migrationStatePath(state.Root)
	var previous []byte
	if signed.Phase == migrationPrepared {
		if _, statErr := os.Lstat(path); statErr == nil {
			return errors.New("a migration recovery record already exists; no state was changed")
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
	} else {
		previous, err = readNormalControlFile(path, "migration recovery record")
		if err != nil {
			return fmt.Errorf("verify existing migration recovery state: %w", err)
		}
		var current migrationState
		if err = json.Unmarshal(previous, &current); err != nil {
			return errors.New("the existing migration recovery record is damaged; it was not replaced")
		}
		if err = validateMigrationStateTransition(current, signed, key); err != nil {
			return err
		}
	}
	tmp := path + ".tmp"
	if _, statErr := os.Lstat(tmp); statErr == nil {
		return errors.New("a migration recovery temporary file already exists; no state was changed")
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	if err = writeNewControlFile(tmp, data, "migration recovery temporary file"); err != nil {
		return fmt.Errorf("write migration recovery state: %w", err)
	}
	if signed.Phase == migrationPrepared {
		if _, statErr := os.Lstat(path); statErr == nil {
			removeExpectedControlFile(tmp, data)
			return errors.New("the migration recovery target changed while it was being created; activation was blocked")
		} else if !os.IsNotExist(statErr) {
			removeExpectedControlFile(tmp, data)
			return statErr
		}
	} else {
		current, readErr := readNormalControlFile(path, "migration recovery record")
		if readErr != nil || !bytes.Equal(current, previous) {
			removeExpectedControlFile(tmp, data)
			return errors.New("the migration recovery record changed while it was being updated; activation was blocked")
		}
	}
	if err = os.Rename(tmp, path); err != nil {
		removeExpectedControlFile(tmp, data)
		return fmt.Errorf("activate migration recovery state: %w", err)
	}
	*state = signed
	return nil
}

func removeExpectedControlFile(path string, expected []byte) {
	current, err := readNormalControlFile(path, "migration recovery temporary file")
	if err != nil || !bytes.Equal(current, expected) {
		return
	}
	_ = os.Remove(path)
}

func validateMigrationStateTransition(current, next migrationState, key []byte) error {
	if err := validateMigrationStateStructure(current, next.Root); err != nil {
		return err
	}
	if err := authenticateMigrationState(current, key); err != nil {
		return err
	}
	expectedPhase := MigrationPhase("")
	switch next.Phase {
	case migrationOriginalMoved:
		expectedPhase = migrationPrepared
	case migrationStageReady:
		expectedPhase = migrationOriginalMoved
	case migrationActivated:
		expectedPhase = migrationStageReady
	default:
		return errors.New("the requested migration recovery transition is invalid")
	}
	if current.Phase != expectedPhase {
		return errors.New("the migration recovery phase changed unexpectedly; the record was not replaced")
	}
	current.AuthTag = ""
	next.AuthTag = ""
	current.Phase = next.Phase
	if !sameMigrationState(current, next) {
		return errors.New("the migration recovery identity changed unexpectedly; the record was not replaced")
	}
	return nil
}

func readMigrationState(root string) (migrationState, []byte, error) {
	data, err := readNormalControlFile(migrationStatePath(root), "migration recovery record")
	if err != nil {
		if os.IsNotExist(err) {
			return migrationState{}, nil, errors.New("this workspace has no unfinished migration to recover")
		}
		return migrationState{}, nil, err
	}
	var state migrationState
	if err = json.Unmarshal(data, &state); err != nil {
		return migrationState{}, nil, errors.New("the migration recovery record is damaged; automatic recovery was blocked")
	}
	if err = validateMigrationStateStructure(state, root); err != nil {
		return migrationState{}, nil, err
	}
	key, err := candidateMigrationKey(state)
	if err != nil {
		return migrationState{}, nil, err
	}
	for _, path := range []string{state.Checkpoint, state.Stage, state.Failed} {
		if _, statErr := os.Lstat(path); statErr == nil {
			if validateErr := validateMigrationDirectory(state, path, true, key); validateErr != nil {
				zeroBytes(key)
				return migrationState{}, nil, fmt.Errorf("an authenticated migration path failed containment checks: %w", validateErr)
			}
		} else if !os.IsNotExist(statErr) {
			zeroBytes(key)
			return migrationState{}, nil, statErr
		}
	}
	return state, key, nil
}

func removeMigrationStage(state migrationState) error {
	return removeMigrationStageWithOps(state, operatingFilesystem)
}

func removeMigrationStageWithOps(state migrationState, ops filesystemOps) error {
	key, err := reloadAuthenticatedMigrationState(state)
	if err != nil {
		return err
	}
	defer zeroBytes(key)
	if _, statErr := os.Lstat(state.Stage); os.IsNotExist(statErr) {
		return removeMigrationRole(state, "stage", key, ops)
	} else if statErr != nil {
		return statErr
	}
	if err = validateMigrationDirectory(state, state.Stage, true, key); err != nil {
		return err
	}
	if err = removeNormalTree(state.Stage, ops.remove); err != nil {
		return fmt.Errorf("remove authenticated migration staging folder: %w", err)
	}
	return removeMigrationRole(state, "stage", key, ops)
}

func copyWorkspaceTree(state migrationState) error {
	source := state.Checkpoint
	destination := state.Stage
	key, err := reloadAuthenticatedMigrationState(state)
	if err != nil {
		return err
	}
	defer zeroBytes(key)
	if err = validateMigrationDirectory(state, source, true, key); err != nil {
		return err
	}
	if err = verifyMigrationWorkspace(source, state, false, key); err != nil {
		return err
	}
	if err = validateMigrationDirectory(state, destination, false, key); err != nil {
		return err
	}
	sourceInfo, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if !sourceInfo.IsDir() || sourceInfo.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(sourceInfo) {
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
	if err = writeStageDirectoryIdentity(state, key); err != nil {
		return err
	}
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
		if fileInfoHasReparsePoint(info) {
			return errors.New("workspace migration will not follow junctions or reparse points")
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
	created, err := output.Stat()
	if err != nil {
		_ = output.Close()
		return err
	}
	ok := false
	defer func() {
		_ = output.Close()
		if !ok {
			removeCreatedNormalFile(destination, created)
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
