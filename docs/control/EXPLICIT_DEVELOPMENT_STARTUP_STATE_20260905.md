# Explicit development startup state — 5 September 2026

**Issue:** #4 clean, explicit workspace state  
**Baseline:** canonical `main` after PR #127  
**Pre-PR qualification:** run `33962912963` — PASS

## Implemented control

Windows development startup no longer silently selects the old shared `EvidenceCaseworkOne/V25N2` workspace.

The exact source candidate derives its normal development workspace from schema, BuildID and embedded source commit. Startup behaviour is explicit when state already exists:

- **Continue this candidate** opens only the current candidate-specific workspace;
- **Open existing** uses the native folder chooser, then performs a read-only ECO workspace marker check before `OpenVault` can touch the selected path;
- **Start clean** never deletes the prior candidate workspace. It acquires that exact workspace's owner, renames the owned directory to a sibling `.previous-*` archive, releases ownership, and lets `OpenVault` create a fresh candidate workspace at the normal route;
- **Cancel** exits without opening or changing workspace state.

If an exact candidate has no workspace, its default first launch is clean. If the legacy `V25N2` workspace exists, ECO tells the user and offers clean start or explicit existing-workspace selection instead of silently importing it.

Both Windows and non-Windows entry points now deliberately close the Vault on normal return so the in-memory key is zeroed and workspace ownership is released by ECO rather than only by process teardown.

## Safety properties tested

Backend tests prove:

- validating an arbitrary folder does not create `vault.key`, `workspace.ecodb` or `objects/`;
- a closed real ECO Vault passes read-only existing-workspace validation;
- Start Clean preserves the prior workspace, all of its Matter state, and leaves it reopenable;
- the fresh replacement workspace starts with no inherited evidence, Matters or questions;
- a live-owned workspace cannot be archived out from under another ECO process;
- a missing candidate workspace makes archive preparation a harmless no-op.

Native source regression tests require the candidate-bound chooser, validation, archive-and-clean flow and explicit `Vault.Close()` path, and reject the former silent fixed-root startup assignment.

## Qualification

Run `33962912963` passed:

- source/offline policy;
- `go test ./...`;
- `go vet ./...`;
- Windows/amd64 internal-package test cross-compile;
- Windows/amd64 native-UI test cross-compile;
- Windows/amd64 application build;
- `git diff --check`.

The first attempt failed before product commit because `main_other.go` retained an unused `path/filepath` import after the root change. That was corrected without changing behaviour and the full gate was rerun successfully.

## Remaining issue #4 work

Issue #4 remains open for complete upgrade/downgrade and incompatible-format policy, migration/checkpoint interruption tests, broader rollback/crash-recovery evidence, and any final reset/history UX needed after real-machine testing.
