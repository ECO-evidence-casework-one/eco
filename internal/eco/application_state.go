package eco

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"
)

const candidateStateSchema = 2

type CandidateEvent struct {
	ID          string    `json:"id"`
	At          time.Time `json:"at"`
	Action      string    `json:"action"`
	WorkspaceID string    `json:"workspace_id,omitempty"`
	Outcome     string    `json:"outcome"`
	Summary     string    `json:"summary"`
	PrevHash    string    `json:"prev_hash,omitempty"`
	Hash        string    `json:"hash"`
}

type CandidateState struct {
	Schema            int              `json:"schema"`
	CandidateID       string           `json:"candidate_id"`
	BuildID           string           `json:"build_id"`
	StateRoot         string           `json:"-"`
	DefaultWorkspace  string           `json:"-"`
	SelectedWorkspace string           `json:"-"`
	Events            []CandidateEvent `json:"events"`
}

type CandidateApplication struct {
	Runtime RuntimeIdentity
	State   CandidateState
	Current WorkspaceSession
	mu      sync.Mutex
}

type CandidateStateWarning struct {
	Action string
	Err    error
}

func (e *CandidateStateWarning) Error() string {
	return fmt.Sprintf("%s completed, but ECO could not save its candidate-specific selection audit: %v", e.Action, e.Err)
}

func (e *CandidateStateWarning) Unwrap() error {
	return e.Err
}

func StartCandidate(baseStateRoot string, runtime RuntimeIdentity) (*CandidateApplication, error) {
	if err := validateRuntime(runtime); err != nil {
		return nil, err
	}
	base, err := normaliseStateRoot(baseStateRoot)
	if err != nil {
		return nil, err
	}
	stateRoot := filepath.Join(base, "candidates", candidateDirectoryName(runtime.CandidateID))
	defaultWorkspace := filepath.Join(stateRoot, "workspaces", "development")
	statePath := filepath.Join(stateRoot, "app-state.json")
	if err = os.MkdirAll(stateRoot, 0700); err != nil {
		return nil, fmt.Errorf("create candidate-specific application state: %w", err)
	}

	state, stateErr := readCandidateState(statePath)
	if stateErr != nil && !os.IsNotExist(stateErr) {
		return nil, stateErr
	}
	app := &CandidateApplication{Runtime: runtime}
	if os.IsNotExist(stateErr) {
		session, recovered, err := startFirstCandidateWorkspace(stateRoot, defaultWorkspace, runtime)
		if err != nil {
			return nil, err
		}
		app.State = CandidateState{
			Schema:            candidateStateSchema,
			CandidateID:       runtime.CandidateID,
			BuildID:           runtime.BuildID,
			StateRoot:         stateRoot,
			DefaultWorkspace:  defaultWorkspace,
			SelectedWorkspace: defaultWorkspace,
			Events:            []CandidateEvent{},
		}
		action := "first-launch"
		summary := "Started this development candidate with a new empty workspace."
		if recovered {
			action = "first-launch-recovery"
			summary = "Recovered this candidate's own completed workspace after application-state creation was interrupted."
		}
		app.addEvent(action, session, "success", summary)
		if err = saveCandidateState(statePath, app.State); err != nil {
			session.Vault.Close()
			return nil, err
		}
		app.Current = session
		return app, nil
	}

	if err = validateCandidateState(state, runtime); err != nil {
		return nil, err
	}
	app.State = state
	app.State.StateRoot = stateRoot
	app.State.DefaultWorkspace = defaultWorkspace
	app.State.SelectedWorkspace = defaultWorkspace
	app.State.BuildID = runtime.BuildID

	if _, err = os.Stat(migrationStatePath(defaultWorkspace)); err == nil {
		session, receipt, recoverErr := RecoverWorkspace(defaultWorkspace, runtime)
		if recoverErr != nil {
			return nil, fmt.Errorf("%s %w", receipt.Message, recoverErr)
		}
		session.Explicit = false
		app.addEvent("workspace-recovered", session, "success", receipt.Message)
		if err = saveCandidateState(statePath, app.State); err != nil {
			session.Vault.Close()
			return nil, err
		}
		app.Current = session
		return app, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect candidate workspace recovery state: %w", err)
	}
	if _, err = os.Stat(restoreStatePath(defaultWorkspace)); err == nil {
		session, receipt, recoverErr := RecoverPortableRestore(defaultWorkspace, runtime)
		if recoverErr != nil {
			return nil, fmt.Errorf("%s %w", receipt.Message, recoverErr)
		}
		session.Explicit = false
		app.addEvent("workspace-recovered", session, "success", receipt.Message)
		if err = saveCandidateState(statePath, app.State); err != nil {
			session.Vault.Close()
			return nil, err
		}
		app.Current = session
		return app, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect candidate portable restore recovery state: %w", err)
	}

	inspected, err := inspectWorkspace(defaultWorkspace, runtime)
	if err != nil {
		return nil, fmt.Errorf("this candidate's development workspace could not be reopened safely: %w", err)
	}
	if inspected.Workspace.CreatedByCandidate != runtime.CandidateID || inspected.Identity.CreatedByCandidate != runtime.CandidateID || inspected.Workspace.BuildID != runtime.BuildID {
		inspected.Close()
		return nil, errors.New("this candidate's automatic workspace belongs to another development candidate; select it deliberately instead")
	}
	v, err := openInspectedVault(inspected, runtime)
	if err != nil {
		return nil, fmt.Errorf("this candidate's development workspace could not be reopened safely: %w", err)
	}
	if err = v.recordLifecycle("system", "workspace-reopened", "Reopened this candidate's own development workspace", map[string]any{
		"build":    runtime.BuildID,
		"explicit": false,
	}); err != nil {
		v.Close()
		return nil, err
	}
	session := sessionFor(v, DispositionReopened, false, "")
	app.addEvent("candidate-restart", session, "success", "Reopened only this development candidate's own workspace.")
	if err = saveCandidateState(statePath, app.State); err != nil {
		v.Close()
		return nil, err
	}
	app.Current = session
	return app, nil
}

