# RepoDoctor 🏥

> **Static Architecture Intelligence for Go Repositories**

RepoDoctor is a CLI tool that analyzes your Go repository's architectural health by evaluating structure, dependency patterns, and maintainability signals. It doesn't lint your syntax—it inspects your engineering decisions.

[Go Version](https://go.dev/)
[License](LICENSE)
[Status](../../tree/main)

---

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/AdemFurkanATA/RepoDoctor.git
cd repodoctor

# Build
go build -o repodoctor.exe

# Run
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

## 🎯 Core Features (v0.1)

### Planned Capabilities

- **File Size Analysis** — Detect unusually large files
- **Function Size Heuristics** — Identify overly complex functions
- **Circular Import Detection** — Catch import cycles in Go packages
- **Layer Validation** — Enforce architectural boundaries
- **Repository Scoring** — Quantitative health metrics (0-100)
- **JSON Reports** — Machine-readable output for CI integration

---

## 📖 Usage (Planned)

```bash
# Analyze current directory
repodoctor analyze .

# Analyze with JSON output
repodoctor analyze ./my-project --format json

# Check specific rules
repodoctor check --rules=circular-imports,size

# Generate health report
repodoctor report --output=health.json
```

### Example Output

```
RepoDoctor v0.1.0
Analyzing: ./my-project

Architecture Health: B+ (85/100)
Maintainability Score: 78/100

Issues Found:
  ⚠ [CIRCULAR] internal/service ↔ internal/repo
  ⚠ [LARGE_FILE] user_handler.go (823 lines)

Checks Passed:
  ✓ Test coverage detected
  ✓ No god objects identified
  
Analysis completed in 234ms
```

---

## 🏗️ Architecture

RepoDoctor philosophy:

> **Clean architecture is not a folder structure. It is discipline.**

RepoDoctor enforces engineering discipline through:

1. **Structure Analysis** — Evaluates package organization
2. **Dependency Graph** — Maps import relationships
3. **Heuristic Rules** — Applies industry best practices
4. **Scoring System** — Quantifies architectural quality

---

## 🗺️ Roadmap

### v0.1 — Core Engine (Current)

- Project initialization
- CLI argument parsing
- Basic file analysis
- Architecture health scoring

### v0.2 — Rule Engine

- Circular import detection
- File/function size thresholds
- Layer violation rules
- Configurable rule sets

### v0.3 — Reporting & CI

- JSON/XML output formats
- GitHub Actions integration
- Custom thresholds
- Trend analysis

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
go test ./...
```

---

## 📁 Project Structure

```
repodoctor/
├── cmd/                 # CLI command definitions
├── internal/            # Core analysis engine
│   ├── analyzer/        # File and package analyzers
│   ├── rules/           # Architecture rule definitions
│   ├── scoring/         # Health scoring logic
│   └── report/          # Output formatters
├── pkg/                 # Public libraries
├── docs/                # Documentation (local only)
├── main.go              # Application entry point
├── go.mod               # Go module definition
└── README.md            # This file
```

> **Note:** The `docs/` directory contains local development documentation and is not committed to version control.

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

