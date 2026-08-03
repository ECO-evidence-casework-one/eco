package eco

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	backupMagic             = "ECOBKP1\x00"
	backupVersion    uint32 = 1
	backupIterations uint32 = 600000
)

type BackupProgress struct {
	Stage, Name    string
	Current, Total int64
	Item, Items    int
}

type BackupReceipt struct {
	Format        string `json:"format"`
	BuildID       string `json:"build_id"`
	EvidenceItems int    `json:"evidence_items"`
	PlainBytes    int64  `json:"plain_bytes"`
	BackupBytes   int64  `json:"backup_bytes"`
	SHA256        string `json:"sha256"`
	Path          string `json:"path"`
}

func (v *Vault) CreatePortableBackup(path, passphrase string, progress func(BackupProgress)) (BackupReceipt, error) {
	v.opMu.RLock()
	defer v.opMu.RUnlock()
	if len([]rune(passphrase)) < 12 {
		return BackupReceipt{}, errors.New("use a backup passphrase of at least 12 characters")
	}
	ws := v.Snapshot()
	objects, err := requiredPreservedObjects(ws)
	if err != nil {
		return BackupReceipt{}, err
	}
	if filepath.Ext(path) == "" {
		path += ".ecobackup"
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return BackupReceipt{}, err
	}
	ok := false
	defer func() {
		f.Close()
		if !ok {
			os.Remove(tmp)
		}
	}()
	bw := bufio.NewWriterSize(f, 1024*1024)
	salt := make([]byte, 16)
	prefix := make([]byte, 4)
	if _, err = rand.Read(salt); err != nil {
		return BackupReceipt{}, err
	}
	if _, err = rand.Read(prefix); err != nil {
		return BackupReceipt{}, err
	}
	key := pbkdf2SHA256([]byte(passphrase), salt, int(backupIterations), 32)
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	if _, err = bw.Write([]byte(backupMagic)); err != nil {
		return BackupReceipt{}, err
	}
	for _, v32 := range []uint32{backupVersion, backupIterations} {
		if err = binary.Write(bw, binary.LittleEndian, v32); err != nil {
			return BackupReceipt{}, err
		}
	}
	if _, err = bw.Write(salt); err != nil {
		return BackupReceipt{}, err
	}
	if _, err = bw.Write(prefix); err != nil {
		return BackupReceipt{}, err
	}
	if err = writeShortString(bw, v.runtime.BuildID); err != nil {
		return BackupReceipt{}, err
	}
	var nonceCounter uint64
	var totalPlain int64
	workspaceBytes, err := json.Marshal(ws)
	if err != nil {
		return BackupReceipt{}, err
	}
	hws := sha256.Sum256(workspaceBytes)
	if progress != nil {
		progress(BackupProgress{Stage: "Encrypting workspace manifest", Name: "workspace", Item: 0, Items: len(objects) + 1})
	}
	if err = writeBackupRecordHeader(bw, 1, "workspace", "workspace.json", int64(len(workspaceBytes)), hws[:]); err != nil {
		return BackupReceipt{}, err
	}
	if err = writeEncryptedChunks(bw, gcm, prefix, &nonceCounter, 1, "workspace", "workspace.json", bytesReader(workspaceBytes), int64(len(workspaceBytes)), progress, 0, len(objects)+1); err != nil {
		return BackupReceipt{}, err
	}
	totalPlain += int64(len(workspaceBytes))
	for i, object := range objects {
		if progress != nil {
			progress(BackupProgress{Stage: "Reading encrypted original", Name: object.Name, Item: i + 1, Items: len(objects) + 1, Total: object.Size})
		}
		if _, verifyErr := v.verifyPreservedObject(object.ID, object.ObjectFile, object.SHA256, object.Size); verifyErr != nil {
			return BackupReceipt{}, fmt.Errorf("%s could not be freshly verified for backup: %w", object.Name, verifyErr)
		}
		if err = writeBackupRecordHeader(bw, 2, object.ID, object.ID, object.Size, mustDecodeHex(object.SHA256)); err != nil {
			return BackupReceipt{}, err
		}
		obj, err := os.Open(filepath.Join(v.Objects, object.ObjectFile))
		if err != nil {
			return BackupReceipt{}, fmt.Errorf("%s: %w", object.Name, err)
		}
		reader, err := newObjectPlainReader(v.key, object.ID, obj)
		if err != nil {
			obj.Close()
			return BackupReceipt{}, fmt.Errorf("%s: %w", object.Name, err)
		}
		err = writeEncryptedChunks(bw, gcm, prefix, &nonceCounter, 2, object.ID, object.ID, reader, object.Size, func(p BackupProgress) {
			if progress != nil {
				p.Name = object.Name
				progress(p)
			}
		}, i+1, len(objects)+1)
		obj.Close()
		if err != nil {
			return BackupReceipt{}, fmt.Errorf("%s: %w", object.Name, err)
		}
		totalPlain += object.Size
	}
	if err = bw.WriteByte(0xff); err != nil {
		return BackupReceipt{}, err
	}
	if err = bw.Flush(); err != nil {
		return BackupReceipt{}, err
	}
	if err = f.Sync(); err != nil {
		return BackupReceipt{}, err
	}
	if err = f.Close(); err != nil {
		return BackupReceipt{}, err
	}
	if err = os.Rename(tmp, path); err != nil {
		return BackupReceipt{}, err
	}
	ok = true
	info, err := os.Stat(path)
	if err != nil {
		return BackupReceipt{}, err
	}
	digest, err := hashFile(path)
	if err != nil {
		return BackupReceipt{}, err
	}
	v.AddChange("user", "portable-backup-created", "Created encrypted portable backup", map[string]any{"file": filepath.Base(path), "items": len(objects), "sha256": digest})
	return BackupReceipt{Format: "ECOBKP1", BuildID: v.runtime.BuildID, EvidenceItems: len(objects), PlainBytes: totalPlain, BackupBytes: info.Size(), SHA256: digest, Path: path}, nil
}

