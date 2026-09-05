package eco

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Every fixture is synthetic and uses the real encrypted workspace format.
func workspaceFormatFixture(t *testing.T) (string, []byte, map[string]json.RawMessage) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "workspace")
	v, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.CreateMatter("Synthetic retained matter"); err != nil {
		_ = v.Close()
		t.Fatal(err)
	}
	v.Workspace.Settings.LowSensory = true
	if err := v.Save(); err != nil {
		_ = v.Close()
		t.Fatal(err)
	}
	key := append([]byte(nil), v.key...)
	data, err := json.Marshal(v.Snapshot())
	closeErr := v.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	t.Cleanup(func() { zeroBytes(key) })
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	return root, key, document
}

func workspaceFormatWrite(t *testing.T, root string, key, plain []byte) {
	t.Helper()
	data, err := encryptBlob(key, metaMagic, plain)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "workspace.ecodb"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func workspaceFormatJSON(t *testing.T, document map[string]json.RawMessage) []byte {
	t.Helper()
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Access times are deliberately excluded: reads may update them. We compare
// directory entries, content hashes, modes, sizes and modification timestamps.
func workspaceFormatTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		value := fmt.Sprintf("%s|%d|%d", info.Mode(), info.Size(), info.ModTime().UnixNano())
		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			value += fmt.Sprintf("|%x", sha256.Sum256(data))
		}
		out[rel] = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestWorkspaceFormatRejectsWithoutMutation(t *testing.T) {
	cases := []string{
		"future_schema", "unsupported_older_schema", "missing_schema",
		"unknown_top", "unknown_nested", "unknown_matter",
		"trailing_json", "malformed_json", "authentication_failure",
		"future_missing_objects", "missing_key", "missing_metadata",
		"objects_only", "unknown_without_owner_file",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			root, key, document := workspaceFormatFixture(t)
			switch name {
			case "future_schema", "future_missing_objects":
				document["schema"] = json.RawMessage(fmt.Sprint(Schema + 1))
			case "unsupported_older_schema":
				document["schema"] = json.RawMessage("0")
			case "missing_schema":
				delete(document, "schema")
			case "unknown_top", "unknown_without_owner_file":
				document["future_extension"] = json.RawMessage(`{"retain":"synthetic"}`)
			case "unknown_nested":
				document["settings"] = json.RawMessage(`{"low_sensory":true,"future_option":true}`)
			case "unknown_matter":
				var matters []map[string]json.RawMessage
				if err := json.Unmarshal(document["matters"], &matters); err != nil {
					t.Fatal(err)
				}
				matters[0]["future_note"] = json.RawMessage(`"retain synthetic note"`)
				data, err := json.Marshal(matters)
				if err != nil {
					t.Fatal(err)
				}
				document["matters"] = data
			}
			plain := workspaceFormatJSON(t, document)
			if name == "trailing_json" {
				plain = append(plain, []byte(` {"schema":1}`)...)
			}
			if name == "malformed_json" {
				plain = []byte(`{"schema":`)
			}
			workspaceFormatWrite(t, root, key, plain)
			switch name {
			case "authentication_failure":
				if err := os.WriteFile(filepath.Join(root, "workspace.ecodb"), []byte("not authenticated metadata"), 0600); err != nil {
					t.Fatal(err)
				}
			case "future_missing_objects":
				if err := os.Remove(filepath.Join(root, "objects")); err != nil {
					t.Fatal(err)
				}
			case "missing_key":
				if err := os.Remove(filepath.Join(root, "vault.key")); err != nil {
					t.Fatal(err)
				}
			case "missing_metadata":
				if err := os.Remove(filepath.Join(root, "workspace.ecodb")); err != nil {
					t.Fatal(err)
				}
			case "objects_only":
				for _, file := range []string{"vault.key", "workspace.ecodb"} {
					if err := os.Remove(filepath.Join(root, file)); err != nil {
						t.Fatal(err)
					}
				}
			case "unknown_without_owner_file":
				if err := os.Remove(filepath.Join(root, workspaceOwnerFilename)); err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
			}
			before := workspaceFormatTree(t, root)
			opened, err := OpenVault(root)
			if opened != nil {
				if closeErr := opened.Close(); closeErr != nil {
					t.Fatal(closeErr)
				}
			}
			if err == nil {
				t.Error("unsupported or incomplete workspace accepted for writing")
			}
			if after := workspaceFormatTree(t, root); !reflect.DeepEqual(before, after) {
				t.Error("refused workspace entries/content/modes/mtime changed")
			}
		})
	}
}

