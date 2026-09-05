from pathlib import Path


def replace_once(path: str, old: str, new: str) -> None:
    p = Path(path)
    text = p.read_text(encoding="utf-8")
    if text.count(old) != 1:
        raise SystemExit(f"expected exactly one patch anchor in {path}, found {text.count(old)}")
    p.write_text(text.replace(old, new, 1), encoding="utf-8")


replace_once(
    "internal/eco/types.go",
    '''type Workspace struct {\n\tSchema        int                  `json:"schema"`\n\tBuildID       string               `json:"build_id"`\n''',
    '''type Workspace struct {\n\tSchema        int                  `json:"schema"`\n\tRevision      uint64               `json:"revision,omitempty"`\n\tLastOwnerTxn  string               `json:"last_owner_txn,omitempty"`\n\tBuildID       string               `json:"build_id"`\n''',
)

replace_once(
    "internal/eco/vault.go",
    '''var ErrVaultClosed = errors.New("vault is closed")\n''',
    '''var (\n\tErrVaultClosed    = errors.New("vault is closed")\n\tErrWorkspaceStale = errors.New("workspace metadata changed since it was loaded")\n)\n''',
)

replace_once(
    "internal/eco/vault.go",
    '''\tkey       []byte\n\towner     *workspaceOwnerLease\n\tclosed    bool\n''',
    '''\tkey                  []byte\n\towner                *workspaceOwnerLease\n\townerTxn             string\n\tpersistedRevision    uint64\n\tpersistedMetaSHA256  string\n\tpersistedChangeHead  string\n\tpersistedOwnerTxn    string\n\tclosed               bool\n''',
)

replace_once(
    "internal/eco/vault.go",
    '''\tv := &Vault{Root: root, Objects: filepath.Join(root, "objects"), owner: owner}\n''',
    '''\tv := &Vault{Root: root, Objects: filepath.Join(root, "objects"), owner: owner, ownerTxn: NewID("OWNER")}\n''',
)

replace_once(
    "internal/eco/vault.go",
    '''\tzeroBytes(v.key)\n\tv.key = nil\n\tv.mu.Unlock()\n''',
    '''\tzeroBytes(v.key)\n\tv.key = nil\n\tv.ownerTxn = ""\n\tv.persistedRevision = 0\n\tv.persistedMetaSHA256 = ""\n\tv.persistedChangeHead = ""\n\tv.persistedOwnerTxn = ""\n\tv.mu.Unlock()\n''',
)

replace_once(
    "internal/eco/vault.go",
    '''\tif os.IsNotExist(err) {\n\t\tv.Workspace = newWorkspace()\n\t\treturn v.Save()\n\t}\n''',
    '''\tif os.IsNotExist(err) {\n\t\tv.persistedRevision = 0\n\t\tv.persistedMetaSHA256 = ""\n\t\tv.persistedChangeHead = ""\n\t\tv.persistedOwnerTxn = ""\n\t\tv.Workspace = newWorkspace()\n\t\treturn v.Save()\n\t}\n''',
)

replace_once(
    "internal/eco/vault.go",
    '''\tif ws.Schema != Schema {\n\t\treturn fmt.Errorf("unsupported workspace schema %d", ws.Schema)\n\t}\n\tv.Workspace = ws\n\treturn nil\n}\n''',
    '''\tif ws.Schema != Schema {\n\t\treturn fmt.Errorf("unsupported workspace schema %d", ws.Schema)\n\t}\n\tv.persistedRevision = ws.Revision\n\tv.persistedMetaSHA256 = workspaceMetadataDigest(data)\n\tv.persistedChangeHead = workspaceChangeHead(ws)\n\tv.persistedOwnerTxn = ws.LastOwnerTxn\n\tv.Workspace = ws\n\treturn nil\n}\n''',
)

