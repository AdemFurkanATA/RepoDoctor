# RepoDoctor v0.6 Upgrade Gap Analizi ve Uygulama Planı

## Kapsam

Bu doküman, `docs/specs/v0.6` altındaki tüm upgrade maddeleri için mevcut kod tabanındaki karşılıkları denetler, net gap’leri çıkarır ve her madde için **ayrı branch + ayrı commit + ayrı push** çalışma planını tanımlar.

## Güvenli ve Ölçeklenebilir Git Uygulama Protokolü

- Force push **yasak** (`git push --force` / `--force-with-lease` kullanılmayacak).
- Destructive komutlar **yasak** (`git reset --hard`, `git clean -fd`, geçmişi yeniden yazan işlemler vb.).
- Branch açma güvenli akışı:
  - Yeni branch: `git switch -c <branch-name>`
  - Branch zaten varsa: `git switch <branch-name>`
- Push standardı:
  - İlk push: `git push -u origin <branch-name>`
  - Sonraki güncellemeler: `git push`
- Her commit öncesi minimum kalite kapısı:
  - İlgili testler + lint/format kontrolleri çalıştırılacak.
  - Secret sızıntısı riski için staged dosyalar gözden geçirilecek (`.env`, token, private key vb. commit edilmeyecek).
  - `git add .` yerine hedefli stage kullanılacak (`git add -A -- <dosya/klasör>`); commit öncesi `git diff --cached --name-only` ile son kontrol yapılacak.
- PR öncesi branch güncelliği korunacak (main ile senkron kontrolü yapılıp çatışmalar branch içinde çözülecek).
- Bağımsız maddeler paralel geliştirilebilir; merge sırası aşağıdaki "Genel uygulama sırası"na göre korunmalıdır.

---

## v0.6-23 — Interactive CLI Mode

- **ID/Başlık:** 23 — Interactive CLI Mode
- **Kaynak dosya:** `docs/specs/v0.6/23-interactive-mode.md`
- **Durum:** Partial

### Gap açıklaması
- `repodoctor interactive` komutu var ve prompt tabanlı akış çalışıyor.
- “Analyze repository” ve “View analysis history” menüleri mevcut.
- Ancak “Configure rules” menüsü işlevsel değil; sadece bilgi mesajı veriyor.
- Spec’teki “guided workflows” hedefi kısmen karşılanıyor, ama konfigurasyon adımı tamamlanmamış.

### Önerilen branch adı
`feature/v0.6-23-interactive-mode-complete-workflow`

### Commit mesaj taslağı
`feat(interactive): complete guided workflow for rule configuration and validation`

### Push notu
- Bu madde için tek odaklı commit sonrası:
  - `git push -u origin feature/v0.6-23-interactive-mode-complete-workflow`

### Kabul kriterleri
- [ ] `repodoctor interactive` akışında analyze + history + configure menüleri gerçek işleve bağlı
- [ ] Geçersiz girişler deterministik ve kullanıcı dostu ele alınıyor
- [ ] Interactive akış mevcut CLI komutlarıyla çakışmadan yeniden kullanım yapıyor

---

## v0.6-24 — Progress Indicators

- **ID/Başlık:** 24 — Progress Indicators
- **Kaynak dosya:** `docs/specs/v0.6/24-progress-bars.md`
- **Durum:** Partial

### Gap açıklaması
- Progress bar altyapısı (`progress.go`) mevcut.
- `runAnalyze` içinde progress kullanımı var; fakat “metrics collection” stage’i görünür/ayrı değil.
- Adım sayıları gerçek iş yüküne bağlı değil (hardcoded), bu nedenle yüzde ilerlemesi doğruluk açısından zayıf.

### Önerilen branch adı
`feature/v0.6-24-progress-indicators-stage-accuracy`

### Commit mesaj taslağı
`feat(progress): add explicit metrics stage and improve progress accuracy per pipeline step`

### Push notu
- Bu madde ayrı commit/push olmalı:
  - `git push -u origin feature/v0.6-24-progress-indicators-stage-accuracy`

### Kabul kriterleri
- [ ] Scanning / Metrics / Dependency Graph / Rules stage’leri ayrı ve etiketli gösteriliyor
- [ ] Progress değerleri stage bazlı deterministik artıyor
- [ ] Verbose ve normal mod çıktıları okunabilir kalıyor

---

## v0.6-25 — Colored CLI Output

- **ID/Başlık:** 25 — Colored CLI Output
- **Kaynak dosya:** `docs/specs/v0.6/25-colored-output.md`
- **Durum:** Partial

### Gap açıklaması
- Renk formatlayıcı (`color.go`) ve colored report metodları var.
- `--no-color` bayrağı mevcut.
- Ancak terminal uyumluluğu tespiti gerçek değil (`isTerminal()` her zaman `true`), ANSI desteklemeyen ortamlarda çıktı bozulabilir.

### Önerilen branch adı
`feature/v0.6-25-colored-output-terminal-compat`

### Commit mesaj taslağı
`fix(color): detect terminal capabilities and disable ANSI when unsupported`

### Push notu
- Tek sorumluluklu commit sonrası push:
  - `git push -u origin feature/v0.6-25-colored-output-terminal-compat`

### Kabul kriterleri
- [ ] INFO/WARN/ERROR/SUCCESS renk eşlemeleri tutarlı
- [ ] `--no-color` tüm text çıktılarında ANSI’yi kapatıyor
- [ ] TTY olmayan/uyumsuz terminalde otomatik fallback var

---

## v0.6-26 — Watch Mode

- **ID/Başlık:** 26 — Watch Mode
- **Kaynak dosya:** `docs/specs/v0.6/26-watch-mode.md`
- **Durum:** Partial

### Gap açıklaması
- `repodoctor analyze --watch` akışı ve fsnotify tabanlı watcher var.
- Debounce uygulanmış.
- Kritik boşluk: `runAnalyze` içinde `os.Exit(...)` kullanımı, watch akışında ihlal çıktığında süreci sonlandırabilir; bu durumda “continuous analysis” kırılır.
- Yeni oluşturulan alt klasörlerin dinamik watcher eklenmesi garanti değil.

### Önerilen branch adı
`feature/v0.6-26-watch-mode-non-terminating-loop`

### Commit mesaj taslağı
`fix(watch): keep watch loop alive on violations and improve dynamic directory watching`

### Push notu
- Bu maddeyi diğerlerinden izole edin:
  - `git push -u origin feature/v0.6-26-watch-mode-non-terminating-loop`

### Kabul kriterleri
- [ ] Dosya değişimi yeni analiz tetikliyor
- [ ] İhlal olsa bile watch loop kapanmıyor
- [ ] Debounce tekrar analiz fırtınasını engelliyor
- [ ] Yeni klasörler için watcher kapsamı güncelleniyor

---

## v0.6-27 — Rule Template Generator

- **ID/Başlık:** 27 — Rule Template Generator
- **Kaynak dosya:** `docs/specs/v0.6/27-rule-template-generator.md`
- **Durum:** Partial

