# ECO progress reporting and audit standard

**Authority:** controlling only when present on canonical `main`  
**Purpose:** preserve understandable, dated and reconstructable records of material ECO work  
**Release effect:** none by itself

## 1. Why this standard exists

ECO work creates technical evidence in commits, pull requests, issues, workflow runs and review comments. Those records are necessary but can become difficult to understand when read separately.

This standard requires a second layer: concise progress reports that explain what changed, why it mattered, what evidence supports it, what remains blocked and what happens next.

The objective is not to create paperwork for every action. The objective is to ensure that a later reviewer can reconstruct material decisions without relying on memory or reading every conversation.

## 2. Record hierarchy

The following records serve different purposes:

1. **Canonical status** — `CURRENT_STATUS.md` records the current public project position.
2. **Release gate** — `docs/control/CURRENT_RELEASE_GATE.md` records the current release decision.
3. **Live Control Board** — issue #22 records active lanes, current blockers and next actions.
4. **Event evidence** — issues, pull requests, commits, reviews, workflow runs and artifact inventories preserve exact event-level evidence.
5. **Dated progress reports** — `docs/control/progress/` preserves readable historical snapshots.
6. **Controlling governance records** — topic-specific records define durable policy and boundaries.

A progress report must not override a controlling issue, release gate, governance record or exact technical evidence. Where records conflict, the more specific and more recently approved controlling record applies until the conflict is corrected.

## 3. When a progress report is required

Create a dated report when:

- a day contains substantial ECO development, inspection, legal, standards or governance work;
- canonical `main` changes materially;
- a P0 or P1 blocker is created, removed or materially redefined;
- an important PR is opened, frozen, reviewed, merged, closed or superseded;
- a release, evidence-use, signing, distribution or deployment gate changes;
- official legal, regulatory or standards material changes a project requirement;
- a significant incident, containment action or failed control action occurs;
- a chat, tool or process failure creates a continuity risk;
- a milestone, candidate or formal work package is completed or rejected.

Do not create repetitive reports merely to state that nothing changed. A no-change record is appropriate only when a scheduled control check was itself required and the negative result matters.

## 4. Report timing

- Use the actual preparation time.
- Record **BST and CEST** while those are the project’s active working time zones.
- Include UTC when the source evidence is an automated service recorded in UTC.
- Use exact calendar dates. Do not rely only on words such as today or yesterday.
- A daily report may be issued during the day and supplemented by a closing report if further material work occurs.

## 5. Mandatory report fields

Every material progress report must state:

1. report title and unique report identity;
2. preparation date and exact time zones;
3. reporting lane and purpose;
4. repository baseline at the start of the reported work;
5. repository state at the end of the report;
6. work completed;
7. decisions made and their scope;
8. evidence references, including relevant issues, PRs, commits and workflow runs;
9. application, binary, model and dependency changes, including an explicit `none` where applicable;
10. gates opened, closed or unchanged;
11. unresolved blockers, risks and dependencies;
12. failed or unavailable actions and their effect;
13. next controlled actions;
14. review status and reviewer relationship;
15. correction history where the report supersedes or corrects another record.

## 6. Status vocabulary

Use the following terms consistently:

- **PASS** — the stated bounded acceptance test was supported by inspected evidence.
- **HOLD** — work may be promising but must not proceed until stated evidence or correction exists.
- **BLOCKED** — a defined condition prevents the activity or gate from proceeding.
- **NOT ASSESSED** — the matter was outside the review scope or evidence was unavailable.
- **NO CHANGE** — a previously recorded position remains unchanged after a required check.
- **SUPERSEDED** — a later controlling record replaces the earlier decision or approach.
- **REJECTED** — the candidate or proposal failed the stated acceptance basis.

A PASS must name its exact scope. It must not be presented as approval of unrelated application, release or compliance matters.

## 7. Evidence and traceability rules

