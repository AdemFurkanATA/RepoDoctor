package main

// LayerViolation represents a layer constraint violation
type LayerViolation struct {
	From    string
	To      string
	Message string
}

// LayerConvention represents the allowed dependency direction
type LayerConvention string

const (
	LayerHandler LayerConvention = "handler"
	LayerService LayerConvention = "service"
	LayerRepo    LayerConvention = "repo"
)

// layerOrder defines the hierarchy (lower index = higher layer)
var layerOrder = map[LayerConvention]int{
	LayerHandler: 0,
	LayerService: 1,
	LayerRepo:    2,
}

// LayerValidationRule enforces architectural layering constraints
type LayerValidationRule struct {
	graph      Graph
	violations []LayerViolation
}

// NewLayerValidationRule creates a new layer validation rule checker
func NewLayerValidationRule(graph Graph) *LayerValidationRule {
	return &LayerValidationRule{
		graph:      graph,
		violations: []LayerViolation{},
	}
}

// Name returns the name of this rule
func (r *LayerValidationRule) Name() string {
	return "layer-validation"
}

// Severity returns the severity level of this rule
func (r *LayerValidationRule) Severity() string {
	return "high"
}

// Check runs the rule and returns true if violations are found
func (r *LayerValidationRule) Check() bool {
	r.violations = []LayerViolation{}

	// Check all edges in the graph
	nodes := r.graph.GetAllNodes()
	for _, node := range nodes {
		deps := r.graph.GetDependencies(node)
		fromLayer := detectLayer(node)

		for _, dep := range deps {
			toLayer := detectLayer(dep)

			// Check if this is an upward import (forbidden)
			if isUpwardImport(fromLayer, toLayer) {
				r.violations = append(r.violations, LayerViolation{
					From:    node,
					To:      dep,
					Message: formatLayerViolation(node, dep, fromLayer, toLayer),
				})
			}
		}
	}

	return len(r.violations) > 0
}

// Violations returns all detected violations
func (r *LayerValidationRule) Violations() []LayerViolation {
	return r.violations
}

// Message returns a formatted message describing the violations
func (r *LayerValidationRule) Message() string {
	if len(r.violations) == 0 {
		return "No layer violations detected"
	}

	msg := "Layer violations found:\n"
	for i, v := range r.violations {
		msg += "[" + string(rune(i+48)) + "] " + v.Message + "\n"
	}

	return msg
}

// detectLayer detects the layer of a package based on its path
func detectLayer(pkgPath string) LayerConvention {
	// Check for layer keywords in the path
	if containsLayerKeyword(pkgPath, "handler") {
		return LayerHandler
	}
	if containsLayerKeyword(pkgPath, "service") {
		return LayerService
	}
	if containsLayerKeyword(pkgPath, "repo") {
		return LayerRepo
	}

	// Default to service layer if no specific layer detected
	return LayerService
}

// containsLayerKeyword checks if a path contains a layer keyword
func containsLayerKeyword(path, keyword string) bool {
	// Simple check: look for /keyword/ or /keyword at end
	if len(path) >= len(keyword) {
		for i := 0; i <= len(path)-len(keyword); i++ {
			if i+len(keyword) <= len(path) {
				substr := path[i : i+len(keyword)]
				if substr == keyword {
					// Check if it's a word boundary
					beforeOK := i == 0 || path[i-1] == '/' || path[i-1] == '\\'
					afterOK := i+len(keyword) == len(path) || path[i+len(keyword)] == '/' || path[i+len(keyword)] == '\\'
					if beforeOK && afterOK {
						return true
					}
				}
			}
		}
	}
	return false
}

// isUpwardImport checks if an import goes upward in the layer hierarchy
func isUpwardImport(from, to LayerConvention) bool {
	fromLevel, fromExists := layerOrder[from]
	toLevel, toExists := layerOrder[to]

	if !fromExists || !toExists {
		return false
	}

	// Upward import: from lower layer (higher number) to higher layer (lower number)
	return toLevel < fromLevel
}

// formatLayerViolation formats a layer violation message
func formatLayerViolation(from, to string, fromLayer, toLayer LayerConvention) string {
	return from + " (" + string(fromLayer) + ") -> " + to + " (" + string(toLayer) + "): upward import not allowed"
}
