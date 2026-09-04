# KeepR-derived evidence occurrence provenance adaptation

Date: 2026-09-04
Upstream: `BlinkingSun/keepr`
Pinned commit inspected: `153a411cd6ad8be28bca7be079d3332662333bb1`
Licence: MIT
ECO integration: design-pattern adaptation only; no KeepR source code copied

## Problem this slice closes

Before this change, ECO correctly detected an exact SHA-256 duplicate, re-verified the retained encrypted object and re-fingerprinted the incoming source, then returned the existing evidence item with `duplicate=true`.

That avoided duplicated preserved bytes, but it discarded an evidential fact: the same content may have been supplied again under a different filename, from another folder, or at another time.

For casework, those are different propositions:

- content identity: these bytes are the same evidence object;
- occurrence provenance: this copy was encountered/supplied at this source location and time.

ECO now preserves both without storing a second encrypted object.

## What KeepR contributed

The pinned KeepR source separates an item from its original source-file metadata. Its `item_source_file` schema records original name, original extension, source SHA-256 and source byte size independently of derived pages/content.

KeepR's import path also performs duplicate detection using the original source SHA-256, rather than page/raster hashes. Its tests explicitly cover duplicate detection by source-file hash.

This is the useful donor pattern: source identity/provenance is distinct from derived content.

## Where ECO deliberately goes further

The pinned KeepR schema keeps one source-file row per item. When KeepR detects a duplicate during import, the duplicate is returned in the import result and skipped; the newly encountered duplicate is not attached as a second durable source occurrence to the existing item.

ECO therefore does **not** copy KeepR's schema literally. It adapts and strengthens the provenance principle for evidential use.

ECO records every accepted exact duplicate as a structured `evidence-occurrence-recorded` event in the existing encrypted, hash-chained change ledger.

No KeepR TypeScript, SQL schema, Electron code, SQLite store or job queue is copied into ECO.

## Occurrence model

`EvidenceOccurrences(evidenceID)` returns a chronological list of `EvidenceOccurrence` records.

The first occurrence is reconstructed from the existing verified `EvidenceItem`:

- evidence ID;
- original filename;
- current source path when still available;
- size;
- SHA-256;
- encrypted object filename;
- import/source-verification timestamps.

This read is synthetic and does not rewrite the workspace. That preserves backward compatibility for every pre-existing ECO vault.

Each later duplicate occurrence is a separately identified, authenticated audit record containing:

- occurrence ID;
- evidence ID;
- occurrence kind `duplicate-import`;
- historical source path;
- original filename at that occurrence;
- exact source size;
- exact source SHA-256;
- reused encrypted object filename;
- observation time;
- incoming-source verification time;
- retained-object verification time;
- explicit marker that the existing preserved object was reused.

Sizes and timestamps are stored as strings inside generic audit details so their exact type survives Go -> JSON -> Go workspace reloads without floating-point conversion ambiguity.

## Duplicate acceptance gate

A duplicate occurrence is recorded only after all of these conditions hold:

1. the incoming regular file is fingerprinted without changing during the read;
2. its SHA-256 matches a currently usable preserved evidence item;
3. its byte size exactly matches that evidence item;
4. ECO re-verifies the retained encrypted `.ecoobj` against the evidence ID, expected SHA-256 and expected size;
5. the incoming file is fingerprinted a second time against the same stable file identity;
6. the second SHA-256 and byte size still exactly match;
7. the occurrence record validates against the retained evidence;
8. the new authenticated change record is successfully persisted.

If the audit persistence fails, the duplicate import returns an error and the new occurrence is not silently accepted.

If a SHA-256 matches but byte sizes differ, ECO refuses deduplication rather than treating the input as the existing evidence. That is an inconsistency/collision boundary, not a normal duplicate.

## One content object, many occurrences

This slice does not create another `.ecoobj` for a duplicate. The preserved bytes remain content-deduplicated while the provenance ledger can grow independently.

This is intentional: repeated receipt of identical content should not multiply storage, but it also should not erase chronology/source history.

## Existing vault compatibility

No workspace schema bump or mass migration is required.

For evidence imported before this slice, `EvidenceOccurrences()` synthesises the initial occurrence from fields already present on `EvidenceItem`. Merely querying occurrence history does not alter the workspace or append an audit event.

When a later duplicate is encountered, only that new occurrence is appended to the authenticated change ledger.

## Portable backup semantics

The occurrence ledger is inside ECO's encrypted workspace change history, so it is automatically included in authenticated portable backups.

ECO's existing restore behavior clears `EvidenceItem.SourcePath` because that field is the current/live source locator and may be invalid on another machine.

A duplicate occurrence's source path is different: it is historical provenance captured at the time of the verified occurrence. This slice intentionally retains that historical path inside the encrypted occurrence audit record across portable backup/restore.

The backup remains encrypted; the path is not written in plaintext outside the encrypted workspace/backup container by this feature.

## Query validation

`EvidenceOccurrences()` does not blindly trust generic audit details. It validates every matching occurrence record before returning it, including:

- bounded occurrence identity;
- exact evidence ID;
- exact retained SHA-256;
- exact retained byte size;
- exact encrypted object filename;
- allowed occurrence kind;
- bounded original name/source path;
- complete timestamps;
- explicit duplicate-object reuse marker;
- unique occurrence IDs.

A malformed occurrence audit record causes the query to fail visibly rather than silently emitting questionable provenance.

## Tests

The integration tests prove:

- two differently named/source-located copies create two occurrences but only one encrypted `.ecoobj`;
- the duplicate import returns the original evidence item;
- every occurrence identifies the exact same SHA-256/size/object;
- duplicate occurrences carry authenticated change IDs and verification timestamps;
- occurrence history survives closing/reopening the encrypted vault;
- repeated imports of the same path create distinct occurrence events rather than being silently discarded;
- legacy initial occurrence queries require no workspace migration;
- malformed audit occurrence metadata is rejected;
- missing evidence queries fail;
- occurrence history survives encrypted portable backup/restore;
- the restored live source locator remains cleared while historical duplicate provenance remains available.

## Non-claims

An occurrence proves only what ECO directly verified at intake: a source file at the recorded locator was read and its bytes exactly matched the retained evidence content at that verification event.

It does not prove who placed the file there, when the file was originally created, why it was supplied, who possessed it before observation, or that two identical copies have independent provenance before ECO encountered them.
