package rules

import (
	"RepoDoctor/internal/model"
)

// CircularDependencyRule detects circular dependencies in a graph
type CircularDependencyRule struct {
	graph DependencyGraph
}

// NewCircularDependencyRule creates a new circular dependency rule checker
func NewCircularDependencyRule(graph DependencyGraph) *CircularDependencyRule {
	return &CircularDependencyRule{
		graph: graph,
	}
}

// ID returns the unique identifier for this rule
func (r *CircularDependencyRule) ID() string {
	return "rule.circular-dependency"
}

// Category returns the category for this rule
func (r *CircularDependencyRule) Category() string {
	return string(CategoryArchitecture)
}

// Severity returns the severity level for this rule
func (r *CircularDependencyRule) Severity() string {
	return string(model.SeverityCritical)
}

func (r *CircularDependencyRule) Capabilities() RuleCapabilities {
	return RuleCapabilities{SupportedLanguages: []string{"Go", "Python", "JavaScript", "TypeScript"}, SupportsMultipleLanguages: true}
}

// Evaluate executes the rule logic against the provided context
func (r *CircularDependencyRule) Evaluate(context AnalysisContext) []model.Violation {
	var violations []model.Violation

	// Use the dependency graph from context or build one from repository files
	graph := r.buildDependencyGraph(context)
	cycles := r.detectCycles(graph)

	for _, cycle := range cycles {
		if len(cycle) > 0 {
			violations = append(violations, model.Violation{
				RuleID:      r.ID(),
				Severity:    model.SeverityCritical,
				Message:     formatCycle(cycle),
				File:        cycle[0],
				Line:        0,
				ScoreImpact: -10.0,
			})
		}
	}

	return violations
}

// buildDependencyGraph builds a dependency graph from the context
func (r *CircularDependencyRule) buildDependencyGraph(context AnalysisContext) DependencyGraph {
	// If context has a dependency graph, use it
	if context.DependencyGraph.Edges != nil {
		return context.DependencyGraph
	}

	// Otherwise, build from repository files
	edges := make(map[string][]string)
	nodes := make([]string, 0)

	for _, file := range context.RepositoryFiles {
		nodes = append(nodes, file.Path)
		// For now, we'll use imports as dependencies
		// In a real implementation, this would need proper package resolution
		edges[file.Path] = file.Imports
	}

	return DependencyGraph{
		Nodes: nodes,
		Edges: edges,
	}
}

// detectCycles performs DFS-based cycle detection
func (r *CircularDependencyRule) detectCycles(graph DependencyGraph) [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, dep := range graph.Edges[node] {
			if !visited[dep] {
				dfs(dep)
			} else if recStack[dep] {
				// Found a cycle
				cycle := extractCycle(path, dep)
				cycles = append(cycles, cycle)
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
	}

	for _, node := range graph.Nodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}

// extractCycle extracts the cycle from the current path
func extractCycle(path []string, start string) []string {
	for i, node := range path {
		if node == start {
			return path[i:]
		}
	}
	return path
}

// formatCycle formats a cycle path for display
func formatCycle(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}

	cyclePath := ""
	for i, pkg := range cycle {
		cyclePath += pkg
		if i < len(cycle)-1 {
			cyclePath += " → "
		}
	}
	// Complete the cycle
	cyclePath += " → " + cycle[0]

	return cyclePath
}
