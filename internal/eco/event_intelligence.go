package eco

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	maxEvidenceEventBatch       = 250_000
	maxEventIdentityText        = 256
	maxEventFieldText           = 4_096
	maxEventLargeFieldText      = 32_768
	defaultCorrelationMaxLinks  = 50_000
	defaultCorrelationMaxScans  = 2_000_000
	maxCorrelationLinksHard     = 250_000
	maxCorrelationScansHard     = 10_000_000
)

const maxLinksPerEntityShare = 0.10

// EvidenceEventExtractor receives only a freshly verified plaintext reading
// copy created by ECO. It must derive EventCandidate values from that file and
// must not retain or modify the temporary path.
type EvidenceEventExtractor func(context.Context, string, SourceReceipt) ([]EventCandidate, error)

type EventCandidate struct {
	Sequence           int        `json:"sequence"`
	Source             string     `json:"source"`
	EventType          string     `json:"event_type"`
	SourceSegmentID    string     `json:"source_segment_id,omitempty"`
	SourceReference    string     `json:"source_reference,omitempty"`
	Timestamp          *time.Time `json:"timestamp,omitempty"`
	TimestampPrecision string     `json:"timestamp_precision,omitempty"`
	User               string     `json:"user,omitempty"`
	Hostname           string     `json:"hostname,omitempty"`
	Device             string     `json:"device,omitempty"`
	SrcIP              string     `json:"src_ip,omitempty"`
	SrcPort            *int       `json:"src_port,omitempty"`
	DstIP              string     `json:"dst_ip,omitempty"`
	DstPort            *int       `json:"dst_port,omitempty"`
	Protocol           string     `json:"protocol,omitempty"`
	ProcessName        string     `json:"process_name,omitempty"`
	ProcessID          *int       `json:"process_id,omitempty"`
	ParentProcess      string     `json:"parent_process,omitempty"`
	CommandLine        string     `json:"command_line,omitempty"`
	FileName           string     `json:"file_name,omitempty"`
	FilePath           string     `json:"file_path,omitempty"`
	FileHash           string     `json:"file_hash,omitempty"`
	Domain             string     `json:"domain,omitempty"`
	URL                string     `json:"url,omitempty"`
	Severity           string     `json:"severity,omitempty"`
	Message            string     `json:"message,omitempty"`
}

type NormalizedEvent struct {
	EventID            string     `json:"event_id"`
	EvidenceID         string     `json:"evidence_id"`
	SourceObject       string     `json:"source_object"`
	SourceSHA256       string     `json:"source_sha256"`
	Source             string     `json:"source"`
	EventType          string     `json:"event_type"`
	SourceSegmentID    string     `json:"source_segment_id,omitempty"`
	SourceReference    string     `json:"source_reference,omitempty"`
	Timestamp          *time.Time `json:"timestamp,omitempty"`
	TimestampPrecision string     `json:"timestamp_precision,omitempty"`
	User               string     `json:"user,omitempty"`
	Hostname           string     `json:"hostname,omitempty"`
	Device             string     `json:"device,omitempty"`
	SrcIP              string     `json:"src_ip,omitempty"`
	SrcPort            *int       `json:"src_port,omitempty"`
	DstIP              string     `json:"dst_ip,omitempty"`
	DstPort            *int       `json:"dst_port,omitempty"`
	Protocol           string     `json:"protocol,omitempty"`
	ProcessName        string     `json:"process_name,omitempty"`
	ProcessID          *int       `json:"process_id,omitempty"`
	ParentProcess      string     `json:"parent_process,omitempty"`
	CommandLine        string     `json:"command_line,omitempty"`
	FileName           string     `json:"file_name,omitempty"`
	FilePath           string     `json:"file_path,omitempty"`
	FileHash           string     `json:"file_hash,omitempty"`
	Domain             string     `json:"domain,omitempty"`
	URL                string     `json:"url,omitempty"`
	Severity           string     `json:"severity,omitempty"`
	Message            string     `json:"message,omitempty"`
}

type EventEntity struct {
	EntityID   string `json:"entity_id"`
	EntityType string `json:"entity_type"`
	Value      string `json:"value"`
}

type EntityEventLink struct {
	EntityID string `json:"entity_id"`
	EventID  string `json:"event_id"`
	Field    string `json:"field"`
}

