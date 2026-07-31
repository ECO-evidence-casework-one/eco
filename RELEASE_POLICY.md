# Release policy

## Release classes

- **Source development snapshot:** source and documentation only.
- **Unsigned prerelease:** automated build for provenance and signing preparation; not recommended for ordinary users.
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

## Prohibitions

Official releases must not:

- ask users to disable Smart App Control or antivirus protection;
- silently download components or models;
- claim OCR or AI capability not actually present;
- include real personal evidence in tests or examples;
- overwrite or destructively migrate a workspace without a recoverable checkpoint.
