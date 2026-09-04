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
$rawSyftJSONPath = Join-Path $dist "ECO.syft.raw.json"
$rawSPDXJSONPath = Join-Path $dist "ECO.spdx.raw.json"
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

function Get-TextSHA256([string]$Text) {
    $sha = [System.Security.Cryptography.SHA256]::Create()
    try {
        $bytes = [System.Text.Encoding]::UTF8.GetBytes($Text)
        return [Convert]::ToHexString($sha.ComputeHash($bytes)).ToLowerInvariant()
    }
    finally {
        $sha.Dispose()
    }
}

try {
    $env:SYFT_CHECK_FOR_APP_UPDATE = "false"
    foreach ($name in $proxyNames) {
        [Environment]::SetEnvironmentVariable($name, $null, "Process")
    }

    Remove-Item -LiteralPath $rawSyftJSONPath, $rawSPDXJSONPath, $syftJSONPath, $spdxJSONPath, $sbomReceiptPath -Force -ErrorAction SilentlyContinue

    # Preserve Syft's unmodified binary-only outputs. ECO may create final
    # reconciled copies below when the Go toolchain records a local replace
    # whose version Syft 1.51.1 exposes as UNKNOWN.
    & $SyftPath scan "file:dist/ECO.exe" -o "syft-json=$rawSyftJSONPath" -o "spdx-json=$rawSPDXJSONPath"
    if ($LASTEXITCODE -ne 0) {
        throw "Syft scan failed with native exit code $LASTEXITCODE"
    }

    foreach ($output in @($rawSyftJSONPath, $rawSPDXJSONPath)) {
        if (-not (Test-Path -LiteralPath $output -PathType Leaf) -or (Get-Item -LiteralPath $output).Length -le 0) {
            throw "Syft did not create a non-empty raw SBOM: $output"
        }
    }

    $syft = Get-Content -LiteralPath $rawSyftJSONPath -Raw | ConvertFrom-Json
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

    # Keep local replace metadata associated with the preceding dep record.
    # Example emitted by Go:
    #   dep github.com/example/module v0.0.0-...
    #   =>  ./third_party/example (devel)
    $compiledDeps = @()
    $currentDep = $null
    foreach ($line in @($buildInfo)) {
        $text = [string]$line
        if ($text -match '^\s*dep\s+(\S+)\s+(\S+)') {
            $currentDep = [pscustomobject]@{
                Name = $Matches[1]
                Version = $Matches[2]
                ReplacePath = ""
                ReplaceVersion = ""
            }
            $compiledDeps += $currentDep
            continue
        }
        if ($null -ne $currentDep -and $text -match '^\s*=>\s+(\S+)(?:\s+(\S+))?') {
            $currentDep.ReplacePath = [string]$Matches[1]
            if ($Matches.Count -gt 2) {
                $currentDep.ReplaceVersion = [string]$Matches[2]
            }
            $currentDep = $null
            continue
        }
        if ($text -match '^\s*(path|mod|build)\s+') {
            $currentDep = $null
        }
    }
    if ($compiledDeps.Count -lt 1) {
        throw "compiled ECO.exe exposed no Go dependency records for SBOM reconciliation"
    }

    $artifacts = @($syft.artifacts)
    if ($artifacts.Count -lt 1) {
        throw "Syft JSON contains no package artifacts"
    }

    $spdx = Get-Content -LiteralPath $rawSPDXJSONPath -Raw | ConvertFrom-Json
    if ([string]$spdx.spdxVersion -ne "SPDX-2.3") {
        throw "unexpected SPDX version: $($spdx.spdxVersion)"
    }
    if ([string]$spdx.dataLicense -ne "CC0-1.0") {
        throw "unexpected SPDX data license: $($spdx.dataLicense)"
    }
    if ([string]::IsNullOrWhiteSpace([string]$spdx.documentNamespace)) {
        throw "SPDX document namespace is missing"
    }

    $thirdPartyRoot = [IO.Path]::GetFullPath((Join-Path $root "third_party"))
    $thirdPartyPrefix = $thirdPartyRoot.TrimEnd([IO.Path]::DirectorySeparatorChar, [IO.Path]::AltDirectorySeparatorChar) + [IO.Path]::DirectorySeparatorChar
    $noticePath = Join-Path $root "THIRD_PARTY_NOTICES.md"
    if (-not (Test-Path -LiteralPath $noticePath -PathType Leaf)) {
        throw "local-replace reconciliation requires THIRD_PARTY_NOTICES.md"
    }
    $noticeText = Get-Content -LiteralPath $noticePath -Raw

    $missing = @()
    $localReplaceRecords = @()
    foreach ($dep in $compiledDeps) {
        $exactArtifact = @($artifacts | Where-Object {
            [string]$_.name -eq [string]$dep.Name -and [string]$_.version -eq [string]$dep.Version
        } | Select-Object -First 1)
        if ($exactArtifact.Count -gt 0) {
            continue
        }

        $unknownArtifact = @($artifacts | Where-Object {
            [string]$_.name -eq [string]$dep.Name -and [string]$_.version -eq "UNKNOWN"
        } | Select-Object -First 1)
        $hasLocalReplace = -not [string]::IsNullOrWhiteSpace([string]$dep.ReplacePath) -and [string]$dep.ReplacePath -match '^\.[\\/]'
        if ($unknownArtifact.Count -lt 1 -or -not $hasLocalReplace) {
            $missing += "$($dep.Name)@$($dep.Version)"
            continue
        }

        # Do not accept arbitrary UNKNOWN packages. A local-replace version may
        # be reconciled only from source that is physically inside ECO's
        # controlled third_party tree and has module, licence, provenance and
        # notice records consistent with the compiled module name/version.
        $localPath = [IO.Path]::GetFullPath((Join-Path $root ([string]$dep.ReplacePath)))
        if (-not $localPath.StartsWith($thirdPartyPrefix, [StringComparison]::OrdinalIgnoreCase)) {
            throw "compiled local replace is outside controlled third_party boundary: $($dep.Name) => $($dep.ReplacePath)"
        }
        if (-not (Test-Path -LiteralPath $localPath -PathType Container)) {
            throw "compiled local replace path is missing: $localPath"
        }

        $localGoMod = Join-Path $localPath "go.mod"
        $licensePath = Join-Path $localPath "LICENSE"
        $provenancePath = Join-Path $localPath "ECO_PROVENANCE.md"
        foreach ($required in @($localGoMod, $licensePath, $provenancePath)) {
            if (-not (Test-Path -LiteralPath $required -PathType Leaf)) {
                throw "compiled local replace lacks required reconciliation record: $required"
            }
        }

        $moduleName = $null
        foreach ($modLine in Get-Content -LiteralPath $localGoMod) {
            if ([string]$modLine -match '^\s*module\s+(\S+)') {
                $moduleName = [string]$Matches[1]
                break
            }
        }
        if ([string]::IsNullOrWhiteSpace($moduleName) -or $moduleName -ne [string]$dep.Name) {
            throw "local replace module identity does not match compiled dependency: expected $($dep.Name), found $moduleName"
        }

        $provenanceText = Get-Content -LiteralPath $provenancePath -Raw
        $moduleTail = ([string]$dep.Name) -replace '^github\.com/', ''
        if ($provenanceText -notmatch [regex]::Escape($moduleTail)) {
            throw "local replace provenance does not identify upstream module $($dep.Name)"
        }
        if ([string]$dep.Version -match '-([0-9a-f]{12})$') {
            $commitPrefix = [string]$Matches[1]
            if ($provenanceText -notmatch [regex]::Escape($commitPrefix)) {
                throw "local replace provenance does not bind compiled pseudo-version commit $commitPrefix"
            }
        }
        $licenceMatch = [regex]::Match($provenanceText, 'Upstream licence:\s*([A-Za-z0-9.+-]+)')
        if (-not $licenceMatch.Success) {
            throw "local replace provenance does not declare an upstream SPDX-style licence identifier"
        }
        $declaredLicence = [string]$licenceMatch.Groups[1].Value
        if ($noticeText -notmatch [regex]::Escape([string]$dep.Name)) {
            throw "THIRD_PARTY_NOTICES.md does not identify compiled local replace $($dep.Name)"
        }

        $sourceRows = @()
        foreach ($sourceFile in Get-ChildItem -LiteralPath $localPath -Recurse -File | Sort-Object FullName) {
            $rel = [IO.Path]::GetRelativePath($localPath, $sourceFile.FullName).Replace('\\', '/')
            $hash = (Get-FileHash -LiteralPath $sourceFile.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
            $sourceRows += "$hash  $rel"
        }
        if ($sourceRows.Count -lt 1) {
            throw "local replace source tree is empty: $localPath"
        }
        $treeFingerprint = Get-TextSHA256 (($sourceRows -join "`n") + "`n")
        $licenseHash = (Get-FileHash -LiteralPath $licensePath -Algorithm SHA256).Hash.ToLowerInvariant()
        $provenanceHash = (Get-FileHash -LiteralPath $provenancePath -Algorithm SHA256).Hash.ToLowerInvariant()

        # Syft found this compiled module but cannot derive the version from the
        # local Go replace. Reconcile only that UNKNOWN version/purl using Go's
        # embedded build-info plus the validated controlled source above.
        $artifact = $unknownArtifact[0]
        $artifact.version = [string]$dep.Version
        if ($artifact.PSObject.Properties.Name -contains 'purl') {
            $artifact.purl = "pkg:golang/$($dep.Name)@$($dep.Version)"
        }

        $spdxPackages = @($spdx.packages | Where-Object { [string]$_.name -eq [string]$dep.Name })
        if ($spdxPackages.Count -lt 1) {
            throw "raw SPDX SBOM does not contain Syft's UNKNOWN local-replace package $($dep.Name)"
        }
        $spdxPackage = $spdxPackages[0]
        $spdxPackage.versionInfo = [string]$dep.Version
        $spdxPackage.licenseDeclared = $declaredLicence
        foreach ($externalRef in @($spdxPackage.externalRefs)) {
            if ([string]$externalRef.referenceType -eq 'purl') {
                $externalRef.referenceLocator = "pkg:golang/$($dep.Name)@$($dep.Version)"
            }
        }

        $localReplaceRecords += [pscustomobject]@{
            name = [string]$dep.Name
            compiled_version = [string]$dep.Version
            syft_raw_version = "UNKNOWN"
            replace_path = [string]$dep.ReplacePath
            replace_version = [string]$dep.ReplaceVersion
            declared_licence = $declaredLicence
            source_tree_sha256 = $treeFingerprint
            licence_sha256 = $licenseHash
            provenance_sha256 = $provenanceHash
        }
    }
    if ($missing.Count -gt 0) {
        throw "Syft SBOM is missing compiled Go dependencies: $($missing -join ', ')"
    }

    if ($localReplaceRecords.Count -gt 0) {
        # Keep raw Syft outputs alongside these final reconciled copies. The
        # receipt hashes both sets, making the correction and its exact inputs
        # auditable rather than silently overwriting scanner evidence.
        $syft | ConvertTo-Json -Depth 100 | Set-Content -LiteralPath $syftJSONPath -Encoding utf8
        $spdx | ConvertTo-Json -Depth 100 | Set-Content -LiteralPath $spdxJSONPath -Encoding utf8
    }
    else {
        Copy-Item -LiteralPath $rawSyftJSONPath -Destination $syftJSONPath
        Copy-Item -LiteralPath $rawSPDXJSONPath -Destination $spdxJSONPath
    }

    foreach ($output in @($syftJSONPath, $spdxJSONPath)) {
        if (-not (Test-Path -LiteralPath $output -PathType Leaf) -or (Get-Item -LiteralPath $output).Length -le 0) {
            throw "SBOM reconciliation did not create a non-empty final SBOM: $output"
        }
    }

    # Final copies must now contain exact name/version coverage for every
    # compiled Go dependency, including any strictly validated local replace.
    $finalSyft = Get-Content -LiteralPath $syftJSONPath -Raw | ConvertFrom-Json
    $finalArtifacts = @($finalSyft.artifacts)
    $finalMissing = @()
    foreach ($dep in $compiledDeps) {
        $found = @($finalArtifacts | Where-Object {
            [string]$_.name -eq [string]$dep.Name -and [string]$_.version -eq [string]$dep.Version
        } | Select-Object -First 1)
        if ($found.Count -lt 1) {
            $finalMissing += "$($dep.Name)@$($dep.Version)"
        }
    }
    if ($finalMissing.Count -gt 0) {
        throw "final reconciled Syft SBOM is missing compiled Go dependencies: $($finalMissing -join ', ')"
    }

    $finalSPDX = Get-Content -LiteralPath $spdxJSONPath -Raw | ConvertFrom-Json
    if ([string]$finalSPDX.spdxVersion -ne "SPDX-2.3" -or [string]$finalSPDX.dataLicense -ne "CC0-1.0") {
        throw "final reconciled SPDX metadata is invalid"
    }
    if (@($finalSPDX.packages).Count -lt $compiledDeps.Count) {
        throw "final SPDX package count is smaller than the compiled dependency count"
    }
    $spdxMissing = @()
    foreach ($dep in $compiledDeps) {
        $found = @($finalSPDX.packages | Where-Object {
            [string]$_.name -eq [string]$dep.Name -and [string]$_.versionInfo -eq [string]$dep.Version
        } | Select-Object -First 1)
        if ($found.Count -lt 1) {
            $spdxMissing += "$($dep.Name)@$($dep.Version)"
        }
    }
    if ($spdxMissing.Count -gt 0) {
        throw "final reconciled SPDX SBOM is missing compiled Go dependencies: $($spdxMissing -join ', ')"
    }

    $rawSyftJSONHash = (Get-FileHash -LiteralPath $rawSyftJSONPath -Algorithm SHA256).Hash.ToLowerInvariant()
    $rawSPDXJSONHash = (Get-FileHash -LiteralPath $rawSPDXJSONPath -Algorithm SHA256).Hash.ToLowerInvariant()
    $syftJSONHash = (Get-FileHash -LiteralPath $syftJSONPath -Algorithm SHA256).Hash.ToLowerInvariant()
    $spdxJSONHash = (Get-FileHash -LiteralPath $spdxJSONPath -Algorithm SHA256).Hash.ToLowerInvariant()
    $sbomReceipt = [ordered]@{
        schema = 2
        status = "PASS"
        generator = "anchore/syft + ECO compiled-local-replace reconciliation"
        generator_version = "1.51.1"
        source_artifact = "ECO.exe"
        source_sha256 = $exeHash
        source_size_bytes = [int64]$exeSize
        source_commit = [string]$buildReceipt.source_commit
        compiled_go_dependencies = [int]$compiledDeps.Count
        syft_packages = [int]$finalArtifacts.Count
        spdx_packages = [int]@($finalSPDX.packages).Count
        local_replace_reconciliations = [int]$localReplaceRecords.Count
        local_replace_modules = @($localReplaceRecords)
        raw_syft_json_sha256 = $rawSyftJSONHash
        raw_spdx_json_sha256 = $rawSPDXJSONHash
        syft_json_sha256 = $syftJSONHash
        spdx_json_sha256 = $spdxJSONHash
    }
    $sbomReceipt | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $sbomReceiptPath -Encoding utf8

    Write-Host "Syft SBOM reconciliation: PASS"
    Write-Host "Artifact SHA-256: $exeHash"
    Write-Host "Compiled Go dependencies reconciled: $($compiledDeps.Count)"
    Write-Host "Local replace versions reconciled: $($localReplaceRecords.Count)"
    Write-Host "Syft packages: $($finalArtifacts.Count)"
    Write-Host "SPDX packages: $(@($finalSPDX.packages).Count)"
    Write-Host "Raw Syft JSON SHA-256: $rawSyftJSONHash"
    Write-Host "Raw SPDX JSON SHA-256: $rawSPDXJSONHash"
    Write-Host "Final Syft JSON SHA-256: $syftJSONHash"
    Write-Host "Final SPDX JSON SHA-256: $spdxJSONHash"
}
finally {
    $env:SYFT_CHECK_FOR_APP_UPDATE = $oldUpdateCheck
    foreach ($name in $proxyNames) {
        [Environment]::SetEnvironmentVariable($name, $oldProxyValues[$name], "Process")
    }
}
