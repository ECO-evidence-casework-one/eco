# Current ECO project status

**Status date:** 4 August 2026  
**Latest control update:** 4 August 2026, 06:39 BST  
**Canonical public status record:** this file  
**Current reviewed `main` commit before this documentation reconciliation:** `4c6929b6b94e4a428ccc69336b50a38ba9fe6271`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Release position:** development only; no approved public binary

## Status authority

This root `CURRENT_STATUS.md` file is the single current public status authority.

The operational control board in issue #22 tracks exact active pull-request heads, independent inspection decisions, owners and next actions. It does not replace this canonical public status record.

Files under `docs/status/` with dates are historical daily records. The former `docs/status/CURRENT_STATUS.md` path is retained only as a pointer to this canonical record.

## What `main` represents

`VERSION` still records the last named source milestone approved under the earlier milestone process: `ECO-V25-20260731-N2-P1`.

The `main` branch also contains later controlled source-development, security and repository-control changes. Those changes do **not** create a new approved source milestone, release candidate or end-user release merely because they were committed.

No private model bundle, diagnostic archive, screenshot set, private workspace or personal evidence is represented by this repository.

## Current controlled development position

### Evidence preservation

PR #10 merged source changes intended to make the verified preserved object the source of truth for extraction, OCR, citation and retrieval.

Issue #3 remains open. Its implementation is materially improved, but independent closure is pending because follow-up work is recorded in issue #12. Until that work and the complete acceptance evidence are independently verified, use with real, sensitive or irreplaceable evidence remains blocked.

### Workspace identity, migration and reset

PR #11 remains draft, unmerged, non-mergeable and independently blocked.

- Last independently reviewed workspace implementation head: `73689717bb08bb8cec0fc1233b92f843b449484a`.
- Current PR head: `61a2004809b341e72f70321843c64c3ff477f549`.
- The commits after the reviewed workspace implementation change only `.github/workflows/ci.yml` to stop public executable upload; they do not correct workspace ownership or concurrency.

Its four current P0 blockers are:

1. stale independently opened writers can overwrite newer workspace metadata without exact revision/CAS conflict detection;
2. path-derived locking can be split by aliases and separate pathname resolution instead of one retained-parent transaction;
3. Linux nested cleanup can close an inspected directory and reopen it by name;
4. workspace creation, first launch and candidate-state writes are not protected by one alias-safe cross-process ownership transaction.

Issue #4 remains open. The branch must be reconciled with current `main` `4c6929b6b94e4a428ccc69336b50a38ba9fe6271` without losing the no-public-artifact workflow, verified fail-fast controls, Windows pointer-safety corrections or the current historical-provenance controls.

### Windows CI truthfulness — issue #27 closed

The former Windows false-green defect was reproduced: a non-zero native `go vet` result could be followed by policy and build steps while the job reported success.

PR #28 isolated and merged the fail-fast correction and the three resulting Windows unsafe-pointer corrections. PR #31 then added a six-stage failure matrix and static assertions for every real native stage in `scripts/build-windows.ps1`.

Final merged issue #27 control commit:

`9b7f3d60b14ff67fbf9dc4e0047ceeb498725e79`

Final independent merged-tree verification used run `30852039542` and proved:

- Linux tests and vet passed;
- source policy passed;
- a controlled non-zero native command terminated the Windows gate;
- separate injected failures at test, vet, source policy, first build, second build and `go version` stopped all later logical stages;
- every real native stage remained wrapped by `Invoke-NativeChecked`;
- ordinary Windows tests, vet, policy and deterministic rebuild passed;
- the run published exactly zero Actions artifacts.

Issue #27 closed as completed on 3 August 2026. This establishes CI truthfulness for the covered commands only; it does not approve the unsigned runner-built executable or open a release gate.

### Public Actions executable containment — issue #24 remains open