### Gap açıklaması
- `repodoctor generate rule <rule-name>` komutu ve dosya üretimi var.
- Ancak üretilen template, mevcut iç mimariyle tam uyumlu değil ve derlenebilirlik kriterini garanti etmiyor (tip/import uyuşmazlığı riski yüksek).
- Spec’in “Generated files compile” başarı kriteri güvenilir şekilde sağlanmıyor.

### Önerilen branch adı
`feature/v0.6-27-rule-template-generator-compile-safe`

### Commit mesaj taslağı
`fix(generator): generate compile-safe rule templates aligned with current rule interfaces`

### Push notu
- Bu madde için ayrı commit/push zorunlu:
  - `git push -u origin feature/v0.6-27-rule-template-generator-compile-safe`

### Kabul kriterleri
- [ ] Üretilen `rules/*_rule.go` dosyası doğrudan derlenebiliyor
- [ ] Template mevcut rule interface sözleşmesine uyuyor
- [ ] En az bir örnek generation için otomatik test eklenmiş

---

## v0.6-28 — CLI Error Improvements

- **ID/Başlık:** 28 — CLI Error Improvements
- **Kaynak dosya:** `docs/specs/v0.6/28-cli-error-improvements.md`
- **Durum:** Partial

### Gap açıklaması
- Yapısal hata modeli (`CLIError`) ve öneri altyapısı mevcut.
- Ancak kullanım kapsamı parçalı: birçok komut hâlâ düz `fmt.Fprintf` ile ham hata basıyor.
- “Did you mean” benzeri öneri sistemi komut/rule seviyesinde sistematik ve merkezi değil.

### Önerilen branch adı
`feature/v0.6-28-cli-errors-unified-handling`

### Commit mesaj taslağı
`refactor(cli-errors): unify structured error handling across all commands with suggestions`

### Push notu
- Ayrı commit/push:
  - `git push -u origin feature/v0.6-28-cli-errors-unified-handling`

### Kabul kriterleri
- [ ] analyze/extract/report/history/generate yolları tek hata formatından geçiyor
- [ ] Error kategorileri net (usage/config/analysis/runtime)
- [ ] Uygun durumlarda öneri veya “did you mean” gösteriliyor

---

## Genel uygulama sırası (önerilen)

1. **v0.6-26 Watch Mode** (yüksek kullanıcı etkisi, continuous akış kırılıyor)
2. **v0.6-27 Rule Template Generator** (derlenebilir template güveni)
3. **v0.6-28 CLI Error Improvements** (tek tip hata deneyimi)
4. **v0.6-24 Progress Indicators** (pipeline görünürlüğü)
5. **v0.6-25 Colored Output** (terminal uyumluluk)
6. **v0.6-23 Interactive Mode** (configure akışını tamamlama)

---

## Risk notları

- **Watch mode düzeltmeleri** ana analiz akışını etkileyeceği için regresyon riski taşır (özellikle exit-code davranışı).
- **Template generator düzeltmesi** mevcut kullanıcıların beklediği çıktı formatını değiştirebilir; migration notu gerekebilir.
- **Hata yönetiminin merkezileştirilmesi** kısa vadede çok dosyaya dokunur; branch başına kapsam dar tutulmalı.
- **Mimari drift riski:** `internal/*` altındaki yeni mimari ile kök `main` paketindeki legacy akış paralel yaşıyor; yeni özellikler eklenirken tek yönlü bağımlılık kuralları netleştirilmezse drift hızlanır.

---

## Uygulama komut planı (23-28)

> Not: Her madde **ayrı branch + ayrı commit + ayrı push** olarak uygulanır.

### Ortak hazırlık (bir kez)

```bash
git status
git switch main
git pull --ff-only origin main
```

### v0.6-26 — Watch Mode

```bash
git switch main
git pull --ff-only origin main
git switch -c feature/v0.6-26-watch-mode-non-terminating-loop

# madde 26 değişikliklerini uygula

go test ./...
git add -A -- main.go internal/ cmd/ # yalnızca madde 26 ile ilgili yolları stage et
git diff --cached --name-only
git commit -m "fix(watch): keep watch loop alive on violations and improve dynamic directory watching"
git push -u origin feature/v0.6-26-watch-mode-non-terminating-loop
```

### v0.6-27 — Rule Template Generator

```bash
git switch main
git pull --ff-only origin main
git switch -c feature/v0.6-27-rule-template-generator-compile-safe

# madde 27 değişikliklerini uygula

go test ./...
git add -A -- generator.go generator_test.go rules/
git diff --cached --name-only
git commit -m "fix(generator): generate compile-safe rule templates aligned with current rule interfaces"
git push -u origin feature/v0.6-27-rule-template-generator-compile-safe
```

### v0.6-28 — CLI Error Improvements

```bash
git switch main
git pull --ff-only origin main
git switch -c feature/v0.6-28-cli-errors-unified-handling

# madde 28 değişikliklerini uygula

go test ./...
git add -A -- main.go cmd/ internal/
git diff --cached --name-only
git commit -m "refactor(cli-errors): unify structured error handling across all commands with suggestions"
git push -u origin feature/v0.6-28-cli-errors-unified-handling
```

### v0.6-24 — Progress Indicators

```bash
git switch main
git pull --ff-only origin main
git switch -c feature/v0.6-24-progress-indicators-stage-accuracy

# madde 24 değişikliklerini uygula

go test ./...
git add -A -- progress.go main.go internal/
git diff --cached --name-only
git commit -m "feat(progress): add explicit metrics stage and improve progress accuracy per pipeline step"
git push -u origin feature/v0.6-24-progress-indicators-stage-accuracy
```

### v0.6-25 — Colored CLI Output

```bash
git switch main
git pull --ff-only origin main
git switch -c feature/v0.6-25-colored-output-terminal-compat

# madde 25 değişikliklerini uygula

go test ./...
git add -A -- color.go main.go internal/
git diff --cached --name-only
git commit -m "fix(color): detect terminal capabilities and disable ANSI when unsupported"
git push -u origin feature/v0.6-25-colored-output-terminal-compat
```

### v0.6-23 — Interactive CLI Mode

```bash
git switch main
git pull --ff-only origin main
git switch -c feature/v0.6-23-interactive-mode-complete-workflow

# madde 23 değişikliklerini uygula

go test ./...
git add -A -- interactive.go internal/ cmd/
git diff --cached --name-only
git commit -m "feat(interactive): complete guided workflow for rule configuration and validation"
git push -u origin feature/v0.6-23-interactive-mode-complete-workflow
```

### PR açma şablonu (her branch için)

```bash
gh pr create --base main --head <branch-name> --title "<pr-title>" --body "$(cat <<'EOF'
## Summary
- specs/todo.md içindeki ilgili v0.6 maddesi uygulandı.
- Ayrı branch/commit/push kuralı korundu.
- go test ./... çalıştırıldı.
EOF
)"
```

---

## CI Stabilizasyon Planı — Function Size ve God Object İhlalleri (Yeni)

