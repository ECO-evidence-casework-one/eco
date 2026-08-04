# ECO progress report — inspection and control room — 4 August 2026

**Report ID:** `ECO-PROGRESS-20260804-INSPECTION-001`  
**Prepared:** 4 August 2026, 10:13 BST / 11:13 CEST  
**Lane:** independent inspection, repository control and release-gate assurance  
**Canonical `main` reviewed:** `f997e1049f8c24ed04848127ec26d55ee784b6f4`  
**Application PR:** #11 at `61a2004809b341e72f70321843c64c3ff477f549`  
**Reviewer relationship:** inspection-lane review; the same technical GitHub account posts records across project lanes

## Executive position

ECO remains blocked from real evidence, ordinary-user executable testing and release. No application source changed during this inspection cycle.

The inspection reviewed documentation/governance merges for intended purpose, progress reporting, publisher/stewardship and a generic public-safe partner pack. Their substantive bounded decisions remain acceptable. Their PR workflow evidence was repeatedly labelled `exact-head` even though raw logs prove GitHub tested synthetic PR merge commits.

A consolidated audit correction and reporting supplement have therefore been prepared. The original PR comments and reports remain preserved as historical evidence.

## Mission alignment

This work belongs in the inspection/control-room lane because it concerns:

- exact tested commit identity;
- reviewer relationship and assurance terminology;
- integrity of progress and audit records;
- continuity-pack verification;
- release-gate truthfulness;
- protection of the blocked PR #11 implementation lane.

No application source, tests, workflow, `VERSION`, model, dependency, licence, executable or evidence fixture changed.

## Repository changes inspected

The following documentation-only merges were reviewed:

| PR | Subject | Branch head | Merge commit |
|---|---|---|---|
| #45 | intended purpose and public claims | `458e0c1a32d4d92db882d68374048e14711c220f` | `d63206f64eff3eb230907e869b5d7530cb6d9f8a` |
| #47 | progress reporting and audit records | `63c0c331cefbd649943ab39fcf7567917f14aee1` | `64ec3549a3a92e1c9c452ad5562480335d9f8f70` |
| #48 | publisher and stewardship gate | `408748aac12c19d5f92135ce7011bbb389034931` | `2bd53b4430d970930710f531d1012c9e65305b98` |
| #49 | publisher/stewardship progress report | `4c4b058604d3dfb8650e9249b3982daa5e0e2eb4` | `62e00ef5e573605b2a96b445a0dc42b0f4e387bd` |
| #51 | generic public-safe partner pack | `c335733309838b8468a9fd102b456fece4f83ab3` | `f997e1049f8c24ed04848127ec26d55ee784b6f4` |

## Workflow evidence correction

Raw checkout logs prove the actual tested commits were synthetic PR merge trees:

| PR | Run | Actual tested checkout | Result |
|---|---:|---|---|
| #45 | `30884028310` | `47897b163e438e5fc9898d5e70efd4c4017c0976` | full gate PASS; zero artifacts |
| #47 | `30885679512` | `8963bef36a166b916c5cbf4074fadc21d1d2db46` | full gate PASS; zero artifacts |
| #48 | `30892431187` | `df7f74bfeadd71d3b6e2945a0037b92b9b4e0341` | full gate PASS; zero artifacts |
| #49 | `30893132522` | `10b35a021d2b1d7a511b69e117598d367e1968c0` | full gate PASS; zero artifacts |
| #51 | `30893783028` | `18a47d0f729d4634a201a7821232f5c264b7775f` | full gate PASS; zero artifacts |

Correct classification: **tested PR merge-tree evidence**, not exact-head CI.

The successful job outcomes remain valid for those tested trees. The later squash comparisons remain separate merged-tree evidence. No separate post-merge CI run is claimed unless independently observed.

## Substantive governance decisions

### Intended purpose

The intended-purpose boundary remains suitable for official ECO governance:

- source-linked, user-controlled assistance with supplied sensitive evidence is permitted in principle;
- diagnosis, treatment, prognosis, triage, clinical-risk and emergency assessment are prohibited;
- reserved legal activities and professional representation are prohibited;
- forensic authenticity/admissibility guarantees are prohibited;
- profiling, scoring and authority-side adverse decisions are prohibited or separately gated.

Issues #16, #20 and #46 remain open because governance adoption does not prove implementation conformance.

### Publisher and stewardship

The publisher/stewardship gate remains accepted in governance scope:

