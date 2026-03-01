package main

import (
	"testing"
)

// TestDependencyGraphAcyclic tests graph with no cycles
func TestDependencyGraphAcyclic(t *testing.T) {
	graph := NewDependencyGraph()

	// Create a simple acyclic graph: A -> B -> C
	graph.AddNode("A")
	graph.AddNode("B")
	graph.AddNode("C")
	graph.AddEdge("A", "B")
	graph.AddEdge("B", "C")

	// Detect cycles - should return empty
	cycles := graph.DetectCycles()

	if len(cycles) != 0 {
		t.Errorf("Expected no cycles in acyclic graph, got %d cycles", len(cycles))
	}

	// Verify node count
	if graph.GetNodeCount() != 3 {
		t.Errorf("Expected 3 nodes, got %d", graph.GetNodeCount())
	}

	// Verify edge count
	if graph.GetEdgeCount() != 2 {
		t.Errorf("Expected 2 edges, got %d", graph.GetEdgeCount())
	}
}

// TestDependencyGraphSimpleCycle tests a simple 2-node cycle
func TestDependencyGraphSimpleCycle(t *testing.T) {
	graph := NewDependencyGraph()

	// Create a simple cycle: A -> B -> A
	graph.AddNode("A")
	graph.AddNode("B")
	graph.AddEdge("A", "B")
	graph.AddEdge("B", "A")

	// Detect cycles - should find one
	cycles := graph.DetectCycles()

	if len(cycles) == 0 {
		t.Error("Expected to detect cycle in graph with A -> B -> A")
	}

	// Verify the cycle contains both nodes
	if len(cycles) > 0 {
		cycle := cycles[0]
		if len(cycle) != 2 {
			t.Errorf("Expected cycle of length 2, got %d", len(cycle))
		}
	}
}

// TestDependencyGraphMultiNodeCycle tests a cycle with multiple nodes
func TestDependencyGraphMultiNodeCycle(t *testing.T) {
	graph := NewDependencyGraph()

	// Create a multi-node cycle: A -> B -> C -> D -> A
	graph.AddNode("A")
	graph.AddNode("B")
	graph.AddNode("C")
	graph.AddNode("D")
	graph.AddEdge("A", "B")
	graph.AddEdge("B", "C")
	graph.AddEdge("C", "D")
	graph.AddEdge("D", "A")

	// Detect cycles
	cycles := graph.DetectCycles()

	if len(cycles) == 0 {
		t.Error("Expected to detect cycle in graph with A -> B -> C -> D -> A")
	}

	// Verify the cycle contains all 4 nodes
	if len(cycles) > 0 {
		cycle := cycles[0]
		if len(cycle) != 4 {
			t.Errorf("Expected cycle of length 4, got %d", len(cycle))
		}
	}
}

// TestDependencyGraphMultipleCycles tests graph with multiple independent cycles
func TestDependencyGraphMultipleCycles(t *testing.T) {
	graph := NewDependencyGraph()

	// Create two independent cycles: A -> B -> A and C -> D -> C
	graph.AddNode("A")
	graph.AddNode("B")
	graph.AddNode("C")
	graph.AddNode("D")
	graph.AddEdge("A", "B")
	graph.AddEdge("B", "A")
	graph.AddEdge("C", "D")
	graph.AddEdge("D", "C")

	// Detect cycles
	cycles := graph.DetectCycles()

	if len(cycles) < 2 {
		t.Errorf("Expected at least 2 cycles, got %d", len(cycles))
	}
}

// TestLayerValidationRuleUpwardImport tests layer violation detection
func TestLayerValidationRuleUpwardImport(t *testing.T) {
	graph := NewDependencyGraph()

	// Create layer structure: handler -> service -> repo
	handlerPath := "project/handler/user_handler.go"
	servicePath := "project/service/user_service.go"
	repoPath := "project/repo/user_repo.go"

	graph.AddNode(handlerPath)
	graph.AddNode(servicePath)
	graph.AddNode(repoPath)

	// Valid: handler -> service (downward)
	graph.AddEdge(handlerPath, servicePath)
	// Valid: service -> repo (downward)
	graph.AddEdge(servicePath, repoPath)

	rule := NewLayerValidationRule(graph)
	hasViolations := rule.Check()

	// Should have no violations for valid downward imports
	if hasViolations {
		t.Errorf("Expected no violations for valid downward imports, got: %v", rule.Violations())
	}
}

// TestLayerValidationRuleRepoToService tests repo -> service violation
func TestLayerValidationRuleRepoToService(t *testing.T) {
	graph := NewDependencyGraph()

	repoPath := "project/repo/user_repo.go"
	servicePath := "project/service/user_service.go"

	graph.AddNode(repoPath)
	graph.AddNode(servicePath)
	// Invalid: repo -> service (upward)
	graph.AddEdge(repoPath, servicePath)

	rule := NewLayerValidationRule(graph)
	hasViolations := rule.Check()

	if !hasViolations {
		t.Error("Expected violation for repo -> service upward import")
	}

	violations := rule.Violations()
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	// Verify violation details
	if len(violations) > 0 {
		v := violations[0]
		if v.From != repoPath {
			t.Errorf("Expected violation From to be %s, got %s", repoPath, v.From)
		}
		if v.To != servicePath {
			t.Errorf("Expected violation To to be %s, got %s", servicePath, v.To)
		}
	}
}

