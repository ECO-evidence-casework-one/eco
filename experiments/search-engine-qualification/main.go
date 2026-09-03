package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blugelabs/bluge"
	"github.com/deagy/recall/bm25"
)

const matterA = "MAT-A"

type document struct {
	ID           string
	MatterID     string
	Text         string
	SourceRegion string
}

type hit struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}

type searchEngine interface {
	Search(query string, limit int) ([]hit, error)
	Update(id, text string) error
	Close() error
}

type goldenQuery struct {
	Query    string `json:"query"`
	Expected string `json:"expected"`
}

type engineReport struct {
	Engine                 string              `json:"engine"`
	MandatoryPass          bool                `json:"mandatory_pass"`
	Error                  string              `json:"error,omitempty"`
	BuildMillis            float64             `json:"build_millis"`
	RetainedHeapBytes      uint64              `json:"retained_heap_bytes"`
	MeanSearchMicros       float64             `json:"mean_search_micros"`
	GoldenTop              map[string]string   `json:"golden_top"`
	GoldenExpected         map[string]string   `json:"golden_expected"`
	MatterIsolationPass    bool                `json:"matter_isolation_pass"`
	ReplacementPass        bool                `json:"replacement_pass"`
	ExternalProvenancePass bool                `json:"external_provenance_pass"`
	DeterminismPass        bool                `json:"determinism_pass"`
	RepeatedSearches       int                 `json:"repeated_searches"`
	Notes                  []string            `json:"notes,omitempty"`
}

type qualificationReport struct {
	Schema      string         `json:"schema"`
	Documents   int            `json:"documents_in_matter"`
	DecoyDocs   int            `json:"decoy_documents_excluded"`
	Iterations  int            `json:"search_iterations_per_query"`
	Reports     []engineReport `json:"reports"`
	Decision    string         `json:"decision"`
	GeneratedAt string         `json:"generated_at"`
}

var golden = []goldenQuery{
	{Query: "warranty confirmation", Expected: "gold-warranty"},
	{Query: "withheld medical records autism", Expected: "gold-medical"},
	{Query: "boiler payment invoice", Expected: "gold-invoice"},
	{Query: "housing local connection rural", Expected: "gold-housing"},
	{Query: "complaint chronology deadline", Expected: "gold-complaint"},
}

func main() {
	all := makeCorpus(5000, 500)
	docs := filterMatter(all, matterA)
	regions := make(map[string]string, len(docs))
	for _, d := range docs {
		regions[d.ID] = d.SourceRegion
	}

	builders := []struct {
		name  string
		build func([]document) (searchEngine, error)
	}{
		{name: "recall-v0.3.6", build: newRecallEngine},
		{name: "bleve-v2.6.1", build: newBleveEngine},
		{name: "bluge-v0.2.2", build: newBlugeEngine},
	}

	reports := make([]engineReport, 0, len(builders))
	for _, candidate := range builders {
		reports = append(reports, qualify(candidate.name, candidate.build, docs, regions))
	}

	decision := "HOLD"
	if len(reports) > 0 && reports[0].MandatoryPass {
		decision = "RETAIN_RECALL_BASELINE"
		if len(reports) > 1 && reports[1].MandatoryPass {
			decision = "RETAIN_RECALL_BASELINE;BLEVE_QUALIFIED_FOR_RICHER_FUTURE_SEARCH"
		}
	}

	out := qualificationReport{
		Schema:      "eco-search-qualification-v1",
		Documents:   len(docs),
		DecoyDocs:   len(all) - len(docs),
		Iterations:  100,
		Reports:     reports,
		Decision:    decision,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		panic(err)
	}

	for _, r := range reports {
		if r.Engine == "recall-v0.3.6" && !r.MandatoryPass {
			os.Exit(2)
		}
	}
}