Bu bölüm, mevcut CI kırılımlarını **mimari sınırlar korunarak** en küçük ve güvenli refactor adımlarıyla kapatmak için hazırlanmıştır.

### Mimari bağlam (mevcut durumdan çıkarım)

- Kod tabanında hibrit bir yapı var: kök `main` paketi içinde CLI orkestrasyonu + UI/çıktı + çalışma akışları birlikte duruyor.
- `internal/*` altında model/language/engine ayrımı oluşmuş; ancak CLI tarafı (`main.go`, `interactive.go`, `generator.go`) hâlâ yoğun ve sorumlulukları iç içe.
- Bu nedenle mevcut ihlaller, yalnızca “satır sayısı” değil, aynı zamanda **sorumluluk ayrımı (SRP)** drift sinyali veriyor.

### Uygulama sırası (önerilen)

1. `main.go` → `executeCommand` parçalama (komut dispatch netleşir, diğer refactorlar için iskelet sağlar)
2. `generator.go` → `generateTemplate` parçalama (template üretim akışı sadeleşir)
3. `interactive.go` → `InteractiveMode` rol ayrıştırma (god object ihlali kapanır)

### Branch/commit önerisi (CI stabilizasyonu için güvenli ve ölçeklenebilir akış)

- **Tercih edilen izolasyon:** Her ihlal için ayrı branch + ayrı PR.
  - `refactor/ci-main-execute-command-split`
  - `refactor/ci-generator-template-split`
  - `refactor/ci-interactive-god-object-split`
- **Alternatif (tek issue zorunluluğu varsa):** Tek branch kullanılabilir; ancak küçük ve bağımsız commitler korunmalı, PR kapsamı yalnızca CI ihlallerini kapatmalı.
- Commit stratejisi (önerilen):
  1. `refactor(cli): split executeCommand into command-specific handlers`
  2. `refactor(generator): split generateTemplate into builder helpers`
  3. `refactor(interactive): extract config and analysis flows from InteractiveMode`
  4. `docs(todo): refine CI stabilization safety/DoD checklist`
- Push: `git push -u origin <secilen-branch>`
- Not: force push yok, destructive git komutu yok, `git add .` yok.

### 1) İhlal: `generator.go` / `generateTemplate` (80 satır limiti)

**Kök neden**
- Tek fonksiyon içinde birden fazla iş birikmiş durumda: template metni, template parse/execute, data hazırlama, hata fallback stratejisi.
- “Template oluşturma” ile “template render etme” aynı seviyede işlendiği için fonksiyon büyümüş.

**Refactor stratejisi (SOLID + mimari sınır uyumu)**
- SRP: `generateTemplate` fonksiyonunu üç yardımcıya ayır:
  - `buildRuleTemplateData(ruleName string) ruleTemplateData`
  - `renderRuleTemplate(data ruleTemplateData) (string, error)`
  - `ruleTemplateText() string` (sabit template kaynağını döndürür)
- `generateTemplate` sadece orkestrasyon yapmalı: typeName/data → render → hata varsa `generateSimpleTemplate` fallback.
- Davranış değişikliği yapma; sadece yapısal parçalama.

**Etkilenen dosyalar**
- `generator.go`
- (gerekirse) `generator_test.go` (davranış regresyonunu kilitlemek için)

**Kabul kriterleri**
- [ ] `generateTemplate` 80 satır altında
- [ ] Mevcut çıktı formatı korunuyor (başarılı durumda)
- [ ] Template parse/execute hatasında fallback hâlâ `generateSimpleTemplate`
- [ ] `go test ./...` başarılı

**Risk / rollback notu**
- Risk: Üretilen template içeriğinde fark oluşması (boşluk/yorum satırı dahil)
- Rollback: Tek commit geri alınarak eski tek-fonksiyon akışına dönülebilir; veri modeli değişmediği için düşük risk.

### 2) İhlal: `main.go` / `executeCommand` (80 satır limiti)

**Kök neden**
- `executeCommand`, parser kurulumunu, flag okumayı, format normalizasyonunu ve iş mantığı çağrılarını tek switch içinde birleştiriyor.
- Komut başına bağımsız sorumluluklar ortak bir megafonksiyonda toplanmış.

**Refactor stratejisi (SOLID + mimari sınır uyumu)**
- OCP + SRP: komut bazlı handler fonksiyonlarına böl:
  - `handleAnalyzeCommand(args []string) error`
  - `handleExtractCommand(args []string) error`
  - `handleReportCommand(args []string) error`
  - `handleHistoryCommand(args []string) error`
  - `handleGenerateCommand(args []string) error`
  - `handleVersionCommand() error` (veya direkt `nil`)
- `executeCommand` yalnızca dispatch + unknown command error üretimi yapmalı.
- Flag set-up ve output format normalizasyonunu handler içine taşı; işlevsel davranışı koru.

**Etkilenen dosyalar**
- `main.go`

**Kabul kriterleri**
- [ ] `executeCommand` 80 satır altında
- [ ] CLI komut davranışlarında geriye dönük uyumluluk (analyze/extract/report/history/generate/version/help)
- [ ] Hata mesajı/suggestion akışı (`Unknown command`) değişmeden çalışıyor
- [ ] `go test ./...` başarılı

**Risk / rollback notu**
- Risk: Flag parse davranışında sessiz kırılma (özellikle `--json` / `-format` etkileşimi)
- Rollback: Handler commit’i izole olduğundan tek commit revert ile hızlı geri dönüş mümkün.

### 3) İhlal: `interactive.go` / `InteractiveMode` (1 field + 14 method, god object)

**Kök neden**
- `InteractiveMode` UI menüsü, kullanıcı girişi, analiz tetikleme, history çağrısı, config yükleme/değiştirme/kaydetme akışlarını birlikte taşıyor.
- Bir nesne hem “view”, hem “input adapter”, hem “use-case orchestrator” rolünü üstlenmiş.

**Refactor stratejisi (SOLID + mimari sınır uyumu)**
- Mevcut paket sınırını bozmadan rol ayrıştır:
  - `InteractiveIO` (girdi/çıktı yardımcıları: readChoice/readString/confirm)
  - `InteractiveConfigController` (configureRules, toggle/set/save akışları)
  - `InteractiveSession` (Run + ana menü dispatch)
- Alternatif minimal yol (dosya eklemeden): `interactive.go` içinde private yardımcı struct’lar ile method dağıtımı.
- Hedef: god object kuralına takılan method sayısını tek struct üzerinde azaltmak; davranış aynı kalmalı.

**Etkilenen dosyalar**
- `interactive.go`
- (opsiyonel, minimal tercih edilirse) `interactive_config.go`, `interactive_io.go` (aynı package `main`)

**Kabul kriterleri**
- [ ] `InteractiveMode` (veya ana etkileşim struct’ı) god object eşiğini aşmıyor
- [ ] Interactive menü akışları aynı çalışıyor (analyze/history/configure)
- [ ] Config değişiklikleri kaydedilebiliyor (`saveConfig` akışı korunuyor)
- [ ] `go test ./...` başarılı

