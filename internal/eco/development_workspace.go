package eco

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ValidateExistingWorkspaceRoot checks an existing ECO workspace without
// creating, modifying or opening it for writing. It is intended for explicit
// user-selected reopen flows before OpenVault is allowed to touch the path.
func ValidateExistingWorkspaceRoot(root string) error {
	if root == "" {
		return errors.New("empty workspace root")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve workspace root: %w", err)
	}
	info, err := os.Lstat(absolute)
	if err != nil {
		return fmt.Errorf("inspect workspace root: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("selected workspace root is not a real directory")
	}
	for _, file := range []string{"vault.key", "workspace.ecodb"} {
		entry, err := os.Lstat(filepath.Join(absolute, file))
		if err != nil {
			return fmt.Errorf("selected folder is not an existing ECO workspace: %s is missing", file)
		}
		if entry.Mode()&os.ModeSymlink != 0 || !entry.Mode().IsRegular() {
			return fmt.Errorf("selected ECO workspace has an unsafe %s entry", file)
		}
	}
	objects, err := os.Lstat(filepath.Join(absolute, "objects"))
	if err != nil {
		return errors.New("selected folder is not an existing ECO workspace: objects directory is missing")
	}
	if objects.Mode()&os.ModeSymlink != 0 || !objects.IsDir() {
		return errors.New("selected ECO workspace has an unsafe objects entry")
	}
	return nil
}

// ArchiveDevelopmentWorkspaceForCleanStart preserves an existing development
// workspace by renaming that exact owned directory to a sibling archive. It
// never deletes the workspace. OpenVault does not silently recreate an archived
// route. Use StartCleanDevelopmentWorkspace for an explicitly requested fresh
// workspace, or OpenVault(archive) to select the preserved prior state.
func ArchiveDevelopmentWorkspaceForCleanStart(root string) (string, error) {
	return archiveDevelopmentWorkspace(root, nil)
}

// boundary is an internal per-call test seam; the normal application leaves nil.
func archiveDevelopmentWorkspace(root string, boundary func(string)) (archivePath string, resultErr error) {
	attemptedArchive := ""
	defer func() {
		resultErr = withRecoveryContext("Archive for clean start", resultErr, root, attemptedArchive, "")
	}()
	if root == "" {
		return "", errors.New("empty workspace root")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve workspace root: %w", err)
	}
	if _, err := os.Lstat(absolute); os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("inspect workspace root: %w", err)
	}
	owner, err := acquireWorkspaceRootOwner(absolute)
	if err != nil {
		return "", err
	}
	closed := false
	defer func() {
		if !closed {
			if closeErr := owner.Close(); closeErr != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("release archive ownership: %w", closeErr))
			}
		}
	}()

	stamp := time.Now().UTC().Format("20060102T150405.000000000Z")
	archive := absolute + ".previous-" + stamp + "-" + developmentStateComponent(NewID("WS"))
	if _, err := os.Lstat(archive); !os.IsNotExist(err) {
		if err == nil {
			return "", errors.New("development workspace archive route already exists")
		}
		return "", fmt.Errorf("inspect development archive route: %w", err)
	}
	attemptedArchive = archive
	if err := os.Rename(absolute, archive); err != nil {
		return "", fmt.Errorf("archive development workspace: %w", err)
	}
	archivePath = archive
	if boundary != nil {
		boundary(archive)
	}
	if err := owner.retarget(archive); err != nil {
		rollbackErr := rollbackArchivedWorkspace(archive, absolute, owner)
		if rollbackErr != nil {
			return archive, fmt.Errorf("retarget archived workspace owner: %w; rollback failed: %w", err, rollbackErr)
		}
		return "", fmt.Errorf("retarget archived workspace owner: %w", err)
	}
	closed = true
	if err := owner.Close(); err != nil {
		return archive, fmt.Errorf("release archived workspace ownership: %w", err)
	}
	closed = true
	return archive, nil
}
