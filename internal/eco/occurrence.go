package eco

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	occurrenceKindInitial   = "initial-import"
	occurrenceKindDuplicate = "duplicate-import"
	occurrenceChangeType    = "evidence-occurrence-recorded"
	maxOccurrencePathRunes  = 32768
)

// EvidenceOccurrence describes one observed copy of already-preserved content.
// ECO preserves one encrypted evidence object per SHA-256 item, while this
// record preserves the fact that the same bytes were encountered again.
type EvidenceOccurrence struct {
	ID                        string    `json:"id"`
	EvidenceID                string    `json:"evidence_id"`
	Kind                      string    `json:"kind"`
	SourcePath                string    `json:"source_path,omitempty"`
	OriginalName              string    `json:"original_name"`
	Size                      int64     `json:"size"`
	SHA256                    string    `json:"sha256"`
	ObjectFile                string    `json:"object_file"`
	ObservedAt                time.Time `json:"observed_at"`
	SourceVerifiedAt          time.Time `json:"source_verified_at"`
	PreservedObjectVerifiedAt time.Time `json:"preserved_object_verified_at"`
	AuditChangeID             string    `json:"audit_change_id,omitempty"`
}

// EvidenceOccurrences returns the initial preserved source plus every later
// exact duplicate that ECO accepted after re-verifying both the incoming bytes
// and the retained encrypted object. Duplicate occurrences are stored as
// structured entries in ECO's encrypted, hash-chained change ledger, avoiding a
// workspace schema migration while still making the history queryable.
func (v *Vault) EvidenceOccurrences(evidenceID string) ([]EvidenceOccurrence, error) {
	evidenceID = strings.TrimSpace(evidenceID)
	if evidenceID == "" {
		return nil, errors.New("evidence occurrence query requires an evidence ID")
	}
	ws := v.Snapshot()
	var item EvidenceItem
	found := false
	for _, candidate := range ws.Evidence {
		if candidate.ID == evidenceID {
			item = candidate
			found = true
			break
		}
	}
	if !found {
		return nil, os.ErrNotExist
	}
	if !preservationUsable(item) {
		return nil, errors.New("evidence occurrence history is unavailable because the preserved evidence is not verified")
	}

	initialObserved := item.ImportedAt
	if initialObserved.IsZero() {
		initialObserved = item.SourceVerifiedAt
	}
	initialVerified := item.SourceVerifiedAt
	if initialVerified.IsZero() {
		initialVerified = initialObserved
	}
	initial := EvidenceOccurrence{
		ID:                        "OCC-INITIAL-" + item.ID,
		EvidenceID:                item.ID,
		Kind:                      occurrenceKindInitial,
		SourcePath:                item.SourcePath,
		OriginalName:              item.OriginalName,
		Size:                      item.Size,
		SHA256:                    item.SHA256,
		ObjectFile:                item.ObjectFile,
		ObservedAt:                initialObserved,
		SourceVerifiedAt:          initialVerified,
		PreservedObjectVerifiedAt: initialVerified,
	}
	if err := validateOccurrence(initial, item); err != nil {
		return nil, fmt.Errorf("initial evidence occurrence is invalid: %w", err)
	}

	out := []EvidenceOccurrence{initial}
	seen := map[string]bool{initial.ID: true}
	// Changes are stored newest first. Read oldest to newest so occurrence
	// history is naturally chronological before the final stable sort.
	for i := len(ws.Changes) - 1; i >= 0; i-- {
		change := ws.Changes[i]
		if change.Type != occurrenceChangeType {
			continue
		}
		if detailString(change.Details, "evidence_id") != item.ID {
			continue
		}
		occ, err := occurrenceFromChange(change)
		if err != nil {
			return nil, fmt.Errorf("occurrence audit record %s is invalid: %w", change.ID, err)
		}
		if err := validateOccurrence(occ, item); err != nil {
			return nil, fmt.Errorf("occurrence audit record %s is inconsistent with evidence: %w", change.ID, err)
		}
		if seen[occ.ID] {
			return nil, fmt.Errorf("duplicate occurrence identifier %q", occ.ID)
		}
		seen[occ.ID] = true
		out = append(out, occ)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ObservedAt.Equal(out[j].ObservedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].ObservedAt.Before(out[j].ObservedAt)
	})
	return out, nil
}