func writeBackupRecordHeader(w io.Writer, typ byte, id, name string, size int64, hash []byte) error {
	if _, err := w.Write([]byte{typ}); err != nil {
		return err
	}
	if err := writeShortString(w, id); err != nil {
		return err
	}
	if err := writeShortString(w, name); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, size); err != nil {
		return err
	}
	if len(hash) != 32 {
		return errors.New("invalid backup record hash")
	}
	_, err := w.Write(hash)
	return err
}
func writeShortString(w io.Writer, s string) error {
	b := []byte(s)
	if len(b) > 65535 {
		return errors.New("backup string too long")
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(len(b))); err != nil {
		return err
	}
	_, err := w.Write(b)
	return err
}

func writeEncryptedChunks(w io.Writer, gcm cipher.AEAD, prefix []byte, counter *uint64, typ byte, id, name string, r io.Reader, total int64, progress func(BackupProgress), item, items int) error {
	buf := make([]byte, chunkSize)
	var done int64
	var chunk uint64
	for {
		n, err := r.Read(buf)
		if n > 0 {
			nonce := make([]byte, gcm.NonceSize())
			copy(nonce, prefix)
			binary.BigEndian.PutUint64(nonce[len(nonce)-8:], *counter)
			(*counter)++
			aad := backupAAD(typ, id, name, chunk)
			sealed := gcm.Seal(nil, nonce, buf[:n], aad)
			if err := binary.Write(w, binary.LittleEndian, uint32(len(sealed))); err != nil {
				return err
			}
			if _, err := w.Write(sealed); err != nil {
				return err
			}
			done += int64(n)
			chunk++
			if progress != nil {
				progress(BackupProgress{Stage: "Encrypting portable backup", Name: name, Current: done, Total: total, Item: item, Items: items})
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return binary.Write(w, binary.LittleEndian, uint32(0))
}
func backupAAD(typ byte, id, name string, chunk uint64) []byte {
	return []byte(fmt.Sprintf("%s:%d:%s:%s:%d", backupMagic, typ, id, name, chunk))
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	hLen := 32
	blocks := (keyLen + hLen - 1) / hLen
	out := make([]byte, 0, blocks*hLen)
	for block := 1; block <= blocks; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], uint32(block))
		mac.Write(b[:])
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iterations; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func mustDecodeHex(s string) []byte { b, _ := hex.DecodeString(s); return b }

type byteSliceReader struct {
	b   []byte
	off int
}

func bytesReader(b []byte) *byteSliceReader { return &byteSliceReader{b: b} }
func (r *byteSliceReader) Read(p []byte) (int, error) {
	if r.off >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.off:])
	r.off += n
	return n, nil
}

type objectPlainReader struct {
	src      io.Reader
	gcm      cipher.AEAD
	prefix   []byte
	objectID string
	index    uint64
	buf      []byte
	off      int
	done     bool
}

func newObjectPlainReader(key []byte, id string, src io.Reader) (*objectPlainReader, error) {
	header := make([]byte, len(objectMagic))
	if _, err := io.ReadFull(src, header); err != nil {
		return nil, err
	}
	if string(header) != objectMagic {
		return nil, errors.New("bad object header")
	}
	prefix := make([]byte, 4)
	if _, err := io.ReadFull(src, prefix); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &objectPlainReader{src: src, gcm: gcm, prefix: prefix, objectID: id}, nil
}
func (r *objectPlainReader) Read(p []byte) (int, error) {
	if r.off < len(r.buf) {
		n := copy(p, r.buf[r.off:])
		r.off += n
		return n, nil
	}
	if r.done {
		return 0, io.EOF
	}
	var n uint32
	err := binary.Read(r.src, binary.LittleEndian, &n)
	if err == io.EOF {
		r.done = true
		return 0, io.EOF
	}
	if err != nil {
		return 0, err
	}
	if n == 0 || n > chunkSize+uint32(r.gcm.Overhead()) {
		return 0, errors.New("unsafe encrypted object chunk")
	}
	sealed := make([]byte, n)
	if _, err = io.ReadFull(r.src, sealed); err != nil {
		return 0, err
	}
	nonce := make([]byte, r.gcm.NonceSize())
	copy(nonce, r.prefix)
	binary.BigEndian.PutUint64(nonce[len(nonce)-8:], r.index)
	plain, err := r.gcm.Open(nil, nonce, sealed, []byte(fmt.Sprintf("%s:%d", r.objectID, r.index)))
	if err != nil {
		return 0, errors.New("object authentication failed")
	}
	r.index++
	r.buf = plain
	r.off = 0
	return r.Read(p)
}

