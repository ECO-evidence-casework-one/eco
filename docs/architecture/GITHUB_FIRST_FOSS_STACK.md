# ECO GitHub-First FOSS Integration Stack

Status: controlled integration plan — **not a release approval**

Branch: `integration/github-first-stack-20260903`

## Rule

ECO must reuse mature free/open-source software before creating a bespoke implementation of a commodity capability. Custom ECO code is reserved for the product-specific trust layer: immutable source identity, provenance, user-confirmed state, contradiction preservation, encrypted Vault storage, bounded recovery, audit receipts, matter isolation, and controlled AI behaviour.

No third-party component is admitted merely because it is popular or available on GitHub. Admission requires an exact upstream repository/ref, compatible licence, defined role, offline/privacy assessment, and an ECO adapter that preserves the trust model.

## Target architecture

### Keep as ECO-owned core

- SHA-256 source identity and preserved originals
- encrypted Vault and encrypted Library catalogue
- `InformationOrigin`, `SourceLink`, source-region and citation controls
- user confirmation/correction/review states
- conflicting-source preservation and missing-evidence caution
- Matter isolation
- resumable jobs, checkpoint DAG, Startup Recovery and bounded Retry/Resume
- Sentinel/diagnostic receipts
- Ask ECO capability boundary and deterministic fact store

### Replace/accelerate with FOSS

| Area | Upstream | Licence | Integration decision |
|---|---|---|---|
| Windows desktop shell | `wailsapp/wails` v2 | MIT | **ADOPT**. Replace hand-drawn Win32 presentation layer with Wails v2 + web frontend while keeping Go backend. |
| Full-text search | `blevesearch/bleve/v2` | Apache-2.0 | **ADOPT**. Matter-scoped memory-first index initially; no plaintext persistent case index until encrypted-at-rest design is proved. |
| Broad text/metadata extraction | `apache/tika` | Apache-2.0 | **ADOPT AS WORKER**. Local crash-isolated parser; derived text/metadata only, never source authority. |
| PDF render/text/search | `klippa-app/go-pdfium` + PDFium | MIT + Apache-2.0 | **ADOPT**. Prefer WASM implementation to avoid CGO/runtime fragility. |
| OCR | `tesseract-ocr/tesseract` | Apache-2.0 | **ADOPT AS WORKER**. OCR output remains `ocr-reading`, never silently promoted to fact. |
| Archive/container traversal | `mholt/archives` | MIT | **ADOPT** under ECO depth/size/path/cancellation limits. |
| EML/MIME parsing | `emersion/go-message` | MIT | **ADOPT**. Preserve raw RFC bytes and parse headers/body/attachments as derived records. |
| MBOX parsing | `emersion/go-mbox` | MIT | **ADOPT** feeding messages to `go-message`. |
| XLSX | `qax-os/excelize` | BSD-3-Clause | **ADOPT** for workbook/sheet/cell inspection and preview. |
| MIME detection | `gabriel-vasile/mimetype` | MIT | **ADOPT** before extension-based classification. |
| Date parsing helper | `araddon/dateparse` | MIT | **ADOPT WITH GUARD**. Only use as candidate parser; ECO ambiguity rules remain controlling. |
| Investigative entity vocabulary | `opensanctions/followthemoney` | MIT | **ADOPT SELECTED SCHEMA VOCABULARY**. Map ECO Person/Organisation/Event/Document/Email to compatible fields without replacing provenance/review states. |
| Frontend data tables | `TanStack/table` | MIT | **ADOPT** for Evidence/Library/Findings tables. |
| Accessible UI primitives | `radix-ui/primitives` | MIT | **ADOPT** for dialogs, tabs, menus, tooltips and keyboard-accessible primitives. |
| Relationship graph | `xyflow/xyflow` | MIT | **ADOPT** for People/Organisations/relationship exploration. |
| React frontend | `facebook/react` | MIT | **ADOPT** as Wails frontend runtime. |
| Timeline visualization | `visjs/vis-timeline` | pending exact licence intake | **HOLD UNTIL LICENCE PINNED**. |

## Initial exact pins

These refs were resolved directly from GitHub before this branch was created:

- Wails v2.14.0 — commit `857398f61118eae7a6d9f5f18ffdd391590703e3` — MIT.
- Bleve v2.6.1 — commit `048761396d42661336db8caa0bed1e98cf2aeaa6` — Apache-2.0.
- Apache Tika 3.2.3 release-preparation commit `8e2ab0c58b62d74245b29689306ac6e1b79f36a1` — Apache-2.0. Tika 4.0.0 is still being treated as a separate upgrade candidate because the current Git ref resolves to an RC-stage commit.

All remaining ADOPT rows require their exact release/tag commit to be appended to the lock before source or binaries are imported.

## Reference systems — benchmark, do not embed wholesale

### IPED (`sepinf-inc/IPED`)

Use as a forensic architecture and acceptance benchmark. It already demonstrates hashing/deduplication, signatures, recursive containers, Tesseract OCR, full-text/metadata indexing, NER, timelines, communication graphs, bookmarks/tags, reports, out-of-process parsing, and resume/restart processing. ECO does not need its disk-image/carving/LE-specific surface wholesale.

### ICIJ Datashare (`ICIJ/datashare`)

Use its ingest/search/OCR/entity-analysis workflow and document-search UX as a benchmark. Do not embed its full server stack: current architecture assumes Java + PostgreSQL + Elasticsearch + Redis and is too heavy for ECO's low-spec, single-user, one-desktop-app target.

### FollowTheMoney

Use selected schema/vocabulary as an interoperability and modelling reference. ECO remains stricter about evidence source, origin, user confirmation, conflicts and uncertainty.

## First vertical slice

The first GitHub-first build must prove one complete user journey instead of many disconnected modules:

1. Add a folder or archive.
2. Preserve original bytes and SHA-256 identity in ECO.
3. Detect MIME from content.
4. Traverse supported containers safely.
5. Extract text/metadata through Tika; parse EML/MBOX/XLSX through specialist Go libraries where useful.
6. OCR image-only content through Tesseract.
7. Store all derived outputs with exact source/provenance labels.
8. Build a Matter-scoped Bleve index in memory from authorised derived text.
9. Search and open the exact source/region.
10. Surface People/Organisations/Events/Documents/Emails with FollowTheMoney-compatible fields plus ECO review/provenance state.
11. Show relationships and timeline in the new Wails UI.
12. Preserve conflicts/missing evidence rather than choosing a winner.
13. Ask ECO only against deterministic retrieval/citations.
14. Create a task/action and export a source-backed case brief.

This journey is the next meaningful acceptance milestone. A candidate that cannot complete it is not to be presented as a major product leap.

## Security/privacy boundaries

- no cloud dependency or telemetry is introduced by this architecture;
- no third-party component may mutate preserved originals;
- external parsers run as bounded local workers where process isolation is useful;
- archive traversal remains subject to ECO limits against path escape, decompression bombs, runaway depth and resource exhaustion;
- OCR/parser/model text is derived/untrusted input;
- search indexes must not create an unencrypted shadow copy of the Matter;
- AI cannot become source of truth or autonomous case actor.

## Branch/release control

This branch is deliberately separate from `main`. The GitHub `main` branch is historical relative to the later private/resilience/Slice development history. Do not merge this integration branch into `main` until the current authoritative source lineage has been reconciled, the vertical slice is working, licences/SBOM are complete, and the existing release gates are explicitly reviewed.
