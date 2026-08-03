package eco

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	migrationFormat            = "ECO-MIGRATION-2"
	migrationRoleFormat        = "ECO-MIGRATION-ROLE-1"
	migrationStageIdentityFile = ".eco-migration-stage.json"
)

type migrationRoleRecord struct {
	Format      string `json:"format"`
	Nonce       string `json:"nonce"`
	Role        string `json:"role"`
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	AuthTag     string `json:"auth_tag"`
}

func newMigrationNonce() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func validMigrationNonce(value string) bool {
	if len(value) != 32 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func migrationSiblingPaths(root, nonce string) (checkpoint, stage, failed string) {
	parent := filepath.Dir(root)
	base := filepath.Base(root)
	return filepath.Join(parent, base+".migration-checkpoint-"+nonce),
		filepath.Join(parent, base+".migration-stage-"+nonce),
		filepath.Join(parent, base+".migration-failed-"+nonce)
}

func migrationRolePath(path string) string { return path + ".eco-role.json" }

func migrationStateAuthentication(state migrationState, key []byte) (string, error) {
	unsigned := state
	unsigned.AuthTag = ""
	data, err := json.Marshal(unsigned)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("ECO-MIGRATION-STATE\x00"))
	_, _ = mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func signMigrationState(state *migrationState, key []byte) error {
	tag, err := migrationStateAuthentication(*state, key)
	if err != nil {
		return err
	}
	state.AuthTag = tag
	return nil
}

func authenticateMigrationState(state migrationState, key []byte) error {
	if len(state.AuthTag) != sha256.Size*2 {
		return errors.New("the migration recovery record has no valid authenticator")
	}
	expected, err := migrationStateAuthentication(state, key)
	if err != nil {
		return err
	}
	provided, err := hex.DecodeString(state.AuthTag)
	if err != nil || !hmac.Equal(provided, mustDecodeHex(expected)) {
		return errors.New("the migration recovery record could not be authenticated; automatic changes were blocked")
	}
	return nil
}

func migrationRoleAuthentication(role migrationRoleRecord, key []byte) (string, error) {
	unsigned := role
	unsigned.AuthTag = ""
	data, err := json.Marshal(unsigned)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("ECO-MIGRATION-ROLE\x00"))
	_, _ = mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func roleForMigration(state migrationState, role string) (migrationRoleRecord, string, error) {
	record := migrationRoleRecord{
		Format:      migrationRoleFormat,
		Nonce:       state.Nonce,
		Role:        role,
		WorkspaceID: state.WorkspaceID,
	}
	var path string
	switch role {
	case "checkpoint":
		record.ID, path = state.CheckpointID, state.Checkpoint
	case "stage":
		record.ID, path = state.StageID, state.Stage
	case "failed":
		record.ID, path = state.FailedID, state.Failed
	default:
		return migrationRoleRecord{}, "", errors.New("unknown migration folder role")
	}
	return record, migrationRolePath(path), nil
}

func writeMigrationRole(state migrationState, roleName string, key []byte) error {
	role, path, err := roleForMigration(state, roleName)
	if err != nil {
		return err
	}
	role.AuthTag, err = migrationRoleAuthentication(role, key)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(role, "", "  ")
	if err != nil {
		return err
	}
	return writeNewControlFile(path, data, "migration folder identity")
}

func writeStageDirectoryIdentity(state migrationState, key []byte) error {
	role, _, err := roleForMigration(state, "stage")
	if err != nil {
		return err
	}
	role.AuthTag, err = migrationRoleAuthentication(role, key)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(role, "", "  ")
	if err != nil {
		return err
	}
	return writeNewControlFile(filepath.Join(state.Stage, migrationStageIdentityFile), data, "migration stage identity")
}

func verifyStageDirectoryIdentityAt(state migrationState, directory string, key []byte) error {
	data, err := readNormalControlFile(filepath.Join(directory, migrationStageIdentityFile), "migration stage identity")
	if err != nil {
		return err
	}
	return verifyStageDirectoryIdentityData(state, key, data)
}

func verifyStageDirectoryIdentityData(state migrationState, key, data []byte) error {
	expected, _, err := roleForMigration(state, "stage")
	if err != nil {
		return err
	}
	var actual migrationRoleRecord
	if err = json.Unmarshal(data, &actual); err != nil {
		return errors.New("the migration stage identity is damaged")
	}
	provided := actual.AuthTag
	actual.AuthTag = ""
	expected.AuthTag = ""
	if actual != expected {
		return errors.New("the migration stage identity does not match the authenticated recovery record")
	}
	tag, err := migrationRoleAuthentication(actual, key)
	providedBytes, decodeErr := hex.DecodeString(provided)
	if err != nil || decodeErr != nil || !hmac.Equal(providedBytes, mustDecodeHex(tag)) {
		return errors.New("the migration stage identity could not be authenticated")
	}
	return nil
}

func verifyMigrationRole(state migrationState, roleName string, key []byte) error {
	_, path, err := roleForMigration(state, roleName)
	if err != nil {
		return err
	}
	data, err := readNormalControlFile(path, "migration folder identity")
	if err != nil {
		return err
	}
	return verifyMigrationRoleData(state, roleName, key, data)
}

func verifyMigrationRoleData(state migrationState, roleName string, key, data []byte) error {
	expected, _, err := roleForMigration(state, roleName)
	if err != nil {
		return err
	}
	var actual migrationRoleRecord
	if err = json.Unmarshal(data, &actual); err != nil {
		return errors.New("a migration folder identity is damaged; filesystem changes were blocked")
	}
	provided := actual.AuthTag
	actual.AuthTag = ""
	expected.AuthTag = ""
	if actual != expected {
		return errors.New("a migration folder identity does not match the authenticated recovery record")
	}
	tag, err := migrationRoleAuthentication(actual, key)
	if err != nil {
		return err
	}
	providedBytes, decodeErr := hex.DecodeString(provided)
	if decodeErr != nil || !hmac.Equal(providedBytes, mustDecodeHex(tag)) {
		return errors.New("a migration folder identity could not be authenticated")
	}
	return nil
}

func writeNewControlFile(path string, data []byte, description string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("%s already exists", description)
	} else if !os.IsNotExist(err) {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	created, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return err
	}
	ok := false
	defer func() {
		_ = f.Close()
		if !ok {
			removeCreatedNormalFile(path, created)
		}
	}()
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

func removeCreatedNormalFile(path string, created os.FileInfo) {
	_ = objectBoundRemoveFile(path, "partly written migration control file", func([]byte) error {
		current, err := os.Lstat(path)
		if err != nil || current.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(current) || !current.Mode().IsRegular() || !os.SameFile(created, current) {
			return errors.New("the partly written migration control file was replaced; cleanup was blocked")
		}
		return nil
	}, nil)
}

func readNormalControlFile(path, description string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(info) || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a normal file", description)
	}
	return os.ReadFile(path)
}

