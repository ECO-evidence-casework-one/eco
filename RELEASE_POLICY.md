# Release policy

## Release classes

- **Source development snapshot:** source and documentation only.
- **Unsigned provenance artifact:** automated build for testing the build chain; not recommended for ordinary users and not published as a normal download.
- **Signed development preview:** trusted signature, synthetic/non-sensitive testing only.
- **Release candidate:** signed, migration-tested and independently inspected.
- **Stable release:** all release gates satisfied and no unresolved critical or high-priority defects.

## Mandatory release contents

- version and build identity;
- source commit;
- signed executable or installer where applicable;
- SHA-256 checksum;
- signature-verification evidence;
- source archive;
- licence and third-party notices;
- SBOM;
- test summary;
- known limitations;
- migration and rollback information.

## Signing gate

No Windows binary is described as suitable for ordinary users until:

- a trusted Authenticode signature validates;
- the signing request was manually approved;
- the signed file matches the automated source build and version metadata;
- Smart App Control testing has passed on a clean Windows system;
- no file was modified after signing.

## Functional and evidence-integrity gates

A signed file is not sufficient by itself. Before release-candidate status, ECO must also have reproducible synthetic test evidence showing that:

- evidence is preserved atomically and all downstream processing uses the verified preserved object;
- incomplete, changed or mismatched evidence is blocked from indexing, citation and retrieval;
- first-run, reopen, migration, reset and rollback behaviour is explicit and recoverable;
- local AI output is instruction-faithful, source-backed and unable to claim unreceipted actions;
- import, hashing, extraction, OCR, indexing and model work do not block the interface;
- core workflows pass keyboard, screen-reader, scrolling and supported high-DPI tests;
- document search reports deterministic page-aware results with visible highlighting and source navigation;
- low-resource and long-duration testing passes on the controlling 8 GB Windows baseline.

## Current release-gate status — 2 August 2026

The public repository remains a source development milestone. Later private candidates are not releases and are not represented by this repository.

The following remain blocked:

- ordinary-user binary distribution;
- use with real, sensitive or irreplaceable evidence;
- release-candidate or stable status;
- claims of reliable generative offline AI, production OCR or complete native PDF investigation.

The active grouped blockers and acceptance tests are tracked in issues #3 through #8. All P0 blockers and all release-relevant P1 blockers must be closed with independent regression evidence before the release position is reconsidered.

## Prohibitions

Official releases must not:

- ask users to disable Smart App Control or antivirus protection;
- silently download components or models;
- claim OCR or AI capability not actually present;
- include real personal evidence in tests or examples;
- overwrite or destructively migrate a workspace without a recoverable checkpoint;
- publish an unsigned CI artifact as a normal end-user download.
