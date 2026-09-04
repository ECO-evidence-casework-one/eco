package eco

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type LlamaCPPAnswerResult struct {
	Question      QuestionRecord    `json:"question"`
	Grounding     GroundingReport   `json:"grounding"`
	EngineVersion string            `json:"engine_version"`
	ModelName     string            `json:"model_name"`
	ModelSHA256   string            `json:"model_sha256"`
	ContextID     string            `json:"context_id"`
	Resources     ResourceAssessment `json:"resources"`
}

type llamaCPPRunner func(context.Context, string, string, GroundingContext) (LlamaCPPModelResult, error)

// AskWithLlamaCPP runs a local GGUF model through llama-cli, but does not trust
// or release the model's free-form draft. The model selects source claims; ECO
// verifies every claim against the exact context it showed, re-verifies the
// preserved objects, then deterministically renders the released answer from
// those grounded claims.
func (v *Vault) AskWithLlamaCPP(question string, scopeIDs []string, executable, modelPath string) (LlamaCPPAnswerResult, error) {
	return v.AskWithLlamaCPPContext(context.Background(), question, scopeIDs, executable, modelPath)
}

func (v *Vault) AskWithLlamaCPPContext(ctx context.Context, question string, scopeIDs []string, executable, modelPath string) (LlamaCPPAnswerResult, error) {
	return v.askWithLlamaCPPRunner(ctx, question, scopeIDs, executable, modelPath, RunLlamaCPP)
}

func (v *Vault) askWithLlamaCPPRunner(ctx context.Context, question string, scopeIDs []string, executable, modelPath string, runner llamaCPPRunner) (LlamaCPPAnswerResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	question = strings.TrimSpace(question)
	if question == "" {
		return LlamaCPPAnswerResult{}, errors.New("local AI question is required")
	}
	if runner == nil {
		return LlamaCPPAnswerResult{}, errors.New("llama.cpp runner is required")
	}
	grounding, err := v.BuildGroundingContext(question, scopeIDs)
	if err != nil {
		return LlamaCPPAnswerResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return LlamaCPPAnswerResult{}, err
	}

	modelResult, err := runner(ctx, executable, modelPath, grounding)
	if err != nil {
		if modelResult.Resources.Blocked {
			_ = v.recordLlamaResourceBlock(question, modelResult, grounding)
		}
		return LlamaCPPAnswerResult{
			EngineVersion: modelResult.EngineVersion,
			ModelName:     modelResult.ModelName,
			ModelSHA256:   modelResult.ModelSHA256,
			ContextID:     grounding.ContextID,
			Resources:     modelResult.Resources,
		}, err
	}
	report, citations, err := v.VerifyGroundingEmission(grounding, modelResult.Emission)
	result := LlamaCPPAnswerResult{
		Grounding:     report,
		EngineVersion: modelResult.EngineVersion,
		ModelName:     modelResult.ModelName,
		ModelSHA256:   modelResult.ModelSHA256,
		ContextID:     grounding.ContextID,
		Resources:     modelResult.Resources,
	}
	if err != nil {
		_ = v.recordLlamaGroundingRejection(question, modelResult, grounding, "verification_error")
		return result, err
	}
	if !report.AllClaimsGrounded {
		if auditErr := v.recordLlamaGroundingRejection(question, modelResult, grounding, "claim_mismatch"); auditErr != nil {
			return result, fmt.Errorf("llama.cpp grounding rejected and audit persistence failed: %w", auditErr)
		}
		return result, errors.New("llama.cpp grounding rejected one or more claims; no model answer was released")
	}
	if len(citations) != len(modelResult.Emission.Claims) {
		return result, errors.New("llama.cpp grounded citation count does not match its claims")
	}

	answer := renderGroundedLlamaAnswer(citations)
	if strings.TrimSpace(answer) == "" {
		return result, errors.New("llama.cpp claims grounded but produced no releasable source wording")
	}
	rec := QuestionRecord{
		ID:                           NewID("Q"),
		AskedAt:                      time.Now().UTC(),
		Question:                     question,
		Intent:                       classifyIntent(question),
		Answer:                       answer,
		Citations:                    citations,
		Support:                      "Local llama.cpp selected the passages; ECO released only deterministically grounded source wording. Source truth, completeness, relevance and legal correctness remain unverified.",
		ScopeIDs:                     append([]string(nil), scopeIDs...),
		ReceiptID:                    NewID("AIR"),
		EvidenceConsidered:           countVerifiedEvidenceForScope(v.Snapshot(), scopeIDs),
		RetrievedSegments:            len(grounding.Records),
		SuspiciousSourcesExcluded:    grounding.SuspiciousSourcesExcluded,
		LowConfidenceSourcesExcluded: grounding.LowConfidenceSourcesExcluded,
		SourceVerificationFailures:   grounding.SourceVerificationFailures,
	}
	if err := v.persistLlamaQuestion(rec, modelResult, grounding, report); err != nil {
		return result, err
	}
	result.Question = rec
	return result, nil
}

func renderGroundedLlamaAnswer(citations []Citation) string {
	if len(citations) == 0 {
		return ""
	}
	lines := make([]string, 0, len(citations))
	seen := map[string]bool{}
	for _, citation := range citations {
		quote := strings.TrimSpace(citation.Quote)
		if quote == "" {
			continue
		}
		key := citation.EvidenceID + "\x00" + citation.SegmentID + "\x00" + quote
		if seen[key] {
			continue
		}
		seen[key] = true
		lines = append(lines, "• "+quote+" ["+citation.Label+"]")
	}
	if len(lines) == 0 {
		return ""
	}
	return "ECO's local model selected these source-grounded findings. Check the cited material before relying on its truth or significance:\r\n\r\n" + strings.Join(lines, "\r\n\r\n")
}

