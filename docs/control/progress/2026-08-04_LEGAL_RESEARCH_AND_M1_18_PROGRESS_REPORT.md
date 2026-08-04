# ECO progress report — legal controls, steward research and M1.18 review — 4 August 2026

**Report ID:** `ECO-PROGRESS-20260804-LRM118-001`  
**Prepared:** 4 August 2026, approximately 11:34 BST / 12:34 CEST  
**Lane:** legal, standards, governance, organisational research and control review  
**Starting canonical baseline:** `f997e1049f8c24ed04848127ec26d55ee784b6f4`  
**Canonical ending state:** `76acdf214ba0bc4ef58dbf2159ae67c063a1a245`  
**Review relationship:** control-lane record through the shared technical account; not an organisationally independent audit or legal opinion

## Executive meaning

This work restored truthful audit terminology, brought ECO's canonical public status and release gate up to date, created an objective method for researching established steward organisations, preserved a ten-organisation public-source longlist with no positive candidate decisions, and imposed legal/privacy integration conditions on the newly merged M1.18 AI-turn orchestrator.

No organisation was contacted or appointed. No company, CIC, charity or other entity was created. No real evidence, executable, signing, release or deployment gate opened.

## Mission alignment

This work belonged in the legal, standards and governance lane because it concerned:

- accuracy of audit and workflow-evidence claims;
- public release-gate truthfulness;
- future publisher/steward responsibility;
- protection against automatic organisational duties for an individual contributor;
- structured public-source organisational research;
- diagnostic privacy and output-safety requirements for local-AI integration;
- consistency between development milestones and issues #14, #20, #46 and #53;
- preservation of P0/P1 blockers and no-outreach controls.

Application implementation was not undertaken in this lane.

## Work completed

### 1. Audit-evidence classification corrected

Historical records had repeatedly described successful pull-request workflows as `exact-head` evidence even where raw logs showed GitHub had checked out synthetic `refs/pull/<number>/merge` commits.

PR #55 consolidated the correction for PRs #45, #47, #48, #49 and #51.

It established four distinct evidence terms:

- **exact-head evidence** — the raw tested SHA equals the branch head;
- **tested PR merge-tree evidence** — GitHub tested a synthetic pull-request merge commit;
- **merged-tree verification** — the final squash/merge tree was separately compared;
- **post-merge CI evidence** — a workflow actually ran against the final merged commit.

PR #55 merged as:

`e6e83dd7adc0269c16afda806c324063264e5e6e`

Its frozen head `78cfc7bf38f9f4c36598308bc7ccc24a3e42eddb` was tested through run `30896127865` at synthetic merge `b20320de8891fb6bdc753045cf8849d5d161e347`. All jobs passed and the artifact inventory was empty.

### 2. Stale audit proposals retired

PR #52 was superseded by the current-main consolidated correction.

PR #54's partner-pack progress report was rebuilt after the correction and then closed unmerged as superseded.

The stale histories were preserved rather than deleted or silently rewritten.

### 3. Partner-pack report reconciled

PR #57 preserved a current-main partner-pack progress report that correctly classified PR #51's workflow evidence and separated it from the final squash-tree comparison.

PR #57 merged as:

`0dc57720c1c3394b03342427bc9b4dca09c1f040`

Its frozen head `3b1d6c92510861e950d64692cee1edc3c2025bf1` was tested through run `30896765934` at synthetic merge `729c3dd85ca97cfad2e33361b86c1d93af45b3b6`. All jobs passed and the artifact inventory was empty.

### 4. Canonical status and release gate repaired

`CURRENT_STATUS.md` and `docs/control/CURRENT_RELEASE_GATE.md` remained materially stale. They still referred to older main identities, described some PRs as open when they were closed, and did not reflect the intended-purpose, publisher, audit or partnership controls.

PR #60 replaced both canonical records with one current and internally consistent position.

It recorded:

- issue #16 governance merged but implementation/claims evidence still open;
- issue #20 and issue #46 still open;
- PRs #18, #19, #52 and #54 closed unmerged;
- PR #11 still blocked by four issue #4 P0 findings;
- issue #15 and issue #24 still P0;
- issue #17 open and no organisation appointed;
- the generic partner pack approved for generic information only;
- the audit-evidence terminology from PR #55;
- every real-evidence, signing, executable, release and deployment block.

PR #60 merged as:

