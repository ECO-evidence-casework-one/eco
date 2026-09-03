param(
    [Parameter(Mandatory=$true)][string]$AppExe,
    [Parameter(Mandatory=$true)][string]$CoreExe,
    [Parameter(Mandatory=$true)][string]$ExpectedCoreSha256,
    [Parameter(Mandatory=$true)][string]$OutputDir,
    [switch]$StubbornCore,
    [switch]$ExpectHashMismatch
)

$absoluteOutput = [System.IO.Path]::GetFullPath($OutputDir)
New-Item -ItemType Directory -Force $absoluteOutput | Out-Null

$params = @{
    AppExe = $AppExe
    CoreExe = $CoreExe
    ExpectedCoreSha256 = $ExpectedCoreSha256
    OutputDir = $absoluteOutput
}
if ($StubbornCore) { $params.StubbornCore = $true }
if ($ExpectHashMismatch) { $params.ExpectHashMismatch = $true }

& (Join-Path $PSScriptRoot 'bridge_probe.ps1') @params
