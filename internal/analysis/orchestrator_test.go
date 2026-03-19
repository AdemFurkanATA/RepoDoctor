package analysis

import (
	"os"
	"path/filepath"
	"testing"

	"RepoDoctor/internal/languages"
)

func TestOrchestrator_Analyze_SelectsAdapterAndBuildsPipeline(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "main.go"), []byte("package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"ok\")}\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	detector := languages.NewRepositoryLanguageDetector()
	detector.RegisterAdapter(languages.NewGoAdapter())
	detector.RegisterAdapter(languages.NewPythonAdapter())

	orchestrator := NewOrchestrator(detector)
	result, err := orchestrator.Analyze(repo)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if result.AdapterName != "Go" {
		t.Fatalf("expected Go adapter, got %s", result.AdapterName)
	}

	if len(result.Files) == 0 {
		t.Fatal("expected detected files")
	}

	if result.Metrics == nil {
		t.Fatal("expected metrics result")
	}

	if result.Graph == nil {
		t.Fatal("expected dependency graph result")
	}
}
