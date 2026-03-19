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
	caps languages.AdapterCapabilities
}

func (a fakeAdapter) Name() string                                  { return "Fake" }
func (a fakeAdapter) FileExtensions() []string                      { return []string{".fake"} }
func (a fakeAdapter) DetectFiles(repoPath string) ([]string, error) { return []string{}, nil }
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
