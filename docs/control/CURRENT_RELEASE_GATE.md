# Current ECO release gate

**Gate record:** `ECO-RELEASE-GATE-20260803-006`  
**Updated:** 3 August 2026, after final issue #27 merged-main verification  
**Canonical public status:** [`../../CURRENT_STATUS.md`](../../CURRENT_STATUS.md)  
**Current `main` commit reviewed:** `9b7f3d60b14ff67fbf9dc4e0047ceeb498725e79`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Signed end-user release:** None

## Controlling release decision

ECO remains a source-development project only.

The following remain blocked:

- use with real, sensitive or irreplaceable evidence;
- ordinary-user binary distribution through GitHub Releases, Actions artifacts or another public channel;
- use or redistribution of historical unsigned Actions executables;
- public preview, release-candidate or stable-release status;
- public-sector or institutional deployment;
- healthcare or clinical deployment;
- EU availability;
- claims of legal, forensic, medical, accessibility or regulatory compliance.

## Current gate status

| Gate | Current position |
|---|---|
| Source identity | `main` contains post-V25 development and repository-control changes; no later named source milestone is approved |
| Windows CI truthfulness | Issue #27 technical acceptance complete on `main` after commits `176bf4a5...` and `9b7f3d60...`; final merged-main run `30852039542` passed the controlled failure self-test, six-stage failure matrix and ordinary Linux/Windows validation with zero artifacts |
| Evidence preservation | PR #10 merged materially improved controls, but issue #3 remains open pending issue #12 and independent closure |
| Ask verification and restore concurrency | Blocked by open issue #12 |
| Ask health-related input boundary | Blocked by P0 issue #20; current Ask source does not yet enforce the proposed restriction |
| Candidate workspace identity | PR #11 contains material candidate-binding improvements but remains draft and blocked by current ownership/concurrency findings |
| Workspace creation and first launch | Blocked by PR #11 exact-head finding: creation and candidate state are not yet protected by one alias-safe cross-process ownership transaction |
| Ordinary workspace writers | Blocked by PR #11 exact-head finding: stale independently opened writers can still erase newer metadata without revision/CAS conflict detection |
| Alias and retained-parent ownership | Blocked by PR #11 exact-head finding: lock, recovery controls and workspace participants are not all derived from one retained object identity |
| Linux nested cleanup | Blocked by PR #11 exact-head finding: an inspected directory can be closed and reopened by name during recursive cleanup |
| Public Actions executable distribution | New uploads stopped in `bdc05df444d21d739abf83fa9cf768fc4ab5dd9a`; corrected pull-request runs prove zero artifacts while internal builds continue; issue #24 remains open for remaining trigger, historical cleanup and stale-branch evidence |
| Diagnostic privacy and offline claims | Blocked by issue #14 |
| Runtime, model, SBOM and licensing | Blocked by issue #15 until the exact packaged artefact is reconciled |
| One-file packaging | Blocked until the actual final embedded executable is supplied and independently inspected |
| Signing order and authenticity | Blocked until the final file is assembled, hashed and Authenticode-signed with no later mutation |
| OCR and local AI | Not approved as reliable production functionality |
| Intended purpose and public claims | Blocked by draft PR #18, issue #16 and the issue #20 implementation dependency |
| Accessibility | Blocked pending issue #7 evidence and truthful conformance documentation |
| Security reporting, publisher and continuity | Blocked by issue #17 |
| Public claims | Only controlled development claims are permitted |

## Issue #27 — Windows native-command failure handling

The former Windows build gate could continue after a failed native command and still report success.

The correction now on `main`:

- routes tests, vet, source policy, both deterministic builds and `go version` through one checked helper;
- captures `$LASTEXITCODE` immediately and terminates on non-zero native exits;
- runs a controlled exit-23 self-test;
- runs a six-stage failure matrix proving that failure at each native stage prevents later work;
- statically verifies that every real native stage is wrapped exactly once and legacy unwrapped command lines are absent;
- corrects the three Windows unsafe-pointer vet findings rather than suppressing them;
- preserves issue #24's no-runnable-artifact workflow.

