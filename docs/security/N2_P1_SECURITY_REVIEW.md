# Version 25 N2 P1 security review

## Scope

This review covers the new image-transformation and OCR-provenance foundations. It does not approve a production OCR engine, model pack or end-user release.

## Controls implemented

- OCR receipts require a bounded engine identity, valid source SHA-256, explicit status, timestamp and confidence values.
- OCR words and lines require bounded text, valid pages, finite confidence values and valid normalised image regions.
- OCR collection sizes and individual text fields are capped.
- Nested OCR words are validated rather than trusting only the top-level word list.
- Source segments must match the validated OCR receipt line-for-line, including text, page, confidence and exact region.
- Duplicate source-segment IDs and non-OCR origins are rejected.
- OCR results are bound to the preserved original evidence SHA-256.
- OCR application is transactional in memory: a failed authenticated workspace write restores the prior evidence and change-log state.
- The original encrypted evidence object is never rewritten by crop, deskew, reading enhancement, perspective correction or OCR.
- Very low-confidence OCR segments are excluded from Ask ECO retrieval.
- Large image previews are bounded before native display processing.
- Perspective-correction points must be finite, inside the source image, convex, ordered and non-degenerate.
- Application source remains free of network imports and embedded credential patterns under the current source-policy scan.

## Remaining risks

- No isolated OCR worker process exists yet.
- No production OCR executable or language data is bundled or approved.
- Derived image renditions are not yet stored with authenticated transformation receipts.
- The current perspective operation is a transformation foundation, not a reviewed interactive editor.
- Image decoders still execute inside the main process.
- Direct Windows exploit, fuzzing, malformed-image and large-batch evidence remains incomplete.
- The current workspace format is an early development format and is not approved for irreplaceable evidence.
- Trusted code signing is pending.

## Disposition

Suitable to publish as source for an early development snapshot. Not suitable to describe as a secure stable release or to distribute as an unsigned end-user executable.
