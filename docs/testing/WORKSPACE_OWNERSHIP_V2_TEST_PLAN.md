# Workspace Ownership V2 test plan

**Status:** controlling hostile-path plan. Slice 1 tests are implemented for root ownership, owned root creation, aliases and root/parent substitution. Remaining sections are mandatory before issue #4 can pass.

## Test principles

- use normal public or intended application APIs for integration proofs;
- use real subprocesses for cross-process behaviour;
- preserve exact source and test identities;
- exercise Windows and Linux separately where filesystem semantics differ;
- use synthetic information only;
- never treat one-process mutex tests as cross-process proof;
- never treat pathname equality as object-identity proof;
- preserve unrelated sentinels during every hostile substitution test;
- fail the test if cleanup or rollback removes an object the transaction did not create or identify.

## Slice 1 — implemented primitive tests

### Existing root ownership

- [x] first owner acquires an existing root;
- [x] second same-process owner is rejected;
- [x] second subprocess owner is rejected;
- [x] owner release allows later acquisition;
- [x] owner lock exists inside the owned root;
- [x] root identity revalidation succeeds while unchanged.

### Alias domain

- [x] Linux symlink alias reaches the same root lock domain;
- [x] Linux parent alias reaches the same creation-claim domain;
- [x] Windows creation keys normalise case aliases;
- [x] Windows creation keys normalise trailing dot/space aliases.

### First creation

- [x] a missing root is created only while a parent creation claim is held;
- [x] root ownership is obtained before the creation claim is released;
- [x] a second acquisition is rejected after first creation;
- [x] a real subprocess concurrent-creation attempt is rejected.

### Substitution

- [x] Linux parent replacement is detected;
- [x] Linux replacement-parent sentinel survives;
- [x] Linux root replacement is detected;
- [x] Linux replacement-root sentinel survives.

## Slice 2 — Vault integration and retained lifetime

- [ ] `OpenVault` acquires or creates the root owner before `objects`, key or metadata files are created;
- [ ] the Vault retains ownership for its complete writable lifetime;
- [ ] a second same-route `OpenVault` fails with `ErrWorkspaceInUse`;
- [ ] a second alias-route `OpenVault` fails;
- [ ] a second subprocess `OpenVault` fails;
- [ ] `Vault.Close` releases exactly once and erases the key;
- [ ] all public operations fail after close;
- [ ] failed opening releases ownership and removes only transaction-owned incomplete children;
- [ ] existing callers and tests close Vaults explicitly or through test cleanup;
- [ ] the real application closes the Vault on normal window shutdown.

## Slice 3 — persisted revision and unique metadata transaction

- [ ] a new workspace begins with a defined persisted revision;
- [ ] opening records the exact authenticated revision;
- [ ] each successful save increments exactly once;
- [ ] a stale expected revision is rejected before file replacement;
- [ ] rejected stale state does not alter memory, metadata or audit;
- [ ] temporary metadata filenames are unique and created exclusively;
- [ ] commit revalidates root identity and temporary-file ownership;
- [ ] failed write, sync or rename leaves the prior authenticated metadata active;
- [ ] expected revision advances only after durable success;
- [ ] abandoned transaction cleanup removes only the owned temporary object.

## Slice 4 — public mutator rollback matrix

Inject deterministic persistence failures into every mutating path, including:

- [ ] `Save`;
- [ ] `AddChange`;
- [ ] selected page and selected evidence;
- [ ] low-sensory and reduced-motion settings;
- [ ] Matter creation and edits;
- [ ] image rotation;
- [ ] evidence import preservation stages;
- [ ] OCR application;
- [ ] Ask ECO question/audit persistence;
- [ ] evidence verification success/failure updates;
- [ ] backup audit changes;
- [ ] reset, migration and restore state updates.

For each path prove:

- [ ] returned error is non-nil and bounded;
- [ ] the in-memory snapshot equals its pre-call snapshot;
- [ ] the authenticated on-disk workspace equals its pre-call state;
- [ ] no success/audit entry remains;
- [ ] no unowned temporary or object file is removed.

## Slice 5 — restore owner transfer

Use a real active Vault and staged restored Vault.

