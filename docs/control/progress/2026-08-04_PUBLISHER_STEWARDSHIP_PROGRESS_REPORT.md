# ECO progress report — publisher and stewardship — 4 August 2026

**Report ID:** `ECO-PROGRESS-20260804-PSG-001`  
**Prepared:** 4 August 2026, approximately 09:41 BST / 10:41 CEST  
**Lane:** legal, standards, governance and organisational readiness  
**Starting canonical baseline:** `64ec3549a3a92e1c9c452ad5562480335d9f8f70`  
**Canonical ending state:** `2bd53b4430d970930710f531d1012c9e65305b98`  
**Review status:** legal/governance control-lane report; no claim of organisational independence or legal advice

## Executive meaning

ECO now has a clear rule for who may eventually stand behind an official release.

The project does **not** require the originating developer or another contributor to create a company, community interest company, charity or other entity, or to become a director, trustee, publisher, support operator, complaints handler, data controller or liability owner merely through contribution.

The preferred future route is an existing established nonprofit, public-interest institution, open-source foundation or equivalent capable organisation. That organisation must complete due diligence, formally accept the role through its authorised governing body and prove that its security, support, complaints, signing, release-withdrawal and continuity arrangements work.

No publisher or organisation has been appointed. No release gate has opened.

## Mission alignment

This work belonged in the legal, standards and governance lane because it concerned:

- future legal and operational responsibility for official ECO releases;
- protection against accidental assignment of organisational duties to an individual contributor;
- security and vulnerability response;
- privacy, accessibility and product complaints;
- signing, repository recovery and continuity;
- open-source rights and official project identity;
- institutional contracts, liability, insurance and exit arrangements;
- future jurisdiction and regulatory assessments;
- audit continuity and current public release-policy wording.

This lane did not change application code, AI behaviour, executable packaging, dependencies or evidence handling.

## Work completed

### 1. Stale publisher proposal inspected

PR #19 was inspected as the historical first publisher and stewardship proposal.

It contained useful work on:

- official-release responsibility;
- security and support routes;
- complaints, signing and continuity;
- institutional contracting;
- protection against assigning duties to the originating developer merely through contribution.

It was nevertheless stale against current `main`, retained obsolete workflow ancestry and required substantive corrections recorded during its earlier review.

### 2. Current official-source boundary checked

The replacement approach was checked against current official material concerning:

- company-director duties and recurring company filings;
- charity-trustee responsibilities;
- CIC status and reporting;
- software-vendor security and incident communication;
- current UK data-protection complaint handling;
- the stated buyer-side scope of PPN 017;
- AI Act roles and intended purpose;
- Cyber Resilience Act roles, including open-source software stewards where applicable.

The sources support a cautious project decision rather than automatic entity formation. Creating or operating a company, charity or CIC would create real office-holder, governance or reporting responsibilities. The project therefore prefers an existing capable organisation and leaves exact legal roles to the actual activity, organisation, product and deployment.

### 3. Clean current-main replacement created

A new branch was created from exact canonical `main`:

`control/issue-17-clean-publisher-stewardship`

The replacement added:

- `docs/governance/PUBLISHER_AND_STEWARDSHIP_GATE.md`;
- `docs/governance/PUBLISHER_ACCEPTANCE_CHECKLIST.md`.

It aligned:

- `GOVERNANCE.md`;
- `MAINTAINERS.md`;
- `RELEASE_POLICY.md`.

Exactly five governance/public-control files changed.

### 4. Project decision recorded

The merged documents state that:

- no official publisher or operating organisation is appointed;
- no legal entity is created;
- an existing established public-interest or open-source organisation is preferred;
- contribution and repository administration alone do not appoint organisational roles;
- actual publishing, signing, contracting, data processing, supply, product claims or applicable law may create responsibilities;
- no individual may perform official supply or publishing acts on behalf of ECO before organisational authority and role allocation are documented;
- a new company, CIC, charity or similar structure would require a separate future decision identifying office-holders, duties, funding and the reason an existing host is unsuitable.

### 5. Steward duties defined

A future steward must be capable of accepting and operating:

- official release, refusal, withdrawal and end-of-support decisions;
- exact source, executable, SBOM, licence, manifest, hash and signature reconciliation;
- private vulnerability reporting and incident communication;
- privacy, accessibility, product-complaint and support routes;
- signing, revocation and recovery;
- protected repository and continuity controls;
- official-project name, release identity and false-endorsement controls;
- institutional contracting, data-role, liability, insurance and exit decisions;
- feature-, role-, market- and deployment-specific legal and regulatory assessment.

The steward must preserve ECO's fully offline, free and open-source and user-data non-custody boundaries.

### 6. Practical acceptance checklist created

The acceptance checklist provides one structured assessment covering:

- legal identity and governing authority;
- mission and product-boundary fit;
- open-source and supply-chain capability;
- security and vulnerability response;
- release, signing and maintenance;
- repository and continuity;
- privacy and support-data roles;
- accessibility, support and product complaints;
- legal, regulatory and public-claims capability;
- institutional procurement and contracting;
- synthetic and tabletop operational tests;
- formal governing-body acceptance.

The checklist directs confidential due-diligence evidence away from the public repository and prohibits uploading private identity records, credentials or sensitive material.

### 7. Release Policy corrected

`RELEASE_POLICY.md` was reconciled with current project controls by:

- replacing the obsolete blanket health-input rule with the issue #20 permitted-assistance and prohibited-output boundary;
- recording issue #27 as closed for the exact Windows commands it covers;
- removing obsolete PR #18 wording;
- adding issue #46 and the publisher acceptance gate;
- preserving every real-evidence, signing, executable-distribution and deployment block.