Final merged-main verification used PR #32 and workflow run `30852039542`. Raw Linux, source-policy and Windows logs passed. The run artifact inventory was exactly empty. The runner-only unsigned executable was not uploaded and is not approved for use or release.

Issue #27 technical acceptance is complete. Closure follows this public-document reconciliation. This result does not open any product or release gate.

## PR #11 and issue #4

PR #11 remains draft and unmerged at the last independently inspected head `73689717bb08bb8cec0fc1233b92f843b449484a`.

The current exact-head blockers are:

1. stale concurrent writers without exact revision/CAS conflict protection;
2. alias-bypassable path-derived locking and split pathname resolution instead of one retained-parent transaction;
3. Linux nested cleanup reopening an inspected directory by name;
4. unowned workspace creation, first launch and candidate-state writes.

PR #11 must not be marked ready, merged or used to close issue #4 until one coherent ownership/concurrency correction is pushed, current `main` is reconciled, raw Windows/Linux evidence is inspected and the merged exact SHA is independently verified.

## Issue #24 — public Actions executable distribution

Independent inspection confirmed that the former public workflow uploaded unsigned runnable `ECO.exe` packages after successful implementation and documentation-only pull-request runs.

Emergency source containment at `bdc05df444d21d739abf83fa9cf768fc4ab5dd9a` removed the `upload-artifact` step while retaining internal Windows compilation and testing.

Corrected validation runs now prove that:

- internal Windows compilation and testing still execute;
- the controlled native-failure self-test and six-stage failure matrix pass;
- successful pull-request runs expose no executable, DLL, installer, model/runtime payload or runnable archive.

Issue #24 remains open until:

- successful push and manual-dispatch paths are evidenced through an independently accessible route;
- any remaining deliberate failure-path cases required by issue #24 are evidenced;
- the four identified historical executable artifacts are deleted by an authorised repository administrator or have expired;
- the private controlled test-handoff process is evidenced;
- affected branches are reconciled with the corrected workflow;
- no later automation reintroduces public binary upload without issue #15, signing and publisher approval.

The identified historical and runner-only executables are unapproved and must not be tested, used or redistributed.

## Issues #3, #12, #14–#17 and #20

- Issue #3 remains open pending issue #12 and complete independent preservation evidence.
- Issue #12 requires bounded verification cost, safe cache invalidation or source-limited verification, and safe serialisation of Ask against restore.
- Issue #14 blocks unsafe diagnostic sharing and inaccurate offline/network claims.
- Issue #15 blocks distribution without actual-build SBOM, licensing, provenance and signing evidence.
- Issue #16 blocks unsupported intended-purpose, legal, medical, forensic, high-risk-decision and compliance claims.
- Issue #17 blocks public or institutional release without an accountable publisher, operational response routes and continuity ownership.
- P0 issue #20 blocks generated processing of health-related or mixed-purpose clinical material before a tested input gate and safe fallback exist.

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

- any public Actions artifact or other channel exposes an unapproved runnable payload;
- the executable is unsigned or is not the actual one-file deliverable;
- accessibility evidence is incomplete;
- application help or AI can invent controls, actions or unsupported conclusions;
- the intended-purpose boundary and current implementation are not aligned;
- no accountable steward accepts security, privacy, complaints, support and continuity duties;
- public claims cannot be supported by the exact released artefact.

### Stop before GitHub Release or public Actions upload

Stop where:

- exact source, immutable commit, build receipt, manifest, SBOM or licence notices are absent;
- final model/runtime provenance is incomplete;
- Smart App Control and clean-machine testing have not passed;
- the file was changed after signing;
- issue #24's distribution controls have not passed;
- any P0 or P1 finding remains unresolved.

Calling a file a test, provenance or temporary artifact does not remove these stop rules. A finding must not be labelled non-release-blocking or accepted as an exception to bypass them.

## Public-record rule

Public status documents may identify defect classes, acceptance tests and release decisions. They must not expose personal evidence, private diagnostics, unapproved binaries, model files, private workspaces or exploit-level instructions.
