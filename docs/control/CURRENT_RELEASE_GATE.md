# Current ECO release gate

**Gate record:** `ECO-RELEASE-GATE-20260803-005`  
**Updated:** 3 August 2026  
**Canonical public status:** [`../../CURRENT_STATUS.md`](../../CURRENT_STATUS.md)  
**Baseline `main` commit reviewed before this intended-purpose update:** `aac8f6f3d5d9314d765cc14546d66ec7fc376aec`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Signed end-user release:** None

## Controlling release decision

ECO remains a source-development project only.

The following remain blocked:

- use with real, sensitive or irreplaceable evidence;
- ordinary-user binary distribution;
- public preview, release-candidate or stable-release status;
- public-sector or institutional deployment;
- healthcare or clinical deployment;
- EU availability;
- claims of legal, forensic, medical, accessibility or regulatory compliance.

## Current gate status

| Gate | Current position |
|---|---|
| Source identity | `main` contains post-V25 development changes; no later named source milestone is approved |
| Evidence preservation | PR #10 merged materially improved controls, but issue #3 is reopened pending issue #12 and independent closure |
| Ask verification and restore concurrency | Blocked by open issue #12 |
| Ask health-related input boundary | Governance boundary defined; technical implementation conformance remains blocked by P0 issue #20 |
| Candidate workspace identity | PR #11 contains material candidate-binding improvements but remains draft and blocked by current ownership/concurrency findings |
| Workspace creation and first launch | Blocked by PR #11 exact-head finding: creation and candidate state are not yet protected by one alias-safe cross-process ownership transaction |
| Ordinary workspace writers | Blocked by PR #11 exact-head finding: stale independently opened writers can still erase newer metadata without revision/CAS conflict detection |
| Alias and retained-parent ownership | Blocked by PR #11 exact-head finding: lock, recovery controls and workspace participants are not all derived from one retained object identity |
| Linux nested cleanup | Blocked by PR #11 exact-head finding: an inspected directory can be closed and reopened by name during recursive cleanup |
| Diagnostic privacy and offline claims | Blocked by issue #14 |
| Runtime, model, SBOM and licensing | Blocked by issue #15 until the exact packaged artefact is reconciled |
| One-file packaging | Blocked until the actual final embedded executable is supplied and independently inspected |
| Signing order and authenticity | Blocked until the final file is assembled, hashed and Authenticode-signed with no later mutation |
| OCR and local AI | Not approved as reliable production functionality |
| Intended purpose and public claims | Governance boundary defined; issue #16 remains open for cross-surface evidence and issue #20 remains the P0 implementation dependency |
| Accessibility | Blocked pending issue #7 evidence and truthful conformance documentation |
| Security reporting, publisher and continuity | Blocked by issue #17 |
| Public claims | Only controlled development claims consistent with the intended-purpose boundary are permitted |

## Issue and pull-request controls

### Issue #3 and issue #12

Issue #3 is open. The implementation merged through PR #10 is not rejected, but it must not be described as independently closed while:

- issue #12 remains open;
- the issue #3 acceptance checklist is not completed with evidence references;
- exact-commit hostile and concurrency testing remains incomplete.

Issue #12 requires bounded verification cost, safe cache invalidation or source-limited verification, and safe serialisation of Ask against restore.

### PR #11 and issue #4

PR #11 remains draft and unmerged at the last independently inspected head `73689717bb08bb8cec0fc1233b92f843b449484a`.

The current exact-head blockers are:

1. stale concurrent writers without exact revision/CAS conflict protection;
2. alias-bypassable path-derived locking and split pathname resolution instead of one retained-parent transaction;
3. Linux nested cleanup reopening an inspected directory by name;
4. unowned workspace creation, first launch and candidate-state writes.

The fail-fast Windows CI correction, object-bound reset/migration operations, authenticated restore phases and issue #3 regression coverage at that head are material progress, but they do not close these four blockers.

PR #11 must not be marked ready, merged or used to close issue #4 until one coherent ownership/concurrency correction is pushed, current `main` is reconciled, raw Windows/Linux evidence is inspected and the merged exact SHA is independently verified.

### Intended-purpose control, issue #16 and issue #20

The [ECO intended purpose and public-claims control](../governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md) is the project-governance boundary when present on the canonical branch.

It operates separately from:

