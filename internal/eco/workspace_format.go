package eco

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// decodeWorkspaceMetadata must never silently discard fields before a writable
// open or restore. Unknown fields may belong to a newer same-schema candidate.
// Older schema-1 workspaces may omit fields that were added with zero defaults.
// BuildID is provenance, not a substitute for an explicit format contract.
func decodeWorkspaceMetadata(data []byte) (Workspace, error) {
	var ws Workspace
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&ws); err != nil {
		return Workspace{}, fmt.Errorf("workspace format is not supported: %w", err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Workspace{}, errors.New("workspace metadata contains trailing or invalid data")
	}
	if ws.Schema != Schema {
		return Workspace{}, fmt.Errorf("unsupported workspace schema %d", ws.Schema)
	}
	return ws, nil
}

// preflightWorkspaceFormat is read-only. It runs before ownership acquisition
// (which can create a Linux lock file), and again under ownership before any
// key/object/metadata creation or recovery. Refusal must not initialise or
// repair an incompatible/incomplete workspace. It does not migrate formats.
func preflightWorkspaceFormat(root string) error {
	if root == "" {
		return errors.New("empty vault root")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve workspace root: %w", err)
	}
	// Apply the same real-directory rule to empty and populated roots. Clean
	// the path first so a trailing separator cannot hide a symlink from Lstat.
	if info, err := os.Lstat(absolute); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return errors.New("selected workspace root is not a real directory")
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect workspace root: %w", err)
	}
	root = absolute
	keyPath := filepath.Join(root, "vault.key")
	metaPath := filepath.Join(root, "workspace.ecodb")
	_, keyErr := os.Lstat(keyPath)
	_, metaErr := os.Lstat(metaPath)
	if os.IsNotExist(keyErr) && os.IsNotExist(metaErr) {
		// An objects directory without either control file is interrupted state,
		// not an empty workspace that can safely receive a replacement key.
		if _, err := os.Lstat(filepath.Join(root, "objects")); !os.IsNotExist(err) {
			if err != nil {
				return fmt.Errorf("inspect workspace objects: %w", err)
			}
			return errors.New("incomplete workspace: objects exist without key and metadata")
		}
		return nil
	}
	if err := ValidateExistingWorkspaceRoot(root); err != nil {
		return fmt.Errorf("existing workspace is incomplete or unsafe: %w", err)
	}
	protected, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("read existing workspace key: %w", err)
	}
	key, err := unprotectLocalKey(protected)
	if err != nil {
		return fmt.Errorf("read existing workspace key: %w", err)
	}
	defer zeroBytes(key)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return fmt.Errorf("read existing workspace metadata: %w", err)
	}
	plain, err := decryptBlob(key, metaMagic, data)
	if err != nil {
		return fmt.Errorf("workspace authentication failed: %w", err)
	}
	defer zeroBytes(plain)
	_, err = decodeWorkspaceMetadata(plain)
	return err
}
