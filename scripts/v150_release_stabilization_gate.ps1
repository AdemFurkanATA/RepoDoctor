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

function Assert-PathExists {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Label
    )

    if (-not (Test-Path $Path)) {
        throw "$Label missing: $Path"
    }
}

$determinismRegex = "TestRunInternalRulePipeline_MultiLanguageMixedFixtureDeterministic|TestJSTSParity_DeterministicAcrossRepeatedRuns"
$dxRegex = "TestBuildReportFromRuleViolations_AttachesRemediationHints|TestInteractiveFixSuggestionAdvisor_.*|TestBuildInteractiveFixSuggestions_.*|TestParseBoolEnv_.*"

Invoke-Step -Name "build" -Command "go build ."
Invoke-Step -Name "core test suite" -Command "go test ./..."
Invoke-Step -Name "static analysis" -Command "go vet ./..."
Invoke-Step -Name "v1.5 remediation + interaction pack" -Command "go test ./... -run '$dxRegex'"
Invoke-Step -Name "architecture boundary pack" -Command "go test . -run 'TestArchitecture_'"

Assert-PathExists -Path "Dockerfile" -Label "distribution dockerfile"
Assert-PathExists -Path "Formula/repodoctor.rb" -Label "distribution formula"
Assert-PathExists -Path ".github/workflows/release-distribution.yml" -Label "distribution workflow"
Assert-PathExists -Path "vscode-repodoctor/package.json" -Label "extension package"

if (Get-Command npm -ErrorAction SilentlyContinue) {
    Invoke-Step -Name "vscode extension compile pack" -Command "npm install --prefix 'vscode-repodoctor' && npm run --prefix 'vscode-repodoctor' compile"
} else {
    Write-Host "==> vscode extension compile pack (skipped: npm not found)"
}

Invoke-Step -Name "v1.4 benchmark/race inheritance gate" -Command "pwsh -File './scripts/v140_release_stabilization_gate.ps1'"
Invoke-Step -Name "determinism confidence suite" -Command "go test ./... -run '$determinismRegex'"
Invoke-Step -Name "self-analysis 100/100 gate" -Command "go run . analyze -path ."

Write-Host "v1.5 release stabilization gate passed."
