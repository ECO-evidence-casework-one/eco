# ECO progress report — inspection and control room — 4 August 2026

**Report ID:** `ECO-PROGRESS-20260804-INSPECTION-001`  
**Prepared:** 4 August 2026, 09:45 BST / 10:45 CEST  
**Lane:** independent inspection, repository control and release-gate assurance  
**Starting reviewed baseline:** `f6c852616deb96f7b71414e9526b8eca376968e5`  
**Intervening canonical `main` before this correction:** `64ec3549a3a92e1c9c452ad5562480335d9f8f70`  
**Correction branch:** `control/inspection-audit-corrections-20260804`  
**Review status:** inspection-lane report; the same technical GitHub account is used for repository posting, but this inspection workstream did not create PR #45 or PR #47

## Executive meaning

ECO remains blocked from real evidence, ordinary-user executable testing and release. No application source changed during this inspection cycle.

Two documentation-only merges materially improved intended-purpose governance and progress reporting. Their substantive boundaries and stop gates are sound, but the audit record overstated two things:

1. pull-request workflow runs were described as `exact-head` evidence even though GitHub tested synthetic PR merge commits;
2. PR #45's same-account legal/governance control review was described as independent review without enough evidence of organisational independence.

This report preserves those findings, corrects the terminology and records the separate inspection-lane review. It also records the repair of a missing loose-file link in the private downloadable control-room pack.

## Mission alignment

This work belongs in the inspection/control-room lane because it concerns:

- exact tested commit identity;
- reviewer relationship and assurance terminology;
- integrity of progress and audit records;
- release-gate truthfulness;
- continuity-pack verification;
- protection of the blocked PR #11 implementation lane.

This lane did not change ECO application code, tests, workflow, `VERSION`, AI models, dependencies, licence text, executable packaging or evidence fixtures.

## Repository identity and lane state

At the start of this inspection supplement, canonical `main` was:

`f6c852616deb96f7b71414e9526b8eca376968e5`

Two later documentation-only merges advanced `main`:

- intended-purpose governance: `d63206f64eff3eb230907e869b5d7530cb6d9f8a`;
- progress-reporting and audit records: `64ec3549a3a92e1c9c452ad5562480335d9f8f70`.

PR #11 remained unchanged at:

`61a2004809b341e72f70321843c64c3ff477f549`

Against `main` `64ec3549...`, PR #11 was 8 commits ahead and 36 commits behind, with common base:

`2dcb44ba8541ab7a319de6e1f14f016aafe2ac1b`

Its four workspace P0 findings remained unchanged.

## Work completed

### 1. PR #45 intended-purpose governance inspection

The complete five-file governance package was inspected after merge.

The substantive boundary remains suitable for official ECO governance:

- source-linked user-controlled assistance with health, legal and other sensitive evidence is permitted in principle;
- diagnosis, treatment, prognosis, triage, monitoring and clinical-risk outputs are prohibited or separately gated;
- reserved legal activities and professional representation are prohibited;
- forensic, authenticity and admissibility guarantees are prohibited;
- profiling, scoring and authority-side adverse decisions are prohibited or require a new assessment;
- governance adoption is separated from application conformance, legal classification, role appointment and release approval;
- GPL and independent-redistribution rights remain distinct from official ECO endorsement.

Issues #16, #20 and #46 remain open. The governance merge does not approve application conformance, real evidence, a publisher, a binary or deployment.

### 2. PR #45 CI evidence correction

PR #45 branch head:

`458e0c1a32d4d92db882d68374048e14711c220f`

Workflow run `30884028310` checked out synthetic PR merge commit:

`47897b163e438e5fc9898d5e70efd4c4017c0976`

The raw log states that the synthetic commit merged the PR head into base `f6c852616deb96f7b71414e9526b8eca376968e5`.

Decision: the run is **tested PR merge-tree evidence**, not exact-head evidence.

The bounded results remain valid:

- source policy PASS;
- Linux tests and vet PASS;
- Windows controlled-failure self-test PASS;
- six-stage failure matrix PASS;
- ordinary Windows tests, vet, policy and deterministic rebuild PASS;
- artifact inventory empty;
- runner-only unsigned executable remained private and unapproved.

### 3. PR #47 reporting-standard inspection

PR #47 branch head:

`63c0c331cefbd649943ab39fcf7567917f14aee1`

Workflow run `30885679512` checked out synthetic PR merge commit:

`8963bef36a166b916c5cbf4074fadc21d1d2db46`

Decision: this is also **tested PR merge-tree evidence**, not exact-head evidence.

The reporting standard is materially useful and preserves adverse findings, failed actions, correction history and reviewer relationships. A supplement was required to make three controls explicit:

- raw checkout identity controls the evidence classification;
- a shared technical posting account does not by itself determine independence;
- substantial control-room reports require verified downloadable packs where appropriate.

### 4. Reviewer-relationship correction

PR #45's original review comments used the word `independent`, but the same connected project account posted both the governance work and the review record.

The supported original description is legal/governance control review, not proven organisational independence.

This later inspection-lane review was performed in a separate control-room workstream that did not create PR #45 or PR #47. The same technical GitHub account is used to post repository records and that limitation is disclosed.

### 5. Downloadable progress-pack verification

The private pack:

`ECO_Control_Room_Progress_and_Continuity_Pack_2026-08-04_0742_BST_v1.0.zip`

was rechecked.

- size: `125492` bytes;
- SHA-256: `76763897af9a3b20b3e8c18c018d625d0542a1bca620f95173db0c1d48fbc7f3`;
- ZIP integrity: PASS;
- ZIP members: seven.