func validateMigrationStateStructure(state migrationState, expectedRoot string) error {
	root, err := canonicalWorkspacePath(expectedRoot)
	if err != nil {
		return err
	}
	if state.Format != migrationFormat || !sameFilesystemPath(state.Root, root) {
		return errors.New("the migration recovery record does not match the selected workspace")
	}
	if !validMigrationNonce(state.Nonce) || !safeRecordID(state.WorkspaceID) || !safeRecordID(state.CheckpointID) || !safeRecordID(state.StageID) || !safeRecordID(state.FailedID) {
		return errors.New("the migration recovery identity is invalid")
	}
	expectedCheckpoint, expectedStage, expectedFailed := migrationSiblingPaths(root, state.Nonce)
	if !sameFilesystemPath(state.Checkpoint, expectedCheckpoint) || !sameFilesystemPath(state.Stage, expectedStage) || !sameFilesystemPath(state.Failed, expectedFailed) {
		return errors.New("the migration recovery paths do not match their authenticated nonce")
	}
	parent := filepath.Dir(root)
	for _, path := range []string{state.Checkpoint, state.Stage, state.Failed} {
		if !sameFilesystemPath(filepath.Dir(filepath.Clean(path)), parent) {
			return errors.New("the migration recovery paths are not direct workspace siblings")
		}
	}
	if state.FromSchema < 1 || state.ToSchema <= state.FromSchema || !validIdentityLabel(state.SourceBuild, 128) || !validIdentityLabel(state.DestinationBuild, 128) || !validIdentityLabel(state.DestinationCandidate, 256) {
		return errors.New("the migration format or build transition is invalid")
	}
	if state.SourceCandidate != "" && !validIdentityLabel(state.SourceCandidate, 256) {
		return errors.New("the migration source candidate identity is invalid")
	}
	switch state.Phase {
	case migrationPrepared, migrationOriginalMoved, migrationStageReady, migrationActivated:
	default:
		return errors.New("the migration recovery phase is invalid")
	}
	return nil
}

