package analysis

import (
	"os"
	"path/filepath"
	"testing"

	"RepoDoctor/internal/domain"
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

	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	detector := languages.NewRepositoryLanguageDetector(strategy)
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

func TestOrchestrator_Analyze_IgnoresVendoredDirectories(t *testing.T) {
	repo := t.TempDir()

	// 1. Create a Go file in the root
	if err := os.WriteFile(filepath.Join(repo, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	// 2. Create node_modules and flood it with python and JS files
	nodeModulesDir := filepath.Join(repo, "node_modules")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}
	for i := 0; i < 50; i++ {
		// Even though there are 50 python files, they are in node_modules, so they should be ignored
		if err := os.WriteFile(filepath.Join(nodeModulesDir, filepath.Base(t.TempDir())+"_script.py"), []byte("print('hello')\n"), 0644); err != nil {
			t.Fatalf("failed to create python file in node_modules: %v", err)
		}
	}

	// 3. Create venv and flood it with python files
	venvDir := filepath.Join(repo, "venv")
	if err := os.MkdirAll(venvDir, 0755); err != nil {
		t.Fatalf("failed to create venv: %v", err)
	}
	for i := 0; i < 50; i++ {
		if err := os.WriteFile(filepath.Join(venvDir, filepath.Base(t.TempDir())+"_lib.py"), []byte("def test(): pass\n"), 0644); err != nil {
			t.Fatalf("failed to create python file in venv: %v", err)
		}
	}

	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	detector := languages.NewRepositoryLanguageDetector(strategy)
	detector.RegisterAdapter(languages.NewGoAdapter())
	detector.RegisterAdapter(languages.NewPythonAdapter())

	orchestrator := NewOrchestrator(detector)
	result, err := orchestrator.Analyze(repo)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	// If node_modules/venv were NOT ignored, Python would dominate (100 files vs 1 Go file)
	// Since they ARE ignored, Go should be the detected language.
	if result.AdapterName != "Go" {
		t.Fatalf("expected Go adapter to dominate due to ignored directories, got %s", result.AdapterName)
	}
}
