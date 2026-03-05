package rules

import (
	"strings"

	"RepoDoctor/internal/model"
)

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
type LayerValidationRule struct{}

// NewLayerValidationRule creates a new layer validation rule checker
func NewLayerValidationRule() *LayerValidationRule {
	return &LayerValidationRule{}
}

// ID returns the unique identifier for this rule
func (r *LayerValidationRule) ID() string {
	return "rule.layer-validation"
}

// Category returns the category for this rule
func (r *LayerValidationRule) Category() string {
	return CategoryArchitecture
}

// Severity returns the severity level for this rule
func (r *LayerValidationRule) Severity() string {
	return string(model.SeverityError)
}

// Evaluate executes the rule logic against the provided context
func (r *LayerValidationRule) Evaluate(context AnalysisContext) []model.Violation {
	var violations []model.Violation

	// Check all files and their imports
	for _, file := range context.RepositoryFiles {
		fromLayer := detectLayer(file.Path)

		for _, imp := range file.Imports {
			toLayer := detectLayer(imp)

			// Check if this is an upward import (forbidden)
			if isUpwardImport(fromLayer, toLayer) {
				violations = append(violations, model.Violation{
					RuleID:      r.ID(),
					Severity:    model.SeverityError,
					Message:     formatLayerViolation(file.Path, imp, fromLayer, toLayer),
					File:        file.Path,
					Line:        0,
					ScoreImpact: -5.0,
				})
			}
		}
	}

	return violations
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