func candidateMigrationKey(state migrationState) ([]byte, error) {
	locations := []struct {
		path string
		base string
	}{
		{state.Root, filepath.Base(state.Root)},
		{state.Checkpoint, filepath.Base(state.Checkpoint)},
		{state.Stage, filepath.Base(state.Stage)},
		{state.Failed, filepath.Base(state.Failed)},
	}
	for _, location := range locations {
		if _, err := validateDirectSibling(state.Root, location.path, location.base, true); err != nil {
			continue
		}
		key, err := loadExistingMasterKey(filepath.Join(location.path, "vault.key"))
		if err != nil {
			continue
		}
		if authenticateMigrationState(state, key) == nil {
			return key, nil
		}
		zeroBytes(key)
	}
	return nil, errors.New("the migration recovery record could not be authenticated by the selected workspace; no filesystem changes were made")
}

func sameMigrationState(a, b migrationState) bool {
	left, _ := json.Marshal(a)
	right, _ := json.Marshal(b)
	return hmac.Equal(left, right)
}

func reloadAuthenticatedMigrationState(expected migrationState) ([]byte, error) {
	actual, key, err := readMigrationState(expected.Root)
	if err != nil {
		return nil, err
	}
	if !sameMigrationState(actual, expected) {
		zeroBytes(key)
		return nil, errors.New("the migration recovery record changed; filesystem changes were blocked")
	}
	return key, nil
}

func reloadAuthenticatedMigrationStateData(expected migrationState, data []byte) ([]byte, error) {
	var actual migrationState
	if err := json.Unmarshal(data, &actual); err != nil {
		return nil, errors.New("the migration recovery record is damaged; automatic recovery was blocked")
	}
	if err := validateMigrationStateStructure(actual, expected.Root); err != nil {
		return nil, err
	}
	key, err := candidateMigrationKey(actual)
	if err != nil {
		return nil, err
	}
	if err = authenticateMigrationState(actual, key); err != nil {
		zeroBytes(key)
		return nil, err
	}
	for _, path := range []string{actual.Checkpoint, actual.Stage, actual.Failed} {
		if _, statErr := os.Lstat(path); statErr == nil {
			if validateErr := validateMigrationDirectory(actual, path, true, key); validateErr != nil {
				zeroBytes(key)
				return nil, fmt.Errorf("an authenticated migration path failed containment checks: %w", validateErr)
			}
		} else if !os.IsNotExist(statErr) {
			zeroBytes(key)
			return nil, statErr
		}
	}
	if !sameMigrationState(actual, expected) {
		zeroBytes(key)
		return nil, errors.New("the migration recovery record changed; filesystem changes were blocked")
	}
	return key, nil
}

func verifyMigrationWorkspace(path string, state migrationState, destination bool, key []byte) error {
	ws, err := readEncryptedWorkspace(path, key)
	if err != nil {
		return err
	}
	if destination {
		if ws.Schema != state.ToSchema || ws.WorkspaceID != state.WorkspaceID || ws.BuildID != state.DestinationBuild || ws.CreatedByCandidate != state.DestinationCandidate {
			return errors.New("the staged or active workspace does not match the authenticated migration destination")
		}
		return nil
	}
	if ws.Schema != state.FromSchema || ws.BuildID != state.SourceBuild || ws.WorkspaceID != state.SourceWorkspaceID || ws.CreatedByCandidate != state.SourceCandidate {
		return errors.New("the checkpoint does not match the authenticated migration source")
	}
	return nil
}

