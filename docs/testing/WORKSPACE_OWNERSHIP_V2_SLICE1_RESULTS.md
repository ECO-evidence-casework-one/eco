# Workspace Ownership V2 — Slice 1 results

**Date:** 4 August 2026  
**Branch:** `development/issue-53-workspace-integrity-repair`  
**Current-main base qualified:** `3c7c69586cac195d146188e6b914db12f6391815`  
**Exact code head qualified before this record:** `b11e7144ee0d0d62fa02b0833f1b87b3876de34d`  
**GitHub Actions run:** `30913011238`

## Decision

**PASS for the isolated Workspace Ownership V2 root-owner and root-creation primitives only.**

This is not a PASS for complete workspace integrity, ordinary `OpenVault`, restore, migration, candidate state, reset or public release.

## Implemented in Slice 1

- OS-backed exclusive workspace-root owner lock;
- owner lease retained until explicit close;
- root identity represented by filesystem object identity rather than a cleaned pathname;
- Linux object identity using device and inode;
- Windows object identity using volume serial number and file index;
- Linux non-blocking exclusive `flock`;
- Windows non-blocking exclusive `LockFileEx` with delete sharing needed for later controlled rename transactions;
- fail-closed behaviour on unsupported platforms;
- parent-object creation claim acquired before a missing workspace root is created;
- root creation while the parent claim is retained;
- root owner lease acquired before the parent creation claim is released;
- Linux alias paths share one root or parent-creation lock domain;
- Windows creation keys normalise case and trailing dot/space aliases;
- root and parent identity revalidation detects path substitution.

## Hostile-path tests

The current tests prove:

- a second owner cannot acquire the same root in one process;
- another process cannot acquire the same existing root;
- releasing the owner permits a later acquisition;
- a Linux symlink alias cannot create a second writable root domain;
- parent aliases cannot create separate first-creation claims;
- one process can create a missing root under a retained claim while a second acquisition is rejected;
- two processes cannot both obtain writable ownership during first root creation;
- parent substitution is detected and an unrelated replacement sentinel is preserved;
- root substitution is detected and an unrelated replacement sentinel is preserved;
- Windows case and trailing dot/space aliases derive the same first-creation key.

## Local focused validation

- Go version used: `go1.23.2`;
- ordinary focused tests: PASS;
- focused race detector: PASS;
- focused `go vet`: PASS;
- Windows amd64 test-package cross-compilation: PASS;
- temporary Windows test binary size: `4,188,672` bytes;
- temporary Windows test binary SHA-256: `0808618d59a848472d35be71374872d4e4ea35c7e97f3aeeb401bb6fcd3a9e1a`;
- temporary test binary uploaded or distributed: NO.

## Exact-head repository validation

GitHub Actions run `30913011238` passed:

- source-policy checks;
- Linux unit and source-regression tests;
- Linux `go vet`;
- Windows native-command failure-stop self-test;
- six-stage Windows failure matrix;
- deterministic Windows build and tests.

The run artifact inventory was empty. No executable or other runnable payload was uploaded.

## Explicit remaining blockers

Slice 1 is not wired into `OpenVault` yet. That is deliberate because portable restore replaces the active root object and must transfer or retarget ownership without creating a second owner or retaining a lease on the pre-restore checkpoint.

The following remain unresolved:

1. safe owner-lease integration into ordinary Vault opening and explicit `Vault.Close`;
2. atomic owner transfer/retargeting during portable restore and rollback;
3. equivalent ownership transactions for migration, reset and candidate application state;
4. persisted monotonic revision and compare-and-swap defence;
5. unique transaction-owned workspace metadata temporary files;
6. complete in-memory rollback for every mutating path;
7. Linux nested cleanup descriptor continuity;
8. full launch → create Matter → import → close → reopen application proof;
9. Acer, keyboard, scaling, screen-reader and public-binary qualification.

None of issue #4's four P0 findings is closed by this slice alone. No real-evidence, executable, signing, source-release, deployment or public-use gate is opened.
