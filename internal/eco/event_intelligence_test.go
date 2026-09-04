package eco

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAnalyzeEvidenceEventsUsesVerifiedReadingCopyAndHydratesProvenance(t *testing.T) {
	v, _ := testGroundingVault(t)
	ws := v.Snapshot()
	if len(ws.Evidence) != 1 {
		t.Fatalf("unexpected evidence count: %d", len(ws.Evidence))
	}
	evidence := ws.Evidence[0]
	seenPath := ""
	seenReceipt := SourceReceipt{}
	zone := time.FixedZone("UTC+2", 2*60*60)
	stamp := time.Date(2026, 8, 12, 14, 0, 0, 0, zone)
	dstPort := 443
	srcPort := 55000
	extractor := func(_ context.Context, path string, source SourceReceipt) ([]EventCandidate, error) {
		seenPath = path
		seenReceipt = source
		if _, err := os.ReadFile(path); err != nil {
			return nil, err
		}
		return []EventCandidate{{
			Sequence: 1, Source: "text-test", EventType: "hearing", Timestamp: &stamp,
			User: "Alice", SrcIP: "10.0.0.1", SrcPort: &srcPort, DstIP: "8.8.8.8", DstPort: &dstPort,
			SourceReference: "line=1", Message: "Hearing event",
		}}, nil
	}
	result, err := v.AnalyzeEvidenceEvents(evidence.ID, extractor, EventCorrelationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if seenPath == "" || seenPath == evidence.SourcePath {
		t.Fatalf("extractor did not receive a derived verified reading copy: %q", seenPath)
	}
	if seenReceipt.ObjectFile != evidence.ObjectFile || seenReceipt.SHA256 != evidence.SHA256 {
		t.Fatalf("extractor receipt not bound to preserved evidence: %+v", seenReceipt)
	}
	if len(result.Events) != 1 {
		t.Fatalf("unexpected event count: %d", len(result.Events))
	}
	event := result.Events[0]
	if event.EventID != "EVT-"+evidence.ID+"-000001" || event.EvidenceID != evidence.ID || event.SourceObject != evidence.ObjectFile || event.SourceSHA256 != evidence.SHA256 {
		t.Fatalf("event provenance was not hydrated by ECO: %+v", event)
	}
	if event.Timestamp == nil || event.Timestamp.Location() != time.UTC || event.Timestamp.Hour() != 12 {
		t.Fatalf("event timestamp was not normalized to UTC: %+v", event.Timestamp)
	}
	if len(result.Entities) == 0 || len(result.EntityEventLinks) == 0 {
		t.Fatal("event entity index was not built")
	}
	for _, entity := range result.Entities {
		if entity.EntityType == "port" && entity.Value == "55000" {
			t.Fatal("ephemeral source port was incorrectly indexed as an entity")
		}
	}
	foundDstPort := false
	foundConnection := false
	for _, entity := range result.Entities {
		if entity.EntityType == "port" && entity.Value == "443" {
			foundDstPort = true
		}
		if entity.EntityType == "network_connection" && entity.Value == "10.0.0.1->8.8.8.8" {
			foundConnection = true
		}
	}
	if !foundDstPort || !foundConnection {
		t.Fatalf("expected destination-port and network-connection entities, got %+v", result.Entities)
	}
}

func TestEventEntityIDsAreDeterministicAndCaseInsensitive(t *testing.T) {
	if EventEntityID("user", "Alice") != EventEntityID("user", " alice ") {
		t.Fatal("deterministic entity ID did not normalize case/whitespace")
	}
	if EventEntityID("user", "Alice") == EventEntityID("hostname", "Alice") {
		t.Fatal("entity type is not part of deterministic identity")
	}
}

func TestEventCorrelationPrioritizesSharedEntityAndLabelsTemporalOnlyWeakly(t *testing.T) {
	base := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)
	at := func(d time.Duration) *time.Time { value := base.Add(d); return &value }
	events := []NormalizedEvent{
		{EventID: "E1", EvidenceID: "A", SourceObject: "A.ecoobj", SourceSHA256: strings.Repeat("a", 64), Timestamp: at(0), User: "Alice"},
		{EventID: "E2", EvidenceID: "A", SourceObject: "A.ecoobj", SourceSHA256: strings.Repeat("a", 64), Timestamp: at(time.Minute), User: "alice"},
		{EventID: "E3", EvidenceID: "A", SourceObject: "A.ecoobj", SourceSHA256: strings.Repeat("a", 64), Timestamp: at(2 * time.Minute), User: "Bob"},
	}
	entities, links := BuildEventEntityIndex(events)
	correlations, stats, err := CorrelateNormalizedEvents(events, entities, links, EventCorrelationOptions{TimeWindow: 5 * time.Minute, MaxLinks: 10, MaxScannedPairs: 100, IncludePossible: true})
	if err != nil {
		t.Fatal(err)
	}
	if stats.ScannedPairs != 3 || len(correlations) != 3 {
		t.Fatalf("unexpected correlation count/stats: correlations=%+v stats=%+v", correlations, stats)
	}
	if correlations[0].RelationshipType != "related" || correlations[0].Confidence != "medium" || !strings.Contains(correlations[0].Basis, "does not establish causality") {
		t.Fatalf("strong correlation is not correctly hedged: %+v", correlations[0])
	}
	for _, link := range correlations[1:] {
		if link.RelationshipType != "possible_relationship" || link.Confidence != "low" || !strings.Contains(link.Basis, "not a confirmed relationship") {
			t.Fatalf("temporal-only relationship is overstated: %+v", link)
		}
	}
}