// TestLayerValidationRuleServiceToHandler tests service -> handler violation
func TestLayerValidationRuleServiceToHandler(t *testing.T) {
	graph := NewDependencyGraph()

	servicePath := "project/service/user_service.go"
	handlerPath := "project/handler/user_handler.go"

	graph.AddNode(servicePath)
	graph.AddNode(handlerPath)
	// Invalid: service -> handler (upward)
	graph.AddEdge(servicePath, handlerPath)

	rule := NewLayerValidationRule(graph)
	hasViolations := rule.Check()

	if !hasViolations {
		t.Error("Expected violation for service -> handler upward import")
	}
}

// TestStructuralScoringCircularPenalty tests scoring with circular dependencies
func TestStructuralScoringCircularPenalty(t *testing.T) {
	graph := NewDependencyGraph()

	// Create a cycle
	graph.AddNode("A")
	graph.AddNode("B")
	graph.AddEdge("A", "B")
	graph.AddEdge("B", "A")

	config := (&ConfigLoader{}).getDefaultConfig()
	scorer := NewStructuralScorer(graph, config, "")
	score := scorer.CalculateScore()

	// Should have penalty for circular dependency
	if score.CircularCount == 0 {
		t.Error("Expected circular dependency to be detected")
	}

	if score.CircularPenalty == 0 {
		t.Error("Expected circular penalty to be applied")
	}

	// Score should be less than max
	if score.TotalScore >= score.MaxScore {
		t.Errorf("Expected score < %.1f, got %.1f", score.MaxScore, score.TotalScore)
	}
}

// TestStructuralScoringLayerPenalty tests scoring with layer violations
func TestStructuralScoringLayerPenalty(t *testing.T) {
	graph := NewDependencyGraph()

	// Create layer violation
	repoPath := "project/repo/user_repo.go"
	servicePath := "project/service/user_service.go"
	graph.AddNode(repoPath)
	graph.AddNode(servicePath)
	graph.AddEdge(repoPath, servicePath)

	config := (&ConfigLoader{}).getDefaultConfig()
	scorer := NewStructuralScorer(graph, config, "")
	score := scorer.CalculateScore()

	// Should have penalty for layer violation
	if score.LayerCount == 0 {
		t.Error("Expected layer violation to be detected")
	}

	if score.LayerPenalty == 0 {
		t.Error("Expected layer penalty to be applied")
	}
}

// TestStructuralScoringDeterministic tests that scoring is deterministic
func TestStructuralScoringDeterministic(t *testing.T) {
	graph := NewDependencyGraph()

	// Create same graph multiple times
	graph.AddNode("A")
	graph.AddNode("B")
	graph.AddNode("C")
	graph.AddEdge("A", "B")
	graph.AddEdge("B", "C")
	graph.AddEdge("C", "A")

	config := (&ConfigLoader{}).getDefaultConfig()

	// Run scoring multiple times
	var scores []*StructuralScore
	for i := 0; i < 3; i++ {
		scorer := NewStructuralScorer(graph, config, "")
		scores = append(scores, scorer.CalculateScore())
	}

	// All scores should be identical (deterministic)
	for i := 1; i < len(scores); i++ {
		if scores[i].TotalScore != scores[0].TotalScore {
			t.Errorf("Score not deterministic: run %d = %.1f, run 0 = %.1f",
				i, scores[i].TotalScore, scores[0].TotalScore)
		}
		if scores[i].CircularCount != scores[0].CircularCount {
			t.Errorf("Circular count not deterministic: run %d = %d, run 0 = %d",
				i, scores[i].CircularCount, scores[0].CircularCount)
		}
	}
}

// TestCircularDependencyRuleCriticalSeverity tests that circular deps have critical severity
func TestCircularDependencyRuleCriticalSeverity(t *testing.T) {
	graph := NewDependencyGraph()
	graph.AddNode("A")
	graph.AddNode("B")
	graph.AddEdge("A", "B")
	graph.AddEdge("B", "A")

	rule := NewCircularDependencyRule(graph)

	if rule.Severity() != "critical" {
		t.Errorf("Expected circular dependency severity to be 'critical', got '%s'", rule.Severity())
	}
}

// TestImportExtractorFiltersStdlib tests that standard library imports are filtered
func TestImportExtractorFiltersStdlib(t *testing.T) {
	// This test verifies the import extractor exists and can be created
	extractor := NewImportExtractor("RepoDoctor")
	if extractor == nil {
		t.Error("Failed to create ImportExtractor")
	}

	// Verify stdlib prefixes are built
	if extractor.stdlibPrefixs == nil || len(extractor.stdlibPrefixs) == 0 {
		t.Error("Expected stdlibPrefixs to be populated")
	}
}

// TestGraphInterface tests basic graph operations
func TestGraphInterface(t *testing.T) {
	graph := NewDependencyGraph()

	// Test AddNode
	graph.AddNode("node1")

	// Verify node count instead of HasNode
	if graph.GetNodeCount() != 1 {
		t.Errorf("Expected 1 node after adding node1, got %d", graph.GetNodeCount())
	}

	// Test AddEdge
	graph.AddNode("node2")
	graph.AddEdge("node1", "node2")

	deps := graph.GetDependencies("node1")
	if len(deps) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(deps))
	}

	if deps[0] != "node2" {
		t.Errorf("Expected dependency to be 'node2', got '%s'", deps[0])
	}
}
