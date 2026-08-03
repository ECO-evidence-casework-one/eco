package eco

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
	"unicode"
)

const (
	workspaceIdentityFormat = "ECO-WORKSPACE-1"
	workspaceIdentityFile   = "workspace.identity.json"
)

type RuntimeIdentity struct {
	CandidateID string `json:"candidate_id"`
	BuildID     string `json:"build_id"`
	Schema      int    `json:"schema"`
}

func CurrentRuntime() RuntimeIdentity {
	revision, modified := currentSourceRevision()
	artifactSHA256, err := currentArtifactSHA256()
	if err != nil {
		return RuntimeIdentity{BuildID: BuildID, Schema: Schema}
	}
	return RuntimeIdentity{CandidateID: candidateIDForSource(BuildID, revision, modified, artifactSHA256), BuildID: BuildID, Schema: Schema}
}

func currentSourceRevision() (string, bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false
	}
	revision := ""
	modified := false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	return revision, modified
}

func currentArtifactSHA256() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate this ECO build: %w", err)
	}
	f, err := os.Open(executable)
	if err != nil {
		return "", fmt.Errorf("open this ECO build: %w", err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", fmt.Errorf("fingerprint this ECO build: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func candidateIDForSource(buildID, revision string, modified bool, artifactSHA256 string) string {
	revision = strings.TrimSpace(revision)
	if revision == "" {
		revision = "unrecorded"
	}
	artifactSHA256 = strings.TrimSpace(artifactSHA256)
	if artifactSHA256 == "" {
		return ""
	}
	id := buildID + "-source-" + revision
	if modified {
		id += "-modified"
	}
	return id + "-artifact-" + artifactSHA256
}

type WorkspaceIdentity struct {
	Format             string    `json:"format"`
	ID                 string    `json:"id"`
	Kind               string    `json:"kind"`
	Schema             int       `json:"schema"`
	CreatedByCandidate string    `json:"created_by_candidate"`
	Name               string    `json:"-"`
	CreatedAt          time.Time `json:"-"`
	CreatedByBuild     string    `json:"-"`
}

type CompatibilityStatus string

const (
	CompatibilityCurrent          CompatibilityStatus = "compatible"
	CompatibilityCompatibleBuild  CompatibilityStatus = "compatible-build"
	CompatibilityMigrationNeeded  CompatibilityStatus = "migration-required"
	CompatibilityTooOld           CompatibilityStatus = "unsupported-older-format"
	CompatibilityDowngradeBlocked CompatibilityStatus = "newer-format-blocked"
	CompatibilityRecoveryNeeded   CompatibilityStatus = "recovery-required"
)

type WorkspaceCompatibility struct {
	Status             CompatibilityStatus `json:"status"`
	Message            string              `json:"message"`
	WorkspaceSchema    int                 `json:"workspace_schema"`
	BuildSchema        int                 `json:"build_schema"`
	WorkspaceBuild     string              `json:"workspace_build"`
	CurrentBuild       string              `json:"current_build"`
	WorkspaceCandidate string              `json:"workspace_candidate"`
	CurrentCandidate   string              `json:"current_candidate"`
	CanOpen            bool                `json:"can_open"`
	CanMigrate         bool                `json:"can_migrate"`
}

type CompatibilityError struct {
	Report WorkspaceCompatibility
}

func (e *CompatibilityError) Error() string {
	return e.Report.Message
}

type RecoveryRequiredError struct {
	Path string
}

func (e *RecoveryRequiredError) Error() string {
	return "This workspace has an unfinished upgrade. Choose Recover workspace before opening it."
}

type WorkspaceDisposition string

const (
	DispositionNew       WorkspaceDisposition = "Newly created"
	DispositionReopened  WorkspaceDisposition = "Reopened"
	DispositionMigrated  WorkspaceDisposition = "Migrated"
	DispositionRecovered WorkspaceDisposition = "Recovered"
	DispositionReset     WorkspaceDisposition = "Reset"
)

type WorkspaceSession struct {
	Vault         *Vault
	Identity      WorkspaceIdentity
	Path          string
	Disposition   WorkspaceDisposition
	Compatibility WorkspaceCompatibility
	Explicit      bool
	Checkpoint    string
}

func (s WorkspaceSession) StatusText() string {
	return fmt.Sprintf("%s. %s", s.Disposition, s.Compatibility.Message)
}

func validateRuntime(runtime RuntimeIdentity) error {
	if !validIdentityLabel(runtime.CandidateID, 256) {
		return errors.New("ECO could not confirm this development candidate's identity, so no workspace was opened")
	}
	if !validIdentityLabel(runtime.BuildID, 128) {
		return errors.New("the build identity is missing")
	}
	if runtime.Schema < 1 {
		return errors.New("the build workspace format is invalid")
	}
	return nil
}

func normaliseWorkspaceRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", errors.New("choose an ECO workspace folder")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve workspace folder: %w", err)
	}
	absolute = filepath.Clean(absolute)
	volume := filepath.VolumeName(absolute)
	if absolute == string(filepath.Separator) || (volume != "" && strings.EqualFold(absolute, volume+string(filepath.Separator))) {
		return "", errors.New("a drive root cannot be used as an ECO workspace")
	}
	return absolute, nil
}

