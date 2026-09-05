# Current ECO project status

**Status date:** 5 September 2026  
**Baseline `main` reviewed for this record:** `34e96faea669ed29e4c8f63b3c333ee642f29dbf`  
**Post-merge CI:** run `33957055962` — source policy, Gitleaks, Linux tests/vet, deterministic Windows build/tests, Syft reconciliation and Cosign signing/tamper rehearsal all passed  
**Release position:** source development only; no approved public source release, signed end-user executable or real-evidence use

## Authority and evidence terminology

This file is ECO's canonical public status summary. Issues, pull requests, exact commits, workflow runs and qualification records remain the detailed evidence. A feature being implemented or a CI run passing does not by itself open a release, real-evidence, accessibility, publisher or deployment gate.

## Live controlled position

| Area | Current position |
|---|---|
| Authoritative source | `main` at `34e96faea669ed29e4c8f63b3c333ee642f29dbf` |
| Qualified GitHub/FOSS stack | Merged through PR #122 after combined CI and controlling 8 GB Acer PASS |
| Native PDF text | Implemented from qualified BSD-3-Clause `ledongthuc/pdf` source; page-aware extraction present |
| PDF visual rendering/navigation | Optional exact-qualified `pdfium-cli` WebAssembly runtime; page count and bounded multi-page preview navigation implemented |
| OCR | Tesseract adapter plus complete verified Windows runtime-bundle registration; real OCR qualification passed |
| Mailbox ingestion | Bounded MBOX reader using MIT `emersion/go-mbox`; mutation fuzz qualification passed |
| Hostile-input testing | Permanent fuzz targets for MBOX, ZIP, EML/MIME, XML/Office text and file sniffing; qualification campaigns passed |
| Local runtime control | Encrypted audited registry for approved Tesseract, Docling, OCRmyPDF, llama.cpp, pdfcpu and pdfium-cli executables |
| Release-integrity rehearsal | Gitleaks, deterministic Windows build, Syft SBOM reconciliation and private Cosign envelope/tamper checks are mandatory CI controls |
| Low-spec PDF gate | PASS on the actual 8 GB Acer: 7.68 GiB RAM, median render 8103 ms, worst 8319 ms, peak renderer working set 291.4 MiB |
| Workspace/state safety | Issue #4 remains open P0; old PR #72 contains useful isolated ownership primitives but is stale and must not be merged wholesale |
| Preserved evidence/source truth | Issue #3 remains open P0; much downstream OCR/PDF/runtime work now uses freshly verified preserved objects, but issue acceptance is not fully closed |
| Ask/restore concurrency | Issue #12 remains open P1 |
| Offline AI reliability | Issue #5 remains open P0 despite bounded grounding and llama.cpp integration; issue acceptance tests are not fully proved |
| Responsiveness | Issue #6 remains open P1 |
| Accessibility | Issue #7 remains open P1 and blocks public preview |
| Page-aware search | Issue #8 remains open P1; PDF page navigation is implemented, but full workspace/document search, match highlighting/receipts and accessible match navigation are not complete |
| Diagnostic privacy/offline claims | Issue #14 remains open P0 |
| Actual released-build provenance | Issue #15 remains open P0; CI rehearsal is materially stronger but does not equal a final packaged, Authenticode-signed release |
| Intended purpose / claims | Issues #16, #20, #23 and #46 remain open; merged technical controls do not close cross-surface claims and high-consequence-output gates |
| Publisher/steward | Issue #17 remains open; no accountable legal publisher/steward is appointed |
| Historical Actions binaries | All four previously listed unsigned artifacts now report `expired: true`; issue #24 remains open for the other distribution-control acceptance items |
| M1.18 hardening | Issue #65 is CLOSED; older status text saying it blocks all integration is obsolete |
| V40 deadline | Issue #69 is a superseded historical control record, not an active delivery deadline or application-parent instruction |
| Real or sensitive evidence | Blocked |
| Signed public executable | None |
| Institutional, healthcare, justice-sector and EU/EEA availability | Blocked |

## September GitHub/FOSS qualification checkpoint

PR #122 promoted the combined qualified stack into `main`. Its exact staging head `85c670e2b3a5fc5a3d87d814e896d749d8cd732b` passed ordinary complete CI in run `33948613469`. After that, the controlling Acer qualification `ECO-PDF-LOW-SPEC-20260905.1` passed on the actual low-spec Windows machine. PR #122 then merged as `34e96faea669ed29e4c8f63b3c333ee642f29dbf`, and post-merge main run `33957055962` passed the complete build/test/release-integrity pipeline.