old_save = '''func (v *Vault) saveUnlocked() error {\n\tif err := v.ensureOpenUnlocked(); err != nil {\n\t\treturn err\n\t}\n\tv.Workspace.UpdatedAt = time.Now().UTC()\n\tv.Workspace.BuildID = BuildID\n\tplain, err := json.MarshalIndent(v.Workspace, "", "  ")\n\tif err != nil {\n\t\treturn err\n\t}\n\tenc, err := encryptBlob(v.key, metaMagic, plain)\n\tif err != nil {\n\t\treturn err\n\t}\n\tpath := filepath.Join(v.Root, "workspace.ecodb")\n\ttmp := path + ".tmp"\n\tif err := os.WriteFile(tmp, enc, 0600); err != nil {\n\t\treturn err\n\t}\n\tif err := os.Rename(tmp, path); err != nil {\n\t\treturn err\n\t}\n\treturn nil\n}\n'''

new_save = '''func (v *Vault) saveUnlocked() error {\n\tif err := v.ensureOpenUnlocked(); err != nil {\n\t\treturn err\n\t}\n\tif err := v.verifyWorkspaceCASUnlocked(); err != nil {\n\t\treturn err\n\t}\n\n\toldRevision := v.Workspace.Revision\n\toldOwnerTxn := v.Workspace.LastOwnerTxn\n\toldUpdatedAt := v.Workspace.UpdatedAt\n\toldBuildID := v.Workspace.BuildID\n\trestoreControl := func() {\n\t\tv.Workspace.Revision = oldRevision\n\t\tv.Workspace.LastOwnerTxn = oldOwnerTxn\n\t\tv.Workspace.UpdatedAt = oldUpdatedAt\n\t\tv.Workspace.BuildID = oldBuildID\n\t}\n\n\tbaseRevision := v.persistedRevision\n\tif v.Workspace.Revision > baseRevision {\n\t\tbaseRevision = v.Workspace.Revision\n\t}\n\tif baseRevision == ^uint64(0) {\n\t\treturn errors.New("workspace revision exhausted")\n\t}\n\tv.Workspace.Revision = baseRevision + 1\n\tv.Workspace.LastOwnerTxn = v.ownerTxn\n\tv.Workspace.UpdatedAt = time.Now().UTC()\n\tv.Workspace.BuildID = BuildID\n\tplain, err := json.MarshalIndent(v.Workspace, "", "  ")\n\tif err != nil {\n\t\trestoreControl()\n\t\treturn err\n\t}\n\tenc, err := encryptBlob(v.key, metaMagic, plain)\n\tif err != nil {\n\t\trestoreControl()\n\t\treturn err\n\t}\n\tpath := filepath.Join(v.Root, "workspace.ecodb")\n\ttmp := path + ".tmp"\n\tif err := os.WriteFile(tmp, enc, 0600); err != nil {\n\t\trestoreControl()\n\t\treturn err\n\t}\n\tif err := os.Rename(tmp, path); err != nil {\n\t\t_ = os.Remove(tmp)\n\t\trestoreControl()\n\t\treturn err\n\t}\n\tv.persistedRevision = v.Workspace.Revision\n\tv.persistedMetaSHA256 = workspaceMetadataDigest(enc)\n\tv.persistedChangeHead = workspaceChangeHead(v.Workspace)\n\tv.persistedOwnerTxn = v.Workspace.LastOwnerTxn\n\treturn nil\n}\n\nfunc (v *Vault) verifyWorkspaceCASUnlocked() error {\n\tpath := filepath.Join(v.Root, "workspace.ecodb")\n\tdata, err := os.ReadFile(path)\n\tif os.IsNotExist(err) {\n\t\tif v.persistedMetaSHA256 == "" && v.persistedRevision == 0 && v.persistedChangeHead == "" && v.persistedOwnerTxn == "" {\n\t\t\treturn nil\n\t\t}\n\t\treturn fmt.Errorf("%w: workspace metadata disappeared", ErrWorkspaceStale)\n\t}\n\tif err != nil {\n\t\treturn fmt.Errorf("read workspace metadata for compare-and-swap: %w", err)\n\t}\n\tdigest := workspaceMetadataDigest(data)\n\tif v.persistedMetaSHA256 == "" || digest != v.persistedMetaSHA256 {\n\t\treturn fmt.Errorf("%w: encrypted metadata digest changed", ErrWorkspaceStale)\n\t}\n\tplain, err := decryptBlob(v.key, metaMagic, data)\n\tif err != nil {\n\t\treturn fmt.Errorf("%w: current workspace authentication failed: %v", ErrWorkspaceStale, err)\n\t}\n\tvar current Workspace\n\tif err := json.Unmarshal(plain, &current); err != nil {\n\t\treturn fmt.Errorf("%w: current workspace format is invalid: %v", ErrWorkspaceStale, err)\n\t}\n\tif current.Revision != v.persistedRevision {\n\t\treturn fmt.Errorf("%w: revision changed from %d to %d", ErrWorkspaceStale, v.persistedRevision, current.Revision)\n\t}\n\tif current.LastOwnerTxn != v.persistedOwnerTxn {\n\t\treturn fmt.Errorf("%w: owner transaction changed", ErrWorkspaceStale)\n\t}\n\tif head := workspaceChangeHead(current); head != v.persistedChangeHead {\n\t\treturn fmt.Errorf("%w: audit-chain head changed", ErrWorkspaceStale)\n\t}\n\treturn nil\n}\n\nfunc workspaceMetadataDigest(data []byte) string {\n\tsum := sha256.Sum256(data)\n\treturn hex.EncodeToString(sum[:])\n}\n\nfunc workspaceChangeHead(ws Workspace) string {\n\tif len(ws.Changes) == 0 {\n\t\treturn ""\n\t}\n\treturn ws.Changes[0].Hash\n}\n'''
replace_once("internal/eco/vault.go", old_save, new_save)

