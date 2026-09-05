$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
. (Join-Path $PSScriptRoot 'prepare-private-local-build.ps1') -LibraryOnly
$root = Join-Path $env:TEMP ('eco-preparer-tests-' + [Guid]::NewGuid().ToString('N'))
[void][IO.Directory]::CreateDirectory($root)
Add-Type -AssemblyName System.IO.Compression
Add-Type -AssemblyName System.IO.Compression.FileSystem
$passed = 0
function Expect-Refusal([scriptblock]$Action) {
    $refused = $false
    try { & $Action | Out-Null } catch { $refused = $true }
    if (-not $refused) { throw 'Negative control unexpectedly accepted.' }
}
function Make-Zip([string]$Path, [string[]]$Names) {
    $zip = [IO.Compression.ZipFile]::Open($Path, [IO.Compression.ZipArchiveMode]::Create)
    try {
        foreach ($name in $Names) {
            $e = $zip.CreateEntry($name)
            $s = $e.Open()
            try { $bytes = [Text.Encoding]::UTF8.GetBytes('synthetic'); $s.Write($bytes,0,$bytes.Length) } finally { $s.Dispose() }
        }
    } finally { $zip.Dispose() }
}
try {
    $file = Join-Path $root 'synthetic.txt'
    [IO.File]::WriteAllText($file,'synthetic')
    $hash = (Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash.ToLowerInvariant()
    if ((Assert-EcoHash $file $hash) -cne $hash) { throw 'Correct checksum refused.' }; $passed++
    Expect-Refusal { Assert-EcoHash $file ('0' * 64) }; $passed++
    Expect-Refusal { Assert-EcoHash $file 'not-a-hash' }; $passed++
    foreach ($name in @('../outside.txt','/absolute.txt','C:/outside.txt','folder./file.txt')) {
        $zip = Join-Path $root ('invalid-' + $passed + '.zip'); Make-Zip $zip @($name)
        $out = Join-Path $root ('out-' + $passed)
        Expect-Refusal { Expand-EcoCheckedZip $zip $out }
        if (Test-Path -LiteralPath $out) { throw 'Refused archive created an extraction directory.' }; $passed++
    }
    $zip = Join-Path $root 'duplicate.zip'; Make-Zip $zip @('same.txt','SAME.txt')
    Expect-Refusal { Expand-EcoCheckedZip $zip (Join-Path $root 'duplicate-out') }; $passed++
    $zip = Join-Path $root 'valid.zip'; Make-Zip $zip @('top/ok.txt')
    $out = Join-Path $root 'valid-out'; Expand-EcoCheckedZip $zip $out 'top/'
    if ((Get-Content -LiteralPath (Join-Path $out 'ok.txt') -Raw) -cne 'synthetic') { throw 'Valid archive did not extract correctly.' }; $passed++
    Expect-Refusal { Expand-EcoCheckedZip $zip $out 'top/' }
    if ((Get-Content -LiteralPath (Join-Path $out 'ok.txt') -Raw) -cne 'synthetic') { throw 'Existing destination changed.' }; $passed++
    Expect-Refusal { Invoke-EcoGo $env:ComSpec @('/c','exit','17') }; $passed++
    if ($passed -ne 11) { throw 'Unexpected preparation test count.' }
    Write-Host "Private preparer controls: $passed/11 PASS"
} finally {
    # Only the unique synthetic directory created above is disposable.
    Remove-Item -LiteralPath $root -Recurse -Force
}
