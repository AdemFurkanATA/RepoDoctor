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
$observabilityRegex = "TestBuildProfilingRequest_.*|TestStartProfiling_.*|TestComposeAnalyzeRequest_ParsesAndNormalizesProfileFlags|TestResolveMetricsExportPath_.*|TestAnalysisService_Run_ExportsMetricsWhenEnabled|TestCLIError_.*|TestCategoryCode_.*"

Invoke-Step -Name "build" -Command "go build ."
Invoke-Step -Name "core test suite" -Command "go test ./..."
Invoke-Step -Name "static analysis" -Command "go vet ./..."
Invoke-Step -Name "observability regression pack" -Command "go test ./... -run '$observabilityRegex'"
Invoke-Step -Name "architecture boundary pack" -Command "go test . -run 'TestArchitecture_'"
Invoke-Step -Name "determinism confidence suite" -Command "go test ./... -run '$determinismRegex'"
Invoke-Step -Name "self-analysis 100/100 gate" -Command "go run . analyze -path ."

Write-Host "v1.3 release stabilization gate passed."
