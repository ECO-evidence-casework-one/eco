# Current project status

**Status date:** 2 August 2026  
**Public source milestone:** V25 / repository source currently on `main`  
**Later private development candidates:** not published  
**Release position:** development only; no approved public binary

## What this repository represents

This public repository remains the controlled source milestone. Later experimental application candidates are being tested privately and must not be inferred to be releases, approved builds or repository-equivalent source snapshots.

No later executable, model, diagnostic bundle, screenshot set, private test workspace or personal evidence is included by this status update.

## Current controlling requirements

ECO is intended to remain:

- completely offline in normal operation;
- free and open source in its own stack and bundled components;
- usable on inexpensive Windows 11 hardware, with an 8 GB low-spec laptop as the controlling baseline;
- distributed eventually as one self-contained Windows application;
- local-only, with no accounts, telemetry, cloud processing or developer access to user evidence;
- explicit about the difference between original evidence, derived readings, app suggestions, user-confirmed facts and user notes.

## Current stop gates

The following remain blocked:

- public end-user binary distribution;
- use with real, sensitive or irreplaceable evidence;
- claims of reliable generative offline AI assistance;
- claims of completed production OCR or native PDF page rendering;
- stable-release or release-candidate status.

## Active grouped defect programme

The current implementation programme is recorded in issues #3 through #8:

1. preserve evidence atomically and use the verified preserved object for all downstream work;
2. prevent unintended loading of prior candidate state;
3. make offline AI instruction-faithful, source-backed and unable to pretend it performed actions;
4. keep the interface responsive during import, OCR and local-model work;
5. complete accessible navigation, scrolling and functioning controls;
6. add page-aware search, visible highlights and exact source navigation.

## Development order

1. Evidence preservation and workspace-state correctness.
2. Deterministic instruction, source and action controls.
3. Background workers, progress, cancellation and recovery.
4. Accessible core workflows and document navigation.
5. Offline model integration and wider synthetic regression testing.
6. Independent inspection, signing, clean-machine testing and release review.

## Public-record rule

Public updates must remain sanitised. They may record defects, controls, acceptance tests and release decisions, but must not expose personal evidence, private diagnostics, exploit-level details, confidential screenshots, unpublished binaries, bundled models or private test workspaces.
