# ECO audit evidence-classification correction — 4 August 2026

**Authority:** controlling audit correction when present on canonical `main`  
**Affected records:** PRs #45 and #47, their review comments, and `ECO-PROGRESS-20260804-LSG-001`  
**Substantive product-boundary effect:** none  
**Release effect:** none

## 1. Purpose

This record corrects the classification of review and continuous-integration evidence used for the 4 August 2026 intended-purpose and progress-reporting merges.

The original pull requests, review comments and dated report remain preserved as historical evidence. They must be read together with this correction rather than silently rewritten or deleted.

## 2. PR #45 workflow evidence

PR #45 proposed the intended-purpose and public-claims control at branch head:

`458e0c1a32d4d92db882d68374048e14711c220f`

Workflow run `30884028310` did not check out that branch head directly. The raw Windows log proves that GitHub checked out synthetic pull-request merge commit:

`47897b163e438e5fc9898d5e70efd4c4017c0976`

That synthetic commit merged head `458e0c1a32d4d92db882d68374048e14711c220f` into base `f6c852616deb96f7b71414e9526b8eca376968e5`.

The run therefore provides **tested PR merge-tree evidence**, not exact-head evidence.

The following results remain valid for the tested merge tree:

- source policy passed;
- Linux tests and `go vet` passed;
- the Windows native-failure self-test passed;
- the six-stage Windows failure matrix passed;
- ordinary Windows tests, vet, source policy and deterministic rebuild passed;
- the Actions artifact inventory was empty;
- the runner-only executable remained unsigned, private and unapproved.

## 3. PR #47 workflow evidence

PR #47 proposed the progress-reporting and audit standard at branch head:

`63c0c331cefbd649943ab39fcf7567917f14aee1`

Workflow run `30885679512` did not check out that branch head directly. The raw Windows log proves that GitHub checked out synthetic pull-request merge commit:

`8963bef36a166b916c5cbf4074fadc21d1d2db46`

That synthetic commit merged head `63c0c331cefbd649943ab39fcf7567917f14aee1` into base `d63206f64eff3eb230907e869b5d7530cb6d9f8a`.

The run therefore provides **tested PR merge-tree evidence**, not exact-head evidence.

The same bounded Linux, Windows, failure-matrix, deterministic-build and zero-artifact results remain valid for that tested merge tree.

## 4. Review relationship correction

PR #45's review and merged-tree comments used the word `independent`. The comments were posted through the same connected project GitHub account used for the governance work, and the legal/governance progress report later described the work as a control-lane review rather than an organisationally independent audit.

A GitHub username or technical posting account does not, by itself, prove or disprove independence. The controlling question is whether the actual reviewer was meaningfully separate from the implementation or evidence-generation workstream and inspected the underlying evidence.

For the original PR #45 pre-merge record, the supported description is:

**legal/governance control review, not proven organisational independence.**

A later dedicated inspection/control-room review separately inspected:

- the full five-file merged governance package;
- the raw PR workflow log and tested checkout identity;
- the empty artifact inventory;
- the relationship between the intended-purpose record, issue #20 and the current release gates;
- the review-independence and evidence-classification defects recorded here.

That later review is a separate inspection workstream, although the same technical GitHub account is used to post repository records. The shared posting identity is disclosed and must not be mistaken for the reviewer relationship itself.

## 5. Intended-purpose governance effect

The substantive intended-purpose and public-claims boundary merged in `d63206f64eff3eb230907e869b5d7530cb6d9f8a` remains accepted as the controlling official ECO governance boundary when this correction is present on canonical `main`.

The phrases in `GOVERNANCE.md` and `docs/governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md` that required an undefined `independent review` are superseded for audit-process purposes by this rule:

> A controlling governance record must be merged after a documented review whose exact scope, evidence and reviewer relationship are stated truthfully. Any separately required independent, legal, regulatory, organisational or release approval remains a distinct gate.

This correction does not establish application conformance. Issues #16, #20 and #46 remain open. It does not approve real evidence, an executable, a model, signing, a publisher, release or deployment.

## 6. Progress-report correction

The dated report `ECO-PROGRESS-20260804-LSG-001` remains a historical legal, standards and governance report.

The following statements in that report are corrected:

- references to PR #43 or PR #45 workflow runs as `exact-head` evidence must be read as tested PR merge-tree evidence unless the raw checkout proves the branch head itself;
- PR #45's review must be read as legal/governance control review, not as organisationally independent review;
- the report's statement that the 10 August artifact reminder had not been created was accurate at preparation time, but a later control action successfully repurposed a duplicate PR #11 watch into a one-time artifact recheck scheduled for 10 August 2026 at 19:00 Europe/London.

These corrections do not change the report's substantive stop gates.

## 7. Private downloadable inspection pack

The private control-room pack:

`ECO_Control_Room_Progress_and_Continuity_Pack_2026-08-04_0742_BST_v1.0.zip`

has verified size `125492` bytes and SHA-256:

`76763897af9a3b20b3e8c18c018d625d0542a1bca620f95173db0c1d48fbc7f3`

Its ZIP integrity passes and it contains seven files, including the continuity snapshot recorded in its manifest.

A delivery-surface defect was found after issue: the continuity snapshot was present inside the ZIP but absent from the loose-file directory, so the individual snapshot link failed. The exact file was restored from the ZIP without changing the ZIP or its hash. The restored file is 2055 bytes with SHA-256:

`23c48b4bfd6b8effc2141fd3e1b7cfba14be731973c4b8f6c6f0a78e7cef311f`

## 8. Future evidence-classification rule

For a GitHub `pull_request` workflow:

- record the actual checkout SHA from the raw log;
- classify the result as PR merge-tree evidence when the checkout is `refs/pull/<number>/merge` or another synthetic merge commit;
- use `exact-head` only when the raw log proves that the exact branch-head SHA was checked out and tested;
- do not infer exact-head testing merely because the workflow run is associated with a PR head;
- state separately whether the final squash or merge commit received its own CI run.

## 9. Gate position

No application, evidence-use, binary, model, signing, publisher, release or deployment gate is opened by this correction.

PR #11 remains draft, unmerged and blocked by its four workspace P0 findings. Issue #24 remains open for historical-artifact expiry or deletion and manual-dispatch proof. Issues #15, #16, #17, #20 and #46 remain open.
