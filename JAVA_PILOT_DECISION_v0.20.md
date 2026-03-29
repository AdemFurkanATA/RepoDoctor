# RepoDoctor v0.20 Java Pilot Decision Memo

## Scope

This memo records the go/no-go decision for the gated Java pilot after completing RD-2001 through RD-2006.

## Evidence Summary

- Java pilot contract is default-off (`language_detection.java_pilot_enabled`), preserving backward compatibility.
- Java adapter skeleton and extraction MVP are implemented with deterministic file discovery and import graph generation.
- Abuse-resistance bounds are in place:
  - max file bytes,
  - bounded scan/import rows,
  - NUL-safe line handling,
  - fail-soft behavior on malformed/oversized files.
- Java pilot fixture corpus and benchmark smoke are available under `internal/languages/testdata/java_pilot_fixture`.
- Core quality gates remain green:
  - `go test ./...`
  - `go vet ./...`
  - `go run . analyze -path .` (100/100)

## Decision

**GO (GATED)**

Proceed with Java support only behind the explicit pilot flag.

## Guardrails

1. Java stays disabled by default for all users.
2. Any performance or determinism regression blocks default enablement.
3. Rule engine remains language-agnostic (no Java-specific rule forks).
4. CI gates and stabilization packs must stay green on Linux and Windows.

## Follow-up

- Keep pilot telemetry/evidence collection active in upcoming iterations.
- Re-evaluate default enablement only after sustained stability across multiple releases.
