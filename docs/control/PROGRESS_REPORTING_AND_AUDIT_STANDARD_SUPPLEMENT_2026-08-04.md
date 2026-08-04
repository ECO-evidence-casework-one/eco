# ECO progress reporting and audit standard supplement — 4 August 2026

**Authority:** controlling supplement to `PROGRESS_REPORTING_AND_AUDIT_STANDARD.md`  
**Purpose:** tested-commit identity, reviewer relationship and downloadable-pack verification  
**Release effect:** none

## 1. Tested commit identity

A report must identify the exact commit actually checked out and tested.

Use:

- **exact-head evidence** — the raw checkout log proves the tested SHA equals the stated branch-head SHA;
- **tested PR merge-tree evidence** — GitHub checked out a synthetic `refs/pull/<number>/merge` commit or equivalent;
- **merged-tree verification** — the final squash/merge tree was compared with the accepted proposal and base;
- **post-merge CI evidence** — a workflow actually ran against the final merged commit.

For every PR workflow record:

- branch head SHA;
- base SHA;
- tested checkout SHA;
- whether the checkout was synthetic;
- run and job IDs;
- artifact inventory;
- final merge SHA;
- whether post-merge CI was observed.

Do not infer exact-head testing merely because a run is associated with a PR head.

## 2. Reviewer relationship and technical posting identity

Independence is determined by the actual relationship between the work preparer, evidence-generation lane and reviewer—not by the GitHub username alone.

Use:

- **self-review** where the same underlying person or agent prepared and reviewed the work;
- **control review** where a control lane reviews its own or closely connected governance/reporting work;
- **inspection-lane review** where a dedicated inspection workstream that did not create the proposal examines the underlying evidence;
- **organisationally independent review** only for a meaningfully separate person or organisation.

Where one technical account posts for multiple lanes, disclose it explicitly.

## 3. Downloadable audit packs

Create a downloadable pack when:

- the user requests a progress, continuity or audit report;
- a substantial inspection cycle reaches a material decision or handoff;
- a chat/tool failure, incident, rollback or premature merge creates recovery risk;
- chat text and GitHub comments alone would be difficult to reconstruct.

A standard full pack should include, where applicable:

- PDF report;
- DOCX report;
- event log CSV;
- source register CSV;
- current-state snapshot;
- SHA-256 manifest;
- receipt;
- ZIP archive;
- adjacent ZIP checksum.

Do not call a pack complete until:

- every promised file exists at the delivery path;
- sizes and hashes are recalculated from delivered files;
- the manifest matches;
- ZIP integrity passes;
- every promised member is present;
- PDF/DOCX rendering is checked where supplied;
- missing or broken loose-file links are corrected or expressly withdrawn.

A private pack may remain outside the public repository. The public control record may state only its non-sensitive identity, reporting window, size, hash and verification result.

## 4. Historical corrections

Preserve the original record and add a dated correction when later review finds:

- a merge tree was labelled exact-head;
- reviewer independence was overstated;
- a promised file or link was missing;
- a later event changed a previously accurate limitation.

State the original record, corrected statement, evidence and decision effect.

## 5. First application

The first application is:

[`AUDIT_EVIDENCE_CLASSIFICATION_CORRECTION_2026-08-04.md`](AUDIT_EVIDENCE_CLASSIFICATION_CORRECTION_2026-08-04.md)

It covers PRs #45, #47, #48, #49 and #51, plus the private 07:42 BST control-room pack.

## 6. Gate position

This supplement changes audit terminology and delivery controls only. It approves no application functionality, real evidence, AI behaviour, executable, signing, publisher, outreach, release or deployment.