- Use full commit SHAs for frozen or merged decisions.
- Record PR and issue numbers.
- Record workflow run and job identifiers when CI evidence is relied upon.
- Record whether the artifact inventory was empty.
- Do not treat a badge, summary or green status alone as sufficient evidence.
- Distinguish exact-head evidence from merged-tree verification.
- State when a squash commit did not receive separate CI and which exact-head evidence is being relied upon.
- Do not download or execute prohibited artifacts merely to improve the report.
- Do not include real evidence, private diagnostics, credentials or identifying case information.

## 8. Review independence

Every report must state who prepared it and how it was reviewed.

The word **independent** may be used only where the reviewer is meaningfully separate from the implementation or evidence-generation lane and has inspected the relevant evidence rather than relying only on the preparer’s summary.

Where the same person, account, agent or workstream prepared and reviewed the record, describe it as `self-review`, `control review` or `merged-tree verification` as applicable. Do not use independence as a decorative label.

Codex or another automated review is supporting evidence, not a substitute for the required human or control-lane decision. An unavailable automated review must be recorded but does not automatically block or approve work unless a controlling rule expressly requires it.

## 9. Corrections and preservation

Dated progress reports are historical records.

After merge:

- do not silently rewrite a material conclusion;
- correct an error through a clearly identified corrective commit or a new report;
- state the original report, the incorrect statement, the corrected statement, the reason and the effect on prior decisions;
- preserve the earlier Git history;
- do not delete an adverse or superseded record merely because the project later improved.

The live Control Board may be updated in place because it is expressly a current-state summary. Its history is preserved by GitHub, while dated reports preserve readable snapshots.

## 10. Daily report structure

Use this order:

1. **Executive meaning** — what the work means for ECO in plain English.
2. **Mission alignment** — why the work belongs in the lane.
3. **Repository identity** — baseline and ending commit.
4. **Completed work** — bounded factual record.
5. **Decisions** — PASS, HOLD, BLOCKED, SUPERSEDED or NO CHANGE.
6. **Evidence index** — issues, PRs, commits, runs and comments.
7. **Gate position** — what remains closed or changed.
8. **Problems and failed actions** — including tool or access limitations.
9. **Next actions** — ordered and controlled.
10. **Review status** — preparer, reviewer relationship and limitations.

## 11. Progress-report template

```markdown
# ECO progress report — <lane and date>

**Report ID:** <unique identifier>  
**Prepared:** <date and BST/CEST time>  
**Lane:** <development / inspection / legal-standards / governance / release>  
**Baseline:** <full commit SHA>  
**Ending state:** <full commit SHA or no repository change>  
**Review status:** <prepared / self-reviewed / control-reviewed / independently reviewed>

## Executive meaning

<Plain-English significance.>

## Mission alignment

<Why this work belongs in the lane and what was deliberately not done.>

## Work completed

<Material actions only.>

## Decisions

| Matter | Decision | Exact scope |
|---|---|---|
| ... | PASS/HOLD/BLOCKED/... | ... |

## Evidence index

- Issue #...
- PR #...
- commit `...`
- workflow run `...`

## Change inventory

- Application source: none / describe
- Workflow: none / describe
- Executable or model: none / describe
- Dependencies or licences: none / describe
- Governance or documentation: none / describe

## Gate position

<Opened, closed and unchanged gates.>

## Problems, limitations and failed actions

<What could not be done and whether it affects the decision.>

## Next controlled actions

1. ...

## Review statement

<Who prepared and reviewed the report, relationship to the work and any limits.>
```

## 12. Audit quality checks

Before merging a progress report, verify:

- timestamps and time zones are correct;
- commit and PR identities match the repository;
- report scope matches the actual changed files;
- no approval is broader than the supporting evidence;
- failed actions are not described as passes;
- superseded records remain discoverable;
- all public and real-evidence stop gates are stated accurately;
- no personal or sensitive evidence is included;
- the report is understandable without specialist knowledge.

## 13. Current application

This standard is introduced because the live Control Board had become stale while detailed evidence was spread across issues, PRs and comments.

The Control Board was refreshed on 4 August 2026. The first dated report under this standard records the legal, standards and governance work leading to canonical `main` `d63206f64eff3eb230907e869b5d7530cb6d9f8a`.
