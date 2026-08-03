package eco

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	restoreStateFormat       = "ECO-RESTORE-1"
	restoreRoleFormat        = "ECO-RESTORE-ROLE-1"
	restoreStageIdentityFile = ".eco-restore-stage.json"
)

type restoreTransactionPhase string

const (
	restorePrepared      restoreTransactionPhase = "prepared"
	restoreStaged        restoreTransactionPhase = "staged"
	restoreOriginalMoved restoreTransactionPhase = "original-moved"
	restoreActivated     restoreTransactionPhase = "activated"
	restoreRecovered     restoreTransactionPhase = "recovered"
)

var errRestoreInterrupted = errors.New("simulated portable restore interruption")

type restoreState struct {
	Format              string                  `json:"format"`
	Root                string                  `json:"root"`
	Checkpoint          string                  `json:"checkpoint"`
	Stage               string                  `json:"stage"`
	Failed              string                  `json:"failed"`
	Nonce               string                  `json:"nonce"`
	OriginalWorkspaceID string                  `json:"original_workspace_id"`
	RestoredWorkspaceID string                  `json:"restored_workspace_id,omitempty"`
	CheckpointID        string                  `json:"checkpoint_id"`
	StageID             string                  `json:"stage_id"`
	FailedID            string                  `json:"failed_id"`
	BuildID             string                  `json:"build_id"`
	CandidateID         string                  `json:"candidate_id"`
	Schema              int                     `json:"schema"`
	SourceSHA256        string                  `json:"source_sha256"`
	Phase               restoreTransactionPhase `json:"phase"`
	StartedAt           time.Time               `json:"started_at"`
	AuthTag             string                  `json:"auth_tag"`
}

type restoreRoleRecord struct {
	Format              string `json:"format"`
	Nonce               string `json:"nonce"`
	Role                string `json:"role"`
	ID                  string `json:"id"`
	OriginalWorkspaceID string `json:"original_workspace_id"`
	SourceSHA256        string `json:"source_sha256"`
	AuthTag             string `json:"auth_tag"`
}

type RestoreRecoveryRequiredError struct{ Path string }

func (e *RestoreRecoveryRequiredError) Error() string {
	return "This workspace has an unfinished portable restore. Choose Recover workspace before opening it."
}

func restoreStatePath(root string) string { return root + ".eco-restore.json" }

func restoreSiblingPaths(root, nonce string) (checkpoint, stage, failed string) {
	parent := filepath.Dir(root)
	base := filepath.Base(root)
	return filepath.Join(parent, base+".restore-checkpoint-"+nonce),
		filepath.Join(parent, base+".restore-stage-"+nonce),
		filepath.Join(parent, base+".restore-failed-"+nonce)
}

func restoreRolePath(path string) string { return path + ".eco-restore-role.json" }

func restoreStateAuthentication(state restoreState, key []byte) (string, error) {
	unsigned := state
	unsigned.AuthTag = ""
	data, err := json.Marshal(unsigned)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("ECO-RESTORE-STATE\x00"))
	_, _ = mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func signRestoreState(state *restoreState, key []byte) error {
	tag, err := restoreStateAuthentication(*state, key)
	if err == nil {
		state.AuthTag = tag
	}
	return err
}

func authenticateRestoreState(state restoreState, key []byte) error {
	expected, err := restoreStateAuthentication(state, key)
	provided, decodeErr := hex.DecodeString(state.AuthTag)
	if err != nil || decodeErr != nil || len(provided) != sha256.Size || !hmac.Equal(provided, mustDecodeHex(expected)) {
		return errors.New("the portable restore recovery record could not be authenticated; automatic changes were blocked")
	}
	return nil
}

func restoreRoleFor(state restoreState, role string) (restoreRoleRecord, string, error) {
	record := restoreRoleRecord{Format: restoreRoleFormat, Nonce: state.Nonce, Role: role, OriginalWorkspaceID: state.OriginalWorkspaceID, SourceSHA256: state.SourceSHA256}
	var path string
	switch role {
	case "checkpoint":
		record.ID, path = state.CheckpointID, state.Checkpoint
	case "stage":
		record.ID, path = state.StageID, state.Stage
	case "failed":
		record.ID, path = state.FailedID, state.Failed
	default:
		return restoreRoleRecord{}, "", errors.New("unknown portable restore folder role")
	}
	return record, restoreRolePath(path), nil
}

