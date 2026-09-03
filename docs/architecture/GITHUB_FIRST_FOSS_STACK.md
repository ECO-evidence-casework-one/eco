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
| Google Takeout / MBOX parsing | `hanshuebner/herold` `internal/import/gmail/mbox.go` | MIT | **ADOPT AS VENDORED ADAPTATION**. Explicitly handles Takeout mboxrd, inter-message blank lines, byte offsets and `>From` unescaping. ECO preserves the root MBOX, preserves each logical RFC 5322 message as encrypted child evidence, records parent/offset/envelope provenance, and lowers the per-message default from upstream 1 GiB to 128 MiB for the low-spec target. |
| Alternative MBOX parser | `emersion/go-mbox` v1.0.4 | Repository metadata MIT; `reader.go` carries a Beer-Ware Revision 42 notice | **EVALUATED, NOT SELECTED** for the bundled parser. Keep only as a compatibility/reference probe unless the mixed file-level notice is deliberately admitted later. |
| XLSX | `qax-os/excelize` | BSD-3-Clause | **ADOPT** for workbook/sheet/cell inspection and preview. |
| MIME detection | `gabriel-vasile/mimetype` | MIT | **ADOPT** before extension-based classification. |
| Date parsing helper | `araddon/dateparse` | MIT | **ADOPT WITH GUARD**. Only use as candidate parser; ECO ambiguity rules remain controlling. |
| Investigative entity vocabulary | `opensanctions/followthemoney` | MIT | **ADOPT SELECTED SCHEMA VOCABULARY**. Map ECO Person/Organisation/Event/Document/Email to compatible fields without replacing provenance/review states. |
| Frontend data tables | `TanStack/table` | MIT | **ADOPT** for Evidence/Library/Findings tables. |
| Accessible UI primitives | `radix-ui/primitives` | MIT | **ADOPT** for dialogs, tabs, menus, tooltips and keyboard-accessible primitives. |
| Relationship graph | `xyflow/xyflow` | MIT | **ADOPT** for People/Organisations/relationship exploration. |
| React frontend | `facebook/react` | MIT | **ADOPT** as Wails frontend runtime. |
| Timeline visualization | `visjs/vis-timeline` | Apache-2.0 OR MIT | **ADOPT**. Upstream `LICENSE.md` explicitly permits either Apache-2.0 or MIT; use the interactive timeline only for derived case chronology, not as source authority. |

## Exact pins established so far

These refs and package versions were resolved directly from GitHub before admission or CI use:

- Wails v2.14.0 — commit `857398f61118eae7a6d9f5f18ffdd391590703e3` — MIT.
- Bleve v2.6.1 — commit `048761396d42661336db8caa0bed1e98cf2aeaa6` — Apache-2.0.
- Apache Tika 3.2.3 release-preparation commit `8e2ab0c58b62d74245b29689306ac6e1b79f36a1` — Apache-2.0. Tika 4.0.0 remains a separate upgrade candidate because the current Git ref resolves to an RC-stage commit.
- Herold Takeout MBOX reader — repository `hanshuebner/herold`, commit `3acbb2a8c298d09d586a0c46899ec8fa42ea5b92`, file `internal/import/gmail/mbox.go`, blob `bf50bd0c5e28e3521b5bb6b19399a33e40f788ea` — MIT.
- React `19.1.0` and React DOM `19.1.0` — exact runtime pins for the Wails frontend proof; the official Wails v2.14.0 React-TypeScript template itself declares React/React DOM `^19.1.0`.
- `@tanstack/react-table` `9.2.4` — MIT. Use the v9 API (`useTable`, `tableFeatures`, `table.FlexRender`), not v8 examples.
- `@radix-ui/react-dialog` `1.1.23` — MIT; upstream peer range explicitly supports React 19.
- `@xyflow/react` `12.11.6` — MIT.
- `vis-timeline` `8.5.4` — Apache-2.0 OR MIT; latest stable GitHub release resolved on 3 September 2026.

The Go-side compatibility probe also pins:

- `github.com/mholt/archives` `v0.1.5`
- `github.com/emersion/go-message` `v0.18.2`
- `github.com/emersion/go-mbox` `v1.0.4` — compatibility/reference probe only; not the selected bundled MBOX implementation
- `github.com/xuri/excelize/v2` `v2.11.0`
- `github.com/gabriel-vasile/mimetype` `v1.4.15`

Remaining ADOPT rows still require exact release/tag lock entries before source or binaries are imported into a releasable ECO build.

## Proven on GitHub Actions — Windows frontend stack

Run `33730332149` at commit `8daff4fbaaa93d5d55ee0e60aa2506b107279bb9` closed all four jobs successfully:

- Linux FOSS probe: PASS
- Windows Go-side FOSS probe: PASS
- Wails v2.14.0 React/TypeScript Windows shell: PASS
- Combined casework frontend: PASS

The combined proof installed the exact React/TanStack/Radix/XYFlow/vis-timeline pins, reported zero npm audit vulnerabilities, compiled the real evidence table + accessible source-details dialog + relationship graph + interactive timeline, and built a Windows executable with SHA-256 `f6f1308bf0055fbed82f9657dd288f4c3bbfbda9bfb5dc7161ad0c54308ef846`.

This is a compatibility proof, not permission to replace the current ECO UI without feature/accessibility parity qualification.

## Proven against the recovered authoritative Slice 10 parent

The later Slice 10 source was recovered from the controlled source-history archive because the GitHub `main` branch and existing development branches did not contain that exact 2 September source state.

Parent checkpoint SHA-256: `5a34cd1e606761aa8b2f57335e99b2d6d92481111f7eda24cf31fb9b018f0860`

Current controlled derivative build ID: `ECO-CASEWORK-20260903-SLICE10-GITHUB-FIRST-INTAKE-DEV`

The first real GitHub-first product integration is now working against that recovered parent:

- root MBOX is preserved and hashed before parsing;
- Google-Takeout-style mboxrd is parsed through the pinned Herold-derived MIT reader;
- each logical message is preserved separately as encrypted EML evidence;
- parent evidence ID, byte offset and bounded envelope line are retained as provenance;
- message-size, intake-entry and total-expanded-byte limits remain blocking;
- temporary staging is private and removed immediately;
- exact duplicate MBOX re-import reuses the completed intake rather than duplicating evidence;
- normal tests, release-tag tests, vet, source/offline policy, FOSS controls and targeted race tests pass;
- connected-casework qualifier remains 66/66 PASS;
- final cross-built Windows development executable is deterministic with SHA-256 `b83a0c6f8a6bf4d102eeab3a912e02f186f1de780039faec2e2cdc000b7d83f8`.

The recovered current source is **not yet represented as a full GitHub branch**. Do not pretend the historical `main` or this FOSS-probe branch is the authoritative product source until the full recovered lineage is uploaded/reconciled through an appropriate controlled route.

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

The MBOX path now proves a real part of steps 1, 2, 4, 5 and 7 against the recovered current parent. The Windows Actions proof closes technical compatibility for step 11's component stack. The remaining journey is still open and must be integrated against the current source rather than presented as complete.

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
