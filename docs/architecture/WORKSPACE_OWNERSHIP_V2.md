# Workspace ownership V2

**Status:** active bounded P0 repair on current `main`; root ownership, owned first creation and object-bound lease retargeting are implemented and qualified as isolated primitives. They are not yet integrated into `OpenVault`, portable restore or public release.

## Purpose

Workspace Ownership V2 replaces pathname-scoped lifecycle locking with one owner-scoped transaction over the underlying filesystem objects.

The complete design must prevent:

- stale writable snapshots replacing newer workspace metadata;
- aliases, junctions, renamed parents or bind mounts creating separate writable ownership domains;
- managed children being created before a parent/root ownership transaction exists;
- cleanup mutating a substituted filesystem object;
- fixed temporary names allowing one transaction to interfere with another;
- restore or migration leaving ownership attached to the wrong pre- or post-activation root.

## Implemented primitive boundary

The current branch provides:

- `acquireWorkspaceRootOwner`;
- `acquireWorkspaceCreationOwner`;
- `acquireOrCreateWorkspaceRootOwner`;
- `workspaceOwnerLease.revalidate`;
- `workspaceOwnerLease.retarget`;
- `workspaceCreationLease.revalidate`.

These primitives are not yet called by `OpenVault`. They are being qualified independently before the root-replacing restore transaction is changed.

### Existing root owner

1. Resolve the supplied route to an absolute path.
2. Require the routed object to be a directory.
3. Capture the underlying filesystem object identity.
4. Acquire non-blocking cross-process ownership for that exact object.
5. Recapture and compare the root identity.
6. Retain the owner until explicit close.

The path is routing information used for later identity revalidation. It is not the ownership identity.

### Missing-root creation claim

1. Identify the existing parent object.
2. Derive a bounded creation identity from the platform-normalised leaf name.
3. Acquire cross-process ownership for that parent/leaf creation role.
4. Revalidate the parent object.
5. Create only the named root directory.
6. Acquire the new root owner.
7. Revalidate the parent.
8. Release the creation claim only after the root owner is secured.

This closes the previously unowned interval before a new workspace root exists at the primitive level.

### Object-bound retargeting

A held root owner may be retargeted after the exact owned directory object is renamed:

1. resolve the proposed new route;
2. capture the object identity at that route;
3. require it to equal the identity already owned;
4. update only the lease's routing path;
5. revalidate the same object at the new route.

Retargeting does not acquire a second owner and does not accept a replacement object. It is required for later active/stage/checkpoint restore handoff.

## Platform controls

### Linux

- object identity: device and inode;
- cross-process exclusion: non-blocking exclusive `flock` on `.eco-owner-v2.lock` or the bounded parent creation-lock file;
- creation leaf identity: case-sensitive leaf bytes;
- symlink aliases resolve to the same target object and lock domain;
- the open lock file does not prevent directory rename on the qualified Linux path.

### Windows

- object identity: volume serial number and file index from `GetFileInformationByHandle`;
- cross-process exclusion: named kernel semaphore derived from the exact object identity and bounded lock role;
- acquisition: `CreateSemaphoreW` plus zero-timeout `WaitForSingleObject`;
- release: `ReleaseSemaphore` plus `CloseHandle`;
- no child lock-file handle remains open inside the owned directory, so controlled directory rename is not pinned by the ownership primitive;
- creation leaf identity is lower-cased with trailing spaces and dots removed, preventing common Win32 aliases from deriving separate creation claims.

#### Rejected Windows design

An earlier Slice 1 attempt used `LockFileEx` on a lock file inside the owned root with read/write/delete sharing. Exact Windows Actions run `30914784651` proved this assumption wrong: `os.Rename(stage, active)` failed with `Access is denied` while the child lock file remained open.

That design is rejected and not controlling. The named-semaphore replacement passed the same Windows rename/retarget tests in exact-head run `30915780542`.

### Other platforms

Ownership acquisition fails closed until equivalent primitives and tests exist.

## Complete target transaction

The final writable Vault lifecycle remains:

1. acquire or create the root under an owned parent transaction;
2. retain the root owner for the complete writable Vault lifetime;
3. open or create managed children only after ownership exists;
4. load authenticated metadata and its persisted revision;
5. before each mutation, revalidate root identity and expected persisted revision;
6. write to a unique transaction-owned temporary object;
7. durably commit and advance revision;
8. roll back all in-memory mutation if persistence fails;
9. during restore/migration, own active, staged and checkpoint objects and transfer ownership atomically;
10. release ownership only through explicit `Vault.Close` or a proven rollback path.

## Restore ownership requirement

Portable restore currently renames the active root to a checkpoint, renames a staged root into the active route and calls ordinary `OpenVault` again. Once root ownership is integrated, that flow cannot remain unchanged.

The replacement must:

- keep the old active-root owner through the rollback window;
- keep the staged-root owner through validation and activation;
- rename both exact owned objects without closing ownership;
- retarget the old owner to the checkpoint route;
- retarget the staged owner to the active route;
- avoid a second ordinary acquisition on the activated stage;
- transfer staged owner, key, objects and workspace into the existing Vault;
- release the old owner only after activation is proven;
- on failure, restore and retarget the old owner while preserving unrelated replacement objects.

## Revision/CAS requirement

Root ownership is the primary writer exclusion. Persisted revision is mandatory defence in depth.

The encrypted workspace metadata must carry a monotonic revision. A Vault stores the exact revision it loaded. Every save must:

1. authenticate and read the currently persisted revision under ownership;
2. require it to equal the Vault's expected revision;
3. reject stale state before replacing any file;
4. write the next revision to a unique temporary file;
5. atomically replace the metadata;
6. update the in-memory expected revision only after durable success.

A revision conflict is a controlled failure and never a last-writer-wins event.

## Temporary-file and rollback requirements

The fixed `workspace.ecodb.tmp` path must be replaced by a unique transaction-owned name created exclusively in the retained root. Cleanup may remove only a file whose identity and transaction nonce it owns.

Every public mutator must preserve the complete pre-mutation state. Persistence failure must restore domain records, selections, settings, audit entries, timestamps, build identity, expected revision and referenced object state where applicable.

## Cleanup requirement

Linux nested cleanup must recurse through retained descriptors or compare any reopened child against the exact object previously inspected. Reopening by name without identity continuity is prohibited.

Windows cleanup must retain equivalent handle/object continuity and reject reparse-point or replacement ambiguity.

## Required proof

- hostile plan: [`../testing/WORKSPACE_OWNERSHIP_V2_TEST_PLAN.md`](../testing/WORKSPACE_OWNERSHIP_V2_TEST_PLAN.md);
- primitive results: [`../testing/WORKSPACE_OWNERSHIP_V2_SLICE1_RESULTS.md`](../testing/WORKSPACE_OWNERSHIP_V2_SLICE1_RESULTS.md).

## Gate effect

The current result proves only isolated ownership, creation and retarget primitives. It does not close issue #4, approve ordinary workspace opening, connect the visible V40 journey, or open any real-evidence, source-release, executable, signing, deployment or public-use gate.
