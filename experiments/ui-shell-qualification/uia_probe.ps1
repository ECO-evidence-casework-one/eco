param(
    [Parameter(Mandatory=$true)][string]$ExePath,
    [Parameter(Mandatory=$true)][string]$OutputDir,
    [Parameter(Mandatory=$true)][string]$Framework
)

$ErrorActionPreference = 'Stop'
New-Item -ItemType Directory -Force $OutputDir | Out-Null
$resultFile = Join-Path $OutputDir 'uia_result.txt'
$inventory = Join-Path $OutputDir 'uia_inventory.txt'
$stdout = Join-Path $OutputDir 'app.stdout.txt'
$stderr = Join-Path $OutputDir 'app.stderr.txt'

Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes
Add-Type -AssemblyName System.Windows.Forms

function All([System.Windows.Automation.AutomationElement]$root) {
    $root.FindAll([System.Windows.Automation.TreeScope]::Descendants,[System.Windows.Automation.Condition]::TrueCondition)
}
function NameOf($e) { try { [string]$e.Current.Name } catch { '' } }
function TopForProcessId([int]$processId) {
    $desktop = [System.Windows.Automation.AutomationElement]::RootElement
    $cond = New-Object System.Windows.Automation.PropertyCondition([System.Windows.Automation.AutomationElement]::ProcessIdProperty,$processId)
    $desktop.FindFirst([System.Windows.Automation.TreeScope]::Children,$cond)
}
function Names($elements) {
    $out = New-Object System.Collections.Generic.List[string]
    foreach ($e in $elements) { $n = NameOf $e; if ($n) { $out.Add($n) } }
    $out
}
function FindSearch($elements) {
    foreach ($e in $elements) {
        if ((NameOf $e) -match 'Search this Matter') { return $e }
    }
    foreach ($e in $elements) {
        if ($e.Current.ControlType -eq [System.Windows.Automation.ControlType]::Edit) { return $e }
    }
    $null
}
function FindTranscript($elements) {
    foreach ($e in $elements) {
        if ((NameOf $e) -eq 'AI conversation transcript') { return $e }
    }
    $null
}
function HasValuePattern($e) {
    [bool]$e.GetCurrentPropertyValue([System.Windows.Automation.AutomationElement]::IsValuePatternAvailableProperty)
}

$savedPath = $env:PATH
$p = $null
try {
    $env:PATH = "$env:SystemRoot\System32;$env:SystemRoot"
    $exe = (Resolve-Path $ExePath).Path
    $p = Start-Process -FilePath $exe -WorkingDirectory (Split-Path $exe) -RedirectStandardOutput $stdout -RedirectStandardError $stderr -PassThru

    $deadline = (Get-Date).AddSeconds(20)
    $root = $null
    do {
        Start-Sleep -Milliseconds 250
        if ($p.HasExited) { throw "candidate exited before UIA window; exit=$($p.ExitCode)" }
        $root = TopForProcessId $p.Id
    } while ($null -eq $root -and (Get-Date) -lt $deadline)
    if ($null -eq $root) { throw 'no top-level Windows UI Automation element within 20 seconds' }

    Start-Sleep -Milliseconds 750
    $elements = All $root
    $lines = foreach ($e in $elements) {
        $value = [bool]$e.GetCurrentPropertyValue([System.Windows.Automation.AutomationElement]::IsValuePatternAvailableProperty)
        $text = [bool]$e.GetCurrentPropertyValue([System.Windows.Automation.AutomationElement]::IsTextPatternAvailableProperty)
        '{0}`t{1}`tFocusable={2}`tValuePattern={3}`tTextPattern={4}' -f $e.Current.ControlType.ProgrammaticName,(NameOf $e),$e.Current.IsKeyboardFocusable,$value,$text
    }
    $lines | Set-Content $inventory -Encoding utf8

    if ($elements.Count -lt 8) { throw "tree_too_small:$($elements.Count)" }
    $buttons = @($elements | Where-Object { $_.Current.ControlType -eq [System.Windows.Automation.ControlType]::Button })
    $focusable = @($elements | Where-Object { $_.Current.IsKeyboardFocusable })
    if ($buttons.Count -lt 5) { throw "buttons_too_few:$($buttons.Count)" }
    if ($focusable.Count -lt 5) { throw "focusable_too_few:$($focusable.Count)" }

    $search = FindSearch $elements
    if ($null -eq $search) { throw 'search_control_missing' }
    if (-not (HasValuePattern $search)) { throw 'search_value_pattern_missing' }
    $search.SetFocus()
    $vp = $search.GetCurrentPattern([System.Windows.Automation.ValuePattern]::Pattern)
    foreach ($q in @('w','wa','war')) { $vp.SetValue($q); Start-Sleep -Milliseconds 300 }
    Start-Sleep -Milliseconds 500

    $filtered = Names (All $root)
    if (($filtered | Where-Object { $_ -match 'Warranty confirmation' }).Count -eq 0) { throw 'incremental_search_warranty_missing' }
    if (($filtered | Where-Object { $_ -eq 'Build the Matter timeline' }).Count -gt 0) { throw 'incremental_search_did_not_filter_timeline' }

    $vp.SetValue('')
    Start-Sleep -Milliseconds 600
    $reset = Names (All $root)
    if (($reset | Where-Object { $_ -eq 'Build the Matter timeline' }).Count -eq 0) { throw 'search_clear_did_not_restore_timeline' }

    $transcript = FindTranscript (All $root)
    if ($null -eq $transcript) { throw 'transcript_control_missing' }
    $textAvailable = [bool]$transcript.GetCurrentPropertyValue([System.Windows.Automation.AutomationElement]::IsTextPatternAvailableProperty)
    $valueAvailable = HasValuePattern $transcript
    if (-not $textAvailable -and -not $valueAvailable) { throw 'transcript_has_no_text_or_value_pattern' }

    $transcriptText = ''
    if ($textAvailable) {
        $tp = $transcript.GetCurrentPattern([System.Windows.Automation.TextPattern]::Pattern)
        $transcriptText = [string]$tp.DocumentRange.GetText(4096)
    } elseif ($valueAvailable) {
        $tv = $transcript.GetCurrentPattern([System.Windows.Automation.ValuePattern]::Pattern)
        $transcriptText = [string]$tv.Current.Value
    }
    if ($transcriptText -notmatch 'Known:.*warranty confirmation') { throw 'transcript_accessible_text_missing' }

    @(
        "framework=$Framework",
        'status=PASS',
        "process_id=$($p.Id)",
        "accessible_descendants=$($elements.Count)",
        "buttons=$($buttons.Count)",
        "keyboard_focusable=$($focusable.Count)",
        'search_value_pattern=PASS',
        'live_search_incremental=PASS',
        'search_clear_reset=PASS',
        "transcript_text_pattern=$textAvailable",
        "transcript_value_pattern=$valueAvailable",
        'transcript_accessible=PASS'
    ) | Set-Content $resultFile -Encoding utf8
}
catch {
    @("framework=$Framework",'status=FAIL',"error=$($_.Exception.Message)") | Set-Content $resultFile -Encoding utf8
}
finally {
    if ($null -ne $p -and -not $p.HasExited) { Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue }
    $env:PATH = $savedPath
    Get-Content $resultFile | Write-Host
    if (Test-Path $inventory) { Write-Host '--- UIA inventory ---'; Get-Content $inventory | Write-Host }
    if (Test-Path $stderr) { Write-Host '--- stderr ---'; Get-Content $stderr | Write-Host }
}
