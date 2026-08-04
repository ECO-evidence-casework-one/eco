# ECO audit evidence-classification correction — 4 August 2026

**Authority:** controlling correction when present on canonical `main`  
**Affected PRs:** #45, #47, #48, #49 and #51  
**Product or release effect:** none

## Purpose

This record corrects workflow-evidence and reviewer-relationship wording while preserving the original pull requests, comments and dated reports as historical evidence.

The substantive intended-purpose, publisher/stewardship and generic partner-pack decisions remain accepted within their bounded documentation/governance scope. No application-conformance, real-evidence, executable, signing, publisher, outreach, release or deployment gate is opened.

## Actual tested checkout identities

GitHub `pull_request` workflows used synthetic merge refs, not the branch heads previously described as `exact-head` evidence.

| PR | Branch head | Base | Actual tested PR merge commit | Run | Correct classification |
|---|---|---|---|---:|---|
| #45 | `458e0c1a32d4d92db882d68374048e14711c220f` | `f6c852616deb96f7b71414e9526b8eca376968e5` | `47897b163e438e5fc9898d5e70efd4c4017c0976` | `30884028310` | tested PR merge-tree evidence |
| #47 | `63c0c331cefbd649943ab39fcf7567917f14aee1` | `d63206f64eff3eb230907e869b5d7530cb6d9f8a` | `8963bef36a166b916c5cbf4074fadc21d1d2db46` | `30885679512` | tested PR merge-tree evidence |
| #48 | `408748aac12c19d5f92135ce7011bbb389034931` | `64ec3549a3a92e1c9c452ad5562480335d9f8f70` | `df7f74bfeadd71d3b6e2945a0037b92b9b4e0341` | `30892431187` | tested PR merge-tree evidence |
| #49 | `4c4b058604d3dfb8650e9249b3982daa5e0e2eb4` | `2bd53b4430d970930710f531d1012c9e65305b98` | `10b35a021d2b1d7a511b69e117598d367e1968c0` | `30893132522` | tested PR merge-tree evidence |
| #51 | `c335733309838b8468a9fd102b456fece4f83ab3` | `62e00ef5e573605b2a96b445a0dc42b0f4e387bd` | `18a47d0f729d4634a201a7821232f5c264b7775f` | `30893783028` | tested PR merge-tree evidence |

For each run, the successful Linux tests/vet, source-policy, Windows controlled-failure test, six-stage failure matrix, ordinary Windows validation and deterministic rebuild remain valid for the tested merge tree. Each inspected Actions artifact inventory was empty. The unsigned runner-built executable remained private and unapproved.

The later squash commits are separate merged-tree evidence. None of those squash commits is represented here as having received a separate post-merge CI run unless such a run is independently observed.

## Review relationship correction

The same technical GitHub account is used to post records from multiple ECO workstreams. A username alone therefore does not establish or disprove reviewer independence.

The original PR reviews for #45, #47, #48, #49 and #51 are supported as **control-lane reviews**. They are not organisationally independent audits or legal opinions.

A later dedicated inspection/control-room workstream separately inspected:

- exact changed-file scope;
- merged governance and reporting documents;
- raw workflow checkout identities;
- job results and artifact inventories;
- interaction with the active release and evidence-use gates;
- the accuracy of the reviewer-relationship wording.

That later work is described as **inspection-lane review**. The shared technical posting account is disclosed and must not be mistaken for the reviewer relationship itself.

## Substantive decision effect

### Intended purpose — PR #45

The intended-purpose record remains the official governance boundary. Issues #16, #20 and #46 remain open because governance adoption does not prove application conformance or cross-surface implementation.

### Progress reporting — PR #47

The progress-reporting standard remains useful. It is supplemented by `PROGRESS_REPORTING_AND_AUDIT_STANDARD_SUPPLEMENT_2026-08-04.md` for tested-commit identity, reviewer relationship and downloadable-pack verification.

### Publisher/stewardship — PRs #48 and #49

The publisher/stewardship gate and dated report remain accepted within governance scope. No organisation or individual is appointed. Contribution does not automatically assign organisational roles, while responsibilities arising from actual conduct or law remain preserved. Issue #17 stays open.

### Generic partner pack — PR #51

The generic public-safe partner pack remains accepted for possible later use. It identifies or contacts no organisation and authorises no research target, outreach, partnership, contracting, repository/signing transfer, real-data pilot, release or deployment. Any actual outreach requires a separate decision and control record.

## Downloadable control-room pack correction

The private pack:

`ECO_Control_Room_Progress_and_Continuity_Pack_2026-08-04_0742_BST_v1.0.zip`

has verified size `125492` bytes and SHA-256:

`76763897af9a3b20b3e8c18c018d625d0542a1bca620f95173db0c1d48fbc7f3`

ZIP integrity passes and the archive contains seven members.

A delivery-surface defect was found after issue: the continuity snapshot was present inside the ZIP and manifest but absent from the loose-file delivery directory. The exact file was restored from the ZIP without changing the ZIP or its hash.

Restored snapshot:

- size: `2055` bytes;
- SHA-256: `23c48b4bfd6b8effc2141fd3e1b7cfba14be731973c4b8f6c6f0a78e7cef311f`;
- manifest match: PASS.

## Historical report corrections

Historical reports and PR comments remain preserved. Read `exact-head` references for the PR runs above as **tested PR merge-tree evidence**.

The legal/governance report's statement that the 10 August artifact reminder had not been created was accurate at preparation time. A later control action repurposed a duplicate PR #11 watch into a one-time artifact recheck scheduled for 10 August 2026 at 19:00 Europe/London.

## Future rule

For every PR workflow, record separately:

- branch head SHA;
- base SHA;
- raw tested checkout SHA;
- whether it is a synthetic PR merge;
- workflow run and job IDs;
- artifact inventory;
- final squash/merge SHA;
- whether the final merge received post-merge CI.

Use `exact-head` only where the raw checkout proves the tested commit equals the stated head.

## Gate position

PR #11 remains draft, unmerged and blocked by four workspace P0 findings. Issues #15, #16, #17, #20, #24 and #46 remain open. No real-evidence, AI, executable, signing, publisher, outreach, release or deployment gate is opened by this correction.
