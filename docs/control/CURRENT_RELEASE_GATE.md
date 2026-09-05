# Current ECO release gate

**Gate record:** `ECO-RELEASE-GATE-20260905-012`  
**Updated:** 5 September 2026  
**Canonical public status:** [`../../CURRENT_STATUS.md`](../../CURRENT_STATUS.md)  
**Baseline `main` reviewed for this record:** `34e96faea669ed29e4c8f63b3c333ee642f29dbf`  
**Post-merge CI:** run `33957055962` — all normal jobs passed  
**Approved public source release:** none  
**Approved signed end-user executable:** none

## Controlling decision

ECO remains a source-development project only. The September GitHub/FOSS stack is now part of canonical `main`, and several capabilities have materially advanced, but the public/real-evidence release gates remain closed.

No current source, workflow, Acer test, runner-built executable, model/runtime adapter or governance record authorises:

- real, sensitive or irreplaceable evidence use;
- reliable professional, clinical, forensic or high-consequence AI claims;
- ordinary-user public executable distribution;
- release-candidate or stable-release status;
- institutional, healthcare, justice-sector or EU/EEA supply/deployment.

## Newly qualified canonical capabilities

PR #122 merged the combined qualified stack as `34e96faea669ed29e4c8f63b3c333ee642f29dbf`. The exact combined source passed complete CI before merge and again on `main` after merge. Canonical `main` now includes:

- native PDF text extraction from a qualified BSD-3-Clause Go reader;
- optional exact-qualified `pdfium-cli` page rendering;
- bounded multi-page PDF preview navigation;
- Tesseract OCR plus verified complete Windows runtime-bundle registration;
- bounded MBOX reading using MIT `emersion/go-mbox`;
- permanent hostile-input fuzz targets for MBOX, ZIP, EML/MIME, XML/Office text and file sniffing;
- approved-local-runtime registration and re-verification controls;
- Gitleaks secret scanning;
- gopsutil resource pressure guard;
- deterministic Windows build checks;
- Syft SBOM generation/reconciliation;
- private Cosign signing/tamper rehearsal.

The controlling 8 GB Acer PDF qualifier also passed with median render 8103 ms, worst render 8319 ms and peak renderer working set 291.4 MiB, within the pre-set limits.

## Gate matrix

| Gate | Current position |
|---|---|
| Preserved evidence/source truth | **OPEN P0 — issue #3.** Many new workflows use freshly verified preserved objects, but issue acceptance is not fully closed. |
| Workspace ownership / clean state | **OPEN P0 — issue #4.** PR #72 has useful stale primitives but is not mergeable wholesale. |
| Offline AI reliability | **OPEN P0 — issue #5.** Grounding/llama.cpp controls exist; full issue acceptance remains incomplete. |
| Responsiveness | **OPEN P1 — issue #6.** |
| Accessibility | **OPEN P1 — issue #7.** Blocks public preview. |
| Page-aware search/source navigation | **OPEN P1 — issue #8.** PDF page navigation is implemented; full search/highlight/receipt/accessibility acceptance is not. |
| Ask verification / restore serialisation | **OPEN P1 — issue #12.** |
| Diagnostic privacy / final offline claims | **OPEN P0 — issue #14.** |
| Final packaged-build provenance | **OPEN P0 — issue #15.** CI rehearsal is strong but does not equal a final packaged Authenticode-signed release. |
| Intended purpose / claims | **OPEN — issues #16, #20, #23 and #46.** |
| Publisher/steward | **OPEN P1 — issue #17.** No accountable legal publisher/steward is appointed. |
| Public runnable-artifact controls | **OPEN P0 — issue #24.** The four historical artifacts now report `expired: true`, but remaining acceptance items are open. |
| M1.18 hardening | **CLOSED — issue #65.** Older text saying it blocks all integration is obsolete. |
| Former V40 target | **SUPERSEDED — issue #69.** Historical control record only; not a current delivery deadline or app-parent instruction. |

## Historical artifact correction

The four unsigned Actions artifacts that earlier control records described as live have now been rechecked through their originating workflow-run inventories. All four report `expired: true`:

| Artifact ID | Origin run | Current status |
|---|---:|---|
| `8854774165` | `30810944362` | expired |
| `8856536245` | `30815339549` | expired |
| `8863951645` | `30833597696` | expired |
| `8865678638` | `30838068198` | expired |

Issue #24 remains open because expiry alone does not prove manual-dispatch no-artifact behaviour, private-handoff controls, future release-automation gating or every stale-branch reconciliation requirement.

## Stale branch controls

### PR #71

PR #71 remains open/draft but belongs to the superseded V40 application direction. Issue #69 explicitly says the V40 shell is rejected as an application parent. Do not revive or merge PR #71 as the product baseline.

### PR #72

PR #72 remains open/draft. It has useful ownership primitives and hostile tests, but it predates current `main`, is not complete for issue #4, and must not be merged wholesale. Reuse small proven components only after reimplementation/requalification against current `main`.

### Issue #65

Issue #65 is closed. Its closure removes the old blanket M1.18 block, but it does not approve real evidence, public release, professional/high-consequence outputs or final model packaging.

## Current binary-release gate

A public executable remains blocked until one exact final candidate proves, at minimum:

- issue #3/#4 state and source-integrity requirements;
- issue #6/#7 usability, responsiveness and accessibility requirements;
- issue #8/#12 search/source navigation and Ask/restore consistency requirements where applicable;
- issue #14 privacy-safe diagnostics and final network/offline evidence;
- issue #15 final packaged-content SBOM/licences/provenance and corresponding source;
- issue #16/#20/#23/#46 intended-purpose, wording and prohibited-output controls;
- issue #17 accountable publisher/steward responsibility;
- issue #24 controlled distribution and release-automation requirements;
- exact clean-machine execution and Acer-baseline application qualification;
- version metadata, final installer/uninstaller where used, rollback/uninstall behaviour and long-duration stability;
- trusted Authenticode signing of the complete final artefact with no post-signing mutation;
- explicit release approval for the exact final SHA-256.

CI Cosign rehearsal does not substitute for Authenticode publisher identity.

## Current source-development sequence

1. Keep canonical status/release records current.
2. Implement **workspace ownership and clean-state control (#4)** on current `main`, selectively reusing qualified PR #72 primitives.
3. Close remaining **preserved-source and Ask/restore consistency (#3/#12)** gaps.
4. Implement **page-aware search/highlighting (#8)** on the qualified PDF/OCR/extraction foundation.
5. Qualify **responsiveness and accessibility (#6/#7)**, including keyboard-only, high-DPI and Narrator/NVDA evidence.
6. Close **diagnostic privacy/offline claims (#14)** and cross-surface intended-purpose/claims gates (#16/#20/#23/#46).
7. Complete **final release packaging/provenance/signing/publisher controls (#15/#17/#24)**.
8. Treat local generative AI (#5) as a bounded optional feature; do not let model packaging delay the deterministic evidence product.

## Public-record rule

Use synthetic and non-sensitive information only. Do not publish personal evidence, private workspaces, credentials, private diagnostics, exploit-level instructions, unapproved executable/model files or sensitive case material in repository records.