func qualify(name string, build func([]document) (searchEngine, error), docs []document, regions map[string]string) engineReport {
	r := engineReport{
		Engine:           name,
		GoldenTop:        map[string]string{},
		GoldenExpected:   map[string]string{},
		RepeatedSearches: 20,
	}
	for _, q := range golden {
		r.GoldenExpected[q.Query] = q.Expected
	}

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	started := time.Now()
	eng, err := build(docs)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	defer eng.Close()
	r.BuildMillis = float64(time.Since(started).Microseconds()) / 1000
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	if after.Alloc > before.Alloc {
		r.RetainedHeapBytes = after.Alloc - before.Alloc
	}

	goldPass := true
	for _, q := range golden {
		hits, searchErr := eng.Search(q.Query, 50)
		if searchErr != nil {
			r.Error = searchErr.Error()
			return r
		}
		if len(hits) == 0 {
			goldPass = false
			r.GoldenTop[q.Query] = ""
			continue
		}
		r.GoldenTop[q.Query] = hits[0].ID
		if hits[0].ID != q.Expected {
			goldPass = false
		}
	}

	isolationPass := true
	for _, q := range golden {
		hits, searchErr := eng.Search(q.Query, 100)
		if searchErr != nil {
			r.Error = searchErr.Error()
			return r
		}
		for _, h := range hits {
			if strings.HasPrefix(h.ID, "decoy-") {
				isolationPass = false
			}
		}
	}
	r.MatterIsolationPass = isolationPass

	provHits, err := eng.Search("warranty confirmation", 10)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	if len(provHits) > 0 && regions[provHits[0].ID] == "Email.eml#body:1" {
		r.ExternalProvenancePass = true
	}

	deterministic := true
	baseline, err := eng.Search("warranty confirmation", 20)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	baseIDs := ids(baseline)
	for i := 0; i < r.RepeatedSearches; i++ {
		hits, searchErr := eng.Search("warranty confirmation", 20)
		if searchErr != nil {
			r.Error = searchErr.Error()
			return r
		}
		if strings.Join(ids(hits), "\x00") != strings.Join(baseIDs, "\x00") {
			deterministic = false
			break
		}
	}
	r.DeterminismPass = deterministic

	iterations := 100
	queries := make([]string, 0, iterations*len(golden))
	for i := 0; i < iterations; i++ {
		for _, q := range golden {
			queries = append(queries, q.Query)
		}
	}
	searchStart := time.Now()
	for _, query := range queries {
		if _, searchErr := eng.Search(query, 20); searchErr != nil {
			r.Error = searchErr.Error()
			return r
		}
	}
	r.MeanSearchMicros = float64(time.Since(searchStart).Microseconds()) / float64(len(queries))

	if err := eng.Update("gold-warranty", "resolved correspondence unrelated to guarantee terms"); err != nil {
		r.Error = err.Error()
		return r
	}
	postUpdate, err := eng.Search("warranty confirmation", 50)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.ReplacementPass = !containsID(postUpdate, "gold-warranty")

	r.MandatoryPass = goldPass && r.MatterIsolationPass && r.ReplacementPass && r.ExternalProvenancePass && r.DeterminismPass
	if name == "recall-v0.3.6" {
		r.Notes = append(r.Notes, "Pure BM25 package; no external dependencies inside bm25; matches current ECO lexical-ranking responsibility.")
	}
	if name == "bleve-v2.6.1" {
		r.Notes = append(r.Notes, "Mem-only mode; richer field/query/vector features than current ECO ranker; larger dependency surface requires justification.")
	}
	if name == "bluge-v0.2.2" {
		r.Notes = append(r.Notes, "Mem-only mode; technical candidate retained despite newest tagged release dating to 2022.")
	}
	return r
}

