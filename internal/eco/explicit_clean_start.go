package eco

import (
	"errors"
	"fmt"
	"path/filepath"
)

// CreateVault is explicit new-state creation, never an existing open.
// The final directory must be newly created by this call and identified
// again under exclusive ownership. Existing routes are never adopted.
func CreateVault(root string) (*Vault, error) {
	if root == "" {
		return nil, errors.New("empty workspace root")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	creation, err := acquireWorkspaceCreationOwner(absolute)
	if err != nil {
		return nil, err
	}
	defer creation.Close()
	created, err := creation.createMissingHierarchy()
	if err != nil {
		cleanupCreatedWorkspaceDirectories(created)
		return nil, err
	}
	var identity workspaceObjectIdentity
	for _, entry := range created {
		if entry.path == absolute {
			identity = entry.identity
		}
	}
	if !identity.valid() {
		cleanupCreatedWorkspaceDirectories(created)
		return nil, errors.New("new workspace directory was not created by this operation")
	}
	owner, err := acquireWorkspaceRootOwner(absolute)
	if err != nil {
		cleanupCreatedWorkspaceDirectories(created)
		return nil, err
	}
	if !identity.equal(owner.identity) {
		_ = owner.Close()
		return nil, errors.New("new workspace directory identity changed")
	}
	if err := creation.revalidate(); err != nil {
		_ = owner.Close()
		return nil, err
	}
	return openOwnedVault(owner, true)
}

// StartCleanDevelopmentWorkspace preserves old state before explicit
// creation. Incomplete operations leave preserved copies for recovery.
func StartCleanDevelopmentWorkspace(root string) (string, error) {
	return startCleanDevelopmentWorkspace(root, nil)
}
func startCleanDevelopmentWorkspace(root string, boundary func(string)) (string, error) {
	if root == "" {
		return "", errors.New("empty workspace root")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	archive, err := ArchiveDevelopmentWorkspaceForCleanStart(absolute)
	if err != nil {
		return archive, err
	}
	fail := func(cause error) (string, error) {
		return archive, fmt.Errorf("explicit clean start did not finish; retain all workspace copies (archived route %q): %w", archive, cause)
	}
	if boundary != nil {
		boundary("reset_archived")
	}
	fresh, err := CreateVault(absolute)
	if err != nil {
		return fail(err)
	}
	defer fresh.Close()
	if boundary != nil {
		boundary("reset_created")
	}
	if err := fresh.Close(); err != nil {
		return fail(err)
	}
	return archive, nil
}
