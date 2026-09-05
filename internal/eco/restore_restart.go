package eco

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var ErrWorkspaceRecoveryRequired = errors.New("workspace recovery requires an explicit choice")

// CheckWorkspaceRecoveryState is read-only. It prevents a restore interrupted
// between the two directory renames from becoming a silent new workspace on
// restart. Checkpoint names are warning evidence only, never trusted paths for
// automatic activation, deletion or migration. A healthy populated workspace
// can still open with its historical pre-restore checkpoints retained.
// A retained .previous-* reset archive likewise requires an explicit choice.
func CheckWorkspaceRecoveryState(root string) error {
	if root == "" {
		return errors.New("empty workspace root")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve recovery-check root: %w", err)
	}
	if _, err := os.Lstat(filepath.Join(absolute, "workspace.ecodb")); err == nil {
		// Normal format/ownership/authentication checks still govern this open.
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect workspace metadata before recovery: %w", err)
	}
	parent, err := os.Open(filepath.Dir(absolute))
	if os.IsNotExist(err) {
		return nil // New candidate whose parent hierarchy does not exist yet.
	}
	if err != nil {
		return fmt.Errorf("inspect recovery checkpoint directory: %w", err)
	}
	defer parent.Close()
	prefixes := []string{filepath.Base(absolute) + ".pre-restore-", filepath.Base(absolute) + ".previous-"}
	for {
		entries, readErr := parent.ReadDir(128)
		for _, entry := range entries {
			name := entry.Name()
			matches := false
			for _, prefix := range prefixes {
				if runtime.GOOS == "windows" {
					matches = matches || strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix))
				} else {
					matches = matches || strings.HasPrefix(name, prefix)
				}
			}
			if matches {
				// Even a link or incomplete entry blocks initialisation. The
				// explicit Open-existing flow must validate the selected folder.
				return fmt.Errorf("%w: the active workspace has no committed metadata, but a retained workspace checkpoint remains at %q. Nothing was created or replaced. Preserve all copies and use Open existing to select and verify the original checkpoint", ErrWorkspaceRecoveryRequired, filepath.Join(filepath.Dir(absolute), name))
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return fmt.Errorf("read recovery checkpoint directory: %w", readErr)
		}
	}
}
