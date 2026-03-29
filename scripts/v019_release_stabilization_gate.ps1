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
Invoke-Step -Name "v0.18 close-out pack" -Command "pwsh -File ./scripts/v018_closeout_gate.ps1"
Invoke-Step -Name "v0.19 benchmark + memory pack" -Command "pwsh -File ./scripts/v019_benchmark_gate.ps1"
Invoke-Step -Name "self-analysis 100/100 gate" -Command "go run . analyze -path ."

Write-Host "v0.19 release stabilization gate passed."
