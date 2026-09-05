package eco

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	searchEngineVersion = "literal-folded-v1"
	maxSearchQueryRunes = 256
	maxSearchMatches    = 5000
	searchSnippetRadius = 120
)

var ErrSearchReceiptStale = errors.New("search receipt no longer matches the verified source reading")

// SearchSourceBinding identifies the preserved source bytes and the exact
// current readable-segment set used by one search receipt.
type SearchSourceBinding struct {
	EvidenceID    string `json:"evidence_id"`
	SafeName      string `json:"safe_name"`
	SourceObject  string `json:"source_object"`
	SourceSHA256  string `json:"source_sha256"`
	ReadingSHA256 string `json:"reading_sha256"`
}

// SearchMatch is one literal occurrence in one verified source segment.
// MatchStartRune/MatchEndRune are offsets within SegmentID's Text, not offsets
// within the original file bytes.
type SearchMatch struct {
	EvidenceID     string            `json:"evidence_id"`
	SafeName       string            `json:"safe_name"`
	SegmentID      string            `json:"segment_id"`
	SegmentOrdinal int               `json:"segment_ordinal"`
	MatchStartRune int               `json:"match_start_rune"`
	MatchEndRune   int               `json:"match_end_rune"`
	MatchText      string            `json:"match_text"`
	Snippet        string            `json:"snippet"`
	Page           int               `json:"page,omitempty"`
	PageHint       string            `json:"page_hint,omitempty"`
	Region         *NormalizedRegion `json:"region,omitempty"`
	Origin         string            `json:"origin,omitempty"`
	Confidence     float64           `json:"confidence,omitempty"`
	SourceObject   string            `json:"source_object"`
	SourceSHA256   string            `json:"source_sha256"`
}

// SearchReceipt is a reproducible description of one bounded deterministic
// search. WorkspaceRevision is audit context only; receipt validity is based on
// the exact source and reading digests so unrelated workspace writes do not
// invalidate otherwise identical search results.
type SearchReceipt struct {
	ID                string                `json:"id"`
	Engine            string                `json:"engine"`
	Query             string                `json:"query"`
	CreatedAt         time.Time             `json:"created_at"`
	WorkspaceRevision uint64                `json:"workspace_revision"`
	ScopeIDs          []string              `json:"scope_ids,omitempty"`
	Sources           []SearchSourceBinding `json:"sources"`
	Matches           []SearchMatch         `json:"matches"`
	Truncated         bool                  `json:"truncated"`
}

// SearchWorkspace performs a case-insensitive literal search over the current
// source-bound readable segments. It does not create a second index/database.
func (v *Vault) SearchWorkspace(query string, scopeIDs []string) (SearchReceipt, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return SearchReceipt{}, errors.New("search query is empty")
	}
	if len([]rune(query)) > maxSearchQueryRunes {
		return SearchReceipt{}, fmt.Errorf("search query exceeds %d characters", maxSearchQueryRunes)
	}

	v.opMu.RLock()
	defer v.opMu.RUnlock()
	v.mu.Lock()
	if err := v.ensureOpenUnlocked(); err != nil {
		v.mu.Unlock()
		return SearchReceipt{}, err
	}
	v.mu.Unlock()
	ws := v.Snapshot()

	allowed := make(map[string]bool, len(scopeIDs))
	for _, id := range scopeIDs {
		if id != "" {
			allowed[id] = true
		}
	}
	useScope := len(allowed) > 0
	needle := foldRunes(query)
	if len(needle) == 0 {
		return SearchReceipt{}, errors.New("search query contains no searchable characters")
	}

	receipt := SearchReceipt{
		ID:                NewID("SRCH"),
		Engine:            searchEngineVersion,
		Query:             query,
		CreatedAt:         time.Now().UTC(),
		WorkspaceRevision: ws.Revision,
		ScopeIDs:          append([]string(nil), scopeIDs...),
		Sources:           []SearchSourceBinding{},
		Matches:           []SearchMatch{},
	}

	for _, item := range ws.Evidence {
		if useScope && !allowed[item.ID] {
			continue
		}
		if !preservationUsable(item) {
			continue
		}
		segments := searchableSegments(item)
		if len(segments) == 0 {
			continue
		}
		readingDigest, err := searchReadingDigest(item, segments)
		if err != nil {
			return SearchReceipt{}, err
		}
		sourceAdded := false
		for _, seg := range segments {
			textRunes := []rune(seg.Text)
			folded := foldRunes(seg.Text)
			for _, start := range findFoldedRuneMatches(folded, needle) {
				end := start + len(needle)
				if end > len(textRunes) {
					continue
				}
				if !sourceAdded {
					receipt.Sources = append(receipt.Sources, SearchSourceBinding{EvidenceID: item.ID, SafeName: item.SafeName, SourceObject: item.ObjectFile, SourceSHA256: item.SHA256, ReadingSHA256: readingDigest})
					sourceAdded = true
				}
				var region *NormalizedRegion
				if seg.Region != nil {
					copyRegion := *seg.Region
					region = &copyRegion
				}
				receipt.Matches = append(receipt.Matches, SearchMatch{
					EvidenceID:     item.ID,
					SafeName:       item.SafeName,
					SegmentID:      seg.ID,
					SegmentOrdinal: seg.Ordinal,
					MatchStartRune: start,
					MatchEndRune:   end,
					MatchText:      string(textRunes[start:end]),
					Snippet:        searchSnippet(textRunes, start, end),
					Page:           seg.Page,
					PageHint:       seg.PageHint,
					Region:         region,
					Origin:         seg.Origin,
					Confidence:     seg.Confidence,
					SourceObject:   item.ObjectFile,
					SourceSHA256:   item.SHA256,
				})
				if len(receipt.Matches) >= maxSearchMatches {
					receipt.Truncated = true
					return receipt, nil
				}
			}
		}
	}
	return receipt, nil
}