type EventCorrelation struct {
	LinkID             string  `json:"link_id"`
	EventIDA           string  `json:"event_id_a"`
	EventIDB           string  `json:"event_id_b"`
	RelationshipType   string  `json:"relationship_type"`
	Basis              string  `json:"basis"`
	SharedEntityID     string  `json:"shared_entity_id,omitempty"`
	SharedEntityType   string  `json:"shared_entity_type,omitempty"`
	SharedEntityValue  string  `json:"shared_entity_value,omitempty"`
	TimeDeltaSeconds   float64 `json:"time_delta_seconds"`
	Confidence         string  `json:"confidence"`
}

type EventCorrelationOptions struct {
	TimeWindow      time.Duration `json:"time_window"`
	MaxLinks        int           `json:"max_links"`
	MaxScannedPairs int           `json:"max_scanned_pairs"`
	IncludePossible bool          `json:"include_possible"`
}

type EventIntelligenceResult struct {
	EvidenceID          string             `json:"evidence_id"`
	SourceObject        string             `json:"source_object"`
	SourceSHA256        string             `json:"source_sha256"`
	Events              []NormalizedEvent  `json:"events"`
	Entities            []EventEntity      `json:"entities"`
	EntityEventLinks    []EntityEventLink  `json:"entity_event_links"`
	Correlations        []EventCorrelation `json:"correlations"`
	ScannedPairs        int                `json:"scanned_pairs"`
	SuppressedPossible  int                `json:"suppressed_possible"`
	CrowdedOutStrong    int                `json:"crowded_out_strong"`
	ScanLimitReached    bool               `json:"scan_limit_reached"`
}

// AnalyzeEvidenceEvents gives a parser a verified ECO reading copy, normalizes
// its candidate events, builds deterministic entities and produces cautious
// time/entity correlations. Results are returned in memory in this slice: ECO
// deliberately does not write a large plaintext analytical database.
func (v *Vault) AnalyzeEvidenceEvents(evidenceID string, extractor EvidenceEventExtractor, options EventCorrelationOptions) (EventIntelligenceResult, error) {
	return v.AnalyzeEvidenceEventsContext(context.Background(), evidenceID, extractor, options)
}

func (v *Vault) AnalyzeEvidenceEventsContext(ctx context.Context, evidenceID string, extractor EvidenceEventExtractor, options EventCorrelationOptions) (EventIntelligenceResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(evidenceID) == "" {
		return EventIntelligenceResult{}, errors.New("event analysis evidence ID is required")
	}
	if extractor == nil {
		return EventIntelligenceResult{}, errors.New("event extractor is required")
	}
	item, record, err := v.eventIntelligenceSource(evidenceID)
	if err != nil {
		return EventIntelligenceResult{}, err
	}
	if item.DetectedType == "executable" || item.Status == "Quarantined" {
		return EventIntelligenceResult{}, errors.New("event extraction is blocked for quarantined or executable evidence")
	}

	var candidates []EventCandidate
	var source SourceReceipt
	err = v.withVerifiedPreservedFile(record, item.SHA256, func(path string, receipt SourceReceipt) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		got, extractErr := extractor(ctx, path, receipt)
		if extractErr != nil {
			return extractErr
		}
		if len(got) > maxEvidenceEventBatch {
			return fmt.Errorf("event extractor returned %d records; the safe per-batch limit is %d", len(got), maxEvidenceEventBatch)
		}
		candidates = got
		source = receipt
		return nil
	})
	if err != nil {
		return EventIntelligenceResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return EventIntelligenceResult{}, err
	}
	events, err := normalizeEventCandidates(item, source, candidates)
	if err != nil {
		return EventIntelligenceResult{}, err
	}
	entities, links := BuildEventEntityIndex(events)
	correlations, stats, err := CorrelateNormalizedEvents(events, entities, links, options)
	if err != nil {
		return EventIntelligenceResult{}, err
	}
	return EventIntelligenceResult{
		EvidenceID:         item.ID,
		SourceObject:       source.ObjectFile,
		SourceSHA256:       source.SHA256,
		Events:             events,
		Entities:           entities,
		EntityEventLinks:   links,
		Correlations:       correlations,
		ScannedPairs:       stats.ScannedPairs,
		SuppressedPossible: stats.SuppressedPossible,
		CrowdedOutStrong:   stats.CrowdedOutStrong,
		ScanLimitReached:   stats.ScanLimitReached,
	}, nil
}

