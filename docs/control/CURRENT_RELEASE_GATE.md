# Current ECO release gate

**Gate record:** `ECO-RELEASE-GATE-20260804-008`  
**Updated:** 4 August 2026, 06:58 BST  
**Canonical public status:** [`../../CURRENT_STATUS.md`](../../CURRENT_STATUS.md)  
**Current reviewed `main` commit before this gate correction:** `c36753f0fff3daef3b94d1e3ae11fd25bea8933e`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Signed end-user release:** None

## Controlling release decision

ECO remains a source-development project only.

The following remain blocked:

- use with real, sensitive or irreplaceable evidence;
- ordinary-user binary distribution through GitHub Releases, Actions artifacts or another public channel;
- use or redistribution of historical unsigned Actions executables;
- public preview, release-candidate or stable-release status;
- public promotion of AI-assisted evidence functions before exact implementation qualification;
- diagnosis, treatment, clinical-risk, professional-representation, profiling, scoring or authority-side decision outputs;
- public-sector or institutional deployment;
- healthcare or clinical deployment;
- EU availability;
- claims of legal, forensic, medical, accessibility or regulatory compliance.

Issue #27 is closed and the covered Windows native-command CI gate is independently trustworthy. That closure does not alter any block above.

## Current gate status

| Gate | Current position |
|---|---|
| Source identity | `main` contains post-V25 development and repository-control changes; no later named source milestone is approved |
| Evidence preservation | PR #10 merged materially improved controls, but issue #3 remains open pending issue #12 and independent closure |
| Ask verification and restore concurrency | Blocked by open issue #12 |
| Evidence assistance with health/legal material | Permitted in principle under issue #20 when source-linked, clearly labelled and user-controlled; current real-evidence and public AI-assisted implementation remains unapproved |
| Clinical, professional and high-consequence outputs | Blocked by P0 issue #20 until exact deterministic and local-model routes prove safe boundary responses and no prohibited diagnosis, treatment, clinical-risk, reserved-legal, profiling, scoring or authority-side decision output |
| Candidate workspace identity | PR #11 contains material candidate-binding improvements but remains draft and blocked by current ownership/concurrency findings |
| Workspace creation and first launch | Blocked by PR #11 finding: creation and candidate state are not protected by one alias-safe cross-process ownership transaction |
| Ordinary workspace writers | Blocked by PR #11 finding: stale independently opened writers can erase newer metadata without revision/CAS conflict detection |
| Alias and retained-parent ownership | Blocked by PR #11 finding: lock, recovery controls and workspace participants are not all derived from one retained object identity |
| Linux nested cleanup | Blocked by PR #11 finding: an inspected directory can be closed and reopened by name during recursive cleanup |
| Windows native-command CI truthfulness | Issue #27 closed; fail-fast helper, controlled failure test, six-stage matrix, static wrapper assertions and merged-tree validation passed |
| Public Actions executable distribution | New uploads stopped and affected open branches have exact-head zero-artifact evidence, but issue #24 remains open for historical artifacts and manual-dispatch proof |
| Diagnostic privacy and offline claims | Blocked by issue #14 |
| Runtime, model, SBOM and licensing | Blocked by issue #15; generic V25 provenance records are explicitly historical/source-level, but no current exact-build package exists |
| One-file packaging | Blocked until the actual final embedded executable is supplied and independently inspected |
| Signing order and authenticity | Blocked until the final file is assembled, hashed and Authenticode-signed with no later mutation |
| OCR and local AI | Not approved as reliable production functionality |
| Intended purpose and public claims | Blocked by issue #16 and corrected P0 issue #20; stale PR #18 is not controlling |
| Accessibility | Blocked pending issue #7 evidence and truthful conformance documentation |
| Security reporting, publisher and continuity | Blocked by issue #17 |
| Public claims | Only controlled development claims are permitted |

## Issue and pull-request controls

### Issue #3 and issue #12

Issue #3 is open. The implementation merged through PR #10 is not rejected, but it must not be described as independently closed while:

- issue #12 remains open;
- the issue #3 acceptance checklist is not completed with evidence references;
- exact-commit hostile and concurrency testing remains incomplete.

Issue #12 requires bounded verification cost, safe cache invalidation or source-limited verification, and safe serialisation of Ask against restore.

### PR #11 and issue #4

PR #11 remains draft, unmerged and non-mergeable.

- Last independently inspected workspace implementation head: `73689717bb08bb8cec0fc1233b92f843b449484a`.
- Current PR head: `61a2004809b341e72f70321843c64c3ff477f549`.
- The later branch commits alter only `.github/workflows/ci.yml` to contain public executable upload; they do not correct the workspace findings.

The current workspace blockers are:

1. stale concurrent writers without exact revision/CAS conflict protection;
2. alias-bypassable path-derived locking and split pathname resolution instead of one retained-parent transaction;
3. Linux nested cleanup reopening an inspected directory by name;
4. unowned workspace creation, first launch and candidate-state writes.

Object-bound reset/migration operations, authenticated restore phases and issue #3 regression coverage at the reviewed implementation head are material progress, but they do not close these four blockers.

PR #11 must reconcile current `main` `c36753f0fff3daef3b94d1e3ae11fd25bea8933e` and preserve:

- issue #24 no-upload controls;
- issue #27 fail-fast helper, controlled failure test and six-stage matrix;
- Windows pointer-safety corrections;
- current canonical status and release-gate records;
- the historical provenance pointer, SBOM and source-notice structure;
- corrected issue #20 wording that permits evidence assistance while targeting prohibited clinical, professional, profiling and high-consequence outputs.

It must not be marked ready, merged or used to close issue #4 until one coherent ownership/concurrency correction is pushed, raw Windows/Linux evidence is inspected and the merged exact SHA is independently verified.

### Issue #27 — Windows native-command CI truthfulness

Issue #27 closed as completed on 3 August 2026.

The evidence chain included:

- reproduction of the original false-green Windows vet result;
- a real vet failure stopping before policy and both builds;
- correction of the three exposed unsafe-pointer findings without suppressing vet;
- successful ordinary Linux and Windows validation;
- a six-stage synthetic failure matrix covering test, vet, policy, first build, second build and `go version`;
- static assertions that every real native stage remains wrapped by `Invoke-NativeChecked`;
- final issue #27 control merge to `main` at `9b7f3d60b14ff67fbf9dc4e0047ceeb498725e79`;
- final merged-tree verification run `30852039542`, with an empty artifact inventory.

This closure permits reliance on the covered CI command results. It does not approve the unsigned runner-built executable or satisfy any application, packaging, signing or release gate.

### Issue #24 — public Actions executable distribution

Independent inspection confirmed that the public workflow uploaded unsigned runnable `ECO.exe` packages after successful implementation and documentation-only pull-request runs.

Emergency source containment at `bdc05df444d21d739abf83fa9cf768fc4ab5dd9a` removed the `upload-artifact` step while retaining internal Windows compilation and testing.

Successful and failed pull-request routes, including exact-head containment checks for PRs #11, #18 and #19, have since exposed exactly zero Actions artifacts.

Issue #24 remains open until:

- the four identified historical executable artifacts are deleted by an authorised repository administrator or have expired;
- a controlled manual-dispatch run produces no downloadable executable, DLL, installer or runnable archive;
- no later automation reintroduces public binary upload without issue #15, signing and publisher approval.

The identified historical executables are unapproved and must not be tested, used or redistributed.

### Issue #15 — actual-build provenance

The generic root repository manifest and preparation receipt are explicit historical pointers. The root SPDX document is explicitly labelled as a historical V25 source-level SBOM, with the exact former bytes preserved under a versioned historical path. Root third-party notices describe current source-level declarations only.

These records cannot support a current executable, signing or release claim. Issue #15 remains open until one exact final executable is reconciled with its source tree, content manifest, actual-build SBOM, complete notices, corresponding source, build receipt and signing record.

### Issue #20 — permitted evidence assistance and prohibited outputs

