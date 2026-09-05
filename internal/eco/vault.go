package eco

import (
	"bytes"
	"context"
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

var (
	ErrVaultClosed    = errors.New("vault is closed")
	ErrWorkspaceStale = errors.New("workspace metadata changed since it was loaded")
)

type Vault struct {
	Root                string
	Objects             string
	key                 []byte
	owner               *workspaceOwnerLease
	ownerTxn            string
	persistedRevision   uint64
	persistedMetaSHA256 string
	persistedChangeHead string
	persistedOwnerTxn   string
	closed              bool
	opMu                sync.RWMutex
	mu                  sync.Mutex
	Workspace           Workspace
}

func OpenVault(root string) (*Vault, error) {
	if root == "" {
		return nil, errors.New("empty vault root")
	}
	owner, err := acquireOrCreateWorkspaceRootOwner(root)
	if err != nil {
		return nil, err
	}
	root = owner.root
	v := &Vault{Root: root, Objects: filepath.Join(root, "objects"), owner: owner, ownerTxn: NewID("OWNER")}
	opened := false
	defer func() {
		if opened {
			return
		}
		zeroBytes(v.key)
		v.key = nil
		_ = owner.Close()
	}()
	if err := os.MkdirAll(v.Objects, 0700); err != nil {
		return nil, err
	}
	key, err := loadOrCreateMasterKey(filepath.Join(root, "vault.key"))
	if err != nil {
		return nil, fmt.Errorf("vault key: %w", err)
	}
	v.key = key
	if err := v.loadWorkspace(); err != nil {
		return nil, err
	}
	if err := cleanupInterruptedReadingCopies(root); err != nil {
		return nil, fmt.Errorf("clean interrupted derived reading state: %w", err)
	}
	if err := v.recoverPreservations(); err != nil {
		return nil, fmt.Errorf("recover preservation state: %w", err)
	}
	opened = true
	return v, nil
}

func (v *Vault) Close() error {
	if v == nil {
		return nil
	}
	v.opMu.Lock()
	defer v.opMu.Unlock()
	v.mu.Lock()
	if v.closed {
		v.mu.Unlock()
		return nil
	}
	v.closed = true
	owner := v.owner
	v.owner = nil
	zeroBytes(v.key)
	v.key = nil
	v.ownerTxn = ""
	v.persistedRevision = 0
	v.persistedMetaSHA256 = ""
	v.persistedChangeHead = ""
	v.persistedOwnerTxn = ""
	v.mu.Unlock()
	if owner != nil {
		return owner.Close()
	}
	return nil
}

func (v *Vault) ensureOpenUnlocked() error {
	if v == nil || v.closed || v.owner == nil {
		return ErrVaultClosed
	}
	if err := v.owner.revalidate(); err != nil {
		return fmt.Errorf("workspace ownership is invalid: %w", err)
	}
	return nil
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
	return Workspace{Schema: Schema, BuildID: BuildID, CreatedAt: now, UpdatedAt: now, Evidence: []EvidenceItem{}, Preservations: []PreservationRecord{}, Matters: []Matter{}, Changes: []ChangeRecord{}, Questions: []QuestionRecord{}, SelectedPage: "home"}
}

func (v *Vault) loadWorkspace() error {
	path := filepath.Join(v.Root, "workspace.ecodb")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		v.persistedRevision = 0
		v.persistedMetaSHA256 = ""
		v.persistedChangeHead = ""
		v.persistedOwnerTxn = ""
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
	v.persistedRevision = ws.Revision
	v.persistedMetaSHA256 = workspaceMetadataDigest(data)
	v.persistedChangeHead = workspaceChangeHead(ws)
	v.persistedOwnerTxn = ws.LastOwnerTxn
	v.Workspace = ws
	return nil
}

func (v *Vault) Save() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.saveUnlocked()
}

