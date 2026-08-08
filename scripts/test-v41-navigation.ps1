param(
    [Parameter(Mandatory = $true)]
    [string]$ExePath,
    [Parameter(Mandatory = $true)]
    [string]$LocalAppDataRoot
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName Accessibility
$accessibilityAssembly = [Accessibility.IAccessible].Assembly.Location

Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;

public sealed class EcoV41A11yInfo
{
    public IntPtr Hwnd { get; set; }
    public string Name { get; set; }
    public string DefaultAction { get; set; }
    public int Role { get; set; }
    public int State { get; set; }
}

public static class EcoV41A11y
{
    private const uint OBJID_CLIENT = 0xFFFFFFFC;
    private const uint BM_CLICK = 0x00F5;

    [DllImport("user32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    public static extern IntPtr FindWindowEx(IntPtr hwndParent, IntPtr hwndChildAfter, string lpszClass, string lpszWindow);

    [DllImport("user32.dll")]
    [return: MarshalAs(UnmanagedType.Bool)]
    public static extern bool IsWindowVisible(IntPtr hwnd);

    [DllImport("user32.dll", CharSet = CharSet.Unicode)]
    private static extern IntPtr SendMessage(IntPtr hwnd, uint msg, IntPtr wParam, IntPtr lParam);

    [DllImport("oleacc.dll")]
    private static extern int AccessibleObjectFromWindow(
        IntPtr hwnd,
        uint dwId,
        ref Guid riid,
        [MarshalAs(UnmanagedType.Interface)] out Accessibility.IAccessible ppvObject);

    private static Accessibility.IAccessible AccessibleFor(IntPtr hwnd)
    {
        Guid iid = typeof(Accessibility.IAccessible).GUID;
        Accessibility.IAccessible acc;
        int hr = AccessibleObjectFromWindow(hwnd, OBJID_CLIENT, ref iid, out acc);
        if (hr < 0) Marshal.ThrowExceptionForHR(hr);
        if (acc == null) throw new InvalidOperationException("No IAccessible object returned.");
        return acc;
    }

    public static EcoV41A11yInfo Inspect(IntPtr hwnd)
    {
        Accessibility.IAccessible acc = AccessibleFor(hwnd);
        object self = 0;
        EcoV41A11yInfo info = new EcoV41A11yInfo();
        info.Hwnd = hwnd;
        info.Role = Convert.ToInt32(acc.get_accRole(self));
        info.State = Convert.ToInt32(acc.get_accState(self));
        try { info.Name = acc.get_accName(self); } catch { }
        try { info.DefaultAction = acc.get_accDefaultAction(self); } catch { }
        return info;
    }

    public static void Click(IntPtr hwnd)
    {
        SendMessage(hwnd, BM_CLICK, IntPtr.Zero, IntPtr.Zero);
    }
}
'@ -ReferencedAssemblies $accessibilityAssembly

$env:LOCALAPPDATA = $LocalAppDataRoot
$vaultRoot = Join-Path $LocalAppDataRoot 'EvidenceCaseworkOne\V41P1'
New-Item -ItemType Directory -Path $vaultRoot -Force | Out-Null
Set-Content -Path (Join-Path $vaultRoot 'whats-seen-V41-P1') -Value 'CI accessibility navigation test' -Encoding ascii

$p = Start-Process -FilePath $ExePath -PassThru
try {
    $deadline = (Get-Date).AddSeconds(25)
    $hwnd = [IntPtr]::Zero
    while ((Get-Date) -lt $deadline -and -not $p.HasExited) {
        Start-Sleep -Milliseconds 200
        $p.Refresh()
        if ($p.MainWindowHandle -ne 0) {
            $hwnd = New-Object IntPtr ([Int64]$p.MainWindowHandle)
            break
        }
    }
    if ($p.HasExited) { throw "V41 exited before navigation test. Exit code: $($p.ExitCode)" }
    if ($hwnd -eq [IntPtr]::Zero) { throw 'V41 did not expose a main window in time.' }

    $ROLE_SYSTEM_PUSHBUTTON = 0x2B
    $STATE_SYSTEM_FOCUSABLE = 0x00100000
    $labels = @('Home', 'Evidence', 'Ask ECO', 'Matters', 'Review', 'Changes', 'Trust & settings')
    $buttons = @{}

    foreach ($label in $labels) {
        $windowText = if ($label -eq 'Trust & settings') { 'Trust && settings' } else { $label }
        $button = [EcoV41A11y]::FindWindowEx($hwnd, [IntPtr]::Zero, 'Button', $windowText)
        if ($button -eq [IntPtr]::Zero) { throw "Native navigation button '$label' was not found." }
        $info = [EcoV41A11y]::Inspect($button)
        if ($info.Role -ne $ROLE_SYSTEM_PUSHBUTTON) { throw "'$label' role was $($info.Role), expected pushbutton $ROLE_SYSTEM_PUSHBUTTON." }
        if (($info.State -band $STATE_SYSTEM_FOCUSABLE) -eq 0) { throw "'$label' is not focusable through MSAA." }
        if ($info.Name -ne $label) { throw "'$label' accessible name mismatch: '$($info.Name)'." }
        if ([string]::IsNullOrWhiteSpace($info.DefaultAction)) { throw "'$label' has no default accessibility action." }
        $buttons[$label] = $button
    }
    Write-Host 'ECO_GATE v41_nav_msaa=PASS_7_NATIVE_BUTTONS'

    # Hosted Windows runners are not interactive desktops, so use the standard
    # Win32 BM_CLICK contract for functional activation after proving the MSAA
    # role/name/focus/default-action metadata above.
    [EcoV41A11y]::Click($buttons['Ask ECO'])
    Start-Sleep -Milliseconds 350
    $edit = [EcoV41A11y]::FindWindowEx($hwnd, [IntPtr]::Zero, 'Edit', $null)
    if ($edit -eq [IntPtr]::Zero) { throw 'Ask ECO question edit control was not found.' }
    if (-not [EcoV41A11y]::IsWindowVisible($edit)) { throw 'Ask ECO navigation did not expose the question edit control.' }
    Write-Host 'ECO_GATE v41_nav_ask_activation=PASS'

    [EcoV41A11y]::Click($buttons['Evidence'])
    Start-Sleep -Milliseconds 350
    if ([EcoV41A11y]::IsWindowVisible($edit)) { throw 'Evidence navigation did not hide Ask ECO controls.' }
    Write-Host 'ECO_GATE v41_nav_evidence_activation=PASS'

    Write-Host 'ECO_GATE v41_windows_navigation=PASS'
}
finally {
    if ($p -and -not $p.HasExited) {
        Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue
    }
}
