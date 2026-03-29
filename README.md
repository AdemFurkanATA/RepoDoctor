# RepoDoctor

> **Static Architecture Analysis for Software Repositories**

RepoDoctor is a CLI tool that analyzes architectural quality in code repositories.
It focuses on **structure-level risks** (dependency cycles, layer violations, oversized units, god objects),
not formatting or style-lint details.

![CLI Version](https://img.shields.io/badge/cli-0.5.0--dev-blue)
![Roadmap Milestone](https://img.shields.io/badge/roadmap-v1.1-complete-brightgreen)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Structural Health](https://img.shields.io/badge/structural--health-100%2F100-brightgreen)]()

---

## Table of Contents

- [Why RepoDoctor?](#why-repodoctor)
- [What This Project Includes Today](#what-this-project-includes-today)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Output & Exit Codes](#output--exit-codes)
- [CI/CD Integration](#cicd-integration)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Quality Gates](#quality-gates)
- [Release & Roadmap Status](#release--roadmap-status)
- [Contributing](#contributing)
- [Privacy & Repository Hygiene](#privacy--repository-hygiene)
- [License](#license)

---

## Why RepoDoctor?

Most tools answer: _“Is this file clean?”_  
RepoDoctor answers: _“Is this system staying healthy over time?”_

It helps teams detect architecture drift early, keep CI objective, and enforce structural standards consistently.

| Question | RepoDoctor Answer |
|---|---|
| Are dependency cycles emerging? | Circular dependency detection |
| Are layers importing upward incorrectly? | Layer validation rules |
| Are files/functions silently growing too much? | Size threshold rules |
| Is complexity concentrating in a few objects? | God-object detection |
| Is architecture quality improving or degrading? | Deterministic score + report trend usage |

---

## What This Project Includes Today

### Core analysis capabilities

- Circular dependency detection
- Layer validation (including custom architecture profiles)
- Size threshold analysis (file/function)
- God object detection
- Structural health scoring (`0-100`)

### Multi-language analysis support

- Go
- Python
- JavaScript / TypeScript
- Java (pilot layer introduced in v0.20; decision memo included)

### Developer and CI workflows

- Human-readable and JSON outputs (`json`, `json-v1`)
- Deterministic execution and ordering
- CI template generation (`generate ci ...`)
- Release gate packs for milestone closeouts

---

## Quick Start

### 1) Clone and build

```bash
git clone https://github.com/AdemFurkanATA/RepoDoctor.git
cd RepoDoctor
go build .
```

### 2) Analyze current repository

```bash
./RepoDoctor analyze -path .
```

Windows PowerShell:

```powershell
go build .
.\RepoDoctor.exe analyze -path .
```

---

## Installation

### Requirements

- Go **1.21+**
- Git

### Build from source

```bash
go build .
```

This produces a local binary (`RepoDoctor` / `RepoDoctor.exe`).

---

## Usage

### Analyze command

```bash
# default text output
repodoctor analyze -path .

# JSON output (v2)
repodoctor analyze -path . -format json

# legacy JSON output (compat mode)
repodoctor analyze -path . -format json-v1

# verbose mode
repodoctor analyze -path . -verbose

# watch mode
repodoctor analyze -path . -watch

# disable colors
repodoctor analyze -path . -no-color
```

### Other commands

```bash
repodoctor interactive
repodoctor extract -path . -module RepoDoctor
repodoctor report -path repodoctor-report.json -format text
repodoctor history -path .
repodoctor generate rule my-custom-rule
repodoctor generate ci github
repodoctor generate ci gitlab
repodoctor generate ci azure
repodoctor version
```

`generate ci <github|gitlab|azure> [--force]` creates deterministic CI starter templates.

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
  circular_severity: critical
  layer_severity: error
  size_severity: warning
  god_object_severity: warning

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
    Java: 0.8
  tie_break_order: [Python, TypeScript, JavaScript, Go, Java]

architecture:
  profile: layered
  custom_layer_order: [handler, service, repo]
  custom_layer_keywords:
    handler: [handler, controller]
    service: [service, usecase]
    repo: [repo, repository, data]
```

Supported architecture profiles:

- `clean`
- `layered`
- `modular-monolith`

---

## Output & Exit Codes

### Exit codes

| Code | Meaning |
|---|---|
| `0` | No critical violations |
| `2` | Critical violations detected (circular/layer) |

### JSON output

RepoDoctor supports:

- `json` (current schema)
- `json-v1` (legacy compatibility)

Use `json-v1` only if an existing integration still depends on old schema.

---

## CI/CD Integration

Current CI pipeline includes:

- Linux + Windows matrix
- `go test ./...`
- `go vet ./...`
- deterministic confidence suites
- self-analysis score gate
- milestone gate packs:
  - `scripts/v018_closeout_gate.ps1`
  - `scripts/v019_release_stabilization_gate.ps1`
  - `scripts/v100_release_stabilization_gate.ps1`
  - `scripts/v110_release_stabilization_gate.ps1`

You can also scaffold CI templates quickly:

```bash
repodoctor generate ci github
```

---

## Architecture

RepoDoctor is a modular monolith with strict boundaries.

High-level flow:

```text
detect language -> select adapter -> collect metrics -> build dependency graph -> execute rules -> score -> report
```

Key modules:

- `internal/analysis` → orchestration
- `internal/languages` → adapters + detection
- `internal/rules` + `internal/engine` → rule registry/execution
- `internal/model` + `internal/domain` → shared contracts and policies

For architecture notes:

- `docs/architecture.md`

---

## Project Structure

```text
RepoDoctor/
├── main.go
├── analysis_service.go
├── runtime_engine.go
├── config.go
├── reporter.go
├── generator.go
├── internal/
│   ├── analysis/
│   ├── languages/
│   ├── rules/
│   ├── engine/
│   ├── model/
│   └── domain/
├── scripts/
└── .github/workflows/
```

---

## Quality Gates

### Mandatory local gates

```bash
go test ./...
go vet ./...
go run . analyze -path .
```

Expected self-analysis result on this repo: **100/100**.

If you touch concurrency-sensitive areas, also run:

```bash
go test -race ./...
```

---

## Release & Roadmap Status

### Completed milestone lines

- v0.18 (M0): Contract hardening
- v0.19 (M1): Incremental + performance baseline
- v0.20 (M2): Java pilot (gated)
- v1.0 (M3): UX & polish
- v1.1+ (M4): Scale (use-case gated)

Recent release PRs:

- v1.0 release: `dev -> main` PR **#250**
- v1.1 release: `dev -> main` PR **#254**

Detailed release tracker:

- `docs/report.md`
- `docs/v1.0-rd-10005-release-stabilization.md`
- `docs/v1.1-rd-11003-release-gate.md`

---

## Contributing

1. Branch from `dev`
2. Keep scope focused (single issue/goal)
3. Run local quality gates
4. Open PR to `dev`
5. Merge to `main` only through green CI release flow

Recommended commit prefixes:

- `feat:`
- `fix:`
- `perf:`
- `docs:`
- `test:`
- `ci:`

---

## Privacy & Repository Hygiene

Do not publish local/private artifacts such as:

- `todo.md` (any location)
- AI planning/protocol notes
- debug artifacts (`debug/`, `*.debug`, `debug.log`, `*.trace`)

---

## License

MIT — see [LICENSE](LICENSE).
