# Building ECO from source

## Requirements

- Go 1.23.2 or a compatible later Go 1.23 maintenance release
- Windows for direct GUI execution testing

The current `main` source uses the Go standard library and Windows system libraries. It has no third-party Go module dependencies.

## Direct source checks

```powershell
go test ./...
go vet ./...
python scripts/check_source_policy.py
```

## Controlled Windows build development

```powershell
./scripts/build-windows.ps1
```

The current `main` build path:

- loads `scripts/native-command.ps1`;
- captures `$LASTEXITCODE` immediately after every native command;
- terminates on every non-zero native exit;
- runs tests, `go vet` and the source-policy check;
- builds the Windows x86-64 GUI twice with deterministic flags;
- rejects the build if the two SHA-256 values differ;
- writes `dist/ECO.exe` locally as an **unsigned private development/provenance output**;
- writes a SHA-256 sidecar and JSON build receipt.

The CI Windows job also runs:

```powershell
./scripts/test-native-command-failure.ps1
./scripts/test-build-windows-failure-matrix.ps1
```

The first test proves that a controlled non-zero native command terminates the gate. The matrix separately injects failure at test, vet, source policy, first build, second build and `go version`, proves no later stage executes, and asserts that the real native commands in `build-windows.ps1` remain wrapped by `Invoke-NativeChecked`.

Issue #27 was independently closed after:

- reproducing the former false-green behaviour;
- proving a real Windows vet failure stopped before policy and build;
- correcting the three resulting unsafe-pointer findings without suppressing vet;
- passing the full failure matrix and ordinary Linux/Windows validation;
- merging the controls to `main` at `9b7f3d60b14ff67fbf9dc4e0047ceeb498725e79`;
- re-running the merged tree with an empty Actions artifact inventory.

This makes the covered Windows CI commands trustworthy as fail-fast checks. It does **not** approve the resulting executable for testing, distribution or release.

## Public Actions artifact rule

P0 issue #24 established that GitHub Actions artifacts are a public binary-distribution surface, even when labelled unsigned, provenance, temporary or test-only.

At `main` commit `bdc05df444d21d739abf83fa9cf768fc4ab5dd9a`, the workflow stopped uploading `ECO.exe`, its checksum and receipt. The Windows job still compiles and tests the executable on the ephemeral hosted runner, after which the runner workspace is discarded.

Successful and failure-path pull-request runs have since been independently inspected and produced exactly zero Actions artifacts. Issue #24 nevertheless remains open because historical executable artifacts must be deleted or expire, manual-dispatch evidence remains outstanding, and affected development branches must preserve the corrected no-upload workflow when reconciled.

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
