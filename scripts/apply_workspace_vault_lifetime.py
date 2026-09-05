from pathlib import Path


def replace_once(text: str, old: str, new: str, label: str) -> str:
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"{label}: expected exactly one anchor, found {count}")
    return text.replace(old, new, 1)


vault_path = Path("internal/eco/vault.go")
vault = vault_path.read_text(encoding="utf-8")

vault = replace_once(
    vault,
    '''type Vault struct {\n\tRoot      string\n\tObjects   string\n\tkey       []byte\n\topMu      sync.RWMutex\n\tmu        sync.Mutex\n\tWorkspace Workspace\n}\n\nfunc OpenVault(root string) (*Vault, error) {\n\tif root == "" {\n\t\treturn nil, errors.New("empty vault root")\n\t}\n\tobjects := filepath.Join(root, "objects")\n\tif err := os.MkdirAll(objects, 0700); err != nil {\n\t\treturn nil, err\n\t}\n\tkey, err := loadOrCreateMasterKey(filepath.Join(root, "vault.key"))\n\tif err != nil {\n\t\treturn nil, fmt.Errorf("vault key: %w", err)\n\t}\n\tv := &Vault{Root: root, Objects: objects, key: key}\n\tif err := v.loadWorkspace(); err != nil {\n\t\treturn nil, err\n\t}\n\tif err := cleanupInterruptedReadingCopies(root); err != nil {\n\t\treturn nil, fmt.Errorf("clean interrupted derived reading state: %w", err)\n\t}\n\tif err := v.recoverPreservations(); err != nil {\n\t\treturn nil, fmt.Errorf("recover preservation state: %w", err)\n\t}\n\treturn v, nil\n}\n''',
    '''var ErrVaultClosed = errors.New("vault is closed")\n\ntype Vault struct {\n\tRoot      string\n\tObjects   string\n\tkey       []byte\n\towner     *workspaceOwnerLease\n\tclosed    bool\n\topMu      sync.RWMutex\n\tmu        sync.Mutex\n\tWorkspace Workspace\n}\n\nfunc OpenVault(root string) (*Vault, error) {\n\tif root == "" {\n\t\treturn nil, errors.New("empty vault root")\n\t}\n\towner, err := acquireOrCreateWorkspaceRootOwner(root)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\troot = owner.root\n\tv := &Vault{Root: root, Objects: filepath.Join(root, "objects"), owner: owner}\n\topened := false\n\tdefer func() {\n\t\tif opened {\n\t\t\treturn\n\t\t}\n\t\tzeroBytes(v.key)\n\t\tv.key = nil\n\t\t_ = owner.Close()\n\t}()\n\tif err := os.MkdirAll(v.Objects, 0700); err != nil {\n\t\treturn nil, err\n\t}\n\tkey, err := loadOrCreateMasterKey(filepath.Join(root, "vault.key"))\n\tif err != nil {\n\t\treturn nil, fmt.Errorf("vault key: %w", err)\n\t}\n\tv.key = key\n\tif err := v.loadWorkspace(); err != nil {\n\t\treturn nil, err\n\t}\n\tif err := cleanupInterruptedReadingCopies(root); err != nil {\n\t\treturn nil, fmt.Errorf("clean interrupted derived reading state: %w", err)\n\t}\n\tif err := v.recoverPreservations(); err != nil {\n\t\treturn nil, fmt.Errorf("recover preservation state: %w", err)\n\t}\n\topened = true\n\treturn v, nil\n}\n\nfunc (v *Vault) Close() error {\n\tif v == nil {\n\t\treturn nil\n\t}\n\tv.opMu.Lock()\n\tdefer v.opMu.Unlock()\n\tv.mu.Lock()\n\tif v.closed {\n\t\tv.mu.Unlock()\n\t\treturn nil\n\t}\n\tv.closed = true\n\towner := v.owner\n\tv.owner = nil\n\tzeroBytes(v.key)\n\tv.key = nil\n\tv.mu.Unlock()\n\tif owner != nil {\n\t\treturn owner.Close()\n\t}\n\treturn nil\n}\n\nfunc (v *Vault) ensureOpenUnlocked() error {\n\tif v == nil || v.closed || v.owner == nil {\n\t\treturn ErrVaultClosed\n\t}\n\tif err := v.owner.revalidate(); err != nil {\n\t\treturn fmt.Errorf("workspace ownership is invalid: %w", err)\n\t}\n\treturn nil\n}\n''',
    "Vault struct/OpenVault",
)

