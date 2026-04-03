package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildDependencyGraph_FromImportMetadata(t *testing.T) {
	imports := map[string]*ImportMetadata{
		"a.go": {Package: "a", Imports: []string{"b"}},
	}

	graph := buildDependencyGraph(imports, false)
	if graph.GetNodeCount() != 2 {
		t.Fatalf("expected two nodes in graph, got %d", graph.GetNodeCount())
	}
	if graph.GetEdgeCount() != 1 {
		t.Fatalf("expected one edge in graph, got %d", graph.GetEdgeCount())
	}

	deps := graph.GetDependencies("a.go")
	if len(deps) != 1 || deps[0] != "b" {
		t.Fatalf("expected dependency a.go -> b, got %v", deps)
	}
}

func TestStructuralScorer_GettersAndExplanation(t *testing.T) {
	scorer := NewStructuralScorer(NewDependencyGraph(), nil, "")
	if scorer.GetCircularRule() == nil || scorer.GetLayerRule() == nil {
		t.Fatal("expected circular and layer rule getters to return non-nil")
	}

	_ = scorer.CalculateScore()
	if !strings.Contains(scorer.GetScoreExplanation(), "Structural Score Breakdown") {
		t.Fatalf("expected score explanation output, got %q", scorer.GetScoreExplanation())
	}

	_ = scorer.HasCriticalViolations()
}

func TestStructuralScorer_HasCriticalViolations_WithCycle(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "a")

	scorer := NewStructuralScorer(graph, nil, "")
	if !scorer.HasCriticalViolations() {
		t.Fatal("expected critical violations for cyclic graph")
	}
}

func TestTrendAnalyzer_LoadHistoryMalformed_StartsFresh(t *testing.T) {
	baseDir := t.TempDir()
	analyzer := NewTrendAnalyzer(baseDir)

	historyDir := filepath.Join(baseDir, ".repodoctor")
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatalf("failed creating history dir: %v", err)
	}

	historyPath := filepath.Join(historyDir, "history.json")
	if err := os.WriteFile(historyPath, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("failed writing malformed history: %v", err)
	}

	if err := analyzer.LoadHistory(); err != nil {
		t.Fatalf("expected malformed history to be tolerated, got error: %v", err)
	}

	if got := len(analyzer.GetAllHistory()); got != 0 {
		t.Fatalf("expected empty history after malformed input, got %d entries", got)
	}
}

func TestRuleMetadata_Accessors(t *testing.T) {
	graph := NewDependencyGraph()

	circular := NewCircularDependencyRule(graph)
	if circular.Name() != "circular-dependency" {
		t.Fatalf("unexpected circular rule name: %s", circular.Name())
	}
	if circular.Message() != "No circular dependencies detected" {
		t.Fatalf("unexpected circular no-violation message: %q", circular.Message())
	}

	layer := NewLayerValidationRule(graph)
	if layer.Name() != "layer-validation" {
		t.Fatalf("unexpected layer rule name: %s", layer.Name())
	}
	if layer.Severity() != "high" {
		t.Fatalf("unexpected layer rule severity: %s", layer.Severity())
	}
}

func TestTrendAnalyzer_GetAllHistory(t *testing.T) {
	analyzer := NewTrendAnalyzer(t.TempDir())
	if err := analyzer.AppendScore(90); err != nil {
		t.Fatalf("failed appending score: %v", err)
	}
	history := analyzer.GetAllHistory()
	if len(history) != 1 {
		t.Fatalf("expected one history entry, got %d", len(history))
	}
}
