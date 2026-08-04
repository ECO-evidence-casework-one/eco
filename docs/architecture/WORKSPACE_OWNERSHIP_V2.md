# Workspace ownership V2

**Status:** active bounded P0 repair on current `main`; Slice 1 root-owner and root-creation primitives implemented, but not yet integrated into `OpenVault` or release-approved.

## Purpose

Workspace Ownership V2 replaces pathname-scoped lifecycle locking with one owner-scoped transaction over the underlying filesystem objects.

The complete design must prevent:

- stale writable snapshots replacing newer workspace metadata;
- aliases, junctions, renamed parents or bind mounts creating separate writable ownership domains;
- managed children being created before a parent/root ownership transaction exists;
- cleanup mutating a substituted filesystem object;
- fixed temporary names allowing one transaction to interfere with another;
- restore or migration leaving ownership attached to the wrong pre- or post-activation root.

## Slice 1 implemented boundary

Slice 1 now provides platform ownership primitives behind narrow internal functions:

- `acquireWorkspaceRootOwner`;
- `acquireWorkspaceCreationOwner`;
- `acquireOrCreateWorkspaceRootOwner`;
- `workspaceOwnerLease.revalidate`;
- `workspaceCreationLease.revalidate`.

These primitives are not yet called by `OpenVault`. They are qualified independently before the root-replacing restore transaction is changed.

### Root owner lease

For an existing workspace root:

1. convert the supplied route to an absolute path;
2. stat the root and require a directory;
3. capture the underlying filesystem object identity;
4. open/create the owner lock file inside that root;
5. acquire a non-blocking exclusive OS lock;
6. recapture and compare the root object identity;
7. retain the lock and object identity until explicit close.

The path string is retained only so the object can be revalidated. It is not treated as the ownership identity.

### Missing-root creation claim

For a root that does not yet exist:

1. identify and retain the existing parent object;
2. derive a bounded creation-lock name from the platform-normalised leaf name;
3. acquire a non-blocking exclusive lock in the parent;
4. revalidate the parent object;
5. create only the named root directory;
6. acquire and retain the new root's owner lease;
7. revalidate the parent again;
8. release the creation claim only after the root lease is secured.

This closes the previously unowned interval before a new workspace root exists at the primitive level.

## Platform identity and locking

### Linux

- root/parent identity: device and inode;
- owner exclusion: non-blocking exclusive `flock`;
- creation leaf identity: case-sensitive leaf bytes;
- symlink aliases resolve to the same target object and lock file.

### Windows

- root/parent identity: volume serial number and file index from `GetFileInformationByHandle`;
- owner exclusion: non-blocking exclusive `LockFileEx`;
- lock handle allows read, write and delete sharing so a later controlled restore transaction can rename owned objects without silently dropping ownership;
- creation leaf identity is lower-cased with trailing spaces and dots removed, preventing common Win32 path aliases from deriving separate creation claims.

### Other platforms

Ownership acquisition fails closed until equivalent primitives and tests exist.

## Complete target transaction

The final writable Vault lifecycle remains:

1. acquire or create root under an owned parent transaction;
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

Portable restore currently renames the active root to a checkpoint, renames a staged root into the active route and calls ordinary `OpenVault` again. Once root ownership is connected, that flow cannot remain unchanged.

The replacement must:

- keep the old active-root lease through the rollback window;
- keep the staged-root lease through validation and activation;
- permit the controlled renames without closing either lease;
- retarget and revalidate the staged lease at the active route after activation;
- avoid opening the activated root through a second ordinary acquisition;
- transfer the staged owner, key and loaded workspace into the existing Vault;
- release the old lease only after activation is proven;
- on failure, restore and retarget the old lease and keep the unrelated replacement object untouched.

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

## Temporary-file requirement

The fixed `workspace.ecodb.tmp` path must be replaced by a unique transaction-owned name created exclusively in the retained root. Cleanup may remove only a file whose identity and transaction nonce it owns.

## Rollback requirement

Every public mutator must preserve the full pre-mutation state needed to undo its change. Persistence failure must restore:

- domain records;
- selection and settings;
- audit entries;
- timestamps and build identity;
- expected revision;
- referenced object ownership state where applicable.

A method must not report success when its audit persistence failed.

## Cleanup requirement

Linux nested cleanup must recurse through retained descriptors or compare any reopened child against the exact object previously inspected. Reopening by name without identity continuity is prohibited.

Windows cleanup must retain equivalent handle/object continuity and reject reparse-point or replacement ambiguity.

## Required proof

The controlling hostile test plan is [`../testing/WORKSPACE_OWNERSHIP_V2_TEST_PLAN.md`](../testing/WORKSPACE_OWNERSHIP_V2_TEST_PLAN.md).

Slice 1 results are recorded in [`../testing/WORKSPACE_OWNERSHIP_V2_SLICE1_RESULTS.md`](../testing/WORKSPACE_OWNERSHIP_V2_SLICE1_RESULTS.md).

## Gate effect

Slice 1 proves only the independent owner and creation primitives. It does not close issue #4, approve a workspace, connect the visible V40 journey, or open any real-evidence, source-release, executable, signing, deployment or public-use gate.
