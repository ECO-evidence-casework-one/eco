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

func TestWorkspaceCreationKeyNormalisesNestedWindowsAliases(t *testing.T) {
	want := platformWorkspaceCreationKey(`Application\Candidate\Casework`)
	for _, alias := range []string{`APPLICATION\candidate\CASEWORK`, `Application/Candidate/Casework.`, `Application\Candidate.\Casework   `} {
		if got := platformWorkspaceCreationKey(alias); got != want {
			t.Fatalf("nested creation key %q != %q for alias %q", got, want, alias)
		}
	}
}
