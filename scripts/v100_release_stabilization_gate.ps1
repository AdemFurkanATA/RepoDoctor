$ErrorActionPreference = 'Stop'

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$Command
    )

    Write-Host "==> $Name"
    Invoke-Expression $Command
}

Invoke-Step -Name "core test suite" -Command "go test ./..."
Invoke-Step -Name "static analysis" -Command "go vet ./..."
Invoke-Step -Name "determinism confidence suite" -Command "go test ./... -run 'TestRunInternalRulePipeline_MultiLanguageMixedFixtureDeterministic|TestJSTSParity_DeterministicAcrossRepeatedRuns'"
Invoke-Step -Name "v1.0 feature contract pack" -Command "go test ./... -run 'TestConfigLoader_RuleSeverityOverrides_DefaultAndValidation|TestLayerValidationRule_CustomLayerPolicyDetectsUpwardImport|TestCITemplateGenerator_GenerateGitHubTemplate|TestDetermineExitCode_NoViolationsReturnsZero|TestDetermineExitCode_CircularOrLayerViolationReturnsTwo|TestDetermineExitCode_SizeAndGodObjectViolationsRemainNonBlocking'"
Invoke-Step -Name "self-analysis 100/100 gate" -Command "go run . analyze -path ."

Write-Host "v1.0 release stabilization gate passed."
