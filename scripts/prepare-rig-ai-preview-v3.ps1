# ECO rig AI preview preparer v3.
# GitHub-first, hash-pinned, private synthetic-data preview.
# Windows PowerShell 5.1 compatible: native stdout/stderr are captured explicitly
# so expected llama.cpp stderr output cannot be promoted to a terminating error.
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
$LlamaCommit = '1269cb1ff1598751f846241be90083ae9ad036fb'
$ModelName = 'qwen2.5-1.5b-instruct-q4_k_m.gguf'
$ModelURL = 'https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf?download=true'
$ModelSHA = '6a1a2eb6d15622bf3c96857206351ba97e1af16c30d7a74ee38970e434e9407e'

function Sha([string]$Path) {
    (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLowerInvariant()
}

function Assert-Sha([string]$Path,[string]$Expected) {
    $actual = Sha $Path
    if ($actual -cne $Expected) {
        throw "SHA-256 mismatch for $([IO.Path]::GetFileName($Path)): $actual"
    }
    $actual
}

function Download-Checked([string]$Url,[string]$Path,[string]$Expected='') {
    if ($Url -notmatch '^https://') { throw 'Only HTTPS downloads are allowed.' }
    if (Test-Path -LiteralPath $Path) { throw "Will not overwrite $Path" }
    [void][IO.Directory]::CreateDirectory((Split-Path -Parent $Path))
    $part = $Path + '.part'
    $old = $ProgressPreference
    $ProgressPreference = 'SilentlyContinue'
    try {
        Invoke-WebRequest -UseBasicParsing -Uri $Url -OutFile $part -TimeoutSec 7200
        if ($Expected) { [void](Assert-Sha $part $Expected) }
        Move-Item -LiteralPath $part -Destination $Path
    } finally {
        $ProgressPreference = $old
        Remove-Item -LiteralPath $part -Force -ErrorAction SilentlyContinue
    }
}

function Expand-Checked([string]$Zip,[string]$Dest) {
    if (Test-Path -LiteralPath $Dest) { throw "Will not overwrite extraction folder $Dest" }
    Expand-Archive -LiteralPath $Zip -DestinationPath $Dest
}

function Run-Native([string]$Label,[string]$Exe,[string[]]$Arguments) {
    & $Exe @Arguments | Out-Host
    if ($LASTEXITCODE -ne 0) { throw "$Label failed with exit code $LASTEXITCODE" }
}

function Capture-Native([string]$Exe,[string[]]$Arguments,[string]$Work) {
    [void][IO.Directory]::CreateDirectory($Work)
    $token = [guid]::NewGuid().ToString('N')
    $stdoutPath = Join-Path $Work ("native-$token.stdout.txt")
    $stderrPath = Join-Path $Work ("native-$token.stderr.txt")
    $oldAction = $ErrorActionPreference
    try {
        # Windows PowerShell 5.1 represents native stderr as its error stream.
        # Continue locally while redirecting both streams to files, then judge
        # success only from the native exit code and the captured bytes/text.
        $ErrorActionPreference = 'Continue'
        & $Exe @Arguments 1> $stdoutPath 2> $stderrPath
        $exitCode = $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $oldAction
    }
    $stdout = ''
    $stderr = ''
    if (Test-Path -LiteralPath $stdoutPath) { $stdout = Get-Content -LiteralPath $stdoutPath -Raw -ErrorAction SilentlyContinue }
    if (Test-Path -LiteralPath $stderrPath) { $stderr = Get-Content -LiteralPath $stderrPath -Raw -ErrorAction SilentlyContinue }
    Remove-Item -LiteralPath $stdoutPath,$stderrPath -Force -ErrorAction SilentlyContinue
    [pscustomobject]@{ ExitCode=[int]$exitCode; Stdout=[string]$stdout; Stderr=[string]$stderr }
}

function Native-Diagnostic([object]$Capture) {
    (($Capture.Stdout + "`n" + $Capture.Stderr).Trim())
}

function Find-PortableGo([string]$Root) {
    $matches = @(Get-ChildItem -LiteralPath $Root -Filter 'go.exe' -File -Recurse | Where-Object { $_.Directory.Name -ieq 'bin' })
    if ($matches.Count -ne 1) { throw "Expected exactly one bin\go.exe in the verified Go archive, found $($matches.Count)." }
    $go = $matches[0].FullName
    $goroot = Split-Path -Parent (Split-Path -Parent $go)
    if (-not (Test-Path -LiteralPath (Join-Path $goroot 'src') -PathType Container)) { throw 'Verified Go archive has an unexpected GOROOT layout.' }
    [pscustomobject]@{ Go=$go; Root=$goroot }
}

function Find-Llama([string]$Root) {
    $matches = @(Get-ChildItem -LiteralPath $Root -Filter 'llama-cli.exe' -File -Recurse)
    if ($matches.Count -ne 1) { throw "Expected exactly one llama-cli.exe, found $($matches.Count)." }
    $matches[0].FullName
}

function Hardware-Receipt([string]$Path) {
    $r=[ordered]@{captured_utc=[DateTimeOffset]::UtcNow.ToString('o');os_64_bit=[Environment]::Is64BitOperatingSystem;os=[Environment]::OSVersion.VersionString;cpu=@();gpu=@();memory_bytes=$null}
    try{$r.cpu=@(Get-CimInstance Win32_Processor|%{[ordered]@{name=$_.Name;cores=$_.NumberOfCores;logical=$_.NumberOfLogicalProcessors;max_mhz=$_.MaxClockSpeed}})}catch{}
    try{$r.gpu=@(Get-CimInstance Win32_VideoController|%{[ordered]@{name=$_.Name;adapter_ram=$_.AdapterRAM;driver=$_.DriverVersion}})}catch{}
    try{$r.memory_bytes=[int64](Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory}catch{}
    $r|ConvertTo-Json -Depth 8|Set-Content -LiteralPath $Path -Encoding UTF8
}

function Write-Utf8NoBom([string]$Path,[string]$Text) {
    $enc = New-Object System.Text.UTF8Encoding($false)
    [IO.File]::WriteAllText($Path,$Text,$enc)
}

function Self-Test {
    $root=Join-Path ([IO.Path]::GetTempPath()) ('eco-ai-v3-'+[guid]::NewGuid().ToString('N'))
    [void][IO.Directory]::CreateDirectory($root)
    try {
        $f=Join-Path $root 'x'
        Set-Content -LiteralPath $f -Value 'x' -NoNewline -Encoding ascii
        $h=Sha $f
        [void](Assert-Sha $f $h)
        $rejected=$false
        try{[void](Assert-Sha $f ('0'*64))}catch{$rejected=$true}
        if(-not $rejected){throw 'Hash rejection self-test failed.'}

        # Reproduce the user-visible failure mode: a successful native process
        # writes its version to stderr. This must remain a successful capture
        # even under Windows PowerShell 5.1 with ErrorActionPreference=Stop.
        $capture = Capture-Native $env:ComSpec @('/d','/c','echo version: ECO_CAPTURE_TEST 1>&2 & exit /b 0') $root
        if ($capture.ExitCode -ne 0 -or $capture.Stderr -notmatch 'ECO_CAPTURE_TEST') {
            throw 'Native stderr capture self-test failed.'
        }
        Write-Host 'Rig AI preparer v3 self-test PASS.'
    } finally {
        Remove-Item -LiteralPath $root -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Build-Eco([string]$Work) {
    $sourceZip=Join-Path $Work 'eco-source.zip'
    $goZip=Join-Path $Work 'go.zip'
    Write-Host '[1/7] Downloading exact ECO source from GitHub.'
    Download-Checked "https://codeload.github.com/ECO-evidence-casework-one/eco/zip/$EcoSource" $sourceZip
    Write-Host '[2/7] Downloading hash-pinned portable Go from GitHub.'
    Download-Checked $GoURL $goZip $GoZipSHA
    $src=Join-Path $Work 'source'
    $compiler=Join-Path $Work 'compiler'
    Expand-Checked $sourceZip $src
    Expand-Checked $goZip $compiler
    $sourceRoots=@(Get-ChildItem -LiteralPath $src -Directory)
    if($sourceRoots.Count-ne 1){throw 'Unexpected ECO source archive layout.'}
    $src=$sourceRoots[0].FullName
    $portable=Find-PortableGo $compiler
    $go=$portable.Go
    $names=@('GOENV','GOTOOLCHAIN','GOWORK','GOTELEMETRY','GOROOT','GOPATH','GOCACHE','GOMODCACHE','GOTMPDIR','GOOS','GOARCH','GOAMD64','CGO_ENABLED','GOMAXPROCS','GOFLAGS','GOEXPERIMENT','GODEBUG','GOPROXY','GOSUMDB','GOPRIVATE','GONOSUMDB','GONOPROXY','TEMP','TMP')
    $saved=@{}
    foreach($n in $names){$saved[$n]=[Environment]::GetEnvironmentVariable($n,'Process')}
    $old=Get-Location
    try {
        $env:GOROOT=$portable.Root
        $env:GOENV='off';$env:GOTOOLCHAIN='local';$env:GOWORK='off';$env:GOTELEMETRY='off'
        $env:GOPATH=Join-Path $Work 'cache';$env:GOCACHE=Join-Path $Work 'cache\build';$env:GOMODCACHE=Join-Path $Work 'cache\modules';$env:GOTMPDIR=Join-Path $Work 'tmp'
        [void][IO.Directory]::CreateDirectory($env:GOTMPDIR)
        $env:TEMP=$env:GOTMPDIR;$env:TMP=$env:GOTMPDIR
        $env:GOOS='windows';$env:GOARCH='amd64';$env:GOAMD64='v1';$env:CGO_ENABLED='0';$env:GOMAXPROCS='2'
        $env:GOFLAGS='-mod=readonly -buildvcs=false';$env:GOEXPERIMENT='';$env:GODEBUG='';$env:GOPRIVATE='';$env:GONOSUMDB='';$env:GONOPROXY='';$env:GOPROXY='https://proxy.golang.org';$env:GOSUMDB='sum.golang.org'
        Set-Location -LiteralPath $src
        $version=(& $go version|Out-String).Trim()
        if($version-cne 'go version go1.23.12 windows/amd64'){throw "Wrong Go identity: $version"}
        Write-Host '[3/7] Verifying dependencies.'
        Run-Native 'go mod download' $go @('mod','download')
        Run-Native 'go mod verify' $go @('mod','verify')
        $env:GOPROXY='off'
        Write-Host '[4/7] Running ECO tests and vet.'
        Run-Native 'go test' $go @('test','-count=1','-p=2','./...')
        Run-Native 'go vet' $go @('vet','-p=2','./...')
        Write-Host '[5/7] Reproducing the qualified archive-source ECO executable twice.'
        $a=Join-Path $Work 'ECO.exe'
        $b=Join-Path $Work 'ECO.rebuild.exe'
        $flags="-s -w -H windowsgui -buildid= -X github.com/ECO-evidence-casework-one/eco/internal/eco.SourceCommit=$EcoSource"
        Run-Native 'ECO build 1' $go @('build','-trimpath','-buildvcs=false','-ldflags',$flags,'-o',$a,'./cmd/eco')
        Run-Native 'ECO build 2' $go @('build','-trimpath','-buildvcs=false','-ldflags',$flags,'-o',$b,'./cmd/eco')
        $observedSha=Sha $a
        $observedSize=(Get-Item -LiteralPath $a).Length
        Write-Host "Archive-source candidate observed SHA-256: $observedSha"
        Write-Host "Archive-source candidate observed size: $observedSize bytes"
        [void](Assert-Sha $a $EcoExeSHA)
        [void](Assert-Sha $b $EcoExeSHA)
        if($observedSize-ne $EcoExeSize){throw "ECO size mismatch: $observedSize"}
        Remove-Item $b -Force
        [pscustomobject]@{Exe=$a;GoVersion=$version;SourceZipSHA=(Sha $sourceZip);Source=$src}
    } finally {
        Set-Location $old
        foreach($n in $names){[Environment]::SetEnvironmentVariable($n,$saved[$n],'Process')}
    }
}

function Install-Llama([string]$Work,[string]$Dest) {
    Write-Host '[6/7] Downloading hash-pinned llama.cpp b10259 from GitHub.'
    $zip=Join-Path $Work 'llama.zip'
    Download-Checked $LlamaURL $zip $LlamaZipSHA
    $tmp=Join-Path $Work 'llama'
    Expand-Checked $zip $tmp
    $llama=Find-Llama $tmp
    Move-Item -LiteralPath $tmp -Destination $Dest
    $llama=Find-Llama $Dest
    $capture=Capture-Native $llama @('--offline','--version') $Work
    $diagnostic=Native-Diagnostic $capture
    if($capture.ExitCode-ne 0-or-not $diagnostic){throw "llama.cpp version probe failed. $diagnostic"}
    $versionLine=@($diagnostic -split "`r?`n" | Where-Object { $_ -match 'version:' } | Select-Object -First 1)
    if($versionLine.Count-ne 1){throw "llama.cpp version output was not recognised. $diagnostic"}
    Write-Host $versionLine[0]
    [pscustomobject]@{Exe=$llama;SHA=(Sha $llama);Version=$versionLine[0]}
}

function Get-Model([string]$Dest) {
    [void][IO.Directory]::CreateDirectory($Dest)
    $p=Join-Path $Dest $ModelName
    if(Test-Path $p){[void](Assert-Sha $p $ModelSHA);return $p}
    Write-Host '[7/7] Downloading official Qwen2.5 1.5B Q4_K_M and verifying its published SHA-256.'
    Download-Checked $ModelURL $p $ModelSHA
    $p
}

function Probe-Qwen([string]$Llama,[string]$Model,[string]$Work) {
    $p=Join-Path $Work 'probe.txt'
    $s=Join-Path $Work 'schema.json'
    Write-Utf8NoBom $p 'Return only the JSON object required by the schema. Set ok to ECO_AI_READY.'
    Write-Utf8NoBom $s '{"type":"object","additionalProperties":false,"required":["ok"],"properties":{"ok":{"type":"string","enum":["ECO_AI_READY"]}}}'
    $arguments=@('--offline','--model',$Model,'--file',$p,'--json-schema-file',$s,'--simple-io','--no-display-prompt','--color','off','--log-disable','--seed','0','--temp','0','--top-k','1','--top-p','1','--min-p','0','--ctx-size','2048','--n-predict','64','--device','none','--n-gpu-layers','0','--fit','off','--no-context-shift','--no-perf')
    $capture=Capture-Native $Llama $arguments $Work
    $stdout=$capture.Stdout.Trim()
    $diagnostic=Native-Diagnostic $capture
    if($capture.ExitCode-ne 0){throw "Qwen smoke test failed with exit code $($capture.ExitCode): $diagnostic"}
    if(-not $stdout){throw "Qwen smoke test produced no standard output. $diagnostic"}
    try{$j=$stdout|ConvertFrom-Json}catch{throw "Qwen smoke test returned non-JSON standard output: $stdout"}
    if($j.ok-cne 'ECO_AI_READY'){throw "Unexpected Qwen result: $stdout"}
    $stdout
}

function Write-Launcher([string]$Root,[string]$Llama,[string]$LlamaSHA) {
    $rel=$Llama.Substring($Root.TrimEnd('\').Length+1)
    @('@echo off','setlocal','cd /d "%~dp0"','if not exist "PreviewUserData" mkdir "PreviewUserData"','set "LOCALAPPDATA=%~dp0PreviewUserData"',"set `"ECO_LLAMA_CPP=%~dp0$rel`"","set `"ECO_LLAMA_CPP_SHA256=$LlamaSHA`"","set `"ECO_LLAMA_MODEL=%~dp0AIAssets\$ModelName`"","set `"ECO_LLAMA_MODEL_SHA256=$ModelSHA`"",'start "" "%~dp0ECO.exe"','endlocal') | Set-Content (Join-Path $Root 'START_ECO_WITH_AI.cmd') -Encoding ascii
}

if($SelfTest){Self-Test;if(-not $QualificationOnly){exit 0}}
if(-not [Environment]::Is64BitOperatingSystem){throw '64-bit Windows is required.'}
if(-not $OutputRoot){$OutputRoot=Join-Path $PSScriptRoot 'ECO_RIG_AI_PREVIEW'}
$OutputRoot=[IO.Path]::GetFullPath($OutputRoot)
if(Test-Path $OutputRoot){
    $existingResult=Join-Path $OutputRoot 'AI_SETUP_RESULT.txt'
    $failed=$false
    if(Test-Path -LiteralPath $existingResult -PathType Leaf){
        $first=Get-Content -LiteralPath $existingResult -TotalCount 1 -ErrorAction SilentlyContinue
        if($first -eq 'ECO RIG AI SETUP STOPPED'){$failed=$true}
    }
    if(-not $failed){throw "Output already exists and is not a known failed setup: $OutputRoot"}
    $archive=$OutputRoot+'.failed-'+(Get-Date -Format 'yyyyMMdd-HHmmss')
    Move-Item -LiteralPath $OutputRoot -Destination $archive
    Write-Host "Preserved previous failed attempt at: $archive"
}
[void][IO.Directory]::CreateDirectory($OutputRoot)
$work=Join-Path $OutputRoot '.work'
[void][IO.Directory]::CreateDirectory($work)
$result=Join-Path $OutputRoot 'AI_SETUP_RESULT.txt'
try {
    Hardware-Receipt (Join-Path $OutputRoot 'RIG_AI_HARDWARE.json')
    $build=Build-Eco $work
    Copy-Item $build.Exe (Join-Path $OutputRoot 'ECO.exe')
    [void](Assert-Sha (Join-Path $OutputRoot 'ECO.exe') $EcoExeSHA)
    $llama=Install-Llama $work (Join-Path $OutputRoot 'Runtime')
    if($QualificationOnly){
        @('RIG AI QUALIFICATION PASS',"ECO archive-source SHA-256: $EcoExeSHA","ECO source commit: $EcoSource","llama archive SHA-256: $LlamaZipSHA","llama-cli SHA-256: $($llama.SHA)","llama version: $($llama.Version)",'Windows PowerShell native stderr capture: PASS','Official Qwen model is intentionally not downloaded in CI qualification.')|Set-Content $result -Encoding utf8
        Get-Content $result|Write-Host
        exit 0
    }
    $model=Get-Model (Join-Path $OutputRoot 'AIAssets')
    $probe=Probe-Qwen $llama.Exe $model $work
    Write-Launcher $OutputRoot $llama.Exe $llama.SHA
    @('ECO RIG AI SETUP PASS','Real offline Qwen generation: PASS',"ECO source commit: $EcoSource","ECO archive-source SHA-256: $EcoExeSHA","Qwen SHA-256: $ModelSHA","llama-cli SHA-256: $($llama.SHA)",'Use START_ECO_WITH_AI.cmd to reopen. Synthetic/test material only.')|Set-Content $result -Encoding utf8
    @('ECO RIG AI PREVIEW','','Double-click START_ECO_WITH_AI.cmd.','This preview uses its own PreviewUserData folder.','Qwen and llama.cpp are local; ECO has no cloud AI fallback.','Synthetic/test material only.')|Set-Content (Join-Path $OutputRoot 'README_FIRST.txt') -Encoding utf8
    Remove-Item $work -Recurse -Force -ErrorAction SilentlyContinue
    Get-Content $result|Write-Host
    if(-not $NoLaunch){Start-Process (Join-Path $OutputRoot 'START_ECO_WITH_AI.cmd') -WorkingDirectory $OutputRoot}
}catch{
    $m=$_.Exception.Message
    if($env:USERPROFILE){$m=$m.Replace($env:USERPROFILE,'[user-profile]')}
    @('ECO RIG AI SETUP STOPPED',$m,'Do not weaken Windows security. Send this file back to the ECO development chat.')|Set-Content $result -Encoding utf8
    Write-Host $m
    exit 1
}
