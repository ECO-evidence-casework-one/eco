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

Run these commands directly and inspect their exit status. The current `main` PowerShell build script does not reliably stop on every non-zero native command exit code.

## Controlled Windows build development

```powershell
./scripts/build-windows.ps1
```

The current `main` script intends to:

- run tests, `go vet` and the source-policy check;
- build the Windows x86-64 GUI twice with deterministic flags;
- reject the build if the two SHA-256 values differ;
- write `dist/ECO.exe` locally as an **unsigned development/provenance output**;
- write a SHA-256 sidecar and JSON build receipt.

However, PowerShell's `$ErrorActionPreference = "Stop"` does not by itself guarantee failure propagation from native commands. On current `main`, a failed native test, vet, policy or build command can be followed by later steps if the script does not check `$LASTEXITCODE` explicitly. Therefore:

- do not treat the script or a green Windows job on `main` as independent proof that every validation passed;
- inspect raw logs and direct command results;
- do not publish, use or rely on an executable produced after any failed command.

Draft PR #11 contains a fail-fast native-command helper and controlled failure self-test, but that correction is not on `main` and the PR remains blocked for separate P0 reasons.

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
