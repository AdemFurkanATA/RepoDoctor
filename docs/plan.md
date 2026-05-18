# RepoDoctor Gelecek Planlama Notları

**Oluşturma Tarihi:** 29 Mart 2026  
**Son Güncelleme:** 27 Nisan 2026 (v1.8 release sonrası güncellendi)  
**Kapsam:** v0.18 → v2.0+ yol haritası için düşünce notları  
**Durum:** v1.8 Tamamlandı, v2.0+ planlama aşamasında

---

## 📌 Özet

Bu belge, RepoDoctor’un ileri seviye (v2.0+) yeteneklerini **CLI-first** yaklaşımını koruyarak nasıl genişleteceğini özetler. Aktif issue takibi için kaynak dosyalar:

- `docs/roadmap.md`
- `docs/report.md`

---

## 🎯 Felsefe

### Temel Prensipler

1. **CLI tool olarak kal** — API server, web UI, database gibi enterprise bloat'a girme.
2. **Deterministic & reliable** — Aynı input = aynı output.
3. **Default-safe/default-off** — riskli özellikler opt-in olmalı.
4. **Modüler monolith sınırları korunur** — orchestration/language/rules/model ayrımı.

### Neler DEĞİL

- ❌ API server
- ❌ Web UI
- ❌ Team dashboard (CLI-first ilkeye ters)

---

## 📊 Mevcut Durum (v1.8 sonrası)

### Başarılar ✅

| Milestone | Durum | İçerik |
|-----------|-------|--------|
| v0.17–v1.1 | ✅ Completed | contract hardening, incremental/perf, java pilot, UX/polish, scale |
| v1.2 | ✅ Completed | test quality & coverage hardening |
| v1.3 | ✅ Completed | observability guardrails |
| v1.4 | ✅ Completed | performance hardening |
| v1.5 | ✅ Completed | developer experience improvements |
| v1.6 | ✅ Completed | quality & security foundation |
| v1.7 | ✅ Completed | risk intelligence (API/churn/hotspot/vuln/docs) |
| v1.8 | ✅ Completed | safe auto-fix + IDE expansion plan |

### 9.3/10 → 9.5/10 Gap Analizi (özet)

| Kategori | Mevcut | Hedef | Gap |
|----------|--------|-------|-----|
| Architecture | 9/10 | 9/10 | ✅ |
| Code Quality | 9/10 | 9/10 | ✅ |
| Testing | 9/10 | 9/10 | ✅ |
| Performance | 9/10 | 9/10 | ✅ |
| Observability | 9/10 | 9/10 | ✅ |
| Advanced Intelligence | 6/10 | 9/10 | -3 🔴 |

**Öncelik:** Advanced intelligence alanını **gated** ve **opt-in** biçimde açmak.

---

## 🚀 v2.0+ Roadmap — Advanced Intelligence (Gated)

**Hedef:** 9.3/10 → 9.5/10  
**Fokus:** Pattern/anomali analizi, risk imza tespiti, trend tabanlı uyarılar  
**Süre:** 3–4 hafta / minor version  
**Toplam Issue:** 5–7 (1–2 Epic)

### Epic 1 — Advanced Intelligence (RD-19001 → RD-19005)

#### RD-19001: Violation Pattern Recognition

- Deterministic signature export (opt-in)
- Pattern confidence skoru + guardrails

#### RD-19002: Anomaly Detection (Gated)

- Tarihsel trend kalite eşiği
- Default-off, fail-soft

#### RD-19003–RD-19005: Research Lane

- Multi-repo analysis / compliance packs / team dashboard yalnızca araştırma
- CLI-first ilkeye aykırı olanlar **defer**

---

## 📌 Notlar

- Bu belge “büyük resim” planıdır; **aktif issue takibi** `docs/roadmap.md` ve `docs/report.md` üzerinden yürür.
- v1.2–v1.8 dönemi **tamamlandı** ve **legacy plan** olarak arşivlendi.
- v2.0+ intelligent features **opt-in** ve **gated** ilerler.
