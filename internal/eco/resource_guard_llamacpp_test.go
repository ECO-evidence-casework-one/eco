package eco

import (
	"context"
	"strings"
	"testing"
)

func TestLlamaCPPWorkflowCarriesResourceAssessmentIntoAudit(t *testing.T) {
	v, _ := testGroundingVault(t)
	resources := ResourceAssessment{
		Level:   "elevated",
		Blocked: false,
		Warnings: []string{"CPU usage is already high"},
		Snapshot: ResourceSnapshot{
			CPUSampled:           true,
			CPUUsedPercent:       96,
			MemorySampled:        true,
			MemoryAvailableBytes: 2 * resourceGiB,
			MemoryUsedPercent:    75,
			DiskSampled:          true,
			DiskFreeBytes:        20 * resourceGiB,
		},
	}
	fake := func(_ context.Context, _, _ string, grounding GroundingContext) (LlamaCPPModelResult, error) {
		var record GroundingRecord
		for _, candidate := range grounding.Records {
			if strings.Contains(candidate.Text, "12 August 2026") {
				record = candidate
				break
			}
		}
		if record.EvidenceID == "" {
			t.Fatal("fake runner did not receive expected grounded wording")
		}
		return LlamaCPPModelResult{
			Emission: GroundingEmission{
				Answer: "draft",
				Claims: []GroundingClaim{{Kind: "value", Text: "12 August 2026", EvidenceID: record.EvidenceID, SegmentID: record.SegmentID}},
			},
			EngineVersion: "llama.cpp test-build",
			ModelName:     "test.gguf",
			ModelSHA256:   strings.Repeat("c", 64),
			Resources:     resources,
		}, nil
	}

	result, err := v.askWithLlamaCPPRunner(context.Background(), "When is the hearing?", nil, "ignored", "ignored", fake)
	if err != nil {
		t.Fatal(err)
	}
	if result.Resources.Level != "elevated" || result.Resources.Blocked {
		t.Fatalf("resource assessment did not survive result boundary: %+v", result.Resources)
	}

	ws := v.Snapshot()
	found := false
	for _, change := range ws.Changes {
		if change.Type != "grounded-local-ai-question" {
			continue
		}
		found = true
		if got, _ := change.Details["resource_level"].(string); got != "elevated" {
			t.Fatalf("audit resource level mismatch: %+v", change.Details)
		}
		if got, _ := change.Details["resource_warning_count"].(int); got != 1 {
			t.Fatalf("audit warning count mismatch: %+v", change.Details)
		}
		if blocked, _ := change.Details["resource_blocked"].(bool); blocked {
			t.Fatalf("audit incorrectly marked warning-only run as blocked: %+v", change.Details)
		}
		break
	}
	if !found {
		t.Fatal("missing grounded local AI audit entry")
	}
}

func TestLlamaCPPWorkflowAuditsCriticalResourceBlockWithoutQuestion(t *testing.T) {
	v, _ := testGroundingVault(t)
	resources := ResourceAssessment{
		Level:   "critical",
		Blocked: true,
		Reasons: []string{"only 128 MiB RAM is available"},
		Snapshot: ResourceSnapshot{
			MemorySampled:        true,
			MemoryAvailableBytes: 128 * resourceMiB,
			MemoryUsedPercent:    99,
		},
	}
	fake := func(_ context.Context, _, _ string, _ GroundingContext) (LlamaCPPModelResult, error) {
		return LlamaCPPModelResult{
			ModelName:   "test.gguf",
			ModelSHA256: strings.Repeat("d", 64),
			Resources:   resources,
		}, &ResourcePressureError{Engine: "llama.cpp", Reasons: resources.Reasons}
	}

	result, err := v.askWithLlamaCPPRunner(context.Background(), "When is the hearing?", nil, "ignored", "ignored", fake)
	if err == nil {
		t.Fatal("expected critical resource pressure to block the local model")
	}
	if !result.Resources.Blocked || result.Resources.Level != "critical" {
		t.Fatalf("blocked assessment was not returned to caller: %+v", result.Resources)
	}
	ws := v.Snapshot()
	if len(ws.Questions) != 0 {
		t.Fatalf("critical resource block created a trusted question record: %+v", ws.Questions)
	}
	found := false
	for _, change := range ws.Changes {
		if change.Type != "local-ai-resource-blocked" {
			continue
		}
		found = true
		if got, _ := change.Details["resource_level"].(string); got != "critical" {
			t.Fatalf("blocked audit resource level mismatch: %+v", change.Details)
		}
		if started, _ := change.Details["generation_started"].(bool); started {
			t.Fatalf("blocked audit claims model generation started: %+v", change.Details)
		}
		break
	}
	if !found {
		t.Fatal("missing critical resource-block audit entry")
	}
}
