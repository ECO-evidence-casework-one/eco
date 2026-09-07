# ECO rig AI preview preparer v3.
# GitHub-first, hash-pinned, Windows PowerShell 5.1 compatible.
# Large Qwen downloads are resumable, visible, retried, stall-detected, then SHA-256 verified.
[CmdletBinding()]
param([string]$OutputRoot='', [switch]$NoLaunch, [switch]$SelfTest, [switch]$QualificationOnly)
$ErrorActionPreference='Stop'; Set-StrictMode -Version Latest

$EcoSource='b03ec2358dbf437deb922ad1cbb96d4e5c6faedb'
$EcoExeSHA='eb6159cb0406a0d1b7195285f03848048026726e6abe3c73b9f7a4f52a9fbee3'
$EcoExeSize=4892672
$GoURL='https://github.com/actions/go-versions/releases/download/1.23.12-16792118003/go-1.23.12-win32-x64.zip'
$GoZipSHA='c27b02f15d4ceb89fbce6ffe2a28df3dd293608cf79e9f12839f672863622845'
$LlamaURL='https://github.com/ggml-org/llama.cpp/releases/download/b10259/llama-b10259-bin-win-cpu-x64.zip'
$LlamaZipSHA='6613d8d56263233ef800fb8f8135231adb5eb851b40558281190242b0b20556b'
$ModelName='qwen2.5-1.5b-instruct-q4_k_m.gguf'
$ModelURL='https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf?download=true'
$ModelSHA='6a1a2eb6d15622bf3c96857206351ba97e1af16c30d7a74ee38970e434e9407e'
$ModelApproxBytes=[int64]1120000000

function Sha([string]$p){(Get-FileHash -LiteralPath $p -Algorithm SHA256).Hash.ToLowerInvariant()}
function Assert-Sha([string]$p,[string]$want){$got=Sha $p;if($got-cne$want){throw "SHA-256 mismatch for $([IO.Path]::GetFileName($p)): $got"};$got}
function Fmt([int64]$b){if($b-ge1GB){'{0:N2} GB'-f($b/1GB)}elseif($b-ge1MB){'{0:N1} MB'-f($b/1MB)}elseif($b-ge1KB){'{0:N1} KB'-f($b/1KB)}else{"$b bytes"}}
function Expand-Checked([string]$z,[string]$d){if(Test-Path $d){throw "Will not overwrite extraction folder $d"};Expand-Archive -LiteralPath $z -DestinationPath $d}
function Run([string]$label,[string]$exe,[string[]]$arguments){& $exe @arguments|Out-Host;if($LASTEXITCODE-ne0){throw "$label failed with exit code $LASTEXITCODE"}}
function Download-Small([string]$u,[string]$p,[string]$want=''){
 if($u-notmatch'^https://'){throw'Only HTTPS downloads are allowed.'};if(Test-Path $p){throw"Will not overwrite $p"}
 [void][IO.Directory]::CreateDirectory((Split-Path -Parent $p));$part=$p+'.part';$old=$ProgressPreference;$ProgressPreference='SilentlyContinue'
 try{Invoke-WebRequest -UseBasicParsing -Uri $u -OutFile $part -TimeoutSec 7200;if($want){[void](Assert-Sha $part $want)};Move-Item $part $p}
 finally{$ProgressPreference=$old;Remove-Item $part -Force -ErrorAction SilentlyContinue}
}
function Capture([string]$exe,[string[]]$arguments,[string]$work){
 [void][IO.Directory]::CreateDirectory($work);$t=[guid]::NewGuid().ToString('N');$o=Join-Path $work "$t.out";$e=Join-Path $work "$t.err";$old=$ErrorActionPreference
 try{$ErrorActionPreference='Continue';& $exe @arguments 1>$o 2>$e;$code=$LASTEXITCODE}finally{$ErrorActionPreference=$old}
 $stdout='';$stderr='';if(Test-Path $o){$stdout=Get-Content $o -Raw -ErrorAction SilentlyContinue};if(Test-Path $e){$stderr=Get-Content $e -Raw -ErrorAction SilentlyContinue}
 Remove-Item $o,$e -Force -ErrorAction SilentlyContinue;[pscustomobject]@{ExitCode=[int]$code;Stdout=[string]$stdout;Stderr=[string]$stderr}
}
function Diag($c){($c.Stdout+"`n"+$c.Stderr).Trim()}
function Curl{ $c=Get-Command curl.exe -ErrorAction SilentlyContinue;if($null-eq$c-or-not$c.Source){throw'Windows curl.exe was not found.'};$c.Source }
function ArgLine([string[]]$a){(($a|ForEach-Object{'"'+([string]$_).Replace('"','\"')+'"'})-join' ')}

