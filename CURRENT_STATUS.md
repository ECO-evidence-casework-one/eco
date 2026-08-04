# Current ECO project status

**Status date:** 4 August 2026  
**Latest control update:** 4 August 2026, approximately 10:50 BST / 11:50 CEST  
**Canonical public status record:** this file  
**Current reviewed `main` before this reconciliation:** `0dc57720c1c3394b03342427bc9b4dca09c1f040`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Release position:** source development only; no approved public binary

## Status authority

This root file is ECO's current public status authority. The live Control Board in issue #22 carries operational detail, exact heads and dated decisions. Dated reports under `docs/control/progress/` and `docs/status/` are historical records and do not replace this file.

`main` contains post-V25 security, governance, provenance and repository-control work. Those commits do not create a new approved milestone, release candidate or end-user release.

## Application implementation lane

PR #11 is the sole active application implementation lane.

- Last inspected workspace implementation: `73689717bb08bb8cec0fc1233b92f843b449484a`.
- Current PR head: `61a2004809b341e72f70321843c64c3ff477f549`.
- Current relation to reviewed `main` `0dc57720...`: 8 commits ahead, 41 commits behind; common base `2dcb44ba8541ab7a319de6e1f14f016aafe2ac1b`.
- State: draft, unmerged and non-mergeable.

The branch commits after the inspected workspace implementation contain workflow upload containment only. They do not correct the workspace findings.

### Four issue #4 P0 blockers

1. Independently opened stale writers can overwrite newer metadata without exact revision/CAS conflict detection.
2. Path-derived locking and separate pathname resolution can split ownership across aliases instead of one retained-parent object identity.
3. Linux nested cleanup can close an inspected directory and reopen it by name.
4. Workspace creation, first launch and candidate-state writes are not protected by one alias-safe cross-process ownership transaction.

PR #11 must be reconstructed from current `main`, deliberately carrying forward only valid workspace work. It must preserve the no-public-artifact workflow, fail-fast and six-stage Windows controls, Windows pointer-safety fixes, current governance, provenance and audit records, and every evidence-use and release stop gate.

No merge, issue #4 closure or real-evidence test is authorised.

## Evidence preservation and Ask

PR #10 materially improved preserved-object source binding for extraction, OCR, citation and retrieval. Issue #3 remains open pending issue #12 and complete acceptance evidence.

Issue #12 still blocks:

- bounded verification cost;
- safe cache invalidation or source-limited verification;
- serialization of Ask against restore;
- persistence-error propagation and truthful rollback.

Use with real, sensitive or irreplaceable evidence remains blocked.

## Windows CI and public artifact control

Issue #27 is closed. The covered Windows native-command path now has:

- `Invoke-NativeChecked` fail-fast handling;
- a controlled non-zero native-command test;
- a six-stage failure matrix for tests, vet, source policy, first build, second build and `go version`;
- Windows unsafe-pointer corrections exposed by truthful vet.

This establishes CI truthfulness for the covered commands only. It does not approve the runner-built executable.

Issue #24 remains open. The workflow no longer uploads runnable artifacts, and inspected PR merge-tree and merged-tree routes have exposed zero artifacts. Four historical unsigned executable archives remain prohibited until administrator deletion or confirmed expiry. A one-time recheck is scheduled for 10 August 2026 at 19:00 Europe/London. Controlled manual-dispatch no-artifact evidence is still outstanding.

No historical or runner-built executable is approved for download, testing, use or redistribution.

## Provenance and packaging

Issue #15 remains P0 and open.

The root manifest, preparation receipt and SPDX record are explicitly historical V25 or source-level records. Exact former bytes are preserved under versioned historical paths. They cannot support a current executable or release claim.

ECO still lacks one reconciled final package containing:

- exact source commit and tree identity;
- clean-source manifest;
- final executable hash and PE inventory;
- actual-build SBOM;
- complete notices and corresponding source;
- reproducible build receipt;
- final model/runtime provenance;
- Authenticode signing record with no post-signing mutation.

The current build remains an unsigned private runner output, not an end-user deliverable.

## Intended purpose and prohibited outputs

The controlling intended-purpose and public-claims boundary is:

[`docs/governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md`](docs/governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md)

ECO may, in principle, provide source-linked, visibly labelled and user-controlled assistance with supplied health, legal, benefits, housing, employment and other sensitive evidence. That may include exact extraction, OCR, search, source navigation, document-stated facts, summaries, chronologies, questions, notes, letters and draft responses.

ECO must not be presented, configured or relied upon to:

- diagnose or infer an unstated diagnosis;
- recommend treatment, medication or clinical action;
- perform prognosis, triage, monitoring, clinical-risk or emergency assessment;
- conduct reserved legal activities or claim professional representation;
- guarantee legal correctness, admissibility or outcome;
- profile or score eligibility, credibility, honesty, dangerousness or entitlement;
- grant, deny, reduce, revoke or reclaim an essential service;
- materially influence an authority-side adverse decision without a fresh assessment and separate approval.

Issues #16, #20 and #46 remain open because governance adoption does not prove application conformance. Current Ask ECO and generated routes are not approved for real evidence or public AI-assisted use.

## Publisher, stewardship and partner controls

The controlling publisher/stewardship gate is:

[`docs/governance/PUBLISHER_AND_STEWARDSHIP_GATE.md`](docs/governance/PUBLISHER_AND_STEWARDSHIP_GATE.md)

No organisation or individual is appointed as publisher, supplier, support operator, complaints handler, controller or liability owner. Contribution or repository administration alone does not appoint those roles, while responsibilities arising from actual publishing, signing, contracting, data processing, claims or law remain preserved.

Issue #17 remains open until a named established organisation completes due diligence, formally accepts the role and proves the operational, security, support, complaints and continuity routes.

The generic public-safe partner pack is present under `docs/partnership/`. It names, researches, ranks and contacts no organisation. It authorises no outreach, relationship, pilot, contract, repository/signing transfer, release or deployment. Any named-organisation research or contact requires a separate issue, evidence and approval decision.

## Audit and progress controls

The controlling workflow-evidence and reviewer-relationship correction is:

[`docs/control/AUDIT_EVIDENCE_CLASSIFICATION_CORRECTION_2026-08-04.md`](docs/control/AUDIT_EVIDENCE_CLASSIFICATION_CORRECTION_2026-08-04.md)

PR workflows must be described as tested PR merge-tree evidence when raw logs show a synthetic `refs/pull/<number>/merge` checkout. `Exact-head` may be used only when the raw checkout proves that exact branch head was tested. Final squash-tree comparison and post-merge CI are separate evidence classes.

The progress standard and supplement require truthful reviewer relationships and verified downloadable audit packs after substantial control cycles. The private 07:42 BST inspection pack is recorded at ZIP SHA-256 `76763897af9a3b20b3e8c18c018d625d0542a1bca620f95173db0c1d48fbc7f3`.

## Current stop gates

The following remain blocked:

- real, sensitive or irreplaceable evidence;
- current Ask ECO or public AI-assisted evidence functions;
- diagnosis, treatment, clinical-risk, professional-representation, profiling, scoring and authority-side adverse-decision outputs;
- ordinary-user executable testing;
- public binary distribution through Releases, Actions or another channel;
- use or redistribution of historical unsigned executables;
- release-candidate or stable-release status;
- claims of reliable production OCR, local AI, native PDF investigation, accessibility, forensic, legal, medical, security or regulatory compliance;
- named outreach, partnership or publisher appointment without separate approval;
- public-sector, institutional, healthcare, justice-sector or EU supply and deployment.

## Other active release-blocking work

- Issue #5: instruction-faithful, source-backed and non-operative offline AI.
- Issue #6: responsive background import, OCR and local-AI work.
- Issue #7: keyboard, screen-reader, DPI, scrolling and cognitive accessibility.
- Issue #8: page-aware search, highlights and source navigation.
- Issue #14: privacy-safe diagnostics and final offline/network claims.
- Issue #15: actual-build SBOM, notices, provenance and signing.
- Issue #16: intended-purpose and public-claims conformance.
- Issue #17: accountable publisher, response routes and continuity.
- Issue #20: targeted clinical, professional, profiling and high-consequence output boundary.
- Issue #24: historical executable expiry/deletion and manual-dispatch proof.
- Issue #46: cross-surface implementation conformance with the governance boundary.

## Next controlled sequence

1. Keep PR #11 frozen until one coherent workspace-correction candidate is built from current `main`.
2. Inspect that candidate delta-first against all four issue #4 P0 boundaries and issue #3 regressions.
3. Complete issue #24 artifact expiry/deletion and manual-dispatch evidence.
4. Continue issue #15 exact-build provenance work without public executable upload.
5. Keep all real-evidence, AI, publisher, outreach, signing, release and deployment gates closed until their acceptance evidence passes.

## Public-record rule

Public repository records may state defect classes, requirements, tests and decisions. They must not contain personal evidence, private diagnostics, confidential screenshots, credentials, unapproved binaries, model files, private workspaces or exploit-level instructions. Use synthetic and non-sensitive information only.
