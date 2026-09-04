# Cosign private release-envelope rehearsal

Date: 2026-09-04

## Mission fit

This is a GitHub/FOSS-first release-integrity slice. It strengthens ECO's deterministic Windows build receipt without opening the public binary release gate.

ECO's qualified chain now produces:

- two deterministic Windows builds and requires identical SHA-256 values;
- `dist/ECO.exe.sha256`;
- `dist/build-receipt.json` containing build ID, source commit, Go version, target, deterministic-rebuild result, artifact SHA-256 and size;
- reconciled Syft/SPDX SBOMs plus `dist/sbom-receipt.json`;
- an explicit `unsigned private provenance build` release class.

Cosign keeps those controls authoritative and adds a cryptographic signing/verification rehearsal around the complete binding.

## Upstream

- Project: `sigstore/cosign`
- Version: `v3.1.3`
- Published: 2026-08-06
- Licence: Apache-2.0
- Windows amd64 asset: `cosign-windows-amd64.exe`
- Required SHA-256: `9fe59be0eca1271873ce019061335eb1ac419b7059202e797828467ddabe33be`

Version 3.1.3 is intentionally pinned because its release notes state that it resolves a verification-bypass vulnerability affecting an unexpected public key in a legacy bundle.

## CI boundary

The Windows qualification job downloads the exact v3.1.3 executable from the official Sigstore/Cosign GitHub release and checks its SHA-256 before execution. No Cosign binary is committed into ECO or shipped with the application.

After the deterministic build and Syft SBOM reconciliation gates succeed, `scripts/test-cosign-release-envelope.ps1`:

1. verifies that `ECO.exe`, `ECO.exe.sha256` and `build-receipt.json` agree on SHA-256 and size;
2. requires the existing deterministic rebuild result to be `PASS`;
3. requires the underlying artifact to remain explicitly unsigned/private;
4. requires `sbom-receipt.json`, `ECO.syft.json` and `ECO.spdx.json` to exist;
5. requires the SBOM receipt to match the artifact SHA-256/size and the build receipt's source commit;
6. independently re-hashes both SBOM files and requires those hashes to match the SBOM receipt;
7. creates a temporary release envelope binding artifact hash/size, build-receipt hash, checksum-file hash, SBOM-receipt hash/status, both SBOM hashes, build ID and source commit;
8. generates a random throwaway password and local CI-only Cosign key pair in the runner temp directory;
9. signs only the temporary release envelope with `cosign sign-blob` using a local bundle and with transparency-log upload disabled;
10. verifies the authentic envelope with the matching temporary public key;
11. modifies a copy of the envelope and requires verification to fail;
12. deletes the temporary key pair, bundle, envelope and tampered copy before the script exits.

No signing material or signing result is uploaded as a GitHub Actions artifact.

## What this proves

A green gate proves that the current ECO build pipeline can:

- bind the deterministic Windows artifact to its provenance receipt;
- bind reconciled machine-readable SBOMs to the same artifact/source commit;
- create a cryptographic signature over the combined artifact/provenance/SBOM binding;
- verify the authentic binding locally;
- reject modified signed content;
- execute the whole path without a public transparency-log entry or public binary publication.

## What this does NOT prove

This rehearsal does not establish a production publisher identity, Authenticode trust, Smart App Control readiness, a durable release-signing key, a public Sigstore identity, a public release, or formal release approval.

The ephemeral CI public key is not a long-term trust root. A later release-governance decision must define who controls a production identity/key, how recovery and revocation work, what public transparency policy applies, and how the final Windows Authenticode path is operated.

## Remaining release-integrity gaps

- The binary SBOM now covers components discoverable in the built `ECO.exe`; separately installed/local engines still need a controlled runtime-component manifest and packaging reconciliation.
- Public release artifacts remain disabled by the existing P0 release gates.
- Production code signing and publisher identity remain intentionally unresolved.
- Full public-release packaging must reconcile the final executable, binary SBOM, runtime-component manifest, licences/notices, release receipt and production signature before distribution.
