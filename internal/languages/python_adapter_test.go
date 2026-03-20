package languages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPythonAdapter_Name(t *testing.T) {
	adapter := NewPythonAdapter()
	if adapter.Name() != "Python" {
		t.Errorf("expected name 'Python', got '%s'", adapter.Name())
	}
}

func TestPythonAdapter_FileExtensions(t *testing.T) {
	adapter := NewPythonAdapter()
	exts := adapter.FileExtensions()
	if len(exts) != 1 || exts[0] != ".py" {
		t.Errorf("expected ['.py'], got %v", exts)
	}
}

func TestPythonAdapter_DetectFiles(t *testing.T) {
	// Create temporary directory with Python files
	tmpDir, err := os.MkdirTemp("", "python-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []string{
		"module1.py",
		"module2.py",
		"test_module.py", // Should be skipped
		"readme.md",      // Should be skipped
	}

	for _, file := range testFiles {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("# test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	adapter := NewPythonAdapter()
	files, err := adapter.DetectFiles(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only find non-test Python files
	if len(files) != 2 {
		t.Errorf("expected 2 Python files, got %d", len(files))
	}
}

func TestPythonAdapter_CollectMetrics(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "python-metrics-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test Python file
	content := `import os
from collections import defaultdict

class MyClass:
    def __init__(self):
        self.value = 0
    
    def my_method(self, param1, param2):
        return param1 + param2

def my_function(x, y):
    return x * y
`
	path := filepath.Join(tmpDir, "test.py")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	adapter := NewPythonAdapter()
	files := []string{path}
	metrics, err := adapter.CollectMetrics(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.TotalFiles != 1 {
		t.Errorf("expected 1 file, got %d", metrics.TotalFiles)
	}

	if len(metrics.Files) > 0 && metrics.Files[0].Imports < 2 {
		t.Errorf("expected at least 2 imports, got %d", metrics.Files[0].Imports)
	}
}

func TestPythonAdapter_ExtractImports(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "python-imports-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	content := `import os
import sys
from collections import defaultdict
from typing import List, Dict
import json
`
	path := filepath.Join(tmpDir, "test.py")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	adapter := NewPythonAdapter()
	imports, pkgName, err := adapter.extractImports(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(imports) < 4 {
		t.Errorf("expected at least 4 imports, got %d: %v", len(imports), imports)
	}

	// Check package name
	expectedPkg := "test"
	if pkgName != expectedPkg {
		t.Errorf("expected package '%s', got '%s'", expectedPkg, pkgName)
	}
}

func TestPythonAdapter_IsStdlibPackage(t *testing.T) {
	adapter := NewPythonAdapter()

	tests := []struct {
		module string
		isStd  bool
	}{
		{"os", true},
		{"sys", true},
		{"json", true},
		{"collections", true},
		{"requests", false}, // Third-party
		{"flask", false},    // Third-party
	}

	for _, test := range tests {
		result := adapter.IsStdlibPackage(test.module)
		if result != test.isStd {
			t.Errorf("module '%s': expected %v, got %v", test.module, test.isStd, result)
		}
	}
}

func TestPythonAdapter_CollectMetrics_EmptyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "python-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "empty.py")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	adapter := NewPythonAdapter()
	metrics, err := adapter.CollectMetrics([]string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.TotalFiles != 1 {
		t.Errorf("expected 1 file, got %d", metrics.TotalFiles)
	}
}

func TestPythonAdapter_DetectFiles_WithSubdirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "python-subdir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Create Python files in root and subdir
	files := []string{
		filepath.Join(tmpDir, "root.py"),
		filepath.Join(subDir, "nested.py"),
	}

	for _, file := range files {
		if err := os.WriteFile(file, []byte("# test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	adapter := NewPythonAdapter()
	detected, err := adapter.DetectFiles(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(detected) != 2 {
		t.Errorf("expected 2 files, got %d", len(detected))
	}
}

func TestPythonAdapter_BuildDependencyGraph(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "python-graph-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	content := `import os
from collections import defaultdict
`
	path := filepath.Join(tmpDir, "test.py")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	adapter := NewPythonAdapter()
	graph, err := adapter.BuildDependencyGraph([]string{path})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if graph == nil {
		t.Fatal("expected non-nil graph")
	}

	if graph.NodeCount() < 1 {
		t.Errorf("expected at least 1 node, got %d", graph.NodeCount())
	}
}

func TestPythonAdapter_ExtractImports_RejectsLargeFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "python-large-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "large.py")
	large := make([]byte, maxPythonFileBytes+1)
	if err := os.WriteFile(path, large, 0644); err != nil {
		t.Fatalf("failed to create large file: %v", err)
	}

	adapter := NewPythonAdapter()
	if _, _, err := adapter.extractImports(path); err == nil {
		t.Fatal("expected extractImports to reject oversized file")
	}
}

func TestPythonAdapter_CollectEvidence_RelativeImports(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "pkg", "app"), 0o755); err != nil {
		t.Fatalf("failed to create package dirs: %v", err)
	}
	for _, rel := range []string{"pkg/__init__.py", "pkg/app/__init__.py"} {
		if err := os.WriteFile(filepath.Join(repo, rel), []byte(""), 0o644); err != nil {
			t.Fatalf("failed to write __init__: %v", err)
		}
	}

	source := filepath.Join(repo, "pkg", "app", "main.py")
	content := "from .service import run\nfrom ..shared import helper\n"
	if err := os.WriteFile(source, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	adapter := NewPythonAdapter()
	signals, warnings, err := adapter.CollectEvidence(repo, []string{source})
	if err != nil {
		t.Fatalf("CollectEvidence failed: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if len(signals) < 2 {
		t.Fatalf("expected evidence signals for relative imports, got %d", len(signals))
	}

	for _, signal := range signals {
		if signal.Language != "Python" {
			t.Fatalf("unexpected language evidence: %s", signal.Language)
		}
		if !strings.HasPrefix(signal.SignalType, "python_import_") {
			t.Fatalf("unexpected signal type: %s", signal.SignalType)
		}
	}
}

func TestPythonAdapter_CollectEvidence_SkipsOutsidePath(t *testing.T) {
	repo := t.TempDir()
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "outside.py")
	if err := os.WriteFile(outsideFile, []byte("import os\n"), 0o644); err != nil {
		t.Fatalf("failed writing outside file: %v", err)
	}

	adapter := NewPythonAdapter()
	signals, warnings, err := adapter.CollectEvidence(repo, []string{outsideFile})
	if err != nil {
		t.Fatalf("CollectEvidence failed: %v", err)
	}
	if len(signals) != 0 {
		t.Fatalf("expected zero signals for outside path, got %d", len(signals))
	}
	if len(warnings) == 0 {
		t.Fatal("expected warning for outside path")
	}
}