func newWorkspaceForRuntime(runtime RuntimeIdentity, id, name string) Workspace {
	now := time.Now().UTC()
	return Workspace{
		Schema:             runtime.Schema,
		BuildID:            runtime.BuildID,
		WorkspaceID:        id,
		WorkspaceName:      name,
		CreatedByBuild:     runtime.BuildID,
		CreatedByCandidate: runtime.CandidateID,
		CreatedAt:          now,
		UpdatedAt:          now,
		Evidence:           []EvidenceItem{},
		Preservations:      []PreservationRecord{},
		Matters:            []Matter{},
		Changes:            []ChangeRecord{},
		Questions:          []QuestionRecord{},
		SelectedPage:       "home",
	}
}

func createVault(root, name string, runtime RuntimeIdentity) (*Vault, error) {
	if err := validateRuntime(runtime); err != nil {
		return nil, err
	}
	absolute, err := normaliseWorkspaceRoot(root)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "ECO development workspace"
	}
	if !validWorkspaceName(name) {
		return nil, errors.New("workspace name must be 1 to 120 visible characters")
	}

	rootCreated := false
	info, statErr := os.Stat(absolute)
	switch {
	case os.IsNotExist(statErr):
		if err = os.MkdirAll(absolute, 0700); err != nil {
			return nil, fmt.Errorf("create workspace folder: %w", err)
		}
		rootCreated = true
	case statErr != nil:
		return nil, fmt.Errorf("inspect workspace folder: %w", statErr)
	case !info.IsDir():
		return nil, errors.New("the selected workspace path is not a folder")
	default:
		linkInfo, linkErr := os.Lstat(absolute)
		if linkErr != nil || linkInfo.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(linkInfo) {
			return nil, errors.New("a new workspace cannot be created through a symbolic link, junction or reparse point")
		}
		entries, readErr := os.ReadDir(absolute)
		if readErr != nil {
			return nil, fmt.Errorf("inspect workspace folder: %w", readErr)
		}
		if len(entries) != 0 {
			return nil, errors.New("a new workspace needs an empty folder; no existing files were changed")
		}
	}

	cleanup := func() {
		_ = os.Remove(filepath.Join(absolute, "workspace.ecodb.tmp"))
		_ = os.Remove(filepath.Join(absolute, "workspace.ecodb"))
		_ = os.Remove(filepath.Join(absolute, workspaceIdentityFile+".tmp"))
		_ = os.Remove(filepath.Join(absolute, workspaceIdentityFile))
		_ = os.Remove(filepath.Join(absolute, "vault.key.tmp"))
		_ = os.Remove(filepath.Join(absolute, "vault.key"))
		_ = os.Remove(filepath.Join(absolute, "objects"))
		if rootCreated {
			_ = os.Remove(absolute)
		}
	}

	objects := filepath.Join(absolute, "objects")
	if err = os.Mkdir(objects, 0700); err != nil {
		cleanup()
		return nil, fmt.Errorf("create encrypted object folder: %w", err)
	}
	key, err := loadOrCreateMasterKey(filepath.Join(absolute, "vault.key"))
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("create workspace key: %w", err)
	}
	id := NewID("WS")
	identity := WorkspaceIdentity{
		Format:             workspaceIdentityFormat,
		ID:                 id,
		Name:               name,
		Kind:               "development",
		Schema:             runtime.Schema,
		CreatedAt:          time.Now().UTC(),
		CreatedByBuild:     runtime.BuildID,
		CreatedByCandidate: runtime.CandidateID,
	}
	v := &Vault{
		Root:      absolute,
		Objects:   objects,
		Identity:  identity,
		key:       key,
		runtime:   runtime,
		Workspace: newWorkspaceForRuntime(runtime, id, name),
	}
	v.Workspace.CreatedAt = identity.CreatedAt
	v.addChangeUnlocked("system", "workspace-created", "Created a new empty ECO development workspace", map[string]any{
		"workspace_id": id,
		"build":        runtime.BuildID,
		"schema":       runtime.Schema,
	})
	if err = v.Save(); err != nil {
		zeroBytes(key)
		cleanup()
		return nil, fmt.Errorf("save new workspace: %w", err)
	}
	if err = writeWorkspaceIdentity(absolute, identity); err != nil {
		zeroBytes(key)
		cleanup()
		return nil, err
	}
	return v, nil
}