**Risk / rollback notu**
- Risk: Menü akışı sırasında kullanıcı girdisi yönlendirmesinde regresyon
- Rollback: Ayrıştırma adımları küçük commitlere bölünürse sorunlu adım geri alınarak etki minimize edilir.

### DoD (CI odaklı)

- [ ] Function size ihlalleri kapanmış (`generateTemplate`, `executeCommand`)
- [ ] God object ihlali kapanmış (`InteractiveMode`)
- [ ] Yeni/var olan testler geçiyor: `go test ./...`
- [ ] Stage edilen dosyalar hedefli ve doğrulanmış (`git diff --cached --name-only`)
- [ ] Secret içeren dosya/anahtar/token commit’e dahil edilmemiş
- [ ] Dokümantasyon güncel: bu plan bölümü korunmuş ve uygulanan adımlar işaretlenmiş

---

## v0.7 Multi-language & Architecture Hardening Backlog

Bu backlog, mevcut kod tabanındaki **mimari drift** noktalarını kapatmak için hazırlanmıştır. Amaç yalnızca yeni özellik eklemek değil; analyze pipeline, language adapter sözleşmesi ve rule execution akışını **gerçekten** ayrıştırıp sürdürülebilir hale getirmektir.

> Zorunlu çalışma kuralı: **Tek issue = tek branch = tek PR (base: `dev`)**.

### v0.7 Güvenli Git + Kalite Kapısı (Zorunlu)

- **Branch izolasyonu zorunlu:** Her issue yalnızca kendi branch’inde geliştirilir; issue dışı dosya değişikliği aynı PR’a alınmaz.
- **PR hedefi sabit:** v0.7 backlog kapsamındaki tüm PR’lar `dev` tabanına açılır (`main` tabanına PR açılmaz).
- **Güvenli git akışı zorunlu:**
  - Force push yasak (`git push --force`, `--force-with-lease` yok).
  - Destructive komutlar yasak (`git reset --hard`, `git clean -fd`, history rewrite içeren riskli akışlar yok).
  - Hedefli stage zorunlu (`git add -A -- <ilgili-dosyalar>`), `git add .` kullanılmaz.
- **Minimum kalite kapısı (merge öncesi):**
  - `go test ./...` başarılı.
  - Mümkünse yarış durumları için `go test -race ./...` başarılı (desteklenmeyen ortamlar PR notunda gerekçelendirilir).
  - İlgili issue için eklenen/etkilenen davranışlar test ile kilitlenir (unit veya integration).
  - CI pipeline’ı yeşil değilse merge yapılmaz.
- **Güvenlik kapısı (OWASP odaklı):**
  - Kullanıcı girdisi alan tüm yeni CLI yollarında path doğrulama/canonicalization yapılır.
  - Repo dışına taşan path/symlink traversal denemeleri kontrollü hata ile reddedilir.
  - Secret/credential dosyaları commit edilmez.
- **PR kapsam kuralı:** Her PR açıklamasında “scope dışı değişiklik yok” beyanı ve test çıktısı özeti bulunur.

### Issue RD-701 — Analyze CLI Path Semantiği Netleştirme (`analyze .` vs `-path`)

- **Issue ID / Başlık:** RD-701 — Analyze CLI path argümanlarının deterministik davranışı
- **Problem ve kök neden:**
  - `analyze` komutunda sadece `-path` flag’i parse ediliyor; `repodoctor analyze .` örneği help içinde var fakat pozisyonel argüman fiilen okunmuyor.
  - Bu uyumsuzluk kullanıcı deneyiminde belirsizlik ve test edilmesi zor davranış üretiyor.
- **Çözüm yaklaşımı (mimari kararlar):**
  - Komut-parsing için tek bir kural belirlenir: `analyze [path]` ve `-path` birlikte desteklenir; çakışma durumunda açık öncelik kuralı uygulanır (`-path` > pozisyonel).
  - Bu kural testlerle kilitlenir (CLI contract testi).
- **Etkilenen dosyalar/modüller:**
  - `main.go`
  - (yeni/ek) CLI davranış test dosyası
- **Kabul kriterleri (testlenebilir checklist):**
  - [ ] `repodoctor analyze .` geçerli path ile çalışır.
  - [ ] `repodoctor analyze -path ./x` çalışır.
  - [ ] `repodoctor analyze ./a -path ./b` durumunda öncelik kuralı deterministik uygulanır ve dokümante edilir.
  - [ ] Geçersiz path ve dosya path’i için hata mesajları tutarlı kalır.
  - [ ] Path canonicalization uygulanır; repo dışına kaçış/traversal denemeleri kontrollü şekilde reddedilir.
- **Risk / rollback notu:**
  - Risk: mevcut scriptlerin argüman alışkanlığı bozulabilir.
  - Rollback: parse değişikliği tek commit’te izole tutulur; revert ile eski davranışa dönülebilir.
- **Önerilen branch adı (tek issue = tek branch):** `feature/v0.7-rd-701-analyze-path-semantics`
- **Önerilen commit mesajı:** `fix(cli): make analyze positional path and -path behavior deterministic`
- **PR hedefi:** `dev`

---

### Issue RD-702 — Analyze Pipeline’ı Language Detector + Adapter Tabanına Taşıma

- **Issue ID / Başlık:** RD-702 — Go-only analyze pipeline’ın adapter-orkestrasyon pipeline’ına geçişi
- **Problem ve kök neden:**
  - `runAnalyze` hâlen kök paketteki Go odaklı import extraction (`extractImports`) ve graph akışına bağlı.
  - `internal/languages` içindeki detector/adapter altyapısı runtime analyze akışında kullanılmıyor.
- **Çözüm yaklaşımı (mimari kararlar):**
  - `AnalyzeOrchestrator` (veya eşdeğer) tanımlanır: `detect -> select adapter -> detect files -> collect metrics -> build graph -> rule execution`.
  - CLI (`main.go`) yalnızca giriş/çıkış ve orchestration çağrısı yapar; domain analiz adımları orchestrator’a taşınır.
  - İlk aşamada tek-dominant dil stratejisi korunur (multi-language merge sonraki issue).
- **Etkilenen dosyalar/modüller:**
  - `main.go`
  - `internal/languages/*`
  - (yeni/ek) `internal/analysis` veya `internal/pipeline` modülü
- **Kabul kriterleri (testlenebilir checklist):**
  - [ ] Analyze akışı adapter seçmeden çalıştırılamaz (detector zorunlu adım).
  - [ ] Seçilen adapter adı verbose çıktıda görünür.
  - [ ] Mevcut Go reposunda skor/ihlal çıktısı geriye uyumlu kalır (kritik farklar testlerde açıklanır).
  - [ ] Pipeline adımları unit test veya integration test ile doğrulanır.
  - [ ] Pipeline adımları dosya sayısına göre gereksiz tekrar tarama yapmaz (kaçınılabilir O(N²) döngüler kaldırılmıştır).
