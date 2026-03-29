package main

import (
	"path/filepath"
	"testing"
)

func TestBuildUnifiedRulesAnalysisContext_DeterministicNodesAndLanguages(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddNode("b.go")
	graph.AddNode("a.go")
	graph.AddEdge("a.go", "b.go")

	ctx := buildUnifiedRulesAnalysisContext(runtimeAnalysisContextInput{
		RepositoryPath:  filepath.Clean("."),
		Graph:           graph,
		PrimaryLanguage: "Python",
	})

	if len(ctx.RepositoryFiles) != 2 {
		t.Fatalf("expected 2 repository files, got %d", len(ctx.RepositoryFiles))
	}
	if ctx.RepositoryFiles[0].Path != "a.go" || ctx.RepositoryFiles[1].Path != "b.go" {
		t.Fatalf("expected deterministic sorted repository files, got %v, %v", ctx.RepositoryFiles[0].Path, ctx.RepositoryFiles[1].Path)
	}
	if len(ctx.Languages) != 1 || ctx.Languages[0] != "Python" {
		t.Fatalf("expected primary language selection [Python], got %v", ctx.Languages)
	}
}

func TestBuildUnifiedRulesAnalysisContext_ConfigurationContractKeys(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddNode("a.go")

	ctx := buildUnifiedRulesAnalysisContext(runtimeAnalysisContextInput{
		RepositoryPath:      filepath.Clean("."),
		Graph:               graph,
		PrimaryLanguage:     "Go",
		ArchitectureProfile: "layered",
	})

	repositoryPath, ok := ctx.Configuration["repositoryPath"]
	if !ok {
		t.Fatalf("configuration key %q missing", "repositoryPath")
	}
	if repositoryPath != filepath.Clean(".") {
		t.Fatalf("configuration key %q drifted: got %v want %v", "repositoryPath", repositoryPath, filepath.Clean("."))
	}

	architectureProfile, ok := ctx.Configuration["architectureProfile"]
	if !ok {
		t.Fatalf("configuration key %q missing", "architectureProfile")
	}
	if architectureProfile != "layered" {
		t.Fatalf("configuration key %q drifted: got %v want %v", "architectureProfile", architectureProfile, "layered")
	}
}

func TestResolveContextLanguages_DefaultIncludesJavaPilotLanguage(t *testing.T) {
	languages := resolveContextLanguages("")
	if len(languages) != 5 {
		t.Fatalf("expected 5 default languages including Java, got %v", languages)
	}
	if languages[4] != "Java" {
		t.Fatalf("expected Java in default language set tail, got %v", languages)
	}
}
