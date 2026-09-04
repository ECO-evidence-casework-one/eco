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


## ledongthuc/pdf — native PDF reader

- Upstream: `https://github.com/ledongthuc/pdf`
- Exact source commit: `b3c860c2375335b0bc6676c430107a553725991d`
- Licence: BSD-3-Clause; verbatim licence retained at `third_party/ledongthuc_pdf/LICENSE`.
- ECO compatibility change: the vendored local module declares Go 1.23 instead of upstream's Go 1.24.1 module directive; qualified source files are otherwise copied from the exact upstream commit.
- Purpose: bounded, page-aware native-text extraction from PDFs. Scanned/image-only PDFs still require a separately registered local OCR path.


## emersion/go-mbox — MBOX message framing

- Upstream: `https://github.com/emersion/go-mbox`
- Exact acquired commit: `1345da99f1254a23f517ffdc979f92359442473d`
- Licence: MIT.
- ECO use: bounded, read-only MBOX message framing before ECO's existing MIME/email extraction.
