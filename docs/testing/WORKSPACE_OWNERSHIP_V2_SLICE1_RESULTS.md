# Workspace Ownership V2 — Slice 1 results

**Date:** 4 August 2026  
**Branch:** `development/issue-53-workspace-integrity-repair`  
**Current-main base:** `e1a4929a679762d6c04416dfdba7094b5cabc703`  
**Corrected exact code head:** `38acecc6ebf906b2c9a27618a511cc7998222a90`  
**Corrected GitHub Actions run:** `30915780542`

## Decision

**PASS for the isolated Workspace Ownership V2 root-owner, owned-root-creation and object-bound owner-retarget primitives only.**

This is not a PASS for complete workspace integrity, ordinary `OpenVault`, portable restore, migration, candidate state, reset or public release.

## Implemented boundary

- OS-backed exclusive workspace-root ownership;
- ownership retained until explicit close;
- Linux device/inode object identity;
- Windows volume/file-index object identity;
- Linux non-blocking exclusive `flock`;
- Windows object-identity-derived named kernel semaphore;
- fail-closed unsupported-platform behaviour;
- parent-object creation claim before a missing root exists;
- root creation while the parent claim remains held;
- root owner secured before parent-claim release;
- Linux alias paths share one root or creation-claim domain;
- Windows creation keys normalise case and trailing dot/space aliases;
- root and parent substitution detection;
- held owner retargeting only to the same renamed directory object.

## Hostile-path proof

Tests prove:

- second owner rejection in one process;
- second owner rejection from a real subprocess;
- owner release and later reacquisition;
- Linux root symlink alias rejection;
- Linux parent alias creation-claim rejection;
- real subprocess concurrent first-creation rejection;
- parent substitution detection with replacement sentinel preservation;
- root substitution detection with replacement sentinel preservation;
- Windows creation-alias normalisation;
- retargeting after directory rename preserves ownership;
- another owner remains blocked at the retargeted route;
- a different object cannot be accepted as the retargeted root;
- a replacement object at the old route remains untouched;
- a closed owner cannot be retargeted.

## Failed Windows design and correction

### Failed attempt

The initial Windows primitive used `LockFileEx` on an open lock file inside the owned directory. It requested read, write and delete sharing, but exact Actions run `30914784651` failed:

- `TestWorkspaceOwnerRetargetsSameRenamedObject`;
- `TestWorkspaceOwnerRetargetLeavesReplacementRouteUntouched`.

Both failed because Windows returned `Access is denied` when renaming the directory while its child lock file remained open.

This was a genuine design failure. The run is not counted as qualification evidence.

### Corrected design

Windows exclusion now uses a named kernel semaphore derived from:

- the owned directory's volume serial number;
- its file index;
- the bounded ownership role.

No child lock-file handle remains open inside the directory. The corrected exact head passed the same rename and retarget tests on Windows.

## Local focused validation of the correction

- ordinary Linux tests: PASS;
- Linux race detector: PASS;
- Linux `go vet`: PASS;
- Windows amd64 test-package cross-compilation: PASS;
- temporary Windows test binary size: `4,213,248` bytes;
- temporary Windows test binary SHA-256: `0f989f3385ba3003d20f5dcbf0b1220f582916c561ca1790bc05a7b457851ddf`;
- temporary binary uploaded or distributed: NO.

Cross-compilation was treated only as compile evidence. Exact Windows Actions remained controlling.

## Exact-head repository validation

Corrected workflow run `30915780542` passed:

- source-policy checks;
- Linux unit and source-regression tests;
- Linux `go vet`;
- Windows native-command failure-stop self-test;
- six-stage Windows failure matrix;
- Windows tests including the rename/retarget tests;
- deterministic Windows build.

Runner-only internal build:

- size: `3,938,304` bytes;
- SHA-256: `158b9066827b0eefefa4d76660dc231b9a53cceaa5e1bcca124f056115f50b16`;
- status: unsigned, private, not an end-user release.

Artifact inventory: empty. No runnable payload was uploaded.

A Node-action deprecation warning was present. It did not fail the source, tests, failure-stop matrix, deterministic build or artifact containment.

## Explicit remaining blockers

The primitives are not wired into `OpenVault` yet. Portable restore replaces the active root object and must transfer the staged and active owners safely before ordinary workspace opening can retain ownership.

Still unresolved:

1. retained ownership in ordinary `OpenVault` and explicit `Vault.Close`;
2. atomic active/stage/checkpoint owner transfer during restore and rollback;
3. migration, reset and candidate-state ownership transactions;
4. persisted monotonic revision and CAS defence;
5. unique transaction-owned metadata temporary files;
6. complete in-memory rollback for every mutating path;
7. Linux nested cleanup descriptor continuity;
8. launch → Matter → import → close → reopen application proof;
9. Acer, keyboard, scaling, screen-reader and public-binary qualification.

None of issue #4's four P0 findings is closed by these primitives alone. No real-evidence, executable, signing, source-release, deployment or public-use gate is opened.
