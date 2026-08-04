# Third-party notices — current source status

This file records the reviewed **source-level** dependency and notice position for current `main`.

It is not the final notice bundle for a released executable and does not satisfy P0 issue #15. Final notices must be generated and reconciled against the exact packaged executable, embedded payload, runtime-created files, SBOM, content manifest and corresponding source.

## Current source declarations

- `go.mod` declares no third-party Go module dependency.
- The source uses the Go standard library, distributed under the Go project's BSD-style licence.
- The Windows source calls Microsoft Windows system libraries supplied by the operating system; these are dependencies of the program but are not presently redistributed by ECO.
- No production OCR engine, OCR language data, AI runtime or AI model is bundled in the current recorded V25 milestone.

These statements describe reviewed source declarations only. They do not replace inspection of the actual PE import table, resources, sections, overlay, embedded payload, runtime extraction or clean-machine behaviour of a future final executable.

## Historical V25 notice record

The exact former V25 N2 P1 notice file is preserved at:

`docs/provenance/historical/ECO_V25_N2_P1_THIRD_PARTY_NOTICES_2026-07-31.md`

That historical record is not authoritative for later source commits or a future release.

## Release authority

P0 issue #15 controls the future actual-build SBOM, licence, notice and provenance package. No executable, public binary, ordinary-user test or release is approved by this source-level notice file.
