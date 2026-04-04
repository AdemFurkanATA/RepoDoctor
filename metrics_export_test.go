package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveMetricsExportPath_DefaultDisabled(t *testing.T) {
	t.Setenv(metricsPathEnv, "")
	path, enabled, err := resolveMetricsExportPath(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enabled || path != "" {
		t.Fatalf("expected metrics export disabled, got enabled=%v path=%q", enabled, path)
	}
}

func TestResolveMetricsExportPath_RootBounded(t *testing.T) {
	root := t.TempDir()
	t.Setenv(metricsPathEnv, filepath.Join(".repodoctor", "metrics", "run.json"))

	path, enabled, err := resolveMetricsExportPath(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !enabled {
		t.Fatal("expected metrics export enabled")
	}
	if !pathWithinRoot(root, path) {
		t.Fatalf("expected metrics path under root, got %s", path)
	}
}

func TestResolveMetricsExportPath_RejectsEscape(t *testing.T) {
	root := t.TempDir()
	t.Setenv(metricsPathEnv, filepath.Join("..", "outside", "metrics.json"))

	_, enabled, err := resolveMetricsExportPath(root)
	if err == nil {
		t.Fatal("expected path escape to fail")
	}
	if enabled {
		t.Fatal("expected disabled result on invalid path")
	}
}

func TestAnalysisService_Run_ExportsMetricsWhenEnabled(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatalf("failed writing go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main(){}\n"), 0o644); err != nil {
		t.Fatalf("failed writing main.go: %v", err)
	}

	t.Setenv(metricsPathEnv, filepath.Join(".repodoctor", "metrics", "run.json"))

	service := NewAnalysisService()
	code := service.Run(AnalyzeRequest{Path: root, Format: "json-v1", ColorEnabled: false})
	if code != 0 {
		t.Fatalf("expected success code, got %d", code)
	}

	metricsPath := filepath.Join(root, ".repodoctor", "metrics", "run.json")
	if info, err := os.Stat(metricsPath); err != nil || info.Size() == 0 {
		t.Fatalf("expected non-empty metrics artifact at %s", metricsPath)
	}
}
