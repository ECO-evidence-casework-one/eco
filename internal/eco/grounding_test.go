package eco

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testGroundingVault(t *testing.T) (*Vault, GroundingContext) {
	t.Helper()
	d := t.TempDir()
	src := filepath.Join(d, "hearing.txt")
	content := []byte("The hearing is on 12 August 2026. The council requested the medical report by 5 August 2026.")
	if err := os.WriteFile(src, content, 0600); err != nil {
		t.Fatal(err)
	}
	v, err := OpenVault(filepath.Join(d, "vault"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := v.ImportFile(src, nil); err != nil {
		t.Fatal(err)
	}
	ctx, err := v.BuildGroundingContext("When is the hearing?", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(ctx.Records) == 0 {
		t.Fatal("expected at least one grounding record")
	}
	return v, ctx
}

func TestGroundingContextKeepsTrustedHashesOutOfModelJSON(t *testing.T) {
	_, ctx := testGroundingVault(t)
	if !strings.HasPrefix(ctx.ContextID, "GCTX-") {
		t.Fatalf("unexpected context ID: %q", ctx.ContextID)
	}
	payload, err := json.Marshal(ctx)
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	if strings.Contains(text, "source_sha256") || strings.Contains(text, ".ecoobj") || strings.Contains(text, "trusted") {
		t.Fatalf("model-facing grounding context leaked trusted source metadata: %s", text)
	}
	if !strings.Contains(text, ctx.Records[0].EvidenceID) || !strings.Contains(text, ctx.Records[0].SegmentID) {
		t.Fatalf("model-facing context omitted its allowed citation vocabulary: %s", text)
	}
}

func TestGroundingEmissionAcceptsSourceTextAndRejectsFabrication(t *testing.T) {
	v, ctx := testGroundingVault(t)
	record := ctx.Records[0]
	claimText := "12 August 2026"
	if !strings.Contains(record.Text, claimText) {
		t.Fatalf("retrieval context did not contain expected text: %q", record.Text)
	}

	grounded := GroundingEmission{
		Answer: "The hearing is on 12 August 2026.",
		Claims: []GroundingClaim{{Kind: "value", Text: claimText, EvidenceID: record.EvidenceID, SegmentID: record.SegmentID}},
	}
	report, citations, err := v.VerifyGroundingEmission(ctx, grounded)
	if err != nil {
		t.Fatal(err)
	}
	if !report.AllClaimsGrounded || report.SemanticTruthVerified || len(report.Checks) != 1 || report.Checks[0].Status != "grounded" || len(citations) != 1 {
		t.Fatalf("unexpected grounded result: report=%+v citations=%+v", report, citations)
	}
	if citations[0].SourceSHA256 == "" || citations[0].SourceObject == "" {
		t.Fatalf("hydrated citation is not rebound to trusted preserved-source metadata: %+v", citations[0])
	}

	fabricated := grounded
	fabricated.Answer = "The hearing is on 13 August 2026."
	fabricated.Claims = []GroundingClaim{{Kind: "value", Text: "13 August 2026", EvidenceID: record.EvidenceID, SegmentID: record.SegmentID}}
	report, citations, err = v.VerifyGroundingEmission(ctx, fabricated)
	if err != nil {
		t.Fatal(err)
	}
	if report.AllClaimsGrounded || len(citations) != 0 || len(report.Checks) != 1 || report.Checks[0].Status != "mismatch" || report.Checks[0].Reason != "text_mismatch" {
		t.Fatalf("fabricated text did not fail closed: report=%+v citations=%+v", report, citations)
	}
}

func TestGroundingEmissionRejectsUnshownIDsAndReconstructedContext(t *testing.T) {
	v, ctx := testGroundingVault(t)
	record := ctx.Records[0]

	_, _, err := v.VerifyGroundingEmission(ctx, GroundingEmission{
		Answer: "A claim.",
		Claims: []GroundingClaim{{Kind: "quote", Text: "12 August 2026", EvidenceID: record.EvidenceID, SegmentID: "invented-segment"}},
	})
	if err == nil || !strings.Contains(err.Error(), "out_of_vocabulary") {
		t.Fatalf("expected invented segment ID to fail closed, got %v", err)
	}

	reconstructed := GroundingContext{
		ContextID: ctx.ContextID,
		Question:  ctx.Question,
		Records:   append([]GroundingRecord(nil), ctx.Records...),
	}
	_, _, err = v.VerifyGroundingEmission(reconstructed, GroundingEmission{
		Answer: "A claim.",
		Claims: []GroundingClaim{{Kind: "quote", Text: "12 August 2026", EvidenceID: record.EvidenceID, SegmentID: record.SegmentID}},
	})
	if err == nil || !strings.Contains(err.Error(), "reconstructed") {
		t.Fatalf("expected reconstructed public JSON context to be non-authoritative, got %v", err)
	}
}

func TestGroundingEmissionRejectsMalformedClaims(t *testing.T) {
	v, ctx := testGroundingVault(t)
	record := ctx.Records[0]
	cases := []GroundingClaim{
		{Kind: "presence", Text: "not allowed", EvidenceID: record.EvidenceID, SegmentID: record.SegmentID},
		{Kind: "table_cell", Text: "value", EvidenceID: record.EvidenceID, SegmentID: record.SegmentID},
		{Kind: "quote", Text: "", EvidenceID: record.EvidenceID, SegmentID: record.SegmentID},
	}
	for i, claim := range cases {
		if _, _, err := v.VerifyGroundingEmission(ctx, GroundingEmission{Answer: "A claim.", Claims: []GroundingClaim{claim}}); err == nil {
			t.Fatalf("case %d: expected malformed claim rejection", i+1)
		}
	}
}

func TestNormalizeGroundingTextOnlyNormalizesWhitespace(t *testing.T) {
	got := normalizeGroundingText("Alpha\n\tBeta   Gamma")
	if got != "Alpha Beta Gamma" {
		t.Fatalf("unexpected normalization: %q", got)
	}
	if normalizeGroundingText("Alpha") == normalizeGroundingText("alpha") {
		t.Fatal("grounding normalization unexpectedly erased case differences")
	}
}