func (v *Vault) eventIntelligenceSource(evidenceID string) (EvidenceItem, PreservationRecord, error) {
	ws := v.Snapshot()
	var item EvidenceItem
	found := false
	for _, candidate := range ws.Evidence {
		if candidate.ID == evidenceID {
			item = cloneEvidenceItem(candidate)
			found = true
			break
		}
	}
	if !found {
		return EvidenceItem{}, PreservationRecord{}, os.ErrNotExist
	}
	if !preservationUsable(item) {
		return EvidenceItem{}, PreservationRecord{}, errors.New("event analysis is blocked because the preserved source is not verified")
	}
	for _, candidate := range ws.Preservations {
		if candidate.EvidenceID == evidenceID && candidate.State == preservationCommitted && candidate.ObjectFile == item.ObjectFile && candidate.PreservedSHA256 == item.SHA256 && candidate.ExpectedSize == item.Size {
			return item, candidate, nil
		}
	}
	return EvidenceItem{}, PreservationRecord{}, errors.New("event analysis is blocked because the committed preservation record is missing or inconsistent")
}

func normalizeEventCandidates(item EvidenceItem, source SourceReceipt, candidates []EventCandidate) ([]NormalizedEvent, error) {
	if source.ObjectFile != item.ObjectFile || source.SHA256 != item.SHA256 || source.VerifiedAt.IsZero() {
		return nil, errors.New("event candidates are not bound to the verified preserved evidence")
	}
	segments := make(map[string]bool, len(item.Segments))
	for _, segment := range item.Segments {
		if segmentBoundToPreservedSource(item, segment) {
			segments[segment.ID] = true
		}
	}
	seenSequence := make(map[int]bool, len(candidates))
	out := make([]NormalizedEvent, 0, len(candidates))
	for i, candidate := range candidates {
		if candidate.Sequence < 1 || candidate.Sequence > 999_999_999 || seenSequence[candidate.Sequence] {
			return nil, fmt.Errorf("event candidate %d has a missing, duplicate or invalid sequence", i+1)
		}
		seenSequence[candidate.Sequence] = true
		sourceName, err := boundedEventRequired(candidate.Source, "source")
		if err != nil {
			return nil, fmt.Errorf("event candidate %d: %w", i+1, err)
		}
		eventType, err := boundedEventRequired(candidate.EventType, "event type")
		if err != nil {
			return nil, fmt.Errorf("event candidate %d: %w", i+1, err)
		}
		if candidate.SourceSegmentID != "" && !segments[candidate.SourceSegmentID] {
			return nil, fmt.Errorf("event candidate %d references a source segment that is not bound to the preserved evidence", i+1)
		}
		if err := validateEventCandidateFields(candidate); err != nil {
			return nil, fmt.Errorf("event candidate %d: %w", i+1, err)
		}
		var timestamp *time.Time
		if candidate.Timestamp != nil {
			t := candidate.Timestamp.UTC()
			timestamp = &t
		}
		out = append(out, NormalizedEvent{
			EventID:            fmt.Sprintf("EVT-%s-%06d", item.ID, candidate.Sequence),
			EvidenceID:         item.ID,
			SourceObject:       item.ObjectFile,
			SourceSHA256:       item.SHA256,
			Source:             sourceName,
			EventType:          eventType,
			SourceSegmentID:    strings.TrimSpace(candidate.SourceSegmentID),
			SourceReference:    strings.TrimSpace(candidate.SourceReference),
			Timestamp:          timestamp,
			TimestampPrecision: strings.TrimSpace(candidate.TimestampPrecision),
			User:               strings.TrimSpace(candidate.User),
			Hostname:           strings.TrimSpace(candidate.Hostname),
			Device:             strings.TrimSpace(candidate.Device),
			SrcIP:              strings.TrimSpace(candidate.SrcIP),
			SrcPort:            cloneIntPointer(candidate.SrcPort),
			DstIP:              strings.TrimSpace(candidate.DstIP),
			DstPort:            cloneIntPointer(candidate.DstPort),
			Protocol:           strings.TrimSpace(candidate.Protocol),
			ProcessName:        strings.TrimSpace(candidate.ProcessName),
			ProcessID:          cloneIntPointer(candidate.ProcessID),
			ParentProcess:      strings.TrimSpace(candidate.ParentProcess),
			CommandLine:        strings.TrimSpace(candidate.CommandLine),
			FileName:           strings.TrimSpace(candidate.FileName),
			FilePath:           strings.TrimSpace(candidate.FilePath),
			FileHash:           strings.TrimSpace(candidate.FileHash),
			Domain:             strings.TrimSpace(candidate.Domain),
			URL:                strings.TrimSpace(candidate.URL),
			Severity:           strings.TrimSpace(candidate.Severity),
			Message:            strings.TrimSpace(candidate.Message),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].EventID < out[j].EventID })
	return out, nil
}