`8cf4da3c1a1bc196280080d147ea5ab833111866`

Its frozen head `13554bb04ab8e7b15e87d946cc52c6790aed1fd4` was tested through run `30897876626` at synthetic merge `2bab13f0dbe878d4cd4a3792de3d0f62047a7f58`. All jobs passed and the artifact inventory was empty.

### 5. PR #11 and issue #4 reconciled to current governance

Current-main hold records were added to PR #11 and issue #4.

PR #11 remains:

- draft;
- unmerged;
- non-mergeable;
- current head `61a2004809b341e72f70321843c64c3ff477f549`;
- last reviewed workspace implementation head `73689717bb08bb8cec0fc1233b92f843b449484a`.

Its four P0 findings remain:

1. stale writers can overwrite newer state;
2. ownership can split through aliases or substituted parents;
3. Linux cleanup can reopen an inspected object by name;
4. creation, first launch and candidate state lack one alias-safe cross-process ownership transaction.

The next branch must be reconstructed or reconciled from current `main`; it is not a rebase-only task.

### 6. Steward-organisation research issue opened

Issue #62 was created as a research-only organisational workstream.

It does not authorise:

- outreach;
- appointment;
- partnership claims;
- a new entity;
- real-data pilots;
- contracts;
- repository/signing transfer;
- release or deployment.

### 7. Objective research method and empty register merged

PR #63 added:

- `docs/partnership/NAMED_ORGANISATION_RESEARCH_METHOD.md`;
- `docs/partnership/CANDIDATE_ORGANISATION_RESEARCH_REGISTER.md`;
- updated partnership-directory controls.

The method includes:

- a primary-source hierarchy;
- freshness and uncertainty rules;
- fatal exclusions;
- non-tradeable ECO boundaries;
- separate assessment domains for legal identity, mission, FOSS, privacy/offline operation, security, accessibility, complaints, release/signing, continuity, sustainability, contracting and regulatory capability;
- `POTENTIAL FIT`, `HOLD`, `UNSUITABLE` and `NOT ASSESSED` decisions;
- separation between longlisting, shortlisting, outreach-target approval and formal issue #17 acceptance.

Missing evidence cannot be scored positively. Reputation, size, mission language and previous partnerships do not prove capability or interest.

PR #63 merged as:

`514ae2dbf2696b9c39f1896676ac60a711af4de9`

Its frozen head `44c260aa281fd5b9c80b162421749f475ac38eb8` was tested through run `30898825396` at synthetic merge `7f7e64519297fe1c9c377b40322f8e155b725208`. All jobs passed and the artifact inventory was empty.

The candidate register remained empty.

### 8. First controlled public-source longlist merged

A role-specific longlist was created for deeper research into:

- mySociety;
- Open Knowledge Foundation;
- Software Freedom Conservancy;
- Software in the Public Interest;
- Code for Science & Society;
- CAST;
- AbilityNet;
- Law for Life;
- Open Rights Group;
- OpenUK.

Every organisation remains `NOT ASSESSED`.

The longlist distinguishes possible roles such as:

- full accountable stewardship;
- FOSS/fiscal/legal hosting;
- civic-technology collaboration;
- accessibility testing and disabled-user participation;
- privacy and digital-rights support;
- public legal education/legal-capability work;
- open-technology specialist networks;
- social-sector digital-capacity support.

The record identifies material unknowns and explains that fiscal sponsorship is not automatically operational stewardship and that specialist capability is not proof of full publisher capacity.

A concurrent application-development merge advanced `main` while the longlist PR was open. The longlist was therefore validated against the resulting current tree rather than the obsolete original base.

PR #64 merged as:

`76acdf214ba0bc4ef58dbf2159ae67c063a1a245`

Its frozen head `749346e7f0aa30daff7c3705a9fc37aabfebe362` was tested through run `30899798883` at synthetic merge `5eb01d6f45d0eb04aba720cbd6fe21844c42d63e`, combining it with current base `b66cab51ad0da118303afd7009065e25a27e6ab7`. All jobs passed, including the new M1.18 tests, and the artifact inventory was empty.

No shortlist or contact route was created.

### 9. Concurrent M1.18 source milestone inspected

PR #58 merged the final isolated AI-turn orchestrator as:

`b66cab51ad0da118303afd7009065e25a27e6ab7`

The package coordinates injected contracts for:

- request validation;
- context compilation;
- model/runtime/adapter routing;
- worker admission;
- streamed generation;
- lease release;
- verification;
- deterministic fallback;
- transient-buffer erasure;
- bounded receipts.

It does not connect a model, runtime, evidence source, application UI or user-visible journey.

The package prevents raw generated bytes from becoming accepted output without an injected verifier returning acceptance. Its own structural validation, however, proves only a safe identifier, non-empty output and size bound. It does not itself establish:

- source support or citations;
- medical/legal/high-consequence compliance;
- output-status separation;
- meaningful user review;
- professional correctness;
- privacy safety of injected receipt identifiers.

### 10. M1.18 workflow terminology corrected

The PR #58 review named synthetic merge `21ae981bc677e65d3bd001f8a91e4442feb64fb6` but also called run `30898880688` exact-head evidence.

A correction comment now records the run as **tested PR merge-tree evidence**, not exact-head CI.

The bounded M1.18 source PASS remains intact.

### 11. M1.18 legal and privacy integration gates added

Control comments were added to:

- PR #58;
- P0 issue #14;
- P0 issue #20;
- P1 issue #46;
- P0 issue #53.

They require the first application integration to prove:

- source-aware compiler/retrieval metadata;
- substantive verifier enforcement of source support and prohibited-output rules;
- reviewed allowlisted non-echoing fallback templates;
- random or approved-tokenised non-content receipt identifiers;
- seeded privacy-canary testing across all success/failure/cancellation/panic paths;
- a typed application result before UI or export;
- preservation of source/OCR/suggestion/confirmation/note status;
- no presentation of `accepted`, `VerificationID` or receipt hashes as factual, legal, medical or forensic approval;
- Phase B workspace repair and Phase C's visible Matter journey before Phase D generated integration.

M1.18 does not close issues #14, #20, #46 or #53.

### 12. Control Board refreshed

Issue #22 was refreshed to canonical `main` `76acdf214ba0bc4ef58dbf2159ae67c063a1a245`.

It now records:

- the merged research method;
- the ten-entry longlist;
- the empty candidate register;
- the M1.18 legal/privacy integration boundary;
- the scheduled issue #24 historical-artifact recheck;
- every unchanged product, evidence and release gate.

## Decisions

| Matter | Decision | Exact scope |
|---|---|---|
| Audit terminology correction | PASS | PR #55 controlling correction |
| Canonical status/release gate | PASS | PR #60 two-file public-control reconciliation |
| PR #11 | BLOCKED | Four unchanged issue #4 P0 findings |
| Steward research method | PASS | Research method and empty register only |
| Named organisation longlist | PASS | Ten entries, all `NOT ASSESSED` |
| Candidate shortlist | NOT CREATED | No candidate assessment completed |
| Outreach | BLOCKED | No target, public contact route or sender role approved |
| M1.18 source milestone | PASS within isolated orchestration scope | Not application, evidence or safety qualification |
| M1.18 issue #14/#20/#46 conformance | BLOCKED | Exact adapters, verifier, fallback, typed result and tests required |
| Real evidence | BLOCKED | No change |
| Executable/signing/release/deployment | BLOCKED | No change |

## Evidence index

### Canonical commits

- starting baseline: `f997e1049f8c24ed04848127ec26d55ee784b6f4`;
- audit correction: `e6e83dd7adc0269c16afda806c324063264e5e6e`;
- partner report: `0dc57720c1c3394b03342427bc9b4dca09c1f040`;
- canonical status/release gate: `8cf4da3c1a1bc196280080d147ea5ab833111866`;
- research method: `514ae2dbf2696b9c39f1896676ac60a711af4de9`;
- M1.18: `b66cab51ad0da118303afd7009065e25a27e6ab7`;
- organisation longlist/current ending state: `76acdf214ba0bc4ef58dbf2159ae67c063a1a245`.

### Principal PRs

- #55 — audit-evidence correction;
- #57 — reconciled partner-pack report;
- #60 — canonical status and release gate;
- #63 — named-organisation research method;
- #58 — M1.18 isolated turn orchestrator;
- #64 — controlled public-source organisation longlist.

### Principal issues

- #4 — workspace ownership/concurrency;
- #14 — diagnostic privacy and offline claims;
- #15 — actual-build provenance/licensing;
- #17 — publisher/stewardship acceptance;
- #20 — prohibited clinical/professional/high-consequence outputs;
- #22 — live Control Board;
- #24 — public executable containment;
- #46 — UI/AI/export intended-purpose conformance;
- #53 — cumulative development and first usable journey;
- #62 — steward-organisation research.

