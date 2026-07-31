# Building ECO from source

## Requirements

- Go 1.23.2 or a compatible later Go 1.23 maintenance release
- Windows for direct GUI execution testing

The current code uses the Go standard library and Windows system libraries. It has no third-party Go module dependencies.

## Test

```powershell
go test ./...
go vet ./...
```

## Build the Windows GUI

```powershell
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -trimpath -ldflags "-s -w -H windowsgui -buildid=" -o dist/ECO.exe ./cmd/eco
```

The repository workflow performs the same controlled build. A locally built executable is unsigned and may be blocked by Windows Smart App Control.

## Verification

```powershell
Get-FileHash dist/ECO.exe -Algorithm SHA256
Get-AuthenticodeSignature dist/ECO.exe
```

Do not describe a locally built unsigned executable as an official release.
