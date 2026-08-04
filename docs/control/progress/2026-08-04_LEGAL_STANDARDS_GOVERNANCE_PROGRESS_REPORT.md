# ECO progress report — legal, standards and governance — 4 August 2026

**Report ID:** `ECO-PROGRESS-20260804-LSG-001`  
**Prepared:** 4 August 2026, 07:50 BST / 08:50 CEST  
**Lane:** legal, standards, governance and release control  
**Starting reviewed baseline:** `f0894289b4bfd492f92df27e842c2879e59fa17a`  
**Canonical ending state:** `d63206f64eff3eb230907e869b5d7530cb6d9f8a`  
**Report-creation branch baseline:** `d63206f64eff3eb230907e869b5d7530cb6d9f8a`  
**Review status:** control-lane report; repository facts checked against exact commits, PRs, issues and workflow evidence; no claim of organisational independence

## Executive meaning

ECO's legal and governance position is now clearer and less restrictive.

ECO is officially intended to be a private, local and user-controlled evidence assistant. It may help a user understand health, legal and other sensitive evidence in principle, but it must not diagnose, recommend treatment, provide professional representation, make forensic guarantees, score people or decide rights and services.

The repository also now describes historical provenance records truthfully and prevents new public GitHub Actions executable uploads. However, no application, AI, executable or public release was approved by this work.

This report also records a process weakness discovered during the work: detailed evidence existed, but the central Control Board and readable progress reporting had fallen behind. The Control Board was refreshed and a permanent reporting standard was prepared.

## Mission alignment

This work belonged in the legal, standards and governance lane because it concerned:

- intended purpose and excluded uses;
- public claims control;
- medical, legal-services, forensic and high-consequence boundaries;
- publisher and stewardship responsibility;
- executable distribution and release gates;
- provenance, SBOM and licensing truthfulness;
- repository audit continuity.

This lane did **not** change application code, AI models, executable packaging, dependencies or development implementation.

## Work completed

### 1. Repository provenance and public-record corrections

The repository was independently rechecked after the earlier control checkpoint.

Material documentation-control merges included:

- PR #39: corrected README wording about issue #27;
- PR #40: marked V25 SBOM and notices as historical source records;
- PR #41: reconciled canonical status and release-gate records;
- PR #43: removed the obsolete blanket health-content restriction from canonical status records;
- PR #45: established the controlling intended-purpose and public-claims record and reconciled principal public documents.

An accidental `noop` commit and its immediate removal were assessed as zero net tree change. They created audit noise but no surviving content defect.

### 2. Public executable containment continuity

Issue #24 remained open.

The current workflow continued to build and test internally without publishing a new executable artifact. Exact-head runs inspected during the governance work returned empty artifact inventories.

Four historical executable artifacts remain subject to exact expiry verification after the latest recorded expiry on 10 August 2026 at 18:45:50 BST / 19:45:50 CEST, plus a reasonable propagation margin.

No historical artifact was downloaded or executed.

### 3. Intended-purpose correction

The project initially treated health-related content too broadly as a reason to block generated assistance.

That position was corrected:

- issue #20 was rewritten around actual prohibited outputs and deployments rather than document subject matter;
- stale PR #18 was closed unmerged and superseded;
- PR #43 corrected the two canonical status documents;
- issue #42 was closed after the correction passed CI with an empty artifact inventory;
- PR #45 merged the controlling intended-purpose and public-claims record.

The controlling purpose now permits source-linked, user-controlled assistance with supplied health, legal and other sensitive material in principle.

The following remain prohibited or separately gated:

- diagnosis and treatment recommendations;
- prognosis, triage, monitoring and clinical-risk assessment;
- professional legal representation and reserved legal activities;
- forensic, authenticity and admissibility guarantees;
- eligibility, credibility, honesty and dangerousness scoring;
- autonomous or authority-side adverse decisions;
- false emergency monitoring or contact claims;
- unsupported compliance, security, accuracy and accessibility claims.

