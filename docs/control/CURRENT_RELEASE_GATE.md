# Current ECO release gate

**Gate record:** `ECO-RELEASE-GATE-20260804-009`  
**Updated:** 4 August 2026, approximately 10:46 BST / 11:46 CEST  
**Canonical public status:** [`../../CURRENT_STATUS.md`](../../CURRENT_STATUS.md)  
**Baseline canonical `main` reviewed for this correction:** `0dc57720c1c3394b03342427bc9b4dca09c1f040`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Signed end-user release:** none

## Controlling release decision

ECO remains a source-development project only.

The following remain blocked:

- use with real, sensitive or irreplaceable evidence;
- ordinary-user executable testing;
- executable distribution through GitHub Releases, Actions artifacts or another public channel;
- use or redistribution of historical unsigned Actions executables;
- public preview, release-candidate or stable-release status;
- public promotion of AI-assisted evidence functions before exact implementation qualification;
- diagnosis, treatment, clinical-risk, professional-representation, forensic-guarantee, profiling, scoring or authority-side decision outputs;
- official institutional, public-sector, healthcare, clinical, justice-sector or EU supply or deployment;
- claims of legal, forensic, medical, security, accuracy, accessibility or regulatory compliance.

Issue #27 is closed for the exact Windows native-command and failure-stage controls it covers. That closure does not approve the runner-built executable or alter any block above.

## Current gate matrix

| Gate | Current position |
|---|---|
| Source identity | `main` contains post-V25 controlled development and governance changes; no later named source milestone is approved |
| Evidence preservation | PR #10 materially improved controls; issue #3 remains open and depends in part on issue #12 |
| Ask verification and restore concurrency | Blocked by issue #12 |
| Workspace ownership and concurrency | Blocked by issue #4 and draft PR #11's four P0 findings |
| Windows native-command CI | Issue #27 closed for covered commands; does not approve the executable |
| Public Actions executable distribution | New uploads contained; P0 issue #24 open for historical artifacts and remaining distribution evidence |
| Diagnostic privacy and offline claims | Blocked by issue #14 |
| Actual-build runtime, model, SBOM and licences | Blocked by P0 issue #15 |
| One-file packaging | Blocked until the exact final embedded executable is built and inspected |
| Signing and authenticity | Blocked until final assembly, hashing and trusted Authenticode signing with no later mutation |
| Intended purpose | Controlling governance merged through PR #45 |
| Intended-purpose implementation and public claims | Blocked by open issue #16 and issue #46 |
| Clinical, professional, profiling and high-consequence output boundary | Blocked by P0 issue #20 |
| OCR and local AI reliability | Not approved as production functionality |
| Accessibility | Blocked pending issue #7 evidence and truthful documentation |
| Publisher, security response and continuity | Blocked by issue #17; no organisation appointed |
| Generic partner pack | Approved for generic public-safe information only; no named research or outreach authorised |
| Named organisational outreach | Blocked pending separate research, target review and contact-route approval |
| Audit evidence | PR #55 correction controls terminology; raw tested checkout identity is required |
| Public claims | Controlled development claims only |

## Evidence classification required for release records

For every relevant workflow and release decision, record separately:

- branch head SHA;
- base SHA;
- actual tested checkout SHA;
- whether the tested checkout is a synthetic pull-request merge;
- workflow run and job IDs;
- artifact inventory;
- final squash or merge SHA;
- whether post-merge CI actually occurred.

Use:

- **exact-head evidence** only when raw checkout proves the tested SHA equals the branch head;
- **tested PR merge-tree evidence** for synthetic pull-request merge refs;
- **merged-tree verification** for a separate comparison of the final merged tree;
- **post-merge CI evidence** only when a workflow actually ran against the final merged commit.

Historical records for PRs #45, #47, #48, #49 and #51 must be read with the controlling correction in `AUDIT_EVIDENCE_CLASSIFICATION_CORRECTION_2026-08-04.md`.

## Application and evidence gates

### Evidence preservation — issues #3 and #12

Issue #3 remains open. PR #10's preserved-object improvements are material but do not establish closure while:

- issue #12 remains unresolved;
- the full issue #3 acceptance record is incomplete;
- hostile, cost-bounded and concurrency evidence is incomplete.

Issue #12 requires safe verification cost, cache or source-limited verification and serialisation between Ask and restore.

Stop before real-evidence use while preserved bytes, hashes, derived readings, retrieval or restore can become inconsistent.

### Workspace identity and ownership — PR #11 / issue #4

PR #11 remains draft, unmerged and non-mergeable.