vault = replace_once(
    vault,
    '''func (v *Vault) saveUnlocked() error {\n\tv.Workspace.UpdatedAt = time.Now().UTC()\n''',
    '''func (v *Vault) saveUnlocked() error {\n\tif err := v.ensureOpenUnlocked(); err != nil {\n\t\treturn err\n\t}\n\tv.Workspace.UpdatedAt = time.Now().UTC()\n''',
    "save ownership validation",
)

vault_path.write_text(vault, encoding="utf-8")

backup_path = Path("internal/eco/backup.go")
backup = backup_path.read_text(encoding="utf-8")
backup = replace_once(
    backup,
    '''\t"path/filepath"\n\t"time"\n''',
    '''\t"path/filepath"\n\t"strings"\n\t"time"\n''',
    "backup strings import",
)
backup = replace_once(
    backup,
    '''\tstageActivated := false\n\tdefer func() {\n\t\tif !stageActivated {\n\t\t\t_ = os.RemoveAll(stageRoot)\n\t\t}\n\t}()\n''',
    '''\tstageActivated := false\n\tdefer func() {\n\t\t_ = stage.Close()\n\t\tif !stageActivated {\n\t\t\t_ = os.RemoveAll(stageRoot)\n\t\t}\n\t}()\n''',
    "stage close defer",
)
old_activation = '''\t// Activation is the only exclusive phase. Read/import/verify/backup file\n\t// operations finish first, and metadata writers are blocked by v.mu.\n\tv.opMu.Lock()\n\tdefer v.opMu.Unlock()\n\tv.mu.Lock()\n\tdefer v.mu.Unlock()\n\n\tparent := filepath.Dir(v.Root)\n\tbase := filepath.Base(v.Root)\n\tpreRestore := filepath.Join(parent, base+".pre-restore-"+time.Now().UTC().Format("20060102T150405Z"))\n\tif err = os.Rename(v.Root, preRestore); err != nil {\n\t\treturn RestoreReceipt{}, fmt.Errorf("could not create pre-restore checkpoint: %w", err)\n\t}\n\tif err = os.Rename(stageRoot, v.Root); err != nil {\n\t\t_ = os.Rename(preRestore, v.Root)\n\t\treturn RestoreReceipt{}, fmt.Errorf("could not activate restored vault; original was rolled back: %w", err)\n\t}\n\tstageActivated = true\n\n\tnewVault, err := OpenVault(v.Root)\n\tif err != nil {\n\t\tfailedPath := stageRoot + ".failed"\n\t\t_ = os.Rename(v.Root, failedPath)\n\t\t_ = os.Rename(preRestore, v.Root)\n\t\treturn RestoreReceipt{}, fmt.Errorf("restored vault could not reopen; original was rolled back: %w", err)\n\t}\n\n\tv.Objects = newVault.Objects\n\tzeroBytes(v.key)\n\tv.key = newVault.key\n\tv.Workspace = newVault.Workspace\n'''
new_activation = '''\t// Activation is the only exclusive phase. The active Vault retains an\n\t// object-bound owner throughout the swap while the staged Vault retains its\n\t// own owner. Ownership follows the directory objects across renames and is\n\t// transferred only after the activated staged vault re-verifies.\n\tv.opMu.Lock()\n\tdefer v.opMu.Unlock()\n\tv.mu.Lock()\n\tdefer v.mu.Unlock()\n\tif err = v.ensureOpenUnlocked(); err != nil {\n\t\treturn RestoreReceipt{}, err\n\t}\n\tif stage.owner == nil {\n\t\treturn RestoreReceipt{}, errors.New("staged vault ownership is missing")\n\t}\n\n\tactiveRoot := v.Root\n\tparent := filepath.Dir(activeRoot)\n\tbase := filepath.Base(activeRoot)\n\tpreRestore := filepath.Join(parent, base+".pre-restore-"+time.Now().UTC().Format("20060102T150405Z"))\n\toldOwner := v.owner\n\tif err = os.Rename(activeRoot, preRestore); err != nil {\n\t\treturn RestoreReceipt{}, fmt.Errorf("could not create pre-restore checkpoint: %w", err)\n\t}\n\tif err = oldOwner.retarget(preRestore); err != nil {\n\t\trollbackErr := rollbackRestoreActivation(activeRoot, stageRoot, preRestore, oldOwner, stage.owner, false)\n\t\treturn RestoreReceipt{}, restoreActivationFailure(fmt.Errorf("active workspace ownership could not follow pre-restore checkpoint: %w", err), rollbackErr)\n\t}\n\tif err = os.Rename(stageRoot, activeRoot); err != nil {\n\t\trollbackErr := rollbackRestoreActivation(activeRoot, stageRoot, preRestore, oldOwner, stage.owner, false)\n\t\treturn RestoreReceipt{}, restoreActivationFailure(fmt.Errorf("could not activate restored vault: %w", err), rollbackErr)\n\t}\n\tif err = stage.owner.retarget(activeRoot); err != nil {\n\t\trollbackErr := rollbackRestoreActivation(activeRoot, stageRoot, preRestore, oldOwner, stage.owner, true)\n\t\treturn RestoreReceipt{}, restoreActivationFailure(fmt.Errorf("staged workspace ownership could not follow activation: %w", err), rollbackErr)\n\t}\n\tstage.Root = activeRoot\n\tstage.Objects = filepath.Join(activeRoot, "objects")\n\tif err = stage.owner.revalidate(); err != nil {\n\t\trollbackErr := rollbackRestoreActivation(activeRoot, stageRoot, preRestore, oldOwner, stage.owner, true)\n\t\treturn RestoreReceipt{}, restoreActivationFailure(fmt.Errorf("activated workspace ownership failed revalidation: %w", err), rollbackErr)\n\t}\n\tif err = verifyStagedVault(stage); err != nil {\n\t\trollbackErr := rollbackRestoreActivation(activeRoot, stageRoot, preRestore, oldOwner, stage.owner, true)\n\t\treturn RestoreReceipt{}, restoreActivationFailure(fmt.Errorf("activated restored vault failed source verification: %w", err), rollbackErr)\n\t}\n\n\tv.Root = activeRoot\n\tv.Objects = stage.Objects\n\tv.owner = stage.owner\n\tstage.owner = nil\n\tzeroBytes(v.key)\n\tv.key = stage.key\n\tstage.key = nil\n\tv.Workspace = stage.Workspace\n\tstageActivated = true\n\t_ = oldOwner.Close()\n'''
backup = replace_once(backup, old_activation, new_activation, "restore activation handoff")
backup = replace_once(
    backup,
    '''type countingReader struct {\n''',
    '''func restoreActivationFailure(cause, rollbackErr error) error {\n\tif rollbackErr == nil {\n\t\treturn cause\n\t}\n\treturn fmt.Errorf("%w; rollback also failed: %v", cause, rollbackErr)\n}\n\nfunc rollbackRestoreActivation(activeRoot, stageRoot, preRestore string, activeOwner, stageOwner *workspaceOwnerLease, stageMoved bool) error {\n\tproblems := []string{}\n\tif stageMoved {\n\t\tif err := os.Rename(activeRoot, stageRoot); err != nil {\n\t\t\tproblems = append(problems, "move failed staged vault back: "+err.Error())\n\t\t} else if stageOwner != nil {\n\t\t\tif err := stageOwner.retarget(stageRoot); err != nil {\n\t\t\t\tproblems = append(problems, "retarget staged owner during rollback: "+err.Error())\n\t\t\t}\n\t\t}\n\t}\n\tif _, err := os.Stat(preRestore); err == nil {\n\t\tif err := os.Rename(preRestore, activeRoot); err != nil {\n\t\t\tproblems = append(problems, "restore original active vault: "+err.Error())\n\t\t} else if activeOwner != nil {\n\t\t\tif err := activeOwner.retarget(activeRoot); err != nil {\n\t\t\t\tproblems = append(problems, "retarget active owner during rollback: "+err.Error())\n\t\t\t}\n\t\t}\n\t} else if !os.IsNotExist(err) {\n\t\tproblems = append(problems, "inspect pre-restore checkpoint: "+err.Error())\n\t}\n\tif len(problems) > 0 {\n\t\treturn errors.New(strings.Join(problems, "; "))\n\t}\n\treturn nil\n}\n\ntype countingReader struct {\n''',
    "restore rollback helpers",
)
backup_path.write_text(backup, encoding="utf-8")

