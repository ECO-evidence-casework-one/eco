# Release policy

## Release and test classes

- **Source development snapshot:** source and documentation only.
- **Private controlled unsigned test handoff:** exact recipient, purpose, source and executable hashes, build identity, warning, expiry or withdrawal record; synthetic and non-sensitive testing only.
- **Signed development preview:** trusted signature, synthetic/non-sensitive testing only.
- **Release candidate:** signed, migration-tested and independently inspected.
- **Stable release:** all release gates satisfied and no unresolved P0 or release-relevant P1 defects.

There is no approved public unsigned-binary class. A GitHub Actions artifact, test archive, provenance package or short-lived download is still a distribution surface when it contains runnable code.

## Distribution surfaces

The same release gates apply to every project-controlled public binary route, including:

- GitHub Release assets;
- GitHub Actions artifacts;
- public preview or test pages;
- public mirrors and archives;
- installer, DLL, model/runtime or executable payload archives.

While issue #24 and the public-binary gate remain open, public GitHub Actions may build and test an executable internally but must not upload runnable payloads. Historical unapproved Actions artifacts must be inventoried and deleted where permitted or allowed to expire with the audit record preserved.

Controlled Acer or user testing uses a private explicit handoff outside the public repository. It is not an official release and must not be represented as ordinary-user distribution.

## Mandatory release contents

- version and build identity;
- source commit;
- signed executable or installer where applicable;
- SHA-256 checksum;
- signature-verification evidence;
- source archive;
- licence and third-party notices;
- actual-build SBOM;
- test summary;
- known limitations;
- migration and rollback information;
- accountable publisher and supported response routes;
- public distribution-surface inventory.

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
- first-run, reopen, creation, migration, reset and rollback behaviour is explicit, isolated and recoverable;
- ordinary concurrent writers cannot silently erase newer workspace state;
- local AI output is instruction-faithful, source-backed and unable to claim unreceipted actions;
- health-related and mixed-purpose clinical material cannot enter generated processing outside the approved boundary;
- import, hashing, extraction, OCR, indexing and model work do not block the interface;
- core workflows pass keyboard, screen-reader, scrolling and supported high-DPI tests;
- document search reports deterministic page-aware results with visible highlighting and source navigation;
- diagnostics are privacy-safe and public offline/network claims match the exact bundled runtime;
- low-resource and long-duration testing passes on the controlling 8 GB Windows baseline.

## Current release-gate status — 3 August 2026

The public repository remains a source-development project. `VERSION` still records the V25 N2 P1 source milestone. Later source commits, draft pull requests and private candidates are not releases.

The following remain blocked:

- ordinary-user binary distribution through GitHub Releases, Actions artifacts or another public route;
- use with real, sensitive or irreplaceable evidence;
- release-candidate or stable status;
- claims of reliable generative offline AI, production OCR or complete native PDF investigation;
- public-sector, institutional, healthcare, clinical or EU deployment;
- claims of legal, medical, forensic, accessibility or regulatory compliance.

Active release-blocking controls include issues #3–#8, #12, #14–#17, #20 and #24, plus issue #4 and draft PR #11. Draft PR #18 does not open a gate. Every P0 and every release-relevant P1 must be independently closed with exact-commit evidence before the release position is reconsidered.

The current `main` Windows build script is not a trustworthy fail-fast release gate because native command exit codes are not explicitly enforced. The correction exists only in blocked draft PR #11. A green Windows job must not be treated as release evidence.

The issue #24 workflow correction stops new public executable uploads but does not validate or approve the built file. Issue #24 remains open until every trigger path is verified and the historical artifact inventory is cleared through deletion or expiry.

## Prohibitions

Official releases and public development workflows must not:

- ask users to disable Smart App Control or antivirus protection;
- silently download components or models;
- claim OCR or AI capability not actually present;
- include real personal evidence in tests or examples;
- overwrite or destructively migrate a workspace without a recoverable checkpoint;
- upload an unsigned executable, DLL, installer, model/runtime payload or runnable archive through public GitHub Actions while release gates are closed;
- publish an unsigned CI output as a normal end-user download;
- rely on a source-level historical SBOM as though it described the final packaged artefact;
- assign publisher, support, complaint, data-role or liability duties to an individual contributor without explicit informed acceptance.
