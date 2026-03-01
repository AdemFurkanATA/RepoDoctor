package main

// CycleViolation represents a circular dependency violation
type CycleViolation struct {
	Path     []string
	Severity string
}

// CircularDependencyRule detects circular dependencies in a graph
type CircularDependencyRule struct {
	graph      Graph
	violations []CycleViolation
}

// NewCircularDependencyRule creates a new circular dependency rule checker
func NewCircularDependencyRule(graph Graph) *CircularDependencyRule {
	return &CircularDependencyRule{
		graph:      graph,
		violations: []CycleViolation{},
	}
}

// Name returns the name of this rule
func (r *CircularDependencyRule) Name() string {
	return "circular-dependency"
}

// Severity returns the severity level of this rule
func (r *CircularDependencyRule) Severity() string {
	return "critical"
}

// Check runs the rule and returns true if violations are found
func (r *CircularDependencyRule) Check() bool {
	r.violations = []CycleViolation{}

	cycles := r.graph.DetectCycles()

	for _, cycle := range cycles {
		if len(cycle) > 0 {
			r.violations = append(r.violations, CycleViolation{
				Path:     cycle,
				Severity: r.Severity(),
			})
		}
	}

	return len(r.violations) > 0
}

// Violations returns all detected violations
func (r *CircularDependencyRule) Violations() []CycleViolation {
	return r.violations
}

// Message returns a formatted message describing the violations
func (r *CircularDependencyRule) Message() string {
	if len(r.violations) == 0 {
		return "No circular dependencies detected"
	}

	msg := ""
	for i, v := range r.violations {
		msg += formatCycle(i+1, v.Path)
	}

	return msg
}

// formatCycle formats a cycle path for display
func formatCycle(index int, path []string) string {
	cyclePath := ""
	for i, pkg := range path {
		cyclePath += pkg
		if i < len(path)-1 {
			cyclePath += " → "
		}
	}
	// Complete the cycle
	if len(path) > 0 {
		cyclePath += " → " + path[0]
	}

	return "[" + string(rune(index)) + "] " + cyclePath + "\n"
}