func openVault(root string, runtime RuntimeIdentity) (*Vault, error) {
	inspected, err := inspectWorkspace(root, runtime)
	if err != nil {
		return nil, err
	}
	return openInspectedVault(inspected, runtime)
}

func openVaultIgnoringRecovery(root string, runtime RuntimeIdentity) (*Vault, error) {
	inspected, err := inspectWorkspaceAt(root, runtime, false)
	if err != nil {
		return nil, err
	}
	return openInspectedVault(inspected, runtime)
}

func openInspectedVault(inspected inspectedWorkspace, runtime RuntimeIdentity) (*Vault, error) {
	if !inspected.Compatibility.CanOpen {
		zeroBytes(inspected.key)
		return nil, &CompatibilityError{Report: inspected.Compatibility}
	}
	objects := filepath.Join(inspected.Path, "objects")
	objectsCanonical, err := validateObjectsDirectory(inspected.Path, objects)
	if err != nil {
		zeroBytes(inspected.key)
		return nil, fmt.Errorf("workspace object folder is unavailable or unsafe: %w", err)
	}
	v := &Vault{
		Root:      inspected.Path,
		Objects:   objectsCanonical,
		Identity:  inspected.Identity,
		key:       inspected.key,
		runtime:   runtime,
		Workspace: inspected.Workspace,
	}
	if err = cleanupInterruptedReadingCopies(v.Root); err != nil {
		zeroBytes(v.key)
		return nil, fmt.Errorf("clean interrupted derived reading state: %w", err)
	}
	if err = v.recoverPreservations(); err != nil {
		zeroBytes(v.key)
		return nil, fmt.Errorf("recover preservation state: %w", err)
	}
	return v, nil
}

type inspectedWorkspace struct {
	Path          string
	Identity      WorkspaceIdentity
	Workspace     Workspace
	Compatibility WorkspaceCompatibility
	key           []byte
}

func InspectWorkspace(root string, runtime RuntimeIdentity) (WorkspaceIdentity, WorkspaceCompatibility, error) {
	inspected, err := inspectWorkspace(root, runtime)
	if err != nil {
		return WorkspaceIdentity{}, WorkspaceCompatibility{}, err
	}
	zeroBytes(inspected.key)
	return inspected.Identity, inspected.Compatibility, nil
}

