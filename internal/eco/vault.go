package eco

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
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
	"sync"
	"time"
)

const (
	objectMagic = "ECOOBJ1\x00"
	metaMagic   = "ECOMETA1"
	chunkSize   = 1024 * 1024
)

type Vault struct {
	Root      string
	Objects   string
	key       []byte
	opMu      sync.RWMutex
	mu        sync.Mutex
	Workspace Workspace
}

func OpenVault(root string) (*Vault, error) {
	if root == "" {
		return nil, errors.New("empty vault root")
	}
	objects := filepath.Join(root, "objects")
	if err := os.MkdirAll(objects, 0700); err != nil {
		return nil, err
	}
	key, err := loadOrCreateMasterKey(filepath.Join(root, "vault.key"))
	if err != nil {
		return nil, fmt.Errorf("vault key: %w", err)
	}
	v := &Vault{Root: root, Objects: objects, key: key}
	if err := v.loadWorkspace(); err != nil {
		return nil, err
	}
	return v, nil
}

func loadOrCreateMasterKey(path string) ([]byte, error) {
	if data, err := os.ReadFile(path); err == nil {
		return unprotectLocalKey(data)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	protected, err := protectLocalKey(key)
	if err != nil {
		return nil, err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, protected, 0600); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp, path); err != nil {
		return nil, err
	}
	return key, nil
}

func newWorkspace() Workspace {
	now := time.Now().UTC()
	return Workspace{Schema: Schema, BuildID: BuildID, CreatedAt: now, UpdatedAt: now, Evidence: []EvidenceItem{}, Matters: []Matter{}, Changes: []ChangeRecord{}, Questions: []QuestionRecord{}, SelectedPage: "home"}
}

func (v *Vault) loadWorkspace() error {
	path := filepath.Join(v.Root, "workspace.ecodb")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		v.Workspace = newWorkspace()
		return v.Save()
	}
	if err != nil {
		return err
	}
	plain, err := decryptBlob(v.key, metaMagic, data)
	if err != nil {
		return fmt.Errorf("workspace authentication failed: %w", err)
	}
	var ws Workspace
	if err := json.Unmarshal(plain, &ws); err != nil {
		return fmt.Errorf("workspace format invalid: %w", err)
	}
	if ws.Schema != Schema {
		return fmt.Errorf("unsupported workspace schema %d", ws.Schema)
	}
	v.Workspace = ws
	return nil
}

func (v *Vault) Save() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.saveUnlocked()
}

func (v *Vault) saveUnlocked() error {
	v.Workspace.UpdatedAt = time.Now().UTC()
	v.Workspace.BuildID = BuildID
	plain, err := json.MarshalIndent(v.Workspace, "", "  ")
	if err != nil {
		return err
	}
	enc, err := encryptBlob(v.key, metaMagic, plain)
	if err != nil {
		return err
	}
	path := filepath.Join(v.Root, "workspace.ecodb")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, enc, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	return nil
}

func (v *Vault) Snapshot() Workspace {
	v.mu.Lock()
	defer v.mu.Unlock()
	// Do not serialise the complete encrypted workspace merely to paint or
	// refresh the native UI. Large extracted documents made the previous JSON
	// round-trip block the Windows message thread long enough to appear hung.
	// Workspace records are append/replace oriented, so a structural slice copy
	// is sufficient for a read-only UI snapshot while immutable strings can be
	// shared safely.
	out := v.Workspace
	out.Evidence = append([]EvidenceItem(nil), v.Workspace.Evidence...)
	for i := range out.Evidence {
		out.Evidence[i].Warnings = append([]string(nil), v.Workspace.Evidence[i].Warnings...)
		out.Evidence[i].Segments = append([]SourceSegment(nil), v.Workspace.Evidence[i].Segments...)
		out.Evidence[i].MatterIDs = append([]string(nil), v.Workspace.Evidence[i].MatterIDs...)
		if v.Workspace.Evidence[i].Image != nil {
			img := *v.Workspace.Evidence[i].Image
			img.Warnings = append([]string(nil), v.Workspace.Evidence[i].Image.Warnings...)
			out.Evidence[i].Image = &img
		}
	}
	out.Matters = append([]Matter(nil), v.Workspace.Matters...)
	for i := range out.Matters {
		out.Matters[i].EvidenceIDs = append([]string(nil), v.Workspace.Matters[i].EvidenceIDs...)
	}
	out.Changes = append([]ChangeRecord(nil), v.Workspace.Changes...)
	out.Questions = append([]QuestionRecord(nil), v.Workspace.Questions...)
	for i := range out.Questions {
		out.Questions[i].Citations = append([]Citation(nil), v.Workspace.Questions[i].Citations...)
		out.Questions[i].ScopeIDs = append([]string(nil), v.Workspace.Questions[i].ScopeIDs...)
	}
	return out
}

