param(
    [Parameter(Mandatory=$true)][string]$ExePath,
    [Parameter(Mandatory=$true)][string]$OutputDir,
    [string]$Framework = 'unknown'
)

$ErrorActionPreference = 'Stop'
New-Item -ItemType Directory -Force $OutputDir | Out-Null
$stdout = Join-Path $OutputDir 'app.stdout.txt'
$stderr = Join-Path $OutputDir 'app.stderr.txt'
$inventory = Join-Path $OutputDir 'uia_inventory.txt'
$resultFile = Join-Path $OutputDir 'uia_result.txt'

Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes
Add-Type -AssemblyName System.Windows.Forms

function Get-AllElements([System.Windows.Automation.AutomationElement]$root) {
    return $root.FindAll(
        [System.Windows.Automation.TreeScope]::Descendants,
        [System.Windows.Automation.Condition]::TrueCondition
    )
}

function Element-Name($element) {
    try { return [string]$element.Current.Name } catch { return '' }
}

function Find-SearchElement($elements) {
    foreach ($e in $elements) {
        $name = Element-Name $e
        if ($name -match 'Search this Matter|ask ECO') {
            if ($e.GetCurrentPropertyValue([System.Windows.Automation.AutomationElement]::IsValuePatternAvailableProperty)) {
                return $e
            }
        }
    }
    foreach ($e in $elements) {
        if ($e.Current.ControlType -eq [System.Windows.Automation.ControlType]::Edit -and
            $e.GetCurrentPropertyValue([System.Windows.Automation.AutomationElement]::IsValuePatternAvailableProperty)) {
            return $e
        }
    }
    return $null
}

function Names-Snapshot($elements) {
    $names = New-Object System.Collections.Generic.List[string]
    foreach ($e in $elements) {
        $n = Element-Name $e
        if ($n) { $names.Add($n) }
    }
    return $names
}