func inspectWorkspace(root string, runtime RuntimeIdentity) (inspectedWorkspace, error) {
	return inspectWorkspaceAt(root, runtime, true)
}

func inspectWorkspaceAt(root string, runtime RuntimeIdentity, checkRecovery bool) (inspectedWorkspace, error) {
	if err := validateRuntime(runtime); err != nil {
		return inspectedWorkspace{}, err
	}
	absolute, err := normaliseWorkspaceRoot(root)
	if err != nil {
		return inspectedWorkspace{}, err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		if os.IsNotExist(err) {
			return inspectedWorkspace{}, errors.New("the selected ECO workspace does not exist")
		}
		return inspectedWorkspace{}, fmt.Errorf("inspect workspace folder: %w", err)
	}
	if !info.IsDir() {
		return inspectedWorkspace{}, errors.New("the selected ECO workspace is not a folder")
	}
	linkInfo, linkErr := os.Lstat(absolute)
	if linkErr != nil || linkInfo.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(linkInfo) {
		return inspectedWorkspace{}, errors.New("the selected ECO workspace cannot be opened through a symbolic link, junction or reparse point")
	}
	absolute, err = canonicalNormalDirectory(absolute, "the selected ECO workspace")
	if err != nil {
		return inspectedWorkspace{}, fmt.Errorf("resolve the selected ECO workspace safely: %w", err)
	}
	if checkRecovery {
		if _, err = os.Lstat(migrationStatePath(absolute)); err == nil {
			return inspectedWorkspace{}, &RecoveryRequiredError{Path: absolute}
		} else if !os.IsNotExist(err) {
			return inspectedWorkspace{}, fmt.Errorf("check workspace recovery state: %w", err)
		}
	}
	key, err := loadExistingMasterKey(filepath.Join(absolute, "vault.key"))
	if err != nil {
		return inspectedWorkspace{}, fmt.Errorf("this folder is not a readable ECO workspace: %w", err)
	}
	ws, err := readEncryptedWorkspace(absolute, key)
	if err != nil {
		zeroBytes(key)
		return inspectedWorkspace{}, err
	}
	identity, identityErr := readWorkspaceIdentity(absolute)
	if identityErr != nil {
		if !(os.IsNotExist(identityErr) && ws.Schema == 1) {
			zeroBytes(key)
			return inspectedWorkspace{}, identityErr
		}
		identity = WorkspaceIdentity{
			Format:             workspaceIdentityFormat,
			ID:                 ws.WorkspaceID,
			Kind:               "development",
			Schema:             ws.Schema,
			CreatedByCandidate: ws.CreatedByCandidate,
		}
	}
	if err = validateWorkspaceIdentity(identity, ws); err != nil {
		zeroBytes(key)
		return inspectedWorkspace{}, err
	}
	identity.Name = ws.WorkspaceName
	identity.CreatedAt = ws.CreatedAt
	identity.CreatedByBuild = ws.CreatedByBuild
	report := compatibilityFor(ws, runtime)
	return inspectedWorkspace{
		Path:          absolute,
		Identity:      identity,
		Workspace:     ws,
		Compatibility: report,
		key:           key,
	}, nil
}

