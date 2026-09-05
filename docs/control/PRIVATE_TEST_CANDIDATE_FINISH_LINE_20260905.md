# ECO private Windows test-candidate finish line

Version 1, 5 September 2026. Source baseline reviewed: `4fc6ea21122058bb63182ff456b71256a21305d1` (merged #130, preserving #129). This record defines a finite delivery target; it does not supersede release gates or retrospectively certify incomplete issue acceptance.

## Deliverable A: private synthetic-data Windows test candidate

One identifiable Windows launch entry point, with a reproducible package/receipt and clear offline-runtime requirements. A tester on the controlling 8 GB Acer can create/select an explicit workspace, organise a matter, import synthetic documents, read PDF/image/mail content, search and navigate to a source, save/close/reopen, export supported records and perform a backup/restore. Interrupted operations must preserve an identifiable last-good state or stop with a usable recovery path. The essential journey must work by keyboard at the tested display scaling without blocked controls or an unresponsive UI.

The candidate remains a controlled private synthetic-data handoff. Packaging must follow the existing private-handoff and signing conditions; passing a build does not itself authorise distribution. Exact allowed delivery conditions must be recorded before handoff.

### Fixed eight delivery checks

| ID | Check / objective completion condition | Baseline position | Evidence / remaining proof |
|---|---|---|---|
| A1 | Exact candidate passes all four normal automated jobs: Linux tests/vet, source policy, secret scan, Windows build/tests/integrity | PASS at reviewed baseline | #130; main run 33972302189. Must rerun for final candidate. |
| A2 | Qualified document-reading foundation is integrated: native PDF text, registered OCR, optional PDF visual runtime and page navigation | PASS for this bounded foundation | #122 combined stack and recorded actual-Acer PDF gate. This is not final-package or all-document acceptance. |
| A3 | Explicit workspace first run/open/reset, ownership, compatibility, recovery and supported version transitions satisfy #4 | PARTIAL | #124-#128 and #130 merged; runtime rollback, process-interruption/restart, migration policy and actual Windows acceptance still need complete mapping. |
| A4 | Synthetic source-preservation and casework round trip: import, verify, organise, save/reopen, supported export, backup/restore; original bytes remain verifiable | PARTIAL | Existing preservation/backup tests; #3/#12 full acceptance and exact end-to-end journey are not certified by this record. |
| A5 | Visible page-aware search and source navigation: search box, scope, match indication, previous/next and source binding | PARTIAL | #129 backend merged; UI/highlighting/accessibility acceptance under #8 remains. Do not duplicate the parallel search lane. |
| A6 | Essential journey is responsive and accessible on Windows | NOT CLOSED | #6/#7: keyboard, focus, scrolling, high-DPI, screen-reader and bounded long-operation/cancellation evidence required. |
| A7 | Exact offline/privacy behaviour is qualified for the candidate and optional tools | NOT CLOSED | #14 and user-facing claim alignment; no network fallback, private diagnostics, source/prompt distinction; source-policy tests alone are insufficient. |
| A8 | Reproducible private handoff and actual-Acer end-to-end acceptance | NOT CLOSED | Exact package/source/runtime/model identities, licences, launch/install/remove route, permitted signing/handoff conditions; test log for the final candidate. Prior PDF performance evidence remains valid within its tested scope. |

At baseline, **2 of 8 delivery checks are closed within their explicitly bounded definitions**. This is a count of acceptance milestones, NOT an effort-weighted percentage, feature-completion percentage, or estimate of time remaining. Partial rows count as zero closed; there is no invented partial credit. A1/A2 must be reconfirmed on the final package. No issue is closed solely by this table.

## Smaller workspace work-package counter

The current implementation sequence has eight packages:

| Package | Position |
|---|---|
| W1 Exclusive workspace owner primitives | Merged #124 |
| W2 Vault lifetime / Close / restore-owner handoff | Merged #125 |
| W3 Authenticated metadata stale-state checks | Merged #126 |
| W4 Candidate-specific development identity | Merged #127 |
| W5 Explicit startup and reversible archive/start-clean | Merged #128 |
| W6 Unsupported-format and unsafe-root refusal | Merged #130 |
| W7 Interruption/recovery qualification and required repairs | IN PROGRESS; bounded runtime rollback tests are only one part |
| W8 Exact Windows/Acer first-run/open/reset/reopen and startup-cost acceptance | NOT CLOSED |

**6/8 packages merged (75% by package count)**. W7 stays open after helper-level tests: handled errors, process termination and power-loss durability are distinct. This counter is not whole-application completion or closure of all six issue #4 acceptance bullets.

## Execution order and stop rule

1. Finish the bounded restore rollback / failed-save test slice. Reproduce observed failures before repairs and keep original/checkpoint files intact.
2. Complete the remaining process-interruption/restart recovery contract and supported version-transition evidence; block unsupported migrations rather than inventing an automatic converter. Close W7 only against explicit tests.
3. Continue #8 search UI in its existing parallel lane. Do not change or duplicate it from the workspace lane.
4. Exercise the single essential synthetic journey and repair the responsiveness/accessibility/privacy defects that prevent A4-A7 acceptance.
5. Freeze one exact private candidate, complete package/handoff checks, then run the combined Windows/Acer acceptance once. Any failure creates a bounded repair and rerun of the affected checks, not a restart.

Once A1-A8 pass for the frozen candidate, stop expanding the private-candidate feature scope and deliver the allowed private handoff. Any proposed new prerequisite must identify the existing failed acceptance requirement or be recorded as a scope change. New parsers, additional AI models, a new UI framework, background self-modifying features, additional donors and unrelated project work do not automatically become prerequisites. Preserve existing features; deferred expansion is not feature deletion. Unqualified optional generative functions must remain explicitly restricted rather than being presented as approved functionality.

## Deliverable B: public end-user release (separate finish line)

A public release requires Deliverable A plus the existing accountable publisher/steward, final exact signed package and SBOM/licences, installer/uninstaller, support/security/recovery arrangements and all applicable real-evidence, accessibility, privacy and intended-purpose release gates (#15/#17/#24 and linked controls). No public binary, production-ready or real-evidence-use claim follows from reaching the private test-candidate milestone.

## Time estimates

There is no defensible calendar completion date in the current evidence: process-interruption and actual-machine defects have not all been measured, and publisher/private-handoff dependencies are not solved by code. Report the remaining failed checks and bounded next change after each pass. Estimate calendar duration only when those dependencies and observed work rates support it; do not convert passing test counts into days remaining.

## Technical reference boundary

The Go `os.Rename` contract warns of OS-specific restrictions and lack of atomicity on non-Unix platforms; `File.Sync` is a separate durability operation. Passing simulated rename-boundary rollback tests must not be labelled proof of Windows power-loss durability. Official references: https://pkg.go.dev/os#Rename and https://pkg.go.dev/os#File.Sync .
