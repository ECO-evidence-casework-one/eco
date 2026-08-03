# Current ECO project status

**Status date:** 3 August 2026  
**Canonical public status record:** this file  
**Baseline `main` commit reviewed before this status update:** `4c130a915d2fd9f4e20274dc4e29bddeb8fb472d`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Release position:** development only; no approved public binary

## Status authority

This root `CURRENT_STATUS.md` file is the single current public status authority.

The operational control board in issue #22 tracks exact active pull-request heads, independent inspection decisions, owners and next actions. It does not replace this canonical public status record.

Files under `docs/status/` with dates are historical daily records. The former `docs/status/CURRENT_STATUS.md` path is retained only as a pointer to this canonical record.

## What `main` represents

`VERSION` still records the last named source milestone approved under the earlier milestone process: `ECO-V25-20260731-N2-P1`.

The `main` branch now also contains later controlled source-development changes. Those changes do **not** create a new approved source milestone, release candidate or end-user release merely because they were merged.

No later private executable, model bundle, diagnostic archive, screenshot set, private workspace or personal evidence is represented by this repository.

## Current controlled development position

### Evidence preservation

PR #10 merged source changes intended to make the verified preserved object the source of truth for extraction, OCR, citation and retrieval.

Issue #3 has been reopened. Its implementation is materially improved, but independent closure is pending because post-merge review identified follow-up work recorded in issue #12. Until that work and the complete acceptance evidence are independently verified, issue #3 remains open and use with real, sensitive or irreplaceable evidence remains blocked.

### Workspace identity, migration and reset

PR #11 remains a draft. It proposes clean candidate-specific workspace state, migration, recovery and reset controls, but independent review identified unresolved P0 ownership, concurrency and filesystem-object boundaries. It must not be treated as approved or merged until those findings are corrected and re-reviewed.

The current independently inspected PR head is `73689717bb08bb8cec0fc1233b92f843b449484a`. The exact operational blockers and acceptance evidence are recorded in issue #22 and on PR #11. Issue #4 remains open.

### Intended-purpose and health-input conformance

PR #18 proposes a controlling intended-purpose and public-claims boundary. Source review found that the current Ask ECO path can rank, truncate, reorder and compose passages without first applying the proposed restriction for health-related source material or health-related or clinical content within mixed-purpose material.

P0 issue #20 records the implementation gate. Until it is independently closed, the current source must not be described as enforcing that boundary, and health-related material must not enter Ask ECO or another generated-answer route except in synthetic tests designed to prove safe rejection.

### Other active grouped work

- Issue #5: instruction-faithful, source-backed and non-operative offline AI.
- Issue #6: responsive background import, OCR and local-AI work.
- Issue #7: keyboard, screen-reader, DPI, scrolling and working-control accessibility.
- Issue #8: page-aware document search, visible highlights and source navigation.
- Issue #12: bounded Ask verification cost and safe serialisation against restore.
- Issue #14: privacy-safe diagnostics and accurate offline/network claims.
- Issue #15: actual-build SBOM, licence notices and release provenance.
- Issue #16: intended purpose, excluded uses and controlled public claims.
- Issue #17: accountable publisher, response routes and project continuity.
- Issue #20: gate Ask and generated processing for health-related inputs.

## Current stop gates

The following remain blocked:

- use with real, sensitive or irreplaceable evidence;
- public end-user binary distribution;
- release-candidate or stable-release status;
- claims of reliable generative offline AI assistance;
- claims that the intended-purpose health-input boundary is implemented;
- claims of completed production OCR or complete native PDF investigation;
- claims of accessibility, forensic, legal, medical or regulatory compliance;
- all public-sector or private-institutional deployment;
- all healthcare or clinical deployment;
- all EU availability, including download, supply or deployment.

## Current public-document assurance note

A 3 August 2026 independent documentation audit found that the canonical status and release stop rules remain protective, but several supporting documents need reconciliation before any release decision:

- `docs/control/CURRENT_RELEASE_GATE.md` still describes older PR #11 findings rather than the current exact-head ownership/concurrency blockers;
- `RELEASE_POLICY.md` still names only issues #3–#8 in its dated current-gate section and omits later release-blocking issues #12, #14–#17 and #20;
- `KNOWN_LIMITATIONS.md` remains narrowly scoped to the recorded V25 N2 P1 milestone and does not summarise the later active workspace, Ask, diagnostic, governance and health-input gates;
- `THREAT_MODEL.md` uses “staged transactional restore” as a current core-control phrase even though `main` still uses pathname-based activation and best-effort rollback and PR #11 remains blocked;
- `BUILDING.md` describes the Windows script as running tests and vet as a controlled gate, but the fail-fast native-command correction is not yet on `main`;
- `ROADMAP.md` needs to distinguish existing bounded Office/email text extraction from the richer structured and receipt-bearing extraction still required;
- maintainer and signing-role documents need an explicit statement that repository roles do not appoint the named individual as legal publisher, supplier, support operator, complaints handler, data controller or liability owner;
- the root V25 SBOM and notices are historical/source-level records, not actual packaged-binary provenance under issue #15.

These documentation findings do not open any gate. They must be reconciled through a controlled documentation change without weakening the current stop rules or restoring stale milestone wording.

## Release prerequisites

Before the release position can be reconsidered, ECO requires objective evidence for at least:

- exact source, build and binary identity;
- complete corresponding source, manifest, SBOM and licence notices;
- verified model and runtime provenance;
- genuine one-file packaging and clean-machine execution;
- Windows stability, low-resource and long-duration testing;
- no external-network proof for all bundled runtime components;
- evidence-integrity and recovery qualification;
- privacy-safe diagnostics and accurate public network claims;
- keyboard, assistive-technology, DPI and cognitive-accessibility evidence;
- trusted Authenticode signing with no post-signing file mutation;
- an authoritative intended-purpose and excluded-use boundary;
- technical enforcement of the health-related generated-processing boundary recorded in issue #20;
- an accountable publisher or steward for security, privacy, complaints, support and continuity;
- independent closure of every P0 and P1 finding.

## Public-record rule

Public updates must remain sanitised. They may record defect classes, controls, acceptance tests and release decisions, but must not publish personal evidence, private diagnostics, exploit-level instructions, confidential screenshots, unpublished binaries, bundled models or private test workspaces.

Use synthetic and non-sensitive information only in this repository.