func restoreRoleAuthentication(role restoreRoleRecord, key []byte) (string, error) {
	unsigned := role
	unsigned.AuthTag = ""
	data, err := json.Marshal(unsigned)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("ECO-RESTORE-ROLE\x00"))
	_, _ = mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func signRestoreRole(role *restoreRoleRecord, key []byte) error {
	tag, err := restoreRoleAuthentication(*role, key)
	if err == nil {
		role.AuthTag = tag
	}
	return err
}

func verifyRestoreRoleData(state restoreState, roleName string, key, data []byte) error {
	expected, _, err := restoreRoleFor(state, roleName)
	if err != nil {
		return err
	}
	var actual restoreRoleRecord
	if err = json.Unmarshal(data, &actual); err != nil {
		return errors.New("a portable restore folder identity is damaged")
	}
	provided := actual.AuthTag
	actual.AuthTag = ""
	if actual != expected {
		return errors.New("a portable restore folder identity does not match the authenticated recovery record")
	}
	tag, err := restoreRoleAuthentication(actual, key)
	providedBytes, decodeErr := hex.DecodeString(provided)
	if err != nil || decodeErr != nil || !hmac.Equal(providedBytes, mustDecodeHex(tag)) {
		return errors.New("a portable restore folder identity could not be authenticated")
	}
	return nil
}

func writeRestoreRole(state restoreState, roleName string, key []byte) error {
	role, path, err := restoreRoleFor(state, roleName)
	if err != nil {
		return err
	}
	if err = signRestoreRole(&role, key); err != nil {
		return err
	}
	data, err := json.MarshalIndent(role, "", "  ")
	if err != nil {
		return err
	}
	return writeNewControlFile(path, data, "portable restore folder identity")
}

func verifyRestoreRole(state restoreState, roleName string, key []byte) error {
	_, path, err := restoreRoleFor(state, roleName)
	if err != nil {
		return err
	}
	data, err := readNormalControlFile(path, "portable restore folder identity")
	if err != nil {
		return err
	}
	return verifyRestoreRoleData(state, roleName, key, data)
}

func writeRestoreStageIdentity(state restoreState, key []byte) error {
	role, _, err := restoreRoleFor(state, "stage")
	if err != nil {
		return err
	}
	if err = signRestoreRole(&role, key); err != nil {
		return err
	}
	data, err := json.MarshalIndent(role, "", "  ")
	if err != nil {
		return err
	}
	return writeNewControlFile(filepath.Join(state.Stage, restoreStageIdentityFile), data, "portable restore stage identity")
}

func verifyRestoreStageIdentityAt(state restoreState, path string, key []byte) error {
	data, err := readNormalControlFile(filepath.Join(path, restoreStageIdentityFile), "portable restore stage identity")
	if err != nil {
		return err
	}
	return verifyRestoreRoleData(state, "stage", key, data)
}

func validateRestoreStateStructure(state restoreState, expectedRoot string) error {
	root, err := canonicalWorkspacePath(expectedRoot)
	if err != nil {
		// A crash may leave the root temporarily absent; its canonical parent and
		// authenticated direct-sibling name are still sufficient here.
		logical, parent, parentErr := canonicalWorkspaceParent(expectedRoot)
		if parentErr != nil {
			return err
		}
		root = filepath.Join(parent, filepath.Base(logical))
	}
	if state.Format != restoreStateFormat || !sameFilesystemPath(state.Root, root) || !validMigrationNonce(state.Nonce) {
		return errors.New("the portable restore recovery record does not match the selected workspace")
	}
	if !safeRecordID(state.OriginalWorkspaceID) || !safeRecordID(state.CheckpointID) || !safeRecordID(state.StageID) || !safeRecordID(state.FailedID) || !validIdentityLabel(state.BuildID, 128) || !validIdentityLabel(state.CandidateID, 256) || state.Schema < 1 || len(state.SourceSHA256) != sha256.Size*2 {
		return errors.New("the portable restore recovery identity is invalid")
	}
	if state.RestoredWorkspaceID != "" && !safeRecordID(state.RestoredWorkspaceID) {
		return errors.New("the restored workspace identity is invalid")
	}
	checkpoint, stage, failed := restoreSiblingPaths(root, state.Nonce)
	if !sameFilesystemPath(checkpoint, state.Checkpoint) || !sameFilesystemPath(stage, state.Stage) || !sameFilesystemPath(failed, state.Failed) {
		return errors.New("the portable restore recovery paths do not match their authenticated nonce")
	}
	switch state.Phase {
	case restorePrepared:
	case restoreStaged, restoreOriginalMoved, restoreActivated, restoreRecovered:
		if state.RestoredWorkspaceID == "" {
			return errors.New("the portable restore recovery record has no restored workspace identity")
		}
	default:
		return errors.New("the portable restore recovery phase is invalid")
	}
	return nil
}