// ValidateSearchReceipt rechecks both the current reading digest and the
// preserved bytes. A changed OCR/extraction reading or preserved object makes
// the old receipt unusable.
func (v *Vault) ValidateSearchReceipt(receipt SearchReceipt) error {
	if receipt.Engine != searchEngineVersion || receipt.ID == "" || strings.TrimSpace(receipt.Query) == "" {
		return errors.New("invalid search receipt")
	}
	v.opMu.RLock()
	defer v.opMu.RUnlock()
	v.mu.Lock()
	if err := v.ensureOpenUnlocked(); err != nil {
		v.mu.Unlock()
		return err
	}
	v.mu.Unlock()
	ws := v.Snapshot()

	byID := make(map[string]EvidenceItem, len(ws.Evidence))
	for _, item := range ws.Evidence {
		byID[item.ID] = item
	}
	for _, source := range receipt.Sources {
		item, ok := byID[source.EvidenceID]
		if !ok || !preservationUsable(item) || item.ObjectFile != source.SourceObject || item.SHA256 != source.SourceSHA256 {
			return fmt.Errorf("%w: source identity changed for %s", ErrSearchReceiptStale, source.EvidenceID)
		}
		segments := searchableSegments(item)
		digest, err := searchReadingDigest(item, segments)
		if err != nil {
			return err
		}
		if digest != source.ReadingSHA256 {
			return fmt.Errorf("%w: readable segments changed for %s", ErrSearchReceiptStale, source.EvidenceID)
		}
		if _, err := v.verifyPreservedObject(item.ID, item.ObjectFile, item.SHA256, item.Size); err != nil {
			return fmt.Errorf("%w: preserved source verification failed for %s: %v", ErrSearchReceiptStale, source.EvidenceID, err)
		}
	}
	return nil
}

func searchableSegments(item EvidenceItem) []SourceSegment {
	segments := make([]SourceSegment, 0, len(item.Segments))
	for _, seg := range item.Segments {
		if strings.TrimSpace(seg.Text) == "" || !segmentBoundToPreservedSource(item, seg) {
			continue
		}
		copySeg := seg
		if seg.Region != nil {
			region := *seg.Region
			copySeg.Region = &region
		}
		segments = append(segments, copySeg)
	}
	sort.SliceStable(segments, func(i, j int) bool {
		if segments[i].Page != segments[j].Page {
			return segments[i].Page < segments[j].Page
		}
		if segments[i].Ordinal != segments[j].Ordinal {
			return segments[i].Ordinal < segments[j].Ordinal
		}
		return segments[i].ID < segments[j].ID
	})
	return segments
}

func searchReadingDigest(item EvidenceItem, segments []SourceSegment) (string, error) {
	type digestSegment struct {
		ID           string            `json:"id"`
		Ordinal      int               `json:"ordinal"`
		Text         string            `json:"text"`
		Page         int               `json:"page"`
		PageHint     string            `json:"page_hint"`
		Region       *NormalizedRegion `json:"region,omitempty"`
		Origin       string            `json:"origin"`
		Confidence   float64           `json:"confidence"`
		SourceObject string            `json:"source_object"`
		SourceSHA256 string            `json:"source_sha256"`
	}
	normalized := make([]digestSegment, 0, len(segments))
	for _, seg := range segments {
		var region *NormalizedRegion
		if seg.Region != nil {
			copyRegion := *seg.Region
			region = &copyRegion
		}
		normalized = append(normalized, digestSegment{ID: seg.ID, Ordinal: seg.Ordinal, Text: seg.Text, Page: seg.Page, PageHint: seg.PageHint, Region: region, Origin: seg.Origin, Confidence: seg.Confidence, SourceObject: seg.SourceObject, SourceSHA256: seg.SourceSHA256})
	}
	payload, err := json.Marshal(struct {
		EvidenceID string          `json:"evidence_id"`
		SHA256     string          `json:"sha256"`
		Segments   []digestSegment `json:"segments"`
	}{EvidenceID: item.ID, SHA256: item.SHA256, Segments: normalized})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func foldRunes(s string) []rune {
	r := []rune(s)
	for i := range r {
		r[i] = unicode.ToLower(r[i])
	}
	return r
}

func findFoldedRuneMatches(haystack, needle []rune) []int {
	if len(needle) == 0 || len(needle) > len(haystack) {
		return nil
	}
	out := make([]int, 0, 4)
	for i := 0; i+len(needle) <= len(haystack); {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			out = append(out, i)
			i += len(needle)
			continue
		}
		i++
	}
	return out
}

func searchSnippet(text []rune, start, end int) string {
	left := start - searchSnippetRadius
	if left < 0 {
		left = 0
	}
	right := end + searchSnippetRadius
	if right > len(text) {
		right = len(text)
	}
	prefix := ""
	suffix := ""
	if left > 0 {
		prefix = "…"
	}
	if right < len(text) {
		suffix = "…"
	}
	return prefix + strings.TrimSpace(string(text[left:right])) + suffix
}