- Last reviewed workspace implementation head: `73689717bb08bb8cec0fc1233b92f843b449484a`.
- Current branch head: `61a2004809b341e72f70321843c64c3ff477f549`.
- The later branch delta is workflow containment only and fixes none of the workspace blockers.

Current P0 blockers:

1. stale writable sessions can overwrite newer metadata without exact revision/CAS protection;
2. pathname-derived locks and separately resolved paths can split ownership across aliases or substituted parents;
3. Linux nested cleanup can close an inspected object and reopen it by name;
4. creation, first launch and candidate-state writes occur before one alias-safe cross-process ownership transaction exists.

PR #11 must reconcile current canonical `main` without losing:

- issue #24 no-public-runnable-artifact controls;
- issue #27's checked native-command and failure-matrix controls;
- current Windows pointer-safety corrections;
- current provenance, SBOM-status and source-notice records;
- intended-purpose, issue #20, issue #46, publisher, partnership and audit controls;
- the canonical status and release gate.

Do not mark PR #11 ready or merge it without one coherent repair, real adversarial/concurrency tests and correctly classified evidence.

## Intended purpose, AI behaviour and claims

### Governance and issue #16

The intended-purpose and claims-control record merged through PR #45 and is controlling governance.

It defines ECO as a private, local, user-controlled evidence assistant. Governance adoption does not prove application conformance, legal classification, release readiness or deployment approval.

Issue #16 remains open until the UI, help, model/system cards, exports, screenshots, release notes, website or partner material and prohibited-claim controls are consistent and evidenced. Issue #46 carries that cross-surface implementation work.

PR #18 is closed unmerged and not controlling.

### Permitted evidence assistance in principle

Subject to source binding, visible output status and meaningful user review, ECO may be designed to:

- preserve, display, extract, OCR, search and navigate supplied evidence;
- identify names, dates, deadlines, events and actions stated in supplied documents;
- produce source-linked summaries, chronologies and comparisons of what documents say;
- help a user prepare questions, notes, letters and draft responses;
- identify possible omissions, conflicts or uncertainty for the user to examine.

Health, legal or other sensitive subject matter alone is not a reason for a blanket refusal.

### P0 issue #20

No deterministic or generated route may present ECO as able to:

- diagnose or infer an unstated diagnosis;
- recommend medication, treatment or a clinical course;
- perform prognosis, triage, monitoring, clinical-risk or emergency assessment;
- conduct reserved legal activities or provide professional representation;
- guarantee legal correctness, admissibility, authenticity or outcome;
- act as a forensic laboratory or expert witness;
- profile or score eligibility, entitlement, credibility, honesty or dangerousness;
- grant, deny, reduce, revoke or reclaim an essential benefit, service, right or opportunity;
- make or materially influence an official adverse decision;
- claim emergency monitoring or contact that did not occur.

P0 issue #20 remains open until deterministic Ask, future local-model routes, retrieval, memory, citations, refusals, drafts and exports pass the required synthetic evidence tests.

### Issue #46

Issue #46 requires all user-facing, model-facing and public surfaces to preserve the same boundary and distinguish:

1. original source text;
2. extraction or OCR;
3. an ECO-generated suggestion;
4. a user confirmation;
5. a user-authored note.

Current Ask ECO and future generative routes are not approved for real/sensitive evidence or public AI-assisted use.

## Provenance, licensing, packaging and signing

### P0 issue #15

Historical/source-level SBOM and notice records do not describe a current packaged release.

Issue #15 remains open until the exact final executable is reconciled with:

- immutable source identity;
- complete packaged-content manifest;
- actual-build SPDX SBOM;
- licences and offline notices for every redistributed component;
- model/runtime provenance, source and hashes;
- corresponding source and build recipe;
- reproducibility or a fully explained difference set;
- clean-machine operation;
- final executable hash, trusted signature and signing record;
- a release receipt proving every record refers to the same artifact.

No executable, model or native runtime may be publicly distributed before this gate passes.

### One-file delivery

The intended official product is one self-contained Windows application. Users must not be required to install or manage a browser server, Python, a model, a native runtime, DLL folders or separate application assets.

The objective remains unproved for a final release candidate.

### Signing

Before ordinary-user distribution:

- the exact complete file must be assembled first;
- its source, manifest, SBOM, notices and hash must be reconciled;
- manual signing approval must be recorded;
- Authenticode verification must succeed;
- no file mutation may occur after signing;
- revocation, certificate and recovery arrangements must be operational.

A signature alone does not satisfy application, evidence-integrity, accessibility, provenance, publisher or distribution gates.