func compatibilityFor(ws Workspace, runtime RuntimeIdentity) WorkspaceCompatibility {
	report := WorkspaceCompatibility{
		WorkspaceSchema:    ws.Schema,
		BuildSchema:        runtime.Schema,
		WorkspaceBuild:     ws.BuildID,
		CurrentBuild:       runtime.BuildID,
		WorkspaceCandidate: ws.CreatedByCandidate,
		CurrentCandidate:   runtime.CandidateID,
	}
	switch {
	case ws.Schema == runtime.Schema && ws.BuildID == runtime.BuildID && ws.CreatedByCandidate == runtime.CandidateID:
		report.Status = CompatibilityCurrent
		report.Message = fmt.Sprintf("Compatible with this build (workspace format %d).", runtime.Schema)
		report.CanOpen = true
	case ws.Schema == runtime.Schema:
		report.Status = CompatibilityCompatibleBuild
		report.Message = fmt.Sprintf("Compatible external workspace from build %s. It was created by another development candidate and is open only because you selected it deliberately.", layBuild(ws.BuildID))
		report.CanOpen = true
	case ws.Schema == 1 && runtime.Schema == 2:
		report.Status = CompatibilityMigrationNeeded
		report.Message = "This older workspace needs a recoverable upgrade before it can be opened. The original will be kept as a checkpoint."
		report.CanMigrate = true
	case ws.Schema < runtime.Schema:
		report.Status = CompatibilityTooOld
		report.Message = fmt.Sprintf("This workspace uses older format %d, which this build cannot upgrade safely. Nothing was changed.", ws.Schema)
	case ws.Schema > runtime.Schema:
		report.Status = CompatibilityDowngradeBlocked
		report.Message = fmt.Sprintf("This workspace uses newer format %d. This build only understands format %d, so opening it has been blocked.", ws.Schema, runtime.Schema)
	}
	return report
}

func layBuild(build string) string {
	if !validIdentityLabel(build, 128) {
		return "an earlier ECO build"
	}
	return build
}

func loadExistingMasterKey(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("the workspace key is missing")
		}
		return nil, err
	}
	key, err := unprotectLocalKey(data)
	if err != nil {
		return nil, errors.New("the workspace key could not be unlocked for this user")
	}
	if len(key) != 32 {
		zeroBytes(key)
		return nil, errors.New("the workspace key is invalid")
	}
	return key, nil
}

func readEncryptedWorkspace(root string, key []byte) (Workspace, error) {
	data, err := os.ReadFile(filepath.Join(root, "workspace.ecodb"))
	if err != nil {
		if os.IsNotExist(err) {
			return Workspace{}, errors.New("the encrypted workspace record is missing")
		}
		return Workspace{}, err
	}
	plain, err := decryptBlob(key, metaMagic, data)
	if err != nil {
		return Workspace{}, errors.New("the workspace record could not be authenticated; no data was opened")
	}
	defer zeroBytes(plain)
	var ws Workspace
	if err = json.Unmarshal(plain, &ws); err != nil {
		return Workspace{}, errors.New("the workspace record is not a supported ECO format")
	}
	if ws.Schema < 1 {
		return Workspace{}, errors.New("the workspace format number is invalid")
	}
	return ws, nil
}

func validateWorkspaceIdentity(identity WorkspaceIdentity, ws Workspace) error {
	if identity.Format != workspaceIdentityFormat {
		return errors.New("the workspace identity file is not a supported ECO identity")
	}
	if identity.Kind != "development" {
		return errors.New("this folder is not marked as an ECO development workspace")
	}
	if identity.Schema != ws.Schema {
		return errors.New("the workspace identity and encrypted record disagree about their format; opening was blocked")
	}
	if ws.Schema >= 2 {
		if !safeRecordID(identity.ID) || !safeRecordID(ws.WorkspaceID) || identity.ID != ws.WorkspaceID {
			return errors.New("the workspace identity does not match the encrypted record; opening was blocked")
		}
		if !validWorkspaceName(ws.WorkspaceName) || !validIdentityLabel(ws.CreatedByBuild, 128) || ws.CreatedAt.IsZero() {
			return errors.New("the authenticated workspace creation record is incomplete; opening was blocked")
		}
		if !validIdentityLabel(identity.CreatedByCandidate, 256) || !validIdentityLabel(ws.CreatedByCandidate, 256) || identity.CreatedByCandidate != ws.CreatedByCandidate {
			return errors.New("the workspace candidate identity is absent or does not match the authenticated record; opening was blocked")
		}
	}
	return nil
}

func validWorkspaceName(name string) bool {
	trimmed := strings.TrimSpace(name)
	return trimmed != "" && trimmed == name && len([]rune(name)) <= 120 && strings.IndexFunc(name, unicode.IsControl) < 0
}

