package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenWorkspaceNeverCreates(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "missing")
	var svc Service
	if err := svc.OpenWorkspace(root); err == nil {
		t.Fatal("open unexpectedly created a missing workspace")
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("missing workspace route changed after open: %v", err)
	}
}

func TestCreateWorkspaceRefusesExistingRoute(t *testing.T) {
	root := filepath.Join(t.TempDir(), "existing")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	var svc Service
	if err := svc.CreateWorkspace(root); err == nil {
		t.Fatal("create unexpectedly adopted an existing route")
	}
	after, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != len(after) {
		t.Fatalf("refused create mutated existing route: before=%d after=%d", len(before), len(after))
	}
}

func TestCloseWorkspaceReleasesOwnership(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	var first, second Service
	if err := first.CreateWorkspace(root); err != nil {
		t.Fatal(err)
	}
	if err := second.OpenWorkspace(root); err == nil {
		t.Fatal("second service opened workspace while first still owned it")
	}
	if err := first.CloseWorkspace(); err != nil {
		t.Fatal(err)
	}
	if err := second.OpenWorkspace(root); err != nil {
		t.Fatalf("workspace did not reopen after close released ownership: %v", err)
	}
	if err := second.CloseWorkspace(); err != nil {
		t.Fatal(err)
	}
}

func TestSnapshotSummaryIsReadOnly(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	var svc Service
	if err := svc.CreateWorkspace(root); err != nil {
		t.Fatal(err)
	}
	defer svc.CloseWorkspace()

	meta := filepath.Join(root, "workspace.ecodb")
	before, err := os.ReadFile(meta)
	if err != nil {
		t.Fatal(err)
	}
	beforeInfo, err := os.Stat(meta)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		if _, err := svc.SnapshotSummary(); err != nil {
			t.Fatal(err)
		}
	}
	after, err := os.ReadFile(meta)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(meta)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("read-only summaries changed workspace metadata bytes")
	}
	if !beforeInfo.ModTime().Equal(afterInfo.ModTime()) || beforeInfo.Mode() != afterInfo.Mode() {
		t.Fatal("read-only summaries changed workspace metadata filesystem state")
	}
}

func TestCreateMatterAndReopenThroughService(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	var svc Service
	if err := svc.CreateWorkspace(root); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreateMatter("Service boundary matter"); err != nil {
		t.Fatal(err)
	}
	summary, err := svc.SnapshotSummary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.MatterCount != 1 || len(summary.Matters) != 1 || summary.Matters[0].Title != "Service boundary matter" {
		t.Fatalf("unexpected service summary after matter creation: %+v", summary)
	}
	if err := svc.CloseWorkspace(); err != nil {
		t.Fatal(err)
	}
	if err := svc.OpenWorkspace(root); err != nil {
		t.Fatal(err)
	}
	defer svc.CloseWorkspace()
	reopened, err := svc.SnapshotSummary()
	if err != nil {
		t.Fatal(err)
	}
	if reopened.MatterCount != 1 || reopened.Matters[0].Title != "Service boundary matter" {
		t.Fatalf("reopened service lost matter state: %+v", reopened)
	}
}

func TestServiceLifecycleErrors(t *testing.T) {
	var svc Service
	if _, err := svc.SnapshotSummary(); !errors.Is(err, ErrNoWorkspaceOpen) {
		t.Fatalf("snapshot without workspace error=%v", err)
	}
	if err := svc.CreateMatter("blocked"); !errors.Is(err, ErrNoWorkspaceOpen) {
		t.Fatalf("matter without workspace error=%v", err)
	}
	root := filepath.Join(t.TempDir(), "workspace")
	if err := svc.CreateWorkspace(root); err != nil {
		t.Fatal(err)
	}
	if err := svc.OpenWorkspace(root); !errors.Is(err, ErrWorkspaceAlreadyOpen) {
		t.Fatalf("second open error=%v", err)
	}
	if err := svc.CloseWorkspace(); err != nil {
		t.Fatal(err)
	}
	if err := svc.CloseWorkspace(); err != nil {
		t.Fatalf("idempotent close failed: %v", err)
	}
}
