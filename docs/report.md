# RepoDoctor Version Report Tracker

Bu dosya sürüm bazlı ilerleme/rapor özetini tek yerde takip etmek için tutulur.

| Version | Milestone | Status | Notes |
|---|---|---|---|
| v0.17 | S8-v0.17 Near-Max Practical Maturity | Completed | Maturity report published (`docs/v0.17-maturity-report.md`). |
| v0.18 | M0: v0.18 Contract Hardening | Completed | RD-1801..RD-1808 merged into `dev`; close-out, path-safety, and CI supply-chain gates active. |
| v0.19 | M1: v0.19 Incremental & Performance | Completed | RD-1901..RD-1908 merged; release stabilization gate and benchmark+memory CI budgets active. |
| v0.20 | M2: v0.20 Java Pilot (Gated) | Completed (Gated GO) | RD-2001..RD-2007 completed; decision memo published (`JAVA_PILOT_DECISION_v0.20.md`), pilot remains default-off. |
| v1.0 | M3: v1.0 UX & Polish | Completed | RD-10001..RD-10005 merged to `dev` and released via `dev -> main` PR #250. |
| v1.1+ | M4: v1.1+ Scale (Use-case gated) | Completed | RD-11001..RD-11003 merged to `dev` and released via `dev -> main` PR #254. |

---

## Milestone Closeout Details (v1.0 / v1.1)

### v1.0 (M3: UX & Polish)

- **RD-10001**: violation message clarity improvements (parser/runtime message quality)
- **RD-10002**: rule severity configuration overrides
- **RD-10003**: custom architecture profiles (`custom_layer_order`, `custom_layer_keywords`)
- **RD-10004**: CI template generator (`generate ci <github|gitlab|azure> [--force]`)
- **RD-10005**: release stabilization + docs sync + v1.0 gate pack

Release path:

- Feature PRs merged into `dev`: #245, #246, #247, #248, #249
- Release PR (`dev -> main`): **#250**

### v1.1 (M4: Scale, use-case gated)

- **RD-11001**: durable incremental snapshot store foundations (file-backed store, fail-closed load behavior, atomic replace)
- **RD-11002**: memory optimization in fingerprint hot path (streaming hash + bounded/batched processing)
- **RD-11003**: dedicated v1.1 release gate orchestration (`scripts/v110_release_stabilization_gate.ps1` + CI workflow wiring)

Release path:

- Feature PRs merged into `dev`: #251, #252, #253
- Release PR (`dev -> main`): **#254**

## Current Branch/Release State

- `main` and `dev` are synchronized to latest completed release line.
- M3 and M4 issue chains are closed.
- Mandatory quality baseline remains green:
  - `go test ./...`
  - `go vet ./...`
  - `go run . analyze -path .` (target `100/100`)

Son güncelleme: 29 Mart 2026
