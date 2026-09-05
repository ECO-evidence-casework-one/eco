$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
. (Join-Path $PSScriptRoot "native-command.ps1")

$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

$buildId = (Get-Content VERSION -Raw).Trim()
$commit = $env:GITHUB_SHA
if ([string]::IsNullOrWhiteSpace($commit)) {
    $commit = "local-unrecorded"
}
$dist = Join-Path $root "dist"
New-Item -ItemType Directory -Force -Path $dist | Out-Null

Write-Host "Testing source"
Invoke-NativeChecked "go test ./..." { go test ./... }
Invoke-NativeChecked "go vet ./..." { go vet ./... }
Invoke-NativeChecked "source-policy check" { python scripts/check_source_policy.py }

$first = Join-Path $dist "ECO.exe"
$second = Join-Path $dist "ECO.rebuild.exe"
$sourceCommitFlag = "-X github.com/ECO-evidence-casework-one/eco/internal/eco.SourceCommit=$commit"
$ldflags = "-s -w -H windowsgui -buildid= $sourceCommitFlag"

Write-Host "Building first deterministic Windows artifact"
Invoke-NativeChecked "first deterministic Windows build" { go build -trimpath -ldflags $ldflags -o $first ./cmd/eco }
Write-Host "Building second deterministic Windows artifact"
Invoke-NativeChecked "second deterministic Windows build" { go build -trimpath -ldflags $ldflags -o $second ./cmd/eco }

$hash1 = (Get-FileHash $first -Algorithm SHA256).Hash.ToLowerInvariant()
$hash2 = (Get-FileHash $second -Algorithm SHA256).Hash.ToLowerInvariant()
if ($hash1 -ne $hash2) {
    throw "Deterministic rebuild failed: $hash1 != $hash2"
}
Remove-Item $second -Force

$size = (Get-Item $first).Length
"$hash1  ECO.exe" | Set-Content -Encoding ascii (Join-Path $dist "ECO.exe.sha256")

$goVersion = Invoke-NativeChecked "go version" { go version }
$receipt = [ordered]@{
    schema = 1
    build_id = $buildId
    release_class = "unsigned private provenance build"
    source_commit = $commit
    go_version = $goVersion
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
Write-Host "This build is unsigned, private and not an end-user release."