## Change inventory

Across this control cycle:

- application source changes by this lane: **none**;
- workflow changes by this lane: **none**;
- executable/model/dependency/SBOM/licence changes by this lane: **none**;
- governance/control/research/reporting documents: **created or reconciled as recorded above**;
- application source merged concurrently through PR #58: **nine isolated M1.18 files, separately inspected**;
- named organisations contacted: **none**;
- candidate assessments completed: **none**;
- real or sensitive evidence used: **none**;
- public runnable artifacts uploaded by reviewed runs: **none**.

## Problems, limitations and adverse findings

### Repeated workflow evidence misclassification

Multiple historical records used `exact-head` wording for synthetic PR merge-tree runs.

Effect: corrected through PR #55 and later event-level comments. Successful tests remain bounded to their actual tested trees.

### Parallel-main movement

Concurrent governance and development merges moved `main` during PR preparation.

Effect: stale PRs #52 and #54 were superseded; PR #64 was revalidated against M1.18 current `main`. Stable-base checks remain necessary before every merge.

### M1.18 verifier is injected and structurally checked only

M1.18's acceptance boundary is not itself the substantive source/professional-safety verifier.

Effect: issue #20 and issue #46 remain open. The first adapter must provide structured evidence support and prohibited-output enforcement.

### Receipt identifiers may be semantically identifying

M1.18 syntax-bounds IDs but does not prove that injected IDs are opaque.

Effect: issue #14 now requires random/tokenised IDs and seeded privacy-canary tests before persistence or export.

### Go string erasure limitation

Request and accepted text cross the M1.18 API as strings and cannot be overwritten in place.

Effect: no claim of complete in-memory erasure is permitted. Adapters must minimise copying, logging and retention.

### No organisation is yet assessed

The longlist is not a shortlist. No public material currently proves that one organisation can satisfy every ECO duty.

Effect: all entries remain `NOT ASSESSED`; no outreach is authorised.

### Codex review unavailable

PR #58 recorded Codex usage-limit exhaustion. It was not treated as approval.

## Gate position

### Completed in this cycle

- controlling audit-evidence correction;
- current canonical status and release gate;
- current-main PR #11 hold record;
- steward-organisation research method;
- empty candidate register;
- ten-entry role-specific longlist;
- M1.18 legal/privacy integration requirements;
- refreshed Control Board.

### Still blocked or open

- PR #11 workspace foundation;
- issue #14 diagnostic privacy;
- issue #15 actual-build provenance and licensing;
- issue #17 organisational acceptance;
- issue #20 prohibited-output enforcement;
- issue #24 historical artifact and distribution evidence;
- issue #46 cross-surface conformance;
- issue #53 visible application journey and integration;
- candidate assessment, shortlist and outreach;
- real or sensitive evidence;
- signing and public executable distribution;
- ordinary-user testing;
- institutional, healthcare, justice-sector or EU supply;
- legal, medical, forensic, security, accuracy, accessibility or compliance claims.

## Next controlled actions

1. Apply the merged research method to a small role-diverse batch, initially one possible wider steward, one FOSS/fiscal host and one specialist collaborator.
2. Verify legal identity from authoritative registers and apply fatal exclusions before domain scoring.
3. Record missing evidence and keep overall status at HOLD or NOT ASSESSED where threshold domains are unproved.
4. Stop for review before creating any shortlist.
5. Do not contact an organisation until a separate outreach-target decision approves the target, role, public route, sender and exact message.
6. In development, complete issue #4 Phase B and issue #53 Phase C before any M1.18 Phase D integration.
7. Carry the issue #14/#20/#46 M1.18 requirements into exact adapter, verifier, fallback, receipt and typed-result tests.
8. Perform the enabled 10 August issue #24 historical-artifact recheck without downloading or executing artifacts.

## Review statement

This report was prepared from repository records, exact file comparisons, raw workflow checkout identities, job results, artifact inventories, current public organisational sources and the controlling ECO governance records.

The same connected project account performs and posts work across project lanes. This report is a control-lane record, not an organisationally independent audit, candidate due diligence, outreach approval or legal opinion.

No organisation, individual, partnership, publisher, contract, real-evidence approval, executable, release or deployment is created by this report.
