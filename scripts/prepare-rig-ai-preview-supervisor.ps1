# ECO rig AI preview restart supervisor.
# Keeps the qualified v3 build/model recipe unchanged while making failed-run
# recovery robust against disposable Windows files that remain open briefly or
# are retained by an orphaned child process.
[CmdletBinding()]
param(
    [string]$OutputRoot = '',
    [switch]$NoLaunch,
    [switch]$SelfTest,
    [switch]$QualificationOnly
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$ModelName = 'qwen2.5-1.5b-instruct-q4_k_m.gguf'

function Format-Bytes([int64]$Bytes) {
    if ($Bytes -ge 1GB) { return ('{0:N2} GB' -f ($Bytes / 1GB)) }
    if ($Bytes -ge 1MB) { return ('{0:N1} MB' -f ($Bytes / 1MB)) }
    if ($Bytes -ge 1KB) { return ('{0:N1} KB' -f ($Bytes / 1KB)) }
    return "$Bytes bytes"
}

function Write-OutputPointer([string]$RequestedRoot, [string]$ActualRoot) {
    $pointer = $RequestedRoot + '.LAST_OUTPUT.txt'
    [IO.File]::WriteAllText($pointer, $ActualRoot + [Environment]::NewLine, [Text.Encoding]::ASCII)
}

function Move-ResumeFile([string]$Source, [string]$Destination) {
    $last = $null
    for ($attempt = 1; $attempt -le 6; $attempt++) {
        try {
            Move-Item -LiteralPath $Source -Destination $Destination
            return
        } catch {
            $last = $_
            if ($attempt -lt 6) { Start-Sleep -Seconds 2 }
        }
    }
    throw "The previous Qwen partial file is still in use after 12 seconds. ECO did not start a second downloader or discard it. Close any older ECO setup/curl window, wait a few seconds, then run this setup again. File: $Source. Last error: $($last.Exception.Message)"
}

function Seed-FailedRetryRoot([string]$SourceRoot, [string]$TargetRoot) {
    if (Test-Path -LiteralPath $TargetRoot) { throw "Retry output already exists: $TargetRoot" }
    [void][IO.Directory]::CreateDirectory($TargetRoot)
    $assets = Join-Path $TargetRoot 'AIAssets'
    [void][IO.Directory]::CreateDirectory($assets)
    @('ECO RIG AI SETUP STOPPED','Supervisor seed for safe partial-download recovery.') | Set-Content -LiteralPath (Join-Path $TargetRoot 'AI_SETUP_RESULT.txt') -Encoding utf8

    $sourcePartial = Join-Path $SourceRoot ('AIAssets\' + $ModelName + '.part')
    if (Test-Path -LiteralPath $sourcePartial -PathType Leaf) {
        $destinationPartial = Join-Path $assets ($ModelName + '.part')
        $bytes = [int64](Get-Item -LiteralPath $sourcePartial).Length
        Move-ResumeFile $sourcePartial $destinationPartial
        Write-Host "Preserved $(Format-Bytes $bytes) of the existing Qwen partial download for resume."
    }
}

function Resolve-RigOutput([string]$RequestedRoot) {
    if (-not (Test-Path -LiteralPath $RequestedRoot)) { return $RequestedRoot }

    $resultPath = Join-Path $RequestedRoot 'AI_SETUP_RESULT.txt'
    $firstLine = ''
    if (Test-Path -LiteralPath $resultPath -PathType Leaf) {
        $firstLine = [string](Get-Content -LiteralPath $resultPath -TotalCount 1 -ErrorAction SilentlyContinue)
    }
    $launcher = Join-Path $RequestedRoot 'START_ECO_WITH_AI.cmd'
    $completedLauncher = Test-Path -LiteralPath $launcher -PathType Leaf

    if ($firstLine -eq 'ECO RIG AI SETUP PASS' -and $completedLauncher) {
        return $RequestedRoot
    }
    if ($firstLine -ne 'ECO RIG AI SETUP STOPPED' -and ($firstLine -or $completedLauncher)) {
        throw "Output exists and is not a failed/interrupted setup: $RequestedRoot"
    }

    $stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
    $suffix = [guid]::NewGuid().ToString('N').Substring(0,8)
    $archive = $RequestedRoot + '.supervisor-interrupted-' + $stamp + '-' + $suffix
    $archiveError = ''
    $archived = $false
    try {
        Move-Item -LiteralPath $RequestedRoot -Destination $archive
        $archived = $true
    } catch {
        $archiveError = $_.Exception.Message
    }

    if ($archived) {
        Write-Host "Preserved previous interrupted attempt at: $archive"
        Seed-FailedRetryRoot $archive $RequestedRoot
        return $RequestedRoot
    }

    $retryRoot = $RequestedRoot + '.retry-' + $stamp + '-' + $suffix
    Write-Host 'Previous failed preview contains a Windows-locked work file. Leaving it untouched instead of failing again.'
    Write-Host "Locked-folder error: $archiveError"
    Write-Host "Fresh retry location: $retryRoot"
    Seed-FailedRetryRoot $RequestedRoot $retryRoot
    return $retryRoot
}

function Invoke-SupervisorSelfTest {
    $root = Join-Path ([IO.Path]::GetTempPath()) ('eco-rig-restart-selftest-' + [guid]::NewGuid().ToString('N'))
    $requested = Join-Path $root 'ECO_RIG_AI_PREVIEW'
    $retryRoot = ''
    [void][IO.Directory]::CreateDirectory((Join-Path $requested 'AIAssets'))
    [void][IO.Directory]::CreateDirectory((Join-Path $requested '.work'))
    @('ECO RIG AI SETUP STOPPED','synthetic failed setup') | Set-Content -LiteralPath (Join-Path $requested 'AI_SETUP_RESULT.txt') -Encoding utf8
    $partial = Join-Path $requested ('AIAssets\' + $ModelName + '.part')
    [IO.File]::WriteAllBytes($partial, [byte[]](1,2,3,4,5,6,7,8))
    $locked = Join-Path $requested '.work\native-synthetic.stderr.txt'
    $stream = [IO.File]::Open($locked, [IO.FileMode]::Create, [IO.FileAccess]::ReadWrite, [IO.FileShare]::None)
    try {
        $retryRoot = Resolve-RigOutput $requested
        if ($retryRoot -eq $requested) { throw 'Locked-root self-test did not select a fresh sibling retry root.' }
        if (-not (Test-Path -LiteralPath $requested -PathType Container)) { throw 'Locked failed root was unexpectedly moved or deleted.' }
        $carried = Join-Path $retryRoot ('AIAssets\' + $ModelName + '.part')
        if (-not (Test-Path -LiteralPath $carried -PathType Leaf)) { throw 'Qwen partial was not preserved into the retry seed.' }
        if ((Get-Item -LiteralPath $carried).Length -ne 8) { throw 'Qwen partial size changed during restart recovery.' }
        $line = [string](Get-Content -LiteralPath (Join-Path $retryRoot 'AI_SETUP_RESULT.txt') -TotalCount 1)
        if ($line -ne 'ECO RIG AI SETUP STOPPED') { throw 'Retry seed is not recognisable by the qualified v3 preparer.' }
    } finally {
        $stream.Dispose()
        Remove-Item -LiteralPath $root -Recurse -Force -ErrorAction SilentlyContinue
    }
    Write-Host 'Rig AI locked-restart supervisor self-test PASS.'
}

if ($SelfTest) {
    Invoke-SupervisorSelfTest
    exit 0
}
if (-not [Environment]::Is64BitOperatingSystem) { throw '64-bit Windows is required.' }
if (-not $OutputRoot) { $OutputRoot = Join-Path $PSScriptRoot 'ECO_RIG_AI_PREVIEW' }
$requestedRoot = [IO.Path]::GetFullPath($OutputRoot)
Write-OutputPointer $requestedRoot $requestedRoot

$resolvedRoot = Resolve-RigOutput $requestedRoot
Write-OutputPointer $requestedRoot $resolvedRoot

$resultPath = Join-Path $resolvedRoot 'AI_SETUP_RESULT.txt'
$existingLauncher = Join-Path $resolvedRoot 'START_ECO_WITH_AI.cmd'
$existingFirst = ''
if (Test-Path -LiteralPath $resultPath -PathType Leaf) { $existingFirst = [string](Get-Content -LiteralPath $resultPath -TotalCount 1 -ErrorAction SilentlyContinue) }
if ($existingFirst -eq 'ECO RIG AI SETUP PASS' -and (Test-Path -LiteralPath $existingLauncher -PathType Leaf)) {
    Write-Host "Existing completed preview found at: $resolvedRoot"
    if (-not $NoLaunch -and -not $QualificationOnly) { Start-Process -FilePath $existingLauncher -WorkingDirectory $resolvedRoot }
    exit 0
}

$v3 = Join-Path $PSScriptRoot 'prepare-rig-ai-preview-v3.ps1'
if (-not (Test-Path -LiteralPath $v3 -PathType Leaf)) { throw "Qualified v3 preparer is missing: $v3" }
$winPS = Join-Path $env:SystemRoot 'System32\WindowsPowerShell\v1.0\powershell.exe'
if (-not (Test-Path -LiteralPath $winPS -PathType Leaf)) { throw 'Windows PowerShell 5.1 executable not found.' }
$args = @('-NoLogo','-NoProfile','-ExecutionPolicy','Bypass','-File',$v3,'-OutputRoot',$resolvedRoot)
if ($NoLaunch) { $args += '-NoLaunch' }
if ($QualificationOnly) { $args += '-QualificationOnly' }
& $winPS @args
$code = $LASTEXITCODE
exit $code
