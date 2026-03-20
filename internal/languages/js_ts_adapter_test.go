package languages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJavaScriptAdapter_DetectFilesAndEvidence(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "package.json"), []byte(`{"name":"demo"}`), 0o644); err != nil {
		t.Fatalf("failed writing package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "app.js"), []byte("import x from 'react'\n"), 0o644); err != nil {
		t.Fatalf("failed writing app.js: %v", err)
	}

	adapter := NewJavaScriptAdapter()
	files, err := adapter.DetectFiles(repo)
	if err != nil {
		t.Fatalf("DetectFiles failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 js file, got %d", len(files))
	}

	provider, ok := adapter.(EvidenceProvider)
	if !ok {
		t.Fatal("adapter must implement EvidenceProvider")
	}
	signals, warnings, err := provider.CollectEvidence(repo, files)
	if err != nil {
		t.Fatalf("CollectEvidence failed: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(signals) == 0 {
		t.Fatal("expected js evidence signals")
	}
}

func TestTypeScriptAdapter_MetadataValidation_BoundedAndGraceful(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "main.ts"), []byte("export const x = 1\n"), 0o644); err != nil {
		t.Fatalf("failed writing ts file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "tsconfig.json"), []byte(`{"compilerOptions": {"strict": true}}`), 0o644); err != nil {
		t.Fatalf("failed writing tsconfig: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "package.json"), []byte(`{"name":"demo","dependencies":{"typescript":"^5"}}`), 0o644); err != nil {
		t.Fatalf("failed writing package.json: %v", err)
	}

	adapter := NewTypeScriptAdapter()
	files, err := adapter.DetectFiles(repo)
	if err != nil {
		t.Fatalf("DetectFiles failed: %v", err)
	}
	provider := adapter.(EvidenceProvider)
	signals, warnings, err := provider.CollectEvidence(repo, files)
	if err != nil {
		t.Fatalf("CollectEvidence failed: %v", err)
	}
	if len(signals) == 0 {
		t.Fatal("expected ts evidence")
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	// malformed metadata must not panic, should produce warning
	if err := os.WriteFile(filepath.Join(repo, "package.json"), []byte(`{"name":`), 0o644); err != nil {
		t.Fatalf("failed writing malformed package.json: %v", err)
	}
	provider = NewTypeScriptAdapter().(EvidenceProvider)
	_, warnings, err = provider.CollectEvidence(repo, files)
	if err != nil {
		t.Fatalf("CollectEvidence should not fail on malformed metadata: %v", err)
	}
	if len(warnings) == 0 {
		t.Fatal("expected warning for malformed metadata")
	}
}

func TestLanguageDetector_DeterministicWithAdapterRegistrationOrder(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "main.ts"), []byte("export const x = 1\n"), 0o644); err != nil {
		t.Fatalf("failed writing ts file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "helper.js"), []byte("module.exports = {}\n"), 0o644); err != nil {
		t.Fatalf("failed writing js file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "tsconfig.json"), []byte(`{"compilerOptions": {"strict": true}}`), 0o644); err != nil {
		t.Fatalf("failed writing tsconfig: %v", err)
	}

	strategy := &fixedIgnore{}

	first := NewRepositoryLanguageDetector(strategy)
	first.RegisterAdapter(NewJavaScriptAdapter())
	first.RegisterAdapter(NewTypeScriptAdapter())
	first.RegisterAdapter(NewGoAdapter())
	first.RegisterAdapter(NewPythonAdapter())

	second := NewRepositoryLanguageDetector(strategy)
	second.RegisterAdapter(NewPythonAdapter())
	second.RegisterAdapter(NewGoAdapter())
	second.RegisterAdapter(NewTypeScriptAdapter())
	second.RegisterAdapter(NewJavaScriptAdapter())

	a, err := first.DetectLanguage(repo)
	if err != nil {
		t.Fatalf("first detector failed: %v", err)
	}
	b, err := second.DetectLanguage(repo)
	if err != nil {
		t.Fatalf("second detector failed: %v", err)
	}

	if a.Name() != b.Name() {
		t.Fatalf("expected deterministic output, got %s and %s", a.Name(), b.Name())
	}
}

type fixedIgnore struct{}

func (f *fixedIgnore) ShouldIgnore(string, string) bool { return false }

func TestJSTSAdapter_CollectEvidence_SkipsOutsideScope(t *testing.T) {
	repo := t.TempDir()
	outside := t.TempDir()

	outsideFile := filepath.Join(outside, "outside.js")
	if err := os.WriteFile(outsideFile, []byte("import x from 'left-pad'\n"), 0o644); err != nil {
		t.Fatalf("failed writing outside fixture: %v", err)
	}

	adapter := NewJavaScriptAdapter().(EvidenceProvider)
	signals, warnings, err := adapter.CollectEvidence(repo, []string{outsideFile})
	if err != nil {
		t.Fatalf("CollectEvidence failed: %v", err)
	}
	if len(signals) != 0 {
		t.Fatalf("expected no signals outside root scope, got %d", len(signals))
	}
	if len(warnings) == 0 {
		t.Fatal("expected warning for outside-scope path")
	}
}

func TestJSTSAdapter_MetadataDepthGuard(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "main.ts"), []byte("export const x = 1\n"), 0o644); err != nil {
		t.Fatalf("failed writing ts file: %v", err)
	}

	deep := `{"a":{"b":{"c":{"d":{"e":{"f":{"g":{"h":{"i":1}}}}}}}}}`
	if err := os.WriteFile(filepath.Join(repo, "package.json"), []byte(deep), 0o644); err != nil {
		t.Fatalf("failed writing deep package.json: %v", err)
	}

	provider := NewTypeScriptAdapter().(EvidenceProvider)
	files, err := NewTypeScriptAdapter().DetectFiles(repo)
	if err != nil {
		t.Fatalf("DetectFiles failed: %v", err)
	}
	_, warnings, collectErr := provider.CollectEvidence(repo, files)
	if collectErr != nil {
		t.Fatalf("CollectEvidence should stay safe with deep metadata: %v", collectErr)
	}

	warnJoined := strings.Join(warnings, "|")
	if !strings.Contains(warnJoined, "metadata nested too deeply") {
		t.Fatalf("expected depth warning, got %v", warnings)
	}
}
