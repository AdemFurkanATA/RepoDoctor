# RepoDoctor 🏥

> **Static Architecture Intelligence for Go Repositories**

RepoDoctor is a CLI tool that analyzes your Go repository's architectural health by evaluating structure, dependency patterns, and maintainability signals. It doesn't lint your syntax—it inspects your engineering decisions.

![Version](https://img.shields.io/badge/version-v0.3.0--dev-blue)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Status](https://img.shields.io/badge/status-dev-yellow)](../../tree/main)

---

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/AdemFurkanATA/RepoDoctor.git
cd RepoDoctor

# Build
go build -o repodoctor.exe

# Run analysis
./repodoctor analyze -path . -format text

# Extract imports
./repodoctor extract -path . -module RepoDoctor

# Show help
./repodoctor --help
```

---

## 📋 Why RepoDoctor?

Most static analysis tools focus on **code style** and **formatting**. RepoDoctor focuses on **structural integrity**:


| Problem                            | RepoDoctor Solution             |
| ---------------------------------- | ------------------------------- |
| ❓ Are layers violating boundaries? | 🔍 Layer validation rules       |
| ❓ Circular dependencies?           | 🔄 Import cycle detection       |
| ❓ God objects emerging?            | 📊 Size heuristics analysis     |
| ❓ Technical debt accumulating?     | 📈 Maintainability scoring      |
| ❓ CI/CD quality gates missing?     | 🛡️ Architecture health reports |


---

## 🎯 Core Features (v0.3)

### Implemented Capabilities

- ✅ **Import Extraction** — AST-based Go import analysis
- ✅ **Dependency Graph** — Graph-based dependency mapping with cycle detection
- ✅ **Circular Dependency Detection** — DFS-based import cycle identification (critical severity)
- ✅ **Layer Validation** — Enforce handler → service → repo architecture (high severity)
- ✅ **Structural Scoring** — Maintainability score (0-100) with penalty weights
- ✅ **Size Threshold Analysis** — Detect oversized files (>500 lines) and functions (>80 lines)
- ✅ **God Object Detection** — Identify structs with too many fields (>15) or methods (>10)
- ✅ **Custom Configuration** — YAML-based config (`.repodoctor/config.yaml`) for rule thresholds
- ✅ **GitHub Actions Integration** — CI/CD workflow with automatic analysis and exit codes
- ✅ **Trend Analysis** — Historical score tracking with `.repodoctor/history.json`
- ✅ **CLI Reports** — Text output and JSON export for CI integration
- ✅ **35+ Unit Tests** — Comprehensive test coverage for all core components

---

## 📖 Usage

### Analyze Command

Analyze your repository for structural violations:

```bash
# Analyze current directory (text output)
repodoctor analyze -path .

# Analyze with JSON output
repodoctor analyze -path ./my-project -format json

# Verbose mode (shows trend analysis)
repodoctor analyze -path . -verbose

# With custom config
repodoctor analyze -path . -verbose
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

### Extract Command

Extract imports from Go files:

```bash
# Extract imports with module normalization
repodoctor extract -path . -module RepoDoctor
```

### Example Text Output

```
╔═══════════════════════════════════════════════════════════╗
║          RepoDoctor Structural Analysis Report           ║
╚═══════════════════════════════════════════════════════════╝

Version: v0.3.0-dev
Path: C:\project

┌───────────────────────────────────────────────────────────┐
│  STRUCTURAL HEALTH SCORE                                  │
└───────────────────────────────────────────────────────────┘
✓ Score: 78.0 / 100.0

┌───────────────────────────────────────────────────────────┐
│  VIOLATIONS SUMMARY                                       │
└───────────────────────────────────────────────────────────┘
Total Violations: 5
  - Circular Dependencies: 1
  - Layer Violations: 2
  - Size Violations: 2
  - God Objects: 0

┌───────────────────────────────────────────────────────────┐
│  CIRCULAR DEPENDENCIES [CRITICAL]                         │
└───────────────────────────────────────────────────────────┘
[1] project/service → project/repo → project/service

┌───────────────────────────────────────────────────────────┐
│  LAYER VIOLATIONS [HIGH]                                  │
└───────────────────────────────────────────────────────────┘
[1] project/repo/user_repo.go (repo) -> project/service/user_service.go (service): upward import not allowed

┌───────────────────────────────────────────────────────────┐
│  SIZE VIOLATIONS [LOW]                                    │
└───────────────────────────────────────────────────────────┘
[1] Function 'largeFunction' in handler.go: 120 lines (threshold: 80)
[2] File utils.go: 650 lines (threshold: 500)

┌───────────────────────────────────────────────────────────┐
│  SCORE BREAKDOWN                                          │
└───────────────────────────────────────────────────────────┘
Base Score:           100.0
Circular Penalty:     -10.0 (1 violations x 10.0)
Layer Penalty:        -10.0 (2 violations x 5.0)
Size Penalty:         -6.0 (2 violations x 3.0)
─────────────────────────────────────────────────
Final Score:          74.0
```

### Example JSON Output

```json
{
  "version": "v0.3.0-dev",
  "path": "C:\\project",
  "score": {
    "total": 74.00,
    "max": 100.00,
    "circularPenalty": 10.00,
    "layerPenalty": 10.00,
    "sizePenalty": 6.00,
    "godObjectPenalty": 0.00
  },
  "violations": {
    "circular": 1,
    "layer": 2,
    "size": 2,
    "godObject": 0
  },
  "circularViolations": [...],
  "layerViolations": [...],
  "sizeViolations": [...],
  "godObjectViolations": [...]
}
```

---

## 🏗️ Architecture

RepoDoctor philosophy:

> **Clean architecture is not a folder structure. It is discipline.**

RepoDoctor enforces engineering discipline through:

1. **Import Extraction** — AST-based parsing of Go files
2. **Dependency Graph** — Adjacency list representation with DFS traversal
3. **Rule Engine** — Pluggable rule interface (CircularDependency, LayerValidation, SizeRule, GodObjectRule)
4. **Configuration System** — YAML-based config with graceful defaults
5. **Scoring System** — Weighted penalty calculation (circular: 10pts, layer: 5pts, size: 3pts, god object: 5pts)
6. **Trend Analysis** — Historical score tracking with delta calculation
7. **Reporter** — Multi-format output (text with ASCII borders, JSON)

---

## 🗺️ Roadmap

### ✅ v0.1 — Core Engine (Completed)

**Goal:** Establish the analysis foundation.

- ✅ Project scaffolding
- ✅ Repository scanner
- ✅ Basic metrics collector
- ✅ Initial rule system
- ✅ Rule execution pipeline
- ✅ Basic scoring engine
- ✅ CLI command structure (`analyze`)
- ✅ Human-readable CLI output

### ✅ v0.2 — Dependency Intelligence (Completed)

**Goal:** Structural awareness of the repository.

- ✅ Import graph builder (Go)
- ✅ Circular dependency detection
- ✅ Layer validation rules
- ✅ Structural scoring adjustments
- ✅ Improved CLI reporting

### ✅ v0.3 — Advanced Analysis & Automation (Completed)

**Goal:** Introduce deeper analysis and automation capabilities.

- ✅ File size threshold detection
- ✅ Function size threshold detection
- ✅ God object detection heuristics
- ✅ Custom rule configuration (`.repodoctor/config.yaml`)
- ✅ GitHub Actions integration
- ✅ Trend analysis with historical scoring
- ✅ Internal state management (`.repodoctor/history.json`)

### 🚧 v0.4 — Rule Engine Evolution (Planned)

**Goal:** Transform the rule system into a scalable analysis engine.

- Rule interface standardization
- Rule registry system
- Rule categories
- Rule execution pipeline
- Standardized violation model
- Migration of existing rules to the new engine

### 🔮 v0.5 — Multi-Language Foundation (Planned)

**Goal:** Prepare RepoDoctor for multi-language analysis.

- Language abstraction layer
- Python repository scanner
- Python import dependency graph
- Python rule implementations
- Shared rule interface across languages
- Language-aware analysis pipeline

### 🔮 v0.6 — Cross-Language Analysis (Planned)

**Goal:** Expand the rule ecosystem across languages.

- JavaScript / TypeScript analysis
- Cross-language rule compatibility
- Unified dependency graph abstraction
- Shared rule categories across languages
- Expanded maintainability heuristics

### 🎯 v1.0 — Extensible Platform (Planned)

**Goal:** Product maturity and extensibility.

- Plugin-based rule system
- External rule packages
- Configurable architecture profiles
- Stable public API
- Official documentation
- Production-ready release

---

**Long-Term Vision:** RepoDoctor aims to become a structural quality gate for repositories, a CI-integrated architecture evaluator, and a developer tool used to maintain engineering discipline across multiple programming languages.

---

## 🚀 GitHub Actions Integration

RepoDoctor can be easily integrated into your CI/CD pipeline using GitHub Actions.

### Basic Workflow

Create `.github/workflows/repodoctor.yml`:

```yaml
name: RepoDoctor Analysis

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  repodoctor:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true

      - name: Install dependencies
        run: go mod download

      - name: Build RepoDoctor
        run: go build -o repodoctor

      - name: Run RepoDoctor analysis
        run: ./repodoctor analyze -path . -format text
```

### Exit Codes

RepoDoctor uses exit codes to indicate analysis results:

- `0` → No critical violations (success)
- `1` → Critical violations detected (failure)

This allows your CI pipeline to fail automatically when architectural violations are found.

### Advanced Configuration

For custom thresholds and rule configuration, create `.repodoctor/config.yaml`:

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

### JSON Output for Further Processing

```yaml
- name: Run RepoDoctor (JSON)
  run: ./repodoctor analyze -path . -format json -verbose
  
- name: Upload analysis results
  uses: actions/upload-artifact@v4
  with:
    name: repodoctor-report
    path: repodoctor-report.json
```

---

## 🛠️ Development

### Prerequisites

- Go 1.25 or higher
- Git

### Build from Source

```bash
git clone https://github.com/AdemFurkanATA/RepoDoctor.git
cd RepoDoctor
go build -o repodoctor.exe
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -v -cover ./...
```

---

## 📁 Project Structure

```
RepoDoctor/
├── main.go                 # CLI entry point (analyze, extract, version commands)
├── import_extractor.go     # AST-based import extraction
├── dependency_graph.go     # Graph data structure with cycle detection
├── circular_rule.go        # Circular dependency rule (critical severity)
├── layer_rule.go           # Layer validation rule (high severity)
├── size_rule.go            # File/function size threshold analysis
├── god_object_rule.go      # God object detection (fields/methods)
├── config.go               # YAML configuration system
├── trend_analyzer.go       # Historical score tracking
├── scoring.go              # Structural scoring system
├── reporter.go             # Output formatter (text, JSON)
├── dependency_test.go      # Comprehensive test suite (35+ tests)
├── docs/                   # Documentation
│   ├── specs/              # Feature specifications
│   ├── architecture.md     # Architecture overview
│   └── roadmap.md          # Development roadmap
├── go.mod                  # Go module definition
└── README.md               # This file
```

---

## 🤝 Contributing

Contributions are welcome! This project is in early development.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📜 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

## 📬 Acknowledgments

Inspired by the need for architectural discipline in growing codebases. Built with ❤️ for Go developers who care about maintainability.

---

**RepoDoctor** — *Enforcing engineering discipline, one repository at a time.*

---

## 📊 Test Coverage

```bash
$ go test -v ./...
=== RUN   TestConfigLoader_DefaultConfig
--- PASS: TestConfigLoader_DefaultConfig (0.00s)
=== RUN   TestSizeRule_DetectLargeFile
--- PASS: TestSizeRule_DetectLargeFile (0.02s)
=== RUN   TestGodObjectRule_DetectManyFields
--- PASS: TestGodObjectRule_DetectManyFields (0.01s)
=== RUN   TestTrendAnalyzer_AppendScore
--- PASS: TestTrendAnalyzer_AppendScore (0.00s)
=== RUN   TestDependencyGraphSimpleCycle
--- PASS: TestDependencyGraphSimpleCycle (0.00s)
=== RUN   TestLayerValidationRuleUpwardImport
--- PASS: TestLayerValidationRuleUpwardImport (0.00s)
=== RUN   TestStructuralScoringDeterministic
--- PASS: TestStructuralScoringDeterministic (0.00s)
PASS
ok      RepoDoctor      0.892s
```

All 35+ tests pass with deterministic output across all v0.3 features.