- **Risk / rollback notu:**
  - Risk: skor farkı ve rapor regresyonu.
  - Rollback: orchestrator entegrasyonu feature-flag veya tek commit revert ile geri alınabilir.
- **Önerilen branch adı (tek issue = tek branch):** `feature/v0.7-rd-702-adapter-based-analyze-pipeline`
- **Önerilen commit mesajı:** `refactor(analyze): route runtime pipeline through language detector and adapters`
- **PR hedefi:** `dev`

---

### Issue RD-703 — Python Adapter’ın Runtime’da Gerçekten Kullanılması

- **Issue ID / Başlık:** RD-703 — Python adapter runtime entegrasyonu
- **Problem ve kök neden:**
  - Python adapter implementasyonu ve testleri var; ancak analyze komutu bu adapter’ı seçip çalıştırmıyor.
  - Sonuç: “Python support” dokümantasyonda var, runtime davranışta yok.
- **Çözüm yaklaşımı (mimari kararlar):**
  - RD-702 sonrası detector, Python-dominant repoda `PythonAdapter` seçmeli.
  - Çıkan metrics/graph mevcut rule context’e dönüştürülmeli.
  - Python için geçici kapsama notu net olmalı (ör. sadece import + temel function/class sayımı).
- **Etkilenen dosyalar/modüller:**
  - `internal/languages/python_adapter.go`
  - `internal/languages/language_detector.go`
  - `main.go` veya yeni analyze orchestrator modülü
- **Kabul kriterleri (testlenebilir checklist):**
  - [ ] Python örnek repository’de analyze komutu Go parser’a düşmeden tamamlanır.
  - [ ] Rapor çıktısında Python adapter seçimi gözlemlenir (verbose).
  - [ ] Python import tabanlı dependency graph üretilir.
  - [ ] En az 1 entegrasyon testi Python path’i için CI’da geçer.
  - [ ] Büyük dosyalarda parse süresi/memory için koruma uygulanır (örn. boyut limiti veya güvenli fallback) ve dokümante edilir.
- **Risk / rollback notu:**
  - Risk: Python parsing’in satır-bazlı yaklaşımı false-positive üretebilir.
  - Rollback: Python adapter seçimi geçici olarak config üzerinden kapatılabilir veya commit revert edilir.
- **Önerilen branch adı (tek issue = tek branch):** `feature/v0.7-rd-703-python-runtime-integration`
- **Önerilen commit mesajı:** `feat(python): enable Python adapter in analyze runtime pipeline`
- **PR hedefi:** `dev`

---

### Issue RD-704 — Language Adapter Sözleşmesini Open/Closed Uyumlu Genişletme

- **Issue ID / Başlık:** RD-704 — Adapter contract hardening (extensible architecture)
- **Problem ve kök neden:**
  - Mevcut `LanguageAdapter` arayüzü temel ama genişletme noktaları zayıf; yeni dil eklemede ortak davranışlar (normalization, capability bildirimi, stdlib ayrımı) adapter içine dağılabilir.
  - Bu durum yeni dil eklendikçe çekirdek kodu değiştirme baskısı oluşturur (OCP ihlali riski).
- **Çözüm yaklaşımı (mimari kararlar):**
  - Sözleşmeye minimal ama stabil genişleme noktaları eklenir (örn. `Capabilities()` veya `NormalizeImport()` benzeri, ihtiyaç kadar).
  - Core pipeline adapter capability’sine göre çalışır; language-specific if/switch büyümez.
  - Geri uyumluluk için adapter kayıt sırasında default capability davranışı sağlanır.
- **Etkilenen dosyalar/modüller:**
  - `internal/languages/language_adapter.go`
  - `internal/languages/go_adapter.go`
  - `internal/languages/python_adapter.go`
  - analyze orchestrator modülü
- **Kabul kriterleri (testlenebilir checklist):**
  - [ ] Yeni bir adapter eklemek için core pipeline’da değişiklik zorunlu değildir.
  - [ ] Arayüz değişikliği sonrası mevcut Go/Python adapterları derlenir ve testleri geçer.
  - [ ] Capability/extension noktaları dokümante edilmiştir.
- **Risk / rollback notu:**
  - Risk: interface değişikliği çok sayıda derleme hatasına yol açabilir.
  - Rollback: interface değişimi ayrı commit olduğundan hızla geri alınabilir.
- **Önerilen branch adı (tek issue = tek branch):** `feature/v0.7-rd-704-language-adapter-ocp-contract`
- **Önerilen commit mesajı:** `refactor(languages): harden adapter contract for open-closed extensibility`
- **PR hedefi:** `dev`

---

### Issue RD-705 — Rule Engine Akışını Tek Kaynağa Toplama (Legacy vs Internal Ayrışması)

- **Issue ID / Başlık:** RD-705 — Rule execution birleştirme ve duplicated orchestration temizliği
- **Problem ve kök neden:**
  - Kök pakette eski rule/scoring akışı çalışırken, `internal/rules` + `internal/engine` tarafında ayrı bir Rule Engine bulunuyor.
  - Bu çift-akış mimari drift ve farklı doğruluk davranışı riski oluşturuyor.
- **Çözüm yaklaşımı (mimari kararlar):**
  - Runtime analyze için tek resmi yol belirlenir: `internal/rules` registry + `internal/engine` executor.
  - Kök paketteki legacy scorer/rule kodu kademeli olarak adapter katmanına dönüştürülür veya deprecate edilir.
  - Geçiş boyunca uyumluluk testi eklenir (eski/yeni sonuç kıyas toleransı net).
- **Etkilenen dosyalar/modüller:**
  - `scoring.go`
  - `main.go`
  - `internal/engine/executor.go`
  - `internal/rules/*`
- **Kabul kriterleri (testlenebilir checklist):**
  - [ ] Analyze runtime, tek bir rule execution yolunu kullanır.
  - [ ] Registry’de kayıtlı rule sayısı ve çalıştırılan rule sayısı raporda izlenebilir.
  - [ ] Sonuçlar deterministic kalır (aynı input -> aynı violation seti).
  - [ ] Rule yürütme sırası deterministik olarak sabitlenir (map iteration kaynaklı rastlantısallık engellenir).
- **Risk / rollback notu:**
  - Risk: Rule mesajları/score hesaplarında kırılma.
  - Rollback: geçiş adımları küçük tutulur; her adım bağımsız revert edilebilir.
- **Önerilen branch adı (tek issue = tek branch):** `feature/v0.7-rd-705-unify-rule-engine-runtime`
- **Önerilen commit mesajı:** `refactor(engine): unify analyze runtime on internal rule registry and executor`
- **PR hedefi:** `dev`

---

### Issue RD-706 — God Object Bölme Planı (CLI Orchestrator Katmanı)

- **Issue ID / Başlık:** RD-706 — `main.go` içindeki mega-akışların gerçek sorumluluklara ayrılması
- **Problem ve kök neden:**
  - CLI parsing, pipeline orchestrasyonu, rapor yazdırma, trend yazımı ve exit kararları tek dosyada/katmanda toplanmış.
  - Bu yapı değişiklikleri riskli hale getiriyor ve çoklu dil geçişinde genişlemeyi zorlaştırıyor.
