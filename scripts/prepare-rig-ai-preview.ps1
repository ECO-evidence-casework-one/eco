# ECO rig AI preview preparer.
# Builds the exact qualified ECO source, installs only hash-pinned local AI assets,
# runs an independent offline Qwen smoke test, and launches an isolated preview.
# No administrator elevation, registry changes, service installation or machine-policy changes.
[CmdletBinding()]
param(
    [string]$OutputRoot = '',
    [switch]$NoLaunch,
    [switch]$SelfTest,
    [switch]$QualificationOnly
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$script:EcoSourceCommit = '24170e4505cfaadf97d37bc172ebf3097af7de55'
$script:EcoExpectedExeSHA256 = '190d06468a9cf282c1837ee08bf854f20724d42e70fc1e0bcb44479a015f36ba'
$script:EcoExpectedExeSize = 4880384
$script:GoURL = 'https://github.com/actions/go-versions/releases/download/1.23.12-16792118003/go-1.23.12-win32-x64.zip'
$script:GoSHA256 = 'c27b02f15d4ceb89fbce6ffe2a28df3dd293608cf79e9f12839f672863622845'
$script:LlamaTag = 'b10259'
$script:LlamaCommit = '1269cb1ff1598751f846241be90083ae9ad036fb'
$script:LlamaURL = 'https://github.com/ggml-org/llama.cpp/releases/download/b10259/llama-b10259-bin-win-cpu-x64.zip'
$script:LlamaArchiveSHA256 = '6613d8d56263233ef800fb8f8135231adb5eb851b40558281190242b0b20556b'
$script:ModelName = 'qwen2.5-1.5b-instruct-q4_k_m.gguf'
$script:ModelURL = 'https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf?download=true'
$script:ModelSHA256 = '6a1a2eb6d15622bf3c96857206351ba97e1af16c30d7a74ee38970e434e9407e'

function Get-EcoSHA256([string]$Path) {
    return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}

function Assert-EcoHash([string]$Path, [string]$Expected) {
    if ($Expected -notmatch '^[a-f0-9]{64}$') { throw 'Invalid expected SHA-256.' }
    $actual = Get-EcoSHA256 $Path
    if ($actual -cne $Expected) {
        throw "SHA-256 mismatch for $([IO.Path]::GetFileName($Path)). Expected $Expected but received $actual. Nothing will be launched."
    }
    return $actual
}

function Get-EcoCheckedDownload([string]$Url, [string]$Destination, [string]$ExpectedSHA256 = '') {
    if ($Url -notmatch '^https://') { throw 'Only HTTPS downloads are accepted.' }
    if (Test-Path -LiteralPath $Destination) { throw "Download destination already exists: $Destination" }
    $parent = Split-Path -Parent $Destination
    if ($parent) { [void][IO.Directory]::CreateDirectory($parent) }
    $part = $Destination + '.part'
    Remove-Item -LiteralPath $part -Force -ErrorAction SilentlyContinue
    $oldProgress = $ProgressPreference
    $ProgressPreference = 'SilentlyContinue'
    try {
        Invoke-WebRequest -UseBasicParsing -Uri $Url -OutFile $part -TimeoutSec 7200
        if ($ExpectedSHA256) { [void](Assert-EcoHash $part $ExpectedSHA256) }
        Move-Item -LiteralPath $part -Destination $Destination
    } finally {
        $ProgressPreference = $oldProgress
        Remove-Item -LiteralPath $part -Force -ErrorAction SilentlyContinue
    }
}

function Expand-EcoCheckedZip([string]$Archive, [string]$Destination, [string]$StripPrefix = '', [long]$MaxTotal = 2GB, [long]$MaxEntry = 1GB) {
    if (Test-Path -LiteralPath $Destination) { throw 'Extraction destination already exists; preserving it.' }
    Add-Type -AssemblyName System.IO.Compression
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $root = [IO.Path]::GetFullPath($Destination).TrimEnd('\') + '\'
    $zip = [IO.Compression.ZipFile]::OpenRead($Archive)
    try {
        if ($zip.Entries.Count -gt 50000) { throw 'Archive entry limit exceeded.' }
        $seen = New-Object 'System.Collections.Generic.HashSet[string]' ([StringComparer]::OrdinalIgnoreCase)
        $plan = New-Object 'System.Collections.Generic.List[object]'
        [long]$total = 0
        foreach ($entry in $zip.Entries) {
            $name = $entry.FullName.Replace('\', '/')
            if ($StripPrefix) {
                if (-not $name.StartsWith($StripPrefix, [StringComparison]::Ordinal)) { throw 'Unexpected archive root.' }
                $name = $name.Substring($StripPrefix.Length)
            }
            if (-not $name) { continue }
            if ($name.StartsWith('/') -or $name.Contains(':') -or $name -match '(^|/)\.\.?(/|$)') { throw 'Unsafe archive path.' }
            foreach ($part in $name.TrimEnd('/').Split('/')) {
                if (-not $part -or $part.EndsWith('.') -or $part.EndsWith(' ')) { throw 'Ambiguous Windows archive path.' }
            }
            $unixType = ($entry.ExternalAttributes -shr 16) -band 0xF000
            if ($unixType -eq 0xA000 -or (($entry.ExternalAttributes -band 0x400) -ne 0)) { throw 'Archive links are not accepted.' }
            $target = [IO.Path]::GetFullPath((Join-Path $Destination $name))
            if (-not $target.StartsWith($root, [StringComparison]::OrdinalIgnoreCase)) { throw 'Archive path escapes its directory.' }
            if (-not $seen.Add($target.TrimEnd('\'))) { throw 'Duplicate archive path.' }
            $total += $entry.Length
            if ($total -gt $MaxTotal -or $entry.Length -gt $MaxEntry) { throw 'Archive size limit exceeded.' }
            $plan.Add([pscustomobject]@{ Entry = $entry; Target = $target; Directory = $name.EndsWith('/') })
        }
        [void][IO.Directory]::CreateDirectory($Destination)
        foreach ($item in $plan) {
            if ($item.Directory) { [void][IO.Directory]::CreateDirectory($item.Target) }
            else {
                [void][IO.Directory]::CreateDirectory([IO.Path]::GetDirectoryName($item.Target))
                [IO.Compression.ZipFileExtensions]::ExtractToFile($item.Entry, $item.Target, $false)
            }
        }
    } finally { $zip.Dispose() }
}

function Invoke-EcoNative([string]$Label, [string]$Executable, [string[]]$Arguments) {
    & $Executable @Arguments | Out-Host
    if ($LASTEXITCODE -ne 0) { throw "$Label failed with exit code $LASTEXITCODE." }
}

function Get-EcoHardwareProfile {
    $profile = [ordered]@{
        captured_utc = [DateTimeOffset]::UtcNow.ToString('o')
        os_64_bit = [Environment]::Is64BitOperatingSystem
        os_version = [Environment]::OSVersion.VersionString
        processor = @()
        graphics = @()
        memory_bytes = $null
        output_drive_free_bytes = $null
    }
    try {
        $profile.processor = @(Get-CimInstance Win32_Processor | ForEach-Object {
            [ordered]@{ name=$_.Name; cores=$_.NumberOfCores; logical_processors=$_.NumberOfLogicalProcessors; max_clock_mhz=$_.MaxClockSpeed }
        })
    } catch { $profile.processor = @([ordered]@{ status='CIM unavailable' }) }
    try {
        $profile.graphics = @(Get-CimInstance Win32_VideoController | ForEach-Object {
            [ordered]@{ name=$_.Name; adapter_ram=$_.AdapterRAM; driver_version=$_.DriverVersion }
        })
    } catch { $profile.graphics = @([ordered]@{ status='CIM unavailable' }) }
    try { $profile.memory_bytes = [int64](Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory } catch { }
    return $profile
}

function Invoke-EcoPreparerSelfTest {
    $root = Join-Path ([IO.Path]::GetTempPath()) ('eco-rig-ai-selftest-' + [Guid]::NewGuid().ToString('N'))
    [void][IO.Directory]::CreateDirectory($root)
    try {
        $sample = Join-Path $root 'sample.txt'
        Set-Content -LiteralPath $sample -Value 'ECO self-test' -NoNewline -Encoding ascii
        $sha = Get-EcoSHA256 $sample
        [void](Assert-EcoHash $sample $sha)
        $mismatchRejected = $false
        try { [void](Assert-EcoHash $sample ('0' * 64)) } catch { $mismatchRejected = $true }
        if (-not $mismatchRejected) { throw 'Hash mismatch self-test did not reject altered identity.' }

        Add-Type -AssemblyName System.IO.Compression
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        $goodZip = Join-Path $root 'good.zip'
        $stream = [IO.File]::Open($goodZip, [IO.FileMode]::CreateNew)
        try {
            $zip = New-Object IO.Compression.ZipArchive($stream, [IO.Compression.ZipArchiveMode]::Create, $false)
            try {
                $entry = $zip.CreateEntry('root/ok.txt')
                $writer = New-Object IO.StreamWriter($entry.Open())
                try { $writer.Write('ok') } finally { $writer.Dispose() }
            } finally { $zip.Dispose() }
        } finally { $stream.Dispose() }
        $goodOut = Join-Path $root 'good-out'
        Expand-EcoCheckedZip $goodZip $goodOut 'root/'
        if ((Get-Content -LiteralPath (Join-Path $goodOut 'ok.txt') -Raw) -cne 'ok') { throw 'Safe extraction self-test failed.' }

        $badZip = Join-Path $root 'bad.zip'
        $stream = [IO.File]::Open($badZip, [IO.FileMode]::CreateNew)
        try {
            $zip = New-Object IO.Compression.ZipArchive($stream, [IO.Compression.ZipArchiveMode]::Create, $false)
            try { [void]$zip.CreateEntry('../escape.txt') } finally { $zip.Dispose() }
        } finally { $stream.Dispose() }
        $unsafeRejected = $false
        try { Expand-EcoCheckedZip $badZip (Join-Path $root 'bad-out') } catch { $unsafeRejected = $true }
        if (-not $unsafeRejected) { throw 'Unsafe archive self-test did not reject traversal.' }
        Write-Host 'Rig AI preparer self-test PASS.'
    } finally { Remove-Item -LiteralPath $root -Recurse -Force -ErrorAction SilentlyContinue }
}

function Invoke-EcoExactBuild([string]$WorkRoot) {
    [void][IO.Directory]::CreateDirectory($WorkRoot)
    $sourceZip = Join-Path $WorkRoot 'eco-source.zip'
    $goZip = Join-Path $WorkRoot 'go.zip'
    $sourceURL = "https://codeload.github.com/ECO-evidence-casework-one/eco/zip/$script:EcoSourceCommit"

    Write-Host '[1/7] Getting the exact merged ECO source from GitHub.'
    Get-EcoCheckedDownload $sourceURL $sourceZip
    $sourceArchiveSHA = Get-EcoSHA256 $sourceZip

    Write-Host '[2/7] Getting the exact GitHub-hosted Go 1.23.12 toolchain.'
    Get-EcoCheckedDownload $script:GoURL $goZip $script:GoSHA256

    $sourceDir = Join-Path $WorkRoot 'source'
    $compilerDir = Join-Path $WorkRoot 'compiler'
    Expand-EcoCheckedZip $sourceZip $sourceDir "eco-$script:EcoSourceCommit/"
    Expand-EcoCheckedZip $goZip $compilerDir
    $go = Join-Path $compilerDir 'go\bin\go.exe'
    if (-not (Test-Path -LiteralPath $go -PathType Leaf)) { throw 'Qualified Go archive did not contain go.exe.' }

    $envNames = @('GOENV','GOTOOLCHAIN','GOWORK','GOTELEMETRY','GOROOT','GOPATH','GOCACHE','GOMODCACHE','GOTMPDIR','GOOS','GOARCH','GOAMD64','CGO_ENABLED','GOMAXPROCS','GOFLAGS','GOEXPERIMENT','GODEBUG','GOPROXY','GOSUMDB','GOPRIVATE','GONOSUMDB','GONOPROXY','TEMP','TMP')
    $saved = @{}
    foreach ($name in $envNames) { $saved[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
    $oldLocation = Get-Location
    try {
        $env:GOROOT = Join-Path $compilerDir 'go'
        $env:GOENV = 'off'; $env:GOTOOLCHAIN = 'local'; $env:GOWORK = 'off'; $env:GOTELEMETRY = 'off'
        $env:GOPATH = Join-Path $WorkRoot 'go-cache'; $env:GOCACHE = Join-Path $WorkRoot 'go-cache\build'
        $env:GOMODCACHE = Join-Path $WorkRoot 'go-cache\modules'; $env:GOTMPDIR = Join-Path $WorkRoot 'temporary'
        [void][IO.Directory]::CreateDirectory($env:GOTMPDIR)
        $env:TEMP = $env:GOTMPDIR; $env:TMP = $env:GOTMPDIR
        $env:GOOS = 'windows'; $env:GOARCH = 'amd64'; $env:GOAMD64 = 'v1'; $env:CGO_ENABLED = '0'
        $env:GOMAXPROCS = '2'; $env:GOFLAGS = '-mod=readonly -buildvcs=false'
        $env:GOEXPERIMENT = ''; $env:GODEBUG = ''; $env:GOPRIVATE = ''; $env:GONOSUMDB = ''; $env:GONOPROXY = ''
        $env:GOPROXY = 'https://proxy.golang.org'; $env:GOSUMDB = 'sum.golang.org'
        Set-Location -LiteralPath $sourceDir
        $version = (& $go version | Out-String).Trim()
        if ($LASTEXITCODE -ne 0 -or $version -cne 'go version go1.23.12 windows/amd64') { throw "Unexpected Go toolchain identity: $version" }

        Write-Host '[3/7] Verifying source-pinned Go dependencies.'
        Invoke-EcoNative 'go mod download' $go @('mod','download')
        Invoke-EcoNative 'go mod verify' $go @('mod','verify')
        $env:GOPROXY = 'off'

        Write-Host '[4/7] Running ECO tests and vet before building.'
        Invoke-EcoNative 'go test' $go @('test','-count=1','-p=2','./...')
        Invoke-EcoNative 'go vet' $go @('vet','-p=2','./...')

        Write-Host '[5/7] Building the exact ECO Windows candidate twice.'
        $first = Join-Path $WorkRoot 'ECO.exe'
        $second = Join-Path $WorkRoot 'ECO.rebuild.exe'
        $ldflags = "-s -w -H windowsgui -buildid= -X github.com/ECO-evidence-casework-one/eco/internal/eco.SourceCommit=$script:EcoSourceCommit"
        Invoke-EcoNative 'first deterministic ECO build' $go @('build','-trimpath','-buildvcs=false','-ldflags',$ldflags,'-o',$first,'./cmd/eco')
        Invoke-EcoNative 'second deterministic ECO build' $go @('build','-trimpath','-buildvcs=false','-ldflags',$ldflags,'-o',$second,'./cmd/eco')
        [void](Assert-EcoHash $first $script:EcoExpectedExeSHA256)
        [void](Assert-EcoHash $second $script:EcoExpectedExeSHA256)
        if ((Get-Item -LiteralPath $first).Length -ne $script:EcoExpectedExeSize) { throw 'ECO executable size differs from the post-merge qualified Windows build.' }
        Remove-Item -LiteralPath $second -Force
        return [pscustomobject]@{ Exe=$first; SourceArchiveSHA256=$sourceArchiveSHA; GoVersion=$version; SourceDir=$sourceDir }
    } finally {
        Set-Location $oldLocation
        foreach ($name in $envNames) { [Environment]::SetEnvironmentVariable($name, $saved[$name], 'Process') }
    }
}

function Install-EcoLlamaRuntime([string]$WorkRoot, [string]$TargetRoot) {
    Write-Host '[6/7] Getting the pinned llama.cpp Windows CPU runtime from GitHub.'
    $archive = Join-Path $WorkRoot 'llama.zip'
    Get-EcoCheckedDownload $script:LlamaURL $archive $script:LlamaArchiveSHA256
    $temp = Join-Path $WorkRoot 'llama-extracted'
    Expand-EcoCheckedZip $archive $temp
    $matches = @(Get-ChildItem -LiteralPath $temp -Filter 'llama-cli.exe' -File -Recurse)
    if ($matches.Count -ne 1) { throw "Expected exactly one llama-cli.exe, found $($matches.Count)." }
    if (Test-Path -LiteralPath $TargetRoot) { throw 'Runtime target already exists.' }
    [IO.Directory]::Move($temp, $TargetRoot)
    $llama = @(Get-ChildItem -LiteralPath $TargetRoot -Filter 'llama-cli.exe' -File -Recurse)[0].FullName
    $llamaSHA = Get-EcoSHA256 $llama
    $versionText = (& $llama '--offline' '--version' 2>&1 | Out-String).Trim()
    if ($LASTEXITCODE -ne 0 -or -not $versionText) { throw 'llama.cpp version probe failed.' }
    return [pscustomobject]@{ Exe=$llama; SHA256=$llamaSHA; Version=($versionText -split "`r?`n")[0] }
}

function Get-EcoModel([string]$TargetRoot) {
    [void][IO.Directory]::CreateDirectory($TargetRoot)
    $model = Join-Path $TargetRoot $script:ModelName
    if (Test-Path -LiteralPath $model -PathType Leaf) {
        [void](Assert-EcoHash $model $script:ModelSHA256)
        return $model
    }
    if ($env:ECO_LLAMA_MODEL -and (Test-Path -LiteralPath $env:ECO_LLAMA_MODEL -PathType Leaf)) {
        try {
            [void](Assert-EcoHash $env:ECO_LLAMA_MODEL $script:ModelSHA256)
            Copy-Item -LiteralPath $env:ECO_LLAMA_MODEL -Destination $model
            [void](Assert-EcoHash $model $script:ModelSHA256)
            return $model
        } catch { }
    }
    Write-Host '[7/7] Getting the official Qwen2.5 1.5B Instruct Q4_K_M model and checking its published SHA-256.'
    Get-EcoCheckedDownload $script:ModelURL $model $script:ModelSHA256
    return $model
}

function Invoke-EcoAIProbe([string]$Llama, [string]$Model, [string]$ProbeRoot) {
    [void][IO.Directory]::CreateDirectory($ProbeRoot)
    $prompt = Join-Path $ProbeRoot 'prompt.txt'
    $schema = Join-Path $ProbeRoot 'schema.json'
    Set-Content -LiteralPath $prompt -Encoding UTF8 -NoNewline -Value 'Return the JSON object required by the schema. Set ok to ECO_AI_READY. Do not add prose.'
    Set-Content -LiteralPath $schema -Encoding UTF8 -NoNewline -Value '{"type":"object","additionalProperties":false,"required":["ok"],"properties":{"ok":{"type":"string","enum":["ECO_AI_READY"]}}}'
    $saved = @{}
    foreach ($name in @('LLAMA_ARG_OFFLINE','HF_HUB_OFFLINE','TRANSFORMERS_OFFLINE','HF_TOKEN','HUGGING_FACE_HUB_TOKEN','HF_ENDPOINT','HTTP_PROXY','HTTPS_PROXY','ALL_PROXY')) {
        $saved[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
    }
    try {
        $env:LLAMA_ARG_OFFLINE='1'; $env:HF_HUB_OFFLINE='1'; $env:TRANSFORMERS_OFFLINE='1'
        foreach ($name in @('HF_TOKEN','HUGGING_FACE_HUB_TOKEN','HF_ENDPOINT','HTTP_PROXY','HTTPS_PROXY','ALL_PROXY')) { [Environment]::SetEnvironmentVariable($name, $null, 'Process') }
        $args = @('--offline','--model',$Model,'--file',$prompt,'--json-schema-file',$schema,'--simple-io','--no-display-prompt','--color','off','--log-disable','--seed','0','--temp','0','--top-k','1','--top-p','1','--min-p','0','--ctx-size','2048','--n-predict','64','--device','none','--n-gpu-layers','0','--fit','off','--no-context-shift','--no-perf')
        $output = (& $Llama @args 2>&1 | Out-String).Trim()
        if ($LASTEXITCODE -ne 0) { throw "Qwen smoke test failed with exit code $LASTEXITCODE. $output" }
        try { $json = $output | ConvertFrom-Json } catch { throw "Qwen smoke test did not return valid JSON. Output: $output" }
        if ($json.ok -cne 'ECO_AI_READY') { throw "Qwen smoke test returned an unexpected result: $output" }
        return $output
    } finally {
        foreach ($name in $saved.Keys) { [Environment]::SetEnvironmentVariable($name, $saved[$name], 'Process') }
    }
}

function Write-EcoLauncher([string]$Root, [string]$LlamaSHA) {
    $runtimeRelative = [IO.Path]::GetFullPath((Get-ChildItem -LiteralPath (Join-Path $Root 'Runtime') -Filter 'llama-cli.exe' -File -Recurse | Select-Object -First 1).FullName).Substring([IO.Path]::GetFullPath($Root).TrimEnd('\').Length + 1)
    $lines = @(
        '@echo off',
        'setlocal',
        'cd /d "%~dp0"',
        'if not exist "PreviewUserData" mkdir "PreviewUserData"',
        'set "LOCALAPPDATA=%~dp0PreviewUserData"',
        "set `"ECO_LLAMA_CPP=%~dp0$runtimeRelative`"",
        "set `"ECO_LLAMA_CPP_SHA256=$LlamaSHA`"",
        "set `"ECO_LLAMA_MODEL=%~dp0AIAssets\$script:ModelName`"",
        "set `"ECO_LLAMA_MODEL_SHA256=$script:ModelSHA256`"",
        'start "" "%~dp0ECO.exe"',
        'endlocal'
    )
    $lines | Set-Content -LiteralPath (Join-Path $Root 'START_ECO_WITH_AI.cmd') -Encoding ascii
}

if ($SelfTest) {
    Invoke-EcoPreparerSelfTest
    if (-not $QualificationOnly) { exit 0 }
}

if (-not [Environment]::Is64BitOperatingSystem) { throw 'A 64-bit Windows machine is required.' }
if (-not $OutputRoot) { $OutputRoot = Join-Path $PSScriptRoot 'ECO_RIG_AI_PREVIEW' }
$OutputRoot = [IO.Path]::GetFullPath($OutputRoot)
if ($OutputRoot.StartsWith('\\')) { throw 'Use a local drive folder for the rig preview.' }
if (Test-Path -LiteralPath $OutputRoot) { throw "Output folder already exists and will not be overwritten: $OutputRoot" }
[void][IO.Directory]::CreateDirectory($OutputRoot)
$work = Join-Path $OutputRoot '.work'
[void][IO.Directory]::CreateDirectory($work)
$resultText = Join-Path $OutputRoot 'AI_SETUP_RESULT.txt'
$resultJSON = Join-Path $OutputRoot 'AI_SETUP_RESULT.json'
$transcript = $false
$oldTls = [Net.ServicePointManager]::SecurityProtocol
try {
    Start-Transcript -Path (Join-Path $OutputRoot 'setup-log.txt') | Out-Null
    $transcript = $true
    [Net.ServicePointManager]::SecurityProtocol = $oldTls -bor [Net.SecurityProtocolType]::Tls12

    $hardware = Get-EcoHardwareProfile
    try {
        $drive = Get-PSDrive -Name ([IO.Path]::GetPathRoot($OutputRoot).Substring(0,1))
        $hardware.output_drive_free_bytes = [int64]$drive.Free
        if ($drive.Free -lt 4GB) { throw 'At least 4 GB of free space is required for the controlled build, runtime and model.' }
    } catch { if ($_.Exception.Message -like 'At least 4 GB*') { throw } }
    $hardware | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $OutputRoot 'RIG_AI_HARDWARE.json') -Encoding UTF8

    $build = Invoke-EcoExactBuild $work
    Copy-Item -LiteralPath $build.Exe -Destination (Join-Path $OutputRoot 'ECO.exe')
    [void](Assert-EcoHash (Join-Path $OutputRoot 'ECO.exe') $script:EcoExpectedExeSHA256)
    if (Test-Path -LiteralPath (Join-Path $build.SourceDir 'LICENSE')) { Copy-Item -LiteralPath (Join-Path $build.SourceDir 'LICENSE') -Destination (Join-Path $OutputRoot 'ECO_LICENSE.txt') }

    $runtime = Install-EcoLlamaRuntime $work (Join-Path $OutputRoot 'Runtime')

    if ($QualificationOnly) {
        $qualification = [ordered]@{
            status='PASS'; qualification_only=$true; source_commit=$script:EcoSourceCommit; source_archive_sha256=$build.SourceArchiveSHA256
            eco_exe_sha256=$script:EcoExpectedExeSHA256; eco_exe_size=$script:EcoExpectedExeSize; go_version=$build.GoVersion
            llama_tag=$script:LlamaTag; llama_commit=$script:LlamaCommit; llama_archive_sha256=$script:LlamaArchiveSHA256
            llama_exe_sha256=$runtime.SHA256; llama_version=$runtime.Version; model_downloaded=$false; app_launched=$false
        }
        $qualification | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $resultJSON -Encoding UTF8
        @('RIG AI PACKAGE QUALIFICATION PASS', "ECO source: $script:EcoSourceCommit", "ECO EXE SHA-256: $script:EcoExpectedExeSHA256", "Source ZIP observed SHA-256: $($build.SourceArchiveSHA256)", "llama.cpp archive SHA-256: $script:LlamaArchiveSHA256", "llama-cli.exe SHA-256: $($runtime.SHA256)", "llama.cpp version: $($runtime.Version)", 'Model not downloaded in CI qualification; official model identity is pinned separately.') | Set-Content -LiteralPath $resultText -Encoding UTF8
        Get-Content -LiteralPath $resultText | Write-Host
        exit 0
    }

    $model = Get-EcoModel (Join-Path $OutputRoot 'AIAssets')
    $modelSHA = Assert-EcoHash $model $script:ModelSHA256
    Write-Host 'Running a real offline Qwen generation smoke test before ECO is allowed to start.'
    $probeOutput = Invoke-EcoAIProbe $runtime.Exe $model (Join-Path $work 'ai-probe')
    Write-EcoLauncher $OutputRoot $runtime.SHA256

    $receipt = [ordered]@{
        status='PASS'; created_utc=[DateTimeOffset]::UtcNow.ToString('o'); source_commit=$script:EcoSourceCommit
        source_archive_sha256=$build.SourceArchiveSHA256; eco_exe_sha256=$script:EcoExpectedExeSHA256; eco_exe_size=$script:EcoExpectedExeSize
        go_version=$build.GoVersion; llama_upstream='ggml-org/llama.cpp'; llama_license='MIT'; llama_tag=$script:LlamaTag; llama_commit=$script:LlamaCommit
        llama_archive_sha256=$script:LlamaArchiveSHA256; llama_exe_sha256=$runtime.SHA256; llama_version=$runtime.Version
        model_upstream='Qwen/Qwen2.5-1.5B-Instruct-GGUF'; model_license='Apache-2.0'; model_file=$script:ModelName; model_sha256=$modelSHA
        offline_generation_probe='PASS'; offline_generation_probe_output=$probeOutput; workspace='isolated PreviewUserData via process-only LOCALAPPDATA'
        public_release=$false; real_evidence_permitted=$false; app_launched=(-not $NoLaunch)
    }
    $receipt | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $resultJSON -Encoding UTF8
    @('ECO RIG AI SETUP PASS', 'The current ECO application and the local Qwen model both passed their checks.', "ECO EXE SHA-256: $script:EcoExpectedExeSHA256", "Qwen model SHA-256: $modelSHA", "llama-cli.exe SHA-256: $($runtime.SHA256)", 'Independent offline Qwen generation: PASS', 'This is an isolated private developer preview. Do not import real evidence yet.', 'Use START_ECO_WITH_AI.cmd to open it again.') | Set-Content -LiteralPath $resultText -Encoding UTF8
    @('ECO RIG AI PREVIEW', '', 'Double-click START_ECO_WITH_AI.cmd.', 'The launcher changes LOCALAPPDATA only for this ECO process, so this preview uses its own PreviewUserData folder.', 'The Qwen model and llama.cpp runtime are local and hash-pinned. ECO itself remains offline.', 'This is not a public release and is for synthetic/test material only.') | Set-Content -LiteralPath (Join-Path $OutputRoot 'README_FIRST.txt') -Encoding UTF8

    Remove-Item -LiteralPath $work -Recurse -Force -ErrorAction SilentlyContinue
    Get-Content -LiteralPath $resultText | Write-Host
    if (-not $NoLaunch) { Start-Process -FilePath (Join-Path $OutputRoot 'START_ECO_WITH_AI.cmd') -WorkingDirectory $OutputRoot }
} catch {
    $message = $_.Exception.Message
    if ($env:USERPROFILE) { $message = $message.Replace($env:USERPROFILE, '[user-profile]') }
    @('ECO RIG AI SETUP STOPPED', $message, 'Nothing changed Windows security or machine policy. Do not disable Defender or Smart App Control.', 'Send AI_SETUP_RESULT.txt and setup-log.txt back to the ECO development chat.') | Set-Content -LiteralPath $resultText -Encoding UTF8
    Write-Host $message
    exit 1
} finally {
    [Net.ServicePointManager]::SecurityProtocol = $oldTls
    if ($transcript) { Stop-Transcript | Out-Null }
}