$savedPath = $env:PATH
$p = $null
try {
    # The consumer probe deliberately removes developer/build paths.
    $env:PATH = "$env:SystemRoot\System32;$env:SystemRoot"
    $exe = (Resolve-Path $ExePath).Path
    $p = Start-Process -FilePath $exe -WorkingDirectory (Split-Path $exe) -RedirectStandardOutput $stdout -RedirectStandardError $stderr -PassThru

    # MainWindowHandle is an IntPtr. In PowerShell, an IntPtr.Zero object is not safely
    # testable with `-not`: the object itself can be truthy even though its numeric handle
    # is zero. Track and compare the numeric handle explicitly so we do not call UIA with
    # IntPtr.Zero before the framework has created its real top-level window.
    $deadline = (Get-Date).AddSeconds(15)
    $proc = $null
    [IntPtr]$hwnd = [IntPtr]::Zero
    do {
        Start-Sleep -Milliseconds 250
        if ($p.HasExited) {
            if (Test-Path $stderr) { Get-Content $stderr | Write-Host }
            throw "UI-0 exited before exposing a window. Exit code: $($p.ExitCode)"
        }
        $proc = Get-Process -Id $p.Id -ErrorAction Stop
        $proc.Refresh()
        [IntPtr]$candidate = $proc.MainWindowHandle
        if ($candidate.ToInt64() -ne 0) {
            $hwnd = $candidate
        }
    } while ($hwnd.ToInt64() -eq 0 -and (Get-Date) -lt $deadline)

    if ($hwnd.ToInt64() -eq 0) { throw 'UI-0 did not expose a main window within 15 seconds.' }

    # Give the framework a short settling interval after HWND creation so its accessible
    # child tree can be populated before UI Automation takes its first snapshot.
    Start-Sleep -Milliseconds 750
    $root = [System.Windows.Automation.AutomationElement]::FromHandle($hwnd)
    if ($null -eq $root) { throw 'UI Automation could not obtain the main window root.' }
    $elements = Get-AllElements $root
    if ($elements.Count -lt 8) { throw "Accessibility tree is unexpectedly small: $($elements.Count) descendants." }

    $lines = New-Object System.Collections.Generic.List[string]
    foreach ($e in $elements) {
        $lines.Add(('{0}`t{1}`tFocusable={2}`tEnabled={3}' -f $e.Current.ControlType.ProgrammaticName,(Element-Name $e),$e.Current.IsKeyboardFocusable,$e.Current.IsEnabled))
    }
    $lines | Set-Content $inventory -Encoding utf8

    $buttons = @($elements | Where-Object { $_.Current.ControlType -eq [System.Windows.Automation.ControlType]::Button })
    $focusable = @($elements | Where-Object { $_.Current.IsKeyboardFocusable })
    if ($buttons.Count -lt 5) { throw "Expected at least 5 accessible buttons, found $($buttons.Count)." }
    if ($focusable.Count -lt 5) { throw "Expected at least 5 keyboard-focusable controls, found $($focusable.Count)." }

    $search = Find-SearchElement $elements
    if ($null -eq $search) { throw 'Could not locate an accessible search text field with ValuePattern.' }
    $search.SetFocus()
    $valuePattern = $search.GetCurrentPattern([System.Windows.Automation.ValuePattern]::Pattern)
    $valuePattern.SetValue('war')
    Start-Sleep -Milliseconds 750

    $filteredElements = Get-AllElements $root
    $filteredNames = Names-Snapshot $filteredElements
    $hasWarranty = ($filteredNames | Where-Object { $_ -match 'warranty|confirmation' }).Count -gt 0
    $hasTimeline = ($filteredNames | Where-Object { $_ -eq 'Build the Matter timeline' }).Count -gt 0
    if (-not $hasWarranty) { throw 'Live search did not expose a warranty/confirmation result after typing war.' }
    if ($hasTimeline) { throw 'Live search did not filter unrelated timeline action after typing war.' }

    $valuePattern.SetValue('')
    Start-Sleep -Milliseconds 750
    $resetNames = Names-Snapshot (Get-AllElements $root)
    if (($resetNames | Where-Object { $_ -eq 'Build the Matter timeline' }).Count -eq 0) {
        throw 'Clearing search did not restore the contextual default actions.'
    }

    # Locate transcript by accessible name first; otherwise locate a text/edit control whose
    # exposed value contains the synthetic answer. This proves the answer exists in the
    # Windows accessibility/text system rather than only as painted pixels.
    $transcript = $null
    foreach ($e in (Get-AllElements $root)) {
        $name = Element-Name $e
        if ($name -eq 'AI conversation transcript') { $transcript = $e; break }
    }
    if ($null -eq $transcript) {
        foreach ($e in (Get-AllElements $root)) {
            if ($e.GetCurrentPropertyValue([System.Windows.Automation.AutomationElement]::IsValuePatternAvailableProperty)) {
                try {
                    $vp = $e.GetCurrentPattern([System.Windows.Automation.ValuePattern]::Pattern)
                    if ([string]$vp.Current.Value -match 'Known:|warranty confirmation') { $transcript = $e; break }
                } catch {}
            }
        }
    }
    if ($null -eq $transcript) { throw 'AI transcript is not exposed as accessible text/value content.' }

    $textAvailable = [bool]$transcript.GetCurrentPropertyValue([System.Windows.Automation.AutomationElement]::IsTextPatternAvailableProperty)
    $valueAvailable = [bool]$transcript.GetCurrentPropertyValue([System.Windows.Automation.AutomationElement]::IsValuePatternAvailableProperty)
    if (-not $textAvailable -and -not $valueAvailable) { throw 'AI transcript exposes neither TextPattern nor ValuePattern.' }

    # Where TextPattern is available, prove that Windows can select the transcript through UIA.
    $selectionProbe = 'value-only'
    if ($textAvailable) {
        $tp = $transcript.GetCurrentPattern([System.Windows.Automation.TextPattern]::Pattern)
        $range = $tp.DocumentRange
        if ($null -eq $range -or [string]::IsNullOrWhiteSpace($range.GetText(4096))) {
            throw 'AI transcript TextPattern exposed no readable text.'
        }
        $range.Select()
        $selectionProbe = 'text-selection-pass'
    }

    @(
        "framework=$Framework",
        'status=PASS',
        "pid=$($p.Id)",
        "hwnd=$($hwnd.ToInt64())",
        "accessible_descendants=$($elements.Count)",
        "buttons=$($buttons.Count)",
        "keyboard_focusable=$($focusable.Count)",
        'live_search=PASS',
        'search_clear_reset=PASS',
        'transcript_accessible=PASS',
        "transcript_selection_probe=$selectionProbe"
    ) | Set-Content $resultFile -Encoding utf8
    Get-Content $resultFile | Write-Host
}
catch {
    @("framework=$Framework",'status=FAIL',"error=$($_.Exception.Message)") | Set-Content $resultFile -Encoding utf8
    Write-Host '--- UIA inventory (if available) ---'
    if (Test-Path $inventory) { Get-Content $inventory | Write-Host }
    Write-Host '--- app stderr ---'
    if (Test-Path $stderr) { Get-Content $stderr | Write-Host }
    throw
}
finally {
    if ($null -ne $p -and -not $p.HasExited) { Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue }
    $env:PATH = $savedPath
}
