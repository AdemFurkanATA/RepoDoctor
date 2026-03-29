package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunAdapterPipeline_JavaPilotEnabled_SelectsJavaAdapter(t *testing.T) {
	repo := t.TempDir()
	javaFile := filepath.Join(repo, "src", "main", "java", "App.java")
	if err := os.MkdirAll(filepath.Dir(javaFile), 0o755); err != nil {
		t.Fatalf("failed creating java fixture dir: %v", err)
	}
	if err := os.WriteFile(javaFile, []byte("package demo;\nimport java.util.List;\npublic class App {}\n"), 0o644); err != nil {
		t.Fatalf("failed writing java fixture: %v", err)
	}

	configPath := filepath.Join(repo, ".repodoctor", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("failed creating config dir: %v", err)
	}
	config := []byte("language_detection:\n  java_pilot_enabled: true\n  weights:\n    Java: 5\n    Go: 1\n    Python: 1\n    JavaScript: 1\n    TypeScript: 1\n")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatalf("failed writing config: %v", err)
	}

	result, err := runAdapterPipeline(repo)
	if err != nil {
		t.Fatalf("runAdapterPipeline failed: %v", err)
	}
	if result.AdapterName != "Java" {
		t.Fatalf("expected Java adapter when pilot is enabled, got %s", result.AdapterName)
	}
}