## Public distribution — issue #24

The former workflow uploaded unsigned runnable packages after Windows jobs. The public upload step has been removed; current workflows may build and test internally on the hosted runner but must not intentionally expose the executable as a downloadable artifact.

Issue #24 remains P0 because:

- four historical unsigned executable artifacts require authorised deletion or exact expiry confirmation;
- controlled manual-dispatch no-artifact evidence is incomplete;
- private controlled test-handoff evidence is incomplete;
- future release automation is not yet gated by issue #15, issue #17, trusted signing and explicit approval.

PRs #18 and #19 are closed unmerged. PR #11 remains a stale implementation branch requiring future reconciliation.

An enabled control task is scheduled for **10 August 2026 at 19:00 Europe/London** to recheck historical artifact IDs `8856536245`, `8854774165`, `8863951645` and `8865678638` without downloading, executing, testing or redistributing them.

A short retention period or names such as `unsigned`, `test`, `temporary` or `provenance` do not permit public runnable distribution.

## Publisher and organisational gate — issue #17

The publisher/stewardship gate and acceptance checklist merged through PR #48.

No organisation or individual is appointed. No contributor becomes the publisher, supplier, director, trustee, support operator, complaints handler, controller or liability owner merely through contribution or repository administration.

This project rule does not prevent responsibilities arising through actual publishing, signing, contracting, supply, data processing, claims or applicable law.

Before an official ordinary-user or institutional release, a named established organisation must:

- complete due diligence;
- formally accept the duties through its governing authority;
- assign authorised role-holders and deputies;
- operate private vulnerability, privacy, accessibility, product-complaint and support routes;
- govern release, signing, withdrawal and end of support;
- prove repository, certificate and maintainer-unavailability recovery;
- decide contracts, data roles, liability, insurance and exit arrangements where relevant.

Issue #17 remains open.

The generic partner pack merged through PR #51 for possible later public-safe, non-binding discussion. It identifies or contacts no organisation. Named research and outreach require a separate issue and approval. PR #19 is closed unmerged and superseded.

## Stop rules

### Stop before real-evidence testing

Stop where any of the following remains unresolved:

- preserved bytes, hashes and derived readings cannot be reconciled;
- Ask, restore, migration, creation or reset can observe or create mixed or unsafe state;
- source attribution and output-status separation are incomplete;
- a route can produce prohibited clinical, professional, forensic, profiling, scoring or adverse-decision output;
- filesystem boundaries are not proven against Windows reparse points, junctions, links, aliases and parent substitution;
- diagnostics or support routes can expose case content;
- any real-evidence P0 or P1 remains open.

### Stop before public preview

Stop where:

- any public channel exposes an unapproved runnable payload;
- the executable is unsigned or is not the actual one-file deliverable;
- accessibility evidence is incomplete;
- the application or AI can invent controls, actions, source support or professional conclusions;
- intended-purpose governance and implementation are not aligned;
- no accountable organisation has accepted publisher and response duties;
- public claims cannot be supported by the exact candidate.

### Stop before GitHub Release or public executable upload

Stop where:

- exact source, build receipt, content manifest, actual-build SBOM or notices are absent;
- model/runtime provenance is incomplete;
- clean-machine, low-resource, stability and Smart App Control testing have not passed;
- the file is unsigned or was modified after signing;
- issue #24 remains incomplete;
- any release-blocking P0 or release-relevant P1 remains unresolved.

No label, disclaimer or risk-acceptance statement may waive these stop rules.

## Gate effect of recent governance and audit work

- PR #45 established intended-purpose governance only.
- PR #48 established publisher/stewardship governance and a future acceptance checklist only.
- PR #51 created a generic public-safe partner pack only.
- PR #55 corrected audit-evidence terminology and reviewer descriptions only.
- PR #57 preserved a reconciled dated report only.

None approved application functionality, real evidence, an executable, signing, a publisher, outreach, release or deployment.

## Next gate actions

1. Reconcile PR #11 metadata to the current canonical state without changing or approving its implementation.
2. Carry issue #46 into the development lane and preserve exact changed-string and synthetic test evidence.
3. Keep issue #24 open and perform the scheduled 10 August historical-artifact check.
4. Only after canonical records agree, create a separate named-organisation research and assessment issue.
5. Do not send outreach until the research method, named target and public contact route are separately reviewed and approved.

## Public-record rule

Public status documents may identify defect classes, controls, tests and decisions. They must not expose personal evidence, private diagnostics, unapproved binaries, model files, private workspaces, credentials or exploit-level instructions.
