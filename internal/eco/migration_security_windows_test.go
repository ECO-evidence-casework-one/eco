//go:build windows

package eco

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationRecoveryRejectsJunctionParticipants(t *testing.T) {
	for _, role := range []string{"stage", "checkpoint"} {
		t.Run(role, func(t *testing.T) {
			root, _, current, state := interruptedMigrationState(t, migrationStageReady)
			participant := state.Stage
			if role == "checkpoint" {
				participant = state.Checkpoint
			}
			kept := participant + ".kept"
			if err := os.Rename(participant, kept); err != nil {
				t.Fatal(err)
			}
			external := t.TempDir()
			sentinel := filepath.Join(external, "keep.txt")
			content := []byte("junction target must remain exact")
			if err := os.WriteFile(sentinel, content, 0600); err != nil {
				t.Fatal(err)
			}
			createTestJunction(t, participant, external)
			if _, _, err := RecoverWorkspace(root, current); err == nil {
				t.Fatal("migration recovery accepted a junction participant")
			}
			assertFileExact(t, sentinel, content)
			if _, err := os.Stat(filepath.Join(kept, "workspace.ecodb")); err != nil {
				t.Fatalf("authentic migration participant was lost: %v", err)
			}
		})
	}
}