func (v *Vault) saveUnlocked() error {
	if err := v.ensureOpenUnlocked(); err != nil {
		return err
	}
	if err := v.verifyWorkspaceCASUnlocked(); err != nil {
		return err
	}

	oldRevision := v.Workspace.Revision
	oldOwnerTxn := v.Workspace.LastOwnerTxn
	oldUpdatedAt := v.Workspace.UpdatedAt
	oldBuildID := v.Workspace.BuildID
	restoreControl := func() {
		v.Workspace.Revision = oldRevision
		v.Workspace.LastOwnerTxn = oldOwnerTxn
		v.Workspace.UpdatedAt = oldUpdatedAt
		v.Workspace.BuildID = oldBuildID
	}

	baseRevision := v.persistedRevision
	if v.Workspace.Revision > baseRevision {
		baseRevision = v.Workspace.Revision
	}
	if baseRevision == ^uint64(0) {
		return errors.New("workspace revision exhausted")
	}
	v.Workspace.Revision = baseRevision + 1
	v.Workspace.LastOwnerTxn = v.ownerTxn
	v.Workspace.UpdatedAt = time.Now().UTC()
	v.Workspace.BuildID = BuildID
	plain, err := json.MarshalIndent(v.Workspace, "", "  ")
	if err != nil {
		restoreControl()
		return err
	}
	enc, err := encryptBlob(v.key, metaMagic, plain)
	if err != nil {
		restoreControl()
		return err
	}
	path := filepath.Join(v.Root, "workspace.ecodb")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, enc, 0600); err != nil {
		restoreControl()
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		restoreControl()
		return err
	}
	v.persistedRevision = v.Workspace.Revision
	v.persistedMetaSHA256 = workspaceMetadataDigest(enc)
	v.persistedChangeHead = workspaceChangeHead(v.Workspace)
	v.persistedOwnerTxn = v.Workspace.LastOwnerTxn
	return nil
}

func (v *Vault) verifyWorkspaceCASUnlocked() error {
	path := filepath.Join(v.Root, "workspace.ecodb")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		if v.persistedMetaSHA256 == "" && v.persistedRevision == 0 && v.persistedChangeHead == "" && v.persistedOwnerTxn == "" {
			return nil
		}
		return fmt.Errorf("%w: workspace metadata disappeared", ErrWorkspaceStale)
	}
	if err != nil {
		return fmt.Errorf("read workspace metadata for compare-and-swap: %w", err)
	}
	digest := workspaceMetadataDigest(data)
	if v.persistedMetaSHA256 == "" || digest != v.persistedMetaSHA256 {
		return fmt.Errorf("%w: encrypted metadata digest changed", ErrWorkspaceStale)
	}
	plain, err := decryptBlob(v.key, metaMagic, data)
	if err != nil {
		return fmt.Errorf("%w: current workspace authentication failed: %v", ErrWorkspaceStale, err)
	}
	var current Workspace
	if err := json.Unmarshal(plain, &current); err != nil {
		return fmt.Errorf("%w: current workspace format is invalid: %v", ErrWorkspaceStale, err)
	}
	if current.Revision != v.persistedRevision {
		return fmt.Errorf("%w: revision changed from %d to %d", ErrWorkspaceStale, v.persistedRevision, current.Revision)
	}
	if current.LastOwnerTxn != v.persistedOwnerTxn {
		return fmt.Errorf("%w: owner transaction changed", ErrWorkspaceStale)
	}
	if head := workspaceChangeHead(current); head != v.persistedChangeHead {
		return fmt.Errorf("%w: audit-chain head changed", ErrWorkspaceStale)
	}
	return nil
}

func workspaceMetadataDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func workspaceChangeHead(ws Workspace) string {
	if len(ws.Changes) == 0 {
		return ""
	}
	return ws.Changes[0].Hash
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
		if v.Workspace.Evidence[i].Extraction != nil {
			extraction := *v.Workspace.Evidence[i].Extraction
			out.Evidence[i].Extraction = &extraction
		}
		if v.Workspace.Evidence[i].Image != nil {
			img := *v.Workspace.Evidence[i].Image
			img.Warnings = append([]string(nil), v.Workspace.Evidence[i].Image.Warnings...)
			if v.Workspace.Evidence[i].Image.SuggestedCrop != nil {
				crop := *v.Workspace.Evidence[i].Image.SuggestedCrop
				img.SuggestedCrop = &crop
			}
			out.Evidence[i].Image = &img
		}
		if v.Workspace.Evidence[i].OCR != nil {
			ocr := *v.Workspace.Evidence[i].OCR
			ocr.Warnings = append([]string(nil), v.Workspace.Evidence[i].OCR.Warnings...)
			ocr.Words = append([]OCRWord(nil), v.Workspace.Evidence[i].OCR.Words...)
			ocr.Lines = append([]OCRLine(nil), v.Workspace.Evidence[i].OCR.Lines...)
			for j := range ocr.Lines {
				ocr.Lines[j].Words = append([]OCRWord(nil), v.Workspace.Evidence[i].OCR.Lines[j].Words...)
			}
			out.Evidence[i].OCR = &ocr
		}
	}
	out.Preservations = append([]PreservationRecord(nil), v.Workspace.Preservations...)
	out.Matters = append([]Matter(nil), v.Workspace.Matters...)
	for i := range out.Matters {
		out.Matters[i].EvidenceIDs = append([]string(nil), v.Workspace.Matters[i].EvidenceIDs...)
	}
	out.Changes = append([]ChangeRecord(nil), v.Workspace.Changes...)
	out.Questions = append([]QuestionRecord(nil), v.Workspace.Questions...)
	for i := range out.Questions {
		out.Questions[i].Citations = append([]Citation(nil), v.Workspace.Questions[i].Citations...)
		for j := range out.Questions[i].Citations {
			if v.Workspace.Questions[i].Citations[j].Region != nil {
				region := *v.Workspace.Questions[i].Citations[j].Region
				out.Questions[i].Citations[j].Region = &region
			}
		}
		out.Questions[i].ScopeIDs = append([]string(nil), v.Workspace.Questions[i].ScopeIDs...)
	}
	return out
}