func startFirstCandidateWorkspace(stateRoot, defaultWorkspace string, runtime RuntimeIdentity) (WorkspaceSession, bool, error) {
	if info, err := os.Stat(defaultWorkspace); err == nil {
		if !info.IsDir() {
			return WorkspaceSession{}, false, errors.New("candidate startup was blocked because its development workspace path is not a folder")
		}
		inspected, inspectErr := inspectWorkspace(defaultWorkspace, runtime)
		if inspectErr != nil {
			return WorkspaceSession{}, false, errors.New("candidate application state is incomplete and the existing development workspace could not be verified; nothing was opened")
		}
		if inspected.Identity.CreatedByCandidate != runtime.CandidateID || inspected.Workspace.CreatedByCandidate != runtime.CandidateID || inspected.Workspace.BuildID != runtime.BuildID {
			inspected.Close()
			return WorkspaceSession{}, false, errors.New("candidate application state is incomplete and the existing workspace belongs to another development candidate; nothing was opened automatically")
		}
		v, openErr := openInspectedVault(inspected, runtime)
		if openErr != nil {
			return WorkspaceSession{}, false, errors.New("candidate application state is incomplete and the existing development workspace could not be verified; nothing was opened")
		}
		if recordErr := v.recordLifecycle("system", "workspace-recovered", "Recovered candidate application state without importing another candidate's data", map[string]any{
			"build": runtime.BuildID,
		}); recordErr != nil {
			v.Close()
			return WorkspaceSession{}, false, recordErr
		}
		return sessionFor(v, DispositionRecovered, false, ""), true, nil
	} else if !os.IsNotExist(err) {
		return WorkspaceSession{}, false, err
	}

	entries, err := os.ReadDir(stateRoot)
	if err != nil {
		return WorkspaceSession{}, false, err
	}
	for _, entry := range entries {
		if entry.Name() != "workspaces" {
			return WorkspaceSession{}, false, errors.New("candidate application state is incomplete and not empty; startup was blocked instead of guessing")
		}
	}
	v, err := createVault(defaultWorkspace, "Development workspace for "+runtime.BuildID, runtime)
	if err != nil {
		return WorkspaceSession{}, false, err
	}
	return sessionFor(v, DispositionNew, false, ""), false, nil
}

func (a *CandidateApplication) CreateWorkspace(path, name string) (WorkspaceSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	v, err := createVault(path, name, a.Runtime)
	if err != nil {
		a.addCandidateEvent("workspace-created", "", "blocked", err.Error())
		_ = saveCandidateState(a.statePath(), a.State)
		return WorkspaceSession{}, err
	}
	session := sessionFor(v, DispositionNew, true, "")
	if err = a.selectSession("workspace-created", session, "Created and selected a new empty ECO development workspace."); err != nil {
		return session, err
	}
	return session, nil
}

