# Ethos grounding adaptation for ECO

Date: 2026-09-04

## Upstream source reviewed

- Project: `docushell/ethos`
- Pinned source commit: `d2423bb189d153ccab037ce734c9b8c275586a4b`
- Upstream licence: Apache-2.0
- Primary reviewed contract: `docs/citation-emission-spec.md`
- Primary reviewed reference implementation: `python/ethos_pdf/emit.py`
- Contract test reviewed: `.github/scripts/test_citation_emission_v1_contract.py`

The Wave-1 donor acquisition audit also records the exact upstream commit and licence-file hashes.

## What ECO adopted

ECO uses an independently written Go implementation of the following control ideas:

1. **Shown-ID vocabulary.** A model may cite only evidence/segment IDs that ECO placed in the exact retrieval context for that answer.
2. **Trusted hydration.** Source hashes, encrypted object names, coordinates and verification state remain application-owned and are added only after the model output has passed deterministic checks.
3. **Fail closed on invented IDs.** An unknown evidence or segment ID rejects the entire emission instead of being guessed, repaired or silently dropped.
4. **Deterministic text grounding.** `quote` and `value` claims must occur in the exact text shown to the model, allowing only whitespace normalisation as a fallback. A fabricated value produces a negative grounding report and no releasable citations.
5. **Presence claims.** A presence claim may identify a shown source without inventing text.
6. **Reverification before release.** ECO rebinds the trusted context to the current workspace and verifies the preserved encrypted object again before returning hydrated citations.
7. **Grounding is not semantic truth.** A green grounding result proves that the asserted source text was actually present in the cited material. It does not prove that the material itself is true, complete, legally correct or relevant.
8. **No retry-to-green rule.** A negative text-grounding verdict is data to preserve and handle, not a reason to regenerate repeatedly until a model happens to pass.

## ECO-specific differences

ECO's native source vocabulary is the pair `(evidence_id, segment_id)` because every `SourceSegment` is already bound to the preserved object's SHA-256 and object filename. The model never emits those trusted hashes.

ECO currently supports model-facing grounding kinds `quote`, `value`, and `presence`. `table_cell` is intentionally rejected until ECO has a controlled native table/cell identity model. Accepting table coordinates before that would create fake precision.

ECO's context is ephemeral. Its public records can be serialised to a local model, but the private trusted mapping cannot. A context reconstructed from model-facing JSON is therefore non-authoritative and verification rejects it.

## Integration boundary

`internal/eco/grounding.go` provides:

- `BuildGroundingContext(...)` — creates the bounded, verified records that may be shown to a local model;
- `VerifyGroundingEmission(...)` — validates model claims against that exact context, re-verifies preserved evidence and returns citations only when every requested claim is grounded.

This boundary is intended to sit directly between ECO retrieval and the planned local `llama.cpp` adapter.

## What this does not claim

This is not an Ethos verifier port and does not claim byte-compatible Ethos reports. ECO is adapting the fail-closed application pattern to its own preserved-evidence model. Ethos remains a separately licensed upstream donor/reference project.
