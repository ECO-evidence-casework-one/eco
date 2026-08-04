# Current ECO project status

**Status date:** 4 August 2026  
**Control update:** approximately 10:46 BST / 11:46 CEST  
**Canonical public status record:** this file  
**Baseline canonical `main` reviewed for this correction:** `0dc57720c1c3394b03342427bc9b4dca09c1f040`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Release position:** source development only; no approved signed end-user release

## Status authority

This root `CURRENT_STATUS.md` is ECO's canonical public status summary.

The live Control Board in issue #22 records current operational detail and next actions. Pull requests, issues, raw workflow logs, artifact inventories and dated reports preserve the supporting audit history. Where a historical record conflicts with a later controlling correction, read it with:

- `docs/control/AUDIT_EVIDENCE_CLASSIFICATION_CORRECTION_2026-08-04.md`;
- `docs/control/PROGRESS_REPORTING_AND_AUDIT_STANDARD_SUPPLEMENT_2026-08-04.md`.

## Evidence terminology

GitHub pull-request workflows commonly test a synthetic merge ref rather than the branch head itself.

ECO therefore distinguishes:

- **exact-head evidence** — raw checkout evidence proves that the tested SHA equals the stated branch head;
- **tested PR merge-tree evidence** — the workflow checked out a synthetic pull-request merge commit;
- **merged-tree verification** — the final squash or merge tree was compared with its accepted proposal and previous base;
- **post-merge CI evidence** — a workflow actually ran against the final merged commit.

A green pull-request check must not be described as exact-head or post-merge evidence without the corresponding raw identity.

## What `main` represents

`VERSION` still records the last named source milestone approved under the earlier milestone process: `ECO-V25-20260731-N2-P1`.

`main` also contains later controlled source-development, repository, governance and audit changes. Those changes do not create a new source milestone, release candidate, stable release or approved executable merely because they were merged.

The public repository must contain only source, synthetic/non-sensitive fixtures and sanitised control records. It does not represent private models, personal evidence, private workspaces or a supported end-user distribution.

## Current control position

| Area | Current position |
|---|---|
| Evidence preservation | PR #10 materially improved preserved-object handling; issue #3 remains open and depends in part on issue #12 |
| Workspace identity and concurrency | PR #11 remains draft, unmerged and non-mergeable with four P0 blockers |
| Windows native-command CI | Issue #27 is closed for the exact commands and failure stages it covers; this does not approve the built executable |
| Public Actions executable distribution | New public runnable uploads are contained; P0 issue #24 remains open for historical artifacts and remaining distribution evidence |
| Actual-build provenance | P0 issue #15 remains open; no authoritative final executable SBOM, manifest, notice bundle, signing record or release receipt exists |
| Intended purpose | Controlling governance merged through PR #45; issue #16 remains open for cross-surface implementation and claims evidence |
| Prohibited high-consequence output boundary | P0 issue #20 remains open |
| UI, AI instructions and export conformance | Issue #46 remains open |
| Publisher and stewardship | Governance and checklist merged through PR #48; issue #17 remains open because no organisation is appointed |
| Generic partner information | Merged through PR #51; issue #50 closed for generic material only |
| Named-organisation research and outreach | Not authorised; requires a separate issue, evidence method and target approval |
| Audit-evidence correction | Merged through PR #55 |
| Dated partner-pack report | Reconciled and merged through PR #57 |
| Real or sensitive evidence | Blocked |
| Signed public executable | None |
| Institutional, healthcare, justice-sector and EU availability | Blocked |

## Application and evidence-integrity position

### Evidence preservation — issues #3 and #12

PR #10 merged material preservation improvements intended to keep the verified preserved object as the source of truth for extraction, OCR, citation and retrieval.

Issue #3 remains open. Issue #12 still requires bounded verification cost, safe cache/source verification and safe serialisation between Ask and restore. Until the complete acceptance evidence is recorded and reviewed, ECO must not be used with real, sensitive or irreplaceable evidence.

### Workspace identity, creation, migration and reset — PR #11 / issue #4

PR #11 remains draft, unmerged and non-mergeable.

- Last reviewed workspace implementation head: `73689717bb08bb8cec0fc1233b92f843b449484a`.
- Current PR head: `61a2004809b341e72f70321843c64c3ff477f549`.
- The later branch delta contains workflow containment only; it does not correct the workspace defects.

The four current P0 blockers are:

1. **stale concurrent writers** — independently opened writable sessions can retain stale state and later overwrite newer workspace metadata;
2. **split or alias-bypassable ownership** — pathname-derived locking and separately resolved paths do not establish one retained object-identity transaction;
3. **Linux nested cleanup reopening by name** — an inspected child can be closed and reopened without identity continuity;
4. **unowned creation and first launch** — managed workspace or candidate state can be created or replaced before one alias-safe cross-process ownership transaction exists.

PR #11 must be rebuilt or reconciled against current canonical `main` without losing:

- issue #24's no-public-runnable-artifact rule;
- issue #27's checked native-command and failure-matrix controls;
- current Windows pointer-safety corrections;
- current provenance, SBOM-status and notice records;
- intended-purpose, publisher/stewardship, partnership and audit controls;
- the current canonical status and release gate.

It must not be marked ready, merged or used to close issue #4 without one coherent repair, real concurrent/adversarial tests and correctly classified workflow evidence.

## Intended purpose and AI/output controls

### Governance boundary — issue #16

The controlling intended-purpose record merged through PR #45 and defines ECO as a private, local, user-controlled evidence and casework assistant.

It permits source-linked assistance with supplied health, legal and other sensitive material in principle. Document subject matter alone is not a reason for a blanket refusal.

Governance adoption does not prove that the current application, deterministic Ask routes, future local model, help, exports, screenshots or partner material conform. Issue #16 therefore remains open, with issue #46 carrying the cross-surface implementation and claims work.

Stale PR #18 is closed unmerged and is not controlling.

### Permitted assistance in principle

Subject to exact source binding, clear output status and meaningful user review, ECO may be designed to:

- preserve, display, extract, OCR, search and navigate supplied material;
- identify document-stated names, dates, deadlines, events and requested actions;
- produce source-linked summaries, chronologies and comparisons of what documents say;
- help a user prepare questions, notes, letters and draft responses;
- identify possible omissions, conflicts or uncertainty for the user to examine.

### Prohibited or separately gated outputs — issue #20

ECO must not be presented, configured or relied upon to:

- diagnose or infer an unstated diagnosis;
- recommend medication, treatment or a clinical course;
- perform prognosis, triage, monitoring, clinical-risk or emergency assessment;
- conduct reserved legal activities or claim professional legal representation;
- guarantee legal correctness, admissibility, authenticity or outcome;
- act as a forensic laboratory or expert witness;
- profile or score eligibility, entitlement, credibility, honesty or dangerousness;
- grant, deny, reduce, revoke or reclaim an essential benefit, service, right or opportunity;
- make or materially influence an official or institutional adverse decision;
- imply emergency monitoring, escalation or contact that did not occur.

P0 issue #20 remains open until deterministic and local-model routes, retrieval, memory, citations, drafts, refusals and exports pass the required synthetic tests. Current Ask ECO and future model-assisted routes are not approved for real or public evidence use.

### Cross-surface conformance — issue #46

Issue #46 requires the UI, onboarding, help, deterministic Ask, future local-model instructions, memory, refusal messages, summaries, drafts, exports, screenshots, release notes and partner material to use the same boundary.

Every surface must distinguish:

1. original source text;
2. extraction or OCR;
3. an ECO-generated suggestion;
4. a user confirmation;
5. a user-authored note.

A warning or disclaimer does not compensate for hidden sources or authoritative presentation.

## Provenance, licensing and distribution

### Actual-build provenance — issue #15

Root SBOM and historical provenance records are explicitly source-level or historical records. They do not describe a final packaged release.

P0 issue #15 remains open until one exact final executable is reconciled with:

- immutable source identity;
- complete content manifest;
- actual-build SPDX SBOM;
- all required licences and notices;
- model/runtime provenance and hashes;
- corresponding source and build recipe;
- clean-machine and reproducibility evidence;
- final executable hash and trusted signature;
- a release receipt proving all records identify the same artifact.

### Public Actions executable containment — issue #24

The public `actions/upload-artifact` executable package was removed. Current workflows build and test internally on the hosted runner without intentionally publishing the executable.

P0 issue #24 remains open because:

- four historical unsigned executable artifacts require authorised deletion or exact expiry confirmation;
- controlled manual-dispatch evidence remains incomplete;
- private controlled test-handoff rules and evidence remain incomplete;
- future release automation must be gated by issue #15, issue #17, trusted signing and explicit release approval.

