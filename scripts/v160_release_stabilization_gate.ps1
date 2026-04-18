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
$v16RulesRegex = "TestSecretDetectionRule_.*|TestCodeDuplicationRule_.*|TestDeadCodeRule_.*|TestErrorHandlingConsistencyRule_.*|TestInterfaceBloatRule_.*|TestCollectCyclomaticComplexitySummary_BandsDeterministic|TestEstimateTechnicalDebt_DeterministicBreakdown|TestInstallManagedPreCommitHook_.*"

Invoke-Step -Name "build" -Command "go build ."
Invoke-Step -Name "core test suite" -Command "go test ./..."
Invoke-Step -Name "static analysis" -Command "go vet ./..."
Invoke-Step -Name "v1.6 rule and dx pack" -Command "go test ./... -run '$v16RulesRegex'"
Invoke-Step -Name "v1.5 stabilization inheritance gate" -Command "pwsh -File './scripts/v150_release_stabilization_gate.ps1'"
Invoke-Step -Name "determinism confidence suite" -Command "go test ./... -run '$determinismRegex'"
Invoke-Step -Name "self-analysis 100/100 gate" -Command "go run . analyze -path ."

Write-Host "v1.6 release stabilization gate passed."