func makeCorpus(matterDocs, decoyDocs int) []document {
	seed := []document{
		{ID: "gold-warranty", MatterID: matterA, Text: "written warranty confirmation warranty confirmation repair commitment preserved email", SourceRegion: "Email.eml#body:1"},
		{ID: "weak-warranty", MatterID: matterA, Text: "repair appointment confirmation schedule", SourceRegion: "Appointment.eml#body:1"},
		{ID: "alt-warranty", MatterID: matterA, Text: "standard warranty details guarantee terms", SourceRegion: "Warranty.pdf#page:1"},
		{ID: "gold-medical", MatterID: matterA, Text: "withheld medical records autism diagnosis missing GP record disclosure", SourceRegion: "RecordsLetter.pdf#page:1"},
		{ID: "gold-invoice", MatterID: matterA, Text: "boiler repair payment invoice invoice receipt payment", SourceRegion: "Invoice.pdf#page:1"},
		{ID: "gold-housing", MatterID: matterA, Text: "housing local connection rural village preference one bedroom", SourceRegion: "HousingEmail.eml#body:1"},
		{ID: "gold-complaint", MatterID: matterA, Text: "complaint response chronology deadline chronology complaint deadline", SourceRegion: "Complaint.txt#line:1"},
	}
	out := append([]document(nil), seed...)
	vocab := []string{"letter", "meeting", "service", "repair", "claim", "response", "date", "casework", "document", "source", "review", "action", "task", "contact", "evidence", "timeline", "status", "note", "message", "record"}
	for i := len(out); i < matterDocs; i++ {
		parts := make([]string, 18)
		for j := range parts {
			parts[j] = vocab[(i*7+j*11)%len(vocab)]
		}
		out = append(out, document{
			ID:           fmt.Sprintf("noise-%05d", i),
			MatterID:     matterA,
			Text:         strings.Join(parts, " "),
			SourceRegion: fmt.Sprintf("Synthetic-%05d.txt#line:1", i),
		})
	}
	for i := 0; i < decoyDocs; i++ {
		text := "other matter generic correspondence service action"
		if i == 0 {
			text = strings.Repeat("warranty confirmation ", 20) + strings.Repeat("withheld medical records autism ", 10)
		}
		out = append(out, document{
			ID:           fmt.Sprintf("decoy-%05d", i),
			MatterID:     "MAT-B",
			Text:         text,
			SourceRegion: fmt.Sprintf("OtherMatter-%05d.txt#line:1", i),
		})
	}
	return out
}

func filterMatter(all []document, matterID string) []document {
	out := make([]document, 0, len(all))
	for _, d := range all {
		if d.MatterID == matterID {
			out = append(out, d)
		}
	}
	return out
}

func ids(hits []hit) []string {
	out := make([]string, len(hits))
	for i := range hits {
		out[i] = hits[i].ID
	}
	return out
}

func containsID(hits []hit, id string) bool {
	for _, h := range hits {
		if h.ID == id {
			return true
		}
	}
	return false
}

func normalizeHits(hits []hit, limit int) []hit {
	sort.SliceStable(hits, func(i, j int) bool {
		if math.Abs(hits[i].Score-hits[j].Score) < 1e-12 {
			return hits[i].ID < hits[j].ID
		}
		return hits[i].Score > hits[j].Score
	})
	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
	}
	return hits
}

// Recall

type recallEngine struct{ idx *bm25.BM25 }

func newRecallEngine(docs []document) (searchEngine, error) {
	idx := bm25.New(bm25.DefaultConfig())
	for _, d := range docs {
		idx.AddDocument(d.ID, d.Text)
	}
	return &recallEngine{idx: idx}, nil
}

func (e *recallEngine) Search(query string, limit int) ([]hit, error) {
	results := e.idx.Search(query)
	hits := make([]hit, 0, len(results))
	for _, r := range results {
		hits = append(hits, hit{ID: r.DocID, Score: r.Score})
	}
	return normalizeHits(hits, limit), nil
}
func (e *recallEngine) Update(id, text string) error { e.idx.AddDocument(id, text); return nil }
func (e *recallEngine) Close() error                  { return nil }

// Bleve

type bleveDoc struct {
	Content string `json:"content"`
}

