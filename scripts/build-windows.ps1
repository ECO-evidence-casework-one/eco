$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

New-Item -ItemType Directory -Force -Path dist | Out-Null

go test ./...
go vet ./...
go build -trimpath -ldflags "-s -w -H windowsgui -buildid=" -o dist/ECO.exe ./cmd/eco
Get-FileHash dist/ECO.exe -Algorithm SHA256 | Format-List
