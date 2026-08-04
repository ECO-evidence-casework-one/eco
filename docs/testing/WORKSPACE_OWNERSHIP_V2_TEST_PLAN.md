# Workspace Ownership V2 test plan

**Issues:** #4, #53 and #69  
**Status:** mandatory pre-implementation and qualification plan  
**Date:** 4 August 2026

The tests below must exercise normal public application APIs. Tests that call hidden lock helpers directly are supporting evidence only and cannot replace subprocess proof.

## Test harness rules

- Use synthetic workspace content only.
- Use separate OS processes for ownership and stale-writer tests.
- Record exact process roles, paths, object identities, revisions, exit codes and final persisted state.
- Use deterministic synchronization barriers rather than arbitrary sleeps wherever possible.
- Preserve a raw event log for each scenario.
- Verify both the intended workspace and unrelated sentinel objects after every hostile scenario.
- Run Linux/amd64 and Windows/amd64 platform suites.
- Unsupported platforms must prove fail-closed behaviour.

## OW-01 — stale writer cannot overwrite newer state

1. Create a workspace and close it.
2. Process A opens writable and pauses after reading revision N.
3. Process B attempts writable open through the same path.
4. Expected primary result: B is refused while A owns the workspace.
5. After A closes, B opens revision N.
6. B pauses with a stale in-memory snapshot.
7. A reopens, saves a Matter change as revision N+1 and closes.
8. B attempts to save its older snapshot with expected revision N.
9. B must receive typed stale-revision failure.
10. B's complete in-memory mutation must roll back.
11. Persisted revision N+1 and A's Matter change must remain intact.

## OW-02 — alias cannot split writable ownership

Run the scenario for every supported alias available on the platform:

- relative versus absolute path;
- case variation on Windows;
- directory symbolic link;
- Windows junction;
- renamed parent with retained handle;
- Linux bind mount where CI capability permits;
- alternate path spelling that resolves to the same object.

Process A opens through path 1. Process B opens through path 2.

Exactly one writable owner may exist. B must be refused or attached as an explicitly read-only view if such a mode is later implemented. It must never receive a second writable domain.

## OW-03 — parent/root substitution

1. Retain the selected parent and workspace root.
2. Pause before the first managed write.
3. Replace or rename the logical parent/root path and place an unrelated directory at the old name.
4. Resume the transaction.
5. The operation must fail with object-identity-changed status.
6. No managed file may be written into the unrelated replacement.
7. An unrelated sentinel must remain byte-identical.
8. Cleanup may affect only exact transaction-owned objects under retained identities.

## OW-04 — Linux nested cleanup substitution

1. Create a transaction-owned nested recovery directory containing an expected child.
2. Cause cleanup to inspect the child and pause before recursion.
3. Rename the inspected child and substitute another same-mount directory containing a sentinel at the original name.
4. Resume cleanup.
5. The substituted sentinel must survive.
6. Cleanup must either recurse through the retained original descriptor or reject the identity mismatch.
7. Recovery state must remain valid and explain what was retained.

## OW-05 — concurrent new workspace creation

1. Select one empty parent and one intended workspace name.
2. Start two normal create-workspace calls at the same barrier.
3. Exactly one process may create the workspace.
4. The second must return a clean owned/conflict result.
5. Final state must contain exactly one:
   - workspace identity;
   - key;
   - metadata file;
   - object directory;
   - initial revision.
6. No process may delete or replace objects created by the other process.
7. The successful workspace must reopen and pass authentication.

## OW-06 — concurrent first candidate launch

1. Use an empty candidate-state parent.
2. Start two normal application first-launch calls simultaneously.
3. Exactly one candidate identity/state transaction becomes authoritative.
4. The other process must reopen the completed state or fail cleanly.
5. No split selection state, duplicate initial workspaces or mixed candidate identity may exist.
6. Every audit event must form one valid chain.

## OW-07 — fixed temporary name attack eliminated

1. Place a hostile file or link at the old fixed `workspace.ecodb.tmp` name.
2. Perform a normal save.
3. The new implementation must not use or remove the hostile object.
4. A unique exclusive transaction temporary object must be used.
5. The sentinel at the old name remains unchanged.

## OW-08 — persistence failure rollback matrix

Inject failure at each durable stage:

- temporary creation;
- write;
- encryption/authentication;
- temporary flush;
- pre-replacement revalidation;
- atomic replacement;
- parent/root flush;
- post-replacement retain/reopen;
- final identity verification.

Repeat for each public mutation family:

- Matter create/edit;
- evidence preservation transition;
- extracted reading;
- OCR result;
- evidence linking;
- Ask/question record;
- audit/change record;
- settings;
- selected page/item;
- migration activation;
- restore activation;
- reset.

For every failure:

- no false completed state;
- complete in-memory rollback where the old state remains authoritative;
- exact recoverable record where rollback cannot safely complete;
- no unrelated cleanup;
- next reopen produces one explicit valid state.

## OW-09 — owner death and recovery

Terminate the owning process at each important phase:

- after owner acquisition;
- after temporary creation;
- after partial write;
- after complete temporary flush;
- immediately before replacement;
- immediately after replacement;
- before parent flush;
- before in-memory revision update.

A new process must either:

- open the last complete authenticated revision; or
- enter explicit recovery for the exact transaction.

It must not guess, merge partial metadata or delete unrelated objects.

## OW-10 — close lifecycle

- `Vault.Close` is idempotent.
- Mutations started after closing fail.
- Active operations cannot outlive released ownership without a controlled cancellation/finish contract.
- Sensitive key material owned by the Vault is zeroed before release completes.
- The OS owner primitive is released last.
- A second process can acquire writable ownership only after close completes.

## OW-11 — read-only snapshot isolation

A `Snapshot` returned to UI code must not allow later mutation of nested slices, maps, regions, receipts or settings inside the live workspace.

Add targeted mutation tests for every nested reference type. This test does not replace process ownership but prevents accidental in-process stale or cross-thread mutation.

## OW-12 — migration, restore and reset regression

Re-run the complete issue #3 suites with V2 ownership active:

- preservation interruption and recovery;
- verified object access;
- authenticated backup;
- staged restore;
- migration checkpoint and rollback;
- selected workspace reset;
- unrelated sentinel preservation;
- source hash and object identity enforcement.

## OW-13 — real UI/application journey

On the exact candidate:

1. launch ECO;
2. create an empty workspace;
3. create a Matter;
4. add synthetic evidence;
5. close ECO cleanly;
6. start a second process while the first is still open and prove writable refusal;
7. close the first process;
8. reopen the same workspace and Matter;
9. verify revision, Matter, evidence and audit continuity.

Run on the Acer baseline and in Windows CI where possible.

## Qualification outputs

The workspace repair PR must provide:

- exact base and head SHAs;
- changed-file inventory;
- ordinary and race-test results;
- raw subprocess event logs;
- Linux and Windows identity/alias scenario results;
- failure-injection matrix;
- preservation/migration/restore/reset regression results;
- Windows deterministic build receipt;
- exact run artifact inventory proving no uploaded executable;
- known limitations;
- independent delta-first review decision.

No single green CI check or unit-test percentage may be used to claim these properties without the scenario evidence above.
