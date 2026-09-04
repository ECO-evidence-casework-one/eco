package eco

import (
	"strings"
	"testing"
)

func TestGroundingEmissionRejectsMutatedShownRecord(t *testing.T) {
	v, ctx := testGroundingVault(t)
	record := ctx.Records[0]
	ctx.Records[0].Text = "fabricated replacement text"
	_, _, err := v.VerifyGroundingEmission(ctx, GroundingEmission{
		Answer: "A claim.",
		Claims: []GroundingClaim{{Kind: "presence", EvidenceID: record.EvidenceID, SegmentID: record.SegmentID}},
	})
	if err == nil || !strings.Contains(err.Error(), "mutated") {
		t.Fatalf("expected modified shown context to fail closed, got %v", err)
	}
}