func boundedEventRequired(value, label string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || len([]rune(value)) > maxEventIdentityText {
		return "", fmt.Errorf("%s is missing or unbounded", label)
	}
	return value, nil
}

func validateEventCandidateFields(c EventCandidate) error {
	identityFields := []string{c.SourceSegmentID, c.TimestampPrecision, c.Severity, c.Protocol}
	for _, value := range identityFields {
		if len([]rune(value)) > maxEventIdentityText {
			return errors.New("event identity metadata is unbounded")
		}
	}
	fields := []string{c.SourceReference, c.User, c.Hostname, c.Device, c.SrcIP, c.DstIP, c.ProcessName, c.ParentProcess, c.FileName, c.FilePath, c.FileHash, c.Domain}
	for _, value := range fields {
		if len([]rune(value)) > maxEventFieldText {
			return errors.New("event field text is unbounded")
		}
	}
	for _, value := range []string{c.CommandLine, c.URL, c.Message} {
		if len([]rune(value)) > maxEventLargeFieldText {
			return errors.New("event large field text is unbounded")
		}
	}
	for _, port := range []*int{c.SrcPort, c.DstPort} {
		if port != nil && (*port < 0 || *port > 65535) {
			return errors.New("event port is outside 0..65535")
		}
	}
	if c.ProcessID != nil && *c.ProcessID < 0 {
		return errors.New("event process ID cannot be negative")
	}
	return nil
}

func cloneIntPointer(in *int) *int {
	if in == nil {
		return nil
	}
	value := *in
	return &value
}

var eventEntityFields = []struct {
	Field      string
	EntityType string
	Value      func(NormalizedEvent) string
}{
	{"user", "user", func(e NormalizedEvent) string { return e.User }},
	{"hostname", "hostname", func(e NormalizedEvent) string { return e.Hostname }},
	{"device", "device", func(e NormalizedEvent) string { return e.Device }},
	{"src_ip", "ip_address", func(e NormalizedEvent) string { return e.SrcIP }},
	{"dst_ip", "ip_address", func(e NormalizedEvent) string { return e.DstIP }},
	{"domain", "domain", func(e NormalizedEvent) string { return e.Domain }},
	{"url", "url", func(e NormalizedEvent) string { return e.URL }},
	{"file_name", "file", func(e NormalizedEvent) string { return e.FileName }},
	{"file_hash", "hash", func(e NormalizedEvent) string { return e.FileHash }},
	{"process_name", "process", func(e NormalizedEvent) string { return e.ProcessName }},
	{"dst_port", "port", func(e NormalizedEvent) string {
		if e.DstPort == nil { return "" }
		return fmt.Sprintf("%d", *e.DstPort)
	}},
}

func EventEntityID(entityType, value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	h := sha256.Sum256([]byte(entityType + ":" + normalized))
	return "ENT-" + entityType + "-" + hex.EncodeToString(h[:])[:12]
}

