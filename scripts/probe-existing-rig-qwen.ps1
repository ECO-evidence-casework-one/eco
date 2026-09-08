# Probe an already-downloaded ECO Qwen model without rebuilding or downloading it again.
# Windows PowerShell 5.1 compatible. Synthetic readiness prompt only.
[CmdletBinding()]
param(
    [string]$SearchRoot = 'E:\',
    [int]$TimeoutSeconds = 120
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$ModelName = 'qwen2.5-1.5b-instruct-q4_k_m.gguf'
$ModelSHA = '6a1a2eb6d15622bf3c96857206351ba97e1af16c30d7a74ee38970e434e9407e'

function Get-SHA256([string]$Path) {
    (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}

function Find-QualifiedRigAssets([string]$Root) {
    if (-not (Test-Path -LiteralPath $Root -PathType Container)) {
        throw "Search root does not exist: $Root"
    }
    $candidateRoots = @(Get-ChildItem -LiteralPath $Root -Directory -Filter 'ECO_RIG_AI_PREVIEW*' -ErrorAction SilentlyContinue | Sort-Object LastWriteTime -Descending)
    foreach ($candidate in $candidateRoots) {
        $model = Join-Path $candidate.FullName ('AIAssets\' + $ModelName)
        if (-not (Test-Path -LiteralPath $model -PathType Leaf)) { continue }
        Write-Host "Found complete-looking Qwen file at: $model"
        Write-Host 'Checking the full model SHA-256 before running anything...'
        $sha = Get-SHA256 $model
        if ($sha -cne $ModelSHA) {
            Write-Host "Skipping candidate because SHA-256 is $sha"
            continue
        }
        $llamas = @(Get-ChildItem -LiteralPath $candidate.FullName -Filter 'llama-cli.exe' -File -Recurse -ErrorAction SilentlyContinue)
        if ($llamas.Count -ne 1) {
            Write-Host "Skipping candidate because it has $($llamas.Count) llama-cli.exe files."
            continue
        }
        return [pscustomobject]@{ Root=$candidate.FullName; Model=$model; ModelSHA=$sha; Llama=$llamas[0].FullName }
    }
    throw "No ECO rig preview under $Root contains both the exact approved Qwen model and exactly one llama-cli.exe."
}

function Quote-ProcessArgument([string]$Value) {
    if ($Value -notmatch '[\s"]') { return $Value }
    return '"' + $Value.Replace('"','\"') + '"'
}

function Join-ProcessArguments([string[]]$Values) {
    (($Values | ForEach-Object { Quote-ProcessArgument ([string]$_) }) -join ' ')
}

$assets = Find-QualifiedRigAssets $SearchRoot
Write-Host "Qwen SHA-256 PASS: $($assets.ModelSHA)"

$versionOut = [IO.Path]::GetTempFileName()
$versionErr = [IO.Path]::GetTempFileName()
try {
    $versionProcess = Start-Process -FilePath $assets.Llama -ArgumentList '--offline --version' -NoNewWindow -PassThru -Wait -RedirectStandardOutput $versionOut -RedirectStandardError $versionErr
    $versionText = ((Get-Content -LiteralPath $versionOut -Raw -ErrorAction SilentlyContinue) + "`n" + (Get-Content -LiteralPath $versionErr -Raw -ErrorAction SilentlyContinue)).Trim()
    if ($versionProcess.ExitCode -ne 0 -or $versionText -notmatch '10259' -or $versionText -notmatch '1269cb1ff') {
        throw "Unexpected llama.cpp identity: $versionText"
    }
    Write-Host "llama.cpp identity PASS: $($versionText -split "`r?`n" | Select-Object -First 1)"
} finally {
    Remove-Item -LiteralPath $versionOut,$versionErr -Force -ErrorAction SilentlyContinue
}

$work = Join-Path ([IO.Path]::GetTempPath()) ('eco-qwen-probe-' + [guid]::NewGuid().ToString('N'))
[void][IO.Directory]::CreateDirectory($work)
$prompt = Join-Path $work 'prompt.txt'
$schema = Join-Path $work 'schema.json'
$stdout = Join-Path $work 'stdout.txt'
$stderr = Join-Path $work 'stderr.txt'
[IO.File]::WriteAllText($prompt, 'Return only the JSON object required by the schema. Set ok to ECO_AI_READY.', (New-Object Text.UTF8Encoding($false)))
[IO.File]::WriteAllText($schema, '{"type":"object","additionalProperties":false,"required":["ok"],"properties":{"ok":{"type":"string","enum":["ECO_AI_READY"]}}}', (New-Object Text.UTF8Encoding($false)))

$args = @(
    '--offline',
    '--no-conversation',
    '--model', $assets.Model,
    '--file', $prompt,
    '--json-schema-file', $schema,
    '--simple-io',
    '--no-display-prompt',
    '--color', 'off',
    '--log-disable',
    '--seed', '0',
    '--temp', '0',
    '--top-k', '1',
    '--top-p', '1',
    '--min-p', '0',
    '--ctx-size', '1024',
    '--n-predict', '32',
    '--device', 'none',
    '--n-gpu-layers', '0',
    '--fit', 'off',
    '--no-context-shift',
    '--no-perf'
)

Write-Host ''
Write-Host 'Starting corrected offline Qwen one-shot probe.'
Write-Host "Hard timeout: $TimeoutSeconds seconds. The process will not be allowed to wait indefinitely."
$started = Get-Date
$process = $null
try {
    $process = Start-Process -FilePath $assets.Llama -ArgumentList (Join-ProcessArguments $args) -NoNewWindow -PassThru -RedirectStandardOutput $stdout -RedirectStandardError $stderr
    while (-not $process.HasExited) {
        Start-Sleep -Seconds 5
        $process.Refresh()
        $elapsed = [int]((Get-Date) - $started).TotalSeconds
        Write-Host "Qwen probe running... ${elapsed}s"
        if ($elapsed -ge $TimeoutSeconds) {
            try { $process.Kill() } catch {}
            try { $process.WaitForExit() } catch {}
            $diag = ''
            if (Test-Path -LiteralPath $stderr) { $diag = (Get-Content -LiteralPath $stderr -Raw -ErrorAction SilentlyContinue).Trim() }
            throw "Qwen probe exceeded the $TimeoutSeconds-second hard timeout. Diagnostic: $diag"
        }
    }
    $process.WaitForExit()
    $elapsed = [int]((Get-Date) - $started).TotalSeconds
    $out = ''
    $err = ''
    if (Test-Path -LiteralPath $stdout) { $out = (Get-Content -LiteralPath $stdout -Raw -ErrorAction SilentlyContinue).Trim() }
    if (Test-Path -LiteralPath $stderr) { $err = (Get-Content -LiteralPath $stderr -Raw -ErrorAction SilentlyContinue).Trim() }
    if ($process.ExitCode -ne 0) { throw "Qwen probe exited $($process.ExitCode). Diagnostic: $err" }
    try { $json = $out | ConvertFrom-Json } catch { throw "Qwen returned non-JSON output: $out`nDiagnostic: $err" }
    if ($json.ok -cne 'ECO_AI_READY') { throw "Qwen returned an unexpected result: $out" }

    $receipt = Join-Path $assets.Root 'QWEN_OFFLINE_PROBE_PASS.txt'
    @(
        'ECO QWEN OFFLINE PROBE PASS',
        "Completed: $([DateTimeOffset]::Now.ToString('o'))",
        "Elapsed seconds: $elapsed",
        "Model: $($assets.Model)",
        "Model SHA-256: $($assets.ModelSHA)",
        "Runtime: $($assets.Llama)",
        "Result: $out",
        'Mode: offline, CPU-only, --no-conversation, bounded one-shot probe'
    ) | Set-Content -LiteralPath $receipt -Encoding utf8
    Write-Host ''
    Write-Host 'QWEN OFFLINE GENERATION PASS'
    Write-Host "Elapsed: ${elapsed}s"
    Write-Host "Receipt: $receipt"
} finally {
    if ($process -and -not $process.HasExited) { try { $process.Kill() } catch {} }
    Remove-Item -LiteralPath $work -Recurse -Force -ErrorAction SilentlyContinue
}
