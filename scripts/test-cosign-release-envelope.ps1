param(
    [Parameter(Mandatory = $true)]
    [string]$CosignPath
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

$dist = Join-Path $root "dist"
$exePath = Join-Path $dist "ECO.exe"
$checksumPath = Join-Path $dist "ECO.exe.sha256"
$receiptPath = Join-Path $dist "build-receipt.json"

foreach ($required in @($CosignPath, $exePath, $checksumPath, $receiptPath)) {
    if (-not (Test-Path -LiteralPath $required -PathType Leaf)) {
        throw "Cosign release-envelope rehearsal requires file: $required"
    }
}

$receipt = Get-Content -LiteralPath $receiptPath -Raw | ConvertFrom-Json
$exeInfo = Get-Item -LiteralPath $exePath
$exeHash = (Get-FileHash -LiteralPath $exePath -Algorithm SHA256).Hash.ToLowerInvariant()
$checksumLine = (Get-Content -LiteralPath $checksumPath -Raw).Trim()
$expectedChecksumLine = "$exeHash  ECO.exe"

if ([string]$receipt.sha256 -ne $exeHash) {
    throw "build-receipt.json SHA-256 does not match ECO.exe"
}
if ([int64]$receipt.size_bytes -ne [int64]$exeInfo.Length) {
    throw "build-receipt.json size does not match ECO.exe"
}
if ($checksumLine -ne $expectedChecksumLine) {
    throw "ECO.exe.sha256 does not match the actual deterministic artifact"
}
if ([string]$receipt.deterministic_rebuild -ne "PASS") {
    throw "release-envelope rehearsal requires a deterministic rebuild PASS receipt"
}
if ([bool]$receipt.signed) {
    throw "private rehearsal expected the underlying ECO.exe to remain unsigned"
}
if ([string]$receipt.release_class -ne "unsigned private provenance build") {
    throw "unexpected build receipt release class"
}
if ([string]::IsNullOrWhiteSpace([string]$receipt.source_commit)) {
    throw "build receipt is missing source commit provenance"
}

$receiptHash = (Get-FileHash -LiteralPath $receiptPath -Algorithm SHA256).Hash.ToLowerInvariant()
$checksumHash = (Get-FileHash -LiteralPath $checksumPath -Algorithm SHA256).Hash.ToLowerInvariant()

$workRoot = Join-Path $env:RUNNER_TEMP ("eco-cosign-rehearsal-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $workRoot | Out-Null
$passwordBytes = New-Object byte[] 32
[System.Security.Cryptography.RandomNumberGenerator]::Fill($passwordBytes)
$password = [Convert]::ToBase64String($passwordBytes)

$keyPrefix = Join-Path $workRoot "eco-rehearsal"
$privateKey = "$keyPrefix.key"
$publicKey = "$keyPrefix.pub"
$envelopePath = Join-Path $workRoot "release-envelope.json"
$bundlePath = Join-Path $workRoot "release-envelope.sigstore.json"
$tamperedPath = Join-Path $workRoot "release-envelope.tampered.json"

$envelope = [ordered]@{
    schema = 1
    release_class = "private cosign verification rehearsal"
    public_release = $false
    public_trust_identity = $false
    transparency_log_uploaded = $false
    signing_key_class = "ephemeral ci-only test key"
    build_id = [string]$receipt.build_id
    source_commit = [string]$receipt.source_commit
    artifact = [ordered]@{
        name = "ECO.exe"
        sha256 = $exeHash
        size_bytes = [int64]$exeInfo.Length
    }
    build_receipt = [ordered]@{
        name = "build-receipt.json"
        sha256 = $receiptHash
    }
    checksum_file = [ordered]@{
        name = "ECO.exe.sha256"
        sha256 = $checksumHash
    }
}

try {
    $envelope | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $envelopePath -Encoding utf8
    $envelopeHashBefore = (Get-FileHash -LiteralPath $envelopePath -Algorithm SHA256).Hash.ToLowerInvariant()

    $password | & $CosignPath generate-key-pair --output-key-prefix $keyPrefix
    if ($LASTEXITCODE -ne 0) {
        throw "Cosign generate-key-pair failed with exit code $LASTEXITCODE"
    }
    if (-not (Test-Path -LiteralPath $privateKey -PathType Leaf) -or -not (Test-Path -LiteralPath $publicKey -PathType Leaf)) {
        throw "Cosign did not create the expected ephemeral key pair"
    }

    $password | & $CosignPath sign-blob --key $privateKey --bundle $bundlePath --use-signing-config=false --tlog-upload=false $envelopePath
    if ($LASTEXITCODE -ne 0) {
        throw "Cosign sign-blob failed with exit code $LASTEXITCODE"
    }
    if (-not (Test-Path -LiteralPath $bundlePath -PathType Leaf) -or (Get-Item -LiteralPath $bundlePath).Length -le 0) {
        throw "Cosign did not create a non-empty local signature bundle"
    }

    & $CosignPath verify-blob --key $publicKey --bundle $bundlePath --insecure-ignore-tlog=true $envelopePath
    if ($LASTEXITCODE -ne 0) {
        throw "Cosign verify-blob rejected the authentic release envelope"
    }

    $envelopeHashAfter = (Get-FileHash -LiteralPath $envelopePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($envelopeHashAfter -ne $envelopeHashBefore) {
        throw "Cosign signing unexpectedly modified the release envelope"
    }

    Copy-Item -LiteralPath $envelopePath -Destination $tamperedPath
    Add-Content -LiteralPath $tamperedPath -Value " " -Encoding utf8
    & $CosignPath verify-blob --key $publicKey --bundle $bundlePath --insecure-ignore-tlog=true $tamperedPath *> $null
    $tamperedExit = $LASTEXITCODE
    if ($tamperedExit -eq 0) {
        throw "Cosign incorrectly accepted a tampered release envelope"
    }
    # The non-zero native exit above is the expected result being asserted.
    # Clear it after the assertion so PowerShell does not report the successful
    # negative-control test itself as the workflow step's final exit status.
    $global:LASTEXITCODE = 0

    Write-Host "Cosign private release-envelope rehearsal: PASS"
    Write-Host "Artifact SHA-256: $exeHash"
    Write-Host "Envelope SHA-256: $envelopeHashBefore"
    Write-Host "Tampered-envelope rejection: PASS"
    Write-Host "No public release, public signer identity or transparency-log upload was performed."
}
finally {
    if ($null -ne $passwordBytes) {
        [Array]::Clear($passwordBytes, 0, $passwordBytes.Length)
    }
    $password = $null
    if (Test-Path -LiteralPath $workRoot) {
        Remove-Item -LiteralPath $workRoot -Recurse -Force
    }
}