- **Çözüm yaklaşımı (mimari kararlar):**
  - `AnalyzeController` (CLI), `AnalysisService` (pipeline), `ReportService` (output), `HistoryService` (state) gibi net roller tanımlanır.
  - `runAnalyze` sadece orchestration çağrısı + exit code politikası seviyesine çekilir.
  - Bağımlılık yönü: CLI -> service; service -> adapters/rules; ters bağımlılık yasak.
- **Etkilenen dosyalar/modüller:**
  - `main.go`
  - (yeni/ek) `internal/analysis/*` veya `internal/app/*`
  - `reporter.go`, `trend_analyzer.go` (gerekirse interface extraction)
- **Kabul kriterleri (testlenebilir checklist):**
  - [ ] `main.go` içinde analyze iş mantığı minimal seviyeye iner (controller seviyesinde kalır).
  - [ ] Yeni service katmanı test edilebilir (mock/stub ile).
  - [ ] Exit code politikası tek yerde tanımlanır ve watch mode ile çakışmaz.
- **Risk / rollback notu:**
  - Risk: refactor sırasında davranış regresyonu.
  - Rollback: dosya-bölme adımları küçük commitlere ayrılır; hatalı adım tekil revert edilir.
- **Önerilen branch adı (tek issue = tek branch):** `refactor/v0.7-rd-706-split-cli-analysis-god-object`
- **Önerilen commit mesajı:** `refactor(cli): split analyze orchestration into controller and analysis services`
- **PR hedefi:** `dev`

---

### Issue RD-707 — God Object Bölme Planı (Interactive ve Generator Ayrıştırma)

- **Issue ID / Başlık:** RD-707 — Interactive/Generator sorumluluk parçalama
- **Problem ve kök neden:**
  - `interactive.go` içinde oturum akışı + config yönetimi + I/O bir arada; `generator.go` içinde template üretim akışı yoğun.
  - Çoklu dil ve yeni komutlar eklendikçe bu dosyalar büyüyerek bakım maliyetini artırıyor.
- **Çözüm yaklaşımı (mimari kararlar):**
  - `interactive` tarafı: session/menu, io adapter, config use-case olarak ayrıştırılır.
  - `generator` tarafı: template data builder, renderer, fallback stratejisi bileşenlerine bölünür.
  - Davranış değişikliği yapılmadan yapı sadeleştirilir (refactor-only).
- **Etkilenen dosyalar/modüller:**
  - `interactive.go`
  - `generator.go`
  - ilgili test dosyaları
- **Kabul kriterleri (testlenebilir checklist):**
  - [ ] Interactive ana struct method/rol yoğunluğu eşiğin altına iner.
  - [ ] Generator akışı fonksiyon boyutu ve SRP açısından sadeleşir.
  - [ ] CLI davranışları ve çıktılar geriye uyumlu kalır.
- **Risk / rollback notu:**
  - Risk: interactive girdilerinde UX regresyonu.
  - Rollback: her dosya ayrıştırması bağımsız commit olarak geri alınabilir.
- **Önerilen branch adı (tek issue = tek branch):** `refactor/v0.7-rd-707-split-interactive-generator-god-objects`
- **Önerilen commit mesajı:** `refactor(interactive,generator): separate io, session, and template responsibilities`
- **PR hedefi:** `dev`

---

## Issue Sırası ve Bağımlılık Grafiği

### Önerilen uygulama sırası

1. **RD-701 (P0)** — CLI path contract sabitlenmeden pipeline geçişi riskli.
2. **RD-702 (P0)** — Adapter tabanlı analyze omurgası.
3. **RD-703 (P0)** — Python adapter runtime aktifleme.
4. **RD-704 (P1)** — Adapter sözleşmesi OCP hardening.
5. **RD-705 (P1)** — Rule engine tekleştirme.
6. **RD-706 (P1)** — main.go orchestration god object parçalama.
7. **RD-707 (P2)** — interactive/generator refactor ve sürdürülebilirlik.

### Bağımlılık grafiği (özet)

- `RD-701 -> RD-702`
- `RD-702 -> RD-703`
- `RD-702 -> RD-704`
- `RD-702 + RD-704 -> RD-705`
- `RD-705 -> RD-706`
- `RD-706 -> RD-707`

### Görsel (ASCII DAG)

```text
RD-701
   |
RD-702
  |  \
  |   \-> RD-704 --\
  v               \
RD-703             > RD-705 -> RD-706 -> RD-707
```

---

## Her Issue için Minimal Git/Pull Request Akışı

Her issue için aşağıdaki kalıp uygulanır:

```bash
git switch dev
git pull --ff-only origin dev
git switch -c <issue-branch>

# sadece ilgili issue değişiklikleri

go test ./...
# (destekleniyorsa) go test -race ./...
git add -A -- <yalnızca-ilgili-dosyalar>
git diff --cached --name-only
git commit -m "<issue-commit-message>"
git push -u origin <issue-branch>

gh pr create --base dev --head <issue-branch> --title "<issue-title>" --body "<issue-summary + test/CI özeti + scope dışı değişiklik yok beyanı>"
```

Notlar:
- Force push yok.
- Destructive git komutu yok.
- Tek issue kapsamı dışına taşan dosya stage edilmeyecek.
- PR merge koşulu: CI yeşil + en az 1 gözden geçirme + açık kritik güvenlik bulgusu olmaması.

---

## v0.8 God Object Elimination & Report Accuracy Backlog

Bu backlog, v0.7 sonrası kalan **god object ihlallerini** ve **rapor doğruluk hatalarını** kapatan sprint'tir.

> Skor etkisi: **67/100 → 97/100** (+30 puan)

### Tamamlanan Issue'lar

#### RD-708 — Report Mapping Placeholder Düzeltmesi ✅
- **PR:** #84 (merged → dev)
- **Problem:** `buildReportFromRuleViolations()` size ve god-object violation'ları için `Lines: 1, Threshold: 1, FieldCount: 1, MethodCount: 1` hardcode ediyordu.
- **Çözüm:** Regex tabanlı `parseSizeViolation()` ve `mergeGodObjectViolation()` helper'ları eklendi; violation mesajlarından gerçek değerler parse ediliyor.
- **Etki:** Rapor artık `"567 lines (threshold: 500)"` ve `"0 fields, 12 methods"` gibi doğru değerler gösteriyor.

#### RD-709 — God Object Rule Name Collision Düzeltmesi ✅
- **PR:** #83 (merged → dev)
- **Problem:** `GodObjectRule.Evaluate()` bare struct name kullanarak map key oluşturuyordu; farklı package'lardaki aynı isimli struct'ların method'ları birleşiyordu (ör. `main.DependencyGraph` + `model.DependencyGraph` = 21 method false positive).
- **Çözüm:** `structKey(filePath, structName)` helper'ı ile `Dir(filePath)+"#"+structName` formatında package-qualified key kullanıldı.
- **Etki:** 3 false positive god object ihlali elendi; skor 67 → 82.

