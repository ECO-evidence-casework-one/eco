param(
    [Parameter(Mandatory = $true)]
    [string]$TargetPath
)

$ErrorActionPreference = 'Stop'
$text = [IO.File]::ReadAllText($TargetPath)

$oldSourceHash = '7f4909b066a3be91bb3308af3bd81c0c64fe14943ab91194a025516ea4ad9195'
$newSourceHash = '7f4909d8720bc6839ba921e762b81676bc445c18f6b825a29117b9ce7d8b2011'
if ([regex]::Matches($text, [regex]::Escape($oldSourceHash)).Count -ne 1) {
    throw 'Expected exactly one stale source-archive hash in the private build script.'
}
$text = $text.Replace($oldSourceHash, $newSourceHash)

$startMarker = '$ocrExePaths = @('
$endMarker = "Write-Step 'Prove OCR/PDF executables load from packaged runtime'"
$start = $text.IndexOf($startMarker, [StringComparison]::Ordinal)
if ($start -lt 0) { throw 'Hard-coded MSYS2 runtime block start marker was not found.' }
$end = $text.IndexOf($endMarker, $start, [StringComparison]::Ordinal)
if ($end -lt 0) { throw 'OCR/PDF runtime verification marker was not found after the runtime block.' }
if ($text.IndexOf($startMarker, $start + $startMarker.Length, [StringComparison]::Ordinal) -ge 0) {
    throw 'More than one hard-coded MSYS2 runtime block start marker was found.'
}

$runtimeReplacement = @'
$msysRoot = (Resolve-Path (Join-Path (Split-Path -Parent $bash) '..\..')).Path
function Resolve-MsysRuntimeFile([string]$Name) {
  $candidates = Get-ChildItem -Path $msysRoot -Recurse -File -Filter $Name -ErrorAction SilentlyContinue |
    Where-Object { $_.FullName -match '\\(ucrt64|mingw64)\\bin\\' } |
    Sort-Object @{Expression={ if ($_.FullName -match '\\ucrt64\\bin\\') { 0 } else { 1 } }}, FullName
  $match = $candidates | Select-Object -First 1
  if ($null -eq $match) { throw "Missing MSYS2 runtime file: $Name under $msysRoot" }
  return $match.FullName
}

$ocrExePaths = @(
  (Resolve-MsysRuntimeFile 'tesseract.exe'),
  (Resolve-MsysRuntimeFile 'pdftotext.exe'),
  (Resolve-MsysRuntimeFile 'pdftoppm.exe')
)
foreach ($p in $ocrExePaths) { Copy-Item -LiteralPath $p -Destination $runtime -Force }

$unixTargets = foreach ($p in $ocrExePaths) {
  $drive = $p.Substring(0, 1).ToLowerInvariant()
  '/' + $drive + '/' + $p.Substring(3).Replace('\', '/')
}
$quotedTargets = ($unixTargets | ForEach-Object { "'$_'" }) -join ' '
$collectScript = Join-Path $stage 'collect-dlls.sh'
$collectScriptText = @"
set -eu
for f in $quotedTargets; do
  ldd "`$f"
done
"@
[IO.File]::WriteAllText($collectScript, $collectScriptText, [Text.UTF8Encoding]::new($false))
$lddOutput = & $bash $collectScript
if ($LASTEXITCODE -ne 0) { throw 'Dependency enumeration failed' }

$dllUnixPaths = [Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase)
foreach ($line in $lddOutput) {
  $candidate = $null
  if ($line -match '=>\s+(\S+\.dll)(?:\s|$)') { $candidate = $Matches[1] }
  elseif ($line -match '^\s*(\S+\.dll)(?:\s|$)') { $candidate = $Matches[1] }
  if ($candidate -and $candidate -match '/(ucrt64|mingw64)/bin/') {
    [void]$dllUnixPaths.Add($candidate)
  }
}
if ($dllUnixPaths.Count -eq 0) { throw 'No non-system OCR/PDF DLL dependencies were discovered.' }

foreach ($unixPath in $dllUnixPaths) {
  $winPath = (& $bash -lc "cygpath -w '$unixPath'").Trim()
  if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath $winPath)) {
    throw "Enumerated DLL could not be resolved: $unixPath"
  }
  Copy-Item -LiteralPath $winPath -Destination $runtime -Force
}

$tessdataSource = Get-ChildItem -Path $msysRoot -Recurse -File -Filter 'eng.traineddata' -ErrorAction SilentlyContinue |
  Where-Object { $_.FullName -match '\\(ucrt64|mingw64)\\share\\tessdata\\' } |
  Select-Object -First 1
if ($null -eq $tessdataSource) { throw "Missing eng.traineddata under $msysRoot" }
$tessdataDir = Join-Path $runtime 'tessdata'
New-Item -ItemType Directory -Force -Path $tessdataDir | Out-Null
Copy-Item -LiteralPath $tessdataSource.FullName -Destination $tessdataDir -Force

$msysLicenses = Get-ChildItem -Path $msysRoot -Directory -Recurse -Filter 'licenses' -ErrorAction SilentlyContinue |
  Where-Object { $_.FullName -match '\\(ucrt64|mingw64)\\share\\licenses$' } |
  Select-Object -First 1
