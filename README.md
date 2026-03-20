# RepoDoctor

> Static architecture analysis for software repositories.

RepoDoctor is a CLI tool that analyzes architectural health (dependency structure, maintainability, and rule violations). It is **not** a style linter; it focuses on structural risks that grow over time.

![Version](https://img.shields.io/badge/version-v0.9.0--dev-blue)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## Quick Start

```bash
git clone https://github.com/AdemFurkanATA/RepoDoctor.git
cd RepoDoctor
go build .

# Analyze current directory
./RepoDoctor analyze -path .

# Analyze any external repository
./RepoDoctor analyze -path /path/to/repository
```

Windows PowerShell:

```powershell
go build .
.\RepoDoctor.exe analyze -path .
```

---

## Core Features

- Circular dependency detection
- Layer validation rules
- File/function size threshold analysis
- God object detection
- Structural health score (0–100)
- JSON output for CI pipelines
- Interactive mode, watch mode, progress bars, colored output

### Language Support

- **Go** (AST-based)
- **Python**
- **JavaScript / TypeScript**

Language support is adapter-based (`LanguageAdapter`) and extensible.

---

## Usage

### Analyze

```bash
# text output (default)
repodoctor analyze -path .

# JSON output
repodoctor analyze -path ./my-repo -format json

# verbose output
repodoctor analyze -path . -verbose

# watch mode
repodoctor analyze -path . -watch

# disable color
repodoctor analyze -path . -no-color
```

### Other Commands

```bash
repodoctor interactive
repodoctor extract -path . -module RepoDoctor
repodoctor history -path .
repodoctor generate rule my-custom-rule
repodoctor version
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0`  | No critical violations |
| `2`  | Critical violations detected |

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
```

---

## Architecture (High Level)

Pipeline:

```text
detect language -> select adapter -> detect files -> collect metrics -> build dependency graph -> execute rules -> score -> report
```

Main building blocks:

- `internal/languages/` → language adapters + language detection
- `internal/analysis/` → orchestration
- `internal/rules/` + `internal/engine/` → rule registry/execution
- `internal/model/` → metrics, graph, violation models

---

## Development

```bash
go test ./...
go vet ./...
go run . analyze -path .
```

---

## Contributing

Please use this workflow:

1. Branch from `dev`
2. Keep scope focused (one issue/feature per branch)
3. Run gates: `go test ./...`, `go vet ./...`, `go run . analyze -path .`
4. Open PR to `dev`
5. Merge to `main` only after CI is green

---

## Privacy & Repository Hygiene

The following are **local/private artifacts** and must not be published:

- `todo.md` (any location)
- AI planning/protocol files
- debug artifacts (`debug/`, `*.debug`, `debug.log`, `*.trace`)

These are ignored via `.gitignore`.

---

## License

MIT — see [LICENSE](LICENSE).