#### RD-710 — DependencyGraph God Object Parçalama ✅
- **PR:** #85 (merged → dev)
- **Problem:** `model.DependencyGraph` 14 method ile god object eşiğini aşıyordu.
- **Çözüm:** `DetectCycles/GetCycles/HasCycles` → `GraphCycleDetector` struct'ına; `GetRoots/GetLeaves` → `FindRoots/FindLeaves` standalone fonksiyonlarına çıkarıldı.
- **Etki:** 14 → 10 method (eşik altı).

#### RD-711 — GoAdapter God Object Parçalama ✅
- **PR:** #86 (merged → dev)
- **Problem:** `GoAdapter` 12 method ile god object eşiğini aşıyordu.
- **Çözüm:** `extractFunctionMetrics`, `extractStructMetrics`, `parseFileAndAddToGraph` → package-level fonksiyonlara (`goExtractFunctionMetrics`, `goExtractStructMetrics`, `goParseFileAndAddToGraph`) dönüştürüldü.
- **Etki:** 12 → 9 method (eşik altı).

#### RD-712 — PythonAdapter God Object Parçalama ✅
- **PR:** #87 (merged → dev)
- **Problem:** `PythonAdapter` 13 method ile god object eşiğini aşıyordu.
- **Çözüm:** `extractFunctionMetrics`, `extractClassMetrics`, `GetPythonVersion` → package-level fonksiyonlara (`pyExtractFunctionMetrics`, `pyExtractClassMetrics`, `DetectPythonVersion`) dönüştürüldü.
- **Etki:** 13 → 10 method (eşikte).

#### RD-713 — main.go Size Violation Kapatma ✅
- **PR:** #89 (merged → dev), #90 (dev → main)
- **Problem:** `main.go` 567 satır ile 500 satır eşiğini aşıyordu (son kalan ihlal).
- **Çözüm:** 6 secondary CLI fonksiyonu (`scanDirectory`, `runReport`, `runHistory`, `runExtract`, `runGenerate`, `runWatch`) `cli_commands.go` dosyasına taşındı.
- **Etki:** `main.go` 567 → 485 satır (eşik altı); skor 97 → **100/100**. Sıfır ihlal.

### v0.8 Sprint Skor Özeti

| PR | Issue | Skor | İhlal Sayısı |
|----|-------|------|-------------|
| #83 | RD-709 | 82/100 | 4 |
| #84 | RD-708 | 82/100 | 4 (accuracy fix) |
| #85 | RD-710 | 87/100 | 3 |
| #86 | RD-711 | 92/100 | 2 |
| #87 | RD-712 | 97/100 | 1 |
| #89 | RD-713 | **100/100** | **0** |

### Kalan İhlaller

✅ **Sıfır ihlal — v0.8 hedefi tamamlandı.**

### Sonraki Adımlar (v0.9 önerileri)

1. JavaScript/TypeScript language adapter eklemesi.
2. JSON rapor formatının zenginleştirilmesi (violation detaylarıyla).
3. Version string'inin build-time injection ile otomatik güncellenmesi.
4. Legacy root-package rule kodunun kademeli deprecation planı.

---

## v0.6 Completion Sprint — Issue-Based Implementation Plan

Bu sprint, v0.6'da kısmen tamamlanmış 6 özelliği tamamlayarak v0.6 milestone'unu kapatır.

> **Kural:** Tek issue = tek branch = tek PR (base: `dev`). Tüm PR'lar merge sonrası `dev→main` PR ile release edilir.

### Kalite Kapısı (Her Issue İçin Zorunlu)

- `go build ./...` başarılı
- `go test ./...` başarılı (75+ test)
- `go vet ./...` temiz
- Self-analysis skoru 100/100 korunacak
- Hedefli `git add` (ilgili dosyalar), `git add .` yasak
- Force push yasak, destructive git komutları yasak

---

### RD-723 — Interactive Mode: God Object Threshold Configuration ✅

- **Branch:** `fix/v0.6-rd-723-interactive-god-object-config`
- **Problem:** Interactive `configureRules()` menüsünde sadece size rule ayarları var (max file lines, max function lines). God object rule eşikleri (max_fields, max_methods) yapılandırılamıyor.
- **Kök neden:** `showConfigMenu` ve `configureRules` sadece 4 seçenek sunuyor; god object threshold set/toggle metodları hiç eklenmemiş.
- **Çözüm:**
  1. `showConfigMenu`'ye 2 yeni seçenek ekle: "Set Max Fields", "Set Max Methods"
  2. `InteractiveConfigController`'a `setMaxFields(config)` ve `setMaxMethods(config)` metodları ekle
  3. Menü numaralarını güncelle (5→Save, 6→Back yerine 7→Save, 8→Back)
  4. `showConfigMenu`'deki current settings bölümüne Max Fields / Max Methods gösterimi ekle
- **Etkilenen dosyalar:** `interactive.go`
- **Commit mesajı:** `feat(interactive): add god object threshold configuration to interactive menu`
- **Kabul kriterleri:**
  - [x] Interactive menüde Max Fields ve Max Methods ayarlanabiliyor
  - [x] Değişiklikler save ile kaydediliyor
  - [x] Mevcut menü davranışları korunuyor
  - [x] `go test ./...` başarılı

---

### RD-724 — Progress Indicators: Real Stage Counts ✅

- **Branch:** `fix/v0.6-rd-724-progress-real-stage-counts`
- **Problem:** `getStageCount()` fonksiyonu "Collecting metrics" ve "Building dependency graph" stage'leri için hardcoded `10` döndürüyor. Bu, progress bar'ın gerçek ilerlemeyi yansıtmamasına neden oluyor.
- **Kök neden:** İlk implementasyonda sadece scanning stage'i gerçek dosya sayısı kullanıyor; diğer stage'ler yaklaşık değer kullanıyor.
- **Çözüm:**
  1. `getStageCount` fonksiyonunda "Collecting metrics" stage'i için `countFiles(repoPath)` kullan (her dosya için metric toplanıyor)
  2. "Building dependency graph" stage'i için de `countFiles(repoPath)` kullan (her dosya graph'a ekleniyor)
  3. "Running rules" stage'i 4 olarak kalsın (gerçekten 4 rule tipi var)
- **Etkilenen dosyalar:** `progress.go`
- **Commit mesajı:** `fix(progress): use real file counts for metrics and graph progress stages`
- **Kabul kriterleri:**
  - [x] Metrics ve graph stage'leri gerçek dosya sayısına göre ilerleme gösteriyor
  - [x] Scanning, metrics, graph, rules stage'leri ayrı ve doğru etiketli
  - [x] `go test ./...` başarılı

---

### RD-725 — Colored Output: Windows Virtual Terminal Processing ✅

