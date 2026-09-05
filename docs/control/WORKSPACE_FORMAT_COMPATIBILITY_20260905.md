# Workspace format compatibility qualification

Refs #4. Synthetic development qualification only; no release or real-evidence approval.

Application baseline: `d9541cd1f2830f175ff0d80bb812276a2ea5864a`.

Qualification run: `33971604200`.

Unchanged baseline: 10 failing regression subcases; supported legacy and compatible backup control fixtures both passed.

Patched candidate: 19 leaf acceptance cases passed. Full Go test suite, go vet, source policy and Windows cross-build also passed. Native Windows execution and the normal full PR pipeline remain separate gates.

## Scope

Strict unknown-field/schema/trailing-data refusal is shared by workspace loading, authenticated metadata checks and portable-restore manifest decoding. A read-only preflight runs before ownership side effects and again under ownership before creation/recovery. Supported schema-1 metadata without revision/owner-transaction fields retains its records and gains those fields only on an explicit successful save.

## Boundaries

No schema migration is introduced. Unknown same-schema extensions are deliberately refused, not deleted. This does not retrofit safety into historical executables, prove all candidate downgrade combinations, or close interruption/crash/rollback, real Windows UX, accessibility or release gates. Standard-library JSON duplicate-key and case-insensitive-name semantics are unchanged in this bounded slice. Extra startup validation cost still needs low-spec measurement. Issue #4 remains open.

## Passed acceptance cases

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
