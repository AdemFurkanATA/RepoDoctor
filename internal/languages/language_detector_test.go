package languages_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"RepoDoctor/internal/domain"
	"RepoDoctor/internal/languages"
	"RepoDoctor/internal/model"
)

// MockIgnoreStrategy used for tests
type MockIgnoreStrategy struct {
	ignored map[string]bool
}

func (m *MockIgnoreStrategy) ShouldIgnore(_ string, dirName string) bool {
	return m.ignored[dirName]
}

// DummyAdapter for testing
type DummyAdapter struct {
	name string
	exts []string
}

func (d *DummyAdapter) Name() string                                  { return d.name }
func (d *DummyAdapter) FileExtensions() []string                      { return d.exts }
func (d *DummyAdapter) DetectFiles(repoPath string) ([]string, error) { return nil, nil }
func (d *DummyAdapter) CollectMetrics(files []string) (*model.RepositoryMetrics, error) {
	return nil, nil
}
func (d *DummyAdapter) BuildDependencyGraph(files []string) (*model.DependencyGraph, error) {
	return nil, nil
}
func (d *DummyAdapter) IsStdlibPackage(importPath string) bool { return false }
func (d *DummyAdapter) Capabilities() languages.AdapterCapabilities {
	return languages.AdapterCapabilities{}
}
func (d *DummyAdapter) NormalizeImport(importPath string) string { return importPath }

func TestLanguageDetector_DetectLanguage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "detector_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "utils.go"), []byte("package utils"), 0644)

	nodeModulesDir := filepath.Join(tempDir, "node_modules")
	_ = os.MkdirAll(nodeModulesDir, 0755)
	for i := 0; i < 5; i++ {
		_ = os.WriteFile(filepath.Join(nodeModulesDir, "index"+strconv.Itoa(i)+".py"), []byte("print('hello')"), 0644)
	}

	mockStrategy := &MockIgnoreStrategy{ignored: map[string]bool{"node_modules": true}}
	detector := languages.NewRepositoryLanguageDetector(mockStrategy)
	detector.RegisterAdapter(&DummyAdapter{name: "Go", exts: []string{".go"}})
	detector.RegisterAdapter(&DummyAdapter{name: "Python", exts: []string{".py"}})

	adapter, err := detector.DetectLanguage(tempDir)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if adapter.Name() != "Go" {
		t.Errorf("Expected dominant language 'Go', got %q", adapter.Name())
	}
}

func TestLanguageDetector_DetectLanguage_DeterministicTieBreak(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "detector_tie_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.WriteFile(filepath.Join(tempDir, "a.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to write go file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "a.py"), []byte("print('x')\n"), 0o644); err != nil {
		t.Fatalf("failed to write python file: %v", err)
	}

	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	detector := languages.NewRepositoryLanguageDetector(strategy)
	detector.RegisterAdapter(&DummyAdapter{name: "Go", exts: []string{".go"}})
	detector.RegisterAdapter(&DummyAdapter{name: "Python", exts: []string{".py"}})

	for i := 0; i < 20; i++ {
		adapter, detectErr := detector.DetectLanguage(tempDir)
		if detectErr != nil {
			t.Fatalf("detect language failed on iteration %d: %v", i, detectErr)
		}
		if adapter.Name() != "Python" {
			t.Fatalf("expected deterministic tie break to pick Python, got %s on iteration %d", adapter.Name(), i)
		}
	}
}

func TestLanguageDetector_DetectLanguage_WeightedProductOverTooling(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "detector_weighted_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "src")
	toolsDir := filepath.Join(tempDir, "tools")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("failed creating src dir: %v", err)
	}
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("failed creating tools dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(srcDir, "app.py"), []byte("def run():\n    return 1\n"), 0o644); err != nil {
		t.Fatalf("failed writing src python file: %v", err)
	}

	for i := 0; i < 30; i++ {
		name := filepath.Join(toolsDir, "tool_file_"+strconv.Itoa(i)+".go")
		if err := os.WriteFile(name, []byte("package tools\n"), 0o644); err != nil {
			t.Fatalf("failed writing tooling go file: %v", err)
		}
	}

	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	detector := languages.NewRepositoryLanguageDetector(strategy)
	detector.RegisterAdapter(&DummyAdapter{name: "Go", exts: []string{".go"}})
	detector.RegisterAdapter(&DummyAdapter{name: "Python", exts: []string{".py"}})

	adapter, detectErr := detector.DetectLanguage(tempDir)
	if detectErr != nil {
		t.Fatalf("detect language failed: %v", detectErr)
	}

	if adapter.Name() != "Python" {
		t.Fatalf("expected product Python to win over tooling Go, got %s", adapter.Name())
	}
}