func encryptBlob(key []byte, aad string, plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, plain, []byte(aad))
	out := append([]byte(aad), 0)
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

func decryptBlob(key []byte, aad string, data []byte) ([]byte, error) {
	prefix := append([]byte(aad), 0)
	if len(data) < len(prefix) || !bytes.Equal(data[:len(prefix)], prefix) {
		return nil, errors.New("bad encrypted blob header")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	data = data[len(prefix):]
	if len(data) < gcm.NonceSize() {
		return nil, errors.New("truncated encrypted blob")
	}
	return gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], []byte(aad))
}

func (v *Vault) ImportFile(path string, progress func(ImportProgress)) (EvidenceItem, bool, error) {
	v.opMu.RLock()
	defer v.opMu.RUnlock()
	info, err := os.Stat(path)
	if err != nil {
		return EvidenceItem{}, false, err
	}
	if !info.Mode().IsRegular() {
		return EvidenceItem{}, false, errors.New("only regular files can be imported")
	}
	if progress != nil {
		progress(ImportProgress{Path: path, Name: info.Name(), Stage: "Checking file type", Total: info.Size()})
	}
	det, err := DetectFile(path)
	if err != nil {
		return EvidenceItem{}, false, err
	}

	f, err := os.Open(path)
	if err != nil {
		return EvidenceItem{}, false, err
	}
	defer f.Close()
	h := sha256.New()
	buf := make([]byte, chunkSize)
	var done int64
	for {
		n, rerr := f.Read(buf)
		if n > 0 {
			_, _ = h.Write(buf[:n])
			done += int64(n)
			if progress != nil {
				progress(ImportProgress{Path: path, Name: info.Name(), Stage: "Hashing original", Current: done, Total: info.Size()})
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return EvidenceItem{}, false, rerr
		}
	}
	hash := hex.EncodeToString(h.Sum(nil))
	v.mu.Lock()
	for _, e := range v.Workspace.Evidence {
		if e.SHA256 == hash {
			v.mu.Unlock()
			return e, true, nil
		}
	}
	v.mu.Unlock()

	id := NewID("EVD")
	objectName := id + ".ecoobj"
	objectPath := filepath.Join(v.Objects, objectName)
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return EvidenceItem{}, false, err
	}
	if progress != nil {
		progress(ImportProgress{Path: path, Name: info.Name(), Stage: "Encrypting original", Total: info.Size()})
	}
	if err := encryptStream(v.key, id, f, objectPath, info.Size(), progress, path, info.Name()); err != nil {
		return EvidenceItem{}, false, err
	}

	item := EvidenceItem{ID: id, OriginalName: info.Name(), SafeName: SafeDisplayName(info.Name()), SourcePath: path, Size: info.Size(), SHA256: hash, DetectedType: det.Type, ExtensionType: det.ExtensionType, TypeMismatch: det.Mismatch, Status: "Preserved", ImportedAt: time.Now().UTC(), ObjectFile: objectName}
	if det.Warning != "" {
		item.Warnings = append(item.Warnings, det.Warning)
	}
	if det.Dangerous {
		item.Status = "Quarantined"
		item.Readable = false
	} else {
		if progress != nil {
			progress(ImportProgress{Path: path, Name: info.Name(), Stage: "Reading safely", Total: info.Size()})
		}
		text, segs, readWarnings := ExtractReadable(path, det.Type)
		item.ExtractedText = text
		item.Segments = segs
		item.Warnings = append(item.Warnings, readWarnings...)
		item.Readable = len(stringsTrim(text)) > 0
		if item.Readable {
			item.Status = "Ready"
		} else if isImageType(det.Type) {
			item.Status = "Image ready"
		} else {
			item.Status = "Preserved — contents not read"
		}
		if isImageType(det.Type) {
			data, rerr := readFileBounded(path, 120*1024*1024)
			if rerr == nil {
				if img, _, derr := DecodeSupportedImage(data); derr == nil {
					a := AssessImage(img)
					item.Image = &a
					item.Warnings = append(item.Warnings, a.Warnings...)
					v.mu.Lock()
					for _, existing := range v.Workspace.Evidence {
						if existing.Image != nil && existing.Image.PerceptualHash != "" && HashDistance(a.PerceptualHash, existing.Image.PerceptualHash) <= 6 {
							item.NearDuplicateOf = existing.ID
							item.Warnings = append(item.Warnings, "This image appears visually similar to "+existing.SafeName+". Review both before excluding either one.")
							break
						}
					}
					v.mu.Unlock()
				} else {
					item.Warnings = append(item.Warnings, "Image preserved, but this native preview could not decode it for visual assessment.")
				}
			}
		}
	}
	v.mu.Lock()
	v.Workspace.Evidence = append([]EvidenceItem{item}, v.Workspace.Evidence...)
	v.Workspace.SelectedID = item.ID
	v.addChangeUnlocked("system", "evidence-imported", "Imported and encrypted "+item.SafeName, map[string]any{"id": item.ID, "sha256": item.SHA256, "type": item.DetectedType, "size": item.Size})
	err = v.saveUnlocked()
	v.mu.Unlock()
	if err != nil {
		return EvidenceItem{}, false, err
	}
	return item, false, nil
}

