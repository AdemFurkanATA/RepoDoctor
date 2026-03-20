# RepoDoctor

> **Static Architecture Analysis for Software Repositories**

RepoDoctor is a CLI tool that analyzes your repository's architectural health by evaluating structure, dependency patterns, and maintainability signals. It doesn't lint your syntax — it inspects your engineering decisions.

![Version](https://img.shields.io/badge/version-v0.8.0-blue)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Self-Analysis](https://img.shields.io/badge/self--analysis-100%2F100-brightgreen)]()
[![Tests](https://img.shields.io/badge/tests-75%20passing-brightgreen)]()

---

## Quick Start

```bash
# Clone the repository
git clone https://github.com/AdemFurkanATA/RepoDoctor.git
cd RepoDoctor

# Build
go build .

# Run analysis on any repository
./RepoDoctor analyze -path /path/to/your/project

# Run analysis on current directory
./RepoDoctor analyze -path .
```

### What You Get

```
Scanning repository [████████████████████] 100%
Collecting metrics [████████████████████] 100%
Building dependency graph [████████████████████] 100%
Running rules [████████████████████] 100%

╔═══════════════════════════════════════════════════════════╗
║          RepoDoctor Structural Analysis Report           ║
╚═══════════════════════════════════════════════════════════╝

Version: 0.5.0-dev
Path: /your/project

┌───────────────────────────────────────────────────────────┐
│  STRUCTURAL HEALTH SCORE                                  │
└───────────────────────────────────────────────────────────┘
✓ Score: 100.0 / 100.0

┌───────────────────────────────────────────────────────────┐
│  VIOLATIONS SUMMARY                                       │
└───────────────────────────────────────────────────────────┘
✓ No violations detected
✨ No structural violations detected! Your architecture is clean.
```

---

## Why RepoDoctor?

Most static analysis tools focus on **code style** and **formatting**. RepoDoctor focuses on **structural integrity** — the kind of problems that compound over time and make codebases unmaintainable.

| Problem | RepoDoctor Solution |
|---------|---------------------|
| Are layers violating boundaries? | Layer validation rules |
| Circular dependencies forming? | DFS-based import cycle detection |
| God objects emerging? | Struct field/method count heuristics |
| Files growing too large? | Size threshold analysis |
| Technical debt accumulating? | Maintainability scoring (0–100) |
| Need CI quality gates? | Exit codes + JSON output for automation |

RepoDoctor eats its own dog food — it analyzes itself and currently scores **100/100** with zero violations.

---

## Features

### Analysis Engine

- **Circular Dependency Detection** — DFS-based import cycle identification with critical severity
- **Layer Validation** — Enforce `handler → service → repo` architecture boundaries
- **Size Threshold Analysis** — Detect oversized files (>500 lines) and functions (>80 lines)
- **God Object Detection** — Identify structs with too many fields (>15) or methods (>10)
- **Structural Scoring** — Maintainability score (0–100) with weighted penalties
- **Trend Analysis** — Historical score tracking across runs

### Multi-Language Support

- **Go** — Full AST-based analysis (imports, functions, structs, dependency graph)
- **Python** — Import analysis, class/function metrics, dependency graph
- **Extensible Adapter Architecture** — Add new languages by implementing `LanguageAdapter` interface

### Developer Experience

- **Interactive Mode** — Guided CLI for analysis workflows
- **Watch Mode** — Continuous analysis with filesystem monitoring
- **Progress Bars** — Visual progress indicators for each pipeline stage
- **Colored Output** — Severity-based color coding with `--no-color` support
- **Rule Template Generator** — Scaffold custom rules from CLI
- **Structured Error Handling** — Actionable error messages with suggestions

### CI/CD Integration

- **Deterministic Exit Codes** — `0` (clean), `2` (critical violations)
- **JSON Output** — Machine-readable reports for pipeline integration
- **GitHub Actions** — Ready-to-use workflow configuration
- **Custom Configuration** — YAML-based rule thresholds

---

## Usage

### Analyze Command

```bash
# Analyze with text output (default)
repodoctor analyze -path .

# Analyze with JSON output
repodoctor analyze -path ./my-project -format json

# Verbose mode (includes trend analysis)
repodoctor analyze -path . -verbose

# Watch mode — re-analyze on file changes
repodoctor analyze -path . -watch

# Disable colored output
repodoctor analyze -path . -no-color
```

### Interactive Mode

```bash
repodoctor interactive
```

Provides a guided menu for:
- Running analysis on a repository
- Viewing analysis history
- Configuring rule thresholds

### Import Extraction

```bash
repodoctor extract -path . -module RepoDoctor
```

### Rule Template Generator

```bash
# Generate a custom rule template
repodoctor generate rule large-interface
# Creates: rules/large_interface_rule.go
```

### View History

```bash
repodoctor history
```

### Configuration

Create `.repodoctor/config.yaml` to customize thresholds:

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
```

### JSON Output

```json
{
  "version": "0.5.0-dev",
  "path": "/your/project",
  "score": {
    "total": 100.00,
    "max": 100.00,
    "circularPenalty": 0.00,
    "layerPenalty": 0.00,
    "sizePenalty": 0.00,
    "godObjectPenalty": 0.00
  },
  "violations": {
    "circular": 0,
    "layer": 0,
    "size": 0,
    "godObject": 0
  }
}
```

---

## Architecture

> **Clean architecture is not a folder structure. It is discipline.**

RepoDoctor is built on SOLID principles with a clear separation of concerns:

```
┌─────────────────────────────────────────────────────┐
│                    CLI Layer                         │
│         main.go / cli_commands.go                   │
│    (command parsing, output, exit codes)             │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│               Orchestration Layer                    │
│     analysis_service.go / runtime_engine.go         │
│     (pipeline coordination, report building)         │
└──────────────────────┬──────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────┐
│                 Internal Core                        │
│                                                      │
│  internal/analysis/    → Pipeline orchestrator       │
│  internal/rules/       → Rule engine (registry +     │
│                          executor pattern)            │
│  internal/languages/   → Language adapters            │
│                          (Go, Python)                 │
│  internal/model/       → Domain models (metrics,     │
│                          violations, dep graph)       │
│  internal/engine/      → Rule execution engine       │
└─────────────────────────────────────────────────────┘
```

### Analysis Pipeline

```
detect language → select adapter → scan files → collect metrics → build dependency graph → execute rules → score → report
```

### Key Design Decisions

- **Adapter Pattern** for language support — new languages don't require core changes
- **Registry + Executor** for rules — pluggable, sorted, deterministic execution
- **Package-qualified keys** for god object detection — prevents cross-package name collisions
- **Regex-based violation parsing** for accurate report mapping

---

## Project Structure

```
RepoDoctor/
├── main.go                     # CLI entry point, analyze pipeline orchestration
├── cli_commands.go             # Secondary CLI commands (scan, report, history, etc.)
├── analysis_service.go         # AnalysisService — full analyze pipeline coordinator
├── runtime_engine.go           # Bridges internal rule engine to legacy report format
│
├── internal/
│   ├── analysis/               # Pipeline orchestrator
│   │   └── orchestrator.go     # Detect → adapt → metrics → graph → rules
│   ├── rules/                  # Unified rule engine (the active path)
│   │   ├── rule.go             # Rule interface, AnalysisContext, domain types
│   │   ├── registry.go         # RuleRegistry with sorted GetAll()
│   │   ├── init.go             # Default registry initialization
│   │   ├── size_rule.go        # File/function size thresholds
│   │   ├── god_object_rule.go  # God object detection (package-qualified keys)
│   │   ├── circular_dependency_rule.go
│   │   └── layer_validation_rule.go
│   ├── engine/
│   │   └── executor.go         # RuleExecutor with panic recovery
│   ├── languages/
│   │   ├── language_adapter.go # LanguageAdapter interface
│   │   ├── language_detector.go
│   │   ├── go_adapter.go       # Go AST-based analysis
│   │   └── python_adapter.go   # Python import/class/function analysis
│   └── model/
│       ├── dependency_graph.go # DependencyGraph (10 methods)
│       ├── graph_cycle_detector.go  # GraphCycleDetector (extracted)
│       ├── graph_analysis.go   # FindRoots, FindLeaves (extracted)
│       ├── violation.go        # Violation with Severity, ScoreImpact
│       └── metrics.go          # Repository/File/Function/Struct metrics
│
├── scoring.go                  # Structural scoring system
├── config.go                   # YAML configuration system
├── reporter.go                 # Output formatter (text, JSON)
├── reporter_methods.go         # Report section writers
├── colored_methods.go          # Colored output section writers
├── color.go                    # ANSI color formatter, terminal detection
├── progress.go                 # Progress bar for pipeline stages
├── watcher.go                  # Filesystem watcher for watch mode
├── interactive.go              # Interactive CLI mode
├── interactive_session.go      # Interactive session management
├── generator.go                # Rule template generator
├── errors.go                   # Structured error system with suggestions
├── trend_analyzer.go           # Historical score tracking
├── import_extractor.go         # AST-based Go import extraction
│
├── .repodoctor/                # Runtime state (gitignored)
│   ├── config.yaml             # User configuration
│   └── history.json            # Score history
├── .github/workflows/
│   └── repodoctor.yml          # CI workflow
├── specs/
│   └── todo.md                 # Sprint planning and backlog
├── go.mod                      # Go 1.21, module RepoDoctor
└── README.md
```

---

## GitHub Actions Integration

Create `.github/workflows/repodoctor.yml`:

```yaml
name: RepoDoctor Analysis

on:
  push:
    branches: [main, dev]
  pull_request:
    branches: [main, dev]

jobs:
  repodoctor:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true

      - name: Install dependencies
        run: go mod download

      - name: Build RepoDoctor
        run: go build .

      - name: Run structural analysis
        run: ./RepoDoctor analyze -path . -format text
```

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | No critical violations — analysis passed |
| `2` | Critical violations detected — pipeline should fail |

### JSON Output for CI Pipelines

```yaml
- name: Run RepoDoctor (JSON)
  run: ./repodoctor analyze -path . -format json > repodoctor-report.json

- name: Upload analysis results
  uses: actions/upload-artifact@v4
  with:
    name: repodoctor-report
    path: repodoctor-report.json
```

---

## Scoring System

RepoDoctor calculates a **Structural Health Score** (0–100) based on weighted penalties:

| Rule | Severity | Penalty per Violation |
|------|----------|----------------------|
| Circular Dependency | Critical | -10 points |
| Layer Violation | High | -5 points |
| God Object | High | -5 points |
| File/Function Size | Medium | -3 points |

The score starts at 100 and decreases with each violation. A score of **100** means zero architectural violations.

---

## Development

### Prerequisites

- Go 1.21 or higher
- Git

### Build from Source

```bash
git clone https://github.com/AdemFurkanATA/RepoDoctor.git
cd RepoDoctor
go build .
```

### Run Tests

```bash
# Run all tests (75 tests across 5 packages)
go test ./...

# Verbose output
go test -v ./...

# With coverage
go test -v -cover ./...

# Static analysis
go vet ./...
```

### Self-Analysis

RepoDoctor analyzes its own codebase:

```bash
go run . analyze -path .
# Score: 100.0 / 100.0
# No violations detected
```

### Mandatory v0.9 Merge Gates

Before merging any issue PR to `dev`, run locally:

```bash
go test ./...
go vet ./...
go run . analyze -path .  # must remain 100/100
```

If the issue touches concurrency/shared-state paths (`internal/languages`, `internal/rules`, `internal/engine`, `internal/analysis`), also run:

```bash
go test -race ./...
```

Workflow policy: one issue = one branch, separate commit(s), separate push, separate PR to `dev`.

---

## Roadmap

### Completed

| Version | Theme | Highlights |
|---------|-------|------------|
| **v0.1** | Core Engine | Repository scanner, basic rule system, scoring engine, CLI |
| **v0.2** | Dependency Intelligence | Import graph builder, circular dependency detection, layer validation |
| **v0.3** | Advanced Analysis | Size thresholds, god object detection, YAML config, GitHub Actions, trend analysis |
| **v0.4** | Rule Engine v2 | Rule interface standardization, registry system, categories, execution pipeline |
| **v0.5** | Multi-Language Foundation | Language adapter architecture, Python support, plugin system, JSON output |
| **v0.6** | CLI & DX | Interactive mode, progress bars, colored output, watch mode, rule templates, error handling |
| **v0.7** | Architecture Hardening | Adapter-based pipeline, unified rule engine, language detector integration |
| **v0.8** | Perfect Score | God object elimination, report accuracy fixes, self-analysis 100/100 |

### v0.8 Sprint Details

The v0.8 sprint focused on eliminating all self-analysis violations:

| Issue | Change | Score Impact |
|-------|--------|-------------|
| RD-709 | Fixed god object cross-package name collision | 67 → 82 |
| RD-708 | Fixed report mapping with regex-based parsers | Accuracy fix |
| RD-710 | Refactored `DependencyGraph` (14 → 10 methods) | 82 → 87 |
| RD-711 | Refactored `GoAdapter` (12 → 9 methods) | 87 → 92 |
| RD-712 | Refactored `PythonAdapter` (13 → 10 methods) | 92 → 97 |
| RD-713 | Extracted CLI commands from `main.go` (567 → 485 lines) | 97 → **100** |

### Planned

| Version | Theme | Goals |
|---------|-------|-------|
| **v0.9** | Expansion | JavaScript/TypeScript adapter, enriched JSON reports, build-time versioning |
| **v1.0** | Platform | Plugin-based rule system, configurable architecture profiles, stable public API |

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch from `dev` (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`) and vet is clean (`go vet ./...`)
5. Commit with conventional messages (`feat:`, `fix:`, `refactor:`, `docs:`)
6. Push to your branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request targeting `dev`

### Branch Strategy

- `main` — stable releases
- `dev` — integration branch for features
- `feature/*`, `refactor/*` — individual work branches

---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

**RepoDoctor** — *Enforcing engineering discipline, one repository at a time.*
