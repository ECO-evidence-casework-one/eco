# ECO progress report — public-safe partner pack — 4 August 2026

**Report ID:** `ECO-PROGRESS-20260804-PPK-001`  
**Original work prepared:** 4 August 2026, approximately 09:58 BST / 10:58 CEST  
**Reconciled after audit correction:** 4 August 2026  
**Lane:** legal, governance and public-safe organisational exploration  
**Starting canonical baseline:** `62e00ef5e573605b2a96b445a0dc42b0f4e387bd`  
**Partner-pack merge:** `f997e1049f8c24ed04848127ec26d55ee784b6f4`  
**Audit-correction baseline for this report:** `e6e83dd7adc0269c16afda806c324063264e5e6e`  
**Review status:** control-lane report; not an organisationally independent audit, partnership authority or legal opinion

## Executive meaning

ECO now has a generic public-safe information and outreach pack that can explain the project to established organisations without implying that a partnership, publisher or endorsement already exists.

The pack requests only a non-binding written discussion. It does not authorise named-organisation research, actual outreach, a real-data pilot, a contract, repository or signing transfer, release or deployment.

No organisation was named, researched, ranked, contacted or appointed through this work.

## Mission alignment

This work followed the publisher and stewardship governance merged through PR #48.

Public outreach can accidentally:

- imply a relationship or endorsement;
- make unsupported product, security or compliance claims;
- invite real evidence or premature pilots;
- weaken offline, open-source or user-data non-custody requirements;
- place organisational duties on an individual contributor;
- create apparent commitments before governing-body approval.

The generic pack creates controls against those risks before any named-organisation work begins.

## Work completed

### 1. Generic-pack issue defined

Issue #50 required a plain-English partner pack and prohibited:

- naming or ranking organisations;
- actual outreach;
- partnership or publisher claims;
- real evidence and real-data pilots;
- entity formation;
- contracts, fundraising or institutional supply;
- product or release approval.

### 2. Partner information pack created

`docs/partnership/PARTNER_INFORMATION_PACK.md` explains:

- what ECO is and the public-interest problem it intends to address;
- fully offline operation;
- free and open-source requirements;
- no required accounts, telemetry, advertising or routine project access to user files;
- one-file Windows and low-spec objectives;
- permitted source-linked evidence assistance;
- prohibited medical, legal, forensic, scoring and official-decision uses;
- current source-development and release limitations;
- why an existing established organisation is preferred;
- what ECO seeks and does not seek;
- staged exploration, due diligence, governing-body acceptance, operational validation and separate release approval.

### 3. Controlled outreach template created

`docs/partnership/EXPLORATORY_OUTREACH_TEMPLATE.md` provides:

- a pre-send checklist;
- a standard written message;
- a short-form written message;
- a response-record template;
- prohibited outreach wording;
- escalation stops.

The template states that ECO has no approved public executable, real-evidence approval or appointed publisher. It asks only for a non-binding written discussion.

### 4. Partnership-use controls created

`docs/partnership/README.md` records:

- no current partner or publisher;
- controlling governance and release records;
- the required sequence before named outreach;
- prohibited uses of the pack;
- the rule that completion opens no product, evidence-use, signing, release or deployment gate.

### 5. PR #51 validation correctly classified

PR #51 branch head was:

`c335733309838b8468a9fd102b456fece4f83ab3`

Workflow run `30893783028` tested synthetic PR merge commit:

`18a47d0f729d4634a201a7821232f5c264b7775f`

The tested merge combined branch head `c3357333...` with base `62e00ef5...`.

The tested PR merge-tree run passed:

- source policy;
- Linux tests and vet;
- controlled Windows failure self-test;
- six-stage Windows failure matrix;
- deterministic Windows build and tests.

The artifact inventory was empty.

This is **tested PR merge-tree evidence**, not branch-head or exact-head CI.

### 6. PR #51 merged and separately compared

PR #51 merged as:

`f997e1049f8c24ed04848127ec26d55ee784b6f4`

The merged squash tree was separately compared with its previous canonical base and contained exactly the three intended partnership documentation files.

The merged-tree comparison is separate evidence from the synthetic PR merge-tree run. No separate post-merge CI is claimed.

### 7. Audit correction made controlling

PR #55 merged as:

`e6e83dd7adc0269c16afda806c324063264e5e6e`

It created the controlling audit-evidence classification correction covering PRs #45, #47, #48, #49 and #51 and the reporting supplement that distinguishes:

- exact-head evidence;
- tested PR merge-tree evidence;
- merged-tree verification;
- post-merge CI evidence.

This report is reconciled to that controlling correction.

### 8. Issue #50 closed within exact scope

Issue #50 was closed as completed for the **generic public-safe pack only**.

