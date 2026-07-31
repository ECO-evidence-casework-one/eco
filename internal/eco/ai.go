package eco

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
)

var suspiciousInstruction = regexp.MustCompile(`(?i)(ignore (all|any|the|previous)|system (message|prompt)|developer message|follow these instructions|do not trust|execute|run (this|the) (code|command)|powershell|cmd\.exe|upload .* to|send .* to|exfiltrat|disable security|reveal (the )?prompt)`)
var datePattern = regexp.MustCompile(`(?i)\b(?:\d{1,2}[/-]\d{1,2}[/-]\d{2,4}|\d{1,2}\s+(?:jan(?:uary)?|feb(?:ruary)?|mar(?:ch)?|apr(?:il)?|may|jun(?:e)?|jul(?:y)?|aug(?:ust)?|sep(?:tember)?|oct(?:ober)?|nov(?:ember)?|dec(?:ember)?)\s+\d{2,4}|(?:jan(?:uary)?|feb(?:ruary)?|mar(?:ch)?|apr(?:il)?|may|jun(?:e)?|jul(?:y)?|aug(?:ust)?|sep(?:tember)?|oct(?:ober)?|nov(?:ember)?|dec(?:ember)?)\s+\d{1,2},?\s+\d{2,4})\b`)
var actionPattern = regexp.MustCompile(`(?i)\b(must|should|need to|required to|please|respond|reply|provide|send|submit|complete|contact|attend|pay|appeal|review|check|confirm)\b`)

type rankedSegment struct {
	Evidence EvidenceItem
	Segment  SourceSegment
	Score    float64
}

func (v *Vault) Ask(question string, scopeIDs []string) QuestionRecord {
	question = strings.TrimSpace(question)
	intent := classifyIntent(question)
	ws := v.Snapshot()
	ranked, excluded, lowConfidenceExcluded := rankSegments(question, ws.Evidence, scopeIDs)
	answer, citations, support := composeAnswer(intent, question, ranked, ws)
	rec := QuestionRecord{ID: NewID("Q"), AskedAt: time.Now().UTC(), Question: question, Intent: intent, Answer: answer, Citations: citations, Support: support, ScopeIDs: scopeIDs, ReceiptID: NewID("AIR"), EvidenceConsidered: len(ws.Evidence), RetrievedSegments: len(ranked), SuspiciousSourcesExcluded: excluded, LowConfidenceSourcesExcluded: lowConfidenceExcluded}
	v.mu.Lock()
	v.Workspace.Questions = append([]QuestionRecord{rec}, v.Workspace.Questions...)
	v.addChangeUnlocked("user", "question-asked", "Asked ECO a source-backed local question", map[string]any{"question": truncate(question, 300), "intent": intent, "sources": len(citations)})
	_ = v.saveUnlocked()
	v.mu.Unlock()
	return rec
}

func classifyIntent(q string) string {
	t := strings.ToLower(q)
	tokens := tokenize(t)
	scores := map[string]float64{"summary": 0.2, "dates": 0, "actions": 0, "compare": 0, "missing": 0, "status": 0, "explain": 0, "integrity": 0, "image": 0}
	weights := map[string]map[string]float64{
		"dates":     {"date": 3, "when": 2.5, "timeline": 3, "chronology": 3, "happened": 1.5, "deadline": 2.5},
		"actions":   {"action": 3, "next": 2, "do": 1.5, "must": 2.5, "should": 2, "respond": 2, "reply": 2, "required": 2},
		"compare":   {"compare": 4, "difference": 3, "different": 2.5, "conflict": 3, "contradict": 3, "versus": 3, "vs": 3},
		"missing":   {"missing": 4, "absent": 3, "gap": 3, "attachment": 2, "not included": 3, "what else": 2},
		"status":    {"status": 4, "where": 1.5, "progress": 3, "current": 2, "overview": 2, "where are we": 4},
		"explain":   {"explain": 4, "meaning": 3, "mean": 3, "plain": 2, "understand": 2},
		"integrity": {"hash": 4, "sha": 4, "integrity": 4, "changed": 2, "verified": 3},
		"image":     {"image": 3, "photo": 3, "photograph": 3, "scan": 3, "quality": 3, "blur": 3, "resolution": 3},
	}
	for intent, m := range weights {
		for _, tok := range tokens {
			scores[intent] += m[tok]
		}
	}
	for phrase, w := range map[string]float64{"where are we": 5, "what happened": 4, "what do i need to do": 5, "what is missing": 5, "image quality": 5, "next action": 5} {
		if strings.Contains(t, phrase) {
			switch phrase {
			case "where are we":
				scores["status"] += w
			case "what happened":
				scores["summary"] += w
			case "what do i need to do", "next action":
				scores["actions"] += w
			case "what is missing":
				scores["missing"] += w
			case "image quality":
				scores["image"] += w
			}
		}
	}
	best := "summary"
	bestScore := scores[best]
	for k, s := range scores {
		if s > bestScore {
			best, bestScore = k, s
		}
	}
	return best
}

