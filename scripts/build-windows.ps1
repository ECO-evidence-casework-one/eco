$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

$buildId = (Get-Content VERSION -Raw).Trim()
$dist = Join-Path $root "dist"
New-Item -ItemType Directory -Force -Path $dist | Out-Null

Write-Host "Testing source"
go test ./...
go vet ./...
python scripts/check_source_policy.py

$first = Join-Path $dist "ECO.exe"
$second = Join-Path $dist "ECO.rebuild.exe"
$ldflags = "-s -w -H windowsgui -buildid="

Write-Host "Building first deterministic Windows artifact"
go build -trimpath -ldflags $ldflags -o $first ./cmd/eco
Write-Host "Building second deterministic Windows artifact"
go build -trimpath -ldflags $ldflags -o $second ./cmd/eco

$hash1 = (Get-FileHash $first -Algorithm SHA256).Hash.ToLowerInvariant()
$hash2 = (Get-FileHash $second -Algorithm SHA256).Hash.ToLowerInvariant()
if ($hash1 -ne $hash2) {
    throw "Deterministic rebuild failed: $hash1 != $hash2"
}
Remove-Item $second -Force

$size = (Get-Item $first).Length
"$hash1  ECO.exe" | Set-Content -Encoding ascii (Join-Path $dist "ECO.exe.sha256")

$commit = $env:GITHUB_SHA
if ([string]::IsNullOrWhiteSpace($commit)) {
    $commit = "local-unrecorded"
}
$receipt = [ordered]@{
    schema = 1
    build_id = $buildId
    release_class = "unsigned provenance artifact"
    source_commit = $commit
    go_version = (go version)
    goos = "windows"
    goarch = "amd64"
    cgo_enabled = "0"
    deterministic_rebuild = "PASS"
    sha256 = $hash1
    size_bytes = $size
    signed = $false
    smart_app_control_ready = $false
}
$receipt | ConvertTo-Json -Depth 4 | Set-Content -Encoding utf8 (Join-Path $dist "build-receipt.json")

Write-Host "Build ID: $buildId"
Write-Host "SHA-256: $hash1"
Write-Host "Size: $size bytes"
Write-Host "This artifact is unsigned and is not an end-user release."
