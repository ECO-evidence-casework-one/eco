# Syft SBOM reconciliation gate

Date: 2026-09-04

## Mission fit

This is a GitHub/FOSS-first release-integrity integration. It inventories the software components actually present in the deterministic Windows `ECO.exe` and binds that inventory back to ECO's existing build provenance and Cosign release-envelope rehearsal.

## Upstream

- Project: `anchore/syft`
- Version: `v1.51.1`
- Published: 2026-08-27
- Licence: Apache-2.0
- Windows amd64 asset: `syft_1.51.1_windows_amd64.zip`
- Required SHA-256: `5e4bc3e6b6344b4625de0f7aa5351aaa72856d11d78462972de0a101ee2c1c8f`

The CI gate downloads only this exact release asset from the official Anchore/Syft GitHub release and verifies the ZIP SHA-256 before extraction/execution. No Syft binary is committed into ECO or shipped with the application.

## Local-only scan boundary

After the deterministic Windows build succeeds, CI runs Syft against `file:dist/ECO.exe`.

For the scan process:

- `SYFT_CHECK_FOR_APP_UPDATE=false` is set, matching Syft's own test practice for avoiding update checks;
- common HTTP/HTTPS/all-proxy environment variables are removed from the child-process environment;
- the source is the already-built local file, not a registry/image/network target.

The network is used only before the scan to fetch the pinned, checksum-verified Syft tool.

## Outputs

The gate creates three ephemeral files under `dist/`:

- `ECO.syft.json` — Syft's native machine-readable SBOM;
- `ECO.spdx.json` — SPDX 2.3 JSON;
- `sbom-receipt.json` — ECO reconciliation receipt.

These files remain on the ephemeral CI runner under the existing closed public-artifact gate.

## Reconciliation controls

`scripts/test-syft-sbom.ps1` requires:

1. `build-receipt.json` SHA-256 and size to match the actual deterministic `ECO.exe`;
2. the build receipt's deterministic-rebuild result to be `PASS`;
3. Syft's native JSON source type to be `file` and its source path to identify `ECO.exe`;
4. Syft's source SHA-256 digest to equal the independently calculated `ECO.exe` SHA-256;
5. `go version -m ECO.exe` to expose at least one compiled Go dependency;
6. every compiled Go dependency path+version reported by the executable to appear in Syft's package inventory;
7. the SPDX document to be SPDX 2.3, use `CC0-1.0` for the SPDX data licence, and have a non-empty document namespace;
8. the SPDX package count not to be smaller than the actual compiled Go dependency count.

The reconciliation receipt records the source artifact SHA-256/size/source commit, compiled dependency count, package counts, Syft version and SHA-256 values for both generated SBOM files.

## Cosign binding

The Cosign private release-envelope rehearsal now refuses to run unless `sbom-receipt.json`, `ECO.syft.json`, and `ECO.spdx.json` exist and agree with the deterministic artifact/build receipt.

The signed temporary envelope includes:

- the deterministic `ECO.exe` SHA-256 and size;
- build-receipt hash;
- checksum-file hash;
- SBOM-receipt hash and qualification status;
- Syft JSON hash;
- SPDX JSON hash;
- build ID and source commit.

This means modifying an SBOM after reconciliation changes the signed envelope inputs and is detected before/at signature verification.

## Deliberate boundary: external local engines

This binary SBOM describes components Syft can discover in the built `ECO.exe`. It does not claim that external optional/local tools are compiled into the executable.

Examples currently connected by ECO adapters but not embedded by this gate include Tesseract, Docling, OCRmyPDF, llama.cpp, pdfcpu and other separately installed/local engines. Their exact installed versions/hashes/licences require a separate runtime-component manifest and later packaging qualification.

## What this does NOT prove

A green SBOM gate does not prove that all components are vulnerability-free, appropriately licensed for every future distribution mode, or present on an end user's machine. It also does not open the public release gate.

Production release still requires controlled runtime-component packaging/inventory, publisher identity/code signing decisions, and the remaining release/adversarial qualification gates.
