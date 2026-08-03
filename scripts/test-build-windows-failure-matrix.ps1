$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "native-command.ps1")

$stages = @(
    "go test ./...",
    "go vet ./...",
    "source-policy check",
    "first deterministic Windows build",
    "second deterministic Windows build",
    "go version"
)

foreach ($failIndex in 0..($stages.Count - 1)) {
    $observed = [System.Collections.Generic.List[string]]::new()
    $caughtExpectedFailure = $false

    try {
        for ($index = 0; $index -lt $stages.Count; $index++) {
            $description = $stages[$index]
            $observed.Add($description)
            $exitCode = if ($index -eq $failIndex) { 23 } else { 0 }

            Invoke-NativeChecked $description {
                & $env:ComSpec /d /c "exit $exitCode"
            }
        }
    } catch {
        $expectedDescription = [regex]::Escape($stages[$failIndex])
        if ($_.Exception.Message -notmatch $expectedDescription -or $_.Exception.Message -notmatch "native exit code 23") {
            throw
        }
        $caughtExpectedFailure = $true
    }

    if (-not $caughtExpectedFailure) {
        throw "Failure matrix stage '$($stages[$failIndex])' did not terminate the pipeline."
    }

    $expectedObserved = $stages[0..$failIndex]
    if ($observed.Count -ne $expectedObserved.Count) {
        throw "Failure matrix stage '$($stages[$failIndex])' executed later stages. Observed: $($observed -join ', ')"
    }

    for ($index = 0; $index -lt $expectedObserved.Count; $index++) {
        if ($observed[$index] -ne $expectedObserved[$index]) {
            throw "Failure matrix order mismatch at index $index for '$($stages[$failIndex])'."
        }
    }
}

$buildScriptPath = Join-Path $PSScriptRoot "build-windows.ps1"
$buildScript = Get-Content $buildScriptPath -Raw
$requiredWrappers = @(
    'Invoke-NativeChecked "go test ./..." { go test ./... }',
    'Invoke-NativeChecked "go vet ./..." { go vet ./... }',
    'Invoke-NativeChecked "source-policy check" { python scripts/check_source_policy.py }',
    'Invoke-NativeChecked "first deterministic Windows build" { go build -trimpath -ldflags $ldflags -o $first ./cmd/eco }',
    'Invoke-NativeChecked "second deterministic Windows build" { go build -trimpath -ldflags $ldflags -o $second ./cmd/eco }',
    '$goVersion = Invoke-NativeChecked "go version" { go version }'
)

foreach ($requiredWrapper in $requiredWrappers) {
    $count = ([regex]::Matches($buildScript, [regex]::Escape($requiredWrapper))).Count
    if ($count -ne 1) {
        throw "Expected exactly one checked native stage in build-windows.ps1: $requiredWrapper (found $count)."
    }
}

$forbiddenUnwrappedLines = @(
    '(?m)^\s*go test \.\/\.\.\.\s*$',
    '(?m)^\s*go vet \.\/\.\.\.\s*$',
    '(?m)^\s*python scripts/check_source_policy\.py\s*$',
    '(?m)^\s*go build .*\.\/cmd\/eco\s*$'
)

foreach ($pattern in $forbiddenUnwrappedLines) {
    if ($buildScript -match $pattern) {
        throw "Found an unwrapped native command in build-windows.ps1 matching: $pattern"
    }
}

Write-Host "Windows native-command failure matrix PASS."
exit 0
