package eco

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestDoclingSourceRequiresKnownEvidence(t *testing.T) {
	v := &Vault{Workspace: Workspace{Evidence: []EvidenceItem{}, Preservations: []PreservationRecord{}}}
	_, _, err := v.doclingSource("EVD-missing")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}

func TestDoclingSourceRejectsInconsistentPreservation(t *testing.T) {
	now := time.Now().UTC()
	item := EvidenceItem{
		ID: "EVD-1", SafeName: "report.pdf", Size: 10,
		SHA256:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ObjectFile: "EVD-1.ecoobj", Preservation: preservationCommitted,
		SourceVerified: true, SourceVerifiedAt: now, DetectedType: "pdf",
	}
	v := &Vault{Workspace: Workspace{
		Evidence: []EvidenceItem{item},
		Preservations: []PreservationRecord{{
			EvidenceID: "EVD-1", ObjectFile: "EVD-1.ecoobj", State: preservationCommitted,
			PreservedSHA256: item.SHA256, ExpectedSize: 11,
		}},
	}}
	_, _, err := v.doclingSource("EVD-1")
	if err == nil {
		t.Fatal("expected inconsistent preservation record to be rejected")
	}
}

func TestDoclingSourceAcceptsVerifiedCommittedEvidence(t *testing.T) {
	now := time.Now().UTC()
	item := EvidenceItem{
		ID: "EVD-1", SafeName: "report.pdf", Size: 10,
		SHA256:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ObjectFile: "EVD-1.ecoobj", Preservation: preservationCommitted,
		SourceVerified: true, SourceVerifiedAt: now, DetectedType: "pdf",
	}
	record := PreservationRecord{
		EvidenceID: "EVD-1", ObjectFile: "EVD-1.ecoobj", State: preservationCommitted,
		PreservedSHA256: item.SHA256, ExpectedSize: item.Size,
	}
	v := &Vault{Workspace: Workspace{Evidence: []EvidenceItem{item}, Preservations: []PreservationRecord{record}}}
	gotItem, gotRecord, err := v.doclingSource("EVD-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotItem.ID != item.ID || gotRecord.EvidenceID != record.EvidenceID {
		t.Fatalf("wrong Docling source returned: %#v %#v", gotItem, gotRecord)
	}
}
