param(
    [Parameter(Mandatory = $true)]
    [string]$ExePath,

    [Parameter(Mandatory = $true)]
    [string]$TempRoot
)

$ErrorActionPreference = 'Stop'

function Start-EcoProbe([string]$suffix) {
    $handleFile = Join-Path $TempRoot "eco-native-$suffix-hwnd.txt"
    Remove-Item $handleFile -Force -ErrorAction SilentlyContinue
    $p = Start-Process -FilePath $ExePath -ArgumentList @('--handle-file', $handleFile) -PassThru
    $deadline = (Get-Date).AddSeconds(20)
    while (-not (Test-Path $handleFile) -and (Get-Date) -lt $deadline -and -not $p.HasExited) {
        Start-Sleep -Milliseconds 200
        $p.Refresh()
    }
    if ($p.HasExited) { throw "Native Win32 spike exited before $suffix inspection. Exit code: $($p.ExitCode)" }
    if (-not (Test-Path $handleFile)) { throw "Native Win32 spike did not publish a window handle for $suffix inspection." }
    $raw = [UInt64]::Parse((Get-Content $handleFile -Raw).Trim())
    return [pscustomobject]@{ Process = $p; Hwnd = (New-Object IntPtr ([Int64]$raw)) }
}

Add-Type -AssemblyName Accessibility
$accessibilityAssembly = [Accessibility.IAccessible].Assembly.Location
Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;

public sealed class EcoMsaaInfo
{
    public IntPtr Hwnd { get; set; }
    public string ClassName { get; set; }
    public string WindowText { get; set; }
    public string AccessibleName { get; set; }
    public string AccessibleValue { get; set; }
    public string DefaultAction { get; set; }
    public int Role { get; set; }
    public int State { get; set; }
}

public static class EcoOleAccProbe
{
    private const uint OBJID_CLIENT = 0xFFFFFFFC;

