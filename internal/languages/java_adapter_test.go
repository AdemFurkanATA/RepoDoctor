package languages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJavaAdapter_CapabilitiesAndNormalizeImport(t *testing.T) {
	adapter := NewJavaAdapter()
	caps := adapter.Capabilities()
	if !caps.SupportsDependencyGraph || !caps.SupportsMetrics {
		t.Fatal("expected Java adapter to support graph and metrics")
	}

	if got := adapter.NormalizeImport(" static java.util.Collections.*;"); got != "java.util.Collections" {
		t.Fatalf("unexpected normalized Java import: %q", got)
	}
}

func TestJavaAdapter_DetectFilesAndSkipBuildDirs(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "src", "main", "java"), 0o755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "target"), 0o755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}

	prod := filepath.Join(repo, "src", "main", "java", "App.java")
	ignored := filepath.Join(repo, "target", "Generated.java")
	if err := os.WriteFile(prod, []byte("package app;\nclass App {}\n"), 0o644); err != nil {
		t.Fatalf("failed writing prod fixture: %v", err)
	}
	if err := os.WriteFile(ignored, []byte("class Generated {}\n"), 0o644); err != nil {
		t.Fatalf("failed writing ignored fixture: %v", err)
	}

	files, err := adapterDetectFilesForTest(NewJavaAdapter(), repo)
	if err != nil {
		t.Fatalf("DetectFiles failed: %v", err)
	}
	if len(files) != 1 || filepath.Clean(files[0]) != filepath.Clean(prod) {
		t.Fatalf("expected only production java file, got %v", files)
	}
}

func TestJavaAdapter_BuildDependencyGraph_Deterministic(t *testing.T) {
	repo := t.TempDir()
	file := filepath.Join(repo, "App.java")
	content := strings.Join([]string{
		"package app.core;",
		"import java.util.List;",
		"import com.acme.shared.Util;",
		"public class App {",
		"  public void run() {}",
		"}",
	}, "\n")
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing java fixture: %v", err)
	}

	adapter := NewJavaAdapter()
	first, err := adapter.BuildDependencyGraph([]string{file})
	if err != nil {
		t.Fatalf("BuildDependencyGraph failed: %v", err)
	}
	baseline := strings.Join(first.GetNode(file).Imports, "|")

	for i := 0; i < 10; i++ {
		next, nextErr := adapter.BuildDependencyGraph([]string{file})
		if nextErr != nil {
			t.Fatalf("BuildDependencyGraph failed on iteration %d: %v", i, nextErr)
		}
		node := next.GetNode(file)
		if node == nil {
			t.Fatalf("expected graph node for %s", file)
		}
		if current := strings.Join(node.Imports, "|"); current != baseline {
			t.Fatalf("java graph imports must be deterministic, baseline=%q current=%q", baseline, current)
		}
	}
}

func TestJavaAdapter_CollectMetrics_ExtractsTypesAndMethods(t *testing.T) {
	repo := t.TempDir()
	file := filepath.Join(repo, "Service.java")
	content := strings.Join([]string{
		"package app.core;",
		"import java.util.List;",
		"public class Service {",
		"  public Service() {}",
		"  public int run(String a, int b) { return b; }",
		"}",
	}, "\n")
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing java fixture: %v", err)
	}

	adapter := NewJavaAdapter()
	metrics, err := adapter.CollectMetrics([]string{file})
	if err != nil {
		t.Fatalf("CollectMetrics failed: %v", err)
	}
	if metrics.TotalFiles != 1 || metrics.TotalFunctions == 0 || metrics.TotalStructs == 0 {
		t.Fatalf("expected java metrics extraction, got files=%d funcs=%d structs=%d", metrics.TotalFiles, metrics.TotalFunctions, metrics.TotalStructs)
	}
}

func adapterDetectFilesForTest(adapter LanguageAdapter, repo string) ([]string, error) {
	return adapter.DetectFiles(repo)
}