Independent inspection found that the public GitHub Actions workflow uploaded unsigned `ECO.exe` packages for successful pull-request and `main` runs, including documentation-only changes. Four unexpired runnable artifacts were identified and recorded in P0 issue #24.

Emergency containment commit `bdc05df444d21d739abf83fa9cf768fc4ab5dd9a` removed the public `upload-artifact` step while preserving internal Windows compilation and testing on the hosted runner.

Multiple independently inspected pull-request routes—including failed native validation, successful builds, affected-branch containment checks and final merged-main verification—have since produced exactly zero Actions artifacts.

Issue #24 remains open because:

- the four historical executable artifacts must be deleted by an authorised repository administrator or expire;
- a controlled manual-dispatch run still requires explicit no-artifact evidence;
- affected development branches must preserve the corrected no-upload workflow when reconciled;
- no later automation may reintroduce public runnable uploads before signing, provenance and publisher gates pass.

No listed Actions executable is approved for testing, use or redistribution.

### Actual-build provenance — issue #15 remains open

The generic root repository manifest and preparation receipt are now explicit pointers to preserved historical V25 records. The root SPDX document is explicitly labelled as a historical V25 source-level SBOM, and the exact former bytes are preserved under a versioned historical path. Root third-party notices now describe current source-level declarations only.

These corrections prevent the historical files from being mistaken for current release evidence. They do not create an authoritative exact-build manifest, executable SBOM, final notice bundle, signing record or release receipt. P0 issue #15 remains open.

### Intended-purpose and health-input conformance

PR #18 proposes a controlling intended-purpose and public-claims boundary. P0 issue #20 records the implementation gate because the current Ask ECO path does not yet enforce the proposed restriction for health-related or mixed-purpose clinical material.

Until issue #20 is independently closed, health-related material must not enter Ask ECO or another generated-answer route except synthetic tests designed to prove safe rejection.

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
- Issue #24: historical executable containment, manual-dispatch proof and branch reconciliation.

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

## Underlying technical and organisational limits

- Current `main` portable restore still uses pathname-based activation and best-effort rollback.
- PR #11 retains four workspace ownership/concurrency P0 blockers.
- Issue #12 blocks safe Ask verification cost and Ask/restore serialisation.
- Issue #14 blocks privacy-safe diagnostics and final bundled-runtime network claims.
- There is no authoritative current exact-build manifest, executable SBOM, final notice bundle, signed receipt or actual packaged-artifact reconciliation required by issue #15.
- Issue #24 remains open despite successful no-artifact PR evidence.
- PR #18 and issues #16 and #20 remain unresolved.
- No accountable established publisher, support operator, complaints handler or continuity owner has accepted the issue #17 duties.
- Accessibility, responsiveness and page-aware search qualification remains incomplete.

## Release prerequisites

Before the release position can be reconsidered, ECO requires objective evidence for at least:

- exact source, build and binary identity;
- complete corresponding source, manifest, SBOM and licence notices;
- verified model and runtime provenance;
- genuine one-file packaging and clean-machine execution;
- Windows stability, low-resource and long-duration testing;
- no-external-network proof for all bundled runtime components;
- evidence-integrity and recovery qualification;
- privacy-safe diagnostics and accurate public network claims;
- keyboard, assistive-technology, DPI and cognitive-accessibility evidence;
- trusted Authenticode signing with no post-signing file mutation;
- an authoritative intended-purpose and excluded-use boundary;
- technical enforcement of the health-related generated-processing boundary recorded in issue #20;
- an accountable publisher or steward for security, privacy, complaints, support and continuity;
- a controlled distribution pipeline that exposes no runnable artifact before approval;
- independent closure of every release-blocking P0 and P1 finding.

## Public-record rule

Public updates must remain sanitised. They may record defect classes, controls, acceptance tests and release decisions, but must not publish personal evidence, private diagnostics, exploit-level instructions, confidential screenshots, unapproved binaries, bundled models or private test workspaces.

Use synthetic and non-sensitive information only in this repository.
