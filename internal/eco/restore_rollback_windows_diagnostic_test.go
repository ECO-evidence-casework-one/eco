//go:build windows

package eco

import (
	"sort"
	"testing"
)

// Keep all existing acceptance assertions intact. This distinguishes a genuine
// rollback mutation from an unstable filesystem snapshot measurement, and prints
// the precise differing values instead of calling every difference byte loss.
func TestRecoveryWindowsSnapshotDiagnostic(t *testing.T) {
	f := newRollbackFixture(t)
	before := workspaceFormatTree(t, f.root)
	immediate := workspaceFormatTree(t, f.root)
	recoveryWindowsReportDifferences(t, "repeated read without restore", before, immediate)
	stageBefore := workspaceFormatTree(t, f.stageRoot)
	f.advance(t, 1)
	if err := rollbackRestoreActivation(f.root, f.stageRoot, f.checkpoint, f.active.owner, f.stage.owner, false); err != nil {
		t.Fatal(err)
	}
	recoveryWindowsReportDifferences(t, "original after boundary-1 rollback", before, workspaceFormatTree(t, f.root))
	recoveryWindowsReportDifferences(t, "untouched stage after boundary-1 rollback", stageBefore, workspaceFormatTree(t, f.stageRoot))
}

func recoveryWindowsReportDifferences(t *testing.T, phase string, before, after map[string]string) {
	t.Helper()
	keys := make(map[string]bool)
	for key := range before {
		keys[key] = true
	}
	for key := range after {
		keys[key] = true
	}
	ordered := make([]string, 0, len(keys))
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	for _, key := range ordered {
		old, hadOld := before[key]
		now, hasNow := after[key]
		if hadOld != hasNow || old != now {
			t.Errorf("%s: entry=%q before=(exists=%t %q) after=(exists=%t %q)", phase, key, hadOld, old, hasNow, now)
		}
	}
}
