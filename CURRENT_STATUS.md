# Current ECO project status

**Status date:** 3 August 2026  
**Latest control update:** 3 August 2026, after final issue #27 merged-main verification  
**Canonical public status record:** this file  
**Current `main` commit reviewed:** `9b7f3d60b14ff67fbf9dc4e0047ceeb498725e79`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Release position:** development only; no approved public binary

## Status authority

This root `CURRENT_STATUS.md` file is the single current public status authority.

The operational control board in issue #22 tracks exact active pull-request heads, independent inspection decisions, owners and next actions. It does not replace this canonical public status record.

Files under `docs/status/` with dates are historical daily records. The former `docs/status/CURRENT_STATUS.md` path is retained only as a pointer to this canonical record.

## What `main` represents

`VERSION` still records the last named source milestone approved under the earlier milestone process: `ECO-V25-20260731-N2-P1`.

The `main` branch also contains later controlled source-development and repository-control changes. Those changes do **not** create a new approved source milestone, release candidate or end-user release merely because they were committed.

No private model bundle, diagnostic archive, screenshot set, private workspace or personal evidence is represented by this repository.

## Current controlled development position

### Evidence preservation

PR #10 merged source changes intended to make the verified preserved object the source of truth for extraction, OCR, citation and retrieval.

Issue #3 remains open. Its implementation is materially improved, but post-merge review identified follow-up work in issue #12. Until that work and the complete acceptance evidence are independently verified, use with real, sensitive or irreplaceable evidence remains blocked.

### Workspace identity, migration and reset

PR #11 remains draft and unmerged. Its last independently inspected head is `73689717bb08bb8cec0fc1233b92f843b449484a`.

The remaining P0 blockers are:

1. stale independently opened writers without exact revision/CAS conflict protection;
2. alias-bypassable path-derived locking and split pathname resolution instead of one retained-parent transaction;
3. Linux nested cleanup reopening an inspected directory by name;
4. unowned workspace creation, first launch and candidate-state writes.

Issue #4 remains open. PR #11 must reconcile current `main`, including the current no-artifact and truthful Windows CI controls, before any new run or merge decision.

### Windows CI truthfulness

Issue #27 recorded that the former Windows PowerShell gate could continue after a failed native command and still report success.

The correction merged in:

- `176bf4a51950c42f83456cc45f33e801dd994303` — checked native-command helper, controlled failure self-test and correction of three Windows unsafe-pointer vet findings;
- `9b7f3d60b14ff67fbf9dc4e0047ceeb498725e79` — six-stage native-failure matrix and static wrapper assertions.

Final merged-main verification used PR #32, head `3ef4730824337d9f4f4d5f4ee8aa71295c967b35`, tested merge ref `2873b7895c9c5435e0bd5687382affe1a821b4d1`, workflow run `30852039542`.

The raw evidence passed:

- source policy;
- Linux tests and vet;
- controlled native exit-23 self-test;
- six-stage failure matrix covering tests, vet, source policy, first build, second build and `go version`;
- Windows tests and vet;
- Windows source policy;
- two deterministic Windows builds;
- empty artifact inventory.

The runner-only unsigned build was 3,937,792 bytes with SHA-256 `1f0d1bc1bdde245562023fc4d7bea6c5bae18b83bb291922654c86520ef8b585`. It was not uploaded and is not approved for testing or release.

Issue #27 technical acceptance is complete. The issue remains open only until this public-document reconciliation is merged and the closure record is posted.

### Public Actions executable containment

P0 issue #24 found that the public GitHub Actions workflow uploaded unsigned `ECO.exe` packages after successful implementation and documentation-only pull-request runs.

Emergency containment commit `bdc05df444d21d739abf83fa9cf768fc4ab5dd9a` removed the public `upload-artifact` step while preserving internal Windows compilation and testing.

Corrected pull-request validation has now proved that internal Windows builds can execute while the artifact inventory remains empty. Issue #24 remains open for:

- successful push-run evidence through an independently accessible route;
- controlled manual-dispatch evidence;
- any remaining deliberate failure-path evidence required by the issue #24 matrix;
- deletion or expiry evidence for the four historical executable artifacts;
- private controlled test-handoff evidence;
- future public-release automation gates;
- reconciliation of PRs #11, #18 and #19 so none restores an upload path.

No historical or runner-only executable is approved for testing, use or redistribution.

### Intended-purpose and health-input conformance

PR #18 proposes a controlling intended-purpose and public-claims boundary. Source review found that the current Ask ECO path can rank, truncate, reorder and compose passages without first applying the proposed restriction for health-related source material or health-related or clinical content within mixed-purpose material.

P0 issue #20 records the implementation gate. Until independently closed, health-related material must not enter Ask ECO or another generated-answer route except in synthetic tests designed to prove safe rejection.

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
- Issue #24: complete public Actions distribution containment and historical cleanup.

## Current stop gates

The following remain blocked:

- use with real, sensitive or irreplaceable evidence;
- public end-user binary distribution through Releases, Actions artifacts or another public channel;
- use or redistribution of historical unsigned Actions executables;
- release-candidate or stable-release status;
- claims of reliable generative offline AI assistance;
- claims that the intended-purpose health-input boundary is implemented;
- claims of completed production OCR or complete native PDF investigation;
- claims of accessibility, forensic, legal, medical or regulatory compliance;
- all public-sector or private-institutional deployment;
- all healthcare or clinical deployment;
- all EU availability, including download, supply or deployment.

## Current underlying release blockers

- issue #3 and issue #12 remain open;
- current `main` portable restore still uses pathname-based activation with best-effort rollback;
- PR #11 remains blocked by the four ownership/concurrency findings above;
- issue #14 still blocks privacy-safe diagnostics and final bundled-runtime network claims;
- the root V25 SBOM and notices remain historical/source-level records rather than the actual packaged-artifact provenance required by issue #15;
- issue #24 remains open for remaining trigger, cleanup and stale-branch evidence;
- PR #18 and issues #16 and #20 remain unresolved;
- no accountable established publisher, support operator, complaints handler or continuity owner has accepted the issue #17 duties;
- accessibility, responsiveness and page-aware search qualification remain incomplete.

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
- a controlled distribution pipeline that exposes no runnable artifact before approval;
- independent closure of every P0 and P1 finding.

## Public-record rule

Public updates must remain sanitised. They may record defect classes, controls, acceptance tests and release decisions, but must not publish personal evidence, private diagnostics, exploit-level instructions, confidential screenshots, unapproved binaries, bundled models or private test workspaces.

Use synthetic and non-sensitive information only in this repository.
