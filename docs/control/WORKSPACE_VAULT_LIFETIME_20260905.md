# Workspace Vault ownership lifetime — 5 September 2026

**Issue:** #4 clean, explicit workspace state  
**Baseline:** `main` after PR #124, commit `c1c82ea33d016507a80df655b8bab2ecec09e2d1`  
**Scope:** current-main lifecycle integration of the qualified workspace-owner primitives

## What this slice changes

- `OpenVault` acquires or creates the workspace root under an exclusive object-bound owner **before** creating `objects/`, `vault.key` or `workspace.ecodb`.
- The owner is retained for the writable Vault lifetime.
- A second writable `OpenVault` for the same workspace is rejected with `ErrWorkspaceInUse` while the first Vault is alive.
- `Vault.Close()` waits for operations using `opMu`, marks the Vault closed, clears the in-memory encryption key and releases the owner.
- Authenticated `saveUnlocked()` validates that the Vault is open and that its owner still identifies the same workspace object before writing metadata.
- Tests that deliberately simulate application restart now explicitly close the old Vault before reopening the same path.

## Restore activation

Portable restore keeps two independent owners during staging:

1. the existing active Vault owns the current workspace object;
2. the staged Vault owns the staged workspace object.

During activation:

- the old active directory is renamed to the pre-restore checkpoint and its owner is retargeted only if the object identity is unchanged;
- the staged directory is renamed into the active path and its owner is retargeted to the same staged object at the new route;
- activated staged ownership and preserved-object contents are reverified;
- only then are the staged owner, key and workspace transferred into the existing active `Vault` object;
- the old checkpoint owner is released after successful transfer;
- activation failures attempt to move both filesystem objects and their owners back together and report rollback failures rather than hiding them.

The previous `OpenVault(v.Root)` reopen-behind-the-existing-Vault pattern is removed from restore activation.

## Qualification before product commit

Controlled integration run `33959910590` passed after the old restart tests were updated to the explicit-close contract:

- source/offline policy check;
- all Go tests on Linux;
- `go vet`;
- Windows/amd64 `internal/eco` test-binary cross-compile;
- `git diff --check`.

Normal pull-request Linux/Windows/Syft/Cosign CI is still required before merge.

## What this does **not** close

Issue #4 remains open. This slice does not yet prove all acceptance requirements for:

- durable persisted workspace revision / CAS semantics against stale writers;
- every reset / switch / migration / upgrade / downgrade path;
- rollback under every filesystem/power-loss interruption point;
- user-facing clean first-run / reopen / reset behaviour on the final Windows candidate;
- complete multi-process read-only/writer policy;
- final release or real-evidence suitability.

This slice is therefore a lifecycle safety improvement, not a claim that workspace-state integrity is complete.