func validIdentityLabel(value string, limit int) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed != "" && trimmed == value && len([]rune(value)) <= limit && strings.IndexFunc(value, unicode.IsControl) < 0
}

func readWorkspaceIdentity(root string) (WorkspaceIdentity, error) {
	data, err := os.ReadFile(filepath.Join(root, workspaceIdentityFile))
	if err != nil {
		return WorkspaceIdentity{}, err
	}
	var identity WorkspaceIdentity
	if err = json.Unmarshal(data, &identity); err != nil {
		return WorkspaceIdentity{}, errors.New("the workspace identity file is damaged; opening was blocked")
	}
	return identity, nil
}

func writeWorkspaceIdentity(root string, identity WorkspaceIdentity) error {
	data, err := json.MarshalIndent(identity, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(root, workspaceIdentityFile)
	tmp := path + ".tmp"
	if err = os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write workspace identity: %w", err)
	}
	if err = os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("activate workspace identity: %w", err)
	}
	return nil
}

func (v *Vault) recordLifecycle(actor, typ, summary string, details map[string]any) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.addChangeUnlocked(actor, typ, summary, details)
	return v.saveUnlocked()
}

func sessionFor(v *Vault, disposition WorkspaceDisposition, explicit bool, checkpoint string) WorkspaceSession {
	return WorkspaceSession{
		Vault:         v,
		Identity:      v.Identity,
		Path:          v.Root,
		Disposition:   disposition,
		Compatibility: compatibilityFor(v.Workspace, v.runtime),
		Explicit:      explicit,
		Checkpoint:    checkpoint,
	}
}

type ResetReceipt struct {
	WorkspaceID       string
	Path              string
	PreviousEvidence  int
	PreviousMatters   int
	PreviousQuestions int
	ObjectsRemoved    int
	CleanupWarnings   []string
}

type ResetPhase string

const resetBeforeObjectCleanup ResetPhase = "before-object-cleanup"

func resetVault(v *Vault) (ResetReceipt, error) {
	return resetVaultWithHook(v, nil)
}

