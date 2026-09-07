# ECO rig AI preview preparer v3.
# GitHub-first, hash-pinned, private synthetic-data preview.
# Windows PowerShell 5.1 compatible. Large Qwen downloads use visible,
# resumable curl.exe transfers with stall detection and exact SHA-256 approval.
[CmdletBinding()]
param(
    [string]$OutputRoot = '',
    [switch]$NoLaunch,
    [switch]$SelfTest,
    [switch]$QualificationOnly
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$EcoSource = 'b03ec2358dbf437deb922ad1cbb96d4e5c6faedb'
$EcoExeSHA = 'eb6159cb0406a0d1b7195285f03848048026726e6abe3c73b9f7a4f52a9fbee3'
$EcoExeSize = 4892672
$GoURL = 'https://github.com/actions/go-versions/releases/download/1.23.12-16792118003/go-1.23.12-win32-x64.zip'
$GoZipSHA = 'c27b02f15d4ceb89fbce6ffe2a28df3dd293608cf79e9f12839f672863622845'
$LlamaURL = 'https://github.com/ggml-org/llama.cpp/releases/download/b10259/llama-b10259-bin-win-cpu-x64.zip'
$LlamaZipSHA = '6613d8d56263233ef800fb8f8135231adb5eb851b40558281190242b0b20556b'
$ModelName = 'qwen2.5-1.5b-instruct-q4_k_m.gguf'
$ModelURL = 'https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf?download=true'
$ModelSHA = '6a1a2eb6d15622bf3c96857206351ba97e1af16c30d7a74ee38970e434e9407e'

function Get-SHA256([string]$Path) {
    (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}

function Assert-SHA256([string]$Path, [string]$Expected) {
    $actual = Get-SHA256 $Path
    if ($actual -cne $Expected) {
        throw "SHA-256 mismatch for $([IO.Path]::GetFileName($Path)). Expected $Expected but received $actual."
    }
    return $actual
}

function Format-Bytes([int64]$Bytes) {
    if ($Bytes -ge 1GB) { return ('{0:N2} GB' -f ($Bytes / 1GB)) }
    if ($Bytes -ge 1MB) { return ('{0:N1} MB' -f ($Bytes / 1MB)) }
    if ($Bytes -ge 1KB) { return ('{0:N1} KB' -f ($Bytes / 1KB)) }
    return "$Bytes bytes"
}

function Download-CheckedSmall([string]$Url, [string]$Path, [string]$ExpectedSHA256 = '') {
    if ($Url -notmatch '^https://') { throw 'Only HTTPS downloads are allowed.' }
    if (Test-Path -LiteralPath $Path) { throw "Will not overwrite $Path" }
    [void][IO.Directory]::CreateDirectory((Split-Path -Parent $Path))
    $part = $Path + '.part'
    $oldProgress = $ProgressPreference
    $ProgressPreference = 'SilentlyContinue'
    try {
        Invoke-WebRequest -UseBasicParsing -Uri $Url -OutFile $part -TimeoutSec 7200
        if ($ExpectedSHA256) { [void](Assert-SHA256 $part $ExpectedSHA256) }
        Move-Item -LiteralPath $part -Destination $Path
    } finally {
        $ProgressPreference = $oldProgress
        Remove-Item -LiteralPath $part -Force -ErrorAction SilentlyContinue
    }
}

function Expand-CheckedZip([string]$Archive, [string]$Destination) {
    if (Test-Path -LiteralPath $Destination) { throw "Will not overwrite extraction folder $Destination" }
    Expand-Archive -LiteralPath $Archive -DestinationPath $Destination
}

function Invoke-Native([string]$Label, [string]$Executable, [string[]]$Arguments) {
    & $Executable @Arguments | Out-Host
    if ($LASTEXITCODE -ne 0) { throw "$Label failed with exit code $LASTEXITCODE" }
}

function Capture-Native([string]$Executable, [string[]]$Arguments, [string]$Work) {
    [void][IO.Directory]::CreateDirectory($Work)
    $token = [guid]::NewGuid().ToString('N')
    $stdoutPath = Join-Path $Work ("native-$token.stdout.txt")
    $stderrPath = Join-Path $Work ("native-$token.stderr.txt")
    $oldAction = $ErrorActionPreference
    try {
        $ErrorActionPreference = 'Continue'
        & $Executable @Arguments 1> $stdoutPath 2> $stderrPath
        $exitCode = $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $oldAction
    }
    $stdout = ''
    $stderr = ''
    if (Test-Path -LiteralPath $stdoutPath) { $stdout = Get-Content -LiteralPath $stdoutPath -Raw -ErrorAction SilentlyContinue }
    if (Test-Path -LiteralPath $stderrPath) { $stderr = Get-Content -LiteralPath $stderrPath -Raw -ErrorAction SilentlyContinue }
    Remove-Item -LiteralPath $stdoutPath,$stderrPath -Force -ErrorAction SilentlyContinue
    [pscustomobject]@{ ExitCode = [int]$exitCode; Stdout = [string]$stdout; Stderr = [string]$stderr }
}

function Native-Diagnostic([object]$Capture) {
    (($Capture.Stdout + "`n" + $Capture.Stderr).Trim())
}

function Find-CurlExe {
    $command = Get-Command -Name 'curl.exe' -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($null -eq $command -or -not $command.Source) { throw 'Windows curl.exe was not found.' }
    return $command.Source
}

function Convert-ToProcessArgumentLine([string[]]$Arguments) {
    return (($Arguments | ForEach-Object { '"' + ([string]$_).Replace('"','\"') + '"' }) -join ' ')
}

function Probe-QwenEndpoint([string]$Work) {
    $curl = Find-CurlExe
    $probe = Join-Path $Work ('qwen-endpoint-' + [guid]::NewGuid().ToString('N') + '.part')
    $headers = $probe + '.headers'
    try {
        $capture = Capture-Native $curl @(
            '--location', '--fail', '--silent', '--show-error',
            '--connect-timeout', '30', '--max-time', '90',
            '--range', '0-0', '--max-filesize', '1048576',
            '--dump-header', $headers, '--output', $probe, $ModelURL
        ) $Work
        if ($capture.ExitCode -ne 0) {
            throw "Official Qwen endpoint probe failed: $(Native-Diagnostic $capture)"
        }
        if (-not (Test-Path -LiteralPath $probe -PathType Leaf)) { throw 'Official Qwen endpoint returned no bounded probe bytes.' }
        $size = (Get-Item -LiteralPath $probe).Length
        if ($size -lt 1 -or $size -gt 1048576) { throw "Official Qwen endpoint ignored the bounded range test ($size bytes)." }
        Write-Host "Official Qwen endpoint/range probe PASS ($size byte(s))."
    } finally {
        Remove-Item -LiteralPath $probe,$headers -Force -ErrorAction SilentlyContinue
    }
}

function Download-QwenModel([string]$Destination) {
    if (Test-Path -LiteralPath $Destination -PathType Leaf) {
        [void](Assert-SHA256 $Destination $ModelSHA)
        Write-Host 'Existing Qwen model SHA-256 PASS.'
        return
    }

    [void][IO.Directory]::CreateDirectory((Split-Path -Parent $Destination))
    $partial = $Destination + '.part'
    $curl = Find-CurlExe
    $maxAttempts = 4

    for ($attempt = 1; $attempt -le $maxAttempts; $attempt++) {
        $existingBytes = [int64]0
        if (Test-Path -LiteralPath $partial -PathType Leaf) { $existingBytes = (Get-Item -LiteralPath $partial).Length }
        if ($existingBytes -gt 0) {
            Write-Host "Qwen transfer attempt $attempt/$maxAttempts — resuming from $(Format-Bytes $existingBytes)."
        } else {
            Write-Host "Qwen transfer attempt $attempt/$maxAttempts — starting from zero."
        }
        Write-Host 'curl progress below shows percent, downloaded bytes, speed and remaining time.'
        Write-Host 'If transfer speed stays below 1 byte/sec for 3 minutes, curl will stop and ECO will retry from the partial file.'

        $arguments = @(
            '--location', '--fail', '--show-error',
            '--connect-timeout', '30',
            '--speed-limit', '1', '--speed-time', '180',
            '--retry', '2', '--retry-delay', '5',
            '--continue-at', '-',
            '--output', $partial,
            $ModelURL
        )
        $process = Start-Process -FilePath $curl -ArgumentList (Convert-ToProcessArgumentLine $arguments) -NoNewWindow -PassThru -Wait
        $exitCode = $process.ExitCode
        if ($exitCode -eq 0) { break }

        $kept = [int64]0
        if (Test-Path -LiteralPath $partial -PathType Leaf) { $kept = (Get-Item -LiteralPath $partial).Length }
        if ($attempt -eq $maxAttempts) {
            throw "Qwen download failed after $maxAttempts attempts. Partial bytes remain at $partial ($(Format-Bytes $kept))."
        }
        Write-Host "curl exited $exitCode. Kept $(Format-Bytes $kept); retrying in 10 seconds."
        Start-Sleep -Seconds 10
    }

    if (-not (Test-Path -LiteralPath $partial -PathType Leaf)) { throw 'Qwen transfer reported success but no downloaded file exists.' }
    Write-Host 'Qwen transfer completed. Verifying the official published SHA-256...'
    $actual = Get-SHA256 $partial
    if ($actual -cne $ModelSHA) {
        $bad = $partial + '.bad-sha-' + (Get-Date -Format 'yyyyMMdd-HHmmss')
        Move-Item -LiteralPath $partial -Destination $bad
        throw "Qwen SHA-256 mismatch. Untrusted bytes were preserved at $bad and will not be used."
    }
    Move-Item -LiteralPath $partial -Destination $Destination
    Write-Host 'Qwen SHA-256 PASS.'
}

function Find-PortableGo([string]$Root) {
    $matches = @(Get-ChildItem -LiteralPath $Root -Filter 'go.exe' -File -Recurse | Where-Object { $_.Directory.Name -ieq 'bin' })
    if ($matches.Count -ne 1) { throw "Expected exactly one bin\go.exe, found $($matches.Count)." }
    $go = $matches[0].FullName
    $goroot = Split-Path -Parent (Split-Path -Parent $go)
    if (-not (Test-Path -LiteralPath (Join-Path $goroot 'src') -PathType Container)) { throw 'Verified Go archive has an unexpected GOROOT layout.' }
    [pscustomobject]@{ Go = $go; Root = $goroot }
}

function Find-Llama([string]$Root) {
    $matches = @(Get-ChildItem -LiteralPath $Root -Filter 'llama-cli.exe' -File -Recurse)
    if ($matches.Count -ne 1) { throw "Expected exactly one llama-cli.exe, found $($matches.Count)." }
    return $matches[0].FullName
}

function Hardware-Receipt([string]$Path) {
    $r = [ordered]@{ captured_utc=[DateTimeOffset]::UtcNow.ToString('o'); os_64_bit=[Environment]::Is64BitOperatingSystem; os=[Environment]::OSVersion.VersionString; cpu=@(); gpu=@(); memory_bytes=$null }
    try { $r.cpu = @(Get-CimInstance Win32_Processor | ForEach-Object { [ordered]@{name=$_.Name;cores=$_.NumberOfCores;logical=$_.NumberOfLogicalProcessors;max_mhz=$_.MaxClockSpeed} }) } catch {}
    try { $r.gpu = @(Get-CimInstance Win32_VideoController | ForEach-Object { [ordered]@{name=$_.Name;adapter_ram=$_.AdapterRAM;driver=$_.DriverVersion} }) } catch {}
    try { $r.memory_bytes = [int64](Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory } catch {}
    $r | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $Path -Encoding UTF8
}

function Write-Utf8NoBom([string]$Path, [string]$Text) {
    [IO.File]::WriteAllText($Path, $Text, (New-Object System.Text.UTF8Encoding($false)))
}

function Invoke-PreparerSelfTest {
    $root = Join-Path ([IO.Path]::GetTempPath()) ('eco-ai-selftest-' + [guid]::NewGuid().ToString('N'))
    [void][IO.Directory]::CreateDirectory($root)
    try {
        $sample = Join-Path $root 'sample.txt'
        Set-Content -LiteralPath $sample -Value 'ECO self-test' -NoNewline -Encoding ascii
        [void](Assert-SHA256 $sample (Get-SHA256 $sample))
        $capture = Capture-Native $env:ComSpec @('/d','/c','echo version: ECO_CAPTURE_TEST 1>&2 & exit /b 0') $root
        if ($capture.ExitCode -ne 0 -or $capture.Stderr -notmatch 'ECO_CAPTURE_TEST') { throw 'Native stderr capture self-test failed.' }
        $curlCapture = Capture-Native (Find-CurlExe) @('--version') $root
        if ($curlCapture.ExitCode -ne 0 -or -not $curlCapture.Stdout) { throw 'curl.exe self-test failed.' }
        Write-Host 'Rig AI preparer self-test PASS.'
    } finally {
        Remove-Item -LiteralPath $root -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Build-Eco([string]$Work) {
    $sourceZip = Join-Path $Work 'eco-source.zip'
    $goZip = Join-Path $Work 'go.zip'
    Write-Host '[1/7] Downloading exact ECO source from GitHub.'
    Download-CheckedSmall "https://codeload.github.com/ECO-evidence-casework-one/eco/zip/$EcoSource" $sourceZip
    Write-Host '[2/7] Downloading hash-pinned portable Go from GitHub.'
    Download-CheckedSmall $GoURL $goZip $GoZipSHA

    $sourceRoot = Join-Path $Work 'source'
    $compilerRoot = Join-Path $Work 'compiler'
    Expand-CheckedZip $sourceZip $sourceRoot
    Expand-CheckedZip $goZip $compilerRoot
    $sourceRoots = @(Get-ChildItem -LiteralPath $sourceRoot -Directory)
    if ($sourceRoots.Count -ne 1) { throw 'Unexpected ECO source archive layout.' }
    $sourceRoot = $sourceRoots[0].FullName
    $portable = Find-PortableGo $compilerRoot
    $go = $portable.Go

    $envNames = @('GOENV','GOTOOLCHAIN','GOWORK','GOTELEMETRY','GOROOT','GOPATH','GOCACHE','GOMODCACHE','GOTMPDIR','GOOS','GOARCH','GOAMD64','CGO_ENABLED','GOMAXPROCS','GOFLAGS','GOEXPERIMENT','GODEBUG','GOPROXY','GOSUMDB','GOPRIVATE','GONOSUMDB','GONOPROXY','TEMP','TMP')
    $saved = @{}
    foreach ($name in $envNames) { $saved[$name] = [Environment]::GetEnvironmentVariable($name,'Process') }
    $oldLocation = Get-Location
    try {
        $env:GOROOT=$portable.Root; $env:GOENV='off'; $env:GOTOOLCHAIN='local'; $env:GOWORK='off'; $env:GOTELEMETRY='off'
        $env:GOPATH=Join-Path $Work 'cache'; $env:GOCACHE=Join-Path $Work 'cache\build'; $env:GOMODCACHE=Join-Path $Work 'cache\modules'; $env:GOTMPDIR=Join-Path $Work 'tmp'
        [void][IO.Directory]::CreateDirectory($env:GOTMPDIR); $env:TEMP=$env:GOTMPDIR; $env:TMP=$env:GOTMPDIR
        $env:GOOS='windows'; $env:GOARCH='amd64'; $env:GOAMD64='v1'; $env:CGO_ENABLED='0'; $env:GOMAXPROCS='2'; $env:GOFLAGS='-mod=readonly -buildvcs=false'
        $env:GOEXPERIMENT=''; $env:GODEBUG=''; $env:GOPRIVATE=''; $env:GONOSUMDB=''; $env:GONOPROXY=''; $env:GOPROXY='https://proxy.golang.org'; $env:GOSUMDB='sum.golang.org'
        Set-Location -LiteralPath $sourceRoot
        $version = (& $go version | Out-String).Trim()
        if ($version -cne 'go version go1.23.12 windows/amd64') { throw "Wrong Go identity: $version" }
        Write-Host '[3/7] Verifying dependencies.'
        Invoke-Native 'go mod download' $go @('mod','download')
        Invoke-Native 'go mod verify' $go @('mod','verify')
        $env:GOPROXY='off'
        Write-Host '[4/7] Running ECO tests and vet.'
        Invoke-Native 'go test' $go @('test','-count=1','-p=2','./...')
        Invoke-Native 'go vet' $go @('vet','-p=2','./...')
        Write-Host '[5/7] Reproducing the qualified archive-source ECO executable twice.'
        $first = Join-Path $Work 'ECO.exe'
        $second = Join-Path $Work 'ECO.rebuild.exe'
        $ldflags = "-s -w -H windowsgui -buildid= -X github.com/ECO-evidence-casework-one/eco/internal/eco.SourceCommit=$EcoSource"
        Invoke-Native 'ECO build 1' $go @('build','-trimpath','-buildvcs=false','-ldflags',$ldflags,'-o',$first,'./cmd/eco')
        Invoke-Native 'ECO build 2' $go @('build','-trimpath','-buildvcs=false','-ldflags',$ldflags,'-o',$second,'./cmd/eco')
        [void](Assert-SHA256 $first $EcoExeSHA)
        [void](Assert-SHA256 $second $EcoExeSHA)
        if ((Get-Item -LiteralPath $first).Length -ne $EcoExeSize) { throw 'ECO size mismatch.' }
        Remove-Item -LiteralPath $second -Force
        return $first
    } finally {
        Set-Location $oldLocation
        foreach ($name in $envNames) { [Environment]::SetEnvironmentVariable($name,$saved[$name],'Process') }
    }
}

function Install-Llama([string]$Work, [string]$Destination) {
    Write-Host '[6/7] Downloading hash-pinned llama.cpp b10259 from GitHub.'
    $archive = Join-Path $Work 'llama.zip'
    Download-CheckedSmall $LlamaURL $archive $LlamaZipSHA
    $temp = Join-Path $Work 'llama'
    Expand-CheckedZip $archive $temp
    Move-Item -LiteralPath $temp -Destination $Destination
    $llama = Find-Llama $Destination
    $capture = Capture-Native $llama @('--offline','--version') $Work
    $diagnostic = Native-Diagnostic $capture
    if ($capture.ExitCode -ne 0 -or -not $diagnostic) { throw "llama.cpp version probe failed. $diagnostic" }
    $versionLine = @($diagnostic -split "`r?`n" | Where-Object { $_ -match 'version:' } | Select-Object -First 1)
    if ($versionLine.Count -ne 1) { throw "llama.cpp version output was not recognised. $diagnostic" }
    Write-Host $versionLine[0]
    [pscustomobject]@{ Exe=$llama; SHA=(Get-SHA256 $llama); Version=$versionLine[0] }
}

function Probe-Qwen([string]$Llama, [string]$Model, [string]$Work) {
    $prompt = Join-Path $Work 'probe.txt'
    $schema = Join-Path $Work 'schema.json'
    Write-Utf8NoBom $prompt 'Return only the JSON object required by the schema. Set ok to ECO_AI_READY.'
    Write-Utf8NoBom $schema '{"type":"object","additionalProperties":false,"required":["ok"],"properties":{"ok":{"type":"string","enum":["ECO_AI_READY"]}}}'
    $arguments = @('--offline','--model',$Model,'--file',$prompt,'--json-schema-file',$schema,'--simple-io','--no-display-prompt','--color','off','--log-disable','--seed','0','--temp','0','--top-k','1','--top-p','1','--min-p','0','--ctx-size','2048','--n-predict','64','--device','none','--n-gpu-layers','0','--fit','off','--no-context-shift','--no-perf')
    $capture = Capture-Native $Llama $arguments $Work
    $stdout = $capture.Stdout.Trim()
    if ($capture.ExitCode -ne 0) { throw "Qwen smoke test failed: $(Native-Diagnostic $capture)" }
    if (-not $stdout) { throw "Qwen smoke test produced no standard output. $(Native-Diagnostic $capture)" }
    try { $json = $stdout | ConvertFrom-Json } catch { throw "Qwen smoke test returned non-JSON standard output: $stdout" }
    if ($json.ok -cne 'ECO_AI_READY') { throw "Unexpected Qwen result: $stdout" }
}

function Write-Launcher([string]$Root, [string]$Llama, [string]$LlamaSHA) {
    $relative = $Llama.Substring($Root.TrimEnd('\').Length + 1)
    @(
        '@echo off','setlocal','cd /d "%~dp0"','if not exist "PreviewUserData" mkdir "PreviewUserData"','set "LOCALAPPDATA=%~dp0PreviewUserData"',
        "set `"ECO_LLAMA_CPP=%~dp0$relative`"","set `"ECO_LLAMA_CPP_SHA256=$LlamaSHA`"","set `"ECO_LLAMA_MODEL=%~dp0AIAssets\$ModelName`"","set `"ECO_LLAMA_MODEL_SHA256=$ModelSHA`"",'start "" "%~dp0ECO.exe"','endlocal'
    ) | Set-Content -LiteralPath (Join-Path $Root 'START_ECO_WITH_AI.cmd') -Encoding ascii
}

if ($SelfTest) {
    Invoke-PreparerSelfTest
    if (-not $QualificationOnly) { exit 0 }
}
if (-not [Environment]::Is64BitOperatingSystem) { throw '64-bit Windows is required.' }
if (-not $OutputRoot) { $OutputRoot = Join-Path $PSScriptRoot 'ECO_RIG_AI_PREVIEW' }
$OutputRoot = [IO.Path]::GetFullPath($OutputRoot)

$partialToCarry = ''
if (Test-Path -LiteralPath $OutputRoot) {
    $resultPath = Join-Path $OutputRoot 'AI_SETUP_RESULT.txt'
    $firstLine = ''
    if (Test-Path -LiteralPath $resultPath -PathType Leaf) { $firstLine = [string](Get-Content -LiteralPath $resultPath -TotalCount 1 -ErrorAction SilentlyContinue) }
    $completedLauncher = Test-Path -LiteralPath (Join-Path $OutputRoot 'START_ECO_WITH_AI.cmd') -PathType Leaf
    if ($firstLine -ne 'ECO RIG AI SETUP STOPPED' -and ($firstLine -or $completedLauncher)) {
        throw "Output exists and is not a failed/interrupted setup: $OutputRoot"
    }
    $archive = $OutputRoot + '.interrupted-' + (Get-Date -Format 'yyyyMMdd-HHmmss')
    Move-Item -LiteralPath $OutputRoot -Destination $archive
    Write-Host "Preserved previous interrupted attempt at: $archive"
    $oldPartial = Join-Path $archive ('AIAssets\' + $ModelName + '.part')
    if (Test-Path -LiteralPath $oldPartial -PathType Leaf) { $partialToCarry = $oldPartial }
}

[void][IO.Directory]::CreateDirectory($OutputRoot)
$work = Join-Path $OutputRoot '.work'
$assets = Join-Path $OutputRoot 'AIAssets'
[void][IO.Directory]::CreateDirectory($work)
[void][IO.Directory]::CreateDirectory($assets)
$result = Join-Path $OutputRoot 'AI_SETUP_RESULT.txt'
$model = Join-Path $assets $ModelName

if ($partialToCarry) {
    $newPartial = $model + '.part'
    Move-Item -LiteralPath $partialToCarry -Destination $newPartial
    Write-Host "Recovered $(Format-Bytes ((Get-Item -LiteralPath $newPartial).Length)) of the previous Qwen partial download."
}

try {
    Hardware-Receipt (Join-Path $OutputRoot 'RIG_AI_HARDWARE.json')
    $built = Build-Eco $work
    Copy-Item -LiteralPath $built -Destination (Join-Path $OutputRoot 'ECO.exe')
    [void](Assert-SHA256 (Join-Path $OutputRoot 'ECO.exe') $EcoExeSHA)
    $llama = Install-Llama $work (Join-Path $OutputRoot 'Runtime')
    Probe-QwenEndpoint $work

    if ($QualificationOnly) {
        @('RIG AI QUALIFICATION PASS',"ECO SHA-256: $EcoExeSHA","llama-cli SHA-256: $($llama.SHA)",'Official Qwen bounded endpoint probe: PASS','Windows PowerShell 5.1 + curl.exe path: PASS') | Set-Content -LiteralPath $result -Encoding utf8
        Get-Content -LiteralPath $result | Out-Host
        exit 0
    }

    Write-Host '[7/7] Downloading official Qwen2.5 1.5B Q4_K_M with visible resumable progress.'
    Download-QwenModel $model
    Probe-Qwen $llama.Exe $model $work
    Write-Launcher $OutputRoot $llama.Exe $llama.SHA
    @('ECO RIG AI SETUP PASS','Real offline Qwen generation: PASS',"Qwen SHA-256: $ModelSHA",'Use START_ECO_WITH_AI.cmd to reopen. Synthetic/test material only.') | Set-Content -LiteralPath $result -Encoding utf8
    Remove-Item -LiteralPath $work -Recurse -Force -ErrorAction SilentlyContinue
    Get-Content -LiteralPath $result | Out-Host
    if (-not $NoLaunch) { Start-Process -FilePath (Join-Path $OutputRoot 'START_ECO_WITH_AI.cmd') -WorkingDirectory $OutputRoot }
} catch {
    $message = $_.Exception.Message
    @('ECO RIG AI SETUP STOPPED',$message,'Any Qwen .part file is preserved for resume. Do not weaken Windows security.') | Set-Content -LiteralPath $result -Encoding utf8
    Write-Host $message
    exit 1
}
