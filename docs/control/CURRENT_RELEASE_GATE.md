# Current ECO release gate

**Gate record:** `ECO-RELEASE-GATE-20260804-010`  
**Updated:** 4 August 2026, 13:49 BST / 14:49 CEST  
**Canonical public status:** [`../../CURRENT_STATUS.md`](../../CURRENT_STATUS.md)  
**Canonical `main` at this checkpoint:** `90825496cda639ef6be8c052a8ad9f2ebe2f1464`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Approved public V40 source release:** none  
**Approved signed end-user executable:** none

## Controlling decision

ECO remains a source-development project only.

No current branch, workflow run, HTML prototype, runner-built executable, governance document, deadline or audit pack authorises:

- real, sensitive or irreplaceable evidence use;
- current Ask ECO or local-model use with public or real evidence;
- a public V40 pre-alpha tag or source release;
- ordinary-user executable testing;
- public executable distribution;
- signing, release-candidate or stable-release status;
- institutional, healthcare, justice-sector or EU supply or deployment.

## Current gate matrix

| Gate | Current position |
|---|---|
| Evidence preservation | Issue #3 open; issue #12 still blocks complete source/restore/Ask consistency |
| Workspace ownership | Issue #4 open; draft PR #72 is documentation-only and on inspection HOLD |
| First usable Matter journey | Draft PR #71 is a synthetic design prototype and on inspection HOLD |
| M1.18 integration | P0 issue #65 blocks every adapter/model/runtime/IPC/Ask ECO/evidence/persistence connection |
| Offline AI controller | Issue #5 open |
| Responsiveness | Issue #6 open |
| Accessibility | Issue #7 blocks public preview |
| Exact page/region source navigation | Issue #8 open; filename-only citations are insufficient |
| Diagnostics and offline claims | Issue #14 open |
| Actual-build provenance | P0 issue #15 open |
| Intended purpose and claims | Issues #16, #20 and #46 open |
| Publisher/steward | Issue #17 open; preferred responsibility model selected but no organisation appointed |
| Public runnable artifacts | P0 issue #24 open; four historical artifacts remain live |
| V40 source-release target | Issue #69 open; 9 August is a target, not a waiver |

Issue #27 remains closed only for the exact Windows native-command and failure-stage controls it covers. It does not approve a runner-built executable.

## Active draft lanes

### PR #72 — Workspace Ownership V2

- Branch head: `a6bc1f0898d529b5f9eebab76f757e0926f85f86`.
- Scope: two new architecture/testing documents only.
- Workflow run: `30905573823`.
- Actual tested checkout: synthetic PR merge `24241f9addf2e6d5f1d68d721b0c5aa492abf228`.
- Artifact inventory: empty.
- Decision: **HOLD**.

Required corrections before the contract can control implementation:

1. rewrite OW-01 so stale-CAS defence does not violate full-lifetime exclusive ownership;
2. bind expected state to authenticated revision/generation, metadata digest, audit/state-chain head, target object identity and owner transaction;
3. distinguish safe retained-object continuation from forbidden pathname redirection;
4. correct workflow and reviewer evidence terminology.

No issue #4 property is implemented or proved by PR #72.

### PR #71 — first usable Matter journey

- Branch head: `05983719b29be02f44ef2b4e7ec09a8166514aa6`.
- Scope: one synthetic offline HTML prototype plus two design/testing records.
- Workflow run: `30905512134`.
- Actual tested checkout: synthetic PR merge `6016792666d7ad7d7e8b6413ad27c1213c44a5d0`.
- Artifact inventory: empty.
- Decision: **HOLD**.

Required corrections before the design can control native implementation:

- real keyboard-operable controls instead of pointer-only cards;
- no dead or misleading visible controls;
- correct modal focus management and focus return;
- accessible progress/live-status semantics and narrow/high-zoom resilience;
- exact page/region-aware citations with OCR provenance and source navigation;
- issue #65 as a hard prerequisite for any M1.18 integration;
- core accessibility failures treated as release blockers;
- qualified offline wording until issue #14 passes.

The prototype is not native application implementation, accessibility evidence or release evidence.

### Superseded PR #11

PR #11 is closed unmerged and superseded. Its stale branch must not be revived, merged or copied wholesale. Useful ideas may be selectively reimplemented from current `main` only after the PR #72 contract is corrected and reviewed.

## Development sequencing gate

Issue #53 controls the sequence:

1. correct and prove the current-main workspace foundation;
2. implement the native first usable Matter journey;
3. correct issue #65;
4. connect one controlled assistance seam;
5. qualify one exact V40 candidate.

No new disconnected engine milestone is authorised. A package test or model producing tokens is not a user-visible feature.

## Source-release gate — issue #69

A public V40 pre-alpha source tag/release remains blocked until one exact candidate passes all applicable source gates, including:

### Workspace and evidence integrity