func TestWorkspaceFormatLegacyDefaultsPreserved(t *testing.T) {
	root, key, document := workspaceFormatFixture(t)
	delete(document, "revision")
	delete(document, "last_owner_txn")
	delete(document, "preservations")
	document["build_id"] = json.RawMessage(`"ECO-SYNTHETIC-OLDER-CANDIDATE"`)
	workspaceFormatWrite(t, root, key, workspaceFormatJSON(t, document))
	before := workspaceFormatTree(t, root)
	v, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if !reflect.DeepEqual(before, workspaceFormatTree(t, root)) {
		t.Fatal("opening supported legacy state rewrote it")
	}
	ws := v.Snapshot()
	if ws.Revision != 0 || len(ws.Matters) != 1 || ws.Matters[0].Title != "Synthetic retained matter" || !ws.Settings.LowSensory {
		t.Fatal("supported legacy data or defaults were lost")
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	got := reopened.Snapshot()
	if got.Revision != 1 || got.LastOwnerTxn == "" || len(got.Matters) != 1 || got.Matters[0].Title != ws.Matters[0].Title || !got.Settings.LowSensory || !got.CreatedAt.Equal(ws.CreatedAt) {
		t.Fatal("legacy save/reopen failed to preserve records or advance ownership revision")
	}
}

// This uses the existing portable-backup writer primitives so the malformed
// manifest still has valid backup authentication and cannot fail merely due
// to a wrong passphrase, bad framing or an invalid digest.
func workspaceFormatBackup(t *testing.T, path, passphrase string, plain []byte) {
	t.Helper()
	const iterations uint32 = 100000
	salt := bytes.Repeat([]byte{0x17}, 16)
	prefix := []byte{1, 2, 3, 4}
	key := pbkdf2SHA256([]byte(passphrase), salt, int(iterations), 32)
	defer zeroBytes(key)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	output.WriteString(backupMagic)
	for _, n := range []uint32{backupVersion, iterations} {
		if err := binary.Write(&output, binary.LittleEndian, n); err != nil {
			t.Fatal(err)
		}
	}
	output.Write(salt)
	output.Write(prefix)
	if err := writeShortString(&output, BuildID); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(plain)
	if err := writeBackupRecordHeader(&output, 1, "workspace", "workspace.json", int64(len(plain)), digest[:]); err != nil {
		t.Fatal(err)
	}
	var counter uint64
	if err := writeEncryptedChunks(&output, gcm, prefix, &counter, 1, "workspace", "workspace.json", bytes.NewReader(plain), int64(len(plain)), nil, 0, 1); err != nil {
		t.Fatal(err)
	}
	output.WriteByte(0xff)
	if err := os.WriteFile(path, output.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceFormatRestoreRejectsWithoutActivation(t *testing.T) {
	for _, name := range []string{"unknown_top", "unknown_nested", "future_schema"} {
		t.Run(name, func(t *testing.T) {
			root, _, document := workspaceFormatFixture(t)
			switch name {
			case "unknown_top":
				document["future_extension"] = json.RawMessage(`true`)
			case "unknown_nested":
				document["settings"] = json.RawMessage(`{"future_option":true}`)
			case "future_schema":
				document["schema"] = json.RawMessage(fmt.Sprint(Schema + 1))
			}
			path := filepath.Join(t.TempDir(), "synthetic.ecobackup")
			passphrase := "synthetic compatibility fixture"
			workspaceFormatBackup(t, path, passphrase, workspaceFormatJSON(t, document))
			v, err := OpenVault(root)
			if err != nil {
				t.Fatal(err)
			}
			defer v.Close()
			before := workspaceFormatTree(t, root)
			view := v.Snapshot()
			if _, err := v.RestorePortableBackup(path, passphrase, nil); err == nil {
				t.Error("incompatible backup manifest was activated")
			}
			if !reflect.DeepEqual(view, v.Snapshot()) || !reflect.DeepEqual(before, workspaceFormatTree(t, root)) {
				t.Error("active workspace changed after incompatible restore")
			}
			for _, pattern := range []string{root + ".restore-*", root + ".pre-restore-*"} {
				matches, err := filepath.Glob(pattern)
				if err != nil {
					t.Fatal(err)
				}
				if len(matches) != 0 {
					t.Errorf("restore left unexpected workspace directories: %v", matches)
				}
			}
		})
	}
}

func TestWorkspaceFormatCompatibleBackupFixture(t *testing.T) {
	root, _, document := workspaceFormatFixture(t)
	path := filepath.Join(t.TempDir(), "synthetic.ecobackup")
	passphrase := "synthetic compatibility fixture"
	workspaceFormatBackup(t, path, passphrase, workspaceFormatJSON(t, document))
	v, err := OpenVault(root)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if _, err := v.RestorePortableBackup(path, passphrase, nil); err != nil {
		t.Fatal(err)
	}
	got := v.Snapshot()
	if len(got.Matters) != 1 || got.Matters[0].Title != "Synthetic retained matter" || !got.Settings.LowSensory {
		t.Fatal("compatible restore lost synthetic records")
	}
}
