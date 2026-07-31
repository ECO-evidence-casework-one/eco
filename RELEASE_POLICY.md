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

## Prohibitions

Official releases must not:

- ask users to disable Smart App Control or antivirus protection;
- silently download components or models;
- claim OCR or AI capability not actually present;
- include real personal evidence in tests or examples;
- overwrite or destructively migrate a workspace without a recoverable checkpoint;
- publish an unsigned CI artifact as a normal end-user download.