func encryptStream(key []byte, objectID string, src io.Reader, dstPath string, size int64, progress func(ImportProgress), sourcePath, name string) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	prefix := make([]byte, 4)
	if _, err := rand.Read(prefix); err != nil {
		return err
	}
	tmp := dstPath + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		out.Close()
		if !ok {
			os.Remove(tmp)
		}
	}()
	if _, err := out.Write([]byte(objectMagic)); err != nil {
		return err
	}
	if _, err := out.Write(prefix); err != nil {
		return err
	}
	buf := make([]byte, chunkSize)
	var index uint64
	var done int64
	for {
		n, rerr := src.Read(buf)
		if n > 0 {
			nonce := make([]byte, gcm.NonceSize())
			copy(nonce, prefix)
			binary.BigEndian.PutUint64(nonce[len(nonce)-8:], index)
			aad := []byte(fmt.Sprintf("%s:%d", objectID, index))
			sealed := gcm.Seal(nil, nonce, buf[:n], aad)
			if err := binary.Write(out, binary.LittleEndian, uint32(len(sealed))); err != nil {
				return err
			}
			if _, err := out.Write(sealed); err != nil {
				return err
			}
			done += int64(n)
			index++
			if progress != nil {
				progress(ImportProgress{Path: sourcePath, Name: name, Stage: "Encrypting original", Current: done, Total: size})
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	if err := out.Sync(); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, dstPath); err != nil {
		return err
	}
	ok = true
	return nil
}

func (v *Vault) ReadEvidence(id string, maxBytes int64) ([]byte, error) {
	v.opMu.RLock()
	defer v.opMu.RUnlock()
	v.mu.Lock()
	var itemCopy EvidenceItem
	found := false
	for i := range v.Workspace.Evidence {
		if v.Workspace.Evidence[i].ID == id {
			itemCopy = v.Workspace.Evidence[i]
			found = true
			break
		}
	}
	v.mu.Unlock()
	var item *EvidenceItem
	if found {
		item = &itemCopy
	}
	if item == nil {
		return nil, os.ErrNotExist
	}
	if maxBytes > 0 && item.Size > maxBytes {
		return nil, fmt.Errorf("evidence is %s; preview limit is %s", HumanBytes(item.Size), HumanBytes(maxBytes))
	}
	f, err := os.Open(filepath.Join(v.Objects, item.ObjectFile))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return decryptObject(v.key, id, f, maxBytes)
}

func decryptObject(key []byte, objectID string, src io.Reader, maxBytes int64) ([]byte, error) {
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
	var out bytes.Buffer
	var index uint64
	for {
		var n uint32
		err := binary.Read(src, binary.LittleEndian, &n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if n == 0 || n > chunkSize+uint32(gcm.Overhead()) {
			return nil, errors.New("unsafe encrypted chunk length")
		}
		sealed := make([]byte, n)
		if _, err := io.ReadFull(src, sealed); err != nil {
			return nil, err
		}
		nonce := make([]byte, gcm.NonceSize())
		copy(nonce, prefix)
		binary.BigEndian.PutUint64(nonce[len(nonce)-8:], index)
		plain, err := gcm.Open(nil, nonce, sealed, []byte(fmt.Sprintf("%s:%d", objectID, index)))
		if err != nil {
			return nil, errors.New("object authentication failed")
		}
		if maxBytes > 0 && int64(out.Len()+len(plain)) > maxBytes {
			return nil, errors.New("decrypted object exceeds preview limit")
		}
		out.Write(plain)
		index++
	}
	return out.Bytes(), nil
}

func (v *Vault) AddChange(actor, typ, summary string, details map[string]any) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.addChangeUnlocked(actor, typ, summary, details)
	_ = v.saveUnlocked()
}

func (v *Vault) addChangeUnlocked(actor, typ, summary string, details map[string]any) {
	prev := ""
	if len(v.Workspace.Changes) > 0 {
		prev = v.Workspace.Changes[0].Hash
	}
	rec := ChangeRecord{ID: NewID("CHG"), At: time.Now().UTC(), Actor: actor, Type: typ, Summary: summary, Details: details, PrevHash: prev}
	b, _ := json.Marshal(struct {
		ID                         string
		At                         time.Time
		Actor, Type, Summary, Prev string
		Details                    map[string]any
	}{rec.ID, rec.At, rec.Actor, rec.Type, rec.Summary, rec.PrevHash, rec.Details})
	h := sha256.Sum256(b)
	rec.Hash = hex.EncodeToString(h[:])
	v.Workspace.Changes = append([]ChangeRecord{rec}, v.Workspace.Changes...)
}

func (v *Vault) SetRotation(id string, degrees int) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	for i := range v.Workspace.Evidence {
		if v.Workspace.Evidence[i].ID == id {
			v.Workspace.Evidence[i].Rotation = ((degrees % 360) + 360) % 360
			v.addChangeUnlocked("user", "image-rotation", "Changed non-destructive image rotation", map[string]any{"id": id, "rotation": v.Workspace.Evidence[i].Rotation})
			return v.saveUnlocked()
		}
	}
	return os.ErrNotExist
}