- one writable owner for the complete Vault lifetime across processes and aliases;
- stale-save rejection against the exact authenticated state;
- safe parent/root/alias/junction/bind-mount handling;
- descriptor/handle-bound cleanup;
- owned concurrent creation and first launch;
- complete persistence rollback;
- issue #3 preservation, migration, restore and reset regressions.

### Native product journey

- a genuinely empty first launch;
- guided native Matter creation;
- responsive synthetic intake with truthful progress and cancellation;
- current position, evidence status and next actions;
- close/reopen continuity without contamination;
- a complete truthful What's New record;
- no dead, fake or placeholder controls.

### Accessibility and source navigation

- keyboard-only completion of the core journey;
- visible and logical focus, including modal behaviour;
- no inaccessible clipping or unreachable scrolling at supported sizes and scaling;
- assistive-technology exposure of controls, errors and progress;
- at least one screen-reader result;
- exact page/region-aware source citations and navigation.

These core accessibility requirements cannot be converted into a narrowed public preview.

### Source, provenance and claims

- exact release commit frozen and zero commits behind its reviewed base;
- correctly classified Linux and Windows workflow evidence;
- empty unapproved-runnable-artifact inventory;
- current README, limitations, changelog and release notes;
- exact source manifest, current SBOM and notices;
- no personal data, real evidence, credentials, private workspace or model file;
- truthful pre-alpha, synthetic-only, unsigned and limitation wording;
- issue #14 offline/diagnostic claims bounded to exact tested evidence.

The 9 August target cannot turn an OPEN or FAIL item into a pass.

## M1.18 integration gate — issue #65

M1.18 may remain on `main` only as isolated unused source while issue #65 is enforced.

Before any adapter or runtime connection, evidence must prove:

- orchestrator-owned sensitive byte buffers are zeroed unconditionally;
- a no-op, erroring or panicking eraser cannot produce a successful erasure receipt;
- API-boundary strings are documented truthfully as non-overwritable;
- non-cooperative workers and cleanup dependencies are bounded through concrete process supervision;
- callback lifetime is enforced or receipt counts are explicitly observational;
- no accepted/fallback output is released while worker or cleanup ownership remains unresolved.

## Public executable distribution — issue #24

Current workflows may build privately on the hosted runner but must not intentionally publish the executable.

These historical unsigned Actions artifacts remain live and prohibited:

| Artifact ID | Run | Expires UTC |
|---|---|---|
| `8854774165` | `30810944362` | 10 August 2026 11:49:38 |
| `8856536245` | `30815339549` | 10 August 2026 12:54:07 |
| `8863951645` | `30833597696` | 10 August 2026 16:47:05 |
| `8865678638` | `30838068198` | 10 August 2026 17:45:50 |

Do not download, execute, test or redistribute them.

Issue #24 also remains open for controlled manual-dispatch no-artifact evidence, private test-handoff controls and future release automation gated by issue #15, issue #17, trusted signing and explicit approval.

## Binary-release gate

A Windows executable remains private unless all source-release gates pass and the exact final file also has:

- a complete packaged-content inventory;
- actual-build SBOM, licences and notices;
- verified runtime/model provenance and corresponding source;
- deterministic or fully explained reproducibility evidence;
- malware/dependency checks;
- clean-machine and Acer-baseline launch, import, close/reopen, memory and crash-recovery evidence;
- keyboard, scaling and screen-reader qualification on the exact file;
- trusted Authenticode signing with no post-signing mutation;
- an explicit release-gate record authorising that exact hash.

An unsigned label, disclaimer or short retention period does not permit public executable distribution.

## Publisher and organisational gate

The preferred future model is one accountable official ECO publisher/steward that may use controlled and replaceable specialists while retaining final authority and responsibility.

No organisation is appointed, shortlisted or authorised for outreach. Candidate research does not show interest or consent. A fragmented arrangement with no accountable lead publisher is rejected.

Issue #17 remains open until a named legal organisation formally accepts and proves release, withdrawal, security, signing, continuity, complaints, data-role, contracting, liability, insurance and regulatory duties.

## Evidence classification required for every decision

Record separately:

- branch head and base;
- actual tested checkout;
- whether it is a synthetic PR merge;
- workflow run and job IDs;
- artifact inventory;
- final merge/squash identity;
- whether post-merge CI actually ran;
- reviewer/control-lane relationship.

Do not use **exact-head**, **post-merge** or **independent** unless the evidence and relationship genuinely support those terms.

## Current next gate actions

1. Correct PR #72 at a new frozen head and re-review the architecture/test contract.
2. Correct PR #71 at a new frozen head and re-review every interactive, accessibility and citation pattern.
3. Implement workspace ownership in small current-main slices with real subprocess evidence.
4. Implement the native Matter journey without M1.18 integration.
5. Correct issue #65 before Phase D.
6. Recheck the four historical artifacts after the expiry window.
7. Freeze and qualify one exact V40 source candidate only when every applicable gate passes.

## Public-record rule

Use synthetic and non-sensitive information only. Do not publish personal evidence, private workspaces, credentials, private diagnostics, exploit-level instructions, unapproved executables or model files.