func countVerifiedEvidenceForScope(ws Workspace, scopeIDs []string) int {
	allowed := make(map[string]bool, len(scopeIDs))
	for _, id := range scopeIDs {
		allowed[id] = true
	}
	count := 0
	for _, item := range ws.Evidence {
		if len(allowed) > 0 && !allowed[item.ID] {
			continue
		}
		if preservationUsable(item) {
			count++
		}
	}
	return count
}

func addResourceAuditDetails(details map[string]any, assessment ResourceAssessment) {
	level := strings.TrimSpace(assessment.Level)
	if level == "" {
		level = "not-recorded"
	}
	details["resource_level"] = level
	details["resource_blocked"] = assessment.Blocked
	details["resource_warning_count"] = len(assessment.Warnings)
	if assessment.Snapshot.MemorySampled {
		details["memory_available_bytes"] = assessment.Snapshot.MemoryAvailableBytes
		details["memory_used_percent"] = assessment.Snapshot.MemoryUsedPercent
	}
	if assessment.Snapshot.DiskSampled {
		details["disk_free_bytes"] = assessment.Snapshot.DiskFreeBytes
	}
	if assessment.Snapshot.CPUSampled {
		details["cpu_used_percent"] = assessment.Snapshot.CPUUsedPercent
	}
}

func (v *Vault) persistLlamaQuestion(rec QuestionRecord, model LlamaCPPModelResult, grounding GroundingContext, report GroundingReport) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	oldQuestions := append([]QuestionRecord(nil), v.Workspace.Questions...)
	oldChanges := append([]ChangeRecord(nil), v.Workspace.Changes...)
	oldUpdatedAt := v.Workspace.UpdatedAt
	oldBuildID := v.Workspace.BuildID
	v.Workspace.Questions = append([]QuestionRecord{rec}, v.Workspace.Questions...)
	details := map[string]any{
		"question_id":             rec.ID,
		"receipt_id":              rec.ReceiptID,
		"context_id":              grounding.ContextID,
		"engine":                  "llama.cpp",
		"engine_version":          truncate(model.EngineVersion, maxOCRIdentityText),
		"model_name":              truncate(model.ModelName, maxOCRIdentityText),
		"model_sha256":            model.ModelSHA256,
		"grounded_claims":         len(report.Checks),
		"semantic_truth_verified": false,
		"model_draft_released":    false,
	}
	addResourceAuditDetails(details, model.Resources)
	v.addChangeUnlocked("local-ai", "grounded-local-ai-question", "Released a local llama.cpp answer after deterministic source grounding", details)
	if err := v.saveUnlocked(); err != nil {
		v.Workspace.Questions = oldQuestions
		v.Workspace.Changes = oldChanges
		v.Workspace.UpdatedAt = oldUpdatedAt
		v.Workspace.BuildID = oldBuildID
		return fmt.Errorf("persist grounded local AI question: %w", err)
	}
	return nil
}

func (v *Vault) recordLlamaGroundingRejection(question string, model LlamaCPPModelResult, grounding GroundingContext, reason string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	oldChanges := append([]ChangeRecord(nil), v.Workspace.Changes...)
	oldUpdatedAt := v.Workspace.UpdatedAt
	oldBuildID := v.Workspace.BuildID
	details := map[string]any{
		"question":                truncate(question, 300),
		"context_id":              grounding.ContextID,
		"engine":                  "llama.cpp",
		"engine_version":          truncate(model.EngineVersion, maxOCRIdentityText),
		"model_name":              truncate(model.ModelName, maxOCRIdentityText),
		"model_sha256":            model.ModelSHA256,
		"reason":                  reason,
		"model_draft_released":    false,
		"semantic_truth_verified": false,
	}
	addResourceAuditDetails(details, model.Resources)
	v.addChangeUnlocked("local-ai", "local-ai-grounding-rejected", "Rejected a local llama.cpp emission before answer release", details)
	if err := v.saveUnlocked(); err != nil {
		v.Workspace.Changes = oldChanges
		v.Workspace.UpdatedAt = oldUpdatedAt
		v.Workspace.BuildID = oldBuildID
		return err
	}
	return nil
}

func (v *Vault) recordLlamaResourceBlock(question string, model LlamaCPPModelResult, grounding GroundingContext) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	oldChanges := append([]ChangeRecord(nil), v.Workspace.Changes...)
	oldUpdatedAt := v.Workspace.UpdatedAt
	oldBuildID := v.Workspace.BuildID
	details := map[string]any{
		"question":      truncate(question, 300),
		"context_id":    grounding.ContextID,
		"engine":        "llama.cpp",
		"engine_version": truncate(model.EngineVersion, maxOCRIdentityText),
		"model_name":    truncate(model.ModelName, maxOCRIdentityText),
		"model_sha256":  model.ModelSHA256,
		"generation_started": false,
	}
	addResourceAuditDetails(details, model.Resources)
	v.addChangeUnlocked("local-ai", "local-ai-resource-blocked", "Blocked local llama.cpp launch because host resources were critically constrained", details)
	if err := v.saveUnlocked(); err != nil {
		v.Workspace.Changes = oldChanges
		v.Workspace.UpdatedAt = oldUpdatedAt
		v.Workspace.BuildID = oldBuildID
		return err
	}
	return nil
}
