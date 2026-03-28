# RepoDoctor

> **Static Architecture Analysis for Software Repositories**

RepoDoctor is a CLI tool that analyzes your repository’s architectural health by evaluating dependency structure, layering discipline, maintainability signals, and long-term architectural risk.

It is intentionally **not** a style linter. RepoDoctor focuses on higher-level design quality: cycles, layering violations, oversized units, and god object drift.

![CLI Version](https://img.shields.io/badge/cli-0.5.0--dev-blue)
![Roadmap Milestone](https://img.shields.io/badge/roadmap-v0.17-complete-brightgreen)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Structural Health](https://img.shields.io/badge/structural--health-100%2F100-brightgreen)]()

---

## Table of Contents

- [Why RepoDoctor?](#why-repodoctor)
- [Quick Start](#quick-start)
- [What You Get](#what-you-get)
- [Core Features](#core-features)
- [Language Support](#language-support)
- [Usage](#usage)
- [Configuration](#configuration)
- [Output & Exit Codes](#output--exit-codes)
- [Architecture Overview](#architecture-overview)
- [Project Structure](#project-structure)
- [Development & Quality Gates](#development--quality-gates)
- [CI Integration (GitHub Actions)](#ci-integration-github-actions)
- [Release Maturity](#release-maturity)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Privacy & Repository Hygiene](#privacy--repository-hygiene)
- [License](#license)

---

## Why RepoDoctor?

Most analysis tools optimize local code quality. RepoDoctor targets **system quality**.

| Question | RepoDoctor Answer |
|---|---|
| Are layers leaking responsibilities? | Layer validation rules |
| Are hidden dependency cycles forming? | Circular dependency analysis |
| Is complexity concentrating in god objects? | Method/field threshold checks |
| Are files/functions silently growing? | Size thresholds with actionable reporting |
| Is architecture quality improving or degrading? | Structural health score + trend-aware workflows |

RepoDoctor itself is continuously validated by running self-analysis on its own codebase.

---

## Quick Start

### 1) Clone & Build

```bash
git clone https://github.com/AdemFurkanATA/RepoDoctor.git
cd RepoDoctor
go build .
```

### 2) Analyze a Repository

```bash
# current directory
./RepoDoctor analyze -path .

# external repository
./RepoDoctor analyze -path /path/to/repository
```

Windows PowerShell:

```powershell
go build .
.\RepoDoctor.exe analyze -path .
```

---

## What You Get

Typical CLI flow:

```text
Scanning repository [████████████████████] 100%
Collecting metrics [████████████████████] 100%
Building dependency graph [████████████████████] 100%
Running rules [████████████████████] 100%

╔═══════════════════════════════════════════════════════════╗
║          RepoDoctor Structural Analysis Report           ║
╚═══════════════════════════════════════════════════════════╝

STRUCTURAL HEALTH SCORE
✓ Score: 100.0 / 100.0

VIOLATIONS SUMMARY
✓ No violations detected
```

---

## Core Features

### Analysis Engine

- **Circular Dependency Detection**
- **Layer Validation**
- **Size Threshold Analysis**
- **God Object Detection**
- **Structural Health Scoring (0–100)**
- **Deterministic rule execution pipeline**

### Developer Experience

- Interactive mode (`interactive`)
- Watch mode (`analyze -watch`)
- Progress bars
- Colored output (`--no-color` supported)
- Rule template generation
- Structured CLI error handling

### CI/CD Alignment

- Machine-readable JSON output
- Deterministic exit codes for pipelines
- Strict quality gates and merge policy

---

## Language Support

RepoDoctor uses an adapter-based architecture (`LanguageAdapter`) for multi-language support.

- **Go** (AST-driven analysis)
- **Python**
- **JavaScript / TypeScript**

Language detection is deterministic and policy-driven, with safeguards against noisy tooling directories.

---

## Usage

### Analyze

```bash
# text output (default)
repodoctor analyze -path .

# JSON output
repodoctor analyze -path ./my-repo -format json

# legacy JSON schema output (compat mode)
repodoctor analyze -path ./my-repo -format json-v1

# verbose mode
repodoctor analyze -path . -verbose

# watch mode
repodoctor analyze -path . -watch

# no color
repodoctor analyze -path . -no-color
```

### Other Commands

```bash
repodoctor interactive
repodoctor extract -path . -module RepoDoctor
repodoctor report -path repodoctor-report.json -format text
repodoctor history -path .
repodoctor generate rule my-custom-rule
repodoctor version
```

---

## Configuration

Create `.repodoctor/config.yaml`:

```yaml
size:
  max_file_lines: 500
  max_function_lines: 80

god_object:
  max_fields: 15
  max_methods: 10

rules:
  enable_size_rule: true
  enable_god_object_rule: true

weights:
  circular: 10
  layer: 5
  size: 3
  god_object: 5

language_detection:
  weights:
    Go: 1.0
    Python: 1.0
    JavaScript: 1.0
    TypeScript: 1.0
  tie_break_order: [Python, TypeScript, JavaScript, Go]
  segment_weights:
    src: 1.0
    app: 1.0
    pkg: 1.0
    tools: 0.2
    scripts: 0.2

architecture:
  profile: layered
```

Supported `architecture.profile` values: `clean`, `layered`, `modular-monolith`.

You can keep defaults and only override needed thresholds.

---

## Output & Exit Codes

### Exit Codes

| Code | Meaning |
|---|---|
| `0` | No critical violations |
| `2` | Critical violations detected |

### JSON Output (v2 example shape)

```json
{
  "version": "0.5.0-dev",
  "schemaVersion": "v2",
  "path": "/repo",
  "score": {
    "total": 100.0,
    "max": 100.0,
    "circularPenalty": 0,
    "layerPenalty": 0,
    "sizePenalty": 0,
    "godObjectPenalty": 0
  },
  "summary": {
    "totalViolations": 0,
    "circular": 0,
    "layer": 0,
    "size": 0,
    "godObject": 0
  },
  "language": {
    "detectedLanguage": "Go",
    "confidence": 1,
    "reasonCodes": ["go.mod"]
  },
  "circularViolations": [],
  "layerViolations": [],
  "sizeViolations": [],
  "godObjectViolations": []
}
```

Use `-format json-v1` only when a legacy integration still depends on the previous schema.

---

## Architecture Overview

### Analysis Pipeline

```text
detect language -> select adapter -> detect files -> collect metrics -> build dependency graph -> execute rules -> score -> report
```

### Layered Design

```text
CLI Layer         : command parsing, request composition, output
Application Layer : orchestration/pipeline control
Domain/Core       : rules, scoring, language policies, models
Infrastructure    : filesystem scanning, adapters, config loading
```

### Core Modules

- `internal/languages/` → adapters + language detection/stats
- `internal/analysis/` → orchestrator
- `internal/rules/` + `internal/engine/` → registry + execution
- `internal/model/` → graph, metrics, violations

---

## Project Structure

```text
RepoDoctor/
├── main.go
├── analysis_service.go
├── runtime_engine.go
├── config.go
├── reporter.go
├── progress.go
├── watcher.go
├── interactive.go
├── generator.go
├── internal/
│   ├── analysis/
│   ├── languages/
│   ├── rules/
│   ├── engine/
│   └── model/
└── .github/workflows/
```

---

## Development & Quality Gates

### Local Gates (mandatory)

```bash
go test ./...
go vet ./...
go run . analyze -path .
```

Expected architectural gate: **100/100** on self-analysis.

If a change touches concurrency/shared-state paths (`internal/languages`, `internal/rules`, `internal/engine`, `internal/analysis`), additionally run:

```bash
go test -race ./...
```

### Merge Discipline

- One issue = one branch
- Separate commit(s), separate push, separate PR
- PR target: `dev`
- `dev -> main` only after CI passes

---

## CI Integration (GitHub Actions)

Current workflow highlights:

- Linux + Windows matrix (`ubuntu-latest`, `windows-latest`)
- `go test ./...`
- `go vet ./...`
- deterministic multi-language confidence suite
- text + JSON analysis output artifact upload

Minimal example:

```yaml
name: RepoDoctor Analysis

on:
  pull_request:
    branches: [dev, main]

jobs:
  repodoctor:
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test ./...
      - run: go vet ./...
      - run: go test ./... -run "TestRunInternalRulePipeline_MultiLanguageMixedFixtureDeterministic|TestJSTSParity_DeterministicAcrossRepeatedRuns"
      - run: go run . analyze -path .
```

---

## Release Maturity

- Roadmap release train (`v0.10` to `v0.17`) is complete.
- Current release maturity report: `RELEASE_MATURITY_v0.17.md`.
- Release close gate status:
  - `go test ./...` pass
  - `go vet ./...` pass
  - `go run . analyze -path .` pass (`100/100`)

---

## Roadmap

### Completed

- v0.8: structural stabilization and 100/100 recovery
- v0.9: architecture hardening, invariants, and output stability improvements
- v0.10-v0.17: 8-sprint roadmap completed (adapter contracts, rule parity, orchestration confidence, incremental foundations, architecture profiles, precision/explainability hardening, release maturity)

### Next

- next release planning starts after v0.17 stabilization window
- expanded rule packs and profile refinements
- tighter CI policy templates and release automation

---

## Contributing

1. Branch from `dev`
2. Keep scope focused (single issue/goal)
3. Run quality gates locally
4. Open PR to `dev`
5. Merge to `main` only via green CI

Conventional commit prefixes are recommended: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`.

---

## Privacy & Repository Hygiene

The following are local/private artifacts and must not be published:

- `todo.md` (any location)
- AI planning/protocol files
- debug artifacts (`debug/`, `*.debug`, `debug.log`, `*.trace`)

These are ignored via `.gitignore`.

---

## License

MIT — see [LICENSE](LICENSE).