if ($null -ne $msysLicenses) {
  Copy-Item -LiteralPath $msysLicenses.FullName -Destination (Join-Path $licenseRoot 'MSYS2-UCRT64') -Force -Recurse
}
'@
$text = $text.Substring(0, $start) + $runtimeReplacement + "`r`n`r`n" + $text.Substring($end)

$smokeStartMarker = "Write-Step 'Run genuine local Qwen inference smoke'"
$smokeEndMarker = "Write-Step 'Run ECO AI adapter against the genuine runtime'"
$smokeStart = $text.IndexOf($smokeStartMarker, [StringComparison]::Ordinal)
if ($smokeStart -lt 0) { throw 'Direct llama smoke start marker was not found.' }
$smokeEnd = $text.IndexOf($smokeEndMarker, $smokeStart, [StringComparison]::Ordinal)
if ($smokeEnd -lt 0) { throw 'ECO adapter marker was not found after direct llama smoke.' }
if ($text.IndexOf($smokeStartMarker, $smokeStart + $smokeStartMarker.Length, [StringComparison]::Ordinal) -ge 0) {
    throw 'More than one direct llama smoke start marker was found.'
}

$smokeReplacement = @'
Write-Step 'Run genuine local Qwen inference smoke'
$llamaPath = Join-Path $runtime 'llama-cli.exe'
$smokeOut = Join-Path $logs 'llama-direct-smoke.txt'
$smokeErr = Join-Path $logs 'llama-direct-smoke.stderr.txt'
$llamaVersionOut = Join-Path $logs 'llama-version.txt'
$llamaVersionErr = Join-Path $logs 'llama-version.stderr.txt'

function Invoke-CapturedProcess([string]$FilePath, [string[]]$Arguments, [string]$StdoutPath, [string]$StderrPath) {
  $psi = [Diagnostics.ProcessStartInfo]::new()
  $psi.FileName = $FilePath
  $psi.UseShellExecute = $false
  $psi.CreateNoWindow = $true
  $psi.RedirectStandardOutput = $true
  $psi.RedirectStandardError = $true
  foreach ($argument in $Arguments) { [void]$psi.ArgumentList.Add($argument) }
  $process = [Diagnostics.Process]::new()
  $process.StartInfo = $psi
  if (-not $process.Start()) { throw "Failed to start $FilePath" }
  $stdoutTask = $process.StandardOutput.ReadToEndAsync()
  $stderrTask = $process.StandardError.ReadToEndAsync()
  $process.WaitForExit()
  $stdout = $stdoutTask.GetAwaiter().GetResult()
  $stderr = $stderrTask.GetAwaiter().GetResult()
  [IO.File]::WriteAllText($StdoutPath, $stdout, [Text.UTF8Encoding]::new($false))
  [IO.File]::WriteAllText($StderrPath, $stderr, [Text.UTF8Encoding]::new($false))
  return $process.ExitCode
}

$versionExit = Invoke-CapturedProcess $llamaPath @('--version') $llamaVersionOut $llamaVersionErr
if ($versionExit -ne 0) {
  Write-Host (Get-Content -Raw $llamaVersionOut -ErrorAction SilentlyContinue)
  Write-Host (Get-Content -Raw $llamaVersionErr -ErrorAction SilentlyContinue)
  throw "llama.cpp version probe failed with exit $versionExit"
}

$sw = [Diagnostics.Stopwatch]::StartNew()
$smokeArgs = @(
  '-m', $modelPath,
  '-p', 'Reply with exactly: Hello from ECO local AI.',
  '--single-turn',
  '--no-display-prompt',
  '--simple-io',
  '--no-warmup',
  '--no-show-timings',
  '-t', '4',
  '-c', '512',
  '-n', '32',
  '--temp', '0'
)
$llamaExit = Invoke-CapturedProcess $llamaPath $smokeArgs $smokeOut $smokeErr
$sw.Stop()
if ($llamaExit -ne 0) {
  Write-Host '--- llama stdout ---'
  Write-Host (Get-Content -Raw $smokeOut -ErrorAction SilentlyContinue)
  Write-Host '--- llama stderr ---'
  Write-Host (Get-Content -Raw $smokeErr -ErrorAction SilentlyContinue)
  throw "llama.cpp direct smoke failed with exit $llamaExit"
}
$smokeText = Get-Content -Raw $smokeOut
if ([string]::IsNullOrWhiteSpace($smokeText)) {
  Write-Host (Get-Content -Raw $smokeErr -ErrorAction SilentlyContinue)
  throw 'llama.cpp direct smoke returned no text'
}
'@
$text = $text.Substring(0, $smokeStart) + $smokeReplacement + "`r`n`r`n" + $text.Substring($smokeEnd)

[IO.File]::WriteAllText($TargetPath, $text, [Text.UTF8Encoding]::new($false))
$hash = (Get-FileHash -Algorithm SHA256 -Path $TargetPath).Hash.ToLowerInvariant()
Write-Host "Corrected private build script SHA-256: $hash"
