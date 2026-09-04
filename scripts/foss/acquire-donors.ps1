[CmdletBinding()]
param(
    [string]$Root = 'E:\ECO\FOSS_DONORS',
    [switch]$IncludeLarge,
    [switch]$Refresh
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$donors = @(
    [pscustomobject]@{ Name='ethos'; Repo='https://github.com/docushell/ethos.git'; Role='citation-grounding'; Mode='adapter/source-review'; ExpectedLicense='Apache-2.0'; Large=$false },
    [pscustomobject]@{ Name='NetForensicAI'; Repo='https://github.com/Sh3n0bi/NetForensicAI.git'; Role='evidence-timeline-entity-graph'; Mode='pattern-adaptation'; ExpectedLicense='MIT'; Large=$false },
    [pscustomobject]@{ Name='keepr'; Repo='https://github.com/BlinkingSun/keepr.git'; Role='content-addressed-library-ocr-queue'; Mode='pattern-adaptation'; ExpectedLicense='MIT'; Large=$false },
    [pscustomobject]@{ Name='docling'; Repo='https://github.com/docling-project/docling.git'; Role='document-conversion-structure'; Mode='external-engine-adapter'; ExpectedLicense='MIT'; Large=$false },
    [pscustomobject]@{ Name='OCRmyPDF'; Repo='https://github.com/ocrmypdf/OCRmyPDF.git'; Role='pdf-ocr-pipeline'; Mode='external-engine-adapter'; ExpectedLicense='MPL-2.0'; Large=$false },
    [pscustomobject]@{ Name='tesseract'; Repo='https://github.com/tesseract-ocr/tesseract.git'; Role='ocr-engine'; Mode='external-engine-adapter'; ExpectedLicense='Apache-2.0'; Large=$false },
    [pscustomobject]@{ Name='pdfcpu'; Repo='https://github.com/pdfcpu/pdfcpu.git'; Role='pdf-inspection-processing'; Mode='go-library-candidate'; ExpectedLicense='Apache-2.0'; Large=$false },
    [pscustomobject]@{ Name='llama.cpp'; Repo='https://github.com/ggml-org/llama.cpp.git'; Role='local-llm-runtime'; Mode='external-engine-adapter'; ExpectedLicense='MIT'; Large=$false },
    [pscustomobject]@{ Name='restic'; Repo='https://github.com/restic/restic.git'; Role='backup-restore'; Mode='external-engine-adapter'; ExpectedLicense='BSD-2-Clause'; Large=$false },
    [pscustomobject]@{ Name='litestream'; Repo='https://github.com/benbjohnson/litestream.git'; Role='sqlite-recovery-replication'; Mode='external-engine-adapter'; ExpectedLicense='Apache-2.0'; Large=$false },
    [pscustomobject]@{ Name='cosign'; Repo='https://github.com/sigstore/cosign.git'; Role='release-signing-provenance'; Mode='build-release-tool'; ExpectedLicense='Apache-2.0'; Large=$false },
    [pscustomobject]@{ Name='gitleaks'; Repo='https://github.com/gitleaks/gitleaks.git'; Role='secret-scanning'; Mode='ci-development-tool'; ExpectedLicense='MIT'; Large=$false },
    [pscustomobject]@{ Name='velopack'; Repo='https://github.com/velopack/velopack.git'; Role='windows-install-update'; Mode='packaging-study'; ExpectedLicense='MIT'; Large=$false },
    [pscustomobject]@{ Name='gopsutil'; Repo='https://github.com/shirou/gopsutil.git'; Role='local-resource-observability'; Mode='go-library-candidate'; ExpectedLicense='BSD-3-Clause'; Large=$false },
    [pscustomobject]@{ Name='PaddleOCR'; Repo='https://github.com/PaddlePaddle/PaddleOCR.git'; Role='advanced-ocr-layout'; Mode='optional-large-external-engine'; ExpectedLicense='Apache-2.0'; Large=$true }
)

function Invoke-Git([string[]]$Args, [string]$WorkingDirectory = $null) {
    if ($WorkingDirectory) {
        Push-Location $WorkingDirectory
    }
    try {
        $output = & git @Args 2>&1
        if ($LASTEXITCODE -ne 0) {
            throw "git $($Args -join ' ') failed:`n$($output -join [Environment]::NewLine)"
        }
        return $output
    }
    finally {
        if ($WorkingDirectory) { Pop-Location }
    }
}

if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    throw 'Git is not installed or not on PATH. Install Git for Windows first, then rerun this script.'
}

New-Item -ItemType Directory -Force -Path $Root | Out-Null
$sourceRoot = Join-Path $Root 'SOURCE'
$licenseRoot = Join-Path $Root 'LICENSE_SNAPSHOTS'
$reportRoot = Join-Path $Root 'REPORTS'
New-Item -ItemType Directory -Force -Path $sourceRoot,$licenseRoot,$reportRoot | Out-Null