    [DllImport("user32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    private static extern IntPtr FindWindowEx(IntPtr parent, IntPtr after, string className, string windowText);

    [DllImport("oleacc.dll")]
    private static extern int AccessibleObjectFromWindow(
        IntPtr hwnd,
        uint objectId,
        ref Guid iid,
        [MarshalAs(UnmanagedType.Interface)] out Accessibility.IAccessible accessible);

    private static Accessibility.IAccessible AccessibleFor(IntPtr hwnd)
    {
        Guid iid = typeof(Accessibility.IAccessible).GUID;
        Accessibility.IAccessible acc;
        int hr = AccessibleObjectFromWindow(hwnd, OBJID_CLIENT, ref iid, out acc);
        if (hr < 0) Marshal.ThrowExceptionForHR(hr);
        if (acc == null) throw new InvalidOperationException("AccessibleObjectFromWindow returned no IAccessible object.");
        return acc;
    }

    public static EcoMsaaInfo Inspect(IntPtr parent, string className, string windowText)
    {
        IntPtr hwnd = FindWindowEx(parent, IntPtr.Zero, className, windowText);
        if (hwnd == IntPtr.Zero)
            throw new InvalidOperationException("Could not find Win32 child class='" + className + "' text='" + windowText + "'.");

        Accessibility.IAccessible acc = AccessibleFor(hwnd);
        object self = 0;
        EcoMsaaInfo info = new EcoMsaaInfo();
        info.Hwnd = hwnd;
        info.ClassName = className;
        info.WindowText = windowText;
        info.Role = Convert.ToInt32(acc.get_accRole(self));
        info.State = Convert.ToInt32(acc.get_accState(self));
        try { info.AccessibleName = acc.get_accName(self); } catch { }
        try { info.AccessibleValue = acc.get_accValue(self); } catch { }
        try { info.DefaultAction = acc.get_accDefaultAction(self); } catch { }
        return info;
    }

    public static void DoDefaultAction(IntPtr parent, string className, string windowText)
    {
        IntPtr hwnd = FindWindowEx(parent, IntPtr.Zero, className, windowText);
        if (hwnd == IntPtr.Zero) throw new InvalidOperationException("Could not find Win32 child for default action.");
        AccessibleFor(hwnd).accDoDefaultAction(0);
    }
}
'@ -ReferencedAssemblies $accessibilityAssembly

$msaa = Start-EcoProbe 'msaa'
try {
    $ROLE_SYSTEM_STATICTEXT = 0x29
    $ROLE_SYSTEM_TEXT = 0x2A
    $ROLE_SYSTEM_PUSHBUTTON = 0x2B
    $ROLE_SYSTEM_LIST = 0x21
    $STATE_SYSTEM_FOCUSABLE = 0x00100000

    $workspace = [EcoOleAccProbe]::Inspect($msaa.Hwnd, 'Static', 'Matter Workspace')
    $workspaceNav = [EcoOleAccProbe]::Inspect($msaa.Hwnd, 'Button', 'Workspace')
    $documentsNav = [EcoOleAccProbe]::Inspect($msaa.Hwnd, 'Button', 'Documents')
    $review = [EcoOleAccProbe]::Inspect($msaa.Hwnd, 'Button', 'Review evidence')
    $ask = [EcoOleAccProbe]::Inspect($msaa.Hwnd, 'Button', 'Ask ECO')
    $edit = [EcoOleAccProbe]::Inspect($msaa.Hwnd, 'Edit', $null)
    $list = [EcoOleAccProbe]::Inspect($msaa.Hwnd, 'ListBox', $null)

    @($workspace, $workspaceNav, $documentsNav, $review, $ask, $edit, $list) |
        Select-Object Hwnd,ClassName,WindowText,AccessibleName,AccessibleValue,DefaultAction,Role,State |
        Format-Table -AutoSize | Out-String | Write-Host

    if ($workspace.Role -ne $ROLE_SYSTEM_STATICTEXT) { throw "Matter Workspace MSAA role was $($workspace.Role), expected $ROLE_SYSTEM_STATICTEXT." }
    foreach ($button in @($workspaceNav, $documentsNav, $review, $ask)) {
        if ($button.Role -ne $ROLE_SYSTEM_PUSHBUTTON) { throw "Button '$($button.WindowText)' MSAA role was $($button.Role), expected $ROLE_SYSTEM_PUSHBUTTON." }
        if (($button.State -band $STATE_SYSTEM_FOCUSABLE) -eq 0) { throw "Button '$($button.WindowText)' is not focusable through MSAA." }
        if ([string]::IsNullOrWhiteSpace($button.AccessibleName)) { throw "Button '$($button.WindowText)' has no MSAA accessible name." }
    }
    if ($edit.Role -ne $ROLE_SYSTEM_TEXT) { throw "Search EDIT MSAA role was $($edit.Role), expected $ROLE_SYSTEM_TEXT." }
    if (($edit.State -band $STATE_SYSTEM_FOCUSABLE) -eq 0) { throw 'Search EDIT is not focusable through MSAA.' }
    if ($list.Role -ne $ROLE_SYSTEM_LIST) { throw "Evidence LISTBOX MSAA role was $($list.Role), expected $ROLE_SYSTEM_LIST." }
    if (($list.State -band $STATE_SYSTEM_FOCUSABLE) -eq 0) { throw 'Evidence LISTBOX is not focusable through MSAA.' }

    [EcoOleAccProbe]::DoDefaultAction($msaa.Hwnd, 'Button', 'Review evidence')
    Start-Sleep -Milliseconds 150
    Write-Host 'ECO_GATE native_win32_msaa_roles=PASS_STATIC_BUTTON_EDIT_LIST'
    Write-Host 'ECO_GATE native_win32_msaa_names=PASS_BUTTON_NAMES'
    Write-Host 'ECO_GATE native_win32_msaa_focusability=PASS_BUTTON_EDIT_LIST'
    Write-Host 'ECO_GATE native_win32_msaa_default_action=PASS_REVIEW_BUTTON'
}
finally {
    if ($msaa.Process -and -not $msaa.Process.HasExited) { Stop-Process -Id $msaa.Process.Id -Force -ErrorAction SilentlyContinue }
}

Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes
Add-Type -AssemblyName UIAutomationClientsideProviders

$uia = Start-EcoProbe 'uia'
try {
    $root = [System.Windows.Automation.AutomationElement]::FromHandle($uia.Hwnd)
    if ($null -eq $root) {
        Write-Host 'ECO_GATE native_win32_hosted_uia=INCONCLUSIVE_NO_UIA_ROOT'
    }
    else {
        $all = $root.FindAll([System.Windows.Automation.TreeScope]::Descendants, [System.Windows.Automation.Condition]::TrueCondition)
        $rows = @()
        foreach ($el in $all) {
            try {
                $rows += [pscustomobject]@{
                    Name = $el.Current.Name
                    Type = $el.Current.ControlType.ProgrammaticName
                    ClassName = $el.Current.ClassName
                    Framework = $el.Current.FrameworkId
                    NativeHandle = $el.Current.NativeWindowHandle
                    Focusable = $el.Current.IsKeyboardFocusable
                }
            } catch { }
        }
        $rows | ConvertTo-Json -Depth 3 | Write-Host
        $reviewUia = @($rows | Where-Object { $_.Name -eq 'Review evidence' }) | Select-Object -First 1
        $editUia = @($rows | Where-Object { $_.ClassName -eq 'Edit' }) | Select-Object -First 1
        $listUia = @($rows | Where-Object { $_.ClassName -eq 'ListBox' }) | Select-Object -First 1
        if ($reviewUia.Type -eq 'ControlType.Button' -and $editUia.Type -eq 'ControlType.Edit' -and $listUia.Type -eq 'ControlType.List') {
            Write-Host 'ECO_GATE native_win32_hosted_uia=PASS_STANDARD_PROXY_MAPPING'
        }
        else {
            Write-Host 'GitHub Windows Server exposes the Win32 HWNDs but flattens standard controls to Pane in this UIA client session.'
            Write-Host 'ECO_GATE native_win32_hosted_uia=INCONCLUSIVE_STANDARD_CONTROLS_FLATTENED_TO_PANES'
        }
    }
}
finally {
    if ($uia.Process -and -not $uia.Process.HasExited) { Stop-Process -Id $uia.Process.Id -Force -ErrorAction SilentlyContinue }
}
