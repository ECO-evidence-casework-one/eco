# ECO progress reporting and audit standard supplement — 4 August 2026

**Authority:** controlling supplement to `PROGRESS_REPORTING_AND_AUDIT_STANDARD.md` when present on canonical `main`  
**Reason:** correct PR workflow evidence classification, clarify reviewer relationships and formalise downloadable audit-pack controls  
**Release effect:** none

This supplement applies together with [`PROGRESS_REPORTING_AND_AUDIT_STANDARD.md`](PROGRESS_REPORTING_AND_AUDIT_STANDARD.md). Where the earlier standard is silent or inconsistent on the subjects below, this more specific supplement controls.

## 1. Tested commit identity

A report must identify the exact Git commit actually checked out and tested.

For GitHub Actions triggered by `pull_request`, the default checkout commonly uses a synthetic commit under `refs/pull/<number>/merge`. `GITHUB_SHA` may therefore identify the synthetic merge tree rather than the branch head.

Use these terms:

- **exact-head evidence** — the raw checkout log proves the tested commit is exactly the stated branch-head SHA;
- **tested PR merge-tree evidence** — the raw checkout log proves a synthetic PR merge commit was tested;
- **merged-tree verification** — the final merged commit or squash tree was compared with the accepted proposal and controlling base;
- **post-merge CI evidence** — a workflow actually ran against the final merged commit.

Do not call a PR workflow `exact-head` merely because it is associated with the PR head. Record:

- branch head SHA;
- base SHA;
- tested checkout SHA;
- whether the checkout was a synthetic merge;
- workflow run and job IDs;
- final merge or squash SHA;
- whether the final merge received its own CI run.

## 2. Reviewer relationship and posting identity

Independence is determined by the actual relationship between the preparer, implementation or evidence-generation workstream and the reviewer.

A shared GitHub username, service account or technical posting account does not by itself prove or defeat independence. The record must identify the real reviewing lane, person or agent and whether that reviewer participated in creating the work or evidence being assessed.

Use:

- **self-review** when the same underlying person or agent prepared and reviewed the work;
- **control review** when a control lane reviews its own or closely connected governance and audit work;
- **inspection-lane review** when a dedicated inspection workstream that did not create the implementation or proposal examines the underlying evidence;
- **organisationally independent review** only where a meaningfully separate person or organisation performs the review.

Where one technical account posts records for multiple lanes, disclose that fact. Do not infer reviewer separation from the posting account alone.

## 3. Downloadable audit packs

A downloadable audit pack is required when:

- the user expressly requests a progress, continuity or audit report;
- a substantial inspection/control-room cycle reaches a material decision or handoff;
- a chat, tool or process failure creates a recovery risk;
- an incident, rollback, premature merge or public-distribution containment action requires a durable bundle;
- the report contains enough evidence that chat text and GitHub comments alone would be difficult to reconstruct.

A standard full pack should include, where applicable:

- fixed-layout PDF report;
- editable DOCX report;
- event log CSV;
- source register CSV;
- current-state continuity snapshot;
- SHA-256 manifest;
- pack receipt;
- ZIP archive;
- adjacent ZIP checksum file.

A report must not be described as complete until:

- every promised file exists at the stated delivery path;
- file sizes and SHA-256 values are recalculated from the delivered files;
- the manifest matches the actual files;
- the ZIP integrity test passes;
- every promised ZIP member is present;
- rendered PDF/DOCX pages are checked for readability where those formats are supplied;
- any missing loose-file link is corrected or clearly withdrawn.

The pack may remain private where publishing it would expose internal control detail, private diagnostics or other non-public material. In that case the public Control Board may record only the pack identity, reporting window, size, SHA-256 and non-sensitive verification result.

## 4. Historical corrections

When a later check finds that:

- a tested merge tree was labelled exact-head;
- review independence was overstated;
- a promised file or delivery link was missing;
- a later event changed a previously accurate limitation;

preserve the original record and issue a dated correction or supplement. State:

- the original record;
- the incorrect or outdated statement;
- the corrected statement;
- the evidence supporting the correction;
- whether the original decision changes.

## 5. Current correction reference

The first application of this supplement is:

[`AUDIT_EVIDENCE_CLASSIFICATION_CORRECTION_2026-08-04.md`](AUDIT_EVIDENCE_CLASSIFICATION_CORRECTION_2026-08-04.md)

It corrects the PR #45 and PR #47 workflow classification, the PR #45 review-relationship wording and the loose-file delivery defect in the private 07:42 BST control-room pack.

## 6. Gate position

This supplement changes audit terminology and delivery controls only. It does not approve application functionality, real evidence, AI behaviour, an executable, signing, a publisher, release or deployment.