### 4. Cross-surface governance reconciliation

PR #45 added:

- `docs/governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md`;
- a governance link and adoption boundary in `GOVERNANCE.md`.

It also corrected stale or contradictory wording in:

- `README.md`;
- `KNOWN_LIMITATIONS.md`;
- `ROADMAP.md`.

`PRIVACY.md` and `SECURITY.md` were inspected and left unchanged because their qualified source-level wording did not contradict the corrected boundary.

### 5. Implementation conformance requirement

Issue #46 was created to convert the governance rule into application requirements.

It requires future inspection and alignment of:

- onboarding and help;
- deterministic Ask ECO;
- local-model system instructions;
- boundary and refusal messages;
- summaries, timelines, comparisons and drafting routes;
- exports and reports;
- screenshots, demonstrations and public or partner wording.

Issues #16 and #20 remain open because the application has not yet proved conformance.

### 6. Audit continuity restoration

The live Control Board in issue #22 was found to be materially stale. It still named an obsolete `main`, active PR #18 and the former health-input restriction.

Issue #22 was refreshed at 4 August 2026, 07:45 BST / 08:45 CEST to record:

- canonical `main` `d63206f64eff3eb230907e869b5d7530cb6d9f8a`;
- active PRs #11 and #19;
- issues #15, #16, #17, #20, #24 and #46;
- unchanged release and evidence-use blocks;
- the controlled next sequence;
- a requirement for daily closing reports on days with substantial ECO work.

The present reporting standard and dated report were then prepared on a separate documentation branch.

## Decisions

| Matter | Decision | Exact scope |
|---|---|---|
| PR #39 README correction | PASS | One-file public-record correction only |
| PR #40 historical SBOM/notices correction | PASS | Historical/source-level provenance truth only; issue #15 remains open |
| Accidental noop sequence | NO CHANGE | Zero net tree difference; audit noise only |
| Blanket health-content restriction | SUPERSEDED | Replaced by output-, purpose- and deployment-based controls |
| PR #18 | SUPERSEDED | Closed unmerged; not controlling |
| PR #43 canonical issue #20 correction | PASS | Two public-control documents only |
| Issue #42 | PASS / CLOSED | Canonical wording corrected; no application approval |
| PR #45 intended-purpose governance | PASS | Governance and public-document boundary only |
| Application conformance | BLOCKED | Issues #16, #20 and #46 remain open |
| Public executable distribution | BLOCKED | Issue #24 remains open |
| Real or sensitive evidence | BLOCKED | No implementation qualification exists |
| Publisher and stewardship | BLOCKED | Issue #17 open; PR #19 stale and non-controlling |
| Formal progress reporting | HOLD pending merge | Standard and first dated report prepared on documentation branch |

## Evidence index

### Canonical commits

- previous control checkpoint: `f0894289b4bfd492f92df27e842c2879e59fa17a`;
- README correction: `c0f6f891174032e4e80274200388dce85df05bbe`;
- historical SBOM/notices correction: `4c6929b6b94e4a428ccc69336b50a38ba9fe6271`;
- canonical status reconciliation: `c36753f0fff3daef3b94d1e3ae11fd25bea8933e`;
- issue #20 canonical correction: `f6c852616deb96f7b71414e9526b8eca376968e5`;
- intended-purpose governance merge: `d63206f64eff3eb230907e869b5d7530cb6d9f8a`.

### Pull requests

- PR #39 — README issue #27 correction;
- PR #40 — historical V25 SBOM and notices;
- PR #41 — canonical status/release-gate reconciliation;
- PR #43 — corrected issue #20 evidence-assistance boundary;
- PR #45 — controlling intended purpose and claims;
- PR #11 — stale blocked workspace implementation lane;
- PR #19 — stale blocked publisher/stewardship lane.

### Issues