func (a *CandidateApplication) OpenWorkspace(path string) (WorkspaceSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	v, err := openVault(path, a.Runtime)
	if err != nil {
		a.addCandidateEvent("workspace-reopened", "", "blocked", err.Error())
		_ = saveCandidateState(a.statePath(), a.State)
		return WorkspaceSession{}, err
	}
	openCompatibility := compatibilityFor(v.Workspace, a.Runtime)
	if err = v.recordLifecycle("user", "workspace-reopened", "Deliberately reopened an existing compatible ECO workspace", map[string]any{
		"build":    a.Runtime.BuildID,
		"explicit": true,
	}); err != nil {
		v.Close()
		return WorkspaceSession{}, err
	}
	session := sessionFor(v, DispositionReopened, true, "")
	session.Compatibility = openCompatibility
	if err = a.selectSession("workspace-reopened", session, "Deliberately selected and reopened this compatible workspace."); err != nil {
		return session, err
	}
	return session, nil
}

func (a *CandidateApplication) MigrateWorkspace(path string) (WorkspaceSession, MigrationReceipt, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	session, receipt, err := MigrateWorkspace(path, a.Runtime)
	if err != nil {
		a.addCandidateEvent("workspace-migration", "", "blocked", err.Error())
		_ = saveCandidateState(a.statePath(), a.State)
		return WorkspaceSession{}, receipt, err
	}
	if err = a.selectSession("workspace-migrated", session, "Migrated and selected the workspace; the original checkpoint was kept."); err != nil {
		return session, receipt, err
	}
	return session, receipt, nil
}

func (a *CandidateApplication) RecoverWorkspace(path string) (WorkspaceSession, RecoveryReceipt, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	absolute, pathErr := normaliseWorkspaceRoot(path)
	if pathErr != nil {
		return WorkspaceSession{}, RecoveryReceipt{}, pathErr
	}
	var session WorkspaceSession
	var receipt RecoveryReceipt
	var err error
	if _, restoreErr := os.Lstat(restoreStatePath(absolute)); restoreErr == nil {
		session, receipt, err = RecoverPortableRestore(absolute, a.Runtime)
	} else if !os.IsNotExist(restoreErr) {
		err = fmt.Errorf("inspect portable restore recovery state: %w", restoreErr)
	} else {
		session, receipt, err = RecoverWorkspace(absolute, a.Runtime)
	}
	if err != nil {
		a.addCandidateEvent("workspace-recovery", "", "attention", receipt.Message+" "+err.Error())
		_ = saveCandidateState(a.statePath(), a.State)
		return WorkspaceSession{}, receipt, err
	}
	if err = a.selectSession("workspace-recovered", session, receipt.Message); err != nil {
		return session, receipt, err
	}
	return session, receipt, nil
}

func (a *CandidateApplication) ResetCurrentWorkspace() (WorkspaceSession, ResetReceipt, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	receipt, err := resetVault(a.Current.Vault)
	if err != nil {
		outcome := "blocked"
		if receipt.WorkspaceID != "" {
			outcome = "attention"
		}
		a.addCandidateEvent("workspace-reset", a.Current.Identity.ID, outcome, err.Error())
		_ = saveCandidateState(a.statePath(), a.State)
		return WorkspaceSession{}, receipt, err
	}
	session := sessionFor(a.Current.Vault, DispositionReset, true, "")
	if err = a.selectSession("workspace-reset", session, "Reset only the selected ECO development workspace."); err != nil {
		return session, receipt, err
	}
	return session, receipt, nil
}

func (a *CandidateApplication) RefreshCurrentAfterRestore() (WorkspaceSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.Current.Vault == nil {
		return WorkspaceSession{}, errors.New("no selected workspace is available after restore")
	}
	session := sessionFor(a.Current.Vault, DispositionRecovered, true, "")
	if err := a.selectSession("portable-backup-restored", session, "Selected the authenticated workspace restored from an encrypted backup."); err != nil {
		return session, err
	}
	return session, nil
}

func (a *CandidateApplication) selectSession(action string, session WorkspaceSession, summary string) error {
	old := a.Current
	previousState := a.State
	a.State.SelectedWorkspace = session.Path
	a.addEvent(action, session, "success", summary)
	if err := saveCandidateState(a.statePath(), a.State); err != nil {
		a.State = previousState
		a.Current = session
		if old.Vault != nil && old.Vault != session.Vault {
			old.Vault.Close()
		}
		return &CandidateStateWarning{Action: action, Err: err}
	}
	a.Current = session
	if old.Vault != nil && old.Vault != session.Vault {
		old.Vault.Close()
	}
	return nil
}

