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
- write `dist/ECO.exe` as an unsigned local development output;
- write a SHA-256 sidecar and JSON build receipt.

However, PowerShell's `$ErrorActionPreference = "Stop"` does not by itself guarantee failure propagation from native commands. On current `main`, a failed native test, vet, policy or build command can be followed by later steps if the script does not check `$LASTEXITCODE` explicitly. Therefore:

- do not treat the script or a green Windows job on `main` as independent proof that every validation passed;
- inspect raw logs and direct command results;
- do not publish or rely on an output produced after any failed command.

Draft PR #11 contains a fail-fast native-command helper and controlled failure self-test, but that correction is not on `main` and the PR remains blocked for separate P0 reasons.

A locally built executable is unsigned and may be blocked by Windows Smart App Control. Do not disable Windows security or describe the output as an official release.

## Public GitHub Actions policy

Public GitHub Actions is a build-and-test environment, not an approved binary handoff.

While issue #24 and the public-binary gate remain open, the workflow must:

- compile and exercise the Windows executable internally;
- upload no executable, DLL, installer, model/runtime payload or runnable archive;
- discard unsigned build outputs before the hosted job ends;
- preserve only independently approved, privacy-safe non-runnable evidence where required.

The workflow correction does not make the build release-ready. It prevents the public CI service from becoming an unapproved executable distribution channel.

Historical executable artifacts from earlier workflow runs may remain until deletion or expiry. Do not download, run or redistribute them.

## Private controlled test handoff

Any unsigned Acer or user-testing candidate must be transferred outside the public repository through an explicit private handoff that records:

- source commit and build identity;
- exact executable SHA-256;
- unsigned and synthetic-testing warnings;
- intended recipient and purpose;
- expiry, withdrawal or supersession status;
- known limitations and relevant open gates.

A private handoff is not an official release and does not permit real, sensitive or irreplaceable evidence.

## Verification after future trusted signing

```powershell
./scripts/verify-signed-release.ps1 -Path path\to\signed\ECO.exe
```

Trusted signing, an actual-build manifest/SBOM, clean-machine testing, an accountable publisher and all current release gates remain mandatory before ordinary-user distribution.
