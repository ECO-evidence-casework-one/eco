# ECO FOSS Donor Register — Wave 1

Date: 2026-09-04
Branch: `integration/foss-acquisition-wave1-20260904`

## Rule

GitHub/FOSS first. ECO should reuse mature local/offline components and proven implementation patterns before new bespoke code is written. Direct vendoring is not automatic: each donor is assigned an integration mode based on language, size, licence obligations, security boundary and whether ECO needs the whole engine or only an adapter/pattern.

## Wave 1

| Project | Upstream | Expected licence | ECO role | Integration mode |
|---|---|---|---|---|
| Ethos | `docushell/ethos` | Apache-2.0 | deterministic citation/evidence grounding | local CLI/library adapter + source review |
| NetForensicAI | `Sh3n0bi/NetForensicAI` | MIT | evidence timeline, entities, graph, evidence-cited findings | adapt patterns; do not import unrelated DFIR/network features |
| KeepR | `BlinkingSun/keepr` | MIT | content-addressed library, occurrence tracking, OCR/job queue patterns | adapt selected patterns |
| Docling | `docling-project/docling` | MIT codebase | broad document conversion and structure extraction | external local engine adapter |
| OCRmyPDF | `ocrmypdf/OCRmyPDF` | MPL-2.0 | scanned PDF OCR pipeline | external local engine adapter; keep licence boundary explicit |
| Tesseract | `tesseract-ocr/tesseract` | Apache-2.0 | OCR engine | external local engine adapter |
| pdfcpu | `pdfcpu/pdfcpu` | Apache-2.0 | PDF inspection/processing | direct Go library candidate after dependency review |
| llama.cpp | `ggml-org/llama.cpp` | MIT | local LLM runtime | external local engine adapter |
| restic | `restic/restic` | BSD-2-Clause | backup/restore | external local engine adapter |
| Litestream | `benbjohnson/litestream` | Apache-2.0 | SQLite recovery/replication patterns | external/local recovery option after architecture review |
| Cosign | `sigstore/cosign` | Apache-2.0 | release signing and provenance | build/release tool, not runtime dependency |
| Gitleaks | `gitleaks/gitleaks` | MIT | secret scanning | development/CI tool |
| Velopack | `velopack/velopack` | MIT | Windows installation/update | packaging/update study and adapter |
| gopsutil | `shirou/gopsutil` | BSD-3-Clause | CPU/RAM/disk/process observability | direct Go library candidate after dependency review |
| PaddleOCR | `PaddlePaddle/PaddleOCR` | Apache-2.0 | advanced OCR/layout | optional large external engine; not cloned by default |

## Acquisition control

`scripts/foss/acquire-donors.ps1` clones the source repositories into `E:\ECO\FOSS_DONORS\SOURCE`, records exact Git commits, copies root licence/notice files into `LICENSE_SNAPSHOTS`, hashes those files, and emits CSV/JSON inventories under `REPORTS`.

PaddleOCR is skipped unless `-IncludeLarge` is explicitly requested because it is a very large repository. Existing dirty checkouts are never overwritten by the refresh mode.

## Integration order

1. Tesseract OCR adapter — first glue slice. ECO already understands Tesseract-compatible TSV, so invoking a selected local Tesseract binary is the smallest useful complete integration.
2. Docling document adapter — use for broad structured extraction where ECO's current handwritten parsers are weaker; keep original bytes and ECO provenance authoritative.
3. Ethos grounding adapter — verify citations/source references before answer/report release.
4. KeepR/NetForensicAI pattern extraction — catalogue/occurrence/job state, entity/timeline/graph patterns; selectively reimplement compatible concepts in Go rather than embedding unrelated applications.
5. restic/Litestream recovery adapter — tested backup and restore paths.
6. llama.cpp runtime adapter — truthful local AI path with no cloud fallback.
7. gopsutil operation telemetry.
8. release tooling — Gitleaks, Cosign and installer/update controls.

## Current implementation in this branch

The first real glue is `internal/eco/tesseract_adapter.go`. It invokes only a caller-selected absolute local Tesseract executable, writes OCR output to a bounded temporary workspace, refuses oversized TSV output, captures the exact Tesseract version, and then feeds the output into ECO's existing provenance-bound `ParseOCRTSV` path. It performs no download and no network action.

## Not approved by this register

This register does not approve blindly copying whole repositories into ECO, enabling cloud services, enabling optional hosted-AI features, or adopting a dependency solely because its upstream project is open source. Exact licence/dependency review remains required before direct vendoring or binary redistribution.