// ApplyOCRResult records a coordinate-bearing local OCR result after the OCR
// worker has completed. The original encrypted evidence object is untouched.
func (v *Vault) ApplyOCRResult(evidenceID string, receipt OCRReceipt, segments []SourceSegment) error {
	if err := ValidateOCRReceipt(receipt); err != nil {
		return err
	}
	if err := validateOCRSegments(receipt, segments); err != nil {
		return err
	}
	v.opMu.RLock()
	defer v.opMu.RUnlock()
	v.mu.Lock()
	var source EvidenceItem
	found := false
	for i := range v.Workspace.Evidence {
		if v.Workspace.Evidence[i].ID == evidenceID {
			source = cloneEvidenceItem(v.Workspace.Evidence[i])
			found = true
			break
		}
	}
	v.mu.Unlock()
	if !found {
		return os.ErrNotExist
	}
	if !preservationUsable(source) {
		return errors.New("OCR is blocked because the preserved source is not verified")
	}
	if receipt.SourceObject != source.ObjectFile || receipt.SourceSHA256 != source.SHA256 {
		return errors.New("OCR receipt does not identify the verified preserved object")
	}
	for _, segment := range segments {
		if segment.SourceObject != source.ObjectFile || segment.SourceSHA256 != source.SHA256 {
			return errors.New("OCR source segment does not identify the verified preserved object")
		}
	}
	if _, err := v.verifyPreservedObject(source.ID, source.ObjectFile, source.SHA256, source.Size); err != nil {
		v.markEvidenceVerificationFailure(source.ID, err)
		return fmt.Errorf("OCR is blocked because preserved source verification failed: %w", err)
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	for i := range v.Workspace.Evidence {
		item := &v.Workspace.Evidence[i]
		if item.ID != evidenceID {
			continue
		}
		if !preservationUsable(*item) || receipt.SourceObject != item.ObjectFile || receipt.SourceSHA256 != item.SHA256 {
			return errors.New("OCR receipt source does not match the verified preserved object")
		}

		// Preserve a complete in-memory rollback point. The active workspace must
		// not retain an OCR result if the authenticated metadata write fails.
		oldItem := cloneEvidenceItem(*item)
		oldChangeLen := len(v.Workspace.Changes)
		oldUpdatedAt := v.Workspace.UpdatedAt
		oldBuildID := v.Workspace.BuildID

		kept := make([]SourceSegment, 0, len(item.Segments)+len(segments))
		for _, seg := range item.Segments {
			if seg.Origin != "ocr" {
				kept = append(kept, seg)
			}
		}
		for _, seg := range segments {
			copySeg := seg
			if seg.Region != nil {
				region := *seg.Region
				copySeg.Region = &region
			}
			kept = append(kept, copySeg)
		}
		item.Segments = kept
		copyReceipt := cloneOCRReceipt(receipt)
		item.OCR = &copyReceipt
		item.Readable = len(item.Segments) > 0
		switch receipt.Status {
		case "ready":
			item.Status = "OCR ready — review required"
		case "no-text":
			item.Status = "OCR found no text — review image"
		case "failed":
			item.Status = "OCR failed safely — original preserved"
		}
		item.Warnings = append(item.Warnings, receipt.Warnings...)
		v.addChangeUnlocked("ocr-worker", "ocr-result-added", "Added coordinate-bearing OCR reading for "+item.SafeName, map[string]any{"id": item.ID, "object_file": item.ObjectFile, "source_sha256": item.SHA256, "engine": receipt.Engine, "engine_version": receipt.EngineVersion, "status": receipt.Status, "lines": len(receipt.Lines), "average_confidence": receipt.AverageConfidence})
		if err := v.saveUnlocked(); err != nil {
			*item = oldItem
			v.Workspace.Changes = v.Workspace.Changes[:oldChangeLen]
			v.Workspace.UpdatedAt = oldUpdatedAt
			v.Workspace.BuildID = oldBuildID
			return fmt.Errorf("persist OCR result: %w", err)
		}
		return nil
	}
	return os.ErrNotExist
}

func cloneOCRReceipt(receipt OCRReceipt) OCRReceipt {
	out := receipt
	out.Warnings = append([]string(nil), receipt.Warnings...)
	out.Words = append([]OCRWord(nil), receipt.Words...)
	out.Lines = append([]OCRLine(nil), receipt.Lines...)
	for i := range out.Lines {
		out.Lines[i].Words = append([]OCRWord(nil), receipt.Lines[i].Words...)
	}
	return out
}

func cloneEvidenceItem(item EvidenceItem) EvidenceItem {
	out := item
	out.Warnings = append([]string(nil), item.Warnings...)
	out.MatterIDs = append([]string(nil), item.MatterIDs...)
	out.Segments = append([]SourceSegment(nil), item.Segments...)
	if item.Extraction != nil {
		extraction := *item.Extraction
		out.Extraction = &extraction
	}
	for i := range out.Segments {
		if item.Segments[i].Region != nil {
			region := *item.Segments[i].Region
			out.Segments[i].Region = &region
		}
	}
	if item.Image != nil {
		img := *item.Image
		img.Warnings = append([]string(nil), item.Image.Warnings...)
		if item.Image.SuggestedCrop != nil {
			crop := *item.Image.SuggestedCrop
			img.SuggestedCrop = &crop
		}
		out.Image = &img
	}
	if item.OCR != nil {
		ocr := cloneOCRReceipt(*item.OCR)
		out.OCR = &ocr
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
	if err := os.Chmod(dstPath, 0400); err != nil {
		return err
	}
	ok = true
	return nil
}

func encryptStreamContext(ctx context.Context, key []byte, objectID string, src io.Reader, dstPath string, size int64, progress func(ImportProgress), sourcePath, name string) (int64, string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, "", err
	}
	prefix := make([]byte, 4)
	if _, err := rand.Read(prefix); err != nil {
		return 0, "", err
	}
	if _, err := os.Lstat(dstPath); err == nil {
		return 0, "", errors.New("preserved object path already exists")
	} else if !os.IsNotExist(err) {
		return 0, "", err
	}
	tmp := dstPath + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return 0, "", err
	}
	closed := false
	keepTemp := false
	defer func() {
		if !closed {
			_ = out.Close()
		}
		if !keepTemp {
			_ = os.Remove(tmp)
		}
	}()
	if _, err := out.Write([]byte(objectMagic)); err != nil {
		return 0, "", err
	}
	if _, err := out.Write(prefix); err != nil {
		return 0, "", err
	}
	buf := make([]byte, chunkSize)
	h := sha256.New()
	var index uint64
	var done int64
	for {
		if err := ctx.Err(); err != nil {
			return done, hex.EncodeToString(h.Sum(nil)), err
		}
		n, readErr := src.Read(buf)
		if n > 0 {
			done += int64(n)
			if done > size {
				return done, hex.EncodeToString(h.Sum(nil)), errors.New("source grew during preservation")
			}
			_, _ = h.Write(buf[:n])
			nonce := make([]byte, gcm.NonceSize())
			copy(nonce, prefix)
			binary.BigEndian.PutUint64(nonce[len(nonce)-8:], index)
			sealed := gcm.Seal(nil, nonce, buf[:n], []byte(fmt.Sprintf("%s:%d", objectID, index)))
			if err := binary.Write(out, binary.LittleEndian, uint32(len(sealed))); err != nil {
				return done, hex.EncodeToString(h.Sum(nil)), err
			}
			if _, err := out.Write(sealed); err != nil {
				return done, hex.EncodeToString(h.Sum(nil)), err
			}
			index++
			if progress != nil {
				progress(ImportProgress{Path: sourcePath, Name: name, Stage: "Preserving immutable original", Current: done, Total: size})
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return done, hex.EncodeToString(h.Sum(nil)), readErr
		}
	}
	if done != size {
		return done, hex.EncodeToString(h.Sum(nil)), errors.New("source ended before the preservation write was complete")
	}
	if err := out.Sync(); err != nil {
		return done, hex.EncodeToString(h.Sum(nil)), err
	}
	if err := out.Close(); err != nil {
		return done, hex.EncodeToString(h.Sum(nil)), err
	}
	closed = true
	keepTemp = true
	if _, err := os.Lstat(dstPath); err == nil {
		return done, hex.EncodeToString(h.Sum(nil)), errors.New("preserved object was unexpectedly created or replaced")
	} else if !os.IsNotExist(err) {
		return done, hex.EncodeToString(h.Sum(nil)), err
	}
	if err := os.Rename(tmp, dstPath); err != nil {
		return done, hex.EncodeToString(h.Sum(nil)), err
	}
	keepTemp = false
	if err := os.Chmod(dstPath, 0400); err != nil {
		return done, hex.EncodeToString(h.Sum(nil)), fmt.Errorf("make preserved object immutable: %w", err)
	}
	return done, hex.EncodeToString(h.Sum(nil)), nil
}

func (v *Vault) ReadEvidence(id string, maxBytes int64) ([]byte, error) {
	data, _, err := v.ReadEvidenceSource(id, maxBytes)
	return data, err
}

// ReadEvidenceSource returns bytes only after a fresh authentication, size and
// SHA-256 check of the immutable preserved object. The receipt is the binding
// downstream preview and OCR work must carry forward.
func (v *Vault) ReadEvidenceSource(id string, maxBytes int64) ([]byte, SourceReceipt, error) {
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
		return nil, SourceReceipt{}, os.ErrNotExist
	}
	if !preservationUsable(*item) {
		return nil, SourceReceipt{}, errors.New("evidence source verification has not succeeded; preview, OCR and retrieval are blocked")
	}
	if maxBytes > 0 && item.Size > maxBytes {
		return nil, SourceReceipt{}, fmt.Errorf("evidence is %s; preview limit is %s", HumanBytes(item.Size), HumanBytes(maxBytes))
	}
	objectPath, err := v.objectPath(item.ID, item.ObjectFile)
	if err != nil {
		v.markEvidenceVerificationFailure(item.ID, err)
		return nil, SourceReceipt{}, err
	}
	f, err := os.Open(objectPath)
	if err != nil {
		v.markEvidenceVerificationFailure(item.ID, err)
		return nil, SourceReceipt{}, err
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil || !before.Mode().IsRegular() {
		err = errors.New("preserved object is not a regular immutable file")
		v.markEvidenceVerificationFailure(item.ID, err)
		return nil, SourceReceipt{}, err
	}
	data, err := decryptObject(v.key, id, f, maxBytes)
	if err == nil {
		after, statErr := f.Stat()
		current, pathErr := os.Stat(objectPath)
		if statErr != nil || pathErr != nil || !sameStableFile(before, after) || !sameStableFile(after, current) {
			err = errors.New("preserved object was replaced or mutated while it was being read")
		}
	}
	if err == nil && int64(len(data)) != item.Size {
		err = errors.New("preserved object size does not match its verified receipt")
	}
	if err == nil {
		h := sha256.Sum256(data)
		if hex.EncodeToString(h[:]) != item.SHA256 {
			err = errors.New("preserved object SHA-256 does not match its verified receipt")
		}
	}
	if err != nil {
		v.markEvidenceVerificationFailure(item.ID, err)
		return nil, SourceReceipt{}, err
	}
	receipt := SourceReceipt{EvidenceID: item.ID, ObjectFile: item.ObjectFile, SHA256: item.SHA256, Size: item.Size, VerifiedAt: time.Now().UTC()}
	return data, receipt, nil
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

func decryptObjectToWriter(key []byte, objectID string, src io.Reader, expectedSize int64, dst io.Writer) error {
	header := make([]byte, len(objectMagic))
	if _, err := io.ReadFull(src, header); err != nil {
		return err
	}
	if string(header) != objectMagic {
		return errors.New("bad object header")
	}
	prefix := make([]byte, 4)
	if _, err := io.ReadFull(src, prefix); err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	var index uint64
	var written int64
	for {
		var n uint32
		err := binary.Read(src, binary.LittleEndian, &n)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if n == 0 || n > chunkSize+uint32(gcm.Overhead()) {
			return errors.New("unsafe encrypted chunk length")
		}
		sealed := make([]byte, n)
		if _, err := io.ReadFull(src, sealed); err != nil {
			return err
		}
		nonce := make([]byte, gcm.NonceSize())
		copy(nonce, prefix)
		binary.BigEndian.PutUint64(nonce[len(nonce)-8:], index)
		plain, err := gcm.Open(nil, nonce, sealed, []byte(fmt.Sprintf("%s:%d", objectID, index)))
		if err != nil {
			return errors.New("object authentication failed")
		}
		written += int64(len(plain))
		if expectedSize >= 0 && written > expectedSize {
			return errors.New("decrypted object exceeds its preserved size receipt")
		}
		if _, err := dst.Write(plain); err != nil {
			return err
		}
		index++
	}
	if expectedSize >= 0 && written != expectedSize {
		return errors.New("decrypted object is incomplete")
	}
	return nil
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
		receipt, err := v.verifyPreservedObject(e.ID, e.ObjectFile, e.SHA256, e.Size)
		if err != nil {
			alerts = append(alerts, e.SafeName+": "+err.Error())
			v.markEvidenceVerificationFailure(e.ID, err)
			continue
		}
		v.markEvidenceVerificationSuccess(e.ID, receipt)
	}
	v.AddChange("system", "integrity-check", "Verified encrypted evidence objects", map[string]any{"alerts": len(alerts), "items": total})
	return alerts
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