func removeAuthenticatedControlFile(path, description string, validate func([]byte) error, ops filesystemOps) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || fileInfoHasReparsePoint(info) || !info.Mode().IsRegular() {
		return fmt.Errorf("%s changed into an unsafe filesystem object", description)
	}
	return objectBoundRemoveFile(path, description, validate, ops.beforeRemove)
}

func migrationRoleNameForPath(state migrationState, path string) (string, bool) {
	switch {
	case sameFilesystemPath(path, state.Checkpoint):
		return "checkpoint", true
	case sameFilesystemPath(path, state.Stage):
		return "stage", true
	case sameFilesystemPath(path, state.Failed):
		return "failed", true
	default:
		return "", false
	}
}

func validateMigrationDirectory(state migrationState, path string, mustExist bool, key []byte) error {
	role, ok := migrationRoleNameForPath(state, path)
	if !ok {
		return errors.New("the migration directory is not an authenticated migration sibling")
	}
	if _, err := validateDirectSibling(state.Root, path, filepath.Base(path), mustExist); err != nil {
		return err
	}
	if err := verifyMigrationRole(state, role, key); err != nil {
		return err
	}
	if role == "stage" && mustExist {
		return verifyStageDirectoryIdentityAt(state, state.Stage, key)
	}
	return nil
}

func removeMigrationRole(state migrationState, role string, key []byte, ops filesystemOps) error {
	_, path, err := roleForMigration(state, role)
	if err != nil {
		return err
	}
	return removeAuthenticatedControlFile(path, "migration folder identity", func(data []byte) error {
		return verifyMigrationRoleData(state, role, key, data)
	}, ops)
}

func removeMigrationMarker(state migrationState, key []byte, ops filesystemOps) error {
	return removeAuthenticatedControlFile(migrationStatePath(state.Root), "migration recovery record", func(data []byte) error {
		currentKey, err := reloadAuthenticatedMigrationStateData(state, data)
		if err != nil {
			return err
		}
		zeroBytes(currentKey)
		return authenticateMigrationState(state, key)
	}, ops)
}

func removeActivatedStageIdentity(state migrationState, key []byte, ops filesystemOps) error {
	return removeAuthenticatedControlFile(filepath.Join(state.Root, migrationStageIdentityFile), "activated migration stage identity", func(data []byte) error {
		return verifyStageDirectoryIdentityData(state, key, data)
	}, ops)
}

func secureMigrationRename(state migrationState, source, destination string, sourceIsDestinationWorkspace bool, destinationRole string, ops filesystemOps) error {
	return objectBoundRename(source, destination, func() error {
		key, err := reloadAuthenticatedMigrationState(state)
		if err != nil {
			return err
		}
		defer zeroBytes(key)

		if sameFilesystemPath(source, state.Root) {
			if _, err = validateDirectSibling(state.Root, source, filepath.Base(state.Root), true); err != nil {
				return err
			}
			if err = verifyMigrationWorkspace(source, state, sourceIsDestinationWorkspace, key); err != nil {
				return err
			}
		} else {
			if err = validateMigrationDirectory(state, source, true, key); err != nil {
				return err
			}
			if err = verifyMigrationWorkspace(source, state, sourceIsDestinationWorkspace, key); err != nil {
				return err
			}
		}

		if sameFilesystemPath(destination, state.Root) {
			if _, err = validateDirectSibling(state.Root, destination, filepath.Base(state.Root), false); err != nil {
				return err
			}
			if _, statErr := os.Lstat(destination); statErr == nil {
				return errors.New("the workspace activation target already exists")
			} else if !os.IsNotExist(statErr) {
				return statErr
			}
		} else {
			if destinationRole == "" {
				return errors.New("the migration rename destination has no authenticated role")
			}
			actualRole, ok := migrationRoleNameForPath(state, destination)
			if !ok || actualRole != destinationRole {
				return errors.New("the migration rename destination role does not match its authenticated path")
			}
			if err = validateMigrationDirectory(state, destination, false, key); err != nil {
				return err
			}
			if _, statErr := os.Lstat(destination); statErr == nil {
				return errors.New("the authenticated migration destination already exists")
			} else if !os.IsNotExist(statErr) {
				return statErr
			}
		}
		return nil
	}, ops.beforeRename)
}
