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

func TestBuildUnifiedRulesAnalysisContext_IncludesCustomArchitectureConfiguration(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddNode("a.go")

	ctx := buildUnifiedRulesAnalysisContext(runtimeAnalysisContextInput{
		RepositoryPath:      filepath.Clean("."),
		Graph:               graph,
		PrimaryLanguage:     "Go",
		ArchitectureProfile: "layered",
		CustomLayerOrder:    []string{"api", "service", "data"},
		CustomLayerKeywords: map[string][]string{
			"api":     []string{"api", "controller"},
			"service": []string{"service"},
			"data":    []string{"repo", "data"},
		},
	})

	rawOrder, ok := ctx.Configuration["customLayerOrder"]
	if !ok {
		t.Fatalf("configuration key %q missing", "customLayerOrder")
	}
	order, ok := rawOrder.([]string)
	if !ok {
		t.Fatalf("configuration key %q has unexpected type %T", "customLayerOrder", rawOrder)
	}
	if len(order) != 3 || order[0] != "api" || order[2] != "data" {
		t.Fatalf("configuration key %q drifted: got %v", "customLayerOrder", order)
	}

	rawKeywords, ok := ctx.Configuration["customLayerKeywords"]
	if !ok {
		t.Fatalf("configuration key %q missing", "customLayerKeywords")
	}
	keywords, ok := rawKeywords.(map[string][]string)
	if !ok {
		t.Fatalf("configuration key %q has unexpected type %T", "customLayerKeywords", rawKeywords)
	}
	if len(keywords["api"]) != 2 || keywords["api"][1] != "controller" {
		t.Fatalf("configuration key %q drifted: got %v", "customLayerKeywords", keywords)
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

func TestBuildUnifiedRulesAnalysisContext_IncludesInterfaceBloatThresholdFromEnv(t *testing.T) {
	t.Setenv("REPODOCTOR_INTERFACE_BLOAT_MAX_METHODS", "14")

	graph := NewDependencyGraph()
	graph.AddNode("a.go")

	ctx := buildUnifiedRulesAnalysisContext(runtimeAnalysisContextInput{
		RepositoryPath:  filepath.Clean("."),
		Graph:           graph,
		PrimaryLanguage: "Go",
	})

	raw, ok := ctx.Configuration["interfaceBloatMaxMethods"]
	if !ok {
		t.Fatalf("configuration key %q missing", "interfaceBloatMaxMethods")
	}
	v, ok := raw.(int)
	if !ok || v != 14 {
		t.Fatalf("unexpected interface bloat threshold in configuration: %v (%T)", raw, raw)
	}
}
