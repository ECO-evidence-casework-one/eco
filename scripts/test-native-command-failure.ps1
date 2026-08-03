$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "native-command.ps1")

$caughtExpectedFailure = $false
try {
    Invoke-NativeChecked "controlled native failure" {
        & $env:ComSpec /d /c "exit 23"
    }
} catch {
    if ($_.Exception.Message -notmatch "native exit code 23") {
        throw
    }
    $caughtExpectedFailure = $true
}

if (-not $caughtExpectedFailure) {
    throw "The native-command failure gate did not stop a failing command."
}

Write-Host "Native-command failure gate self-test PASS."
exit 0
