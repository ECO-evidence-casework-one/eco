# Current ECO release gate

**Gate record:** `ECO-RELEASE-GATE-20260804-011`  
**Updated:** 4 August 2026, approximately 14:28 BST / 15:28 CEST  
**Canonical public status:** [`../../CURRENT_STATUS.md`](../../CURRENT_STATUS.md)  
**Baseline `main` reviewed for this record:** `9c98588387f5aed6f33371fefbf1eacbc514a5e9`  
**Live canonical source:** the repository's current `main` ref  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Approved public V40 source release:** none  
**Approved signed end-user executable:** none

## Controlling decision

ECO remains a source-development project only. The baseline SHA identifies the exact tree reviewed when this gate record was prepared; it is not intended to self-reference a future squash commit.

No branch, workflow, synthetic prototype, runner-built executable, governance record, deadline or audit pack currently authorises:

- real, sensitive or irreplaceable evidence use;
- current Ask ECO or local-model use with public or real evidence;
- a public V40 source tag or pre-alpha release;
- ordinary-user executable testing;
- public executable distribution;
- signing, release-candidate or stable-release status;
- institutional, healthcare, justice-sector or EU supply or deployment.

## Gate matrix

| Gate | Current position |
|---|---|
| Preserved evidence/source truth | Issue #3 open; issue #12 still blocks complete Ask/restore/source consistency |
| Workspace ownership | Issue #4 open; PR #72 documentation-only and held |
| First usable native Matter journey | PR #71 is design-only and held |
| M1.18 integration | P0 issue #65 blocks every adapter/model/runtime/IPC/Ask ECO/evidence/persistence connection |
| Offline AI controller | Issue #5 open |
| Responsiveness | Issue #6 open |
| Accessibility | Issue #7 blocks public preview |
| Exact source navigation | Issue #8 open; filename-only citations are insufficient |
| Diagnostic privacy/offline claims | Issue #14 open |
| Actual-build provenance | P0 issue #15 open |
| Intended purpose and public claims | Issues #16, #20 and #46 open |
| Publisher/steward | Issue #17 open; no organisation appointed; official source and binary release blocked |
| Public runnable artifacts | P0 issue #24 open; four historical artifacts remain live |
| V40 target | Issue #69 open; 9 August is a target, not a waiver |

Issue #27 is closed only for the exact Windows native-command and failure-stage controls it covers. It does not approve an executable.

## Active draft lanes

### PR #72 — Workspace Ownership V2

- Head: `a6bc1f0898d529b5f9eebab76f757e0926f85f86`.
- Scope: two new architecture/testing documents only.
- Run: `30905573823`.
- Actual tested checkout: synthetic merge `24241f9addf2e6d5f1d68d721b0c5aa492abf228`.
- Artifact inventory: empty.
- Decision: **HOLD**.

Required corrections:

1. rewrite OW-01 so stale-CAS defence does not violate full-lifetime exclusive ownership;
2. bind expected state to authenticated revision/generation, metadata digest, audit/state-chain head, target identity and owner transaction;
3. distinguish safe retained-object continuation from forbidden pathname redirection;
4. correct workflow and reviewer evidence terminology.

No issue #4 property is implemented or proved by PR #72.

### PR #71 — V40 Matter journey

- Head: `05983719b29be02f44ef2b4e7ec09a8166514aa6`.
- Scope: one synthetic offline HTML prototype and two design/testing records.
- Run: `30905512134`.
- Actual tested checkout: synthetic merge `6016792666d7ad7d7e8b6413ad27c1213c44a5d0`.
- Artifact inventory: empty.
- Decision: **HOLD**.

Required corrections:

- keyboard-operable controls instead of pointer-only cards;
- no dead or misleading visible controls;
- correct modal focus management and focus return;
- accessible progress/live-status semantics and narrow/high-zoom resilience;
- exact page/region-aware citations with OCR provenance and navigation;
- issue #65 as a hard prerequisite for M1.18;
- core accessibility failures treated as release blockers;
- qualified offline wording until issue #14 passes.

The prototype is not native implementation, accessibility evidence or release evidence.

### Superseded PR #11

PR #11 is closed unmerged and superseded. Its stale branch must not be revived, merged or copied wholesale. Useful work may be selectively reimplemented only from current `main` after the PR #72 contract is corrected and reviewed.

## Development sequence

Issue #53 controls the sequence:

1. correct and prove workspace ownership;
2. implement the native first usable Matter journey;
3. correct issue #65;
4. connect one controlled assistance seam;
5. qualify one exact V40 candidate.

No new disconnected engine milestone is authorised.

## Public V40 source-release gate — issue #69

A V40 source tag/release remains blocked until one exact candidate passes all applicable gates.

### Workspace and evidence integrity

- one writable owner for the full Vault lifetime across processes and aliases;
- stale-save rejection against the exact authenticated state;
- safe alias, junction, bind-mount, parent and root handling;
- descriptor/handle-bound cleanup;
- owned concurrent creation and first launch;
- complete persistence rollback;
- issue #3 preservation, migration, restore and reset regressions.