type bleveEngine struct{ idx bleve.Index }

func newBleveEngine(docs []document) (searchEngine, error) {
	mapping := bleve.NewIndexMapping()
	docMapping := bleve.NewDocumentMapping()
	textMapping := bleve.NewTextFieldMapping()
	textMapping.Store = false
	textMapping.IncludeTermVectors = false
	docMapping.AddFieldMappingsAt("content", textMapping)
	mapping.DefaultMapping = docMapping
	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}
	eng := &bleveEngine{idx: idx}
	for _, d := range docs {
		if err := eng.Update(d.ID, d.Text); err != nil {
			idx.Close()
			return nil, err
		}
	}
	return eng, nil
}

func (e *bleveEngine) Search(query string, limit int) ([]hit, error) {
	q := bleve.NewMatchQuery(query)
	q.SetField("content")
	req := bleve.NewSearchRequest(q)
	if limit > 0 {
		req.Size = limit
	}
	res, err := e.idx.Search(req)
	if err != nil {
		return nil, err
	}
	hits := make([]hit, 0, len(res.Hits))
	for _, h := range res.Hits {
		hits = append(hits, hit{ID: h.ID, Score: h.Score})
	}
	return normalizeHits(hits, limit), nil
}
func (e *bleveEngine) Update(id, text string) error { return e.idx.Index(id, bleveDoc{Content: text}) }
func (e *bleveEngine) Close() error                  { return e.idx.Close() }

// Bluge

type blugeEngine struct {
	writer *bluge.Writer
	reader *bluge.Reader
}

func newBlugeEngine(docs []document) (searchEngine, error) {
	writer, err := bluge.OpenWriter(bluge.InMemoryOnlyConfig())
	if err != nil {
		return nil, err
	}
	eng := &blugeEngine{writer: writer}
	for _, d := range docs {
		if err := eng.updateNoRefresh(d.ID, d.Text); err != nil {
			writer.Close()
			return nil, err
		}
	}
	if err := eng.refresh(); err != nil {
		writer.Close()
		return nil, err
	}
	return eng, nil
}

func (e *blugeEngine) updateNoRefresh(id, text string) error {
	doc := bluge.NewDocument(id).AddField(bluge.NewTextField("content", text))
	return e.writer.Update(doc.ID(), doc)
}

func (e *blugeEngine) refresh() error {
	if e.reader != nil {
		_ = e.reader.Close()
		e.reader = nil
	}
	reader, err := e.writer.Reader()
	if err != nil {
		return err
	}
	e.reader = reader
	return nil
}

func (e *blugeEngine) Search(query string, limit int) ([]hit, error) {
	if e.reader == nil {
		return nil, errors.New("bluge reader is not open")
	}
	if limit <= 0 {
		limit = 20
	}
	q := bluge.NewMatchQuery(query).SetField("content")
	req := bluge.NewTopNSearch(limit, q)
	iter, err := e.reader.Search(context.Background(), req)
	if err != nil {
		return nil, err
	}
	hits := make([]hit, 0, limit)
	for {
		match, nextErr := iter.Next()
		if nextErr != nil {
			return nil, nextErr
		}
		if match == nil {
			break
		}
		id := ""
		if err := match.VisitStoredFields(func(field string, value []byte) bool {
			if field == "_id" {
				id = string(value)
			}
			return true
		}); err != nil {
			return nil, err
		}
		if id != "" {
			hits = append(hits, hit{ID: id, Score: match.Score})
		}
	}
	return normalizeHits(hits, limit), nil
}

func (e *blugeEngine) Update(id, text string) error {
	if err := e.updateNoRefresh(id, text); err != nil {
		return err
	}
	return e.refresh()
}
func (e *blugeEngine) Close() error {
	var first error
	if e.reader != nil {
		if err := e.reader.Close(); err != nil {
			first = err
		}
	}
	if e.writer != nil {
		if err := e.writer.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