func (v *Vault) recordDuplicateOccurrence(item EvidenceItem, sourcePath string, sourceInfo os.FileInfo, sourceHash string, sourceVerifiedAt time.Time, preserved SourceReceipt) error {
	if !preservationUsable(item) {
		return errors.New("cannot record a duplicate occurrence against unverified evidence")
	}
	if sourceInfo == nil || !sourceInfo.Mode().IsRegular() {
		return errors.New("duplicate occurrence source is not a regular file")
	}
	if sourceHash != item.SHA256 || sourceInfo.Size() != item.Size {
		return errors.New("duplicate occurrence bytes do not exactly match the retained evidence")
	}
	if preserved.EvidenceID != item.ID || preserved.ObjectFile != item.ObjectFile || preserved.SHA256 != item.SHA256 || preserved.Size != item.Size || preserved.VerifiedAt.IsZero() {
		return errors.New("duplicate occurrence is not backed by a fresh preserved-object verification")
	}
	if sourceVerifiedAt.IsZero() {
		return errors.New("duplicate source verification time is missing")
	}
	cleanPath := filepath.Clean(sourcePath)
	occ := EvidenceOccurrence{
		ID:                        NewID("OCC"),
		EvidenceID:                item.ID,
		Kind:                      occurrenceKindDuplicate,
		SourcePath:                cleanPath,
		OriginalName:              sourceInfo.Name(),
		Size:                      sourceInfo.Size(),
		SHA256:                    sourceHash,
		ObjectFile:                item.ObjectFile,
		ObservedAt:                time.Now().UTC(),
		SourceVerifiedAt:          sourceVerifiedAt,
		PreservedObjectVerifiedAt: preserved.VerifiedAt,
	}
	if err := validateOccurrence(occ, item); err != nil {
		return err
	}

	details := map[string]any{
		"occurrence_id":                occ.ID,
		"evidence_id":                  occ.EvidenceID,
		"kind":                         occ.Kind,
		"source_path":                  occ.SourcePath,
		"original_name":                occ.OriginalName,
		"source_size":                  strconv.FormatInt(occ.Size, 10),
		"source_sha256":                occ.SHA256,
		"object_file":                  occ.ObjectFile,
		"observed_at":                  occ.ObservedAt.Format(time.RFC3339Nano),
		"source_verified_at":           occ.SourceVerifiedAt.Format(time.RFC3339Nano),
		"preserved_object_verified_at": occ.PreservedObjectVerifiedAt.Format(time.RFC3339Nano),
		"duplicate_reused_object":      true,
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	oldChanges := append([]ChangeRecord(nil), v.Workspace.Changes...)
	oldUpdatedAt := v.Workspace.UpdatedAt
	oldBuildID := v.Workspace.BuildID
	v.addChangeUnlocked("system", occurrenceChangeType, "Recorded a verified duplicate occurrence for "+item.SafeName, details)
	if err := v.saveUnlocked(); err != nil {
		v.Workspace.Changes = oldChanges
		v.Workspace.UpdatedAt = oldUpdatedAt
		v.Workspace.BuildID = oldBuildID
		return fmt.Errorf("persist duplicate evidence occurrence: %w", err)
	}
	return nil
}

func occurrenceFromChange(change ChangeRecord) (EvidenceOccurrence, error) {
	if change.Type != occurrenceChangeType {
		return EvidenceOccurrence{}, errors.New("wrong change type")
	}
	sizeText := detailString(change.Details, "source_size")
	size, err := strconv.ParseInt(sizeText, 10, 64)
	if err != nil {
		return EvidenceOccurrence{}, errors.New("source size is missing or invalid")
	}
	observedAt, err := parseOccurrenceTime(detailString(change.Details, "observed_at"))
	if err != nil {
		return EvidenceOccurrence{}, fmt.Errorf("observed time: %w", err)
	}
	sourceVerifiedAt, err := parseOccurrenceTime(detailString(change.Details, "source_verified_at"))
	if err != nil {
		return EvidenceOccurrence{}, fmt.Errorf("source verification time: %w", err)
	}
	preservedVerifiedAt, err := parseOccurrenceTime(detailString(change.Details, "preserved_object_verified_at"))
	if err != nil {
		return EvidenceOccurrence{}, fmt.Errorf("preserved-object verification time: %w", err)
	}
	if reused, ok := change.Details["duplicate_reused_object"].(bool); !ok || !reused {
		return EvidenceOccurrence{}, errors.New("duplicate reuse marker is missing")
	}
	return EvidenceOccurrence{
		ID:                        detailString(change.Details, "occurrence_id"),
		EvidenceID:                detailString(change.Details, "evidence_id"),
		Kind:                      detailString(change.Details, "kind"),
		SourcePath:                detailString(change.Details, "source_path"),
		OriginalName:              detailString(change.Details, "original_name"),
		Size:                      size,
		SHA256:                    detailString(change.Details, "source_sha256"),
		ObjectFile:                detailString(change.Details, "object_file"),
		ObservedAt:                observedAt,
		SourceVerifiedAt:          sourceVerifiedAt,
		PreservedObjectVerifiedAt: preservedVerifiedAt,
		AuditChangeID:             change.ID,
	}, nil
}

func validateOccurrence(occ EvidenceOccurrence, item EvidenceItem) error {
	if len(occ.ID) < 1 || len(occ.ID) > 160 {
		return errors.New("occurrence ID is missing or unbounded")
	}
	if occ.EvidenceID != item.ID || occ.SHA256 != item.SHA256 || occ.Size != item.Size || occ.ObjectFile != item.ObjectFile {
		return errors.New("occurrence does not identify the exact preserved evidence bytes")
	}
	if occ.Kind != occurrenceKindInitial && occ.Kind != occurrenceKindDuplicate {
		return errors.New("occurrence kind is invalid")
	}
	if len([]rune(occ.OriginalName)) < 1 || len([]rune(occ.OriginalName)) > 512 {
		return errors.New("occurrence original name is missing or unbounded")
	}
	if len([]rune(occ.SourcePath)) > maxOccurrencePathRunes {
		return errors.New("occurrence source path is unbounded")
	}
	if occ.ObservedAt.IsZero() || occ.SourceVerifiedAt.IsZero() || occ.PreservedObjectVerifiedAt.IsZero() {
		return errors.New("occurrence verification timestamps are incomplete")
	}
	return nil
}

func detailString(details map[string]any, key string) string {
	if details == nil {
		return ""
	}
	value, ok := details[key]
	if !ok {
		return ""
	}
	s, _ := value.(string)
	return strings.TrimSpace(s)
}

func parseOccurrenceTime(text string) (time.Time, error) {
	if strings.TrimSpace(text) == "" {
		return time.Time{}, errors.New("timestamp is missing")
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}
