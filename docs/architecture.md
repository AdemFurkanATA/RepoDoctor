# RepoDoctor Architecture (v1.0 Stabilization Snapshot)

This document reflects the currently enforced architecture in the repository.

## Architectural Style

RepoDoctor is a **modular monolith CLI** with explicit package boundaries and deterministic analysis behavior.

## Module Responsibilities

- **`main` (root package)**
  - CLI command parsing/composition
  - runtime wiring and report emission
- **`internal/analysis`**
  - orchestration of language detection + adapter execution
- **`internal/languages`**
  - language adapters and detection policy logic
- **`internal/rules` + `internal/engine`**
  - rule implementations, registry, execution pipeline
- **`internal/model` + `internal/domain`**
  - domain policies and shared data structures

## Dependency Direction

High-level direction is preserved as:

`main -> internal/analysis -> internal/languages -> internal/rules/internal/engine -> internal/model/internal/domain`

Boundary expectations are continuously validated by architectural boundary tests.

## Runtime Rule Context Flow

1. Config is loaded from `.repodoctor/config.yaml`
2. Runtime context is built with deterministic ordering and selected language evidence
3. Rules are executed through the internal engine
4. Violations are normalized and rendered in deterministic report order

## v1.0 Capability Notes

- Rule-level severity overrides are supported via `rules.*_severity`
- Architecture profile supports `clean`, `layered`, `modular-monolith`
- Custom architecture policy supports:
  - `architecture.custom_layer_order`
  - `architecture.custom_layer_keywords`
- CI template scaffolding is available via `generate ci <github|gitlab|azure> [--force]`

## Determinism and Safety Invariants

- Deterministic ordering for graph/rule/report output is mandatory
- Analyzer does not execute repository code and does not require network access
- Path handling remains canonicalized and root-bounded
- Release gates must keep `go test`, `go vet`, and self-analysis (`100/100`) green
