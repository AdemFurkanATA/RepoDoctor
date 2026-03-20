package languages_test

import (
	"os"
	"path/filepath"
	"testing"

	"RepoDoctor/internal/domain"
	"RepoDoctor/internal/languages"
	"RepoDoctor/internal/model"
)

// MockIgnoreStrategy used for tests
type MockIgnoreStrategy struct {
	ignored map[string]bool
}

func (m *MockIgnoreStrategy) ShouldIgnore(dirName string) bool {
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