$runId = Get-Date -Format 'yyyyMMdd_HHmmss'
$results = New-Object System.Collections.Generic.List[object]

foreach ($donor in $donors) {
    if ($donor.Large -and -not $IncludeLarge) {
        $results.Add([pscustomobject]@{
            name=$donor.Name; repo=$donor.Repo; role=$donor.Role; mode=$donor.Mode;
            expected_license=$donor.ExpectedLicense; status='SKIPPED_LARGE'; commit=''; branch='';
            license_files=''; license_sha256=''; path=''; error='Use -IncludeLarge to acquire this repository.'
        })
        continue
    }

    $dest = Join-Path $sourceRoot $donor.Name
    Write-Host "`n=== $($donor.Name) ===" -ForegroundColor Cyan
    try {
        if (-not (Test-Path (Join-Path $dest '.git'))) {
            Invoke-Git @('clone','--depth','1',$donor.Repo,$dest) | Out-Host
        }
        elseif ($Refresh) {
            $dirty = (Invoke-Git @('status','--porcelain') $dest) -join ''
            if ($dirty) {
                throw 'Existing checkout is dirty; refusing to refresh it automatically.'
            }
            Invoke-Git @('fetch','--depth','1','origin') $dest | Out-Host
            $defaultRef = (Invoke-Git @('symbolic-ref','refs/remotes/origin/HEAD') $dest | Select-Object -First 1).Trim()
            $defaultBranch = $defaultRef -replace '^refs/remotes/origin/',''
            Invoke-Git @('checkout',$defaultBranch) $dest | Out-Host
            Invoke-Git @('reset','--hard',"origin/$defaultBranch") $dest | Out-Host
        }

        $commit = (Invoke-Git @('rev-parse','HEAD') $dest | Select-Object -First 1).Trim()
        $branch = (Invoke-Git @('branch','--show-current') $dest | Select-Object -First 1).Trim()
        $origin = (Invoke-Git @('remote','get-url','origin') $dest | Select-Object -First 1).Trim()

        $licenseFiles = Get-ChildItem -LiteralPath $dest -File -ErrorAction SilentlyContinue |
            Where-Object { $_.Name -match '^(LICENSE|LICENCE|COPYING|NOTICE)(\..*)?$' }

        $snapDir = Join-Path $licenseRoot $donor.Name
        New-Item -ItemType Directory -Force -Path $snapDir | Out-Null
        $licenseNames = @()
        $licenseHashes = @()
        foreach ($license in $licenseFiles) {
            $target = Join-Path $snapDir $license.Name
            Copy-Item -LiteralPath $license.FullName -Destination $target -Force
            $licenseNames += $license.Name
            $licenseHashes += ((Get-FileHash -Algorithm SHA256 -LiteralPath $license.FullName).Hash.ToLowerInvariant())
        }

        $results.Add([pscustomobject]@{
            name=$donor.Name; repo=$origin; role=$donor.Role; mode=$donor.Mode;
            expected_license=$donor.ExpectedLicense; status='ACQUIRED'; commit=$commit; branch=$branch;
            license_files=($licenseNames -join ';'); license_sha256=($licenseHashes -join ';');
            path=$dest; error=''
        })
    }
    catch {
        Write-Warning "$($donor.Name): $($_.Exception.Message)"
        $results.Add([pscustomobject]@{
            name=$donor.Name; repo=$donor.Repo; role=$donor.Role; mode=$donor.Mode;
            expected_license=$donor.ExpectedLicense; status='FAILED'; commit=''; branch='';
            license_files=''; license_sha256=''; path=$dest; error=$_.Exception.Message
        })
    }
}

$csv = Join-Path $reportRoot "donor_inventory_$runId.csv"
$json = Join-Path $reportRoot "donor_inventory_$runId.json"
$results | Export-Csv -NoTypeInformation -Encoding UTF8 -Path $csv
$results | ConvertTo-Json -Depth 5 | Set-Content -Encoding UTF8 -Path $json

$summary = [pscustomobject]@{
    run_id = $runId
    root = $Root
    acquired = @($results | Where-Object status -eq 'ACQUIRED').Count
    failed = @($results | Where-Object status -eq 'FAILED').Count
    skipped_large = @($results | Where-Object status -eq 'SKIPPED_LARGE').Count
    csv = $csv
    json = $json
}
$summary | ConvertTo-Json -Depth 3 | Set-Content -Encoding UTF8 -Path (Join-Path $reportRoot "summary_$runId.json")

Write-Host "`nAcquisition complete." -ForegroundColor Green
Write-Host "Acquired: $($summary.acquired)  Failed: $($summary.failed)  Skipped large: $($summary.skipped_large)"
Write-Host "Inventory: $csv"

if ($summary.failed -gt 0) { exit 2 }
exit 0