type RestoreReceipt struct {
	Format          string `json:"format"`
	SourcePath      string `json:"source_path"`
	SourceBuildID   string `json:"source_build_id"`
	EvidenceItems   int    `json:"evidence_items"`
	RestoredBytes   int64  `json:"restored_bytes"`
	PreRestoreVault string `json:"pre_restore_vault"`
	SourceSHA256    string `json:"source_sha256"`
}

type RestorePhase string

const (
	restoreStageReadyHook    RestorePhase = "stage-ready"
	restoreStagedHook        RestorePhase = "staged"
	restoreOriginalMovedHook RestorePhase = "original-moved"
	restoreActivatedHook     RestorePhase = "activated"
	restoreRecoveredHook     RestorePhase = "recovered"
)

// Preserve the original internal test seam name.
const restoreStageReady = restoreStageReadyHook

// RestorePortableBackup authenticates and validates a portable backup in a
// separate staging vault. The active vault is replaced only after every
// record, hash, relationship and staged encrypted object has passed checks.
// If activation fails, the original vault is rolled back.
func (v *Vault) RestorePortableBackup(path, passphrase string, progress func(BackupProgress)) (RestoreReceipt, error) {
	return v.restorePortableBackup(path, passphrase, progress, nil)
}

func (v *Vault) restorePortableBackup(path, passphrase string, progress func(BackupProgress), hook func(RestorePhase, *Vault) error) (RestoreReceipt, error) {
	return v.restorePortableBackupWithOps(path, passphrase, progress, hook, operatingFilesystem)
}