func TestLanguageDetector_DetectLanguage_PythonRelativeImportsPreferred(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "detector_py_relative")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.WriteFile(filepath.Join(tempDir, "pyproject.toml"), []byte("[project]\nname='demo'\n"), 0o644); err != nil {
		t.Fatalf("failed writing pyproject marker: %v", err)
	}

	appDir := filepath.Join(tempDir, "src", "pkg", "app")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("failed creating package dirs: %v", err)
	}

	for _, rel := range []string{"src/pkg/__init__.py", "src/pkg/app/__init__.py"} {
		if err := os.WriteFile(filepath.Join(tempDir, rel), []byte(""), 0o644); err != nil {
			t.Fatalf("failed writing __init__.py: %v", err)
		}
	}

	pythonMain := "from .service import run\nfrom ..shared import util\n"
	if err := os.WriteFile(filepath.Join(appDir, "main.py"), []byte(pythonMain), 0o644); err != nil {
		t.Fatalf("failed writing python source: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "scripts.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed writing go script: %v", err)
	}

	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	detector := languages.NewRepositoryLanguageDetector(strategy)
	detector.RegisterAdapter(languages.NewGoAdapter())
	detector.RegisterAdapter(languages.NewPythonAdapter())

	for i := 0; i < 10; i++ {
		adapter, detectErr := detector.DetectLanguage(tempDir)
		if detectErr != nil {
			t.Fatalf("detect language failed: %v", detectErr)
		}
		if adapter.Name() != "Python" {
			t.Fatalf("expected Python, got %s on iteration %d", adapter.Name(), i)
		}
	}
}

func TestLanguageDetector_DetectLanguage_SkipsSymlinkOutsideRoot(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "detector_symlink")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outside, err := os.MkdirTemp("", "detector_outside")
	if err != nil {
		t.Fatalf("failed creating outside dir: %v", err)
	}
	defer os.RemoveAll(outside)

	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed writing go file: %v", err)
	}

	if err := os.WriteFile(filepath.Join(outside, "danger.py"), []byte("print('x')\n"), 0o644); err != nil {
		t.Fatalf("failed writing outside file: %v", err)
	}

	if err := os.Symlink(outside, filepath.Join(tempDir, "linked_outside")); err != nil {
		t.Skipf("symlink not supported in this environment: %v", err)
	}

	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	detector := languages.NewRepositoryLanguageDetector(strategy)
	detector.RegisterAdapter(languages.NewGoAdapter())
	detector.RegisterAdapter(languages.NewPythonAdapter())

	adapter, detectErr := detector.DetectLanguage(tempDir)
	if detectErr != nil {
		t.Fatalf("detect language failed: %v", detectErr)
	}

	if adapter.Name() != "Go" {
		t.Fatalf("expected Go after skipping outside symlink, got %s", adapter.Name())
	}
}

func TestLanguageDetector_DetectLanguage_DeterministicAcrossRuns(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "detector_deterministic")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.WriteFile(filepath.Join(tempDir, "app.py"), []byte("from .mod import run\n"), 0o644); err != nil {
		t.Fatalf("failed writing app.py: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "mod.py"), []byte("def run():\n    return 1\n"), 0o644); err != nil {
		t.Fatalf("failed writing mod.py: %v", err)
	}

	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	detector := languages.NewRepositoryLanguageDetector(strategy)
	detector.RegisterAdapter(languages.NewGoAdapter())
	detector.RegisterAdapter(languages.NewPythonAdapter())

	first, err := detector.DetectLanguage(tempDir)
	if err != nil {
		t.Fatalf("initial detection failed: %v", err)
	}

	for i := 0; i < 100; i++ {
		adapter, detectErr := detector.DetectLanguage(tempDir)
		if detectErr != nil {
			t.Fatalf("detection failed on iteration %d: %v", i, detectErr)
		}
		if adapter.Name() != first.Name() {
			t.Fatalf("non-deterministic result: got %s, expected %s", adapter.Name(), first.Name())
		}
	}
}

