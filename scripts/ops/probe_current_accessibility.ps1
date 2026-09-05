param(
    [Parameter(Mandatory=$true)][string]$ExePath,
    [Parameter(Mandatory=$true)][string]$TempRoot
)
$ErrorActionPreference='Stop'
$oldLocal=$env:LOCALAPPDATA
$env:LOCALAPPDATA=Join-Path $TempRoot 'localappdata'
New-Item -ItemType Directory -Force -Path $env:LOCALAPPDATA | Out-Null

Add-Type -AssemblyName Accessibility
$accAssembly=[Accessibility.IAccessible].Assembly.Location
Add-Type -TypeDefinition @'
using System;
using System.Runtime.InteropServices;
using System.Text;
public sealed class EcoA11yRow { public IntPtr Hwnd; public int ControlId; public string ClassName; public string WindowText; public bool Visible; public string AccessibleName; public string DefaultAction; public int Role; public int State; }
public static class EcoA11yProbe {
  const uint OBJID_CLIENT=0xFFFFFFFC; delegate bool EnumProc(IntPtr h,IntPtr l);
  [DllImport("user32.dll")] static extern bool EnumWindows(EnumProc cb,IntPtr l);
  [DllImport("user32.dll")] static extern bool EnumChildWindows(IntPtr p,EnumProc cb,IntPtr l);
  [DllImport("user32.dll")] static extern uint GetWindowThreadProcessId(IntPtr h,out uint pid);
  [DllImport("user32.dll",CharSet=CharSet.Unicode)] static extern int GetClassName(IntPtr h,StringBuilder b,int n);
  [DllImport("user32.dll",CharSet=CharSet.Unicode)] static extern int GetWindowText(IntPtr h,StringBuilder b,int n);
  [DllImport("user32.dll")] static extern int GetWindowTextLength(IntPtr h);
  [DllImport("user32.dll")] static extern bool IsWindowVisible(IntPtr h);
  [DllImport("user32.dll")] static extern int GetDlgCtrlID(IntPtr h);
  [DllImport("oleacc.dll")] static extern int AccessibleObjectFromWindow(IntPtr h,uint id,ref Guid iid,[MarshalAs(UnmanagedType.Interface)] out Accessibility.IAccessible acc);
  static string C(IntPtr h){var b=new StringBuilder(256);GetClassName(h,b,b.Capacity);return b.ToString();}
  static string T(IntPtr h){var b=new StringBuilder(Math.Max(1,GetWindowTextLength(h)+1));GetWindowText(h,b,b.Capacity);return b.ToString();}
  static EcoA11yRow R(IntPtr h){var r=new EcoA11yRow{Hwnd=h,ControlId=GetDlgCtrlID(h),ClassName=C(h),WindowText=T(h),Visible=IsWindowVisible(h)};try{Guid g=typeof(Accessibility.IAccessible).GUID;Accessibility.IAccessible a; if(AccessibleObjectFromWindow(h,OBJID_CLIENT,ref g,out a)>=0&&a!=null){object self=0;try{r.AccessibleName=a.get_accName(self);}catch{}try{r.DefaultAction=a.get_accDefaultAction(self);}catch{}try{r.Role=Convert.ToInt32(a.get_accRole(self));}catch{}try{r.State=Convert.ToInt32(a.get_accState(self));}catch{}}}catch{}return r;}
  public static IntPtr Main(uint pid,string cls){IntPtr f=IntPtr.Zero;EnumWindows((h,l)=>{uint p;GetWindowThreadProcessId(h,out p);if(p==pid&&C(h)==cls){f=h;return false;}return true;},IntPtr.Zero);return f;}
  public static EcoA11yRow[] Children(IntPtr p){IntPtr[] h=new IntPtr[256];int c=0;EnumChildWindows(p,(x,l)=>{if(c<h.Length)h[c++]=x;return true;},IntPtr.Zero);var r=new EcoA11yRow[c];for(int i=0;i<c;i++)r[i]=R(h[i]);return r;}
}
'@ -ReferencedAssemblies $accAssembly

$p=Start-Process -FilePath $ExePath -PassThru
try {
  $deadline=(Get-Date).AddSeconds(20);$main=[IntPtr]::Zero
  while($main -eq [IntPtr]::Zero -and (Get-Date) -lt $deadline -and -not $p.HasExited){Start-Sleep -Milliseconds 200;$main=[EcoA11yProbe]::Main([uint32]$p.Id,'ECO_V25_NATIVE_MAIN')}
  if($main -eq [IntPtr]::Zero){throw 'Current ECO main HWND was not found.'}
  $rows=@([EcoA11yProbe]::Children($main))
  $rows|Select ControlId,ClassName,WindowText,Visible,AccessibleName,DefaultAction,Role,State|Format-Table -AutoSize|Out-String|Write-Host
  $expected=@{1001='Edit';1002='Button';1003='Edit';1004='Edit';1005='Button';1006='Button';1007='Button';1008='Button';1009='Button'}
  $semanticPass=$true
  foreach($id in $expected.Keys){$row=@($rows|?{$_.ControlId -eq $id})|Select -First 1;if($null -eq $row -or $row.ClassName -ne $expected[$id] -or $row.Role -eq 0){$semanticPass=$false}}
  $painted=@('Home','Evidence','Matters','Review','Changes','Trust & settings','+  Add files','+  Add folder','Paste image','Open native preview')
  $missing=@();foreach($label in $painted){if(-not @($rows|?{$_.WindowText -eq $label -or $_.AccessibleName -eq $label}).Count){$missing+=$label}}
  $report=@('# ECO current-main accessibility baseline','',('Source commit: `'+$env:GITHUB_SHA+'`'),('Standard child HWNDs inspected: **'+$rows.Count+'**'),('Ask/search native semantic controls: **'+$(if($semanticPass){'PASS'}else{'FAIL'})+'**'),('Known custom-painted interactive labels with no child HWND semantic peer: **'+$missing.Count+' / '+$painted.Count+'**'),'','## Missing native semantic peers in this baseline')
  foreach($label in $missing){$report+=('- '+$label)}
  $report+=@('','## Interpretation','- Existing native EDIT/BUTTON controls expose established Windows accessibility contracts.','- The listed custom-painted interactions are not represented as child HWND semantic controls and therefore are not qualified for Narrator/NVDA from this evidence.','- Hosted-runner MSAA evidence is a structural baseline only; physical Windows 11 Narrator/NVDA remains mandatory.')
  Set-Content -LiteralPath 'docs/testing/CURRENT_MAIN_ACCESSIBILITY_BASELINE_20260905.md' -Value ($report -join "`n") -Encoding UTF8
  if(-not $semanticPass){throw 'Existing native Ask/search controls failed MSAA semantic baseline.'}
  Write-Host 'ECO_GATE current_main_native_controls_msaa=PASS'
  Write-Host ('ECO_GATE current_main_custom_painted_missing_semantics='+$missing.Count)
} finally { if($p -and -not $p.HasExited){Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue};$env:LOCALAPPDATA=$oldLocal }
