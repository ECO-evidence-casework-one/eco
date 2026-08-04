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

$replacement = @'
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

$text = $text.Substring(0, $start) + $replacement + "`r`n`r`n" + $text.Substring($end)
[IO.File]::WriteAllText($TargetPath, $text, [Text.UTF8Encoding]::new($false))

$hash = (Get-FileHash -Algorithm SHA256 -Path $TargetPath).Hash.ToLowerInvariant()
Write-Host "Corrected private build script SHA-256: $hash"
