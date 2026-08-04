# Workspace ownership V2

**Status:** implementation architecture for issue #53 Phase B and issue #4.  
**Branch:** `development/issue-53-workspace-integrity-repair`  
**Date:** 4 August 2026

## Problem statement

The current application protects a Vault only with process-local mutexes. The stale PR #11 implementation added cross-process primitives but released ordinary open ownership before returning the writable Vault and keyed its process registry by pathname. Those behaviours allow stale writers and alias-split ownership.

Workspace Ownership V2 must treat writable access as one retained transaction over underlying filesystem objects, not as a short-lived pathname check.

## Security property

For one underlying workspace object, at most one writable ECO ownership domain may exist at a time across all processes and all path aliases.

A process that does not hold the retained owner lease must not:

- create managed workspace children;
- replace workspace metadata;
- replace candidate-state selection;
- migrate, restore or reset the workspace;
- clean interrupted transaction objects;
- index or expose a mutation as completed.

A stale process must fail before replacing a newer persisted revision.

## Ownership scopes

### Parent creation claim

Creating a workspace begins by retaining and identifying the selected parent directory before any ECO-managed child exists.

The claim binds:

- retained parent handle/descriptor;
- exact parent object identity;
- intended child name;
- a transaction nonce;
- a parent-local exclusive claim object created without following links.

The new workspace directory and all managed children are created relative to that retained parent. Cleanup may remove only objects created by the same nonce and exact identities.

### Workspace owner lease

Opening an existing workspace retains:

- parent directory;
- workspace root directory;
- encrypted object directory;
- key and metadata control files;
- an exclusive OS ownership primitive;
- exact object identities for every retained role.

The lease remains attached to the writable `Vault` until `Vault.Close` completes.

It is not released after inspection, authentication, first save or UI startup.

### Candidate-state owner lease

Candidate-specific application state uses the same principles:

- retain the state parent;
- acquire a creation/open lease before reading missing state as authoritative;
- bind replacement to exact current object identity and revision;
- write through a unique transaction-owned temporary file;
- fsync the replacement and parent where supported;
- reject stale revisions.

## Object identity

Logical path strings are display and routing information only. They are not ownership identity.

### Linux/amd64

Identity must include enough kernel-provided information to detect substitution and aliases, including:

- device;
- inode;
- mount identity where available;
- object type.

Operations must be descriptor-relative with no-follow flags. Recursive cleanup must recurse through retained child descriptors or verify that a reopened child matches the exact previously inspected identity before mutation.

### Windows/amd64

Identity must use retained handles opened without unsafe reparse traversal and stable file identity information, including volume and file ID.

The exclusive owner primitive must be shared across path aliases by deriving or storing it relative to a retained object/parent identity rather than the user-supplied logical path.

Junctions, symbolic links and substituted parents must fail closed unless the retained target identity is exactly the authorised workspace object.

### Unsupported platforms

Writable workspace ownership fails closed until equivalent object-bound primitives and tests exist.

## Persisted revision and CAS

Cross-process exclusion is the primary control. Persisted compare-and-swap is defence in depth and protects against stale in-memory state, lost locks and implementation errors.

Workspace metadata includes a monotonically increasing `revision`.

Every mutating transaction records:

- expected revision;
- next revision;
- current metadata object identity;
- replacement transaction nonce.

Before replacement, the implementation verifies that:

- the owner lease is still valid;
- the root and control-file identities remain valid;
- the persisted revision equals the expected revision;
- no newer replacement has appeared.

A revision mismatch returns a typed stale-workspace error. The in-memory mutation is rolled back to the complete pre-operation snapshot.

## Atomic replacement

Fixed shared temporary names such as `workspace.ecodb.tmp` are prohibited.

Each replacement uses a unique owner-scoped name containing an unpredictable nonce. The temporary object must be created exclusively and without following links.

Required order:

1. validate owner lease and expected revision;
2. create unique temporary object relative to retained root;
3. write complete authenticated replacement;
4. flush temporary object;
5. revalidate owner lease, root identity, target identity and expected revision;
6. atomically replace the authorised target;
7. flush retained parent/root where supported;
8. retain or reopen the new target by exact identity;
9. update in-memory revision only after durable replacement;
10. remove transaction-owned leftovers.

## Mutation rollback

Every public mutating path must establish a complete in-memory rollback point before mutation.

This includes:

- Matter creation and editing;
- evidence preservation state transitions;
- OCR and extraction results;
- Ask records and citations;
- settings;
- audit/change records;
- selection state;
- backup/restore activation;
- migration;
- reset.

If persistence or final verification fails, the entire affected in-memory state returns to its prior value. Partial audit entries, counters, selections and derived segments must not remain.

## Lifecycle

`Vault` gains explicit lifecycle state:

- writable owner lease;
- retained object binding;
- persisted revision;
- closed flag.

All mutating methods require an active writable owner lease. `Save` on a closed or non-owner Vault fails.

`Vault.Close`:

1. prevents new operations;
2. waits for active operations to finish or cancels them according to contract;
3. zeroes sensitive key material owned by the Vault;
4. closes retained control-file and directory handles/descriptors;
5. releases the OS owner primitive last;
6. is idempotent;
7. reports cleanup errors without reopening or mutating by pathname.

## Creation and first launch

Workspace and candidate-state creation are transactions, not optimistic setup.

Concurrent creation must produce:

- one successful owner and complete state; and
- one clean refusal or safe reopen decision.

It must never produce two keys, two identities, split object directories, mixed metadata or one process deleting the other process's files.

## Migration, restore and reset

These operations must use the same owner lease and revision model.

A checkpoint or staging directory is transaction-owned and object-bound. Recovery records may identify transaction paths for routing but cleanup requires retained identity or exact revalidation.

Reset may remove only:

- the selected workspace's referenced encrypted objects; and
- exact transaction-owned leftovers.

It may not recursively clean arbitrary path contents.

## Error model

Typed errors must distinguish:

- workspace already owned;
- stale revision;
- object identity changed;
- alias/reparse traversal blocked;
- recovery required;
- persistence failed with rollback completed;
- persistence failed and recovery record retained;
- unsupported platform.

User-facing messages must explain what was protected and whether anything changed.

## Integration sequence

1. Add platform owner-lease and object-identity primitives with focused tests.
2. Add revision/CAS metadata and backward-compatible read logic.
3. Introduce `Vault.Close` and retain ownership for full lifetime.
4. Convert `OpenVault` and creation to owner-scoped transactions.
5. Convert `Save` and every public mutation to CAS plus rollback.
6. Convert candidate state.
7. Convert migration, restore, reset and cleanup.
8. Add real subprocess tests.
9. Run all preservation and application regressions.
10. Freeze one exact SHA for independent delta-first review.

## Prohibited shortcuts

- pathname-derived ownership domains;
- lock files inside an unowned new workspace as the first claim;
- releasing ownership before returning a writable Vault;
- treating process-local mutexes as cross-process safety;
- fixed temporary filenames;
- recursive cleanup by pathname after closing the inspected descriptor/handle;
- updating in-memory state before durable persistence without rollback;
- allowing the public-release deadline to narrow or waive these properties.