// BuildEventEntityIndex follows the donor's signal/noise distinction: source
// ports are not entities, destination ports are indexable but not useful for
// correlation, and a directional host pair becomes a network_connection.
func BuildEventEntityIndex(events []NormalizedEvent) ([]EventEntity, []EntityEventLink) {
	entityMap := map[string]EventEntity{}
	linkMap := map[string]EntityEventLink{}
	for _, event := range events {
		for _, spec := range eventEntityFields {
			value := strings.TrimSpace(spec.Value(event))
			if value == "" {
				continue
			}
			id := EventEntityID(spec.EntityType, value)
			if _, exists := entityMap[id]; !exists {
				entityMap[id] = EventEntity{EntityID: id, EntityType: spec.EntityType, Value: value}
			}
			key := id + "\x00" + event.EventID + "\x00" + spec.Field
			linkMap[key] = EntityEventLink{EntityID: id, EventID: event.EventID, Field: spec.Field}
		}
		if event.SrcIP != "" && event.DstIP != "" {
			value := event.SrcIP + "->" + event.DstIP
			id := EventEntityID("network_connection", value)
			if _, exists := entityMap[id]; !exists {
				entityMap[id] = EventEntity{EntityID: id, EntityType: "network_connection", Value: value}
			}
			key := id + "\x00" + event.EventID + "\x00src_ip+dst_ip"
			linkMap[key] = EntityEventLink{EntityID: id, EventID: event.EventID, Field: "src_ip+dst_ip"}
		}
	}
	entities := make([]EventEntity, 0, len(entityMap))
	for _, entity := range entityMap {
		entities = append(entities, entity)
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].EntityID < entities[j].EntityID })
	links := make([]EntityEventLink, 0, len(linkMap))
	for _, link := range linkMap {
		links = append(links, link)
	}
	sort.Slice(links, func(i, j int) bool {
		if links[i].EventID != links[j].EventID { return links[i].EventID < links[j].EventID }
		if links[i].Field != links[j].Field { return links[i].Field < links[j].Field }
		return links[i].EntityID < links[j].EntityID
	})
	return entities, links
}

type correlationStats struct {
	ScannedPairs       int
	SuppressedPossible int
	CrowdedOutStrong   int
	ScanLimitReached   bool
}

func CorrelateNormalizedEvents(events []NormalizedEvent, entities []EventEntity, links []EntityEventLink, options EventCorrelationOptions) ([]EventCorrelation, correlationStats, error) {
	options, err := normalizeCorrelationOptions(options)
	if err != nil {
		return nil, correlationStats{}, err
	}
	entityByID := make(map[string]EventEntity, len(entities))
	for _, entity := range entities {
		if entity.EntityID == "" || entity.EntityType == "" {
			return nil, correlationStats{}, errors.New("correlation entity identity is incomplete")
		}
		entityByID[entity.EntityID] = entity
	}
	linksByEvent := map[string][]EntityEventLink{}
	for _, link := range links {
		entity, ok := entityByID[link.EntityID]
		if !ok || link.EventID == "" {
			return nil, correlationStats{}, errors.New("correlation entity-event link is incomplete")
		}
		_ = entity
		linksByEvent[link.EventID] = append(linksByEvent[link.EventID], link)
	}
	for eventID := range linksByEvent {
		sort.SliceStable(linksByEvent[eventID], func(i, j int) bool {
			return correlationEntityPriority(entityByID[linksByEvent[eventID][i].EntityID].EntityType) < correlationEntityPriority(entityByID[linksByEvent[eventID][j].EntityID].EntityType)
		})
	}

	timestamped := make([]NormalizedEvent, 0, len(events))
	seenEvents := map[string]bool{}
	for _, event := range events {
		if event.EventID == "" || event.EvidenceID == "" || event.SourceObject == "" || !sha256TextPattern.MatchString(event.SourceSHA256) {
			return nil, correlationStats{}, errors.New("correlation event provenance is incomplete")
		}
		if seenEvents[event.EventID] {
			return nil, correlationStats{}, errors.New("correlation event IDs are not unique")
		}
		seenEvents[event.EventID] = true
		if event.Timestamp != nil {
			timestamped = append(timestamped, event)
		}
	}
	sort.Slice(timestamped, func(i, j int) bool {
		if timestamped[i].Timestamp.Equal(*timestamped[j].Timestamp) { return timestamped[i].EventID < timestamped[j].EventID }
		return timestamped[i].Timestamp.Before(*timestamped[j].Timestamp)
	})

	strong := make([]EventCorrelation, 0)
	possible := make([]EventCorrelation, 0)
	stats := correlationStats{}
	perEntity := map[string]int{}
	entityCeiling := max(1, int(float64(options.MaxLinks)*maxLinksPerEntityShare))
	windowStart := 0
	stop := false
	for b := 0; b < len(timestamped) && !stop; b++ {
		for windowStart < b && timestamped[b].Timestamp.Sub(*timestamped[windowStart].Timestamp) > options.TimeWindow {
			windowStart++
		}
		for a := b - 1; a >= windowStart; a-- {
			if stats.ScannedPairs >= options.MaxScannedPairs {
				stats.ScanLimitReached = true
				stop = true
				break
			}
			stats.ScannedPairs++
			first := timestamped[a]
			second := timestamped[b]
			delta := second.Timestamp.Sub(*first.Timestamp).Seconds()
			shared, ok := strongestSharedEntity(first.EventID, second.EventID, linksByEvent, entityByID)
			if ok {
				seen := perEntity[shared.EntityID]
				if seen >= entityCeiling {
					stats.CrowdedOutStrong++
					continue
				}
				perEntity[shared.EntityID] = seen + 1
				if len(strong) >= options.MaxLinks {
					stop = true
					break
				}
				if len(strong)+len(possible) >= options.MaxLinks && len(possible) > 0 {
					possible = possible[:len(possible)-1]
					stats.SuppressedPossible++
				}
				strong = append(strong, makeEventCorrelation(first, second, delta, "related", "medium", shared, options.TimeWindow))
				continue
			}
			if !options.IncludePossible {
				continue
			}
			if len(strong)+len(possible) >= options.MaxLinks {
				stats.SuppressedPossible++
				continue
			}
			possible = append(possible, makeEventCorrelation(first, second, delta, "possible_relationship", "low", EventEntity{}, options.TimeWindow))
		}
	}
	out := make([]EventCorrelation, 0, len(strong)+len(possible))
	out = append(out, strong...)
	out = append(out, possible...)
	return out, stats, nil
}

