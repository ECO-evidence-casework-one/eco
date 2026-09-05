package eco

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func importSearchTestSource(t *testing.T, v *Vault, dir, name string) EvidenceItem {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("search test source"), 0600); err != nil {
		t.Fatal(err)
	}
	item, _, err := v.ImportFile(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func setSearchTestSegments(t *testing.T, v *Vault, evidenceID string, segments []SourceSegment) {
	t.Helper()
	v.mu.Lock()
	found := false
	for i := range v.Workspace.Evidence {
		item := &v.Workspace.Evidence[i]
		if item.ID != evidenceID {
			continue
		}
		for j := range segments {
			segments[j].SourceObject = item.ObjectFile
			segments[j].SourceSHA256 = item.SHA256
		}
		item.Segments = append([]SourceSegment(nil), segments...)
		item.Readable = len(segments) > 0
		found = true
		break
	}
	v.mu.Unlock()
	if !found {
		t.Fatalf("evidence %s not found", evidenceID)
	}
	if err := v.Save(); err != nil {
		t.Fatal(err)
	}
}

func TestSearchWorkspaceReturnsEveryPageAwareMatchInOrder(t *testing.T) {
	dir := t.TempDir()
	v, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	item := importSearchTestSource(t, v, dir, "one.txt")
	region := &NormalizedRegion{X: 0.1, Y: 0.2, Width: 0.4, Height: 0.1}
	setSearchTestSegments(t, v, item.ID, []SourceSegment{
		{ID: "SEG-3", Ordinal: 3, Page: 3, Text: "Final ALPHA result", Origin: "ocr", Confidence: 0.42, Region: region},
		{ID: "SEG-1", Ordinal: 1, Page: 1, Text: "Alpha first and alpha second", Origin: "native"},
		{ID: "SEG-2", Ordinal: 2, Page: 2, Text: "Nothing on this page", Origin: "native"},
	})

	receipt, err := v.SearchWorkspace("alpha", nil)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Engine != searchEngineVersion || receipt.Query != "alpha" {
		t.Fatalf("unexpected receipt identity: %+v", receipt)
	}
	if len(receipt.Sources) != 1 || len(receipt.Matches) != 3 {
		t.Fatalf("sources=%d matches=%d", len(receipt.Sources), len(receipt.Matches))
	}
	if receipt.Matches[0].Page != 1 || receipt.Matches[1].Page != 1 || receipt.Matches[2].Page != 3 {
		t.Fatalf("unexpected match order/pages: %+v", receipt.Matches)
	}
	if receipt.Matches[0].MatchText != "Alpha" || receipt.Matches[1].MatchText != "alpha" || receipt.Matches[2].MatchText != "ALPHA" {
		t.Fatalf("original match casing was not preserved: %+v", receipt.Matches)
	}
	last := receipt.Matches[2]
	if last.Origin != "ocr" || last.Confidence != 0.42 || last.Region == nil || !last.Region.Valid() {
		t.Fatalf("OCR provenance lost: %+v", last)
	}
	if last.SourceSHA256 != item.SHA256 || last.SourceObject != item.ObjectFile {
		t.Fatalf("source binding lost: %+v", last)
	}
	if err := v.ValidateSearchReceipt(receipt); err != nil {
		t.Fatalf("fresh receipt did not validate: %v", err)
	}
}

func TestSearchWorkspaceHonoursEvidenceScope(t *testing.T) {
	dir := t.TempDir()
	v, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	first := importSearchTestSource(t, v, dir, "first.txt")
	second := importSearchTestSource(t, v, dir, "second.txt")
	setSearchTestSegments(t, v, first.ID, []SourceSegment{{ID: "SEG-A", Ordinal: 1, Page: 1, Text: "alpha in first"}})
	setSearchTestSegments(t, v, second.ID, []SourceSegment{{ID: "SEG-B", Ordinal: 1, Page: 1, Text: "alpha in second"}})

	receipt, err := v.SearchWorkspace("alpha", []string{second.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Matches) != 1 || receipt.Matches[0].EvidenceID != second.ID {
		t.Fatalf("scope not enforced: %+v", receipt.Matches)
	}
	if len(receipt.Sources) != 1 || receipt.Sources[0].EvidenceID != second.ID {
		t.Fatalf("source scope not reflected in receipt: %+v", receipt.Sources)
	}
}

func TestSearchReceiptInvalidatesWhenReadableSegmentsChange(t *testing.T) {
	dir := t.TempDir()
	v, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	item := importSearchTestSource(t, v, dir, "one.txt")
	setSearchTestSegments(t, v, item.ID, []SourceSegment{{ID: "SEG-1", Ordinal: 1, Page: 1, Text: "alpha current wording"}})
	receipt, err := v.SearchWorkspace("alpha", nil)
	if err != nil {
		t.Fatal(err)
	}

	setSearchTestSegments(t, v, item.ID, []SourceSegment{{ID: "SEG-1", Ordinal: 1, Page: 1, Text: "beta replacement wording"}})
	if err := v.ValidateSearchReceipt(receipt); !errors.Is(err, ErrSearchReceiptStale) {
		t.Fatalf("changed reading validation error=%v", err)
	}
}

func TestSearchReceiptSurvivesUnrelatedWorkspaceWrite(t *testing.T) {
	dir := t.TempDir()
	v, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	item := importSearchTestSource(t, v, dir, "one.txt")
	setSearchTestSegments(t, v, item.ID, []SourceSegment{{ID: "SEG-1", Ordinal: 1, Page: 1, Text: "alpha stable wording"}})
	receipt, err := v.SearchWorkspace("alpha", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.SetSelectedPage("evidence"); err != nil {
		t.Fatal(err)
	}
	if err := v.ValidateSearchReceipt(receipt); err != nil {
		t.Fatalf("unrelated workspace write invalidated source receipt: %v", err)
	}
}

func TestSearchReceiptRejectsPreservedObjectTamper(t *testing.T) {
	dir := t.TempDir()
	v, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	item := importSearchTestSource(t, v, dir, "one.txt")
	setSearchTestSegments(t, v, item.ID, []SourceSegment{{ID: "SEG-1", Ordinal: 1, Page: 1, Text: "alpha stable wording"}})
	receipt, err := v.SearchWorkspace("alpha", nil)
	if err != nil {
		t.Fatal(err)
	}
	objectPath := filepath.Join(v.Objects, item.ObjectFile)
	data, err := os.ReadFile(objectPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 2 {
		t.Fatal("encrypted object unexpectedly short")
	}
	data[len(data)-1] ^= 0xff
	if err := os.WriteFile(objectPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := v.ValidateSearchReceipt(receipt); !errors.Is(err, ErrSearchReceiptStale) {
		t.Fatalf("tampered source validation error=%v", err)
	}
}

func TestSearchWorkspaceBoundsMatchCount(t *testing.T) {
	dir := t.TempDir()
	v, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	item := importSearchTestSource(t, v, dir, "one.txt")
	setSearchTestSegments(t, v, item.ID, []SourceSegment{{ID: "SEG-1", Ordinal: 1, Page: 1, Text: strings.Repeat("a ", maxSearchMatches+500)}})
	receipt, err := v.SearchWorkspace("a", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !receipt.Truncated || len(receipt.Matches) != maxSearchMatches {
		t.Fatalf("match bound not enforced: truncated=%v matches=%d", receipt.Truncated, len(receipt.Matches))
	}
}

func TestSearchWorkspaceRejectsClosedVault(t *testing.T) {
	dir := t.TempDir()
	v, err := OpenVault(filepath.Join(dir, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := v.SearchWorkspace("alpha", nil); !errors.Is(err, ErrVaultClosed) {
		t.Fatalf("closed search error=%v", err)
	}
}
