package eco

import (
	"errors"
	"fmt"
	"path/filepath"
)

// StartCleanDevelopmentWorkspace is an explicit user-requested operation.
// It preserves prior state, creates and owns a new directory, then writes
// the fresh workspace before returning. A process stop cannot grant a
// future ordinary OpenVault permission to silently initialise old state.
func StartCleanDevelopmentWorkspace(root string) (string, error) {
	return startCleanDevelopmentWorkspace(root, nil)
}

// boundary is a per-call internal test observation seam, never set by UI.
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
	creation, err := acquireWorkspaceCreationOwner(absolute)
	if err != nil {
		return fail(err)
	}
	defer creation.Close()
	created, err := creation.createMissingHierarchy()
	if err != nil {
		cleanupCreatedWorkspaceDirectories(created)
		return fail(err)
	}
	var identity workspaceObjectIdentity
	for _, entry := range created {
		if entry.path == absolute {
			identity = entry.identity
		}
	}
	if !identity.valid() {
		cleanupCreatedWorkspaceDirectories(created)
		return fail(errors.New("clean-start route was not created by this operation"))
	}
	owner, err := acquireWorkspaceRootOwner(absolute)
	if err != nil {
		cleanupCreatedWorkspaceDirectories(created)
		return fail(err)
	}
	if !identity.equal(owner.identity) {
		_ = owner.Close()
		return fail(errors.New("new clean-start directory identity changed"))
	}
	if err := creation.revalidate(); err != nil {
		_ = owner.Close()
		return fail(err)
	}
	fresh, err := openOwnedVault(owner, true)
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