func TestSharedPortDoesNotCreateStrongCorrelation(t *testing.T) {
	base := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)
	second := base.Add(time.Second)
	port := 80
	events := []NormalizedEvent{
		{EventID: "E1", EvidenceID: "A", SourceObject: "A.ecoobj", SourceSHA256: strings.Repeat("a", 64), Timestamp: &base, DstPort: &port},
		{EventID: "E2", EvidenceID: "A", SourceObject: "A.ecoobj", SourceSHA256: strings.Repeat("a", 64), Timestamp: &second, DstPort: &port},
	}
	entities, links := BuildEventEntityIndex(events)
	correlations, _, err := CorrelateNormalizedEvents(events, entities, links, EventCorrelationOptions{IncludePossible: true, MaxLinks: 10, MaxScannedPairs: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(correlations) != 1 || correlations[0].RelationshipType != "possible_relationship" {
		t.Fatalf("shared destination port incorrectly became a strong relationship: %+v", correlations)
	}
}

func TestCorrelationCapsHighDegreeEntity(t *testing.T) {
	base := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)
	events := make([]NormalizedEvent, 0, 5)
	for i := 0; i < 5; i++ {
		stamp := base.Add(time.Duration(i) * time.Second)
		events = append(events, NormalizedEvent{EventID: string(rune('A' + i)), EvidenceID: "A", SourceObject: "A.ecoobj", SourceSHA256: strings.Repeat("a", 64), Timestamp: &stamp, User: "hub"})
	}
	entities, links := BuildEventEntityIndex(events)
	correlations, stats, err := CorrelateNormalizedEvents(events, entities, links, EventCorrelationOptions{MaxLinks: 10, MaxScannedPairs: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(correlations) != 1 || stats.CrowdedOutStrong == 0 {
		t.Fatalf("high-degree entity cap did not protect the link budget: correlations=%d stats=%+v", len(correlations), stats)
	}
}

func TestAnalyzeEvidenceEventsRejectsDanglingSourceSegment(t *testing.T) {
	v, _ := testGroundingVault(t)
	evidenceID := v.Snapshot().Evidence[0].ID
	extractor := func(_ context.Context, _ string, _ SourceReceipt) ([]EventCandidate, error) {
		return []EventCandidate{{Sequence: 1, Source: "test", EventType: "event", SourceSegmentID: "invented-segment"}}, nil
	}
	if _, err := v.AnalyzeEvidenceEvents(evidenceID, extractor, EventCorrelationOptions{}); err == nil || !strings.Contains(err.Error(), "source segment") {
		t.Fatalf("expected dangling source segment rejection, got %v", err)
	}
}

func TestEventCandidateRejectsInvalidPortsAndDuplicateSequence(t *testing.T) {
	v, _ := testGroundingVault(t)
	evidenceID := v.Snapshot().Evidence[0].ID
	badPort := 70000
	extractor := func(_ context.Context, _ string, _ SourceReceipt) ([]EventCandidate, error) {
		return []EventCandidate{{Sequence: 1, Source: "test", EventType: "event", DstPort: &badPort}}, nil
	}
	if _, err := v.AnalyzeEvidenceEvents(evidenceID, extractor, EventCorrelationOptions{}); err == nil || !strings.Contains(err.Error(), "port") {
		t.Fatalf("expected invalid port rejection, got %v", err)
	}
	extractor = func(_ context.Context, _ string, _ SourceReceipt) ([]EventCandidate, error) {
		return []EventCandidate{{Sequence: 1, Source: "test", EventType: "event"}, {Sequence: 1, Source: "test", EventType: "event"}}, nil
	}
	if _, err := v.AnalyzeEvidenceEvents(evidenceID, extractor, EventCorrelationOptions{}); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate sequence rejection, got %v", err)
	}
}
