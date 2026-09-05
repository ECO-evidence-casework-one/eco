# Workspace format compatibility qualification

Refs #4. Synthetic development qualification only; no release or real-evidence approval.

## Initial before/after qualification

Application baseline: `d9541cd1f2830f175ff0d80bb812276a2ea5864a`.

Qualification run: `33971604200`.

Unchanged baseline: 10 failing regression subcases; supported legacy and compatible backup control fixtures both passed.

Initial patched candidate: 19 leaf acceptance cases passed. Full Go test suite, go vet, source policy and Windows cross-build also passed. The initial normal PR pipeline then passed all four jobs in run `33971782124`, including native Windows tests/build and integrity checks. Those results apply to initial head `b79de74b09bf5559968b9fc333ee46b8f461f737`, before the review follow-up below.

## Scope

Strict unknown-field/schema/trailing-data refusal is shared by workspace loading, authenticated metadata checks and portable-restore manifest decoding. A read-only preflight runs before ownership side effects and again under ownership before creation/recovery. Supported schema-1 metadata without revision/owner-transaction fields retains its records and gains those fields only on an explicit successful save.

## Review follow-up: unsafe empty root

Automatic review of PR #130 identified that an empty symlink root could be initialized but then refused on reopening. Four additional root cases were added before the fix: an empty-directory symlink, that route with a trailing separator, a dangling symlink, and an unrelated regular file.

Run `33972019455` on test-first head `3f30faf4c59d592a84f9d92ad82edb3709ec31b1` reproduced three failing symlink subcases. The Linux log showed unwanted initialization/mutation; the unrelated-file rejection passed. This was an expected red regression run, not a passing qualification.

The follow-up validates the cleaned absolute root with `Lstat` before accepting empty state or acquiring any owner. Symlink and non-directory roots are refused consistently before workspace creation. The four cases compare the complete parent/target tree before and after refusal.

The expanded suite comprises 23 selected Linux leaf cases (the original 19 plus four root cases). Final qualification must refer to the latest PR #130 head and its actual CI/review results, not reuse the earlier green head. On Windows, symlink fixture creation is explicitly skipped if the runner cannot create it; do not infer Windows symlink acceptance coverage from a package pass alone.

## Boundaries

No schema migration is introduced. Unknown same-schema extensions are deliberately refused, not deleted. This does not retrofit safety into historical executables, prove all candidate downgrade combinations, or close interruption/crash/rollback, real Windows UX, accessibility or release gates. Standard-library JSON duplicate-key and case-insensitive-name semantics are unchanged in this bounded slice. Extra startup validation cost still needs low-spec measurement. Issue #4 remains open.

## Initial 19 passed acceptance cases

- `TestWorkspaceFormatCompatibleBackupFixture`
- `TestWorkspaceFormatLegacyDefaultsPreserved`
- `TestWorkspaceFormatRejectsWithoutMutation/authentication_failure`
- `TestWorkspaceFormatRejectsWithoutMutation/future_missing_objects`
- `TestWorkspaceFormatRejectsWithoutMutation/future_schema`
- `TestWorkspaceFormatRejectsWithoutMutation/malformed_json`
- `TestWorkspaceFormatRejectsWithoutMutation/missing_key`
- `TestWorkspaceFormatRejectsWithoutMutation/missing_metadata`
- `TestWorkspaceFormatRejectsWithoutMutation/missing_schema`
- `TestWorkspaceFormatRejectsWithoutMutation/objects_only`
- `TestWorkspaceFormatRejectsWithoutMutation/trailing_json`
- `TestWorkspaceFormatRejectsWithoutMutation/unknown_matter`
- `TestWorkspaceFormatRejectsWithoutMutation/unknown_nested`
- `TestWorkspaceFormatRejectsWithoutMutation/unknown_top`
- `TestWorkspaceFormatRejectsWithoutMutation/unknown_without_owner_file`
- `TestWorkspaceFormatRejectsWithoutMutation/unsupported_older_schema`
- `TestWorkspaceFormatRestoreRejectsWithoutActivation/future_schema`
- `TestWorkspaceFormatRestoreRejectsWithoutActivation/unknown_nested`
- `TestWorkspaceFormatRestoreRejectsWithoutActivation/unknown_top`
