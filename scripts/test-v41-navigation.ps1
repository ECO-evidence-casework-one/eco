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

public static class EcoV41A11y
{
    private const uint OBJID_CLIENT = 0xFFFFFFFC;

    [DllImport("user32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
    public static extern IntPtr FindWindowEx(IntPtr hwndParent, IntPtr hwndChildAfter, string lpszClass, string lpszWindow);

    [DllImport("user32.dll")]
    [return: MarshalAs(UnmanagedType.Bool)]
    public static extern bool IsWindowVisible(IntPtr hwnd);

    [DllImport("oleacc.dll")]
    private static extern int AccessibleObjectFromWindow(
        IntPtr hwnd,
        uint dwId,
        ref Guid riid,
        [MarshalAs(UnmanagedType.Interface)] out Accessibility.IAccessible ppvObject);

    public static Accessibility.IAccessible AccessibleFor(IntPtr hwnd)
    {
        Guid iid = typeof(Accessibility.IAccessible).GUID;
        Accessibility.IAccessible acc;
        int hr = AccessibleObjectFromWindow(hwnd, OBJID_CLIENT, ref iid, out acc);
        if (hr < 0) Marshal.ThrowExceptionForHR(hr);
        if (acc == null) throw new InvalidOperationException("No IAccessible object returned.");
        return acc;
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
        $button = [EcoV41A11y]::FindWindowEx($hwnd, [IntPtr]::Zero, 'Button', $label)
        if ($button -eq [IntPtr]::Zero) { throw "Native navigation button '$label' was not found." }
        $acc = [EcoV41A11y]::AccessibleFor($button)
        $self = 0
        $role = [int]$acc.get_accRole($self)
        $state = [int]$acc.get_accState($self)
        $name = $acc.get_accName($self)
        $action = $acc.get_accDefaultAction($self)
        if ($role -ne $ROLE_SYSTEM_PUSHBUTTON) { throw "'$label' role was $role, expected pushbutton $ROLE_SYSTEM_PUSHBUTTON." }
        if (($state -band $STATE_SYSTEM_FOCUSABLE) -eq 0) { throw "'$label' is not focusable through MSAA." }
        if ($name -ne $label) { throw "'$label' accessible name mismatch: '$name'." }
        if ([string]::IsNullOrWhiteSpace($action)) { throw "'$label' has no default accessibility action." }
        $buttons[$label] = $acc
    }
    Write-Host 'ECO_GATE v41_nav_msaa=PASS_7_NATIVE_BUTTONS'

    $buttons['Ask ECO'].accDoDefaultAction(0)
    Start-Sleep -Milliseconds 350
    $edit = [EcoV41A11y]::FindWindowEx($hwnd, [IntPtr]::Zero, 'Edit', $null)
    if ($edit -eq [IntPtr]::Zero) { throw 'Ask ECO question edit control was not found.' }
    if (-not [EcoV41A11y]::IsWindowVisible($edit)) { throw 'Ask ECO navigation did not expose the question edit control.' }
    Write-Host 'ECO_GATE v41_nav_ask_activation=PASS'

    $buttons['Evidence'].accDoDefaultAction(0)
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
