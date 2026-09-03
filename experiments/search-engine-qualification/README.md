# ECO search-engine qualification — Recall / Bleve / Bluge

Status: synthetic qualification only. This branch does not modify the sealed current ECO product source.

## GitHub-first pins

- `github.com/deagy/recall` v0.3.6 — MIT — tag resolves to commit `cbccb179b965a6325e4bd3f853c8492da0acb72d`.
- `github.com/blevesearch/bleve/v2` v2.6.1 — Apache-2.0 — tag resolves to commit `048761396d42661336db8caa0bed1e98cf2aeaa6`.
- `github.com/blugelabs/bluge` v0.2.2 — Apache-2.0 — tag resolves to commit `57414197005148539c5dc5db8ab581594969df79`.

Recall v0.3.6 and Bleve v2.6.1 are current 2026 releases. Bluge's newest tag is from July 2022; later repository activity does not change that release-age disadvantage.

## Privacy architecture

All three engines are evaluated in memory only:

- Recall BM25 is inherently in-memory;
- Bleve uses `NewMemOnly`;
- Bluge uses `InMemoryOnlyConfig`.

The corpus includes only one synthetic Matter at a time. Provenance/source-region metadata remains in a separate Go map keyed by document ID and is not stored in the search engine. This mirrors ECO's rule that the search engine must not become a second plaintext case database.

## Mandatory gate

Each engine must prove:

1. expected first result for five fixed casework-style queries;
2. no result from a second synthetic Matter because scope is applied before indexing;
3. document replacement removes the old term contribution;
4. external source-region metadata can be recovered from the returned ID;
5. repeated searches produce the same normalized ranked IDs;
6. no filesystem index is created by the qualification code;
7. benchmark completes on Linux and Windows.

The harness also records build time, retained Go heap after build/GC, repeated-query latency and Go dependency count. Those are decision evidence, not standalone pass/fail thresholds.

## Decision rule

ECO does not replace a smaller proven component merely because a larger library exists. If current Recall passes the mandatory gate and remains materially lighter, it stays the default lexical ranker for the current architecture. Bleve may still qualify as the richer future engine if fielded/prefix/fuzzy/vector features justify its dependency and memory cost. Bluge must overcome both the same technical gate and its older release lineage.