- #15 — actual-build SBOM, licences and provenance;
- #16 — intended purpose and cross-surface conformance;
- #17 — accountable publisher and continuity;
- #20 — permitted evidence assistance and prohibited outputs;
- #22 — live Control Board;
- #24 — public executable artifact containment;
- #42 — canonical intended-purpose correction, closed;
- #46 — UI, AI-instruction and export alignment.

### Workflow evidence

- PR #43 exact-head run `30882430585`: source policy, Linux and Windows passed; artifact inventory empty;
- PR #45 final exact-head run `30884028310`: source policy, Linux and Windows passed; artifact inventory empty.

## Change inventory

- Application source: **none**.
- Application tests: **none**.
- Workflow: **none** in the intended-purpose and reporting work.
- Executable: **none created or distributed by this lane**.
- AI model or runtime: **none**.
- Dependencies: **none**.
- Licence text: **none**.
- Governance and documentation: **material changes described above**.
- Real or sensitive evidence: **none used**.

## Gate position

### Newly established

- A controlling intended-purpose and public-claims record is now present on canonical `main`.

### Still open or blocked

- issue #15 actual-build provenance and licensing;
- issue #16 implementation and cross-surface conformance;
- issue #17 publisher and stewardship;
- issue #20 behavioural output boundary;
- issue #24 historical artifacts and remaining distribution evidence;
- issue #46 UI, AI-instruction and export alignment;
- real and sensitive evidence use;
- public executable distribution and GitHub Release;
- signing;
- ordinary-user testing;
- institutional, healthcare, justice-sector and EU supply;
- legal, medical, forensic, security, accuracy, accessibility and compliance claims.

No gate was opened merely because governance wording was adopted.

## Problems, limitations and failed actions

### Scheduled artifact reminder

A requested scheduled task for the 10 August artifact-expiry check was not created because the account had already reached its limit of five active tasks.

Effect: the underlying audit requirement remains recorded in issue #24 and issue #22, but there is no successfully created reminder task.

### Manual workflow dispatch

A controlled manual-dispatch run could not be started because the available GitHub connection did not expose a fresh workflow-dispatch action, the local `gh` command was unavailable and no suitable token was present.

Effect: issue #24's manual-dispatch acceptance item remains open. It was not treated as failed or passed.

### Codex review availability

The Codex connector reported that code-review usage limits had been reached during PR #45.

Effect: the bounded control review, exact-head CI and merged-tree comparison were completed, but no separate Codex review was claimed. Lack of Codex review was not treated as approval.

### Audit-summary staleness

Issue #22 had become stale despite detailed event evidence existing elsewhere.

Effect: a reviewer relying only on the Control Board could have misunderstood the current project. The board was refreshed and the present reporting framework was prepared.

### Timestamp correction

The first draft used the user-message time, 07:41 BST / 08:41 CEST, as the preparation and Control Board update time. GitHub's authoritative write timestamps showed that the Control Board was updated at approximately 07:45 BST and the audit PR was opened at approximately 07:50 BST.

Effect: the report was corrected before merge. No substantive project decision changed.

## Next controlled actions

1. Review and merge the progress-reporting and audit-standard documentation branch if its exact scope and CI are clean.
2. Carry issue #46 into the development lane as the controlling implementation-conformance requirement.
3. Rebuild or reconcile the issue #17 publisher and stewardship proposal from current `main`, without appointing an individual contributor by default.
4. Keep issue #24 open and perform the exact historical-artifact expiry check after 10 August 2026 at 18:45:50 BST / 19:45:50 CEST plus propagation margin.
5. Reconcile or replace stale PR #11 before any further workspace-integration decision.
6. Continue trigger-based official legal, standards, accessibility, licensing and procurement monitoring.

## Review statement

This report was prepared in the ECO legal, standards and governance control lane using GitHub repository records, exact commit comparisons, PR and issue state, workflow job results and artifact inventories.

The same connected project account performed much of the repository work and prepared this report. The report therefore does not claim organisational or personnel independence. It records a control-lane review and must be assessed on the cited evidence and exact repository identities.

No legal opinion, regulatory classification, release approval or professional assurance is created by this report.