func BenchmarkLanguageDetector_WalkDir(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "benchmark_test")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ignoredDir := filepath.Join(tempDir, "node_modules", "package")
	_ = os.MkdirAll(ignoredDir, 0755)
	for i := 0; i < 1000; i++ {
		_ = os.WriteFile(filepath.Join(ignoredDir, "file"+strconv.Itoa(i)+".py"), []byte("pass"), 0644)
	}

	srcDir := filepath.Join(tempDir, "src")
	_ = os.MkdirAll(srcDir, 0755)
	for i := 0; i < 100; i++ {
		_ = os.WriteFile(filepath.Join(srcDir, "main"+strconv.Itoa(i)+".go"), []byte("package main\n\nfunc main() {}"), 0644)
	}

	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	detector := languages.NewRepositoryLanguageDetector(strategy)
	detector.RegisterAdapter(&DummyAdapter{name: "Go", exts: []string{".go"}})
	detector.RegisterAdapter(&DummyAdapter{name: "Python", exts: []string{".py"}})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = detector.DetectLanguage(tempDir)
	}
}

func TestLanguageDetector_WithPolicy_MonorepoSegmentAware(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "detector_monorepo")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.MkdirAll(filepath.Join(tempDir, "apps", "frontend"), 0o755); err != nil {
		t.Fatalf("failed creating frontend dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "tools"), 0o755); err != nil {
		t.Fatalf("failed creating tools dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "apps", "frontend", "main.ts"), []byte("export const app = 1\n"), 0o644); err != nil {
		t.Fatalf("failed writing ts file: %v", err)
	}
	for i := 0; i < 20; i++ {
		if err := os.WriteFile(filepath.Join(tempDir, "tools", "tool_"+strconv.Itoa(i)+".go"), []byte("package tools\n"), 0o644); err != nil {
			t.Fatalf("failed writing tooling go file: %v", err)
		}
	}

	policy := languages.DetectionPolicy{
		LanguageWeights: map[string]float64{"Go": 0.5, "TypeScript": 3.0, "JavaScript": 1.0, "Python": 1.0},
		TieBreakOrder:   []string{"TypeScript", "Python", "JavaScript", "Go"},
		SegmentWeights:  map[string]float64{"tools": 0.1, "apps": 1.2},
	}

	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	detector := languages.NewRepositoryLanguageDetectorWithPolicy(strategy, policy)
	detector.RegisterAdapter(languages.NewGoAdapter())
	detector.RegisterAdapter(languages.NewTypeScriptAdapter())
	detector.RegisterAdapter(languages.NewJavaScriptAdapter())
	detector.RegisterAdapter(languages.NewPythonAdapter())

	adapter, detectErr := detector.DetectLanguage(tempDir)
	if detectErr != nil {
		t.Fatalf("detect language failed: %v", detectErr)
	}
	if adapter.Name() != "TypeScript" {
		t.Fatalf("expected TypeScript with configured monorepo policy, got %s", adapter.Name())
	}
}

func TestLanguageDetector_DetectLanguage_MixedFixtureParity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "detector_mixed_parity")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.MkdirAll(filepath.Join(tempDir, "src"), 0o755); err != nil {
		t.Fatalf("failed creating src dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "scripts"), 0o755); err != nil {
		t.Fatalf("failed creating scripts dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "src", "app.py"), []byte("from .service import run\n"), 0o644); err != nil {
		t.Fatalf("failed writing python file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "scripts", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed writing go file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "pyproject.toml"), []byte("[project]\nname='demo'\n"), 0o644); err != nil {
		t.Fatalf("failed writing python marker: %v", err)
	}

	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	detector := languages.NewRepositoryLanguageDetector(strategy)
	detector.RegisterAdapter(languages.NewGoAdapter())
	detector.RegisterAdapter(languages.NewPythonAdapter())

	adapter, detectErr := detector.DetectLanguage(tempDir)
	if detectErr != nil {
		t.Fatalf("detect language failed: %v", detectErr)
	}
	if adapter.Name() != "Python" {
		t.Fatalf("expected Python on mixed fixture parity test, got %s", adapter.Name())
	}
}
