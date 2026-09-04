package eco

import (
	"context"
	"strings"
	"testing"
)

func TestLlamaCPPWorkflowReleasesOnlyGroundedSourceText(t *testing.T) {
	v, _ := testGroundingVault(t)
	fake := func(_ context.Context, _, _ string, grounding GroundingContext) (LlamaCPPModelResult, error) {
		var record GroundingRecord
		for _, candidate := range grounding.Records {
			if strings.Contains(candidate.Text, "12 August 2026") {
				record = candidate
				break
			}
		if record.EvidenceID == "" {
			t.Fatal("fake runner did not receive the expected source wording")
		}
		return LlamaCPPModelResult{
			Emission: GroundingEmission{
				Answer: "FABRICATED FREE-FORM DRAFT THAT ECO MUST NOT RELEASE",
				Claims: []GroundingClaim{{Kind: "value", Text: "12 August 2026", EvidenceID: record.EvidenceID, SegmentID: record.SegmentID}},
			},
			EngineVersion: "llama.cpp test-build",
			ModelName:     "test.gguf",
			ModelSHA256:   strings.Repeat("a", 64),
		}, nil
	}

	result, err := v.askWithLlamaCPPRunner(context.Background(), "When is the hearing?", nil, "ignored", "ignored", fake)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result.Question.Answer, "FABRICATED") {
		t.Fatalf("model draft escaped into released answer: %q", result.Question.Answer)
	}
	if !strings.Contains(result.Question.Answer, "12 August 2026") {
		t.Fatalf("released answer omitted grounded source wording: %q", result.Question.Answer)
	}
	if !result.Grounding.AllClaimsGrounded || result.Grounding.SemanticTruthVerified {
		t.Fatalf("unexpected grounding report: %+v", result.Grounding)
	}
	if len(result.Question.Citations) != 1 || result.Question.Citations[0].SourceSHA256 == "" {
		t.Fatalf("released question is not bound to a verified source: %+v", result.Question)
	}
	ws := v.Snapshot()
	if len(ws.Questions) != 1 || ws.Questions[0].ID != result.Question.ID {
		t.Fatalf("grounded local AI question was not persisted exactly once: %+v", ws.Questions)
	}
	foundAudit := false
	for _, change := range ws.Changes {
		if change.Type == "grounded-local-ai-question" {
			foundAudit = true
			if released, _ := change.Details["model_draft_released"].(bool); released {
				t.Fatal("audit log incorrectly claims the free-form model draft was released")
			}
			break
		}
	}
	if !foundAudit {
		t.Fatal("missing grounded local AI audit record")
	}
}

func TestLlamaCPPWorkflowRejectsFabricatedClaimWithoutPersistingQuestion(t *testing.T) {
	v, _ := testGroundingVault(t)
	fake := func(_ context.Context, _, _ string, grounding GroundingContext) (LlamaCPPModelResult, error) {
		record := grounding.Records[0]
		return LlamaCPPModelResult{
			Emission: GroundingEmission{
				Answer: "The hearing is on 13 August 2026.",
				Claims: []GroundingClaim{{Kind: "value", Text: "13 August 2026", EvidenceID: record.EvidenceID, SegmentID: record.SegmentID}},
			},
			EngineVersion: "llama.cpp test-build",
			ModelName:     "test.gguf",
			ModelSHA256:   strings.Repeat("b", 64),
		}, nil
	}

	result, err := v.askWithLlamaCPPRunner(context.Background(), "When is the hearing?", nil, "ignored", "ignored", fake)
	if err == nil || !strings.Contains(err.Error(), "no model answer was released") {
		t.Fatalf("expected fabricated claim to fail closed, got result=%+v err=%v", result, err)
	}
	if result.Grounding.AllClaimsGrounded {
		t.Fatalf("fabricated claim unexpectedly grounded: %+v", result.Grounding)
	}
	ws := v.Snapshot()
	if len(ws.Questions) != 0 {
		t.Fatalf("rejected model emission created a trusted question record: %+v", ws.Questions)
	}
	foundRejection := false
	for _, change := range ws.Changes {
		if change.Type == "local-ai-grounding-rejected" {
			foundRejection = true
			break
		}
	}
	if !foundRejection {
		t.Fatal("rejected model emission was not recorded in the audit chain")
	}
}

func TestRenderGroundedLlamaAnswerDeduplicatesExactClaims(t *testing.T) {
	citation := Citation{EvidenceID: "E1", SegmentID: "S1", Label: "source", Quote: "Exact source wording"}
	answer := renderGroundedLlamaAnswer([]Citation{citation, citation})
	if strings.Count(answer, "Exact source wording") != 1 {
		t.Fatalf("duplicate grounded claim was rendered more than once: %q", answer)
	}
}
