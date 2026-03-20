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
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "detector_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some normal files
	os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tempDir, "utils.go"), []byte("package utils"), 0644)

	// Create an ignored directory with many files
	nodeModulesDir := filepath.Join(tempDir, "node_modules")
	os.MkdirAll(nodeModulesDir, 0755)
	for i := 0; i < 5; i++ {
		os.WriteFile(filepath.Join(nodeModulesDir, "index"+string(rune(i+48))+".py"), []byte("print('hello')"), 0644)
	}

	mockStrategy := &MockIgnoreStrategy{
		ignored: map[string]bool{
			"node_modules": true,
		},
	}

	detector := languages.NewRepositoryLanguageDetector(mockStrategy)

	goAdapter := &DummyAdapter{name: "Go", exts: []string{".go"}}
	pyAdapter := &DummyAdapter{name: "Python", exts: []string{".py"}}

	detector.RegisterAdapter(goAdapter)
	detector.RegisterAdapter(pyAdapter)

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

func BenchmarkLanguageDetector_WalkDir(b *testing.B) {
	// Create a mock filesystem
	tempDir, err := os.MkdirTemp("", "benchmark_test")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create deep ignored directory structure
	ignoredDir := filepath.Join(tempDir, "node_modules", "package")
	os.MkdirAll(ignoredDir, 0755)
	for i := 0; i < 1000; i++ {
		os.WriteFile(filepath.Join(ignoredDir, "file"+string(rune(i))+".py"), []byte("pass"), 0644)
	}

	// Create some go files
	srcDir := filepath.Join(tempDir, "src")
	os.MkdirAll(srcDir, 0755)
	for i := 0; i < 100; i++ {
		os.WriteFile(filepath.Join(srcDir, "main"+string(rune(i))+".go"), []byte("package main\n\nfunc main() {}"), 0644)
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
