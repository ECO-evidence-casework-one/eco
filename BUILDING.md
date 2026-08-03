# Building ECO from source

## Requirements

- Go 1.23.2 or a compatible later Go 1.23 maintenance release
- Windows for direct GUI execution testing

The current `main` source uses the Go standard library and Windows system libraries. It has no third-party Go module dependencies.

## Test

```powershell
go test ./...
go vet ./...
```

Run these commands directly and inspect their exit status.

## Controlled Windows build development

```powershell
./scripts/build-windows.ps1
```

The current `main` script:

- runs tests, `go vet` and the source-policy check;
- routes every native test, vet, policy, build and `go version` command through `Invoke-NativeChecked`;
- captures `$LASTEXITCODE` immediately and terminates on every non-zero native exit;
- builds the Windows x86-64 GUI twice with deterministic flags;
- rejects the build if the two SHA-256 values differ;
- writes `dist/ECO.exe` locally as an **unsigned private development/provenance output**;
- writes a SHA-256 sidecar and JSON build receipt.

The Windows workflow also runs:

```powershell
./scripts/test-native-command-failure.ps1
./scripts/test-build-windows-failure-matrix.ps1
```

The first script proves a controlled native exit code stops the gate. The second tests failure at each native stage—tests, vet, source policy, first build, second build and `go version`—and verifies that no later logical stage executes. It also checks that the real build script contains exactly one checked wrapper for each native stage and no legacy unwrapped test, vet, policy or build line.

The fail-fast correction merged in commits `176bf4a51950c42f83456cc45f33e801dd994303` and `9b7f3d60b14ff67fbf9dc4e0047ceeb498725e79`. Final merged-main verification through PR #32, workflow run `30852039542`, passed Linux tests/vet, source policy, the controlled native-failure self-test, the six-stage failure matrix, Windows tests/vet, source policy and deterministic rebuild. The run exposed no downloadable artifact.

A green job is evidence only for the exact source and commands actually reviewed. It is not a release approval, signing result, SBOM, accessibility result, real-evidence approval or assurance that unrelated open P0/P1 findings are resolved.

## Public Actions artifact rule

P0 issue #24 established that GitHub Actions artifacts are a public binary-distribution surface, even when labelled unsigned, provenance, temporary or test-only.

At `main` commit `bdc05df444d21d739abf83fa9cf768fc4ab5dd9a`, the workflow stopped uploading `ECO.exe`, its checksum and receipt. The Windows job still compiles and tests the executable on the ephemeral hosted runner, after which the runner workspace is discarded.

While the public-binary release gate is closed:

- do not add `dist/ECO.exe`, DLLs, installers, model/runtime payloads or runnable archives to `actions/upload-artifact`;
- do not create a manual-dispatch or pull-request route that exposes a runnable payload;
- do not treat short retention, an `unsigned` name or warning text as permission to distribute it;
- do not upload a receipt or report until it has been checked for private paths, usernames, tokens, source text and other sensitive content;
- use a separately authorised private handoff for controlled Acer testing, with exact source SHA, executable SHA-256, build identity, unsigned warning, recipient/purpose and withdrawal or expiry record.

Four historical unsigned executable artifacts identified by issue #24 remain unapproved until deleted or expired. Do not download or run them as test candidates.

A locally built executable is unsigned and may be blocked by Windows Smart App Control. Do not disable Windows security or describe the file as an official release.

## Verification after future trusted signing

```powershell
./scripts/verify-signed-release.ps1 -Path path\to\signed\ECO.exe
```

Trusted signing, an actual-build manifest/SBOM, clean-machine testing, publisher approval and all current release gates remain mandatory before ordinary-user distribution or any public executable artifact.