function Endpoint([string]$work){
 $curl=Curl;$probe=Join-Path $work ('qwen-probe-'+[guid]::NewGuid().ToString('N'));$headers=$probe+'.headers'
 try{
  $c=Capture $curl @('--location','--fail','--silent','--show-error','--connect-timeout','30','--max-time','90','--range','0-0','--max-filesize','1048576','--dump-header',$headers,'--output',$probe,$ModelURL) $work
  if($c.ExitCode-ne0){throw "Qwen endpoint probe failed: $(Diag $c)"}
  if(-not(Test-Path $probe)){throw'Qwen endpoint returned no probe file.'};$n=(Get-Item $probe).Length;if($n-lt1-or$n-gt1048576){throw"Qwen endpoint ignored the bounded range request ($n bytes)."}
  [int64]$total=0;$h=Get-Content $headers -Raw -ErrorAction SilentlyContinue
  $m=[regex]::Matches($h,'(?im)^content-range:\s*bytes\s+\d+-\d+/(\d+)\s*$');if($m.Count){$total=[int64]$m[$m.Count-1].Groups[1].Value}
  if($total-le0){$m=[regex]::Matches($h,'(?im)^x-linked-size:\s*(\d+)\s*$');if($m.Count){$total=[int64]$m[$m.Count-1].Groups[1].Value}}
  if($total-le0){$total=$ModelApproxBytes};Write-Host("Qwen endpoint range/resume PASS · "+(Fmt $total));$total
 }finally{Remove-Item $probe,$headers -Force -ErrorAction SilentlyContinue}
}

function Download-Model([string]$path,[string]$work,[int64]$total){
 if(Test-Path $path){[void](Assert-Sha $path $ModelSHA);return}
 $curl=Curl;$part=$path+'.part';$max=4;if($total-le0){$total=$ModelApproxBytes}
 for($attempt=1;$attempt-le$max;$attempt++){
  [int64]$before=0;if(Test-Path $part){$before=(Get-Item $part).Length}
  if($before){Write-Host("Qwen attempt $attempt/$max · resuming from "+(Fmt $before))}else{Write-Host"Qwen attempt $attempt/$max · starting."}
  $ca=@('--location','--fail','--show-error','--connect-timeout','30','--retry','2','--retry-delay','5','--output',$part);if($before){$ca+=@('--continue-at','-')};$ca+=@($ModelURL)
  $p=Start-Process -FilePath $curl -ArgumentList (ArgLine $ca) -NoNewWindow -PassThru
  [int64]$last=$before;[int64]$sample=$before;$changed=Get-Date;$reported=Get-Date;$sampleAt=Get-Date;$stalled=$false
  while(-not$p.HasExited){
   Start-Sleep 5;[int64]$size=0;if(Test-Path $part){$size=(Get-Item $part).Length};$now=Get-Date
   if($size-gt$last){$last=$size;$changed=$now}
   if(($now-$reported).TotalSeconds-ge15){$secs=[math]::Max(1,($now-$sampleAt).TotalSeconds);$speed=[math]::Max(0,($size-$sample)/$secs);$pct=[math]::Min(100,100*$size/$total);Write-Host('Qwen: {0} / {1} ({2:N1}%) · {3:N2} MB/s'-f(Fmt $size),(Fmt $total),$pct,($speed/1MB));$reported=$now;$sampleAt=$now;$sample=$size}
   if(($now-$changed).TotalSeconds-ge180){Write-Host'No Qwen download progress for 3 minutes; retrying without deleting partial bytes.';Stop-Process $p.Id -Force -ErrorAction SilentlyContinue;$stalled=$true;break}
  }
  try{$p.WaitForExit()}catch{};$code=-1;try{$code=$p.ExitCode}catch{}
  if(-not$stalled-and$code-eq0){break}
  if($attempt-eq$max){[int64]$kept=0;if(Test-Path $part){$kept=(Get-Item $part).Length};throw"Qwen download failed after $max attempts. Partial bytes kept: $part ($(Fmt $kept))."}
  Write-Host"curl exit $code; retrying in 10 seconds. Partial bytes are preserved.";Start-Sleep 10
 }
 if(-not(Test-Path $part)){throw'Qwen transfer reported success but no partial file exists.'}
 Write-Host'Qwen transfer finished. Verifying published SHA-256...';$got=Sha $part
 if($got-cne$ModelSHA){$bad=$part+'.bad-sha-'+(Get-Date -Format'yyyyMMdd-HHmmss');Move-Item $part $bad;throw"Qwen SHA-256 mismatch; untrusted bytes preserved at $bad."}
 Move-Item $part $path;Write-Host'Qwen SHA-256 PASS.'
}