replace_once(
    "internal/eco/backup.go",
    '''\tv.Workspace = stage.Workspace\n\tstageActivated = true\n\t_ = oldOwner.Close()\n''',
    '''\tv.Workspace = stage.Workspace\n\tv.ownerTxn = stage.ownerTxn\n\tv.persistedRevision = stage.persistedRevision\n\tv.persistedMetaSHA256 = stage.persistedMetaSHA256\n\tv.persistedChangeHead = stage.persistedChangeHead\n\tv.persistedOwnerTxn = stage.persistedOwnerTxn\n\tstage.ownerTxn = ""\n\tstage.persistedRevision = 0\n\tstage.persistedMetaSHA256 = ""\n\tstage.persistedChangeHead = ""\n\tstage.persistedOwnerTxn = ""\n\tstageActivated = true\n\t_ = oldOwner.Close()\n''',
)

Path("internal/eco/workspace_cas_test.go").write_text(r'''package eco

import (
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
'''.replace('"testing"\n)', '"testing"\n    "encoding/json"\n)'), encoding="utf-8")

Path("docs/control/WORKSPACE_METADATA_CAS_20260905.md").write_text('''# Workspace metadata compare-and-swap — 5 September 2026\n\n**Issue:** #4 clean, explicit workspace state  \n**Baseline:** current `main` after PR #125  \n**Scope:** stale authenticated metadata rejection for the live writable Vault\n\n## Control\n\nEach persisted workspace now carries a monotonically advancing `revision` and the identifier of the owner transaction that wrote it. A live Vault also retains the SHA-256 of the exact encrypted `workspace.ecodb` bytes it loaded or wrote and the hash-chain head of the persisted `ChangeRecord` ledger.\n\nBefore an authenticated metadata save, ECO revalidates the object-bound workspace owner and then requires the on-disk metadata to match the expected:\n\n- encrypted metadata SHA-256;\n- workspace revision;\n- last owner transaction;\n- audit-chain head.\n\nIf any of those values changed, disappeared or fail authentication, the save returns `ErrWorkspaceStale` and does not overwrite the unexpected metadata. A successful write advances the revision and records the current owner transaction.\n\nPortable restore transfers the staged Vault's CAS state together with its owner/key/workspace when the staged object becomes active, so the next ordinary save continues from the activated metadata rather than from stale pre-restore state.\n\n## Compatibility\n\nLegacy schema-1 workspaces that do not contain `revision` or `last_owner_txn` load as revision zero and are upgraded on their next successful authenticated save. Unknown JSON fields remain compatible with older schema-1 readers, so this slice does not change the workspace schema number.\n\n## Limits\n\nIssue #4 remains open. This slice does not yet implement the user-facing first-run/open/reset chooser, complete migration/upgrade/downgrade evidence, all interruption-point rollback proofs, or the final read-only/writer policy.\n''', encoding="utf-8")
