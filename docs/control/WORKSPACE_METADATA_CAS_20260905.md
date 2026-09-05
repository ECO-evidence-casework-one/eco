# Workspace metadata compare-and-swap — 5 September 2026

**Issue:** #4 clean, explicit workspace state  
**Baseline:** current `main` after PR #125  
**Scope:** stale authenticated metadata rejection for the live writable Vault

## Control

Each persisted workspace now carries a monotonically advancing `revision` and the identifier of the owner transaction that wrote it. A live Vault also retains the SHA-256 of the exact encrypted `workspace.ecodb` bytes it loaded or wrote and the hash-chain head of the persisted `ChangeRecord` ledger.

Before an authenticated metadata save, ECO revalidates the object-bound workspace owner and then requires the on-disk metadata to match the expected:

- encrypted metadata SHA-256;
- workspace revision;
- last owner transaction;
- audit-chain head.

If any of those values changed, disappeared or fail authentication, the save returns `ErrWorkspaceStale` and does not overwrite the unexpected metadata. A successful write advances the revision and records the current owner transaction.

Portable restore transfers the staged Vault's CAS state together with its owner/key/workspace when the staged object becomes active, so the next ordinary save continues from the activated metadata rather than from stale pre-restore state.

## Compatibility

Legacy schema-1 workspaces that do not contain `revision` or `last_owner_txn` load as revision zero and are upgraded on their next successful authenticated save. Unknown JSON fields remain compatible with older schema-1 readers, so this slice does not change the workspace schema number.

## Limits

Issue #4 remains open. This slice does not yet implement the user-facing first-run/open/reset chooser, complete migration/upgrade/downgrade evidence, all interruption-point rollback proofs, or the final read-only/writer policy.