function Find-Go([string]$r){$m=@(Get-ChildItem $r -Filter go.exe -File -Recurse|Where-Object{$_.Directory.Name-ieq'bin'});if($m.Count-ne1){throw"Expected one bin\go.exe, found $($m.Count)."};$g=$m[0].FullName;$root=Split-Path -Parent(Split-Path -Parent $g);[pscustomobject]@{Go=$g;Root=$root}}
function Find-Llama([string]$r){$m=@(Get-ChildItem $r -Filter llama-cli.exe -File -Recurse);if($m.Count-ne1){throw"Expected one llama-cli.exe, found $($m.Count)."};$m[0].FullName}
function Hardware([string]$p){$r=[ordered]@{captured_utc=[DateTimeOffset]::UtcNow.ToString('o');os=[Environment]::OSVersion.VersionString;cpu=@();gpu=@();memory_bytes=$null};try{$r.cpu=@(Get-CimInstance Win32_Processor|%{[ordered]@{name=$_.Name;cores=$_.NumberOfCores;logical=$_.NumberOfLogicalProcessors}})}catch{};try{$r.gpu=@(Get-CimInstance Win32_VideoController|%{[ordered]@{name=$_.Name;driver=$_.DriverVersion}})}catch{};try{$r.memory_bytes=[int64](Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory}catch{};$r|ConvertTo-Json -Depth 6|Set-Content $p -Encoding UTF8}
function Utf8([string]$p,[string]$t){[IO.File]::WriteAllText($p,$t,(New-Object Text.UTF8Encoding($false)))}

function SelfTest{
 $r=Join-Path([IO.Path]::GetTempPath())('eco-ai-'+[guid]::NewGuid().ToString('N'));[void][IO.Directory]::CreateDirectory($r)
 try{$f=Join-Path $r x;Set-Content $f x -NoNewline -Encoding ascii;$h=Sha $f;[void](Assert-Sha $f $h);$c=Capture $env:ComSpec @('/d','/c','echo version: ECO_CAPTURE_TEST 1>&2 & exit /b 0') $r;if($c.ExitCode-ne0-or$c.Stderr-notmatch'ECO_CAPTURE_TEST'){throw'native stderr capture failed'};$cc=Capture (Curl) @('--version') $r;if($cc.ExitCode-ne0){throw'curl self-test failed'};Write-Host'Rig AI preparer self-test PASS.'}
 finally{Remove-Item $r -Recurse -Force -ErrorAction SilentlyContinue}
}

function BuildEco([string]$w){
 $sz=Join-Path $w source.zip;$gz=Join-Path $w go.zip;Write-Host'[1/7] Downloading exact ECO source from GitHub.';Download-Small "https://codeload.github.com/ECO-evidence-casework-one/eco/zip/$EcoSource" $sz
 Write-Host'[2/7] Downloading hash-pinned portable Go from GitHub.';Download-Small $GoURL $gz $GoZipSHA;$sd=Join-Path $w source;$gd=Join-Path $w compiler;Expand-Checked $sz $sd;Expand-Checked $gz $gd;$roots=@(Get-ChildItem $sd -Directory);if($roots.Count-ne1){throw'Unexpected ECO source archive layout.'};$src=$roots[0].FullName;$pg=Find-Go $gd;$go=$pg.Go
 $names=@('GOENV','GOTOOLCHAIN','GOWORK','GOTELEMETRY','GOROOT','GOPATH','GOCACHE','GOMODCACHE','GOTMPDIR','GOOS','GOARCH','GOAMD64','CGO_ENABLED','GOMAXPROCS','GOFLAGS','GOPROXY','GOSUMDB','TEMP','TMP');$saved=@{};foreach($n in $names){$saved[$n]=[Environment]::GetEnvironmentVariable($n,'Process')};$old=Get-Location
 try{$env:GOROOT=$pg.Root;$env:GOENV='off';$env:GOTOOLCHAIN='local';$env:GOWORK='off';$env:GOTELEMETRY='off';$env:GOPATH=Join-Path $w cache;$env:GOCACHE=Join-Path $w 'cache\build';$env:GOMODCACHE=Join-Path $w 'cache\modules';$env:GOTMPDIR=Join-Path $w tmp;[void][IO.Directory]::CreateDirectory($env:GOTMPDIR);$env:TEMP=$env:GOTMPDIR;$env:TMP=$env:GOTMPDIR;$env:GOOS='windows';$env:GOARCH='amd64';$env:GOAMD64='v1';$env:CGO_ENABLED='0';$env:GOMAXPROCS='2';$env:GOFLAGS='-mod=readonly -buildvcs=false';$env:GOPROXY='https://proxy.golang.org';$env:GOSUMDB='sum.golang.org';Set-Location $src
  $v=(& $go version|Out-String).Trim();if($v-cne'go version go1.23.12 windows/amd64'){throw"Wrong Go identity: $v"};Write-Host'[3/7] Verifying dependencies.';Run mod $go @('mod','download');Run verify $go @('mod','verify');$env:GOPROXY='off';Write-Host'[4/7] Running ECO tests and vet.';Run test $go @('test','-count=1','-p=2','./...');Run vet $go @('vet','-p=2','./...')
  Write-Host'[5/7] Reproducing qualified ECO executable twice.';$a=Join-Path $w ECO.exe;$b=Join-Path $w ECO2.exe;$lf="-s -w -H windowsgui -buildid= -X github.com/ECO-evidence-casework-one/eco/internal/eco.SourceCommit=$EcoSource";Run build1 $go @('build','-trimpath','-buildvcs=false','-ldflags',$lf,'-o',$a,'./cmd/eco');Run build2 $go @('build','-trimpath','-buildvcs=false','-ldflags',$lf,'-o',$b,'./cmd/eco');[void](Assert-Sha $a $EcoExeSHA);[void](Assert-Sha $b $EcoExeSHA);if((Get-Item $a).Length-ne$EcoExeSize){throw'ECO size mismatch'};Remove-Item $b -Force;$a
 }finally{Set-Location $old;foreach($n in $names){[Environment]::SetEnvironmentVariable($n,$saved[$n],'Process')}}
}
function InstallLlama([string]$w,[string]$d){Write-Host'[6/7] Downloading hash-pinned llama.cpp b10259 from GitHub.';$z=Join-Path $w llama.zip;Download-Small $LlamaURL $z $LlamaZipSHA;$t=Join-Path $w llama;Expand-Checked $z $t;Move-Item $t $d;$ll=Find-Llama $d;$c=Capture $ll @('--offline','--version') $w;$diag=Diag $c;if($c.ExitCode-ne0-or-not$diag){throw"llama.cpp version probe failed: $diag"};$line=@($diag-split"`r?`n"|?{$_-match'version:'}|Select-Object -First 1);if($line.Count-ne1){throw"llama.cpp version not recognised: $diag"};Write-Host $line[0];[pscustomobject]@{Exe=$ll;SHA=(Sha $ll);Version=$line[0]}}
function Probe([string]$ll,[string]$m,[string]$w){$p=Join-Path $w prompt.txt;$s=Join-Path $w schema.json;Utf8 $p 'Return only the JSON object required by the schema. Set ok to ECO_AI_READY.';Utf8 $s '{"type":"object","additionalProperties":false,"required":["ok"],"properties":{"ok":{"type":"string","enum":["ECO_AI_READY"]}}}';$a=@('--offline','--model',$m,'--file',$p,'--json-schema-file',$s,'--simple-io','--no-display-prompt','--color','off','--log-disable','--seed','0','--temp','0','--top-k','1','--top-p','1','--min-p','0','--ctx-size','2048','--n-predict','64','--device','none','--n-gpu-layers','0','--fit','off','--no-context-shift','--no-perf');$c=Capture $ll $a $w;$o=$c.Stdout.Trim();if($c.ExitCode-ne0){throw"Qwen smoke test failed: $(Diag $c)"};try{$j=$o|ConvertFrom-Json}catch{throw"Qwen returned non-JSON: $o"};if($j.ok-cne'ECO_AI_READY'){throw"Unexpected Qwen result: $o"}}
function Launcher([string]$r,[string]$ll,[string]$sha){$rel=$ll.Substring($r.TrimEnd('\').Length+1);@('@echo off','setlocal','cd /d "%~dp0"','if not exist "PreviewUserData" mkdir "PreviewUserData"','set "LOCALAPPDATA=%~dp0PreviewUserData"', "set `"ECO_LLAMA_CPP=%~dp0$rel`"", "set `"ECO_LLAMA_CPP_SHA256=$sha`"", "set `"ECO_LLAMA_MODEL=%~dp0AIAssets\$ModelName`"", "set `"ECO_LLAMA_MODEL_SHA256=$ModelSHA`"", 'start "" "%~dp0ECO.exe"','endlocal')|Set-Content(Join-Path $r START_ECO_WITH_AI.cmd)-Encoding ascii}

if($SelfTest){SelfTest;if(-not$QualificationOnly){exit 0}}
if(-not[Environment]::Is64BitOperatingSystem){throw'64-bit Windows is required.'};if(-not$OutputRoot){$OutputRoot=Join-Path $PSScriptRoot ECO_RIG_AI_PREVIEW};$OutputRoot=[IO.Path]::GetFullPath($OutputRoot)
$carry='';if(Test-Path $OutputRoot){$res=Join-Path $OutputRoot AI_SETUP_RESULT.txt;$first='';if(Test-Path $res){$first=[string](Get-Content $res -TotalCount 1 -ErrorAction SilentlyContinue)};$launcher=Test-Path(Join-Path $OutputRoot START_ECO_WITH_AI.cmd);if($first-ne'ECO RIG AI SETUP STOPPED'-and($first-or$launcher)){throw"Output exists and is not a failed/interrupted setup: $OutputRoot"};$arc=$OutputRoot+'.interrupted-'+(Get-Date -Format'yyyyMMdd-HHmmss');Move-Item $OutputRoot $arc;Write-Host"Preserved interrupted attempt: $arc";$old=Join-Path $arc ('AIAssets\'+$ModelName+'.part');if(Test-Path $old){$carry=$old}}
[void][IO.Directory]::CreateDirectory($OutputRoot);$work=Join-Path $OutputRoot .work;[void][IO.Directory]::CreateDirectory($work);$result=Join-Path $OutputRoot AI_SETUP_RESULT.txt;$assets=Join-Path $OutputRoot AIAssets;[void][IO.Directory]::CreateDirectory($assets)
if($carry){$target=(Join-Path $assets $ModelName)+'.part';Move-Item $carry $target;$n=(Get-Item $target).Length;Write-Host("Recovered "+(Fmt $n)+" partial Qwen download; retry will resume it.")}
try{
 Hardware(Join-Path $OutputRoot RIG_AI_HARDWARE.json);$exe=BuildEco $work;Copy-Item $exe (Join-Path $OutputRoot ECO.exe);[void](Assert-Sha(Join-Path $OutputRoot ECO.exe)$EcoExeSHA);$ll=InstallLlama $work (Join-Path $OutputRoot Runtime);$total=Endpoint $work
 if($QualificationOnly){@('RIG AI QUALIFICATION PASS',"ECO SHA-256: $EcoExeSHA","llama-cli SHA-256: $($ll.SHA)",'Official Qwen range/resume endpoint: PASS',"Qwen reported bytes: $total")|Set-Content $result -Encoding utf8;Get-Content $result|Out-Host;exit 0}
 Write-Host'[7/7] Downloading official Qwen2.5 1.5B Q4_K_M with visible resumable progress.';$model=Join-Path $assets $ModelName;Download-Model $model $work $total;Probe $ll.Exe $model $work;Launcher $OutputRoot $ll.Exe $ll.SHA
 @('ECO RIG AI SETUP PASS','Real offline Qwen generation: PASS',"Qwen SHA-256: $ModelSHA",'Use START_ECO_WITH_AI.cmd to reopen. Synthetic/test material only.')|Set-Content $result -Encoding utf8;Remove-Item $work -Recurse -Force -ErrorAction SilentlyContinue;Get-Content $result|Out-Host;if(-not$NoLaunch){Start-Process(Join-Path $OutputRoot START_ECO_WITH_AI.cmd)-WorkingDirectory $OutputRoot}
}catch{$m=$_.Exception.Message;@('ECO RIG AI SETUP STOPPED',$m,'Any Qwen .part file is preserved for resume. Do not weaken Windows security.')|Set-Content $result -Encoding utf8;Write-Host $m;exit 1}