- no organisation or individual is appointed;
- Karl or another contributor is not required to form a company, CIC, charity or other entity;
- contribution and repository administration do not automatically assign organisational duties;
- responsibilities arising from actual publishing, signing, contracting, data processing, claims or law remain preserved;
- the preferred route is an established compatible organisation;
- GPL rights remain separate from official-project endorsement controls;
- issue #17 remains open until a named organisation formally accepts and operationally proves the duties.

### Generic partner pack

The generic partner pack is accepted for possible later use only:

- no organisation is named, ranked or contacted;
- no relationship, appointment, endorsement or certification is implied;
- outreach language is non-binding and written-only;
- real evidence, pilots, contracting, repository/signing transfer and release are excluded;
- actual named-organisation research or contact requires a separate decision and control record.

PR #54 remains draft and on hold pending this broader correction.

## Reviewer relationship

The original reviews are supported as control-lane reviews, not organisationally independent audits or legal opinions.

This later inspection/control-room workstream did not create PRs #45, #47, #48, #49 or #51 and inspected the underlying files, raw logs and artifact inventories. The same technical GitHub account is used to post records across lanes; that fact is disclosed and must not be mistaken for the reviewer relationship itself.

## Downloadable pack verification

Private pack:

`ECO_Control_Room_Progress_and_Continuity_Pack_2026-08-04_0742_BST_v1.0.zip`

- size: `125492` bytes;
- SHA-256: `76763897af9a3b20b3e8c18c018d625d0542a1bca620f95173db0c1d48fbc7f3`;
- ZIP integrity: PASS;
- members: seven.

The continuity snapshot was present inside the ZIP and manifest but missing from the loose delivery directory. The exact member was restored without changing the ZIP.

- restored size: `2055` bytes;
- restored SHA-256: `23c48b4bfd6b8effc2141fd3e1b7cfba14be731973c4b8f6c6f0a78e7cef311f`;
- manifest match: PASS.

## Incident and hold record

A first audit-correction branch and PR #52 were created from `64ec3549...`. Main then advanced through PRs #48, #49 and #51, making PR #52 stale and non-mergeable.

A temporary governance/documentation merge hold was recorded on issue #22 at 10:00 BST. The hold prevents further governance, partner, outreach or progress-report merges until this consolidated correction is based on stable `main` and reviewed.

PR #52 must close unmerged as superseded. The replacement branch is based on `f997e1049f8c24ed04848127ec26d55ee784b6f4`.

## PR #11 continuity

PR #11 remains unchanged at `61a2004809b341e72f70321843c64c3ff477f549`, draft, unmerged and non-mergeable.

Its four P0 blockers remain:

1. stale concurrent writers;
2. alias-bypassable and split ownership;
3. Linux nested cleanup reopening by name;
4. unowned creation, first launch and candidate state.

No repeated source review was performed because no new implementation SHA exists.

## Decisions

| Matter | Decision |
|---|---|
| PRs #45/#47/#48/#49/#51 substantive bounded documentation | PASS within stated scope |
| Their `exact-head` workflow wording | CORRECTED to tested PR merge-tree evidence |
| Original same-account reviews | control review, not organisationally independent audit |
| Publisher/stewardship gate | governance PASS; issue #17 remains open |
| Generic partner pack | documentation PASS; no actual outreach authorised |
| Private 07:42 BST ZIP | PASS |
| Loose snapshot delivery | corrected and hash verified |
| PR #52 | stale; close unmerged as superseded |
| PR #54 | HOLD pending consolidated correction |
| PR #11 | BLOCKED |
| Real evidence, binary, signing, release, deployment | BLOCKED |

## Problems and unavailable actions

- Codex review capacity was unavailable and is not treated as approval.
- The connected interface still does not expose deletion of historical Actions artifacts or creation of a fresh manual workflow dispatch.
- Canonical `CURRENT_STATUS.md`, release-gate records and PR #11 metadata require reconciliation after this correction merges.
- Rapid parallel governance merges created moving-base risk; the temporary hold addresses that control failure.

## Next controlled actions

1. freeze and inspect the replacement correction PR;
2. verify its exact changed-file scope;
3. record its actual tested PR merge checkout SHA and empty artifact inventory;
4. merge only after no new hold or conflicting main movement;
5. close PR #52 unmerged as superseded;
6. reconcile PR #54, canonical status/release gate and PR #11 metadata;
7. lift the documentation/governance hold only after those records agree;
8. await a genuine PR #11 workspace-correction SHA.

## Gate position

No application, real-evidence, AI, executable, signing, publisher, outreach, release or deployment gate is opened by this report.