func rankSegments(q string, evidence []EvidenceItem, scope []string) ([]rankedSegment, int, int) {
	allowed := map[string]bool{}
	for _, id := range scope {
		allowed[id] = true
	}
	useScope := len(scope) > 0
	query := tokenize(q)
	if len(query) == 0 {
		return nil, 0, 0
	}
	type doc struct {
		e      EvidenceItem
		s      SourceSegment
		tf     map[string]int
		length int
	}
	docs := []doc{}
	excluded := 0
	lowConfidenceExcluded := 0
	df := map[string]int{}
	for _, e := range evidence {
		if useScope && !allowed[e.ID] {
			continue
		}
		for _, s := range e.Segments {
			if s.Origin == "ocr" && s.Confidence > 0 && s.Confidence < 0.35 {
				lowConfidenceExcluded++
				continue
			}
			if suspiciousInstruction.MatchString(s.Text) {
				excluded++
				continue
			}
			toks := tokenize(s.Text)
			if len(toks) == 0 {
				continue
			}
			tf := map[string]int{}
			seen := map[string]bool{}
			for _, t := range toks {
				tf[t]++
				seen[t] = true
			}
			for t := range seen {
				df[t]++
			}
			docs = append(docs, doc{e, s, tf, len(toks)})
		}
	}
	if len(docs) == 0 {
		return nil, excluded, lowConfidenceExcluded
	}
	avg := 0.0
	for _, d := range docs {
		avg += float64(d.length)
	}
	avg /= float64(len(docs))
	k1, b := 1.5, 0.75
	out := make([]rankedSegment, 0, len(docs))
	for _, d := range docs {
		score := 0.0
		for _, t := range query {
			freq := float64(d.tf[t])
			if freq == 0 {
				continue
			}
			idf := math.Log(1 + (float64(len(docs))-float64(df[t])+0.5)/(float64(df[t])+0.5))
			score += idf * (freq * (k1 + 1)) / (freq + k1*(1-b+b*float64(d.length)/avg))
		}
		if score > 0 {
			out = append(out, rankedSegment{d.e, d.s, score})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if len(out) > 12 {
		out = out[:12]
	}
	return out, excluded, lowConfidenceExcluded
}

func composeAnswer(intent, q string, ranked []rankedSegment, ws Workspace) (string, []Citation, string) {
	if intent == "status" {
		return statusAnswer(ws), nil, "Workspace-derived"
	}
	if intent == "integrity" {
		return integrityAnswer(ws), nil, "Workspace-derived"
	}
	if intent == "image" {
		if a, c := imageAnswer(ws, ranked); a != "" {
			return a, c, "Directly supported by local image assessment"
		}
	}
	if len(ranked) == 0 {
		return "ECO could not find readable source text that supports an answer. Image files are preserved and quality-checked in this preview, but automatic OCR is not yet bundled. Try selecting a readable text, Office, OpenDocument, email or archive item.", nil, "Not sufficiently supported"
	}
	selected := diversify(ranked, 6)
	cites := make([]Citation, 0, len(selected))
	sentences := []string{}
	for _, r := range selected {
		candidate := bestSentence(r.Segment.Text, q, intent)
		if candidate == "" {
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
		cites = append(cites, Citation{EvidenceID: r.Evidence.ID, SegmentID: r.Segment.ID, Label: label, Quote: truncate(candidate, 500), Score: r.Score, Page: r.Segment.Page, Region: region, Origin: r.Segment.Origin, Confidence: r.Segment.Confidence})
		sentences = append(sentences, candidate+" ["+label+"]")
	}
	if len(sentences) == 0 {
		return "ECO found related evidence, but not enough clear wording to answer safely.", nil, "Not sufficiently supported"
	}
	prefix := "The strongest source-backed passages are:"
	switch intent {
	case "dates":
		prefix = "These date-related source passages are the strongest matches:"
	case "actions":
		prefix = "These possible action passages require your review:"
	case "compare":
		prefix = "These passages are the best comparison starting points; ECO has not decided which is correct:"
	case "missing":
		prefix = "These passages may indicate missing or referenced material:"
	case "explain":
		prefix = "In plain language, the source passages most relevant to your question say:"
	}
	support := "Directly supported by cited source passages"
	for _, c := range cites {
		if c.Origin == "ocr" {
			support = "Supported by coordinate-bearing OCR suggestions — check wording against the highlighted image regions"
			break
		}
	}
	return prefix + "\r\n\r\n• " + strings.Join(sentences, "\r\n\r\n• "), cites, support
}

func diversify(in []rankedSegment, maxN int) []rankedSegment {
	out := []rankedSegment{}
	per := map[string]int{}
	for _, r := range in {
		if per[r.Evidence.ID] >= 2 {
			continue
		}
		out = append(out, r)
		per[r.Evidence.ID]++
		if len(out) >= maxN {
			break
		}
	}
	return out
}

func bestSentence(text, q, intent string) string {
	parts := splitSentences(text)
	qt := tokenSet(q)
	best := ""
	score := -1.0
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if len(s) < 20 {
			continue
		}
		if intent == "actions" && !actionPattern.MatchString(s) {
			continue
		}
		if intent == "actions" && suspiciousInstruction.MatchString(s) {
			continue
		}
		if intent == "dates" && !datePattern.MatchString(s) {
			continue
		}
		st := tokenize(s)
		hit := 0.0
		for _, t := range st {
			if qt[t] {
				hit++
			}
		}
		hit /= math.Sqrt(float64(len(st)) + 1)
		if hit > score {
			score = hit
			best = s
		}
	}
	if best == "" && intent != "actions" && intent != "dates" {
		best = truncate(strings.TrimSpace(text), 650)
	}
	return truncate(best, 650)
}

func splitSentences(s string) []string {
	f := func(r rune) bool { return r == '.' || r == '!' || r == '?' || r == '\n' }
	return strings.FieldsFunc(s, f)
}
func tokenSet(s string) map[string]bool {
	m := map[string]bool{}
	for _, t := range tokenize(s) {
		m[t] = true
	}
	return m
}
func tokenize(s string) []string {
	var b strings.Builder
	out := []string{}
	flush := func() {
		if b.Len() > 1 {
			t := b.String()
			if !stopwords[t] {
				out = append(out, t)
			}
		}
		b.Reset()
	}
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}

var stopwords = map[string]bool{"the": true, "and": true, "for": true, "that": true, "this": true, "with": true, "from": true, "are": true, "was": true, "were": true, "have": true, "has": true, "had": true, "you": true, "your": true, "what": true, "when": true, "where": true, "which": true, "who": true, "how": true, "into": true, "about": true, "can": true, "could": true, "would": true, "should": true, "not": true, "but": true, "all": true, "any": true, "its": true, "our": true, "they": true, "them": true, "then": true, "than": true, "been": true, "being": true, "also": true, "only": true, "there": true, "here": true}

func statusAnswer(ws Workspace) string {
	readable, images, attention := 0, 0, 0
	for _, e := range ws.Evidence {
		if e.Readable {
			readable++
		}
		if e.Image != nil {
			images++
		}
		if len(e.Warnings) > 0 || e.Status == "Quarantined" {
			attention++
		}
	}
	openMatters := 0
	for _, m := range ws.Matters {
		if strings.EqualFold(m.Status, "active") {
			openMatters++
		}
	}
	return fmt.Sprintf("Current local workspace: %d preserved evidence items, %d readable items, %d assessed images, %d active matters and %d items needing attention. The newest build is %s. No cloud or network service is used.", len(ws.Evidence), readable, images, openMatters, attention, BuildID)
}
func integrityAnswer(ws Workspace) string {
	return fmt.Sprintf("ECO has %d encrypted evidence objects recorded. Use Trust & settings → Verify encrypted evidence to recalculate each decrypted SHA-256 and authentication tag.", len(ws.Evidence))
}
func imageAnswer(ws Workspace, ranked []rankedSegment) (string, []Citation) {
	for _, e := range ws.Evidence {
		if e.Image != nil {
			a := e.Image
			return fmt.Sprintf("%s is %d × %d pixels (%.1f MP), %s. Local assessment: brightness %.0f/255, contrast %.0f and blur variance %.0f. %s", e.SafeName, a.Width, a.Height, a.Megapixels, a.Orientation, a.Brightness, a.Contrast, a.BlurVariance, a.QualityLabel), nil
		}
	}
	return "", nil
}
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
