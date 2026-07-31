# Building ECO from source

## Requirements

- Go 1.23.2 or a compatible later Go 1.23 maintenance release
- Windows for direct GUI execution testing

The current source uses the Go standard library and Windows system libraries. It has no third-party Go module dependencies.

## Test

```powershell
go test ./...
go vet ./...
```

## Controlled Windows build

```powershell
./scripts/build-windows.ps1
```

The script:

- runs tests and `go vet`;
- builds the Windows x86-64 GUI twice with deterministic flags;
- rejects the build if the two SHA-256 values differ;
- writes `dist/ECO.exe` as an **unsigned provenance artifact**;
- writes a SHA-256 sidecar and JSON build receipt.

A locally built executable is unsigned and may be blocked by Windows Smart App Control. Do not disable Windows security or describe the artifact as an official release.

## Verification after future trusted signing

```powershell
./scripts/verify-signed-release.ps1 -Path path\to\signed\ECO.exe
```
