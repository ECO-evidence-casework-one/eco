# W7 recovery acceptance map

5 September 2026. Applies to the recovery-reporting continuation of merged baseline `726be7dc9ab6c808325173db7c236915d3554376` (#133).

## Decision rule and boundary

This is the final source-acceptance mapping for the existing W7 work package in `PRIVATE_TEST_CANDIDATE_FINISH_LINE_20260905.md`. It does not add a ninth delivery check, redefine the essential journey or close issue #4.

The fixed requirement is that interrupted operations preserve an identifiable last-good state or stop with a usable recovery path. W7 source qualification can be recorded complete only after the reporting changes and tests below pass normal native Linux/Windows CI, review findings are addressed, the exact source is merged, and its post-merge CI passes. Until those observations exist, W7 is a closure candidate, not complete. Record the actual final head and run in the PR completion checkpoint; do not infer a pass from this document.

W8 remains the exact Windows/Acer first-run, reopen, reset, recovery-choice and startup-cost acceptance. A3 and A8, and the public/real-evidence gates, remain open. Hosted native tests do not replace interaction with the final packaged application.

## Requirement-to-evidence map

| Existing obligation | Implementation and tests | What the evidence does not prove |
|---|---|---|
| A failed handled restore retains the original and does not substitute an unrelated directory | #132 identity-bound rollback and cleanup; `TestRecoveryRollbackRenameBoundaries`, `TestRecoveryRollbackRefusesUnsafeState`, `TestRecoveryStageCleanupOwnership` | No atomic filesystem transaction or defence against every adversarial concurrent path change. |
| A failed/uncommitted metadata save does not replace the last committed workspace | `TestRecoveryMetadataWriteInterruption`; authenticated metadata CAS tests from #126 | No storage-controller/power-loss durability claim. |
| A forced process stop leaves the original recoverable and a predictable restart result | #133 `TestRestoreForcedStopRestart`: five actual restore stop boundaries; `TestResetForcedStopRestart`: two actual reset stop boundaries. Parent kills and waits for the real child operation, without manually reconstructing its directory state. | Seven selected boundaries, not arbitrary timing or whole-machine reboot testing. |
| Missing or empty selected state is never silently created by an ordinary open | #133 strict `OpenVault`; `TestStrictOpenNeverCreates`, `TestExplicitCreateRefusesExistingRoutes`, `TestExplicitCreateThenStrictOpen`, recovery guard tests | Final first-run UI and installation behaviour remains W8. |
| Deliberate Start clean remains usable, preserves old records and respects existing ownership | #133 combined reset/explicit creation; `TestExplicitCleanStartWithRestoreHistory`, `TestExplicitCleanStartRefusesLiveOwner`; this continuation adds identity-bound archive rollback and failure locations | No permission to reset arbitrary folders; actual selection and prompt acceptance stays W8. |
| Supported format transitions retain records, and unknown formats are refused rather than rewritten | #130 schema/unknown-field/trailing-data tests; new `TestRecoveryReportingLegacyStrictOpen` and `TestRecoveryReportingStrictUnknownFormat` exercise production strict OpenVault, not the fixture create-or-open helper | No automatic schema converter is implemented. Unsupported migrations are blocked; historical executables are not retroactively made safe. Any future converter requires its own checkpoint/interruption tests. |
| Restore/reset failures name the local recovery leads, preserve the real causes and explain the next action | `WorkspaceRecoveryError`; returned-error regression for real failed-stage cleanup; archive failure/rollback tests. The primary and rollback/cleanup error identities remain inspectable. | A named path is not claimed to exist or be complete after an arbitrary external move. The error explicitly says locations are leads to verify. |
| A committed restore with a release/cleanup problem is not mislabelled as failed or silently successful | `recordRestoreFinalization`, receipt recovery warnings and `RestoreCompletionNotice`; three finalization-state controls and a normal actual restore control | Owner-release fault semantics are unit-tested; this is not a physical hardware-failure simulation. |

## Reporting change and reproduced defects

Before/after qualifier `33978398733` ran two tests against unchanged baseline #133: rollback secondary-error identity was lost, and a real failed restore discarded cleanup location/action details. Both failed before the patch. The patched source passed nine selected reporting cases, full Linux tests, vet, source-policy and Windows cross-build.

The new branch also adds two strict production-API transition controls, bringing this continuation's selected reporting/transition scope to eleven cases. Those additional controls, and the final complete tree, require ordinary final-head native CI. The prior nine-case receipt is historical evidence for its own qualification stage, not proof that later changes passed.

The archive failure path now validates the owned source and vacant destination before rollback rather than moving a substituted folder. Returned context identifies both the requested route and the attempted archive. This extends the same already-required identity-bound recovery behaviour, not a new product feature.

Cleanup runs while the stage is still owned; its error is joined before the final local recovery message is formed. A successful activation remains a successful receipt with a visible warning if final release fails. The Windows completion path uses the same tested notice formatter; its normal title no longer makes an unqualified 'restored safely' claim.

## Test-set accounting

#132's 17 handled-error cases, #133's 23 opening/restart cases and this continuation's eleven reporting/transition cases answer different questions. The seven forced-stop cases are already included in #133's 23. Do not describe these counts as independent defects, exhaustive coverage, equal amounts of engineering effort or a whole-product completion percentage.

Some pre-existing tests use the #133 test-only fixture helper. The two added transition controls and the real restart/open-contract tests use the production strict API explicitly. No production creation bypass is added by this continuation.

## Remaining work after W7 source qualification

Perform W8 with the frozen, permitted private synthetic-data candidate and existing machine checklist: verify the visible workspace/action, opening and clean-start choices, recovery instructions, keyboard/display behaviour and low-spec timings. Preserve the existing A4-A7 end-to-end, responsiveness/accessibility and offline/privacy obligations; do not count them as satisfied by W7.

Keep source/release status documents distinct from historical checkpoints. A passed source work package is not permission to distribute an unsigned binary, use real evidence, claim a public release or bypass the remaining packaging/publisher requirements.

## Technical references checked for the reporting design

- Go error trees and joining: https://pkg.go.dev/errors#Join
- Named return values and deferred calls: https://go.dev/ref/spec#Defer_statements
- Rename atomicity/durability boundary: https://pkg.go.dev/os#Rename

Recovery locations are local UI material. This patch adds no network logging or automatic diagnostics export; do not copy real paths or evidence into public issue reports.