func (v *Vault) SetSelectedPage(page string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.Workspace.SelectedPage = page
	return v.saveUnlocked()
}
func (v *Vault) SetSelectedEvidence(id string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.Workspace.SelectedID = id
	return v.saveUnlocked()
}
func (v *Vault) ToggleLowSensory() (bool, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.Workspace.Settings.LowSensory = !v.Workspace.Settings.LowSensory
	return v.Workspace.Settings.LowSensory, v.saveUnlocked()
}
func (v *Vault) CreateMatter(title string) (Matter, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	now := time.Now().UTC()
	m := Matter{ID: NewID("MAT"), Title: title, Status: "Active", NextAction: "Review and assign relevant evidence", CreatedAt: now, UpdatedAt: now}
	v.Workspace.Matters = append([]Matter{m}, v.Workspace.Matters...)
	v.addChangeUnlocked("user", "matter-created", "Created "+title, map[string]any{"id": m.ID})
	return m, v.saveUnlocked()
}

func (v *Vault) VerifyAll(progress func(current, total int, name string)) []string {
	v.opMu.RLock()
	defer v.opMu.RUnlock()
	alerts := []string{}
	ws := v.Snapshot()
	total := len(ws.Evidence)
	for i, e := range ws.Evidence {
		if progress != nil {
			progress(i, total, e.SafeName)
		}
		f, err := os.Open(filepath.Join(v.Objects, e.ObjectFile))
		if err != nil {
			alerts = append(alerts, e.SafeName+": encrypted object missing")
			continue
		}
		data, err := decryptObject(v.key, e.ID, f, 0)
		f.Close()
		if err != nil {
			alerts = append(alerts, e.SafeName+": authentication failed")
			continue
		}
		h := sha256.Sum256(data)
		if hex.EncodeToString(h[:]) != e.SHA256 {
			alerts = append(alerts, e.SafeName+": SHA-256 mismatch")
		}
	}
	v.AddChange("system", "integrity-check", "Verified encrypted evidence objects", map[string]any{"alerts": len(alerts), "items": total})
	return alerts
}

func stringsTrim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\n' || s[0] == '\r' || s[0] == '\t') {
		s = s[1:]
	}
	return s
}
func isImageType(t string) bool {
	switch t {
	case "jpeg", "png", "gif", "bmp", "tiff", "webp":
		return true
	}
	return false
}
func readFileBounded(path string, max int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() > max {
		return nil, fmt.Errorf("file exceeds bounded reader limit")
	}
	return io.ReadAll(io.LimitReader(f, max+1))
}
