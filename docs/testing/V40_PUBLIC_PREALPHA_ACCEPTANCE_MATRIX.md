# V40 public pre-alpha acceptance matrix

**Issue:** #69  
**Status:** open qualification matrix  
**Date:** 4 August 2026

This matrix controls the first public V40 source release. A checked item requires retained evidence at one exact commit. Statements in a PR description are not evidence by themselves.

## A. Workspace integrity — release blocker

| ID | Requirement | Required evidence | Status |
|---|---|---|---|
| WS-01 | A writable workspace has one cross-process owner for the full Vault lifetime. | Real subprocess contention test on Windows and Linux. | OPEN |
| WS-02 | A stale process cannot replace newer workspace metadata. | Process A saves; process B attempts stale save; exact revision/CAS rejection and rollback proven. | OPEN |
| WS-03 | Aliases cannot create separate writable ownership domains. | Symlink/junction/bind-mount/renamed-parent tests using the same underlying object. | OPEN |
| WS-04 | New workspace creation is owned before any managed child is created. | Concurrent create test; one success, one clean refusal; no partial foreign cleanup. | OPEN |
| WS-05 | First candidate-state creation and replacement is owned and revision checked. | Concurrent first-launch subprocess test. | OPEN |
| WS-06 | Linux nested cleanup preserves exact child identity. | Same-mount substitution test preserves unrelated sentinel. | OPEN |
| WS-07 | Failed persistence rolls back in-memory mutation. | Forced write/rename/fsync failure tests across every mutating public path. | OPEN |
| WS-08 | Open, migrate, restore and reset retain issue #3 preservation guarantees. | Full regression suite. | OPEN |

## B. First usable Matter journey

| ID | Requirement | Required evidence | Status |
|---|---|---|---|
| UX-01 | Home identifies the current workspace and immediate next actions. | Native screenshot plus UI regression assertions. | OPEN |
| UX-02 | Create Matter asks only for name, optional reference and objective. | Keyboard-operated native test. | OPEN |
| UX-03 | Matter overview shows Current position, evidence wording, next actions and evidence status. | Native screenshot and semantic assertions. | OPEN |
| UX-04 | Evidence intake shows exact current stage and progress. | Timed synthetic import evidence. | OPEN |
| UX-05 | Cancellation creates no false completed evidence record. | Cancellation at every stage with persisted-state inspection. | OPEN |
| UX-06 | Close/reopen returns to the same workspace and Matter. | Acer and CI subprocess journey. | OPEN |
| UX-07 | What’s New lists all material changes and limitations. | Native first-launch and reopen evidence. | OPEN |
| UX-08 | No button or menu item is dead or placeholder-only. | Control inventory and action tests. | OPEN |

## C. Controlled assistance

| ID | Requirement | Required evidence | Status |
|---|---|---|---|
| AI-01 | The app uses one explicit adapter to M1.18. | Dependency mapping and package import inspection. | OPEN |
| AI-02 | Generated text cannot bypass verification. | Hostile adapter tests. | OPEN |
| AI-03 | Responses show current position, suggested action and named sources. | Synthetic Matter journey. | OPEN |
| AI-04 | Failure, timeout and cancellation return controlled output. | Subprocess and unit tests. | OPEN |
| AI-05 | No real evidence or model is used for public-pre-alpha qualification. | Fixture inventory and repository scan. | OPEN |

## D. Accessibility and performance

| ID | Requirement | Required evidence | Status |
|---|---|---|---|
| AP-01 | Core journey is keyboard operable. | Recorded key sequence and focus assertions. | OPEN |
| AP-02 | Focus is visible and ordered. | Native UI inspection at all core screens. | OPEN |
| AP-03 | Layout works at 125%, 150% and 200% scaling. | Windows screenshots and clipping checks. | OPEN |
| AP-04 | Low-sensory and reduced-motion settings work. | Before/after evidence. | OPEN |
| AP-05 | Long work never blocks the message thread. | Responsiveness telemetry and cancellation tests. | OPEN |
| AP-06 | Acer baseline can complete the journey within bounded memory. | Acer test receipt with peak memory and timings. | OPEN |
| AP-07 | At least one screen reader can identify the core controls and statuses. | Narrator or NVDA evidence. | OPEN |

## E. Source-release qualification

| ID | Requirement | Required evidence | Status |
|---|---|---|---|
| SR-01 | Exact commit is frozen and zero commits behind current main. | Compare receipt. | OPEN |
| SR-02 | Linux unit, race, vet and source-policy gates pass. | Exact-head Actions evidence. | OPEN |
| SR-03 | Windows failure-stop matrix and deterministic build pass. | Exact-head Actions evidence. | OPEN |
| SR-04 | No unapproved runnable artifact is uploaded by CI. | Exact run artifact inventory. | OPEN |
| SR-05 | README, limitations, changelog and release notes are truthful. | Delta review. | OPEN |
| SR-06 | SBOM, notices and dependency licences are current. | Exact source manifest and notices inspection. | OPEN |
| SR-07 | Repository contains no personal data, real evidence, credentials or private workspace objects. | Repository and fixture scan. | OPEN |
| SR-08 | Public tag and release identify the build as pre-alpha source. | Release record. | OPEN |

## F. Additional binary gates

These are not required for a source-only public release. They are mandatory before attaching a Windows executable.

| ID | Requirement | Required evidence | Status |
|---|---|---|---|
| BIN-01 | Exact executable is reproducible and hashed. | Two-build equality and receipt. | OPEN |
| BIN-02 | Malware and dependency checks pass. | Retained reports. | OPEN |
| BIN-03 | Acer launch, import, close/reopen, memory and crash-recovery tests pass. | Acer qualification bundle. | OPEN |
| BIN-04 | Keyboard, scaling and screen-reader checks pass on the exact executable. | Accessibility bundle. | OPEN |
| BIN-05 | Binary publication is explicitly authorised by the controlling release gate. | Updated release-gate record. | OPEN |
| BIN-06 | Unsigned/signing status is unambiguous. | Release wording and file metadata inspection. | OPEN |

## Release decision rules

- Any OPEN or FAIL item in section A blocks source and binary release.
- Any OPEN or FAIL item in section B blocks V40 public release.
- AI integration may be omitted only if the app clearly says unavailable and contains no fake or placeholder AI action. If included, all section C items apply.
- Section D failures may block or narrow the release depending on severity, but keyboard access, clipping and message-thread freezes are blocking.
- All section E items must pass for a source release.
- All section F items must pass for a binary release.
- No deadline can convert a failed gate into a pass.