- implementation conformance by an exact application build;
- medical-device or other legal and regulatory classification;
- release or deployment approval.

Adoption of that document does not establish those matters or close issue #16 or P0 issue #20.

The implementation observation is tied to application source at commit `001e4fc2d55f8c29a9eaa20a0f741162433aadad`. Later documentation commits do not establish that the Ask implementation changed.

P0 issue #20 blocks Ask ECO, a local model or another semantic/generated-answer route from operating on a vault containing health-related or mixed-purpose clinical material unless independently verified isolation proves that restricted material cannot enter ranking, prompts, context, memory, citations or output. Question scope to an unrelated file or topic is not sufficient isolation.

Issue #16 remains open until README, application UI, model or system cards, exports, release material, screenshots, partner material and prohibited-claim controls are reconciled and independently evidenced.

Before any official EU availability, the accountable organisation must perform a feature-level and role-specific assessment under Regulation (EU) 2024/1689 as amended by Regulation (EU) 2026/1744 and the European Commission's final Article 50 guidelines of 20 July 2026. No EU AI Act compliance claim is approved.

### Issues #14, #15 and #17

- Issue #14 blocks unsafe diagnostic sharing and inaccurate offline/network claims.
- Issue #15 blocks distribution without actual-build SBOM, licensing, provenance and signing evidence.
- Issue #17 blocks public or institutional release without an accountable publisher, operational response routes and continuity ownership.

## Public-document assurance result

The 3 August 2026 public-document assurance reconciliation is complete. The README, current status, release gate, known limitations, release policy, roadmap, building guide, privacy policy, security policy, threat model and maintainer-role record now describe the blocked development position consistently.

That reconciliation changed documentation and control wording only and did not change source behaviour, workflows, `VERSION`, binaries, models, the historical V25 SBOM, licences or third-party notices. It did not approve a build or resolve a technical gate.

The following underlying limits remain release-blocking:

- current `main` still lacks explicit native exit-code enforcement in the Windows PowerShell build gate;
- current `main` portable restore still uses pathname-based activation and best-effort rollback;
- PR #11 still has the four exact-head ownership/concurrency blockers above and is currently non-mergeable pending reconciliation;
- issue #12 still blocks Ask verification and restore concurrency;
- issue #14 still blocks diagnostic and final network qualification;
- issue #15 still requires actual packaged-artifact SBOM, licence and provenance reconciliation;
- issues #16, #17 and #20 remain open;
- accessibility, responsiveness and page-aware search qualification remains incomplete.

## Stop rules

### Stop before real-evidence testing

Stop where any of the following remains unresolved:

- preserved bytes, recorded hashes and derived readings cannot be reconciled exactly;
- Ask, restore, migration, creation or reset can observe or create mixed or unsafe state;
- health-related or mixed-purpose clinical material can enter semantic Ask or generated-answer processing without the issue #20 gate;
- filesystem boundaries are not proven against Windows reparse points, junctions, links, aliases and parent substitution;
- diagnostic or runtime records may expose case content without a safe redacted mode;
- endpoint, runtime or local IPC risks remain unqualified;
- any real-evidence P0 or P1 finding remains open.

### Stop before public preview

Stop where:

- the executable is unsigned or is not the actual one-file deliverable;
- accessibility evidence is incomplete;
- application help or AI can invent controls, actions or unsupported conclusions;
- the intended-purpose boundary and current implementation are not aligned;
- no accountable steward accepts security, privacy, complaints, support and continuity duties;
- public claims cannot be supported by the exact released artefact.

### Stop before GitHub Release, official binary or institutional supply

Stop where:

- exact source, immutable commit, build receipt, manifest, SBOM or licence notices are absent;
- final model/runtime provenance is incomplete;
- Smart App Control and clean-machine testing have not passed;
- the file was changed after signing;
- contracting, data-role, support, liability or exit arrangements required for institutional supply are unresolved;
- any P0 or P1 finding remains unresolved.

A finding must not be labelled non-release-blocking or accepted as an exception to bypass this stop rule. The unresolved-P0/P1 stop applies to every official ECO distribution and project-endorsed institutional supply, not only a public GitHub Release.

## Public-record rule

Public status documents may identify defect classes, acceptance tests and release decisions. They must not expose personal evidence, private diagnostics, unpublished binaries, model files, private workspaces or exploit-level instructions.