### Native product journey

- genuinely empty first launch;
- guided native Matter creation;
- responsive synthetic intake with truthful progress and cancellation;
- current position, evidence status and next actions;
- close/reopen continuity without contamination;
- complete truthful What's New record;
- no dead, fake or placeholder controls.

### Accessibility and source navigation

- keyboard-only completion;
- visible and logical focus, including modal behaviour;
- no inaccessible clipping or unreachable scrolling;
- assistive-technology exposure of controls, errors and progress;
- at least one screen-reader result;
- exact page/region-aware citations and source navigation.

Core accessibility failures cannot be converted into a narrowed public preview.

### Source, provenance, claims and accountability

- exact release commit frozen and compared with its reviewed base;
- correctly classified Linux and Windows evidence;
- empty unapproved-runnable-artifact inventory;
- current README, limitations, changelog and release notes;
- exact source manifest, current SBOM and notices;
- no personal data, real evidence, credentials, private workspace or model file;
- truthful pre-alpha, synthetic-only, unsigned and limitation wording;
- issue #14 claims bounded to exact tested evidence;
- a named accountable publisher/steward has formally accepted official source publishing, withdrawal, security-response, complaints and continuity duties, satisfying issue #17.

The 9 August target cannot turn an OPEN or FAIL item into a pass.

## M1.18 integration gate — issue #65

M1.18 may remain on `main` only as isolated unused source while issue #65 is enforced.

Before any connection, evidence must prove:

- unconditional zeroing of orchestrator-owned sensitive byte buffers;
- no-op, erroring or panicking erasers cannot produce a successful receipt;
- API-boundary strings are documented truthfully as non-overwritable;
- concrete process supervision bounds non-cooperative workers and cleanup;
- callback lifetime is enforced or receipt counts are explicitly observational;
- no accepted/fallback output is released while ownership remains unresolved.

## Public executable distribution — issue #24

Current workflows may build privately on the hosted runner but must not intentionally publish the executable.

These unsigned historical artifacts remain live and prohibited:

| Artifact ID | Run | Expires UTC |
|---|---|---|
| `8854774165` | `30810944362` | 10 August 2026 11:49:38 |
| `8856536245` | `30815339549` | 10 August 2026 12:54:07 |
| `8863951645` | `30833597696` | 10 August 2026 16:47:05 |
| `8865678638` | `30838068198` | 10 August 2026 17:45:50 |

Do not download, execute, test or redistribute them.

Issue #24 also remains open for controlled manual-dispatch no-artifact evidence, private test-handoff controls and future release automation gated by issue #15, issue #17, trusted signing and explicit approval.

## Binary-release gate

An executable remains private unless every source-release gate passes and the exact final file also has:

- complete packaged-content inventory;
- actual-build SBOM, licences and notices;
- verified runtime/model provenance and corresponding source;
- deterministic or fully explained reproducibility evidence;
- malware/dependency checks;
- clean-machine and Acer-baseline launch, import, close/reopen, memory and crash-recovery evidence;
- keyboard, scaling and screen-reader qualification on the exact file;
- trusted Authenticode signing with no post-signing mutation;
- explicit release approval for that exact hash.

An unsigned label, disclaimer or short retention period does not permit public executable distribution.

## Publisher and organisational gate

The preferred future model is one accountable official publisher/steward that may use controlled and replaceable specialists while retaining final authority and responsibility.

No organisation is appointed, shortlisted or authorised for outreach. Open Knowledge Foundation's deeper public-source review remains `HOLD` and creates no relationship, endorsement or contact authority. A fragmented arrangement with no accountable lead publisher is rejected.

Issue #17 remains open and blocks an official source or binary release until a named legal organisation formally accepts and proves release, withdrawal, security, signing where applicable, continuity, complaints, data-role, contracting, liability, insurance and regulatory duties.

## Evidence classification

Every relevant decision must record:

- branch head and base;
- actual tested checkout and whether it is a synthetic PR merge;
- workflow run and jobs;
- artifact inventory;
- final merge/squash identity;
- whether post-merge CI ran;
- reviewer/control-lane relationship.

Do not use **exact-head**, **post-merge** or **independent** unless the evidence and relationship support those terms.

## Next gate actions

1. Correct PR #72 at a new frozen head and re-review it.
2. Correct PR #71 at a new frozen head and re-review it.
3. Implement workspace ownership in small current-main slices with real subprocess evidence.
4. Implement the native Matter journey without M1.18.
5. Correct issue #65 before assistance integration.
6. Recheck the historical artifacts after their expiry window.
7. Freeze and qualify one exact V40 source candidate only when every applicable gate passes.

## Public-record rule

Use synthetic and non-sensitive information only. Do not publish personal evidence, private workspaces, credentials, private diagnostics, exploit-level instructions, unapproved executables or model files.