Its closure does not approve named-organisation research, actual outreach or any relationship.

## Decisions

| Matter | Decision | Exact scope |
|---|---|---|
| Generic public-safe partner pack | PASS | Three documentation files only |
| Plain-English project explanation | PASS | Current intended purpose, boundaries and limitations |
| Non-binding written outreach template | PASS | Generic template only; no recipient approved |
| PR #51 workflow evidence | PASS, correctly classified | Tested synthetic PR merge `18a47d0f...` |
| PR #51 merged tree | PASS | Separate exact three-file comparison |
| Named organisations | NOT ASSESSED | No organisation identified, researched or ranked |
| Actual outreach | BLOCKED | Requires separate research and target approval |
| Publisher or partner appointment | BLOCKED | Issue #17 acceptance process remains incomplete |
| Real-data or institutional pilot | BLOCKED | Evidence-use and deployment gates remain closed |
| Issue #50 | PASS / CLOSED | Generic pack scope only |

## Evidence index

### Canonical commits

- partner-pack starting baseline: `62e00ef5e573605b2a96b445a0dc42b0f4e387bd`;
- generic partner-pack merge: `f997e1049f8c24ed04848127ec26d55ee784b6f4`;
- controlling audit correction: `e6e83dd7adc0269c16afda806c324063264e5e6e`.

### Pull requests and issues

- PR #51 — generic public-safe partner information and outreach pack, merged;
- PR #55 — audit-evidence classification correction, merged;
- issue #17 — publisher and stewardship acceptance, remains open;
- issue #22 — live Control Board;
- issue #50 — generic public-safe partner pack, closed completed;
- PR #54 — stale report proposal, to close unmerged as superseded by the current-main reconstruction.

### Workflow evidence

- PR #51 branch head: `c335733309838b8468a9fd102b456fece4f83ab3`;
- workflow run: `30893783028`;
- actual tested checkout: synthetic PR merge `18a47d0f729d4634a201a7821232f5c264b7775f`;
- all jobs passed;
- artifact inventory empty.

## Change inventory

- Application source: **none**.
- Application tests: **none**.
- Workflow: **none**.
- Executable or installer: **none created or distributed**.
- AI model or runtime: **none**.
- Dependencies: **none**.
- SBOM, notices or licence text: **none**.
- Partnership documentation: **three files**.
- Named organisations or contact records: **none**.
- Real or sensitive evidence: **none used**.

## Gate position

### Completed

- generic partner information pack;
- generic non-binding written outreach template;
- response-record and escalation controls;
- issue #50 generic scope;
- controlling audit-classification correction for PR #51 evidence.

### Still open or blocked

- named-organisation research and prioritisation;
- actual outreach;
- publisher or partner appointment;
- issue #17 organisational acceptance and operational evidence;
- real or sensitive evidence;
- signing and public executable release;
- ordinary-user testing;
- institutional, healthcare, justice-sector or EU supply;
- legal, medical, forensic, security, accuracy, accessibility and compliance claims.

## Problems and limitations

### Historical evidence was initially misclassified

The first PR #51 review described run `30893783028` as exact-head evidence.

Raw log inspection showed that GitHub tested synthetic PR merge `18a47d0f729d4634a201a7821232f5c264b7775f`.

Effect: the successful results remain valid for that tested PR merge tree. The final squash tree was separately compared. The controlling correction is now on canonical `main`.

### Generic material only

The pack describes organisation categories but contains no evidence that any named body is interested, suitable or capable.

Effect: it must not be used to imply a relationship.

### No outreach authority

The standard message exists, but no recipient, public contact route or sender role has been approved.

Effect: no message may be sent merely because the template exists.

### Current product immaturity

ECO has no approved public executable, real-evidence approval or institutional deployment.

Effect: any request for a live or real-data trial must return to the active gates.

## Next controlled actions

1. Close stale PR #54 unmerged as superseded after this current-main report is safely preserved.
2. Reconcile the live Control Board and canonical status/release-gate records.
3. Create a separate named-organisation research issue only after the public records agree.
4. Define objective assessment and exclusion criteria derived from the publisher acceptance checklist.
5. Research organisations using public sources without contacting them.
6. Stop for review before approving any outreach target or sending a message.
7. Continue issue #46 implementation conformance and issue #24 artifact controls in their appropriate lanes.

## Review statement

This report was prepared from repository records, PR #51 tested PR merge-tree evidence, raw checkout logs, artifact inventory, merged-tree comparison and the controlling PR #55 audit correction.

The same connected project account performs and posts work across project lanes. This is a control-lane record, not an organisationally independent audit, partnership approval or legal opinion.

No organisation, individual, partnership, publisher, contract, release or deployment is created by this report.
