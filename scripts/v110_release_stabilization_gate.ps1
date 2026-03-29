$ErrorActionPreference = 'Stop'

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$Command
    )

    Write-Host "==> $Name"
    Invoke-Expression $Command
}

$determinismRegex = "TestRunInternalRulePipeline_MultiLanguageMixedFixtureDeterministic|TestJSTSParity_DeterministicAcrossRepeatedRuns"
$incrementalRegex = "TestIncrementalBoundaryService_.*|TestFileIncrementalSnapshotStore_.*|TestIncrementalCorrectness_.*|TestBuildIncrementalCacheKey_.*|TestIncrementalCacheSnapshot_.*|TestHashFileSHA256Streaming_MatchesReferenceHash"

Invoke-Step -Name "build" -Command "go build ."
Invoke-Step -Name "core test suite" -Command "go test ./..."
Invoke-Step -Name "static analysis" -Command "go vet ./..."
Invoke-Step -Name "v1.1 incremental regression pack" -Command "go test ./internal/analysis -run '$incrementalRegex'"
Invoke-Step -Name "determinism confidence suite" -Command "go test ./... -run '$determinismRegex'"
Invoke-Step -Name "self-analysis 100/100 gate" -Command "go run . analyze -path ."

Write-Host "v1.1 release stabilization gate passed."
