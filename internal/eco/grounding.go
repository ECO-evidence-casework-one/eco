package eco

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	maxGroundingClaims     = 256
	maxGroundingClaimText  = 4096
	maxGroundingSourceText = 8192
	maxGroundingIDText     = 256
)

// GroundingRecord is the model-facing source vocabulary for one answer. It
// deliberately exposes opaque ECO IDs and the exact text shown to the model,
// but not trusted source hashes or internal object paths.
type GroundingRecord struct {
	EvidenceID string  `json:"evidence_id"`
	SegmentID  string  `json:"segment_id"`
	Display    string  `json:"display"`
	Text       string  `json:"text"`
	Page       int     `json:"page,omitempty"`
	Origin     string  `json:"origin,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// GroundingContext is created by ECO immediately before a local model call.
// trusted is intentionally not serialised: model output cannot recreate or
// authorise a context by copying JSON back to ECO.
type GroundingContext struct {
	ContextID                    string            `json:"context_id"`
	Question                     string            `json:"question"`
	Records                      []GroundingRecord `json:"records"`
	SuspiciousSourcesExcluded    int               `json:"suspicious_sources_excluded"`
	LowConfidenceSourcesExcluded int               `json:"low_confidence_sources_excluded"`
	SourceVerificationFailures   int               `json:"source_verification_failures"`
	trusted                      map[string]groundingTrustedSource
}

type groundingTrustedSource struct {
	EvidenceID   string
	SegmentID    string
	SourceObject string
	SourceSHA256 string
	ShownText    string
	Label        string
	Page         int
	Region       *NormalizedRegion
	Origin       string
	Confidence   float64
	Score        float64
}

// GroundingEmission is the narrow model-facing output contract. The model is
// allowed to point only at evidence/segment IDs that ECO placed in the current
// context. It cannot provide trusted hashes, bounding boxes or verification
// results.
type GroundingEmission struct {
	Answer string           `json:"answer"`
	Claims []GroundingClaim `json:"claims"`
}

type GroundingClaim struct {
	Kind       string `json:"kind"`
	Text       string `json:"text,omitempty"`
	EvidenceID string `json:"evidence_id"`
	SegmentID  string `json:"segment_id"`
}

type GroundingCheck struct {
	Index      int    `json:"index"`
	Kind       string `json:"kind"`
	EvidenceID string `json:"evidence_id"`
	SegmentID  string `json:"segment_id"`
	Status     string `json:"status"`
	Method     string `json:"method"`
	Reason     string `json:"reason,omitempty"`
}

type GroundingReport struct {
	ContextID             string           `json:"context_id"`
	AllClaimsGrounded     bool             `json:"all_claims_grounded"`
	SemanticTruthVerified bool             `json:"semantic_truth_verified"`
	Checks                []GroundingCheck `json:"checks"`
}

// BuildGroundingContext creates the exact bounded source vocabulary that may
// be shown to a local model. Suspicious document instructions and very-low-
// confidence OCR are removed by the normal ECO retrieval path first.
func (v *Vault) BuildGroundingContext(question string, scopeIDs []string) (GroundingContext, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return GroundingContext{}, errors.New("grounding question is required")
	}
	verificationFailures := v.verifyEvidenceForUse(scopeIDs)
	ws := v.Snapshot()
	ranked, excluded, lowConfidenceExcluded := rankSegments(question, ws.Evidence, scopeIDs)
	if len(ranked) == 0 {
		return GroundingContext{}, errors.New("no verified source segments support a grounding context")
	}

	ctx := GroundingContext{
		Question:                     question,
		SuspiciousSourcesExcluded:    excluded,
		LowConfidenceSourcesExcluded: lowConfidenceExcluded,
		SourceVerificationFailures:   verificationFailures,
		trusted:                      make(map[string]groundingTrustedSource),
	}
	for _, r := range ranked {
		if !segmentBoundToPreservedSource(r.Evidence, r.Segment) {
			continue
		}
		shown := boundGroundingText(r.Segment.Text, maxGroundingSourceText)
		if strings.TrimSpace(shown) == "" {
			continue
		}
		key := groundingKey(r.Evidence.ID, r.Segment.ID)
		if _, exists := ctx.trusted[key]; exists {
			continue
		}
		label := fmt.Sprintf("%s · §%d", r.Evidence.SafeName, r.Segment.Ordinal)
		if r.Segment.Origin == "ocr" {
			label = fmt.Sprintf("%s · OCR page %d · %.0f%% confidence", r.Evidence.SafeName, max(1, r.Segment.Page), r.Segment.Confidence*100)
		}
		var region *NormalizedRegion
		if r.Segment.Region != nil {
			copyRegion := *r.Segment.Region
			region = &copyRegion
		}
		record := GroundingRecord{
			EvidenceID: r.Evidence.ID,
			SegmentID:  r.Segment.ID,
			Display:    label,
			Text:       shown,
			Page:       r.Segment.Page,
			Origin:     r.Segment.Origin,
			Confidence: r.Segment.Confidence,
		}
		ctx.Records = append(ctx.Records, record)
		ctx.trusted[key] = groundingTrustedSource{
			EvidenceID:   r.Evidence.ID,
			SegmentID:    r.Segment.ID,
			SourceObject: r.Evidence.ObjectFile,
			SourceSHA256: r.Evidence.SHA256,
			ShownText:    shown,
			Label:        label,
			Page:         r.Segment.Page,
			Region:       region,
			Origin:       r.Segment.Origin,
			Confidence:   r.Segment.Confidence,
			Score:        r.Score,
		}
	}
	if len(ctx.Records) == 0 {
		return GroundingContext{}, errors.New("no verified source text remained after grounding controls")
	}
	ctx.ContextID = groundingContextID(ctx.trusted)
	return ctx, nil
}

// VerifyGroundingEmission deterministically resolves every model claim only
// through the app-owned context created above. Unknown IDs and malformed
// claims fail the entire batch. Text mismatches return a negative report and
// no releasable citations; callers must not regenerate merely to obtain green.
func (v *Vault) VerifyGroundingEmission(ctx GroundingContext, emission GroundingEmission) (GroundingReport, []Citation, error) {
	report := GroundingReport{ContextID: ctx.ContextID, SemanticTruthVerified: false}
	if ctx.ContextID == "" || len(ctx.trusted) == 0 || groundingContextID(ctx.trusted) != ctx.ContextID {
		return report, nil, errors.New("grounding context is missing, reconstructed or mutated")
	}
	if err := validateGroundingContextRecords(ctx); err != nil {
		return report, nil, err
	}
	if strings.TrimSpace(emission.Answer) == "" {
		return report, nil, errors.New("grounding answer is required")
	}
	if len(emission.Claims) == 0 || len(emission.Claims) > maxGroundingClaims {
		return report, nil, fmt.Errorf("grounding claims must contain 1 to %d items", maxGroundingClaims)
	}

	// Re-bind every trusted context entry to the current verified workspace so a
	// stale or replaced derived segment cannot be accepted after retrieval.
	ws := v.Snapshot()
	current := make(map[string]groundingTrustedSource, len(ctx.trusted))
	for key, trusted := range ctx.trusted {
		item, segment, ok := findGroundingSource(ws, trusted.EvidenceID, trusted.SegmentID)
		if !ok || !preservationUsable(item) || !segmentBoundToPreservedSource(item, segment) {
			return report, nil, fmt.Errorf("grounding source %s/%s is no longer verified", trusted.EvidenceID, trusted.SegmentID)
		}
		if item.ObjectFile != trusted.SourceObject || item.SHA256 != trusted.SourceSHA256 || boundGroundingText(segment.Text, maxGroundingSourceText) != trusted.ShownText {
			return report, nil, fmt.Errorf("grounding source %s/%s changed after retrieval", trusted.EvidenceID, trusted.SegmentID)
		}
		current[key] = trusted
	}

	verifiedEvidence := map[string]bool{}
	citations := make([]Citation, 0, len(emission.Claims))
	allGrounded := true
	for i, claim := range emission.Claims {
		index := i + 1
		if err := validateGroundingClaim(claim, index); err != nil {
			return report, nil, err
		}
		key := groundingKey(claim.EvidenceID, claim.SegmentID)
		trusted, ok := current[key]
		if !ok {
			return report, nil, fmt.Errorf("claim %d: out_of_vocabulary: evidence/segment ID was not shown", index)
		}
		if !verifiedEvidence[trusted.EvidenceID] {
			item, _, _ := findGroundingSource(ws, trusted.EvidenceID, trusted.SegmentID)
			if _, err := v.verifyPreservedObject(item.ID, item.ObjectFile, item.SHA256, item.Size); err != nil {
				v.markEvidenceVerificationFailure(item.ID, err)
				return report, nil, fmt.Errorf("claim %d: preserved source verification failed: %w", index, err)
			}
			verifiedEvidence[trusted.EvidenceID] = true
		}

		check := GroundingCheck{Index: index, Kind: claim.Kind, EvidenceID: claim.EvidenceID, SegmentID: claim.SegmentID}
		grounded := false
		switch claim.Kind {
		case "presence":
			grounded = true
			check.Status = "grounded"
			check.Method = "presence_only"
		case "quote", "value":
			if strings.Contains(trusted.ShownText, claim.Text) {
				grounded = true
				check.Status = "grounded"
				check.Method = "exact_text_contains"
			} else if strings.Contains(normalizeGroundingText(trusted.ShownText), normalizeGroundingText(claim.Text)) {
				grounded = true
				check.Status = "grounded"
				check.Method = "normalized_text_contains"
			} else {
				check.Status = "mismatch"
				check.Method = "none"
				check.Reason = "text_mismatch"
			}
		}
		report.Checks = append(report.Checks, check)
		if !grounded {
			allGrounded = false
			continue
		}
		quote := claim.Text
		if claim.Kind == "presence" {
			quote = boundGroundingText(trusted.ShownText, 500)
		}
		var region *NormalizedRegion
		if trusted.Region != nil {
			copyRegion := *trusted.Region
			region = &copyRegion
		}
		citations = append(citations, Citation{
			EvidenceID:   trusted.EvidenceID,
			SegmentID:    trusted.SegmentID,
			Label:        trusted.Label,
			Quote:        boundGroundingText(quote, 500),
			Score:        trusted.Score,
			Page:         trusted.Page,
			Region:       region,
			Origin:       trusted.Origin,
			Confidence:   trusted.Confidence,
			SourceObject: trusted.SourceObject,
			SourceSHA256: trusted.SourceSHA256,
		})
	}
	report.AllClaimsGrounded = allGrounded && len(report.Checks) == len(emission.Claims)
	if !report.AllClaimsGrounded {
		return report, nil, nil
	}
	return report, citations, nil
}

func validateGroundingContextRecords(ctx GroundingContext) error {
	if len(ctx.Records) != len(ctx.trusted) {
		return errors.New("grounding context records no longer match the trusted source set")
	}
	seen := make(map[string]bool, len(ctx.Records))
	for _, record := range ctx.Records {
		key := groundingKey(record.EvidenceID, record.SegmentID)
		if seen[key] {
			return errors.New("grounding context contains duplicate shown source IDs")
		}
		seen[key] = true
		trusted, ok := ctx.trusted[key]
		if !ok || record.Display != trusted.Label || record.Text != trusted.ShownText || record.Page != trusted.Page || record.Origin != trusted.Origin || record.Confidence != trusted.Confidence {
			return errors.New("grounding context shown records were mutated after retrieval")
		}
	}
	return nil
}

func validateGroundingClaim(claim GroundingClaim, index int) error {
	if strings.TrimSpace(claim.EvidenceID) == "" || strings.TrimSpace(claim.SegmentID) == "" || len([]rune(claim.EvidenceID)) > maxGroundingIDText || len([]rune(claim.SegmentID)) > maxGroundingIDText {
		return fmt.Errorf("claim %d: invalid_locator: bounded evidence_id and segment_id are required", index)
	}
	switch claim.Kind {
	case "presence":
		if strings.TrimSpace(claim.Text) != "" {
			return fmt.Errorf("claim %d: invalid_claim: presence claims must not contain text", index)
		}
	case "quote", "value":
		if strings.TrimSpace(claim.Text) == "" || len([]rune(claim.Text)) > maxGroundingClaimText {
			return fmt.Errorf("claim %d: invalid_claim: quote/value text is missing or unbounded", index)
		}
	default:
		return fmt.Errorf("claim %d: unsupported_claim_kind: %q", index, claim.Kind)
	}
	return nil
}

func findGroundingSource(ws Workspace, evidenceID, segmentID string) (EvidenceItem, SourceSegment, bool) {
	for _, item := range ws.Evidence {
		if item.ID != evidenceID {
			continue
		}
		for _, segment := range item.Segments {
			if segment.ID == segmentID {
				return item, segment, true
			}
		}
		return EvidenceItem{}, SourceSegment{}, false
	}
	return EvidenceItem{}, SourceSegment{}, false
}

func groundingKey(evidenceID, segmentID string) string {
	return evidenceID + "\x00" + segmentID
}

func groundingContextID(trusted map[string]groundingTrustedSource) string {
	keys := make([]string, 0, len(trusted))
	for key := range trusted {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, key := range keys {
		s := trusted[key]
		_, _ = h.Write([]byte(s.EvidenceID))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(s.SegmentID))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(s.SourceObject))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(s.SourceSHA256))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(s.ShownText))
		_, _ = h.Write([]byte{0xff})
	}
	return "GCTX-" + hex.EncodeToString(h.Sum(nil))
}

func normalizeGroundingText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func boundGroundingText(s string, limit int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return strings.TrimSpace(string(runes[:limit]))
}
