//go:build windows

package eco

import "testing"

func TestWorkspaceCreationKeyNormalisesWindowsAliases(t *testing.T) {
	want := platformWorkspaceCreationKey("Casework")
	for _, alias := range []string{"CASEWORK", "casework", "Casework.", "Casework   "} {
		if got := platformWorkspaceCreationKey(alias); got != want {
			t.Fatalf("creation key %q != %q for alias %q", got, want, alias)
		}
	}
}