This checkpoint means the PDF/OCR/mailbox/runtime/fuzz work is no longer merely branch research. It is part of canonical `main`. It does **not** mean ECO is a public release or suitable for real evidence.

## Superseded or stale lanes

### PR #71 — V40 Matter journey

PR #71 remains open/draft, but issue #69 explicitly supersedes the V40 application shell as an application parent. Do not revive or merge PR #71 as the product baseline. Useful UX requirements may be reimplemented selectively against current `main`.

### PR #72 — Workspace Ownership V2

PR #72 remains open/draft and contains useful ownership primitives and hostile tests, but its branch predates current `main` and its own description says it does not complete issue #4. Do not merge it wholesale. Reuse/adapt the proven primitives in small current-main slices, with Vault/restore lifecycle integration, stale-state/CAS controls and current CI.

### Issue #65

Issue #65 is closed for its bounded M1.18 repair/recovered-worker proof. Older documents saying it blocks every adapter/model/runtime connection are stale. Closure does not approve real evidence, public release or high-consequence outputs.

### Issue #69

Issue #69 remains open only as a visible supersession/history record. The former 9 August V40 target is not an active delivery instruction.

## Current release blockers

### P0 engineering / assurance

- #3 preserved evidence/source truth — open.
- #4 clean explicit workspace state / ownership — open.
- #5 offline AI instruction/source/reliability boundary — open.
- #14 privacy-safe diagnostics and exact offline/network claims — open.
- #15 final packaged-build SBOM/licences/provenance/signing — open.
- #20 high-consequence/clinical/professional-output gate — open.
- #24 public runnable-artifact controls — open, although the four historical artifacts are now expired.

### P1 product / usability

- #6 responsiveness and long-running-work behaviour — open.
- #7 keyboard, screen-reader, scrolling and high-DPI accessibility — open.
- #8 page-aware search/highlighting/source navigation — open.
- #12 Ask verification cost and restore serialisation — open.
- #16 intended-purpose/claims governance — open.
- #17 accountable publisher/steward — open.
- #23 user-facing claim reconciliation — open.
- #46 UI/AI/export intended-purpose conformance — open.

## What is now proved on the low-spec Acer

The PDF visual-rendering/navigation workload completed successfully on the controlling 8 GB Acer even though the machine began the test with only about 1.63 GiB available RAM and about 78% RAM usage. The pre-set acceptance limits were met:

- median ECO page render: 8103 ms (`<= 10000` target);
- slowest page render: 8319 ms (`<= 15000` target);
- peak renderer working set: 291.4 MiB (`<= 768` target);
- page count and `1 -> 2 -> 3 -> 2 -> 1` navigation passed;
- out-of-range page rejection passed.

This proves the qualified PDF visual path on that machine. It does not prove all ECO workflows, long-duration use, accessibility, clean-install packaging or model performance on the Acer.

## Development sequence from current `main`

1. Reconcile canonical status/release documents to this checkpoint.
2. Rebuild **workspace/state ownership (#4)** from current `main`, selectively reusing the proven PR #72 primitives rather than reviving the stale branch.
3. Close remaining **preserved-source / Ask-restore consistency (#3/#12)** gaps with current-main tests.
4. Build the missing **page-aware search/highlighting (#8)** layer on the now-qualified extraction/OCR/PDF foundation.
5. Qualify **responsiveness and accessibility (#6/#7)** on the actual Windows UI, including keyboard-only, high-DPI and Narrator/NVDA evidence.
6. Close **diagnostic privacy/offline claims (#14)** and user-facing claim alignment (#16/#20/#23/#46).
7. Complete final **packaging, actual-release SBOM/licences, Authenticode signing, installer/uninstaller and publisher/steward controls (#15/#17/#24)** before any public binary decision.
8. Reassess local generative AI (#5) only against the exact low-spec model/runtime package and the permitted-output boundaries; do not let model packaging delay the deterministic evidence product.

## Stop gates

The following remain blocked:

- real, sensitive or irreplaceable evidence use;
- claims that ECO is production-ready, court-ready, forensic, legally/medically compliant or accessible;
- public executable distribution or stable-release status;
- trusted signing claims before the final exact artefact is Authenticode-signed and reconciled;
- official institutional, healthcare, justice-sector or EU/EEA supply/deployment;
- high-consequence scoring, diagnosis/treatment, professional representation or authority-side adverse-decision use.

## Public-record rule

Use synthetic and non-sensitive information only. Do not publish personal evidence, private workspaces, credentials, private diagnostics, exploit-level instructions, unapproved executable/model files or sensitive case material in repository records.
