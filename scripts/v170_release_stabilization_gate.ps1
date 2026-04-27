$ErrorActionPreference = 'Stop'

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$Command
    )

    Write-Host "==> $Name"
    Invoke-Expression $Command
    if (-not $? -or $LASTEXITCODE -ne 0) {
        throw "step failed: $Name (exit code $LASTEXITCODE)"
    }
}

$determinismRegex = "TestRunInternalRulePipeline_MultiLanguageMixedFixtureDeterministic|TestJSTSParity_DeterministicAcrossRepeatedRuns"
$v17PackRegex = "TestComputeAPIBreakingChanges_.*|TestComputeGitChurnSummary_.*|TestCollectHotspotSummary_DeterministicRiskOrdering|TestComputeDependencyVulnerabilities_.*|TestDocsGenerator_.*|TestParseGenerateDocsArgs_ValidAndInvalid"

Invoke-Step -Name "build" -Command "go build ."
Invoke-Step -Name "core test suite" -Command "go test ./..."
Invoke-Step -Name "static analysis" -Command "go vet ./..."
Invoke-Step -Name "v1.7 feature pack" -Command "go test ./... -run '$v17PackRegex'"
Invoke-Step -Name "determinism confidence suite" -Command "go test ./... -run '$determinismRegex'"
Invoke-Step -Name "self-analysis 100/100 gate" -Command "go run . analyze -path ."

Write-Host "v1.7 release stabilization gate passed."