PRs #18 and #19 are closed unmerged. PR #11 remains the only listed stale implementation branch requiring future current-main reconciliation.

An enabled one-time control task is scheduled for **10 August 2026 at 19:00 Europe/London** to recheck the four historical artifact IDs without downloading, executing, testing or redistributing them.

No historical or runner-only executable is approved for use or redistribution.

## Publisher, stewardship and partnership position

The controlling publisher/stewardship gate and acceptance checklist merged through PR #48.

No organisation or individual is appointed as ECO's official publisher, supplier, support operator, complaints handler, controller or liability owner.

The project does not presently require the originating developer or another contributor to form a company, CIC, charity or other entity or become a director, trustee or equivalent office-holder merely through contribution.

Contribution alone does not prevent responsibilities arising through actual publishing, signing, contracting, supply, data processing, public claims or applicable law. No individual may perform official ECO publishing or supply acts before authority and role allocation are documented.

The preferred future route is an existing established nonprofit, public-interest institution, open-source foundation or equivalent capable organisation. Issue #17 remains open until a named organisation:

- completes due diligence;
- formally accepts the role through its governing authority;
- funds and assigns role-holders and deputies;
- proves security, complaints, release withdrawal, signing and continuity routes through synthetic or tabletop exercises.

The generic partner pack merged through PR #51. It identifies and contacts no organisation and authorises no research target or outreach. Informal interest would not constitute organisational acceptance.

Stale PR #19 is closed unmerged and superseded by PR #48.

## Audit and reporting position

PR #55 made the workflow-evidence correction controlling. Historical PRs #45, #47, #48, #49 and #51 tested identified synthetic PR merge trees rather than the branch heads previously labelled exact-head.

Their successful jobs and empty artifact inventories remain evidence for those tested trees. Final squash-tree comparisons are separate evidence. No unobserved post-merge CI is claimed.

PR #57 preserved the reconciled partner-pack progress report. Stale PRs #52 and #54 are closed unmerged as superseded.

## Current stop gates

The following remain blocked:

- real, sensitive or irreplaceable evidence use;
- ordinary-user testing with an ECO executable;
- public executable distribution through Releases, Actions artifacts or another public channel;
- use or redistribution of historical unsigned executables;
- release-candidate or stable-release status;
- public promotion of AI-assisted evidence functions before issue #20 and issue #46 qualification;
- diagnosis, treatment, clinical-risk, professional-representation, forensic-guarantee, profiling, scoring or authority-side decision outputs;
- claims of production OCR, reliable generative local AI or complete native document investigation;
- claims of legal, medical, forensic, security, accuracy, accessibility or regulatory compliance;
- official institutional, public-sector, healthcare, clinical, justice-sector or EU supply and deployment;
- named partner outreach without a separate approved research and outreach control.

## Release prerequisites

Before the release position can be reconsidered, ECO requires objective evidence for at least:

- exact source, build and binary identity;
- complete actual-build manifest, SBOM, licences and notices;
- verified model and runtime provenance;
- genuine one-file packaging and clean-machine operation;
- Windows stability, low-resource and long-duration testing;
- no-external-network proof for the exact bundled runtime;
- evidence-integrity, restore, migration and recovery qualification;
- privacy-safe diagnostics;
- keyboard, assistive-technology, DPI, scrolling and cognitive-accessibility evidence;
- page-aware search and source navigation evidence;
- technical enforcement of issues #20 and #46;
- trusted Authenticode signing with no post-signing mutation;
- an accountable accepted publisher/steward;
- a controlled distribution pipeline exposing no runnable payload before approval;
- closure of every release-blocking P0 and release-relevant P1 finding with correctly classified evidence.

## Next controlled sequence

1. Reconcile PR #11's public metadata to this canonical state without changing or approving its implementation.
2. Carry issue #46 into the development lane.
3. Keep issue #24 open and perform the scheduled 10 August historical-artifact check.
4. Create a separate named-organisation research and assessment issue only after the canonical status and release gate merge and agree.
5. Do not contact any organisation until the research method, target and public contact route are separately reviewed and approved.

## Public-record rule

Public updates may record defect classes, controls, tests and decisions. They must not publish personal evidence, private diagnostics, credentials, exploit-level instructions, unapproved binaries, models or private workspaces.

Use synthetic and non-sensitive information only in the public repository.
