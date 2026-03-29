$ErrorActionPreference = 'Stop'

function Get-BudgetNs {
    param(
        [Parameter(Mandatory = $true)][string]$EnvName,
        [Parameter(Mandatory = $true)][double]$Default
    )

    $raw = [Environment]::GetEnvironmentVariable($EnvName)
    if ([string]::IsNullOrWhiteSpace($raw)) {
        return $Default
    }

    $parsed = 0.0
    if (-not [double]::TryParse($raw, [System.Globalization.NumberStyles]::Float, [System.Globalization.CultureInfo]::InvariantCulture, [ref]$parsed)) {
        throw "Invalid numeric budget in ${EnvName}: '$raw'"
    }
    return $parsed
}

function Assert-BenchmarkUnderBudget {
    param(
        [Parameter(Mandatory = $true)][string]$Output,
        [Parameter(Mandatory = $true)][string]$BenchmarkName,
        [Parameter(Mandatory = $true)][double]$BudgetNs
    )

    $regex = [regex]("(?m)^" + [regex]::Escape($BenchmarkName) + "(?:-\d+)?\s+\d+\s+([0-9]+(?:\.[0-9]+)?)\s+ns/op")
    $match = $regex.Match($Output)
    if (-not $match.Success) {
        throw "Benchmark output missing expected line: $BenchmarkName"
    }

    $ns = 0.0
    if (-not [double]::TryParse($match.Groups[1].Value, [System.Globalization.NumberStyles]::Float, [System.Globalization.CultureInfo]::InvariantCulture, [ref]$ns)) {
        throw "Unable to parse ns/op for $BenchmarkName"
    }

    if ($ns -gt $BudgetNs) {
        throw "$BenchmarkName breached budget: ${ns}ns/op > ${BudgetNs}ns/op"
    }
}

$buildMapBudgetNs = Get-BudgetNs -EnvName "RD_BUDGET_BUILD_MAP_NS" -Default 2000000000
$diffStateBudgetNs = Get-BudgetNs -EnvName "RD_BUDGET_DIFF_STATE_NS" -Default 10000000

Write-Host "Running v0.19 benchmark budget gate..."

$buildMapOutput = go test ./internal/analysis -run ^$ -bench BenchmarkBuildGoFingerprintMap_LargeRepoFixture -benchmem -count=1 | Out-String
$diffStateOutput = go test ./internal/analysis -run ^$ -bench BenchmarkDiffFingerprintStates_LargeMap -benchmem -count=1 | Out-String

$combinedOutput = @(
    "# BenchmarkBuildGoFingerprintMap_LargeRepoFixture",
    $buildMapOutput,
    "# BenchmarkDiffFingerprintStates_LargeMap",
    $diffStateOutput
) -join [Environment]::NewLine

Set-Content -Path "benchmark-gate.txt" -Value $combinedOutput -Encoding utf8

Assert-BenchmarkUnderBudget -Output $buildMapOutput -BenchmarkName "BenchmarkBuildGoFingerprintMap_LargeRepoFixture" -BudgetNs $buildMapBudgetNs
Assert-BenchmarkUnderBudget -Output $diffStateOutput -BenchmarkName "BenchmarkDiffFingerprintStates_LargeMap" -BudgetNs $diffStateBudgetNs

Write-Host "v0.19 benchmark gate passed."