- **Branch:** `fix/v0.6-rd-725-colored-output-windows-vtp`
- **Problem:** `isTerminal()` Windows'ta WT_SESSION/ANSICON/ConEmuANSI ortam değişkenlerine bakıyor ama standart Windows Terminal (cmd.exe) bu değişkenleri set etmiyor. Modern Windows 10+ VTP destekliyor ama bu kontrol yapılmıyor.
- **Kök neden:** Windows terminal desteği sadece env-var bazlı kontrol ediyor; `golang.org/x/sys/windows` ile VTP enable etme veya Go 1.21+ `os.Stdout.Stat()` fallback'i yeterli olabilir.
- **Çözüm:**
  1. Windows'ta `os.Stdout.Stat()` fallback'inin `ModeCharDevice` kontrolünü güçlendir — bu zaten mevcut ama env var branch'i `return true` yapıp atlıyor
  2. Windows branch'ini düzelt: env var'lar varsa `true` dön, yoksa `ModeCharDevice` kontrolüne düş (mevcut kodda zaten böyle — kontrol et)
  3. `isTerminal()` fonksiyonuna açıklayıcı yorum ekle
- **Etkilenen dosyalar:** `color.go`
- **Commit mesajı:** `fix(color): improve Windows terminal detection and add TERM_PROGRAM support`
- **Kabul kriterleri:**
  - [x] `NO_COLOR` env var set edildiğinde renkler kapalı
  - [x] `TERM=dumb` durumunda renkler kapalı
  - [x] Pipe/redirect durumunda renkler kapalı (ModeCharDevice)
  - [x] `go test ./...` başarılı

---

### RD-726 — Watch Mode: Graceful Shutdown with Signal Handling ✅

- **Branch:** `fix/v0.6-rd-726-watch-graceful-shutdown`
- **Problem:** `WatchAndAnalyze` fonksiyonu `select {}` ile sonsuza kadar bekliyor. Ctrl+C ile kapatıldığında watcher düzgün temizlenmiyor (Close() çağrılmıyor).
- **Kök neden:** Signal handling implementasyonu eksik; `select {}` bloğu yerine `os/signal` kullanılarak graceful shutdown yapılmalı.
- **Çözüm:**
  1. `WatchAndAnalyze`'da `os/signal.Notify` ile `SIGINT`, `SIGTERM` dinle
  2. `select {}` yerine signal channel'ını bekle
  3. Signal geldiğinde `watcher.Stop()` çağrısı ile temiz kapatma yap
  4. Kapatma mesajı yazdır
- **Etkilenen dosyalar:** `watcher.go`
- **Commit mesajı:** `fix(watch): add graceful shutdown with OS signal handling`
- **Kabul kriterleri:**
  - [x] Ctrl+C ile düzgün kapatılıyor, watcher.Close() çağrılıyor
  - [x] Kapatma sırasında kullanıcıya mesaj gösteriliyor
  - [x] Mevcut watch loop davranışı korunuyor
  - [x] `go test ./...` başarılı

---

### RD-727 — Rule Template Generator: Align with Internal Rule Interface ✅

- **Branch:** `fix/v0.6-rd-727-generator-rule-interface-alignment`
- **Problem:** Üretilen rule template'i kendi `Violation` struct'ını tanımlıyor ve `Evaluate(rootPath string) ([]Violation, error)` imzası kullanıyor. Bu, `internal/rules.Rule` interface'iyle uyumlu değil: `Evaluate(context AnalysisContext) []model.Violation`.
- **Kök neden:** Template, Rule interface standartlaştırılmadan önce yazılmış ve güncellenmemiş.
- **Çözüm:**
  1. `ruleTemplateText()` template'ini güncelle: `Evaluate(context AnalysisContext) []model.Violation` imzası kullan
  2. Template'den bağımsız `Violation` struct tanımını kaldır
  3. Import'lara `"RepoDoctor/internal/model"` ve `"RepoDoctor/internal/rules"` (veya sadece AnalysisContext için rules) ekle
  4. `simpleRuleTemplate` sabitini de aynı şekilde güncelle
  5. Test'i güncelle: derlenebilirlik kontrolü `internal/rules` interface'ine uyumluluğu doğrulamalı
- **Etkilenen dosyalar:** `generator.go`, `generator_test.go`
- **Commit mesajı:** `fix(generator): align rule template with internal Rule interface contract`
- **Kabul kriterleri:**
  - [x] Üretilen template `internal/rules.Rule` interface'ini implemente ediyor
  - [x] Template'de bağımsız Violation struct yok
  - [x] Üretilen dosya Go syntax olarak geçerli (parser.ParseFile)
  - [x] `go test ./...` başarılı

---

### RD-728 — CLI Error Improvements: Eliminate Duplicate Error Messages ✅

- **Branch:** `fix/v0.6-rd-728-cli-errors-deduplicate`
- **Problem:** `validatePath()` fonksiyonu her hata durumunda hem `err.Display()` çağırıp hem de `fmt.Fprintf(os.Stderr, ColorError(...))` ile aynı mesajı tekrar basıyor. Bu, kullanıcıya çift hata mesajı gösteriyor.
- **Kök neden:** Hata altyapısı (`CLIError.Display()`) eklendikten sonra eski `fmt.Fprintf` satırları kaldırılmamış.
- **Çözüm:**
  1. `validatePath()`'teki her hata bloğunda `CLIError.Display()` çağrısını tut, duplicate `fmt.Fprintf` satırlarını kaldır
  2. `runWatch()` fonksiyonunda raw `fmt.Fprintf + os.Exit(1)` yerine `WrapError` kullan
  3. `Display()` metodunun renk formatlamasını kullandığından emin ol
- **Etkilenen dosyalar:** `main.go`, `cli_commands.go`
- **Commit mesajı:** `fix(cli-errors): remove duplicate error messages in validatePath and unify runWatch error handling`
- **Kabul kriterleri:**
  - [x] Her hata durumunda tek bir mesaj basılıyor
  - [x] Tüm komut yolları `CLIError` formatından geçiyor
  - [x] `go test ./...` başarılı
  - [x] `main.go` 500 satır altında kalıyor

---

### Uygulama Sırası

Bağımsız issue'lar — herhangi bir sırada uygulanabilir:

```text
RD-723 (interactive)  ─┐
RD-724 (progress)     ─┤
RD-725 (color)        ─┤── hepsi bağımsız ──→ dev merge ──→ dev→main PR
RD-726 (watch)        ─┤
RD-727 (generator)    ─┤
RD-728 (cli-errors)   ─┘
```

### Git Workflow

```bash
# Her issue için:
git switch dev && git pull origin dev
git switch -c <issue-branch>
# ... değişiklikleri uygula ...
go test ./... && go vet ./... && go build ./...
go run . analyze -path .  # skor 100/100 kalmalı
git add -A -- <ilgili-dosyalar>
git commit -m "<commit-mesajı>"
git push -u origin <issue-branch>
gh pr create --base dev --head <issue-branch> --title "<pr-title>" --body "<özet>"
gh pr merge <pr-no> --merge --delete-branch

# Tümü bittikten sonra:
gh pr create --base main --head dev --title "release: v0.6 completion sprint"
gh pr merge <pr-no> --merge
```