Issue #20 supersedes the earlier blanket restriction on generated processing whenever health-related material is present. ECO may assist a user with supplied health-related, legal or other sensitive evidence through source-linked factual assistance and user-controlled drafting. Document subject matter alone is not a reason to disable assistance.

Permitted evidence assistance, subject to source binding, visible output status and meaningful user review, may include extraction, OCR, exact search, source navigation, document-stated facts, source-linked summaries and chronologies, and user-controlled questions, notes, letters and draft responses.

ECO must not be presented, configured or relied upon to:

- diagnose or infer an unstated diagnosis;
- recommend treatment, medication or a clinical course of action;
- perform prognosis, triage, monitoring, clinical-risk or emergency assessment;
- conduct reserved legal activities or claim professional legal representation;
- guarantee legal correctness, admissibility or outcome;
- profile or score eligibility, credibility, honesty, dangerousness or entitlement;
- grant, deny, reduce, revoke or reclaim benefits, housing, healthcare, employment, education or another essential service;
- materially influence an authority-side or institutional adverse decision without a fresh regulatory and deployment assessment.

A whole-vault or whole-document health-content ban must not substitute for testing these actual output boundaries. Real or sensitive evidence and public AI-assisted use remain blocked until the exact implementation passes issue #20's acceptance tests. Issue #16 remains open for the clean controlling intended-purpose record. Stale PR #18 is not controlling.

### Issues #14, #16 and #17

- Issue #14 blocks unsafe diagnostic sharing and inaccurate offline/network claims.
- Issue #16 blocks unsupported intended-purpose, legal, medical, forensic, high-risk-decision and compliance claims and requires one clean controlling record.
- Issue #17 blocks public or institutional release without an accountable publisher, operational response routes and continuity ownership.

## Underlying release-blocking limits

- Current `main` portable restore still uses pathname-based activation and best-effort rollback.
- PR #11 retains the four ownership/concurrency blockers above.
- Issue #12 still blocks Ask verification and restore concurrency.
- Issue #14 still blocks diagnostic and final network qualification.
- Issue #15 still requires actual packaged-artifact SBOM, licence and provenance reconciliation.
- Issue #20's targeted deterministic and generated-output boundary is not implemented and independently qualified for real or public use.
- Issue #24 still requires historical artifact removal/expiry and manual-dispatch evidence.
- Issues #16 and #17 remain open; PRs #18 and #19 remain drafts, and PR #18's stale blanket health-content approach is not controlling.
- Accessibility, responsiveness and page-aware search qualification remains incomplete.

## Stop rules

### Stop before real-evidence testing

Stop where any of the following remains unresolved:

- preserved bytes, recorded hashes and derived readings cannot be reconciled exactly;
- Ask, restore, migration, creation or reset can observe or create mixed or unsafe state;
- deterministic or generated routes can produce diagnosis, treatment, clinical-risk, reserved-legal, representation, profiling, scoring or authority-side adverse-decision outputs;
- supplied health or legal evidence cannot be handled with exact source attribution and visible separation of source text, OCR, generated suggestion, user confirmation and user note;
- filesystem boundaries are not proven against Windows reparse points, junctions, links, aliases and parent substitution;
- diagnostic or runtime records may expose case content without a safe redacted mode;
- endpoint, runtime or local IPC risks remain unqualified;
- any real-evidence P0 or P1 finding remains open.

### Stop before public preview

Stop where:

- any public Actions artifact or other channel exposes an unapproved runnable payload;
- the executable is unsigned or is not the actual one-file deliverable;
- accessibility evidence is incomplete;
- application help or AI can invent controls, actions, professional conclusions or unsupported high-consequence outputs;
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
- any release-blocking P0 or P1 finding remains unresolved.

Calling a file a test, provenance or temporary artifact does not remove these stop rules. A finding must not be labelled non-release-blocking or accepted as an exception to bypass them.

## Public-record rule

Public status documents may identify defect classes, acceptance tests and release decisions. They must not expose personal evidence, private diagnostics, unapproved binaries, model files, private workspaces or exploit-level instructions.
