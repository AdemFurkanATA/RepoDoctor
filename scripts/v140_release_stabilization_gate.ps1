$ErrorActionPreference = 'Stop'

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$Command
    )

    Write-Host "==> $Name"
    Invoke-Expression $Command
    if ($LASTEXITCODE -ne 0) {
        throw "step failed: $Name (exit code $LASTEXITCODE)"
    }
}

function Get-Percentile {
    param(
        [Parameter(Mandatory = $true)][long[]]$Values,
        [Parameter(Mandatory = $true)][double]$Percentile
    )

    if ($Values.Count -eq 0) {
        throw "no samples provided"
    }

    $sorted = $Values | Sort-Object
    $rank = [Math]::Ceiling($Percentile * $sorted.Count)
    if ($rank -lt 1) { $rank = 1 }
    if ($rank -gt $sorted.Count) { $rank = $sorted.Count }
    return [long]$sorted[$rank - 1]
}

function Assert-BenchmarkBudget {
    param(
        [Parameter(Mandatory = $true)][string]$BenchmarkOutput
    )

    $nsSamples = [System.Collections.Generic.List[long]]::new()
    $allocSamples = [System.Collections.Generic.List[long]]::new()

    $regex = [regex]'BenchmarkGoAdapter_ParallelParsing-\d+\s+\d+\s+(\d+) ns/op\s+(\d+) B/op'
    foreach ($line in ($BenchmarkOutput -split "`n")) {
        $match = $regex.Match($line)
        if ($match.Success) {
            [void]$nsSamples.Add([long]::Parse($match.Groups[1].Value))
            [void]$allocSamples.Add([long]::Parse($match.Groups[2].Value))
        }
    }

    if ($nsSamples.Count -lt 3) {
        throw "insufficient benchmark samples for budget check"
    }

    $p50 = Get-Percentile -Values $nsSamples.ToArray() -Percentile 0.50
    $p95 = Get-Percentile -Values $nsSamples.ToArray() -Percentile 0.95
    $maxAlloc = ($allocSamples | Sort-Object -Descending | Select-Object -First 1)

    $p50Limit = 25000000
    $p95Limit = 50000000
    $allocLimit = 3000000

    Write-Host "Benchmark budget summary: p50=$p50 ns/op, p95=$p95 ns/op, maxAlloc=$maxAlloc B/op"

    if ($p50 -gt $p50Limit) {
        throw "p50 budget exceeded: $p50 > $p50Limit"
    }
    if ($p95 -gt $p95Limit) {
        throw "p95 budget exceeded: $p95 > $p95Limit"
    }
    if ($maxAlloc -gt $allocLimit) {
        throw "memory budget exceeded: $maxAlloc > $allocLimit"
    }
}

$determinismRegex = "TestRunInternalRulePipeline_MultiLanguageMixedFixtureDeterministic|TestJSTSParity_DeterministicAcrossRepeatedRuns"
$budgetRegex = "TestOrchestrator_Analyze_AppliesMemoryFileBudget|TestOrchestrator_Analyze_InvalidMemoryBudgetFailSoft|TestFileIncrementalSnapshotStore_WarmupFromFilesystem_.*"

Invoke-Step -Name "build" -Command "go build ."
Invoke-Step -Name "core test suite" -Command "go test ./..."
Invoke-Step -Name "static analysis" -Command "go vet ./..."
Invoke-Step -Name "v1.4 memory/cache guard pack" -Command "go test ./internal/analysis -run '$budgetRegex'"

Write-Host "==> benchmark regression pack"
$bench = go test ./internal/languages -bench 'BenchmarkGoAdapter_ParallelParsing' -benchmem -run '^$' -count 5
$benchExit = $LASTEXITCODE
$benchText = ($bench | Out-String)
Write-Host $benchText
if ($benchExit -ne 0) {
    throw "step failed: benchmark regression pack (exit code $benchExit)"
}
Assert-BenchmarkBudget -BenchmarkOutput $benchText

Invoke-Step -Name "race detector pack" -Command "go test -race ./..."
Invoke-Step -Name "determinism confidence suite" -Command "go test ./... -run '$determinismRegex'"
Invoke-Step -Name "self-analysis 100/100 gate" -Command "go run . analyze -path ."

Write-Host "v1.4 release stabilization benchmark gate passed."
