# Release policy

## Release classes

- **Source development snapshot:** source and documentation only.
- **Unsigned private provenance build:** controlled build-chain evidence or an explicitly authorised private test handoff; never a public download while binary gates are closed.
- **Signed development preview:** trusted signature, synthetic/non-sensitive testing only.
- **Release candidate:** signed, migration-tested and independently inspected.
- **Stable release:** all release gates satisfied and no unresolved P0 or release-relevant P1 defects.

A GitHub Actions artifact is a distribution surface. Labelling an executable `unsigned`, `provenance`, `temporary` or `test` does not make public upload permissible.

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
- first-run, reopen, creation, migration, reset and rollback behaviour is explicit, isolated and recoverable;
- ordinary concurrent writers cannot silently erase newer workspace state;
- local AI output is instruction-faithful, source-backed and unable to claim unreceipted actions;
- supplied health, legal and other sensitive evidence can receive permitted source-linked assistance while diagnosis, treatment, clinical-risk, reserved-legal, professional-representation, profiling, scoring and authority-side adverse-decision outputs are blocked;
- import, hashing, extraction, OCR, indexing and model work do not block the interface;
- core workflows pass keyboard, screen-reader, scrolling and supported high-DPI tests;
- document search reports deterministic page-aware results with visible highlighting and source navigation;
- diagnostics are privacy-safe and public offline/network claims match the exact bundled runtime;
- low-resource and long-duration testing passes on the controlling 8 GB Windows baseline.

## Distribution surfaces

The same release gates apply to every channel that can convey runnable ECO code, including:

- GitHub Release assets;
- GitHub Actions artifacts;
- workflow-dispatch outputs;
- installers and update packages;
- public object storage or website downloads;
- partner mirrors;
- archives containing an executable, DLL, model/runtime launcher or another runnable payload.

A controlled private Acer test handoff is not a public release, but it still requires an exact source SHA, executable SHA-256, build receipt, explicit unsigned/development warning, named recipient and purpose, and withdrawal or expiry control.

## Publisher and stewardship gate

An official ordinary-user or institutional release requires an accountable established organisation that has passed:

- [`docs/governance/PUBLISHER_AND_STEWARDSHIP_GATE.md`](docs/governance/PUBLISHER_AND_STEWARDSHIP_GATE.md);
- [`docs/governance/PUBLISHER_ACCEPTANCE_CHECKLIST.md`](docs/governance/PUBLISHER_ACCEPTANCE_CHECKLIST.md).

No organisation is currently appointed.

The preferred route is an existing nonprofit, public-interest institution, open-source foundation or equivalent capable organisation. The originating developer or another contributor is not required to form a company, CIC, charity or other entity or to accept director, trustee, publisher, support, complaints, controller or liability duties merely through contribution.

Contribution alone does not appoint legal roles, but responsibilities may arise from actual publishing, signing, contracting, data processing, supply, claims or applicable law. No individual maintainer may perform official supply or publishing acts on behalf of ECO before authority and role allocation are documented.

Organisational acceptance does not itself approve a release. The exact release candidate must still pass every technical, evidence-integrity, security, accessibility, provenance, licensing, signing and distribution gate.

## Current release-gate status — 4 August 2026

The public repository remains a source-development project. `VERSION` still records the V25 N2 P1 source milestone. Later source commits, draft pull requests and private candidates are not releases.

The following remain blocked:

- ordinary-user binary distribution;
- public executable upload through Actions or another non-Release route;
- use with real, sensitive or irreplaceable evidence;
- release-candidate or stable status;
- claims of reliable generative offline AI, production OCR or complete native PDF investigation;
- public-sector, institutional, healthcare, clinical, justice-sector or EU deployment;
- claims of legal, medical, forensic, accessibility, security, accuracy or regulatory compliance.

Active release-blocking controls include issues #3–#8, #12, #14–#17, #20, #24 and #46, plus issue #4 and draft PR #11. Every P0 and every release-relevant P1 must be closed with exact-commit evidence before the release position is reconsidered.

P0 issue #24 identified four historical unsigned Actions executable packages. Current `main` no longer uploads public runnable artifacts, but issue #24 remains open for historical-artifact expiry or deletion evidence, controlled manual-dispatch evidence, private test-handoff rules and future release-automation controls.

Issue #27 is closed for the exact Windows commands it covers. Current `main` checks native command failures, runs a controlled non-zero self-test and a six-stage failure matrix, and has passed artifact-free validation. This establishes CI truthfulness for those covered commands only; it does not approve the runner-built executable or open a release gate.

The controlling intended-purpose record is merged, but issues #16, #20 and #46 remain open because the application, AI instructions and exports have not yet proved conformance.

Issue #17 remains open. The publisher/stewardship governance record may be adopted without appointing an organisation; the issue closes only after a named organisation completes due diligence, formally accepts the role and passes the required operational tests.

## Prohibitions

Official or public development distribution must not:

- ask users to disable Smart App Control or antivirus protection;
- silently download components or models;
- claim OCR or AI capability not actually present;
- include real personal evidence in tests or examples;
- overwrite or destructively migrate a workspace without a recoverable checkpoint;
- publish an unsigned executable through GitHub Releases, Actions artifacts or another public channel while release gates are closed;
- rely on short retention, warning text or a provenance label to bypass distribution controls;
- rely on a source-level historical SBOM as though it described the final packaged artefact;
- assign publisher, support, complaint, data-role, contracting or liability duties to an individual contributor merely through contribution;
- describe an independent fork or build as an official ECO release without authority;
- announce an official publisher or institutional partner before formal governing-body acceptance and operational validation.