func (v *Vault) restorePortableBackupWithOps(path, passphrase string, progress func(BackupProgress), hook func(RestorePhase, *Vault) error, ops filesystemOps) (receipt RestoreReceipt, resultErr error) {
	if len([]rune(passphrase)) < 12 {
		return RestoreReceipt{}, errors.New("backup passphrase must be at least 12 characters")
	}

	sourceHash, err := hashFile(path)
	if err != nil {
		return RestoreReceipt{}, err
	}
	if v.binding != nil || v.lifecycle != nil {
		return RestoreReceipt{}, errors.New("another workspace lifecycle transaction is already active")
	}
	if err = attachWorkspaceGuards(v); err != nil {
		return RestoreReceipt{}, fmt.Errorf("begin authenticated portable restore: %w", err)
	}
	if err = verifyBindingMatchesVault(v, v.binding); err != nil {
		v.releaseWorkspaceBinding()
		v.lifecycle.Close()
		v.lifecycle = nil
		return RestoreReceipt{}, fmt.Errorf("begin authenticated portable restore: %w", err)
	}
	defer func() {
		_ = v.releaseWorkspaceBinding()
		if v.lifecycle != nil {
			_ = v.lifecycle.Close()
			v.lifecycle = nil
		}
	}()
	f, err := os.Open(path)
	if err != nil {
		return RestoreReceipt{}, err
	}
	defer f.Close()
	br := bufio.NewReaderSize(f, 1024*1024)

	magic := make([]byte, len(backupMagic))
	if _, err = io.ReadFull(br, magic); err != nil {
		return RestoreReceipt{}, err
	}
	if string(magic) != backupMagic {
		return RestoreReceipt{}, errors.New("not a supported ECO encrypted backup")
	}

	var version uint32
	if err = binary.Read(br, binary.LittleEndian, &version); err != nil {
		return RestoreReceipt{}, err
	}
	if version != backupVersion {
		return RestoreReceipt{}, fmt.Errorf("unsupported backup version %d", version)
	}
	var iterations uint32
	if err = binary.Read(br, binary.LittleEndian, &iterations); err != nil {
		return RestoreReceipt{}, err
	}
	if iterations < 100000 || iterations > 1000000 {
		return RestoreReceipt{}, errors.New("backup key-derivation setting is outside safe bounds")
	}

	salt := make([]byte, 16)
	prefix := make([]byte, 4)
	if _, err = io.ReadFull(br, salt); err != nil {
		return RestoreReceipt{}, err
	}
	if _, err = io.ReadFull(br, prefix); err != nil {
		return RestoreReceipt{}, err
	}
	sourceBuild, err := readShortString(br)
	if err != nil {
		return RestoreReceipt{}, err
	}
	if len(sourceBuild) == 0 || len(sourceBuild) > 128 {
		return RestoreReceipt{}, errors.New("invalid source build identity")
	}

	password := []byte(passphrase)
	key := pbkdf2SHA256(password, salt, int(iterations), 32)
	zeroBytes(password)
	defer zeroBytes(key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return RestoreReceipt{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return RestoreReceipt{}, err
	}
	var nonceCounter uint64

	state, err := newRestoreState(v, sourceHash)
	if err != nil {
		return RestoreReceipt{}, err
	}
	restoreKey := append([]byte(nil), v.key...)
	defer zeroBytes(restoreKey)
	if err = writeRestoreState(&state, restoreKey); err != nil {
		return RestoreReceipt{}, err
	}
	stateStarted := true
	for _, role := range []string{"checkpoint", "stage", "failed"} {
		if err = writeRestoreRole(state, role, restoreKey); err != nil {
			return RestoreReceipt{}, fmt.Errorf("create authenticated portable restore folder identity: %w", err)
		}
	}
	var stage *Vault
	transactionComplete := false
	interrupted := false
	defer func() {
		if !stateStarted || transactionComplete || interrupted {
			return
		}
		if stage != nil {
			stage.Close()
		}
		_ = v.releaseWorkspaceBinding()
		_, cleanupErr := rollbackRestoreState(state, ops)
		rebindErr := v.rebindWorkspaceObjects()
		resultErr = errors.Join(resultErr, cleanupErr, rebindErr)
	}()

	stage, err = createVault(state.Stage, "Portable restore staging workspace", v.runtime)
	if err != nil {
		return RestoreReceipt{}, err
	}
	if err = writeRestoreStageIdentity(state, restoreKey); err != nil {
		return RestoreReceipt{}, err
	}

	var ws Workspace
	var requiredObjects []preservedObjectSpec
	workspaceSeen := false
	restored := make(map[string]bool)
	var totalBytes int64
	records := 0

	for {
		typ, readErr := br.ReadByte()
		if readErr != nil {
			return RestoreReceipt{}, readErr
		}
		if typ == 0xff {
			break
		}
		records++
		if records > 100001 {
			return RestoreReceipt{}, errors.New("backup contains too many records")
		}

		id, err := readShortString(br)
		if err != nil {
			return RestoreReceipt{}, err
		}
		name, err := readShortString(br)
		if err != nil {
			return RestoreReceipt{}, err
		}
		var size int64
		if err = binary.Read(br, binary.LittleEndian, &size); err != nil {
			return RestoreReceipt{}, err
		}
		expectedHash := make([]byte, sha256.Size)
		if _, err = io.ReadFull(br, expectedHash); err != nil {
			return RestoreReceipt{}, err
		}
		if size < 0 || size > 2*1024*1024*1024*1024 {
			return RestoreReceipt{}, errors.New("backup record size is unsafe")
		}
		if !safeRecordID(id) || len(name) > 512 {
			return RestoreReceipt{}, errors.New("backup record identity is invalid")
		}

		plainReader := &backupPlainReader{
			src:      br,
			gcm:      gcm,
			prefix:   prefix,
			counter:  &nonceCounter,
			typ:      typ,
			id:       id,
			name:     name,
			maxPlain: size,
		}
		hasher := sha256.New()
		counted := &countingReader{r: io.TeeReader(plainReader, hasher)}

		switch typ {
		case 1:
			if workspaceSeen || records != 1 || id != "workspace" || name != "workspace.json" {
				return RestoreReceipt{}, errors.New("workspace manifest is misplaced or duplicated")
			}
			if size > 64*1024*1024 {
				return RestoreReceipt{}, errors.New("workspace manifest is too large")
			}
			data, err := io.ReadAll(io.LimitReader(counted, size+1))
			if err != nil {
				return RestoreReceipt{}, fmt.Errorf("wrong passphrase, altered backup or unreadable workspace: %w", err)
			}
			if int64(len(data)) != size || counted.n != size {
				return RestoreReceipt{}, errors.New("workspace size mismatch")
			}
			if !hmac.Equal(hasher.Sum(nil), expectedHash) {
				return RestoreReceipt{}, errors.New("workspace hash mismatch")
			}
			if err = json.Unmarshal(data, &ws); err != nil {
				return RestoreReceipt{}, errors.New("workspace schema is invalid")
			}
			if err = validateRestoredWorkspace(&ws, v.runtime.Schema); err != nil {
				return RestoreReceipt{}, err
			}
			requiredObjects, err = requiredPreservedObjects(ws)
			if err != nil {
				return RestoreReceipt{}, fmt.Errorf("backup preservation manifest is invalid: %w", err)
			}
			workspaceSeen = true

		case 2:
			if !workspaceSeen {
				return RestoreReceipt{}, errors.New("evidence record appeared before workspace manifest")
			}
			if name != id || restored[id] {
				return RestoreReceipt{}, errors.New("duplicate or non-opaque evidence record")
			}
			object, referenced := preservedObjectForID(requiredObjects, id)
			if !referenced {
				return RestoreReceipt{}, errors.New("backup contains evidence not referenced by the workspace")
			}
			displayName := object.Name
			if progress != nil {
				progress(BackupProgress{
					Stage: "Restoring encrypted original",
					Name:  displayName,
					Item:  len(restored) + 1,
					Items: len(requiredObjects),
					Total: size,
				})
			}
			objectPath := filepath.Join(stage.Objects, object.ObjectFile)
			err = encryptStream(stage.key, id, counted, objectPath, size, func(p ImportProgress) {
				if progress != nil {
					progress(BackupProgress{
						Stage:   "Re-encrypting into staged vault",
						Name:    displayName,
						Current: p.Current,
						Total:   size,
						Item:    len(restored) + 1,
						Items:   len(requiredObjects),
					})
				}
			}, "", displayName)
			if err != nil {
				return RestoreReceipt{}, fmt.Errorf("wrong passphrase, altered backup or failed evidence restore: %w", err)
			}
			if counted.n != size {
				return RestoreReceipt{}, errors.New("restored evidence size mismatch")
			}
			if !hmac.Equal(hasher.Sum(nil), expectedHash) {
				return RestoreReceipt{}, errors.New("restored evidence SHA-256 mismatch")
			}
			restored[id] = true
			totalBytes += size

		default:
			return RestoreReceipt{}, errors.New("backup contains an unknown record type")
		}
	}

	if _, err := br.Peek(1); err != io.EOF {
		if err == nil {
			return RestoreReceipt{}, errors.New("backup contains trailing data")
		}
		return RestoreReceipt{}, err
	}
	if !workspaceSeen {
		return RestoreReceipt{}, errors.New("backup does not contain a workspace manifest")
	}
	if len(restored) != len(requiredObjects) {
		return RestoreReceipt{}, fmt.Errorf("backup has %d preserved objects but workspace requires %d", len(restored), len(requiredObjects))
	}
	for _, object := range requiredObjects {
		if !restored[object.ID] {
			return RestoreReceipt{}, fmt.Errorf("missing preserved object %s", object.ID)
		}
	}
	for i := range ws.Evidence {
		e := &ws.Evidence[i]
		e.SourcePath = ""
	}
	for i := range ws.Preservations {
		ws.Preservations[i].SourcePath = ""
	}

	stage.mu.Lock()
	stage.identityTransition = true
	stage.Workspace = ws
	stage.Workspace.BuildID = v.runtime.BuildID
	stage.Workspace.CreatedByCandidate = v.runtime.CandidateID
	stage.Identity = WorkspaceIdentity{Format: workspaceIdentityFormat, ID: ws.WorkspaceID, Name: ws.WorkspaceName, Kind: "development", Schema: ws.Schema, CreatedAt: ws.CreatedAt, CreatedByBuild: ws.CreatedByBuild, CreatedByCandidate: v.runtime.CandidateID}
	stage.addChangeUnlocked("system", "portable-backup-restored", "Restored and validated encrypted portable backup", map[string]any{
		"source_build":  sourceBuild,
		"source_sha256": sourceHash,
		"items":         len(restored),
	})
	err = stage.saveUnlocked()
	stage.mu.Unlock()
	if err != nil {
		return RestoreReceipt{}, err
	}
	if err = writeWorkspaceIdentityForVault(stage, stage.Identity); err != nil {
		return RestoreReceipt{}, err
	}
	stage.identityTransition = false
	if hook != nil {
		if err = hook(restoreStageReadyHook, stage); err != nil {
			if errors.Is(err, errRestoreInterrupted) {
				interrupted = true
			}
			return RestoreReceipt{}, err
		}
	}
	if err = verifyStagedVault(stage); err != nil {
		return RestoreReceipt{}, err
	}
	nextState := state
	nextState.RestoredWorkspaceID = stage.Identity.ID
	nextState.Phase = restoreStaged
	if err = writeRestoreState(&nextState, restoreKey); err != nil {
		return RestoreReceipt{}, err
	}
	state = nextState
	if hook != nil {
		if err = hook(restoreStagedHook, stage); err != nil {
			if errors.Is(err, errRestoreInterrupted) {
				interrupted = true
			}
			return RestoreReceipt{}, err
		}
	}

	// Activation is the only exclusive phase. Read/import/verify/backup file
	// operations finish first, and metadata writers are blocked by v.mu.
	v.opMu.Lock()
	opLocked := true
	defer func() {
		if opLocked {
			v.opMu.Unlock()
		}
	}()
	v.mu.Lock()
	metadataLocked := true
	defer func() {
		if metadataLocked {
			v.mu.Unlock()
		}
	}()

	if err = v.releaseWorkspaceBinding(); err != nil {
		return RestoreReceipt{}, fmt.Errorf("release the authenticated workspace handles before portable restore: %w", err)
	}
	if err = stage.releaseWorkspaceBinding(); err != nil {
		return RestoreReceipt{}, fmt.Errorf("release the authenticated staged-workspace handles before portable restore: %w", err)
	}
	if stage.lifecycle != nil {
		if err = stage.lifecycle.Close(); err != nil {
			return RestoreReceipt{}, err
		}
		stage.lifecycle = nil
	}
	if err = secureRestoreRename(state, state.Root, state.Checkpoint, "root-original", "checkpoint", ops); err != nil {
		return RestoreReceipt{}, fmt.Errorf("could not create pre-restore checkpoint: %w", err)
	}
	nextState = state
	nextState.Phase = restoreOriginalMoved
	if err = writeRestoreState(&nextState, restoreKey); err != nil {
		return RestoreReceipt{}, fmt.Errorf("could not record the original workspace checkpoint: %w", err)
	}
	state = nextState
	if hook != nil {
		if err = hook(restoreOriginalMovedHook, stage); err != nil {
			if errors.Is(err, errRestoreInterrupted) {
				interrupted = true
			}
			return RestoreReceipt{}, err
		}
	}
	if err = secureRestoreRename(state, state.Stage, state.Root, "stage", "root", ops); err != nil {
		return RestoreReceipt{}, fmt.Errorf("could not activate the authenticated restored workspace: %w", err)
	}
	nextState = state
	nextState.Phase = restoreActivated
	if err = writeRestoreState(&nextState, restoreKey); err != nil {
		return RestoreReceipt{}, fmt.Errorf("could not record restored-workspace activation: %w", err)
	}
	state = nextState
	if hook != nil {
		if err = hook(restoreActivatedHook, stage); err != nil {
			if errors.Is(err, errRestoreInterrupted) {
				interrupted = true
			}
			return RestoreReceipt{}, err
		}
	}
	if err = v.rebindWorkspaceObjects(); err != nil {
		return RestoreReceipt{}, fmt.Errorf("the restored workspace could not be rebound safely: %w", err)
	}
	if err = v.binding.Verify(); err != nil {
		return RestoreReceipt{}, fmt.Errorf("the restored workspace changed during activation: %w", err)
	}
	keyData, err := v.binding.ReadFile("vault.key")
	if err != nil {
		return RestoreReceipt{}, err
	}
	newKey, err := loadExistingMasterKeyData(keyData)
	if err != nil {
		return RestoreReceipt{}, err
	}
	workspaceData, err := v.binding.ReadFile("workspace.ecodb")
	if err != nil {
		zeroBytes(newKey)
		return RestoreReceipt{}, err
	}
	activeWorkspace, err := readEncryptedWorkspaceData(workspaceData, newKey)
	if err != nil {
		zeroBytes(newKey)
		return RestoreReceipt{}, err
	}
	identityData, err := v.binding.ReadFile(workspaceIdentityFile)
	if err != nil {
		zeroBytes(newKey)
		return RestoreReceipt{}, err
	}
	activeIdentity, err := readWorkspaceIdentityData(identityData)
	if err != nil || validateWorkspaceIdentity(activeIdentity, activeWorkspace) != nil || activeWorkspace.WorkspaceID != state.RestoredWorkspaceID {
		zeroBytes(newKey)
		return RestoreReceipt{}, errors.New("the activated workspace identity did not match the authenticated staged workspace")
	}
	zeroBytes(v.key)
	v.key = newKey
	v.Objects = filepath.Join(v.Root, "objects")
	v.Identity = activeIdentity
	v.Workspace = activeWorkspace
	zeroBytes(stage.key)
	stage.key = nil
	stage = nil
	v.mu.Unlock()
	metadataLocked = false
	v.opMu.Unlock()
	opLocked = false
	if err = v.recoverPreservations(); err != nil {
		return RestoreReceipt{}, fmt.Errorf("the activated workspace could not recover verified preservation state: %w", err)
	}

	nextState = state
	nextState.Phase = restoreRecovered
	if err = writeRestoreState(&nextState, restoreKey); err != nil {
		return RestoreReceipt{}, err
	}
	state = nextState
	if hook != nil {
		if err = hook(restoreRecoveredHook, v); err != nil {
			if errors.Is(err, errRestoreInterrupted) {
				interrupted = true
			}
			return RestoreReceipt{}, err
		}
	}
	if err = removeRestoreStageIdentity(state, state.Root, restoreKey, ops); err != nil {
		return RestoreReceipt{}, err
	}
	if err = cleanupRestoreControls(state, restoreKey, ops); err != nil {
		return RestoreReceipt{}, err
	}
	transactionComplete = true

	return RestoreReceipt{
		Format:          "ECOBKP1",
		SourcePath:      path,
		SourceBuildID:   sourceBuild,
		EvidenceItems:   len(restored),
		RestoredBytes:   totalBytes,
		PreRestoreVault: state.Checkpoint,
		SourceSHA256:    sourceHash,
	}, nil
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

type backupPlainReader struct {
	src      io.Reader
	gcm      cipher.AEAD
	prefix   []byte
	counter  *uint64
	typ      byte
	id       string
	name     string
	chunk    uint64
	buf      []byte
	off      int
	done     bool
	plain    int64
	maxPlain int64
}

func (r *backupPlainReader) Read(p []byte) (int, error) {
	if r.off < len(r.buf) {
		n := copy(p, r.buf[r.off:])
		r.off += n
		return n, nil
	}
	if r.done {
		return 0, io.EOF
	}

	var encryptedSize uint32
	if err := binary.Read(r.src, binary.LittleEndian, &encryptedSize); err != nil {
		return 0, err
	}
	if encryptedSize == 0 {
		r.done = true
		if r.plain != r.maxPlain {
			return 0, errors.New("record plaintext size mismatch")
		}
		return 0, io.EOF
	}
	if encryptedSize > chunkSize+uint32(r.gcm.Overhead()) {
		return 0, errors.New("encrypted backup chunk is unsafe")
	}

	sealed := make([]byte, encryptedSize)
	if _, err := io.ReadFull(r.src, sealed); err != nil {
		return 0, err
	}
	nonce := make([]byte, r.gcm.NonceSize())
	copy(nonce, r.prefix)
	binary.BigEndian.PutUint64(nonce[len(nonce)-8:], *r.counter)
	(*r.counter)++
	plain, err := r.gcm.Open(nil, nonce, sealed, backupAAD(r.typ, r.id, r.name, r.chunk))
	if err != nil {
		return 0, errors.New("wrong passphrase or altered backup")
	}
	r.chunk++
	r.plain += int64(len(plain))
	if r.plain > r.maxPlain {
		return 0, errors.New("record exceeds declared size")
	}
	r.buf = plain
	r.off = 0
	return r.Read(p)
}

func readShortString(r io.Reader) (string, error) {
	var n uint16
	if err := binary.Read(r, binary.LittleEndian, &n); err != nil {
		return "", err
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return "", err
	}
	return string(b), nil
}

func safeRecordID(s string) bool {
	if len(s) < 1 || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

type preservedObjectSpec struct {
	ID         string
	ObjectFile string
	Name       string
	Size       int64
	SHA256     string
}

func requiredPreservedObjects(ws Workspace) ([]preservedObjectSpec, error) {
	objects := make([]preservedObjectSpec, 0, len(ws.Evidence)+len(ws.Preservations))
	byEvidence := map[string]preservedObjectSpec{}
	for _, item := range ws.Evidence {
		if !safeRecordID(item.ID) || !safeManagedObjectName(item.ObjectFile) || item.Size < 0 || len(item.SHA256) != sha256.Size*2 {
			return nil, errors.New("workspace evidence has an invalid preserved-object receipt")
		}
		spec := preservedObjectSpec{ID: item.ID, ObjectFile: item.ObjectFile, Name: item.SafeName, Size: item.Size, SHA256: item.SHA256}
		if _, exists := byEvidence[item.ID]; exists {
			return nil, errors.New("workspace evidence has duplicate preserved-object identities")
		}
		byEvidence[item.ID] = spec
		objects = append(objects, spec)
	}
	for _, record := range ws.Preservations {
		switch record.State {
		case preservationCommitted:
			spec, exists := byEvidence[record.EvidenceID]
			if !exists || spec.ObjectFile != record.ObjectFile || spec.Size != record.ExpectedSize || spec.SHA256 != record.PreservedSHA256 {
				return nil, errors.New("a committed preservation receipt does not match its usable evidence record")
			}
		case preservationFailed:
			continue
		case preservationRecoverable:
			if !safeRecordID(record.EvidenceID) || !safeManagedObjectName(record.ObjectFile) || record.ExpectedSize < 0 || record.BytesPreserved != record.ExpectedSize || record.PreservedSHA256 == "" || record.PreservedSHA256 != record.IntakeSHA256 || record.VerifiedAt.IsZero() {
				return nil, errors.New("a pending preservation is not complete enough for a truthful backup")
			}
			if _, exists := byEvidence[record.EvidenceID]; exists {
				return nil, errors.New("a pending preservation duplicates an evidence object")
			}
			spec := preservedObjectSpec{ID: record.EvidenceID, ObjectFile: record.ObjectFile, Name: record.SafeName, Size: record.ExpectedSize, SHA256: record.PreservedSHA256}
			byEvidence[record.EvidenceID] = spec
			objects = append(objects, spec)
		default:
			return nil, errors.New("backup was blocked because a preservation operation is incomplete")
		}
	}
	return objects, nil
}

func preservedObjectForID(objects []preservedObjectSpec, id string) (preservedObjectSpec, bool) {
	for _, object := range objects {
		if object.ID == id {
			return object, true
		}
	}
	return preservedObjectSpec{}, false
}

func validateRestoredWorkspace(ws *Workspace, expectedSchema int) error {
	if ws.Schema != expectedSchema {
		return fmt.Errorf("unsupported restored workspace schema %d", ws.Schema)
	}
	if !safeRecordID(ws.WorkspaceID) || !validWorkspaceName(ws.WorkspaceName) || !validIdentityLabel(ws.CreatedByBuild, 128) || !validIdentityLabel(ws.CreatedByCandidate, 256) || ws.CreatedAt.IsZero() {
		return errors.New("restored workspace identity is invalid")
	}
	if len(ws.Evidence) > 100000 || len(ws.Matters) > 100000 || len(ws.Changes) > 500000 || len(ws.Questions) > 100000 {
		return errors.New("restored workspace record counts are unsafe")
	}

	evidenceIDs := make(map[string]bool, len(ws.Evidence))
	var totalExtracted int64
	for i := range ws.Evidence {
		e := &ws.Evidence[i]
		if !safeRecordID(e.ID) || evidenceIDs[e.ID] {
			return errors.New("restored evidence identifiers are invalid or duplicated")
		}
		evidenceIDs[e.ID] = true
		if len([]rune(e.SafeName)) == 0 || len([]rune(e.SafeName)) > 512 || e.Size < 0 || e.Size > 2*1024*1024*1024*1024 {
			return errors.New("restored evidence metadata is unsafe")
		}
		if len(e.SHA256) != 64 {
			return errors.New("restored evidence SHA-256 is invalid")
		}
		if _, err := hex.DecodeString(e.SHA256); err != nil {
			return errors.New("restored evidence SHA-256 is invalid")
		}
		if int64(len(e.ExtractedText)) > maxExtractBytes {
			return errors.New("restored extracted text exceeds safety bound")
		}
		totalExtracted += int64(len(e.ExtractedText))
		if totalExtracted > 2*1024*1024*1024 {
			return errors.New("restored total extracted text exceeds safety bound")
		}
		if len(e.Segments) > 100000 {
			return errors.New("restored evidence has too many source segments")
		}
		for _, segment := range e.Segments {
			if len(segment.Text) > 5000 || len(segment.ID) > 128 || len(segment.PageHint) > 256 {
				return errors.New("restored source segment exceeds safety bounds")
			}
		}
		if len(e.Warnings) > 1000 || len(e.MatterIDs) > 100000 {
			return errors.New("restored evidence relationships exceed safety bounds")
		}
	}

	matterIDs := make(map[string]bool, len(ws.Matters))
	for _, matter := range ws.Matters {
		if !safeRecordID(matter.ID) || matterIDs[matter.ID] || len([]rune(matter.Title)) == 0 || len([]rune(matter.Title)) > 1000 {
			return errors.New("restored matter metadata is invalid")
		}
		matterIDs[matter.ID] = true
		if len(matter.EvidenceIDs) > 100000 {
			return errors.New("restored matter has too many evidence relationships")
		}
		for _, evidenceID := range matter.EvidenceIDs {
			if !evidenceIDs[evidenceID] {
				return errors.New("restored matter references missing evidence")
			}
		}
	}
	for _, e := range ws.Evidence {
		for _, matterID := range e.MatterIDs {
			if !matterIDs[matterID] {
				return errors.New("restored evidence references a missing matter")
			}
		}
	}
	for _, q := range ws.Questions {
		if !safeRecordID(q.ID) || len(q.Question) > 20000 || len(q.Answer) > 200000 || len(q.Citations) > 10000 {
			return errors.New("restored question record is unsafe")
		}
		for _, citation := range q.Citations {
			if !evidenceIDs[citation.EvidenceID] || len(citation.Quote) > 10000 {
				return errors.New("restored citation is invalid")
			}
		}
	}
	return nil
}

func verifyStagedVault(v *Vault) error {
	ws := v.Snapshot()
	objects, err := requiredPreservedObjects(ws)
	if err != nil {
		return err
	}
	for _, object := range objects {
		receipt, verifyErr := v.verifyPreservedObject(object.ID, object.ObjectFile, object.SHA256, object.Size)
		if verifyErr != nil {
			return fmt.Errorf("staged object verification failed for %s: %w", object.Name, verifyErr)
		}
		if receipt.ObjectFile != object.ObjectFile || receipt.SHA256 != object.SHA256 || receipt.Size != object.Size {
			return fmt.Errorf("staged object receipt mismatch for %s", object.Name)
		}
	}
	return nil
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
