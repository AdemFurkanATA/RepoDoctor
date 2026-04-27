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
| v1.2 | M5: v1.2 Test Quality & Coverage | Completed | RD-12001..RD-12006 merged into `dev`; root/model/rules coverage hardening, parser fuzz/property determinism tests, and v1.2 gate pack activated. |
| v1.3 | M6: v1.3 Observability | Completed | RD-13001..RD-13005 merged into `dev`; structured logging, safe profiling hooks, error taxonomy hardening, deterministic metrics export, and v1.3 gate pack activated. |
| v1.4 | M7: v1.4 Performance Hardening | Completed | RD-14001..RD-14004 merged into `dev`; parallel parsing tuning, memory budget fail-soft controls, filesystem-safe cache warmup, and benchmark/race gate automation activated. |
| v1.5 | M8: v1.5 Developer Experience | Completed | RD-15001..RD-15005 merged into `dev`; released to `main` via `dev -> main` PR #305. |

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

### v1.2 (M5: Test Quality & Coverage)

- **RD-12001**: `internal/model` package test expansion and deterministic coverage lock
- **RD-12002**: root package test expansion (coverage gate >=70%) + `json-v1` report path alignment
- **RD-12003**: `internal/rules` edge/boundary and threshold behavior coverage
- **RD-12004**: parser/extractor fuzz harnesses + malformed corpus fail-soft assertions
- **RD-12005**: property-based determinism invariants for ordering helpers
- **RD-12006**: v1.2 stabilization gate orchestration (`scripts/v120_release_stabilization_gate.ps1` + CI workflow wiring)

Release path:

- Feature PRs merged into `dev`: #282, #283, #284, #285, #286, #287
- Release PR (`dev -> main`): pending (create after v1.2 gate PR merges and CI remains green)

### v1.3 (M6: Observability)

- **RD-13001**: structured logging boundaries (`internal/logger`) with default-off behavior
- **RD-13002**: opt-in CPU/heap profiling hooks with root-bounded path safeguards
- **RD-13003**: stable error taxonomy (class+code), wrapped-context preservation, debug-gated details
- **RD-13004**: deterministic metrics export (`internal/metrics`) with default-off activation
- **RD-13005**: v1.3 stabilization gate orchestration (`scripts/v130_release_stabilization_gate.ps1` + CI workflow wiring)

Release path:

- Feature PRs merged into `dev`: #289, #290, #291, #292, #293
- Release PR (`dev -> main`): pending (create after v1.3 gate PR merges and CI remains green)

### v1.4 (M7: Performance Hardening)

- **RD-14001**: parallel parsing worker tuning with deterministic merge behavior
- **RD-14002**: fail-soft memory budget enforcement (`REPODOCTOR_MEMORY_FILE_BUDGET`)
- **RD-14003**: filesystem-safe incremental cache warmup (default-off)
- **RD-14004**: benchmark/race/memory stabilization gate (`scripts/v140_release_stabilization_gate.ps1` + CI wiring)

Release path:

- Feature PRs merged into `dev`: #295, #296, #297, #298
- Release PR (`dev -> main`): pending (create after v1.4 gate PR merges and CI remains green)

### v1.5 (M8: Developer Experience)

- **RD-15001**: actionable remediation hints in text/colored outputs
- **RD-15002**: default-off interactive remediation suggestions (confirmation-gated, advisory-only)
- **RD-15003**: Homebrew + Docker distribution foundation with tag-triggered GHCR workflow
- **RD-15004**: isolated VS Code extension skeleton under `vscode-repodoctor/` consuming CLI JSON output only
- **RD-15005**: v1.5 stabilization gate orchestration (`scripts/v150_release_stabilization_gate.ps1` + CI wiring)

Release path:

- Feature PRs merged into `dev`: #300, #301, #302, #303, #304
- Release PR (`dev -> main`): **#305**
- Release checklist: `docs/v1.5-release-pr-checklist.md`

## Current Branch/Release State

- Latest completed release line is synchronized on `main`; `dev` may temporarily run ahead with planning/docs-only deltas before the next release merge.
- `dev` ve `main` v1.5 release sonrası aynı committe hizalı; bir sonraki icra hattı v1.6 (M9) planı.
- Mandatory quality baseline remains green:
  - `go test ./...`
  - `go vet ./...`
  - `go run . analyze -path .` (target `100/100`)

## Next Planned Execution Line (Ready)

1. Create/validate milestones: **M9 v1.6**, **M10 v1.7**, **M11 v1.8**.
2. Open roadmap-governed issues: **RD-16001..RD-16009**, **RD-17001..RD-17006**, **RD-18001..RD-18004**.
3. Enforce workflow for every implementation issue:
   - branch-per-issue from `dev`
   - PR target `dev`
   - mandatory + conditional checks green
   - merge
4. After each milestone chain closes on `dev`, open release PR `dev -> main` and keep `dev` branch intact.

## Process/Gate Risk Watchlist

- **Open:** Release distribution workflow'unda manuel tetikleme (`workflow_dispatch`) ile release akışı bypass riski; tag/ref doğrulaması ve protection kontrolü sıkılaştırılmalı.
- **Mitigated (v1.7):** CI akışında stabilization script zinciri tek bir konsolide kapıya (`scripts/v170_release_stabilization_gate.ps1`) indirildi; mükerrer test/vet döngüleri azaltıldı.
- **Plan (CI maliyet notu):**
  - PR tarafında hızlı kapı (`repodoctor-fast.yml`) düşük maliyetli geri bildirim için birincil kalacak.
  - Tam matris kapısı (`repodoctor.yml`) konsolide stabilization gate ile korunacak.
  - Çalışma süresi/runner tüketimi 2 sprint boyunca izlenip eşik dışına taşarsa determinism alt testleri için ayrı gece-job ayrıştırması yapılacak.
- **Open:** Branch governance doğrulaması (required status checks / required review / up-to-date) repo ayarı seviyesinde izlenmeli; yalnız doküman kuralı yeterli değil.

Son güncelleme: 11 Nisan 2026 (roadmap v1.6+ ile hizalandı)
