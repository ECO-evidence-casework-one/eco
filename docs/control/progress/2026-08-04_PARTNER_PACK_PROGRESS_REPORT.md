# ECO progress report — public-safe partner pack — 4 August 2026

**Report ID:** `ECO-PROGRESS-20260804-PPK-001`  
**Prepared:** 4 August 2026, approximately 09:58 BST / 10:58 CEST  
**Lane:** legal, governance and public-safe organisational exploration  
**Starting canonical baseline:** `62e00ef5e573605b2a96b445a0dc42b0f4e387bd`  
**Canonical ending state:** `f997e1049f8c24ed04848127ec26d55ee784b6f4`  
**Review status:** control-lane report; no claim of organisational independence, partnership authority or legal advice

## Executive meaning

ECO now has a generic, public-safe information pack that can explain the project to established organisations without implying that any partnership, publisher or endorsement already exists.

The pack protects the project's core boundaries and the originating developer's position. It requests only a non-binding written discussion and does not authorise named-organisation research, actual outreach, real-data pilots, contracts or transfer of repository or signing authority.

No organisation was named, researched, ranked, contacted or appointed.

## Mission alignment

This work followed the publisher and stewardship governance merged through PR #48.

It belonged in the legal and governance lane because public outreach can accidentally:

- imply a partnership or endorsement;
- make unsupported product or compliance claims;
- invite real evidence or premature pilots;
- weaken offline, open-source or user-data non-custody requirements;
- place organisational duties on an individual contributor;
- create apparent commitments before governing-body approval.

The pack creates controls against those risks before any named-organisation work begins.

## Work completed

### 1. Issue #50 created

Issue #50 defined the required generic pack and prohibited:

- naming or ranking organisations;
- actual outreach;
- partnership or publisher claims;
- real evidence and real-data pilots;
- entity formation;
- contracts, fundraising or institutional supply;
- product or release approval.

### 2. Partner information pack drafted

`docs/partnership/PARTNER_INFORMATION_PACK.md` explains:

- what ECO is;
- the public-interest problem it intends to address;
- fully offline operation;
- free and open-source requirements;
- no accounts, telemetry, advertising or routine access to user files;
- one-file Windows and low-spec objectives;
- permitted evidence assistance and prohibited medical, legal, forensic, scoring and official-decision uses;
- current source-development and release limitations;
- why an existing established organisation is preferred;
- what ECO seeks and does not seek;
- staged exploration, due diligence, governing-body acceptance, operational validation and separate release approval;
- principal public repository records;
- questions suitable for a first non-binding discussion.

### 3. Outreach template drafted

`docs/partnership/EXPLORATORY_OUTREACH_TEMPLATE.md` provides:

- a pre-send checklist;
- a standard written message;
- a short-form written message;
- a response-record template;
- prohibited outreach wording;
- escalation stops.

The template requests only a non-binding written discussion and expressly states that ECO has no approved public executable, real-evidence approval or appointed publisher.

### 4. Partnership usage controls added

`docs/partnership/README.md` records:

- no current partner or publisher;
- controlling governance and release records;
- the required sequence before named outreach;
- prohibited uses of the pack;
- the rule that completion opens no product, evidence-use, signing, release or deployment gate.

### 5. Exact-head validation completed

Draft PR #51 froze exact head:

`c335733309838b8468a9fd102b456fece4f83ab3`

Workflow run `30893783028` passed:

- source policy;
- Linux tests and vet;
- controlled Windows failure self-test;
- six-stage Windows failure matrix;
- deterministic Windows build and tests.

The run's artifact inventory was exactly empty.

A control-lane review recorded a bounded PASS for the three-document generic pack.

### 6. PR #51 merged and verified

PR #51 merged as:

`f997e1049f8c24ed04848127ec26d55ee784b6f4`

Merged-tree comparison confirmed:

- exactly one squash commit;
- exactly three new partnership documentation files;
- no application source, tests, workflow, `VERSION`, executable, model, dependency, SBOM, notices, licence text or evidence fixture changes.

### 7. Issue #50 closed within its exact scope

Issue #50 was closed as completed for the **generic public-safe pack only**.

Its closing record states that named-organisation research and actual outreach require a later separate issue and approval decision.

## Decisions

| Matter | Decision | Exact scope |
|---|---|---|
| Generic public-safe partner pack | PASS | Three documentation files only |
| Plain-English project explanation | PASS | Current intended purpose, boundaries and limitations |
| Non-binding written outreach template | PASS | Generic use only; no named recipient approved |
| Named organisations | NOT ASSESSED | No organisation identified, researched or ranked |
| Actual outreach | BLOCKED | Requires separate target research and approval |
| Partnership or publisher appointment | BLOCKED | Issue #17 acceptance process remains incomplete |
| Real-data or institutional pilot | BLOCKED | Current evidence-use and deployment gates remain closed |
| Issue #50 | PASS / CLOSED | Generic pack scope only |

## Evidence index

### Canonical commits

- starting canonical baseline: `62e00ef5e573605b2a96b445a0dc42b0f4e387bd`;
- generic partner pack merge: `f997e1049f8c24ed04848127ec26d55ee784b6f4`.

### Pull requests

- PR #51 — generic public-safe partner information and outreach pack, merged.

### Issues

- #17 — publisher and stewardship acceptance, remains open;
- #22 — live Control Board;
- #50 — generic public-safe partner pack, closed completed.

### Workflow evidence

- PR #51 exact-head run `30893783028` — all jobs passed; artifact inventory empty.

## Change inventory

- Application source: **none**.
- Application tests: **none**.
- Workflow: **none**.
- Executable or installer: **none created or distributed**.
- AI model or runtime: **none**.
- Dependencies: **none**.
- SBOM, notices or licence text: **none**.
- Documentation: **three new public-safe partnership files**.
- Named organisations or contact records: **none**.
- Real or sensitive evidence: **none used**.

## Gate position

### Newly completed

- generic partner information pack;
- generic non-binding written outreach template;
- response-record and escalation controls;
- issue #50 generic scope.

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

### Generic material only

The pack describes objective organisation categories but contains no evidence that any named body is interested, suitable or capable.

Effect: it must not be used to imply a relationship. Named research requires a separate evidence-based process.

### No outreach authority

The standard message exists, but no recipient, public contact route or sender role has been approved.

Effect: no message should be sent merely because the template has merged.

### Current product immaturity

The pack accurately states that ECO has no approved public executable, no real-evidence approval and no institutional deployment.

Effect: any organisation requesting a live or real-data trial must be referred back to the active gates.

## Next controlled actions

1. Create a separate named-organisation research issue.
2. Define objective scoring and exclusion criteria derived from the publisher acceptance checklist.
3. Research organisations using current public sources without contacting them.
4. Record evidence for mission fit, governance capability, open-source compatibility and conflicts.
5. Stop for review before approving any outreach target or sending any message.
6. Continue issue #46 implementation conformance and issue #24 artifact controls in their appropriate lanes.

## Review statement

This report was prepared from repository records, PR #51 exact-head CI, artifact inventory and merged-tree comparison.

The same connected project account performed the work and prepared this report. It is a control-lane record, not an organisationally independent audit, partnership approval or legal opinion.

No organisation, individual, partnership, publisher, contract, release or deployment is created by this report.