func normalizeCorrelationOptions(options EventCorrelationOptions) (EventCorrelationOptions, error) {
	if options.TimeWindow <= 0 {
		options.TimeWindow = 5 * time.Minute
	}
	if options.TimeWindow > 24*time.Hour {
		return options, errors.New("correlation time window exceeds 24 hours")
	}
	if options.MaxLinks <= 0 {
		options.MaxLinks = defaultCorrelationMaxLinks
	}
	if options.MaxLinks > maxCorrelationLinksHard {
		return options, fmt.Errorf("correlation max links exceeds hard limit %d", maxCorrelationLinksHard)
	}
	if options.MaxScannedPairs <= 0 {
		options.MaxScannedPairs = defaultCorrelationMaxScans
	}
	if options.MaxScannedPairs > maxCorrelationScansHard {
		return options, fmt.Errorf("correlation scanned-pair limit exceeds hard limit %d", maxCorrelationScansHard)
	}
	return options, nil
}

func strongestSharedEntity(eventA, eventB string, linksByEvent map[string][]EntityEventLink, entityByID map[string]EventEntity) (EventEntity, bool) {
	idsB := map[string]bool{}
	for _, link := range linksByEvent[eventB] {
		idsB[link.EntityID] = true
	}
	for _, link := range linksByEvent[eventA] {
		entity := entityByID[link.EntityID]
		if entity.EntityType == "port" {
			continue
		}
		if idsB[link.EntityID] {
			return entity, true
		}
	}
	return EventEntity{}, false
}

func correlationEntityPriority(entityType string) int {
	switch entityType {
	case "user": return 0
	case "hostname": return 1
	case "device": return 2
	case "ip_address": return 3
	case "domain": return 4
	case "url": return 5
	case "file": return 6
	case "hash": return 7
	case "process": return 8
	case "network_connection": return 9
	case "port": return 100
	default: return 50
	}
}

func makeEventCorrelation(a, b NormalizedEvent, delta float64, relationshipType, confidence string, shared EventEntity, window time.Duration) EventCorrelation {
	basis := fmt.Sprintf("occurred within %.0fs of each other with no shared correlating entity; temporal proximity only, not a confirmed relationship", window.Seconds())
	if relationshipType == "related" {
		basis = fmt.Sprintf("share %s %q and occurred within %.0fs of each other; proximity does not establish causality", shared.EntityType, shared.Value, window.Seconds())
	}
	h := sha256.Sum256([]byte(a.EventID + ":" + b.EventID + ":" + relationshipType))
	return EventCorrelation{
		LinkID:            "LINK-" + hex.EncodeToString(h[:])[:16],
		EventIDA:          a.EventID,
		EventIDB:          b.EventID,
		RelationshipType:  relationshipType,
		Basis:             basis,
		SharedEntityID:    shared.EntityID,
		SharedEntityType:  shared.EntityType,
		SharedEntityValue: shared.Value,
		TimeDeltaSeconds:  delta,
		Confidence:        confidence,
	}
}