func newRestoreState(v *Vault, sourceSHA string) (restoreState, error) {
	nonce, err := newMigrationNonce()
	if err != nil {
		return restoreState{}, err
	}
	checkpoint, stage, failed := restoreSiblingPaths(v.Root, nonce)
	state := restoreState{
		Format: restoreStateFormat, Root: v.Root, Checkpoint: checkpoint, Stage: stage, Failed: failed,
		Nonce: nonce, OriginalWorkspaceID: v.Identity.ID, CheckpointID: NewID("RCP"), StageID: NewID("RST"), FailedID: NewID("RFL"),
		BuildID: v.runtime.BuildID, CandidateID: v.runtime.CandidateID, Schema: v.runtime.Schema, SourceSHA256: sourceSHA,
		Phase: restorePrepared, StartedAt: time.Now().UTC(),
	}
	return state, validateRestoreStateStructure(state, v.Root)
}

func sameRestoreState(left, right restoreState) bool {
	a, _ := json.Marshal(left)
	b, _ := json.Marshal(right)
	return hmac.Equal(a, b)
}

func writeRestoreState(state *restoreState, key []byte) error {
	if err := validateRestoreStateStructure(*state, state.Root); err != nil {
		return err
	}
	signed := *state
	if err := signRestoreState(&signed, key); err != nil {
		return err
	}
	data, err := json.MarshalIndent(signed, "", "  ")
	if err != nil {
		return err
	}
	path := restoreStatePath(state.Root)
	var previous []byte
	if signed.Phase == restorePrepared {
		if _, statErr := os.Lstat(path); statErr == nil {
			return errors.New("an unfinished portable restore already exists; recover it before starting another")
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
	} else {
		previous, err = readNormalControlFile(path, "portable restore recovery record")
		if err != nil {
			return err
		}
		var current restoreState
		if err = json.Unmarshal(previous, &current); err != nil || authenticateRestoreState(current, key) != nil {
			return errors.New("the existing portable restore recovery record is damaged or unauthenticated")
		}
		current.AuthTag, signed.AuthTag = "", ""
		allowed := (current.Phase == restorePrepared && signed.Phase == restoreStaged) ||
			(current.Phase == restoreStaged && signed.Phase == restoreOriginalMoved) ||
			(current.Phase == restoreOriginalMoved && signed.Phase == restoreActivated) ||
			(current.Phase == restoreActivated && signed.Phase == restoreRecovered)
		current.Phase = signed.Phase
		if current.Phase == restoreStaged {
			current.RestoredWorkspaceID = signed.RestoredWorkspaceID
		}
		if !allowed || !sameRestoreState(current, signed) {
			return errors.New("the portable restore recovery transition is not valid")
		}
		if err = signRestoreState(&signed, key); err != nil {
			return err
		}
		data, _ = json.MarshalIndent(signed, "", "  ")
	}
	tmp := path + ".tmp-" + NewID("RST")
	if err = writeNewControlFile(tmp, data, "portable restore recovery temporary file"); err != nil {
		return err
	}
	if signed.Phase != restorePrepared {
		current, readErr := readNormalControlFile(path, "portable restore recovery record")
		if readErr != nil || !bytes.Equal(current, previous) {
			removeExpectedControlFile(tmp, data)
			return errors.New("the portable restore recovery record changed while it was being updated")
		}
	}
	if err = os.Rename(tmp, path); err != nil {
		removeExpectedControlFile(tmp, data)
		return err
	}
	*state = signed
	return nil
}

func candidateRestoreKey(state restoreState) ([]byte, error) {
	for _, path := range []string{state.Root, state.Checkpoint, state.Failed} {
		if _, err := validateDirectSibling(state.Root, path, filepath.Base(path), true); err != nil {
			continue
		}
		key, err := loadExistingMasterKey(filepath.Join(path, "vault.key"))
		if err == nil && authenticateRestoreState(state, key) == nil {
			return key, nil
		}
		zeroBytes(key)
	}
	return nil, errors.New("the portable restore recovery record could not be authenticated by the original workspace")
}

func readRestoreState(root string) (restoreState, []byte, error) {
	data, err := readNormalControlFile(restoreStatePath(root), "portable restore recovery record")
	if err != nil {
		return restoreState{}, nil, err
	}
	var state restoreState
	if err = json.Unmarshal(data, &state); err != nil {
		return restoreState{}, nil, errors.New("the portable restore recovery record is damaged")
	}
	if err = validateRestoreStateStructure(state, root); err != nil {
		return restoreState{}, nil, err
	}
	key, err := candidateRestoreKey(state)
	if err != nil {
		return restoreState{}, nil, err
	}
	return state, key, nil
}

func reloadAuthenticatedRestoreState(expected restoreState) ([]byte, error) {
	actual, key, err := readRestoreState(expected.Root)
	if err != nil {
		return nil, err
	}
	if !sameRestoreState(actual, expected) {
		zeroBytes(key)
		return nil, errors.New("the portable restore recovery record changed; filesystem changes were blocked")
	}
	return key, nil
}

func verifyOriginalRestoreWorkspace(path string, state restoreState, key []byte) error {
	ws, err := readEncryptedWorkspace(path, key)
	if err != nil {
		return err
	}
	if ws.WorkspaceID != state.OriginalWorkspaceID {
		return errors.New("the original workspace does not match the authenticated portable restore record")
	}
	return nil
}

func verifyRestoredWorkspaceAt(path string, state restoreState) error {
	key, err := loadExistingMasterKey(filepath.Join(path, "vault.key"))
	if err != nil {
		return err
	}
	defer zeroBytes(key)
	ws, err := readEncryptedWorkspace(path, key)
	if err != nil {
		return err
	}
	identity, err := readWorkspaceIdentity(path)
	if err != nil {
		return err
	}
	if err = validateWorkspaceIdentity(identity, ws); err != nil {
		return err
	}
	if ws.WorkspaceID != state.RestoredWorkspaceID || ws.Schema != state.Schema || ws.BuildID != state.BuildID || ws.CreatedByCandidate != state.CandidateID {
		return errors.New("the staged workspace does not match the authenticated portable restore destination")
	}
	return nil
}

func validateRestoreRoleDirectory(state restoreState, path, role string, mustExist bool, key []byte) error {
	if _, err := validateDirectSibling(state.Root, path, filepath.Base(path), mustExist); err != nil {
		return err
	}
	if err := verifyRestoreRole(state, role, key); err != nil {
		return err
	}
	if role == "stage" && mustExist {
		return verifyRestoreStageIdentityAt(state, path, key)
	}
	return nil
}

func secureRestoreRename(state restoreState, source, destination, sourceRole, destinationRole string, ops filesystemOps) error {
	return objectBoundRename(source, destination, func() error {
		key, err := reloadAuthenticatedRestoreState(state)
		if err != nil {
			return err
		}
		defer zeroBytes(key)
		if sourceRole == "root-original" {
			if _, err = validateDirectSibling(state.Root, source, filepath.Base(state.Root), true); err != nil {
				return err
			}
			if err = verifyOriginalRestoreWorkspace(source, state, key); err != nil {
				return err
			}
		} else {
			if err = validateRestoreRoleDirectory(state, source, sourceRole, true, key); err != nil {
				return err
			}
			if sourceRole == "checkpoint" {
				err = verifyOriginalRestoreWorkspace(source, state, key)
			} else if sourceRole == "stage" {
				err = verifyRestoredWorkspaceAt(source, state)
			}
			if err != nil {
				return err
			}
		}
		if destinationRole == "root" {
			_, err = validateDirectSibling(state.Root, destination, filepath.Base(state.Root), false)
		} else {
			err = validateRestoreRoleDirectory(state, destination, destinationRole, false, key)
		}
		if err != nil {
			return err
		}
		if _, statErr := os.Lstat(destination); statErr == nil {
			return errors.New("the portable restore rename destination already exists")
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
		return nil
	}, ops.beforeRename)
}

func removeRestoreRole(state restoreState, role string, key []byte, ops filesystemOps) error {
	_, path, err := restoreRoleFor(state, role)
	if err != nil {
		return err
	}
	return removeAuthenticatedControlFile(path, "portable restore folder identity", func(data []byte) error {
		return verifyRestoreRoleData(state, role, key, data)
	}, ops)
}

func removeRestoreStageIdentity(state restoreState, root string, key []byte, ops filesystemOps) error {
	return removeAuthenticatedControlFile(filepath.Join(root, restoreStageIdentityFile), "portable restore stage identity", func(data []byte) error {
		return verifyRestoreRoleData(state, "stage", key, data)
	}, ops)
}

func removeRestoreMarker(state restoreState, key []byte, ops filesystemOps) error {
	return removeAuthenticatedControlFile(restoreStatePath(state.Root), "portable restore recovery record", func(data []byte) error {
		var actual restoreState
		if err := json.Unmarshal(data, &actual); err != nil || !sameRestoreState(actual, state) {
			return errors.New("the portable restore recovery record changed before cleanup")
		}
		return authenticateRestoreState(actual, key)
	}, ops)
}

func removeRestoreStage(state restoreState, key []byte, ops filesystemOps) error {
	if _, err := os.Lstat(state.Stage); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return objectBoundRemoveTree(state.Stage, func() error {
		currentKey, err := reloadAuthenticatedRestoreState(state)
		if err != nil {
			return err
		}
		defer zeroBytes(currentKey)
		if err = verifyRestoreRole(state, "stage", currentKey); err != nil {
			return err
		}
		return verifyRestoreStageIdentityAt(state, state.Stage, currentKey)
	}, ops.beforeRemove)
}

func cleanupRestoreControls(state restoreState, key []byte, ops filesystemOps) error {
	var result error
	for _, role := range []string{"checkpoint", "stage", "failed"} {
		result = errors.Join(result, removeRestoreRole(state, role, key, ops))
	}
	result = errors.Join(result, removeRestoreMarker(state, key, ops))
	return result
}

func rollbackRestoreState(state restoreState, ops filesystemOps) (bool, error) {
	key, err := reloadAuthenticatedRestoreState(state)
	if err != nil {
		return false, err
	}
	defer zeroBytes(key)
	rootOriginal := verifyOriginalRestoreWorkspace(state.Root, state, key) == nil
	checkpointOriginal := verifyOriginalRestoreWorkspace(state.Checkpoint, state, key) == nil
	if !rootOriginal && checkpointOriginal {
		if _, statErr := os.Lstat(state.Root); statErr == nil {
			if err = secureRestoreRename(state, state.Root, state.Failed, "stage", "failed", ops); err != nil {
				return false, fmt.Errorf("move failed restored workspace aside: %w", err)
			}
		}
		if err = secureRestoreRename(state, state.Checkpoint, state.Root, "checkpoint", "root", ops); err != nil {
			return false, fmt.Errorf("restore the original workspace checkpoint: %w", err)
		}
		rootOriginal = true
	}
	if !rootOriginal {
		return false, errors.New("the authenticated original workspace is not available at its root or checkpoint; the recovery record was retained")
	}
	if err = removeRestoreStage(state, key, ops); err != nil {
		return true, fmt.Errorf("the original workspace is available, but staged restore cleanup needs attention: %w", err)
	}
	if err = cleanupRestoreControls(state, key, ops); err != nil {
		return true, fmt.Errorf("the original workspace is available, but restore control cleanup needs attention: %w", err)
	}
	return true, nil
}

func RecoverPortableRestore(root string, runtime RuntimeIdentity) (WorkspaceSession, RecoveryReceipt, error) {
	absolute, err := canonicalWorkspacePath(root)
	if err != nil {
		return WorkspaceSession{}, RecoveryReceipt{}, err
	}
	lease, err := acquireWorkspaceLifecycleLease(absolute)
	if err != nil {
		return WorkspaceSession{}, RecoveryReceipt{}, err
	}
	defer lease.Close()
	state, key, err := readRestoreState(absolute)
	if err != nil {
		return WorkspaceSession{}, RecoveryReceipt{}, err
	}
	defer zeroBytes(key)
	receipt := RecoveryReceipt{Path: absolute, Checkpoint: state.Checkpoint}
	if state.BuildID != runtime.BuildID || state.CandidateID != runtime.CandidateID || state.Schema != runtime.Schema {
		return WorkspaceSession{}, receipt, errors.New("this unfinished portable restore belongs to another development candidate or build; recovery was blocked")
	}
	stageIdentityValid := verifyRestoreStageIdentityAt(state, state.Root, key) == nil
	rootRestored := verifyRestoredWorkspaceAt(state.Root, state) == nil && (stageIdentityValid || state.Phase == restoreRecovered)
	checkpointOriginal := verifyOriginalRestoreWorkspace(state.Checkpoint, state, key) == nil
	if rootRestored && checkpointOriginal {
		if state.Phase == restoreOriginalMoved {
			next := state
			next.Phase = restoreActivated
			if err = writeRestoreState(&next, key); err != nil {
				return WorkspaceSession{}, receipt, err
			}
			state = next
		}
		v, openErr := openVaultIgnoringRecovery(state.Root, runtime)
		if openErr != nil {
			return WorkspaceSession{}, receipt, openErr
		}
		if err = v.recordLifecycle("system", "workspace-recovered", "Recovered and verified an interrupted portable restore", map[string]any{
			"checkpoint": filepath.Base(state.Checkpoint),
			"phase":      state.Phase,
		}); err != nil {
			v.Close()
			return WorkspaceSession{}, receipt, err
		}
		if stageIdentityValid {
			if err = removeRestoreStageIdentity(state, state.Root, key, operatingFilesystem); err != nil {
				v.Close()
				return WorkspaceSession{}, receipt, err
			}
		}
		if state.Phase != restoreRecovered {
			next := state
			next.Phase = restoreRecovered
			if err = writeRestoreState(&next, key); err != nil {
				v.Close()
				return WorkspaceSession{}, receipt, err
			}
			state = next
		}
		if err = cleanupRestoreControls(state, key, operatingFilesystem); err != nil {
			v.Close()
			return WorkspaceSession{}, receipt, err
		}
		receipt.MigrationKept = true
		receipt.WorkspaceOpened = true
		receipt.Message = "The authenticated portable restore was recovered. The original checkpoint was kept."
		receipt.Compatibility = compatibilityFor(v.Workspace, runtime)
		return sessionFor(v, DispositionRecovered, true, state.Checkpoint), receipt, nil
	}
	restored, rollbackErr := rollbackRestoreState(state, operatingFilesystem)
	if rollbackErr != nil {
		return WorkspaceSession{}, receipt, rollbackErr
	}
	v, err := openVaultIgnoringRecovery(state.Root, runtime)
	if err != nil {
		return WorkspaceSession{}, receipt, err
	}
	if err = v.recordLifecycle("system", "workspace-recovered", "Recovered the original workspace after an interrupted portable restore", map[string]any{
		"phase": state.Phase,
	}); err != nil {
		v.Close()
		return WorkspaceSession{}, receipt, err
	}
	receipt.OriginalRestored = restored
	receipt.WorkspaceOpened = true
	receipt.Message = "The interrupted portable restore was rolled back. The original workspace was restored."
	receipt.Compatibility = compatibilityFor(v.Workspace, runtime)
	return sessionFor(v, DispositionRecovered, true, ""), receipt, nil
}
