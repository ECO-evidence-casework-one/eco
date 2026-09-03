param(
    [Parameter(Mandatory=$true)][string]$AppExe,
    [Parameter(Mandatory=$true)][string]$CoreExe,
    [Parameter(Mandatory=$true)][string]$ExpectedCoreSha256,
    [Parameter(Mandatory=$true)][string]$OutputDir,
    [switch]$StubbornCore,
    [switch]$ExpectHashMismatch
)

$ErrorActionPreference = 'Stop'
New-Item -ItemType Directory -Force $OutputDir | Out-Null
$resultFile = Join-Path $OutputDir 'bridge_result.txt'
$pidFile = Join-Path $OutputDir 'core.pid'
Remove-Item $pidFile -Force -ErrorAction SilentlyContinue

Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes

function TopForProcessId([int]$processId) {
    $desktop = [System.Windows.Automation.AutomationElement]::RootElement
    $cond = New-Object System.Windows.Automation.PropertyCondition([System.Windows.Automation.AutomationElement]::ProcessIdProperty,$processId)
    $desktop.FindFirst([System.Windows.Automation.TreeScope]::Children,$cond)
}

function FindNamedEdit([System.Windows.Automation.AutomationElement]$root, [string]$name) {
    $all = $root.FindAll([System.Windows.Automation.TreeScope]::Descendants,[System.Windows.Automation.Condition]::TrueCondition)
    foreach ($e in $all) {
        if ($e.Current.ControlType -eq [System.Windows.Automation.ControlType]::Edit -and [string]$e.Current.Name -eq $name) {
            return $e
        }
    }
    return $null
}

function ReadValue($element) {
    if ($null -eq $element) { return $null }
    $available = [bool]$element.GetCurrentPropertyValue([System.Windows.Automation.AutomationElement]::IsValuePatternAvailableProperty)
    if (-not $available) { throw "ValuePattern unavailable for $($element.Current.Name)" }
    $pattern = $element.GetCurrentPattern([System.Windows.Automation.ValuePattern]::Pattern)
    return [string]$pattern.Current.Value
}

function WaitForValue([System.Windows.Automation.AutomationElement]$root, [string]$name, [string]$pattern, [int]$seconds = 20) {
    $deadline = (Get-Date).AddSeconds($seconds)
    do {
        $e = FindNamedEdit $root $name
        if ($null -ne $e) {
            $value = ReadValue $e
            if ($value -match $pattern) { return $value }
        }
        Start-Sleep -Milliseconds 200
    } while ((Get-Date) -lt $deadline)
    throw "Timed out waiting for $name to match $pattern"
}

function AssertNoNetworkEndpoint([int]$processId, [string]$label) {
    $tcp = @(Get-NetTCPConnection -OwningProcess $processId -ErrorAction SilentlyContinue)
    $udp = @(Get-NetUDPEndpoint -OwningProcess $processId -ErrorAction SilentlyContinue)
    if ($tcp.Count -gt 0 -or $udp.Count -gt 0) {
        throw "$label unexpectedly owns network endpoints: tcp=$($tcp.Count) udp=$($udp.Count)"
    }
}

$saved = @{
    PATH = $env:PATH
    DOTNET_ROOT = $env:DOTNET_ROOT
    CORE_EXE = $env:ECO_BRIDGE_CORE_EXE
    CORE_SHA = $env:ECO_BRIDGE_CORE_SHA256
    PID_FILE = $env:ECO_BRIDGE_PID_FILE
    IGNORE_EOF = $env:ECO_FIXTURE_IGNORE_EOF
}

