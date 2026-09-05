# Private synthetic-data build preparation. Never launches ECO or changes machine policy.
# Requires Windows PowerShell 5.1 or later. All build files stay in a new output directory.
[CmdletBinding()]
param(
    [string]$ManifestPath = (Join-Path $PSScriptRoot 'private-build-manifest.json'),
    [string]$OutputRoot = '',
    [switch]$LibraryOnly,
    [switch]$NoPrompt
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Assert-EcoHash([string]$Path, [string]$Expected) {
    if ($Expected -notmatch '^[a-fA-F0-9]{64}$') { throw 'Invalid expected SHA-256.' }
    $actual = (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actual -cne $Expected.ToLowerInvariant()) { throw "SHA-256 mismatch for $([IO.Path]::GetFileName($Path)). Nothing will be launched." }
    return $actual
}

function Expand-EcoCheckedZip([string]$Archive, [string]$Destination, [string]$StripPrefix = '') {
    if (Test-Path -LiteralPath $Destination) { throw 'Extraction destination already exists; preserving it.' }
    Add-Type -AssemblyName System.IO.Compression
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $root = [IO.Path]::GetFullPath($Destination).TrimEnd('\') + '\'
    $zip = [IO.Compression.ZipFile]::OpenRead($Archive)
    try {
        if ($zip.Entries.Count -gt 40000) { throw 'Archive entry limit exceeded.' }
        $seen = New-Object 'System.Collections.Generic.HashSet[string]' ([StringComparer]::OrdinalIgnoreCase)
        $plan = New-Object 'System.Collections.Generic.List[object]'
        [long]$total = 0
        foreach ($entry in $zip.Entries) {
            $name = $entry.FullName.Replace('\', '/')
            if ($StripPrefix) {
                if (-not $name.StartsWith($StripPrefix, [StringComparison]::Ordinal)) { throw 'Unexpected source archive root.' }
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
            if ($total -gt 1GB -or $entry.Length -gt 128MB) { throw 'Archive size limit exceeded.' }
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

function Get-EcoCheckedDownload([string]$Url, [string]$Destination, [string]$Sha256) {
    if (Test-Path -LiteralPath $Destination) { throw 'Download destination already exists.' }
    $oldProgress = $ProgressPreference
    $ProgressPreference = 'SilentlyContinue'
    try {
        Invoke-WebRequest -UseBasicParsing -Uri $Url -OutFile $Destination -TimeoutSec 300
        [void](Assert-EcoHash $Destination $Sha256)
    } finally { $ProgressPreference = $oldProgress }
}

function Invoke-EcoGo([string]$Go, [string[]]$Arguments) {
    & $Go @Arguments | Out-Host
    if ($LASTEXITCODE -ne 0) { throw "Go command failed (exit $LASTEXITCODE): $($Arguments -join ' ')" }
}

function Invoke-EcoPrivateBuild([string]$Manifest, [string]$Destination) {
    $m = Get-Content -LiteralPath $Manifest -Raw | ConvertFrom-Json
    $source = '8b69a669b003fe30e84f1d344aa7533eb9cd9045'
    if ($m.source_commit -cne $source) { throw 'This preparer is restricted to the frozen PR #134 source.' }
    foreach ($hash in @($m.source_archive_sha256, $m.executable_sha256)) {
        if ($hash -notmatch '^[a-f0-9]{64}$') { throw 'Manifest lacks a qualified fingerprint.' }
    }
    if ([DateTimeOffset]::UtcNow -gt [DateTimeOffset]::Parse($m.expires_utc)) { throw 'Private preparation window expired. Obtain a current handoff before rebuilding.' }
    if (-not [Environment]::Is64BitOperatingSystem) { throw 'A 64-bit Windows machine is required.' }
    if (Test-Path -LiteralPath $Destination) { throw 'Build destination already exists; nothing has been overwritten.' }
    $work = [IO.Path]::GetFullPath($Destination)
    if ($work.Length -gt 110 -or $work.StartsWith('\\')) { throw 'Extract the preparation kit to a short local folder, then try again.' }
    [void][IO.Directory]::CreateDirectory($work)
    $envNames = @('GOENV','GOTOOLCHAIN','GOWORK','GOTELEMETRY','GOROOT','GOPATH','GOCACHE','GOMODCACHE','GOTMPDIR','GOOS','GOARCH','GOAMD64','CGO_ENABLED','GOMAXPROCS','GOFLAGS','GOEXPERIMENT','GODEBUG','GOPROXY','GOSUMDB','GOPRIVATE','GONOSUMDB','GONOPROXY','TEMP','TMP')
    $saved = @{}
    foreach ($name in $envNames) { $saved[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
    $oldLocation = Get-Location
    $oldTls = [Net.ServicePointManager]::SecurityProtocol
    $transcript = $false
    try {
        Start-Transcript -Path (Join-Path $work 'build-log.txt') | Out-Null
        $transcript = $true
        [Net.ServicePointManager]::SecurityProtocol = $oldTls -bor [Net.SecurityProtocolType]::Tls12
        Write-Host '[1/6] Downloading the fixed source and portable Go compiler. No installer is run.'
        $srcZip = Join-Path $work 'source.zip'
        $goZip = Join-Path $work 'go.zip'
        Get-EcoCheckedDownload "https://codeload.github.com/ECO-evidence-casework-one/eco/zip/$source" $srcZip $m.source_archive_sha256
        Get-EcoCheckedDownload 'https://go.dev/dl/go1.23.12.windows-amd64.zip' $goZip '07c35866cdd864b81bb6f1cfbf25ac7f87ddc3a976ede1bf5112acbb12dfe6dc'
        Write-Host '[2/6] Checking and extracting the two archives into this build folder.'
        $src = Join-Path $work 'source'
        $sdk = Join-Path $work 'compiler'
        Expand-EcoCheckedZip $srcZip $src "eco-$source/"
        Expand-EcoCheckedZip $goZip $sdk
        $go = Join-Path $sdk 'go\bin\go.exe'
        $env:GOROOT = Join-Path $sdk 'go'
        $env:GOENV = 'off'; $env:GOTOOLCHAIN = 'local'; $env:GOWORK = 'off'; $env:GOTELEMETRY = 'off'
        $env:GOPATH = Join-Path $work 'cache'; $env:GOCACHE = Join-Path $work 'cache\build'
        $env:GOMODCACHE = Join-Path $work 'cache\modules'; $env:GOTMPDIR = Join-Path $work 'temporary'
        [void][IO.Directory]::CreateDirectory($env:GOTMPDIR)
        $env:TEMP = $env:GOTMPDIR; $env:TMP = $env:GOTMPDIR
        $env:GOOS = 'windows'; $env:GOARCH = 'amd64'; $env:GOAMD64 = 'v1'; $env:CGO_ENABLED = '0'
        $env:GOMAXPROCS = '2'; $env:GOFLAGS = '-mod=readonly -buildvcs=false'
        $env:GOEXPERIMENT = ''; $env:GODEBUG = ''; $env:GOPRIVATE = ''; $env:GONOSUMDB = ''; $env:GONOPROXY = ''
        $env:GOPROXY = 'https://proxy.golang.org'; $env:GOSUMDB = 'sum.golang.org'
        Set-Location -LiteralPath $src
        $version = & $go version
        if ($LASTEXITCODE -ne 0 -or $version -cne 'go version go1.23.12 windows/amd64') { throw 'Compiler version does not match the qualified recipe.' }
        Write-Host '[3/6] Downloading and verifying the source-pinned Go dependencies.'
        Invoke-EcoGo $go @('mod','download')
        Invoke-EcoGo $go @('mod','verify')
        $env:GOPROXY = 'off'
        Write-Host '[4/6] Running the Windows source tests. This does not open your personal workspaces.'
        Invoke-EcoGo $go @('test','-count=1','-p=2','./...')
        Invoke-EcoGo $go @('vet','-p=2','./...')
        Write-Host '[5/6] Building twice and checking the qualified executable fingerprint.'
        $candidate = Join-Path $work 'candidate'
        $stage = Join-Path $work 'build-staging'
        [void][IO.Directory]::CreateDirectory($stage)
        $first = Join-Path $stage 'ECO.exe'; $second = Join-Path $work 'ECO.rebuild.exe'
        $flags = "-s -w -H windowsgui -buildid= -X github.com/ECO-evidence-casework-one/eco/internal/eco.SourceCommit=$source"
        Invoke-EcoGo $go @('build','-p=2','-trimpath','-buildvcs=false','-ldflags',$flags,'-o',$first,'./cmd/eco')
        Invoke-EcoGo $go @('build','-p=2','-trimpath','-buildvcs=false','-ldflags',$flags,'-o',$second,'./cmd/eco')
        $sha = Assert-EcoHash $first $m.executable_sha256
        [void](Assert-EcoHash $second $sha)
        $signature = Get-AuthenticodeSignature -LiteralPath $first
        if ($signature.Status.ToString() -ne 'NotSigned') { throw 'Unexpected signature status for the fixed unsigned test candidate.' }
        $receipt = [ordered]@{
            source_commit=$source; executable_sha256=$sha; size_bytes=(Get-Item -LiteralPath $first).Length
            source_archive_sha256=$m.source_archive_sha256; compiler=$version; recipe='archive-source, buildvcs=false, trimpath, explicit SourceCommit'
            tests='PASS'; vet='PASS'; deterministic_rebuild='PASS'; expected_fingerprint='PASS'
            signature_status=$signature.Status.ToString(); recipient=$m.recipient; purpose=$m.purpose
            expires_utc=$m.expires_utc; created_utc=[DateTimeOffset]::UtcNow.ToString('o')
            app_launched=$false; real_evidence_permitted=$false; public_distribution_permitted=$false
            optional_runtimes='Not bundled or registered by this preparer'; actual_machine_acceptance='NOT RUN'
        }
        $receipt | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath (Join-Path $stage 'build-receipt.json') -Encoding UTF8
        Copy-Item -LiteralPath (Join-Path $src 'LICENSE') -Destination (Join-Path $stage 'LICENSE')
        # Publish the whole qualified directory only after every check succeeds.
        [IO.Directory]::Move($stage, $candidate)
        Write-Host '[6/6] PREPARATION PASSED. ECO was not launched. Keep the compiler and source folders private.'
        Write-Host "Executable SHA-256: $sha"
        return [pscustomobject]$receipt
    } finally {
        Set-Location $oldLocation
        foreach ($name in $envNames) { [Environment]::SetEnvironmentVariable($name, $saved[$name], 'Process') }
        [Net.ServicePointManager]::SecurityProtocol = $oldTls
        if ($transcript) { Stop-Transcript | Out-Null }
    }
}

if (-not $LibraryOnly) {
    $resultFile = Join-Path $PSScriptRoot 'BUILD_RESULT.txt'
    try {
        if (-not $NoPrompt) {
            Write-Host 'This creates an UNSIGNED, private synthetic-data test build. It is NOT a public release.'
            Write-Host 'It downloads public source/compiler/dependencies, runs tests, and builds locally.'
            Write-Host 'No account, administrator elevation, registry change, upload, or automatic ECO launch.'
            if ((Read-Host 'Type BUILD to prepare, or press Enter to stop') -cne 'BUILD') { throw 'Preparation cancelled; no build requested.' }
        }
        if (-not $OutputRoot) { $OutputRoot = Join-Path $PSScriptRoot ('build-' + [Guid]::NewGuid().ToString('N').Substring(0,8)) }
        $r = Invoke-EcoPrivateBuild $ManifestPath $OutputRoot
        @('PREPARATION PASSED - ECO NOT LAUNCHED', "Source: $($r.source_commit)", "EXE SHA-256: $($r.executable_sha256)", "Size: $($r.size_bytes)", 'Tests / vet / repeat build / expected hash: PASS', "Local output folder: $([IO.Path]::GetFileName($OutputRoot))\candidate", 'Signature: UNSIGNED PRIVATE TEST BUILD', 'Next: send this result back in the private chat. Do not import real evidence.') | Set-Content -LiteralPath $resultFile -Encoding UTF8
        Get-Content -LiteralPath $resultFile | Write-Host
    } catch {
        $message = $_.Exception.Message
        if ($env:USERPROFILE) { $message = $message.Replace($env:USERPROFILE, '[user-profile]') }
        @('PREPARATION STOPPED - ECO NOT LAUNCHED', $message, 'Do not change Defender, Smart App Control or machine security settings. Send this result to the private chat.') | Set-Content -LiteralPath $resultFile -Encoding UTF8
        Write-Host $message
        exit 1
    }
}
