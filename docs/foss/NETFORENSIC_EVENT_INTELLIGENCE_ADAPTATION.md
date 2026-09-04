# NetForensicAI event-intelligence adaptation for ECO

Date: 2026-09-04

## Upstream source reviewed

- Project: `Sh3n0bi/NetForensicAI`
- Pinned source commit: `70cd1b98a4433ab6fe2db6280799d33673acea9e`
- Upstream licence: MIT
- Reviewed modules:
  - `netforensicai/core/event.py`
  - `netforensicai/core/entities.py`
  - `netforensicai/core/correlation.py`
  - `netforensicai/core/evidence.py`
  - `netforensicai/core/case.py`
  - `netforensicai/core/store.py`

## Useful upstream control ideas

The pinned donor has several patterns worth carrying into ECO:

1. One normalized event shape across source formats, with event identity and evidence traceability mandatory.
2. Deterministic entity IDs derived from entity type plus normalized value, so the same real-world identifier does not multiply simply because it appears in several records.
3. Source ports are deliberately not entities because they are ephemeral; destination ports remain indexable because service-port queries can be useful.
4. Destination ports are indexable but deliberately excluded from correlation because a shared common port is generally noise, not a meaningful relationship.
5. A directional source/destination host pair is represented as a `network_connection` entity.
6. Stronger correlation requires both a shared correlating entity and time proximity.
7. Temporal proximity without a shared entity is a separate low-confidence `possible_relationship` tier and must not be described as causality.
8. Pair scanning and retained-link counts are bounded, and one high-degree entity is prevented from consuming the whole relationship budget.
9. Re-parsing an evidence item should replace its derived event set rather than accumulate duplicates.
10. Large event collections belong outside a monolithic case/workspace JSON object.

## ECO-specific implementation in this slice

`internal/eco/event_intelligence.go` independently implements the reusable event/entity/correlation controls in Go.

The main entry point is `Vault.AnalyzeEvidenceEvents(...)` / `AnalyzeEvidenceEventsContext(...)`.

A parser is not handed the user's original file path. ECO first resolves the committed preservation record and uses `withVerifiedPreservedFile` to produce a freshly verified temporary reading copy. The parser receives that path plus an app-owned `SourceReceipt`. ECO then hydrates every normalized event with:

- `evidence_id`;
- preserved `source_object`;
- preserved `source_sha256`;
- a deterministic event ID based on the evidence ID and parser sequence.

If a candidate names an ECO source segment, that segment must already be bound to the same preserved object.

The engine then builds deterministic entities and bounded correlations. Correlation output uses only `medium` confidence for shared-entity-plus-time proximity and `low` confidence for temporal-only proximity. Its explanatory text explicitly states that proximity does not establish causality.

## Storage decision: intentionally not copied from the donor

NetForensicAI stores large event collections in a per-case DuckDB database. That is sensible for its architecture and is much better than putting tens of thousands of events in one JSON case file.

ECO cannot simply copy that design because ECO currently keeps evidence-derived text and workspace metadata encrypted at rest. Adding a normal DuckDB file would create a potentially large plaintext forensic derivative containing timestamps, usernames, IP addresses, file names, commands, URLs and other sensitive material.

Therefore this first adaptation **returns event intelligence in memory and does not create a plaintext analytical database**. This is deliberate, not an unfinished accidental write path. A scalable persistent event store must preserve ECO's at-rest security model before it is connected.

## Bounded first-slice limits

- Maximum candidate events per verified extraction call: 250,000.
- Default correlation window: 5 minutes; hard maximum: 24 hours.
- Default retained relationship budget: 50,000; hard maximum: 250,000.
- Default scanned-pair CPU budget: 2,000,000; hard maximum: 10,000,000.
- No single shared entity may consume more than 10% of the retained-link budget.
- Event field lengths, ports, process IDs, source references and parser identities are bounded and validated.

## What this does not claim

A `related` event pair means only that the events share a non-noisy entity and occurred within the configured time window. A `possible_relationship` means only temporal proximity. Neither means one event caused the other, that the events are part of the same incident, or that any allegation is proved.
