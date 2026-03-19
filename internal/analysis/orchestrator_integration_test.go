package analysis

import (
	"os"
	"path/filepath"
	"testing"

	"RepoDoctor/internal/languages"
)

func TestOrchestrator_Analyze_PythonDominantRepository(t *testing.T) {
	repo := t.TempDir()

	if err := os.WriteFile(filepath.Join(repo, "app.py"), []byte("import os\nfrom collections import defaultdict\n\n\ndef main():\n    return os.getcwd()\n"), 0644); err != nil {
		t.Fatalf("failed to create app.py: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repo, "util.py"), []byte("import json\n"), 0644); err != nil {
		t.Fatalf("failed to create util.py: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repo, "helper.go"), []byte("package helper\n"), 0644); err != nil {
		t.Fatalf("failed to create helper.go: %v", err)
	}

	detector := languages.NewRepositoryLanguageDetector()
	detector.RegisterAdapter(languages.NewGoAdapter())
	detector.RegisterAdapter(languages.NewPythonAdapter())

	orchestrator := NewOrchestrator(detector)
	result, err := orchestrator.Analyze(repo)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if result.AdapterName != "Python" {
		t.Fatalf("expected Python adapter, got %s", result.AdapterName)
	}

	if result.Graph == nil || result.Graph.NodeCount() == 0 {
		t.Fatal("expected non-empty dependency graph for python repository")
	}
}