- [ ] active root owner is retained through the rollback window;
- [ ] staged root owner is retained through validation and activation;
- [ ] controlled rename succeeds with retained Windows handles and Linux descriptors;
- [ ] stage owner is retargeted and revalidated at the active route;
- [ ] old owner is retargeted at the checkpoint route;
- [ ] no second ordinary `OpenVault` acquisition is attempted on the activated stage;
- [ ] existing Vault receives staged owner, key, objects and workspace atomically;
- [ ] success releases the old checkpoint owner only after activation proof;
- [ ] failure restores the old root and ownership;
- [ ] failure preserves unrelated replacement sentinels;
- [ ] no lock remains attached to the wrong root after success or rollback;
- [ ] restored Vault can close and reopen normally.

## Slice 6 — migration, reset and candidate state

### Migration

- [ ] source, stage, checkpoint and recovery roles are object-bound;
- [ ] concurrent migration/open attempts cannot split ownership;
- [ ] interruption recovery identifies exact objects, not only names;
- [ ] rollback preserves unrelated replacements.

### Reset

- [ ] selected workspace ownership is retained for the full reset;
- [ ] only referenced or transaction-owned objects are removed;
- [ ] alias or parent substitution fails closed;
- [ ] unrelated sentinels survive;
- [ ] persistence failure restores the pre-reset workspace.

### Candidate application state

- [ ] parent claim exists before first candidate state is created;
- [ ] concurrent first launch permits one creator;
- [ ] candidate-state revision/CAS prevents stale replacement;
- [ ] selected workspace identity cannot be overwritten by an older process;
- [ ] aliases cannot create multiple candidate-state domains.

## Slice 7 — Linux nested cleanup continuity

Construct a nested directory tree containing transaction-owned cleanup candidates and unrelated sentinels.

- [ ] recursion proceeds through retained descriptors; or
- [ ] every reopened child is compared to the exact previously inspected object;
- [ ] substitute a same-mount child after inspection;
- [ ] cleanup rejects the substituted child;
- [ ] replacement sentinel survives;
- [ ] original transaction object remains recoverable or is safely cleaned through its retained identity;
- [ ] symlinks, bind-mount-like aliases and non-directory replacements fail closed.

## Slice 8 — owner death and recovery

- [ ] subprocess holding a root owner is terminated;
- [ ] OS releases the lock;
- [ ] later opening revalidates the root and authenticated metadata;
- [ ] abandoned unique transaction files are identified by bounded transaction records;
- [ ] cleanup never removes a file merely because its name resembles a temporary file;
- [ ] recovery either completes or rolls back deterministically.

## Slice 9 — preservation and lifecycle regressions

Rerun the complete existing suite for:

- [ ] encrypted import and duplicate handling;
- [ ] immutable source verification;
- [ ] preservation interruption and recovery;
- [ ] OCR provenance and rollback;
- [ ] portable backup creation;
- [ ] restore round trip and wrong-passphrase isolation;
- [ ] migration compatibility and interruption recovery;
- [ ] reset isolation;
- [ ] source-policy and offline checks;
- [ ] Windows failure-stop and deterministic build checks.

## Slice 10 — actual application journey

On a clean synthetic workspace:

1. [ ] launch ECO;
2. [ ] create a new workspace;
3. [ ] create a named Matter;
4. [ ] import synthetic evidence with visible progress;
5. [ ] observe preserved and verified evidence state;
6. [ ] close ECO through the ordinary window path;
7. [ ] verify the owner is released;
8. [ ] reopen the same workspace;
9. [ ] verify the same Matter and evidence state;
10. [ ] attempt a second process and alias path while the first is open;
11. [ ] verify both are rejected without damaging the active workspace.

## Qualification outputs

Before issue #4 or public release can pass, preserve:

- exact base and head SHAs;
- exact changed-file inventory;
- source hashes or Git blob identities;
- local ordinary/race/vet results;
- Linux subprocess raw output;
- Windows subprocess raw output;
- exact-head GitHub Actions jobs and conclusions;
- artifact inventory;
- temporary binary identity and deletion/containment status;
- known limitations;
- same-account versus independent-review relationship.

## Stop rule

Any failed ownership, revision, rollback, substitution, cleanup or lifecycle property blocks merge and public release. The 9 August target cannot downgrade or waive a failed P0 property.