$app = $null
$corePid = 0
try {
    $resolvedApp = (Resolve-Path $AppExe).Path
    $resolvedCore = (Resolve-Path $CoreExe).Path
    $env:PATH = "$env:SystemRoot\System32;$env:SystemRoot"
    $env:DOTNET_ROOT = 'Z:\definitely-not-a-dotnet-runtime'
    $env:ECO_BRIDGE_CORE_EXE = $resolvedCore
    $env:ECO_BRIDGE_CORE_SHA256 = $ExpectedCoreSha256
    $env:ECO_BRIDGE_PID_FILE = $pidFile
    $env:ECO_FIXTURE_IGNORE_EOF = $(if ($StubbornCore) { '1' } else { '0' })

    $app = Start-Process -FilePath $resolvedApp -WorkingDirectory (Split-Path $resolvedApp) -PassThru
    $deadline = (Get-Date).AddSeconds(20)
    $root = $null
    do {
        Start-Sleep -Milliseconds 250
        if ($app.HasExited) { throw "Avalonia shell exited before UIA was available: $($app.ExitCode)" }
        $root = TopForProcessId $app.Id
    } while ($null -eq $root -and (Get-Date) -lt $deadline)
    if ($null -eq $root) { throw 'No Avalonia top-level UIA element within 20 seconds.' }

    if ($ExpectHashMismatch) {
        $status = WaitForValue $root 'Core status' '^Bridge failed: Core executable SHA-256 mismatch\.$'
        if (Test-Path $pidFile) { throw 'Core PID file was created despite hash mismatch.' }
        AssertNoNetworkEndpoint $app.Id 'Avalonia shell'
        @(
            'status=PASS',
            'mode=hash_mismatch',
            "core_status=$status",
            'core_started=false',
            'network_endpoints=0'
        ) | Set-Content $resultFile -Encoding utf8
    }
    else {
        $status = WaitForValue $root 'Core status' '^ECO core ready\.$'
        $title = WaitForValue $root 'Matter title' '^Synthetic Bridge Matter$'
        $revision = WaitForValue $root 'Core revision' '^REV-SYNTH-1$'
        $summary = WaitForValue $root 'Evidence summary' '^Records 3; readable 2; unresolved 1$'

        $pidDeadline = (Get-Date).AddSeconds(10)
        while (-not (Test-Path $pidFile) -and (Get-Date) -lt $pidDeadline) { Start-Sleep -Milliseconds 100 }
        if (-not (Test-Path $pidFile)) { throw 'Core PID file was not written.' }
        $corePid = [int](Get-Content $pidFile -Raw).Trim()
        if ($corePid -le 0 -or $null -eq (Get-Process -Id $corePid -ErrorAction SilentlyContinue)) { throw 'Core process is not running after successful projection.' }

        AssertNoNetworkEndpoint $app.Id 'Avalonia shell'
        AssertNoNetworkEndpoint $corePid 'Go core'

        @(
            'status=PASS',
            "mode=$(if ($StubbornCore) { 'stubborn_core' } else { 'normal' })",
            "core_status=$status",
            "matter_title=$title",
            "revision=$revision",
            "evidence_summary=$summary",
            "app_pid=$($app.Id)",
            "core_pid=$corePid",
            'network_endpoints=0'
        ) | Set-Content $resultFile -Encoding utf8
    }

    if (-not $app.CloseMainWindow()) { throw 'Avalonia shell did not accept CloseMainWindow.' }
    if (-not $app.WaitForExit(12000)) { throw 'Avalonia shell did not exit within 12 seconds.' }

    if (-not $ExpectHashMismatch -and $corePid -gt 0) {
        $stopDeadline = (Get-Date).AddSeconds(5)
        do {
            $alive = Get-Process -Id $corePid -ErrorAction SilentlyContinue
            if ($null -eq $alive) { break }
            Start-Sleep -Milliseconds 100
        } while ((Get-Date) -lt $stopDeadline)
        if ($null -ne (Get-Process -Id $corePid -ErrorAction SilentlyContinue)) {
            throw 'Core process remained alive after Avalonia shell exited.'
        }
        Add-Content $resultFile 'core_termination=confirmed'
        if ($StubbornCore) { Add-Content $resultFile 'forced_kill_path=confirmed' }
        else { Add-Content $resultFile 'stdin_eof_shutdown=confirmed' }
    }
}
catch {
    @('status=FAIL',"error=$($_.Exception.Message)") | Set-Content $resultFile -Encoding utf8
    throw
}
finally {
    if ($null -ne $app -and -not $app.HasExited) {
        Stop-Process -Id $app.Id -Force -ErrorAction SilentlyContinue
        try { $app.WaitForExit(3000) | Out-Null } catch { }
    }
    if ($corePid -gt 0 -and $null -ne (Get-Process -Id $corePid -ErrorAction SilentlyContinue)) {
        Stop-Process -Id $corePid -Force -ErrorAction SilentlyContinue
    }
    $env:PATH = $saved.PATH
    $env:DOTNET_ROOT = $saved.DOTNET_ROOT
    $env:ECO_BRIDGE_CORE_EXE = $saved.CORE_EXE
    $env:ECO_BRIDGE_CORE_SHA256 = $saved.CORE_SHA
    $env:ECO_BRIDGE_PID_FILE = $saved.PID_FILE
    $env:ECO_FIXTURE_IGNORE_EOF = $saved.IGNORE_EOF
    if (Test-Path $resultFile) { Get-Content $resultFile | Write-Host }
}
