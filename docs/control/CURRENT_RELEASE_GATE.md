# Current ECO release gate

**Gate record:** `ECO-RELEASE-GATE-20260804-009`  
**Updated:** 4 August 2026, approximately 10:50 BST / 11:50 CEST  
**Canonical public status:** [`../../CURRENT_STATUS.md`](../../CURRENT_STATUS.md)  
**Current reviewed `main` before this reconciliation:** `0dc57720c1c3394b03342427bc9b4dca09c1f040`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Signed end-user release:** none

## Controlling decision

ECO remains a source-development project only. No executable, real-evidence use, AI-assisted public function, publisher, outreach target, partnership, release or deployment is approved.

The following remain blocked:

- real, sensitive or irreplaceable evidence;
- ordinary-user executable testing;
- public binary distribution through Releases, Actions artifacts or another channel;
- use or redistribution of historical unsigned executables;
- public preview, release-candidate or stable-release status;
- public promotion of current Ask ECO or other AI-assisted evidence functions;
- diagnosis, treatment, clinical-risk, professional-representation, profiling, scoring or authority-side decision outputs;
- named outreach or partnership activity without separate research and approval;
- public-sector, institutional, healthcare, justice-sector or EU supply and deployment;
- legal, forensic, medical, accessibility, security or regulatory-compliance claims.

## Gate matrix

| Gate | Current position |
|---|---|
| Source identity | `main` contains post-V25 development and control work; no later named milestone is approved |
| Evidence preservation | PR #10 materially improved controls; issue #3 remains open pending issue #12 and complete acceptance evidence |
| Ask verification and restore | Blocked by issue #12 |
| Workspace lifecycle | PR #11 remains draft and blocked by four P0 ownership/concurrency findings |
| Windows CI truthfulness | Issue #27 closed for covered native commands; does not approve the executable |
| Public Actions distribution | New uploads stopped; inspected PR merge-tree and merged-tree routes expose zero artifacts; issue #24 remains open |
| Diagnostics and offline claims | Blocked by issue #14 |
| Final build provenance | Blocked by issue #15 |
| One-file Windows packaging | Not qualified on the actual final embedded executable |
| Authenticode signing | No final signed file; no post-signing integrity evidence |
| OCR and local AI | Not approved as reliable production functionality |
| Intended purpose | Controlling governance record exists; issues #16, #20 and #46 remain open for implementation conformance |
| Publisher and stewardship | Controlling gate exists; issue #17 remains open and no organisation is appointed |
| Generic partner material | Present and public-safe; no named research, outreach or relationship authorised |
| Accessibility | Blocked pending issue #7 evidence and truthful conformance documentation |
| Public claims | Controlled development claims only |

## Application lane — PR #11 and issue #4

PR #11 remains open, draft, unmerged and non-mergeable.

- Last inspected workspace implementation: `73689717bb08bb8cec0fc1233b92f843b449484a`.
- Current branch head: `61a2004809b341e72f70321843c64c3ff477f549`.
- Against reviewed `main` `0dc57720...`: 8 commits ahead and 41 commits behind; common base `2dcb44ba8541ab7a319de6e1f14f016aafe2ac1b`.

The later branch commits contain issue #24 workflow containment only. They do not correct:

1. stale independently opened writers without exact revision/CAS conflict protection;
2. alias-bypassable path-derived locking and split ownership;
3. Linux recursive cleanup reopening an inspected directory by name;
4. unowned workspace creation, first launch and candidate-state writes.

The branch must be reconstructed from exact current `main`, not resolved through a wholesale branch/main choice or casual conflict-button merge. Current-main workflow, Windows memory safety, governance, provenance and audit controls must survive intact.

Do not mark PR #11 ready, merge it, close issue #4 or use it with real evidence before one coherent correction and fresh hostile/concurrency review pass.

## Evidence preservation — issues #3 and #12

Issue #3 remains open. PR #10's preserved-object controls are material progress but are not final closure evidence.

Issue #12 requires bounded verification cost and safe Ask/restore serialization. Persistence errors must reach callers, in-memory state must roll back and the UI must remain truthful.

Real-evidence use remains blocked until both issues and their interaction tests pass.

## Windows CI — issue #27

Issue #27 is closed. The accepted control includes:

- fail-fast native-command handling;
- a deliberate non-zero native-command self-test;
- a six-stage failure matrix;
- static coverage of every native stage in `scripts/build-windows.ps1`;
- Windows pointer-safety fixes exposed by truthful `go vet`;
- successful tested merge-tree and merged-tree validation with zero artifacts.

This permits reliance on the covered command results only. It does not approve any runner-built executable.

## Public executable containment — issue #24

