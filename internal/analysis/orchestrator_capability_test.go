package analysis

import (
	"fmt"
	"testing"

	"RepoDoctor/internal/languages"
	"RepoDoctor/internal/model"
)

type fakeDetector struct {
	adapter languages.LanguageAdapter
}

func (d fakeDetector) DetectLanguage(repoPath string) (languages.LanguageAdapter, error) {
	if d.adapter == nil {
		return nil, fmt.Errorf("no adapter")
	}
	return d.adapter, nil
}

func (d fakeDetector) GetSupportedLanguages() []string { return []string{"fake"} }

type fakeAdapter struct {
	caps  languages.AdapterCapabilities
	files []string
}

func (a fakeAdapter) Name() string             { return "Fake" }
func (a fakeAdapter) FileExtensions() []string { return []string{".fake"} }
func (a fakeAdapter) DetectFiles(repoPath string) ([]string, error) {
	if a.files == nil {
		return []string{}, nil
	}
	return append([]string(nil), a.files...), nil
}
func (a fakeAdapter) CollectMetrics(files []string) (*model.RepositoryMetrics, error) {
	return model.NewRepositoryMetrics(), nil
}
func (a fakeAdapter) BuildDependencyGraph(files []string) (*model.DependencyGraph, error) {
	return model.NewDependencyGraph(), nil
}
func (a fakeAdapter) IsStdlibPackage(importPath string) bool      { return false }
func (a fakeAdapter) Capabilities() languages.AdapterCapabilities { return a.caps }
func (a fakeAdapter) NormalizeImport(importPath string) string    { return importPath }

func TestOrchestrator_Analyze_RejectsMissingCapabilities(t *testing.T) {
	orchestrator := NewOrchestrator(fakeDetector{adapter: fakeAdapter{caps: languages.AdapterCapabilities{}}})
	if _, err := orchestrator.Analyze(t.TempDir()); err == nil {
		t.Fatal("expected orchestrator to reject missing capabilities")
	}
}

func TestOrchestrator_Analyze_AppliesMemoryFileBudget(t *testing.T) {
	t.Setenv(memoryFileBudgetEnv, "2")
	orchestrator := NewOrchestrator(fakeDetector{adapter: fakeAdapter{
		caps:  languages.AdapterCapabilities{SupportsDependencyGraph: true, SupportsMetrics: true},
		files: []string{"c.go", "a.go", "b.go"},
	}})

	result, err := orchestrator.Analyze(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected analyze error: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected bounded file count 2, got %d", len(result.Files))
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected memory budget warning")
	}
}

func TestOrchestrator_Analyze_InvalidMemoryBudgetFailSoft(t *testing.T) {
	t.Setenv(memoryFileBudgetEnv, "not-a-number")
	orchestrator := NewOrchestrator(fakeDetector{adapter: fakeAdapter{
		caps:  languages.AdapterCapabilities{SupportsDependencyGraph: true, SupportsMetrics: true},
		files: []string{"a.go"},
	}})

	result, err := orchestrator.Analyze(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected analyze error: %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected file set to remain unchanged, got %d", len(result.Files))
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected invalid-budget warning for fail-soft behavior")
	}
}
