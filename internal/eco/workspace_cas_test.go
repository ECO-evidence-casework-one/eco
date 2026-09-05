package eco

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceCASRevisionAdvancesAndOwnerTxnChangesOnReopen(t *testing.T) {
	root := filepath.Join(t.TempDir(), "vault")
	first, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	firstSnapshot := first.Snapshot()
	if firstSnapshot.Revision == 0 || firstSnapshot.LastOwnerTxn == "" {
		t.Fatalf("initial persisted CAS control missing: revision=%d owner=%q", firstSnapshot.Revision, firstSnapshot.LastOwnerTxn)
	}
	firstOwner := firstSnapshot.LastOwnerTxn
	firstRevision := firstSnapshot.Revision
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	loaded := reopened.Snapshot()
	if loaded.Revision != firstRevision || loaded.LastOwnerTxn != firstOwner {
		t.Fatalf("reopen changed persisted CAS state before save: got rev=%d owner=%q want rev=%d owner=%q", loaded.Revision, loaded.LastOwnerTxn, firstRevision, firstOwner)
	}
	if err := reopened.Save(); err != nil {
		t.Fatal(err)
	}
	after := reopened.Snapshot()
	if after.Revision <= firstRevision {
		t.Fatalf("revision did not advance: before=%d after=%d", firstRevision, after.Revision)
	}
	if after.LastOwnerTxn == "" || after.LastOwnerTxn == firstOwner {
		t.Fatalf("new owner transaction was not persisted: old=%q new=%q", firstOwner, after.LastOwnerTxn)
	}
}

func TestWorkspaceCASRejectsOlderValidEncryptedMetadata(t *testing.T) {
	root := filepath.Join(t.TempDir(), "vault")
	v, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	path := filepath.Join(root, "workspace.ecodb")
	older, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, older, 0600); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); !errors.Is(err, ErrWorkspaceStale) {
		t.Fatalf("stale valid metadata save error=%v", err)
	}
	stillOlder, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if workspaceMetadataDigest(stillOlder) != workspaceMetadataDigest(older) {
		t.Fatal("stale metadata was overwritten after CAS rejection")
	}
}

func TestWorkspaceCASRejectsSameRevisionDifferentAuthenticatedMetadata(t *testing.T) {
	root := filepath.Join(t.TempDir(), "vault")
	v, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	path := filepath.Join(root, "workspace.ecodb")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := decryptBlob(v.key, metaMagic, data)
	if err != nil {
		t.Fatal(err)
	}
	var ws Workspace
	if err := json.Unmarshal(plain, &ws); err != nil {
		t.Fatal(err)
	}
	ws.SelectedPage = "tampered-but-authenticated"
	changedPlain, err := json.MarshalIndent(ws, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	changed, err := encryptBlob(v.key, metaMagic, changedPlain)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, changed, 0600); err != nil {
		t.Fatal(err)
	}
	if err := v.Save(); !errors.Is(err, ErrWorkspaceStale) {
		t.Fatalf("same-revision changed metadata save error=%v", err)
	}
}