The workflow no longer uploads `ECO.exe`, DLLs, installers or runnable archives. Earlier affected branches were checked through controlled PR merge-tree routes and published zero artifacts after containment.

Four historical unsigned executable artifacts remain prohibited until administrator deletion or confirmed expiry. A one-time expiry recheck is scheduled for 10 August 2026 at 19:00 Europe/London.

Issue #24 remains open until:

- all four historical artifacts are deleted or confirmed expired;
- a controlled manual-dispatch run exposes no runnable payload;
- future automation remains unable to publish a binary before provenance, signing, publisher and explicit release approval.

Calling an output temporary, test or provenance material does not bypass this gate.

## Actual-build provenance — issue #15

Historical root manifest, preparation-receipt and SPDX records are explicitly labelled historical/source-level and cannot support a current release.

Issue #15 requires one exact final executable reconciled with:

- immutable commit and source tree;
- clean tracked-file manifest;
- build inputs and action identities;
- executable hash, PE imports, sections, resources, manifest and Go build information;
- actual-build SBOM and complete notices;
- corresponding source and model/runtime provenance;
- reproducible receipt;
- final Authenticode signature with no later mutation.

No current runner build satisfies this gate.

## Intended purpose — issues #16, #20 and #46

The controlling governance record is:

[`../governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md`](../governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md)

Permitted in principle: source-linked, visibly labelled and user-controlled assistance with supplied evidence, including extraction, OCR, exact search, document-stated facts, summaries, chronologies and user-controlled drafting.

Prohibited or separately gated:

- diagnosis, treatment, prognosis, triage, monitoring, clinical-risk or emergency assessment;
- reserved legal activities, professional representation or outcome guarantees;
- forensic authenticity, provenance or admissibility guarantees;
- profiling or scoring of eligibility, credibility, honesty, dangerousness or entitlement;
- grant, denial, reduction, revocation or recovery of essential services;
- authority-side adverse decisions without a fresh assessment and separate approval.

Issues #16, #20 and #46 remain open because the application, help, AI responses, diagnostics, packaging and public claims have not passed cross-surface conformance testing. Current Ask ECO remains blocked for real evidence and public AI-assisted use.

## Publisher, stewardship and partner controls — issue #17

The controlling gate is:

[`../governance/PUBLISHER_AND_STEWARDSHIP_GATE.md`](../governance/PUBLISHER_AND_STEWARDSHIP_GATE.md)

No individual or organisation is appointed. Contribution or repository administration alone does not assign publisher, supplier, support, complaints, controller or liability roles; actual conduct and applicable law remain controlling.

Issue #17 closes only after a named established organisation completes due diligence, formally accepts the role and proves operational response, security, support, complaints and continuity controls.

The generic pack under `docs/partnership/` is informational only. It authorises no named research target, message, relationship, pilot, contract, repository/signing transfer, release or deployment.

## Audit evidence and reporting

The controlling correction is:

[`AUDIT_EVIDENCE_CLASSIFICATION_CORRECTION_2026-08-04.md`](AUDIT_EVIDENCE_CLASSIFICATION_CORRECTION_2026-08-04.md)

Use:

- `exact-head evidence` only when the raw checkout proves the exact branch-head SHA was tested;
- `tested PR merge-tree evidence` for synthetic PR merge checkouts;
- `merged-tree verification` for final squash/merge comparisons;
- `post-merge CI evidence` only where a workflow actually ran against the final merged commit.

Reviewer relationship must identify the actual person, agent or workstream. A shared technical posting account does not by itself establish independence.

## Stop before real evidence

Stop where any of the following remains unresolved:

- preserved bytes, hashes and derived readings cannot be reconciled exactly;
- Ask, restore, migration, creation or reset can observe or create mixed state;
- filesystem boundaries are unproven against aliases, links, junctions, reparse points or parent substitution;
- output routes can produce prohibited clinical, professional, profiling, scoring or authority-side conclusions;
- source text, OCR, generated suggestion, user confirmation and user note are not visibly separated;
- diagnostics may expose case content;
- any real-evidence P0 or P1 finding remains open.

## Stop before public preview or release

Stop where:

- an unapproved runnable payload is public;
- the final executable is unsigned or not the one-file deliverable;
- source, manifest, SBOM, notices, model/runtime provenance or build receipt do not reconcile;
- the file changed after signing;
- accessibility and low-resource qualification are incomplete;
- application behaviour and public claims do not conform to the intended-purpose boundary;
- no accountable publisher accepts operational duties;
- any release-blocking P0 or P1 finding remains unresolved.

## Public-record rule

Public records may identify defect classes, controls, tests and decisions. They must not expose personal evidence, private diagnostics, credentials, confidential screenshots, unapproved binaries, model files, private workspaces or exploit-level instructions. Use synthetic and non-sensitive material only.
