$ErrorActionPreference = 'Stop'

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$Command
    )

    Write-Host "==> $Name"
    Invoke-Expression $Command
}

$closeoutRegex = "TestRunInternalRulePipeline_MultiLanguageMixedFixtureDeterministic|TestJSTSParity_DeterministicAcrossRepeatedRuns|TestArchitecture_InternalDependencyBoundaries|TestArchitecture_HighRiskImportsForbiddenInModelAndDomain|TestArchitecture_PurityBoundary_NoSideEffectImportsInSensitivePackages|TestReleaseCompatibility_JSONSchemaRequiredFields|TestReporter_JSONV1_CompatibilitySwitch|TestReporter_JSONV1_GoldenParity|TestReporter_JSONV1_DeterministicAcrossInputOrdering|TestReporter_JSONV1_WithViolations_StaysValidJSON|TestReporter_JSONV2_ContainsSchemaAndSummary|TestReporter_JSONV2_GoldenStableOrderingAndSchema|TestNormalizePathWithinRoot_RejectsAbsoluteOutsidePath|TestNormalizePathWithinRoot_RejectsParentTraversalEscape|TestNormalizePathWithinRoot_AllowsPathInsideRoot|TestNormalizePathWithinRoot_RejectsSymlinkOutsideRoot|TestNormalizePathWithinRoot_SymlinkLoopFailsSoftWithoutEscape"

Invoke-Step -Name "v0.18 close-out deterministic + boundary + compatibility + path-safety tests" -Command "go test ./... -run '$closeoutRegex'"
Invoke-Step -Name "v0.18 close-out self analysis gate" -Command "go run . analyze -path ."

Write-Host "v0.18 close-out gate pack passed."