func resetVaultWithHook(v *Vault, hook func(ResetPhase) error) (ResetReceipt, error) {
	if v == nil {
		return ResetReceipt{}, errors.New("no ECO development workspace is selected")
	}
	absolute, err := normaliseWorkspaceRoot(v.Root)
	if err != nil {
		return ResetReceipt{}, err
	}
	if absolute != v.Root {
		return ResetReceipt{}, errors.New("the selected workspace path changed unexpectedly; reset was blocked")
	}
	if info, statErr := os.Lstat(v.Root); statErr != nil || info.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(info) {
		return ResetReceipt{}, errors.New("reset was blocked because the selected workspace folder is unavailable or is a symbolic link, junction or reparse point")
	}
	objectsCanonical, err := validateObjectsDirectory(v.Root, v.Objects)
	if err != nil {
		return ResetReceipt{}, fmt.Errorf("reset was blocked before records changed because the encrypted object folder is unsafe: %w", err)
	}
	v.opMu.Lock()
	defer v.opMu.Unlock()
	v.mu.Lock()
	identity, err := readWorkspaceIdentity(v.Root)
	if err != nil {
		v.mu.Unlock()
		return ResetReceipt{}, errors.New("reset was blocked because the selected folder has no valid ECO workspace identity")
	}
	if err = validateWorkspaceIdentity(identity, v.Workspace); err != nil {
		v.mu.Unlock()
		return ResetReceipt{}, fmt.Errorf("reset was blocked: %w", err)
	}
	if identity.ID != v.Identity.ID || identity.Kind != "development" {
		v.mu.Unlock()
		return ResetReceipt{}, errors.New("reset was blocked because the selected workspace identity changed")
	}
	old := v.Workspace
	managed := make(map[string]bool)
	for _, item := range old.Evidence {
		managed[item.ObjectFile] = true
	}
	for _, record := range old.Preservations {
		managed[record.ObjectFile] = true
		managed[record.ObjectFile+".part"] = true
		managed[record.ObjectFile+".tmp"] = true
	}
	for name := range managed {
		if _, targetErr := managedObjectTarget(objectsCanonical, name); targetErr != nil {
			v.mu.Unlock()
			return ResetReceipt{}, fmt.Errorf("reset was blocked before records changed: %w", targetErr)
		}
	}
	receipt := ResetReceipt{
		WorkspaceID:       identity.ID,
		Path:              v.Root,
		PreviousEvidence:  len(old.Evidence),
		PreviousMatters:   len(old.Matters),
		PreviousQuestions: len(old.Questions),
	}
	reset := newWorkspaceForRuntime(v.runtime, identity.ID, identity.Name)
	reset.CreatedAt = old.CreatedAt
	reset.CreatedByBuild = old.CreatedByBuild
	reset.CreatedByCandidate = old.CreatedByCandidate
	previousAudit := ""
	if len(old.Changes) > 0 {
		previousAudit = old.Changes[0].Hash
	}
	v.Workspace = reset
	v.addChangeWithPreviousUnlocked(previousAudit, "user", "workspace-reset", "Reset only the selected ECO development workspace", map[string]any{
		"workspace_id":       identity.ID,
		"previous_evidence":  receipt.PreviousEvidence,
		"previous_matters":   receipt.PreviousMatters,
		"previous_questions": receipt.PreviousQuestions,
	})
	if err = v.saveUnlocked(); err != nil {
		v.Workspace = old
		v.mu.Unlock()
		return ResetReceipt{}, fmt.Errorf("the workspace could not be reset; its records were left unchanged: %w", err)
	}
	v.mu.Unlock()

	if hook != nil {
		if err = hook(resetBeforeObjectCleanup); err != nil {
			return receipt, err
		}
	}
	objectsCanonical, err = validateObjectsDirectory(v.Root, v.Objects)
	if err != nil {
		return receipt, fmt.Errorf("workspace records were reset, but object cleanup was blocked because the objects folder changed: %w", err)
	}
	for name := range managed {
		target, targetErr := managedObjectTarget(objectsCanonical, name)
		if targetErr != nil {
			return receipt, fmt.Errorf("workspace records were reset, but no objects were deleted: %w", targetErr)
		}
		if info, statErr := os.Lstat(target); statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(info) || !info.Mode().IsRegular() {
				return receipt, errors.New("workspace records were reset, but no objects were deleted because a managed object target is not a normal file")
			}
		} else if !os.IsNotExist(statErr) {
			return receipt, fmt.Errorf("workspace records were reset, but no objects were deleted because a managed object could not be inspected: %w", statErr)
		}
	}
	for name := range managed {
		currentObjects, containmentErr := validateObjectsDirectory(v.Root, v.Objects)
		if containmentErr != nil || !sameFilesystemPath(currentObjects, objectsCanonical) {
			return receipt, errors.New("workspace records were reset, but remaining object cleanup stopped because the objects folder changed")
		}
		path, targetErr := managedObjectTarget(currentObjects, name)
		if targetErr != nil {
			return receipt, targetErr
		}
		err = os.Remove(path)
		if err == nil {
			receipt.ObjectsRemoved++
		} else if !os.IsNotExist(err) {
			receipt.CleanupWarnings = append(receipt.CleanupWarnings, "An unreferenced encrypted object could not be removed.")
		}
	}
	return receipt, nil
}

func safeManagedObjectName(name string) bool {
	if name == "" || filepath.Base(name) != name || strings.ContainsAny(name, `/\`) {
		return false
	}
	return strings.HasSuffix(name, ".ecoobj") || strings.HasSuffix(name, ".ecoobj.part") || strings.HasSuffix(name, ".ecoobj.tmp")
}

func (v *Vault) Close() {
	if v == nil {
		return
	}
	v.opMu.Lock()
	defer v.opMu.Unlock()
	v.mu.Lock()
	defer v.mu.Unlock()
	zeroBytes(v.key)
	v.key = nil
}
