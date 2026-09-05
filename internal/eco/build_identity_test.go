package eco

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDevelopmentWorkspaceRootSeparatesSourceCandidates(t *testing.T) {
	base := t.TempDir()
	first := DevelopmentWorkspaceRoot(base, BuildID, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Schema)
	second := DevelopmentWorkspaceRoot(base, BuildID, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Schema)
	if first == second {
		t.Fatalf("distinct source candidates shared one development root: %q", first)
	}
	if !strings.Contains(first, filepath.Join("Development", "schema-1")) {
		t.Fatalf("development root does not include schema boundary: %q", first)
	}
	if !strings.HasSuffix(first, "aaaaaaaaaaaa") {
		t.Fatalf("source commit was not reduced to a stable 12-character identity: %q", first)
	}
}

func TestDevelopmentWorkspaceRootSeparatesSchemasAndBuilds(t *testing.T) {
	base := t.TempDir()
	a := DevelopmentWorkspaceRoot(base, "ECO-V25-A", "0123456789abcdef", 1)
	b := DevelopmentWorkspaceRoot(base, "ECO-V25-B", "0123456789abcdef", 1)
	c := DevelopmentWorkspaceRoot(base, "ECO-V25-A", "0123456789abcdef", 2)
	if a == b || a == c || b == c {
		t.Fatalf("build/schema identity collision: a=%q b=%q c=%q", a, b, c)
	}
}

func TestDevelopmentWorkspaceRootSanitisesUnsafeComponents(t *testing.T) {
	root := DevelopmentWorkspaceRoot(t.TempDir(), `ECO:V25/unsafe\\name`, `abc:def/ghi`, 1)
	if strings.Contains(filepath.Base(filepath.Dir(root)), ":") || strings.Contains(filepath.Base(root), ":") {
		t.Fatalf("unsafe path punctuation survived sanitisation: %q", root)
	}
}

func TestDefaultDevelopmentWorkspaceRootUsesEmbeddedIdentity(t *testing.T) {
	old := SourceCommit
	SourceCommit = "fedcba9876543210fedcba9876543210fedcba98"
	t.Cleanup(func() { SourceCommit = old })
	root := DefaultDevelopmentWorkspaceRoot(t.TempDir())
	if !strings.HasSuffix(root, "fedcba987654") {
		t.Fatalf("default root did not use embedded source identity: %q", root)
	}
}