lifecycle_test = Path("internal/eco/vault_lifetime_test.go")
if lifecycle_test.exists():
    raise SystemExit("vault_lifetime_test.go already exists")
lifecycle_test.write_text(r'''package eco

import (
    "errors"
    "os"
    "path/filepath"
    "testing"
)

func TestOpenVaultExcludesSecondWriterUntilClose(t *testing.T) {
    root := filepath.Join(t.TempDir(), "vault")
    first, err := OpenVault(root)
    if err != nil {
        t.Fatal(err)
    }
    second, err := OpenVault(root)
    if second != nil {
        _ = second.Close()
    }
    if !errors.Is(err, ErrWorkspaceInUse) {
        _ = first.Close()
        t.Fatalf("second OpenVault error=%v", err)
    }
    if err := first.Close(); err != nil {
        t.Fatal(err)
    }
    reopened, err := OpenVault(root)
    if err != nil {
        t.Fatal(err)
    }
    if err := reopened.Close(); err != nil {
        t.Fatal(err)
    }
}

func TestVaultCloseMakesSaveFailClosed(t *testing.T) {
    root := filepath.Join(t.TempDir(), "vault")
    v, err := OpenVault(root)
    if err != nil {
        t.Fatal(err)
    }
    if err := v.Close(); err != nil {
        t.Fatal(err)
    }
    if err := v.Save(); !errors.Is(err, ErrVaultClosed) {
        t.Fatalf("Save after Close error=%v", err)
    }
    if len(v.key) != 0 {
        t.Fatal("closed Vault retained its encryption key")
    }
}

func TestRestoreTransfersActiveWorkspaceOwnership(t *testing.T) {
    root := t.TempDir()
    sourceFile := filepath.Join(root, "source.txt")
    if err := os.WriteFile(sourceFile, []byte("restored source bytes"), 0600); err != nil {
        t.Fatal(err)
    }
    sourceRoot := filepath.Join(root, "source-vault")
    source, err := OpenVault(sourceRoot)
    if err != nil {
        t.Fatal(err)
    }
    item, _, err := source.ImportFile(sourceFile, nil)
    if err != nil {
        _ = source.Close()
        t.Fatal(err)
    }
    backupPath := filepath.Join(root, "portable.ecobackup")
    if _, err := source.CreatePortableBackup(backupPath, "workspace-owner-test-passphrase", nil); err != nil {
        _ = source.Close()
        t.Fatal(err)
    }
    if err := source.Close(); err != nil {
        t.Fatal(err)
    }

    activeRoot := filepath.Join(root, "active-vault")
    active, err := OpenVault(activeRoot)
    if err != nil {
        t.Fatal(err)
    }
    receipt, err := active.RestorePortableBackup(backupPath, "workspace-owner-test-passphrase", nil)
    if err != nil {
        _ = active.Close()
        t.Fatal(err)
    }
    if receipt.PreRestoreVault == "" {
        _ = active.Close()
        t.Fatal("restore did not retain a pre-restore checkpoint path")
    }
    snapshot := active.Snapshot()
    if len(snapshot.Evidence) != 1 || snapshot.Evidence[0].ID != item.ID {
        _ = active.Close()
        t.Fatalf("active Vault did not receive restored workspace: %+v", snapshot.Evidence)
    }
    other, err := OpenVault(activeRoot)
    if other != nil {
        _ = other.Close()
    }
    if !errors.Is(err, ErrWorkspaceInUse) {
        _ = active.Close()
        t.Fatalf("restored active root lost exclusive ownership: %v", err)
    }
    if err := active.Save(); err != nil {
        _ = active.Close()
        t.Fatalf("restored active Vault could not persist through transferred owner: %v", err)
    }
    if err := active.Close(); err != nil {
        t.Fatal(err)
    }
    reopened, err := OpenVault(activeRoot)
    if err != nil {
        t.Fatal(err)
    }
    defer reopened.Close()
    reopenedSnapshot := reopened.Snapshot()
    if len(reopenedSnapshot.Evidence) != 1 || reopenedSnapshot.Evidence[0].ID != item.ID {
        t.Fatalf("restored workspace did not survive close/reopen: %+v", reopenedSnapshot.Evidence)
    }
}
''', encoding="utf-8")
