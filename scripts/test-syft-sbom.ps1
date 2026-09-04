param(
    [Parameter(Mandatory = $true)]
    [string]$SyftPath
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$dist = Join-Path $root "dist"
$exePath = Join-Path $dist "ECO.exe"
$receiptPath = Join-Path $dist "build-receipt.json"
$syftJSONPath = Join-Path $dist "ECO.syft.json"
$spdxJSONPath = Join-Path $dist "ECO.spdx.json"
$sbomReceiptPath = Join-Path $dist "sbom-receipt.json"

foreach ($required in @($SyftPath, $exePath, $receiptPath)) {
    if (-not (Test-Path -LiteralPath $required -PathType Leaf)) {
        throw "Syft SBOM reconciliation requires file: $required"
    }
}

$buildReceipt = Get-Content -LiteralPath $receiptPath -Raw | ConvertFrom-Json
$exeHash = (Get-FileHash -LiteralPath $exePath -Algorithm SHA256).Hash.ToLowerInvariant()
$exeSize = (Get-Item -LiteralPath $exePath).Length
if ([string]$buildReceipt.sha256 -ne $exeHash -or [int64]$buildReceipt.size_bytes -ne [int64]$exeSize) {
    throw "Syft SBOM scan is blocked because build receipt does not match ECO.exe"
}
if ([string]$buildReceipt.deterministic_rebuild -ne "PASS") {
    throw "Syft SBOM scan requires a deterministic rebuild PASS receipt"
}

# Syft's own test harness disables update checks. Keep the scan local-only and
# remove common proxy routes from the child process environment as defense in depth.
$oldUpdateCheck = $env:SYFT_CHECK_FOR_APP_UPDATE
$proxyNames = @("HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "all_proxy", "no_proxy")
$oldProxyValues = @{}
foreach ($name in $proxyNames) {
    $oldProxyValues[$name] = [Environment]::GetEnvironmentVariable($name, "Process")
}

try {
    $env:SYFT_CHECK_FOR_APP_UPDATE = "false"
    foreach ($name in $proxyNames) {
        [Environment]::SetEnvironmentVariable($name, $null, "Process")
    }

    Remove-Item -LiteralPath $syftJSONPath, $spdxJSONPath, $sbomReceiptPath -Force -ErrorAction SilentlyContinue

    & $SyftPath scan "file:dist/ECO.exe" -o "syft-json=$syftJSONPath" -o "spdx-json=$spdxJSONPath"
    if ($LASTEXITCODE -ne 0) {
        throw "Syft scan failed with native exit code $LASTEXITCODE"
    }

    foreach ($output in @($syftJSONPath, $spdxJSONPath)) {
        if (-not (Test-Path -LiteralPath $output -PathType Leaf) -or (Get-Item -LiteralPath $output).Length -le 0) {
            throw "Syft did not create a non-empty SBOM: $output"
        }
    }

    $syft = Get-Content -LiteralPath $syftJSONPath -Raw | ConvertFrom-Json
    if ([string]$syft.source.type -ne "file") {
        throw "Syft JSON source type is not file"
    }
    if ([string]$syft.source.metadata.path -notmatch 'ECO\.exe$') {
        throw "Syft JSON source path does not identify ECO.exe"
    }

    $sourceSHA = $null
    foreach ($digest in @($syft.source.metadata.digests)) {
        if ([string]$digest.algorithm -ieq "sha256") {
            $sourceSHA = ([string]$digest.value).ToLowerInvariant()
            break
        }
    }
    if ([string]::IsNullOrWhiteSpace($sourceSHA) -or $sourceSHA -ne $exeHash) {
        throw "Syft source SHA-256 does not match the deterministic ECO.exe artifact"
    }

    $buildInfo = & go version -m $exePath
    if ($LASTEXITCODE -ne 0) {
        throw "go version -m failed while reconciling compiled dependencies"
    }
    $compiledDeps = @()
    foreach ($line in @($buildInfo)) {
        $text = [string]$line
        if ($text -match '^\s*dep\s+(\S+)\s+(\S+)') {
            $compiledDeps += [pscustomobject]@{ Name = $Matches[1]; Version = $Matches[2] }
        }
    }
    if ($compiledDeps.Count -lt 1) {
        throw "compiled ECO.exe exposed no Go dependency records for SBOM reconciliation"
    }

    $artifacts = @($syft.artifacts)
    if ($artifacts.Count -lt 1) {
        throw "Syft JSON contains no package artifacts"
    }
    $missing = @()
    foreach ($dep in $compiledDeps) {
        $found = $false
        foreach ($artifact in $artifacts) {
            if ([string]$artifact.name -eq [string]$dep.Name -and [string]$artifact.version -eq [string]$dep.Version) {
                $found = $true
                break
            }
        }
        if (-not $found) {
            $missing += "$($dep.Name)@$($dep.Version)"
        }
    }
    if ($missing.Count -gt 0) {
        throw "Syft SBOM is missing compiled Go dependencies: $($missing -join ', ')"
    }

    $spdx = Get-Content -LiteralPath $spdxJSONPath -Raw | ConvertFrom-Json
    if ([string]$spdx.spdxVersion -ne "SPDX-2.3") {
        throw "unexpected SPDX version: $($spdx.spdxVersion)"
    }
    if ([string]$spdx.dataLicense -ne "CC0-1.0") {
        throw "unexpected SPDX data license: $($spdx.dataLicense)"
    }
    if ([string]::IsNullOrWhiteSpace([string]$spdx.documentNamespace)) {
        throw "SPDX document namespace is missing"
    }
    if (@($spdx.packages).Count -lt $compiledDeps.Count) {
        throw "SPDX package count is smaller than the compiled dependency count"
    }

    $syftJSONHash = (Get-FileHash -LiteralPath $syftJSONPath -Algorithm SHA256).Hash.ToLowerInvariant()
    $spdxJSONHash = (Get-FileHash -LiteralPath $spdxJSONPath -Algorithm SHA256).Hash.ToLowerInvariant()
    $sbomReceipt = [ordered]@{
        schema = 1
        status = "PASS"
        generator = "anchore/syft"
        generator_version = "1.51.1"
        source_artifact = "ECO.exe"
        source_sha256 = $exeHash
        source_size_bytes = [int64]$exeSize
        source_commit = [string]$buildReceipt.source_commit
        compiled_go_dependencies = [int]$compiledDeps.Count
        syft_packages = [int]$artifacts.Count
        spdx_packages = [int]@($spdx.packages).Count
        syft_json_sha256 = $syftJSONHash
        spdx_json_sha256 = $spdxJSONHash
    }
    $sbomReceipt | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $sbomReceiptPath -Encoding utf8

    Write-Host "Syft SBOM reconciliation: PASS"
    Write-Host "Artifact SHA-256: $exeHash"
    Write-Host "Compiled Go dependencies reconciled: $($compiledDeps.Count)"
    Write-Host "Syft packages: $($artifacts.Count)"
    Write-Host "SPDX packages: $(@($spdx.packages).Count)"
    Write-Host "Syft JSON SHA-256: $syftJSONHash"
    Write-Host "SPDX JSON SHA-256: $spdxJSONHash"
}
finally {
    $env:SYFT_CHECK_FOR_APP_UPDATE = $oldUpdateCheck
    foreach ($name in $proxyNames) {
        [Environment]::SetEnvironmentVariable($name, $oldProxyValues[$name], "Process")
    }
}
