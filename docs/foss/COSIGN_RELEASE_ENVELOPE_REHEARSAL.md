# Cosign private release-envelope rehearsal

Date: 2026-09-04

## Mission fit

This is a GitHub/FOSS-first release-integrity slice. It strengthens ECO's existing deterministic Windows build receipt without opening the public binary release gate.

ECO already produces:

- two deterministic Windows builds and requires identical SHA-256 values;
- `dist/ECO.exe.sha256`;
- `dist/build-receipt.json` containing build ID, source commit, Go version, target, deterministic-rebuild result, artifact SHA-256 and size;
- an explicit `unsigned private provenance build` release class.

This integration keeps those controls authoritative and adds a cryptographic signing/verification rehearsal around them.

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

After ECO's existing deterministic build gate succeeds, `scripts/test-cosign-release-envelope.ps1`:

1. verifies that `ECO.exe`, `ECO.exe.sha256` and `build-receipt.json` agree on SHA-256 and size;
2. requires the existing deterministic rebuild result to be `PASS`;
3. requires the underlying artifact to remain explicitly unsigned/private;
4. creates a temporary release envelope binding the artifact hash/size, build-receipt hash, checksum-file hash, build ID and source commit;
5. generates a random throwaway password and local CI-only Cosign key pair in the runner temp directory;
6. signs only the temporary release envelope with `cosign sign-blob` using a local bundle and with transparency-log upload disabled;
7. verifies the authentic envelope with the matching temporary public key;
8. modifies a copy of the envelope and requires verification to fail;
9. deletes the temporary key pair, bundle, envelope and tampered copy before the script exits.

No signing material or signing result is uploaded as a GitHub Actions artifact.

## What this proves

A green gate proves that the current ECO build pipeline can:

- bind the deterministic Windows artifact to its existing provenance receipt;
- create a cryptographic signature over that binding;
- verify the authentic binding locally;
- reject modified signed content;
- execute the whole path without a public transparency-log entry or public binary publication.

## What this does NOT prove

This rehearsal does not establish a production publisher identity, Authenticode trust, Smart App Control readiness, a durable release-signing key, a public Sigstore identity, a public release, or formal release approval.

The ephemeral CI public key is not a long-term trust root. A later release-governance decision must define who controls a production identity/key, how recovery and revocation work, what public transparency policy applies, and how the final Windows Authenticode path is operated.

## Remaining release-integrity gaps

- ECO still needs a controlled SBOM for its own compiled dependency set and bundled/external runtime components.
- Public release artifacts remain disabled by the existing P0 release gates.
- Production code signing and publisher identity remain intentionally unresolved.
- Full release-manifest/SBOM/signature reconciliation must be qualified before public distribution.
