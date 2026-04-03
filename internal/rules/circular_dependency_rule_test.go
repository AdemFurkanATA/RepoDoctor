package rules

import (
	"strings"
	"testing"
)

func TestCircularDependencyRule_MetadataAndCapabilities(t *testing.T) {
	rule := NewCircularDependencyRule(DependencyGraph{})

	if rule.ID() != "rule.circular-dependency" {
		t.Fatalf("unexpected id: %s", rule.ID())
	}
	if rule.Category() != string(CategoryArchitecture) {
		t.Fatalf("unexpected category: %s", rule.Category())
	}
	if rule.Severity() != "critical" {
		t.Fatalf("unexpected severity: %s", rule.Severity())
	}

	caps := rule.Capabilities()
	if !caps.SupportsMultipleLanguages {
		t.Fatal("expected multi-language capability")
	}
	if len(caps.SupportedLanguages) == 0 {
		t.Fatal("expected non-empty supported language list")
	}
}

func TestCircularDependencyRule_Evaluate_UsesContextDependencyGraph(t *testing.T) {
	rule := NewCircularDependencyRule(DependencyGraph{})
	ctx := AnalysisContext{
		DependencyGraph: DependencyGraph{
			Nodes: []string{"a", "b"},
			Edges: map[string][]string{
				"a": []string{"b"},
				"b": []string{"a"},
			},
		},
	}

	violations := rule.Evaluate(ctx)
	if len(violations) != 1 {
		t.Fatalf("expected one circular violation, got %d", len(violations))
	}

	v := violations[0]
	if v.RuleID != rule.ID() {
		t.Fatalf("unexpected rule id: %s", v.RuleID)
	}
	if v.Severity != "critical" {
		t.Fatalf("unexpected severity: %s", v.Severity)
	}
	if v.File != "a" {
		t.Fatalf("expected first cycle node as file, got %s", v.File)
	}
	if !strings.Contains(v.Message, "a") || !strings.Contains(v.Message, "b") {
		t.Fatalf("expected cycle message to include nodes, got %q", v.Message)
	}
}

func TestCircularDependencyRule_Evaluate_BuildsGraphFromRepositoryFiles(t *testing.T) {
	rule := NewCircularDependencyRule(DependencyGraph{})
	ctx := AnalysisContext{
		RepositoryFiles: []RepositoryFile{
			{Path: "a.go", Imports: []string{"b.go"}},
			{Path: "b.go", Imports: []string{"a.go"}},
		},
	}

	violations := rule.Evaluate(ctx)
	if len(violations) != 1 {
		t.Fatalf("expected one circular violation from file imports, got %d", len(violations))
	}
}

func TestCircularDependencyRule_Evaluate_NoCycle(t *testing.T) {
	rule := NewCircularDependencyRule(DependencyGraph{})
	ctx := AnalysisContext{
		DependencyGraph: DependencyGraph{
			Nodes: []string{"a", "b", "c"},
			Edges: map[string][]string{
				"a": []string{"b"},
				"b": []string{"c"},
				"c": []string{},
			},
		},
	}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %v", violations)
	}
}

func TestExtractCycle_StartNotFoundReturnsPath(t *testing.T) {
	path := []string{"a", "b", "c"}
	cycle := extractCycle(path, "x")
	if len(cycle) != len(path) {
		t.Fatalf("expected fallback full path, got %v", cycle)
	}
}

func TestFormatCycle_EmptyCycle(t *testing.T) {
	if got := formatCycle(nil); got != "" {
		t.Fatalf("expected empty message for empty cycle, got %q", got)
	}
}