A delivery-surface defect was found: the continuity snapshot existed inside the ZIP and manifest but was missing from the loose-file delivery directory, so the individual snapshot link failed.

The exact snapshot was restored from the ZIP without changing the ZIP or its hash.

- restored file size: `2055` bytes;
- restored SHA-256: `23c48b4bfd6b8effc2141fd3e1b7cfba14be731973c4b8f6c6f0a78e7cef311f`;
- manifest match: PASS.

### 6. Artifact-expiry reminder correction

The legal/governance report correctly recorded that the requested 10 August task had not been created at its preparation time because all active task slots were occupied.

A later task audit found two duplicate active PR #11 watches. The older duplicate was repurposed into a one-time artifact recheck scheduled for 10 August 2026 at 19:00 Europe/London.

The recheck is prohibited from downloading, executing, testing or redistributing the artifacts.

## Decisions

| Matter | Decision | Exact scope |
|---|---|---|
| PR #45 substantive intended-purpose boundary | PASS | Governance boundary only; no implementation or release approval |
| PR #45 original `independent` label | CORRECTED | Supported as legal/governance control review; later inspection-lane review recorded separately |
| PR #45 run `30884028310` | CORRECTED / PASS | Tested PR merge-tree CI and empty artifact inventory, not exact-head CI |
| PR #47 reporting standard | PASS WITH SUPPLEMENT | Useful standard; evidence-classification, reviewer-relationship and downloadable-pack controls supplemented |
| PR #47 run `30885679512` | CORRECTED / PASS | Tested PR merge-tree CI and empty artifact inventory, not exact-head CI |
| Private 07:42 BST ZIP pack | PASS | ZIP, members, manifest and hash verified |
| Loose continuity-snapshot link | CORRECTED | Exact ZIP member restored and hash verified |
| 10 August artifact recheck | PASS / SCHEDULED | One-time 19:00 Europe/London task created by repurposing a duplicate watch |
| PR #11 workspace implementation | BLOCKED | No new implementation; four P0 findings remain |
| Public binary, real evidence and deployment | BLOCKED | No gate changed |

## Evidence index

### Commits

- `d63206f64eff3eb230907e869b5d7530cb6d9f8a` — intended-purpose governance merge;
- `64ec3549a3a92e1c9c452ad5562480335d9f8f70` — progress-reporting merge;
- PR #11 head `61a2004809b341e72f70321843c64c3ff477f549`.

### Pull requests

- PR #45 — intended purpose and public claims;
- PR #47 — progress reporting and audit records;
- PR #11 — blocked workspace lane.

### Workflow evidence

- PR #45 run `30884028310`, Windows job `91911334234`, tested checkout `47897b163e438e5fc9898d5e70efd4c4017c0976`, artifact inventory empty;
- PR #47 run `30885679512`, Windows job `91916338275`, tested checkout `8963bef36a166b916c5cbf4074fadc21d1d2db46`, artifact inventory empty.

### Control records

- `docs/governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md`;
- `docs/control/PROGRESS_REPORTING_AND_AUDIT_STANDARD.md`;
- `docs/control/AUDIT_EVIDENCE_CLASSIFICATION_CORRECTION_2026-08-04.md`;
- `docs/control/PROGRESS_REPORTING_AND_AUDIT_STANDARD_SUPPLEMENT_2026-08-04.md`;
- issue #22 Control Board;
- issues #16, #20, #24 and #46.

## Change inventory

- Application source: **none**.
- Application tests: **none**.
- Workflow: **none**.
- `VERSION`: **none**.
- Executable or model: **none distributed or approved**.
- Dependencies or licences: **none**.
- Governance and audit documentation: **correction package only**.
- Real or sensitive evidence: **none used**.

## Gate position

No release or evidence-use gate opened.

Still blocked:

- PR #11 and issue #4;
- real, sensitive or irreplaceable evidence;
- current Ask ECO and public AI-assisted evidence functions;
- public executable distribution and signing;
- issue #15 final provenance;
- issue #16 and issue #46 cross-surface conformance;
- issue #20 implementation qualification;
- issue #17 publisher and continuity;
- accessibility, institutional, healthcare, justice-sector and EU deployment.

## Problems, limitations and failed actions

- Codex code-review capacity remained unavailable for PRs #45 and #47. No Codex review is claimed.
- The connected GitHub interface still does not expose deletion of historical Actions artifacts or creation of a fresh manual workflow dispatch.
- Canonical `CURRENT_STATUS.md` and `CURRENT_RELEASE_GATE.md` still identify an earlier reviewed `main` in their headers. Their substantive stop gates remain current, but a bounded canonical reconciliation is required after this correction package.
- PR #11 is now 8 commits ahead and 36 commits behind `main` `64ec3549...`; it must not be reconciled by a wholesale ours/theirs choice.

## Next controlled actions

1. Review this correction branch's exact changed files and PR merge-tree CI.
2. Merge only if the audit corrections are internally consistent and the artifact inventory is empty.
3. Issue a bounded canonical status/release-gate reconciliation against the resulting `main`.
4. Update PR #11's reconciliation map to the new canonical `main` without changing its branch.
5. Await a genuine workspace-correction SHA before resuming application source review.
6. Recheck historical artifact expiry on 10 August 2026 through the scheduled task.

## Review statement

This report was prepared in the dedicated ECO inspection/control-room lane. That lane did not create PR #45 or PR #47 and inspected their merged files, raw Windows logs, tested checkout SHAs and artifact inventories.

The same technical GitHub account is used to post records across project lanes. This report therefore discloses the posting identity and bases its reviewer relationship on actual workstream separation rather than the username alone.

This is not a legal opinion, organisational certification, release approval or professional assurance.