func (a *CandidateApplication) statePath() string {
	return filepath.Join(a.State.StateRoot, "app-state.json")
}

func (a *CandidateApplication) addEvent(action string, session WorkspaceSession, outcome, summary string) {
	a.addCandidateEvent(action, session.Identity.ID, outcome, summary)
}

func (a *CandidateApplication) addCandidateEvent(action, workspaceID, outcome, summary string) {
	previous := ""
	if len(a.State.Events) > 0 {
		previous = a.State.Events[0].Hash
	}
	event := CandidateEvent{
		ID:          NewID("APP"),
		At:          time.Now().UTC(),
		Action:      action,
		WorkspaceID: workspaceID,
		Outcome:     outcome,
		Summary:     candidateAuditSummary(action, outcome, summary),
		PrevHash:    previous,
	}
	event.Hash = candidateEventHash(event)
	a.State.Events = append([]CandidateEvent{event}, a.State.Events...)
}

func candidateAuditSummary(action, outcome, successSummary string) string {
	if outcome == "success" {
		return successSummary
	}
	switch action {
	case "workspace-created":
		return "Workspace creation was blocked; the selected workspace did not change."
	case "workspace-reopened":
		return "Workspace reopening was blocked; nothing was opened by guesswork."
	case "workspace-migration":
		return "Workspace migration was blocked; unverified state was not selected."
	case "workspace-recovery":
		return "Workspace recovery needs attention; unauthenticated state was not selected."
	case "workspace-reset":
		return "Workspace reset needs attention; unsafe object cleanup was blocked."
	default:
		return "The workspace action needs attention; no workspace was selected by guesswork."
	}
}

func candidateEventHash(event CandidateEvent) string {
	value := struct {
		ID, Action, WorkspaceID, Outcome, Summary, PrevHash string
		At                                                  time.Time
	}{
		ID:          event.ID,
		Action:      event.Action,
		WorkspaceID: event.WorkspaceID,
		Outcome:     event.Outcome,
		Summary:     event.Summary,
		PrevHash:    event.PrevHash,
		At:          event.At,
	}
	data, _ := json.Marshal(value)
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func normaliseStateRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", errors.New("the application state folder is unavailable")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolute), nil
}

func candidateDirectoryName(candidateID string) string {
	var name strings.Builder
	for _, r := range strings.ToLower(candidateID) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			name.WriteRune(r)
		} else if name.Len() > 0 && !strings.HasSuffix(name.String(), "-") {
			name.WriteByte('-')
		}
		if name.Len() >= 32 {
			break
		}
	}
	prefix := strings.Trim(name.String(), "-")
	if prefix == "" {
		prefix = "candidate"
	}
	digest := sha256.Sum256([]byte(candidateID))
	return prefix + "-" + hex.EncodeToString(digest[:6])
}

func readCandidateState(path string) (CandidateState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return CandidateState{}, err
	}
	var state CandidateState
	if err = json.Unmarshal(data, &state); err != nil {
		return CandidateState{}, errors.New("this candidate's application state is damaged; startup was blocked instead of opening a workspace by guesswork")
	}
	return state, nil
}

func validateCandidateState(state CandidateState, runtime RuntimeIdentity) error {
	if state.Schema != candidateStateSchema {
		return errors.New("this candidate's application state has an unsupported format; startup was blocked")
	}
	if state.CandidateID != runtime.CandidateID {
		return errors.New("this application state belongs to another development candidate; startup was blocked")
	}
	if state.BuildID != runtime.BuildID {
		return errors.New("this candidate's application state names another build; startup was blocked")
	}
	previous := ""
	for i := len(state.Events) - 1; i >= 0; i-- {
		event := state.Events[i]
		if event.PrevHash != previous || candidateEventHash(event) != event.Hash {
			return errors.New("this candidate's application audit record is inconsistent; startup was blocked")
		}
		previous = event.Hash
	}
	return nil
}

func saveCandidateState(path string, state CandidateState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err = os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("save candidate-specific application state: %w", err)
	}
	if err = os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("activate candidate-specific application state: %w", err)
	}
	return nil
}
