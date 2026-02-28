# RepoDoctor 🏥

> **Static Architecture Intelligence for Go Repositories**

RepoDoctor is a CLI tool that analyzes your Go repository's architectural health by evaluating structure, dependency patterns, and maintainability signals. It doesn't lint your syntax—it inspects your engineering decisions.

![Version](https://img.shields.io/badge/version-v0.2.0--dev-blue)
[![Go Version](https://img.shields.io/badge/go-1.25+-00ADD8)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

[Go Version](https://go.dev/)
[License](LICENSE)
[Status](../../tree/main)

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

## 🎯 Core Features (v0.2)

### Implemented Capabilities

- ✅ **Import Extraction** — AST-based Go import analysis with AST parsing
- ✅ **Dependency Graph** — Graph-based dependency mapping with cycle detection
- ✅ **Circular Dependency Detection** — DFS-based import cycle identification (critical severity)
- ✅ **Layer Validation** — Enforce handler → service → repo architecture (high severity)
- ✅ **Structural Scoring** — Maintainability score (0-100) with penalty weights
- ✅ **CLI Reports** — Beautiful text output and JSON export for CI integration
- ✅ **13 Unit Tests** — Comprehensive test coverage for all core components

---

## 📖 Usage

### Analyze Command

Analyze your repository for structural violations:

```bash
# Analyze current directory (text output)
repodoctor analyze -path .

# Analyze with JSON output
repodoctor analyze -path ./my-project -format json

# Verbose mode
repodoctor analyze -path . -verbose
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

Version: v0.2.0-dev
Path: C:\project

┌───────────────────────────────────────────────────────────┐
│  STRUCTURAL HEALTH SCORE                                  │
└───────────────────────────────────────────────────────────┘
✓ Score: 85.0 / 100.0

┌───────────────────────────────────────────────────────────┐
│  VIOLATIONS SUMMARY                                       │
└───────────────────────────────────────────────────────────┘
Total Violations: 3
  - Circular Dependencies: 1
  - Layer Violations: 2

┌───────────────────────────────────────────────────────────┐
│  CIRCULAR DEPENDENCIES [CRITICAL]                         │
└───────────────────────────────────────────────────────────┘
[1] project/service → project/repo → project/service

┌───────────────────────────────────────────────────────────┐
│  LAYER VIOLATIONS [HIGH]                                  │
└───────────────────────────────────────────────────────────┘
[1] project/repo/user_repo.go (repo) -> project/service/user_service.go (service): upward import not allowed

┌───────────────────────────────────────────────────────────┐
│  SCORE BREAKDOWN                                          │
└───────────────────────────────────────────────────────────┘
Base Score:           100.0
Circular Penalty:     -10.0 (1 violations x 10.0)
Layer Penalty:        -10.0 (2 violations x 5.0)
─────────────────────────────────────────────────
Final Score:          80.0
```

### Example JSON Output

```json
{
  "version": "v0.2.0-dev",
  "path": "C:\\project",
  "score": {
    "total": 80.00,
    "max": 100.00,
    "circularPenalty": 10.00,
    "layerPenalty": 10.00
  },
  "violations": {
    "circular": 1,
    "layer": 2
  },
  "circularViolations": [...],
  "layerViolations": [...]
}
```

---

## 🏗️ Architecture

RepoDoctor philosophy:

> **Clean architecture is not a folder structure. It is discipline.**

RepoDoctor enforces engineering discipline through:

1. **Import Extraction** — AST-based parsing of Go files
2. **Dependency Graph** — Adjacency list representation with DFS traversal
3. **Rule Engine** — Pluggable rule interface (CircularDependency, LayerValidation)
4. **Scoring System** — Weighted penalty calculation (circular: 10pts, layer: 5pts)
5. **Reporter** — Multi-format output (text with ASCII borders, JSON)

---

## 🗺️ Roadmap

### v0.1 — Core Engine ✅ (Completed)

- ✅ Project initialization
- ✅ CLI argument parsing
- ✅ Import extraction with AST
- ✅ Dependency graph construction

### v0.2 — Rule Engine ✅ (Current)

- ✅ Circular import detection (DFS-based)
- ✅ Layer violation rules (handler → service → repo)
- ✅ Structural scoring system
- ✅ Text and JSON output formats
- ✅ Comprehensive test suite (13 tests)

### v0.3 — Advanced Analysis (Planned)

- File/function size thresholds
- God object detection
- Custom rule configuration
- GitHub Actions integration
- Trend analysis over time

---

## 🛠️ Development

### Prerequisites

- Go 1.25 or higher
- Git

### Build from Source

```bash
git clone https://github.com/yourusername/repodoctor.git
cd repodoctor
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
├── scoring.go              # Structural scoring system
├── reporter.go             # Output formatter (text, JSON)
├── dependency_test.go      # Comprehensive test suite (13 tests)
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
=== RUN   TestDependencyGraphAcyclic
--- PASS: TestDependencyGraphAcyclic (0.00s)
=== RUN   TestDependencyGraphSimpleCycle
--- PASS: TestDependencyGraphSimpleCycle (0.00s)
=== RUN   TestDependencyGraphMultiNodeCycle
--- PASS: TestDependencyGraphMultiNodeCycle (0.00s)
=== RUN   TestLayerValidationRuleUpwardImport
--- PASS: TestLayerValidationRuleUpwardImport (0.00s)
=== RUN   TestLayerValidationRuleRepoToService
--- PASS: TestLayerValidationRuleRepoToService (0.00s)
=== RUN   TestStructuralScoringDeterministic
--- PASS: TestStructuralScoringDeterministic (0.00s)
PASS
ok      RepoDoctor      0.367s
```

All 13 tests pass with deterministic output.

