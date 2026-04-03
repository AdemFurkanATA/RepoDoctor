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
$rulesEdgeRegex = "TestCircularDependencyRule_.*|TestGodObjectRule_.*|TestCategories_AllAndValidations|TestExampleRule_ImplementsContract|TestLoadFromDir_FiltersHiddenAndNonGoFiles"
$parserExtractorRegex = "TestImportExtractor_ExtractFromFile_MalformedFailSoft|TestImportExtractor_ExtractFromDir_SkipsMalformedKeepsValid|FuzzImportExtractor_ExtractFromFile_NoPanic|FuzzParsePythonImportLine_NoPanic|FuzzParsePythonDynamicImport_NoPanic|TestPythonAdapter_CollectEvidence_FailSoftOnMalformedAndOversized"

Invoke-Step -Name "build" -Command "go build ."
Invoke-Step -Name "core test suite" -Command "go test ./..."
Invoke-Step -Name "static analysis" -Command "go vet ./..."
Invoke-Step -Name "v1.2 rules edge-case regression pack" -Command "go test ./internal/rules -run '$rulesEdgeRegex'"
Invoke-Step -Name "v1.2 parser/extractor malformed-corpus pack" -Command "go test ./... -run '$parserExtractorRegex'"
Invoke-Step -Name "v1.2 property-based determinism pack" -Command "go test . -run 'TestProperty_'"
Invoke-Step -Name "determinism confidence suite" -Command "go test ./... -run '$determinismRegex'"
Invoke-Step -Name "self-analysis 100/100 gate" -Command "go run . analyze -path ."

Write-Host "v1.2 release stabilization gate passed."