### 8. Exact-head validation completed

Draft PR #48 froze exact head:

`408748aac12c19d5f92135ce7011bbb389034931`

Workflow run `30892431187` passed:

- source policy;
- Linux unit tests and vet;
- controlled Windows failure self-test;
- six-stage Windows failure matrix;
- deterministic Windows build and tests.

The run's artifact inventory was exactly empty.

A control-lane review recorded a bounded PASS for the five-file governance package. It expressly did not claim organisational independence, legal advice, product approval or release approval.

### 9. Stale PR #19 retired

PR #19 was closed unmerged as **SUPERSEDED**.

Its history and review findings remain available as audit evidence. It was not described as valueless or silently deleted.

### 10. PR #48 merged and verified

PR #48 merged as:

`2bd53b4430d970930710f531d1012c9e65305b98`

Merged-tree comparison confirmed:

- exactly one squash commit;
- exactly the five expected governance/public-control files;
- no application source, tests, workflow, `VERSION`, executable, model, dependency, SBOM, notices, licence text or evidence fixture changes.

Issue #17 remained open after merge.

## Decisions

| Matter | Decision | Exact scope |
|---|---|---|
| Require the originating developer to form a company, CIC or charity now | REJECTED | Not required for source development; would create real office-holder and operational responsibilities |
| Preferred future route | PASS | Existing established nonprofit, public-interest institution, open-source foundation or equivalent capable organisation |
| Automatic contributor appointment | BLOCKED | Contribution and repository administration alone do not appoint organisational roles |
| Duties arising only through explicit consent | REJECTED | Actual conduct and applicable law may create responsibilities |
| PR #19 | SUPERSEDED | Closed unmerged; preserved as historical audit evidence |
| PR #48 exact-head governance package | PASS | Five governance/public-control files only |
| Publisher acceptance checklist | PASS | Usable due-diligence and operational-evidence framework |
| Publisher appointment | BLOCKED | No named organisation has completed or accepted the process |
| Issue #17 closure | BLOCKED | Requires actual organisation, governing-body acceptance and operational testing |
| Ordinary-user or institutional release | BLOCKED | Organisational and all other technical/release gates remain open |

## Evidence index

### Canonical commits

- starting canonical baseline: `64ec3549a3a92e1c9c452ad5562480335d9f8f70`;
- publisher/stewardship governance merge: `2bd53b4430d970930710f531d1012c9e65305b98`.

### Pull requests

- PR #19 — historical publisher/stewardship proposal, closed unmerged as superseded;
- PR #48 — clean current-main publisher and stewardship gate, merged.

### Issues

- #17 — accountable publisher and project continuity, remains open;
- #20 — permitted evidence assistance and prohibited outputs;
- #22 — live Control Board;
- #24 — public executable containment;
- #46 — UI, AI-instruction and export conformance.

### Workflow evidence

- PR #48 exact-head run `30892431187` — all jobs passed; artifact inventory empty.

## Change inventory

- Application source: **none**.
- Application tests: **none**.
- Workflow: **none**.
- Executable or installer: **none created or distributed by this work**.
- AI model or runtime: **none**.
- Dependencies: **none**.
- SBOM or third-party notices: **none**.
- Licence text: **none**.
- Governance/public documentation: **five files changed as recorded above**.
- Real or sensitive evidence: **none used**.

## Gate position

### Newly established

- controlling publisher and stewardship governance record;
- practical publisher acceptance checklist;
- current Governance, Maintainer and Release Policy alignment.

### Still open or blocked

- no official publisher or steward appointed;
- issue #17 operational and organisational acceptance;
- issue #15 actual-build SBOM, licensing and provenance;
- issues #16, #20 and #46 implementation conformance;
- issue #24 historical artifact and distribution evidence;
- application workspace integrity blockers;
- real or sensitive evidence use;
- signing;
- public executable and GitHub Release;
- ordinary-user testing;
- institutional, healthcare, justice-sector and EU supply;
- legal, medical, forensic, security, accuracy, accessibility and compliance claims.

Governance adoption did not open any release or deployment gate.

## Problems, limitations and failed actions

### Codex review unavailable

When PR #48 was marked ready, the Codex connector reported that code-review usage limits had been reached.

Effect: no Codex review was claimed. Exact-head CI, the five-file comparison and the control-lane review were used for the bounded governance decision. Lack of Codex review was not treated as approval.

### Wrong issue-update endpoint

One attempt to refresh the Control Board used an issue-comment update action with an invalid comment identifier and returned `404 Not Found`.

Effect: the failed action made no repository change. The Control Board was then updated through the correct issue-update action.

### No organisation assessed

The checklist is a framework only. No candidate organisation has been identified, contacted, assessed or appointed through this work.

Effect: issue #17 and every public-release and institutional gate remain open or blocked.

## Next controlled actions

1. Prepare a public-safe partner information and outreach pack based on the merged gate and checklist.
2. Define objective categories for suitable organisations without announcing or implying a partnership.
3. Carry issue #46 into the development lane as the UI/AI/export conformance requirement.
4. Keep issue #24 open and perform the historical-artifact check after the recorded 10 August deadline.
5. Reconcile or replace stale PR #11 before any workspace-integration decision.
6. Continue trigger-based legal, standards, accessibility, licensing and procurement monitoring.

## Review statement

This report was prepared in the ECO legal, standards and governance control lane using current repository records, exact commit comparison, issue and PR state, workflow results and artifact inventories.

The same connected project account performed the repository work and prepared the report. This is therefore a control-lane record, not an organisationally independent audit or legal opinion.

No legal entity, publisher, supplier, controller, professional role or release approval is created by this report.