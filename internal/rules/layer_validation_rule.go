package rules

import (
	"RepoDoctor/internal/model"
	"fmt"
	"strings"
)

// LayerConvention represents the allowed dependency direction
type LayerConvention string

const (
	LayerHandler LayerConvention = "handler"
	LayerService LayerConvention = "service"
	LayerRepo    LayerConvention = "repo"
)

// layerOrder defines the default hierarchy (lower index = higher layer)
var layerOrder = map[LayerConvention]int{LayerHandler: 0, LayerService: 1, LayerRepo: 2}

type layerPolicy struct {
	order      []LayerConvention
	rank       map[LayerConvention]int
	keywords   map[LayerConvention][]string
	defaultFor LayerConvention
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
	return string(CategoryArchitecture)
}

// Severity returns the severity level for this rule
func (r *LayerValidationRule) Severity() string {
	return string(model.SeverityError)
}

func (r *LayerValidationRule) Capabilities() RuleCapabilities {
	return RuleCapabilities{SupportedLanguages: []string{"Go", "Python", "JavaScript", "TypeScript"}, SupportsMultipleLanguages: true}
}

// Evaluate executes the rule logic against the provided context
func (r *LayerValidationRule) Evaluate(context AnalysisContext) []model.Violation {
	var violations []model.Violation
	profile := resolveArchitectureProfile(context)
	policy := resolveLayerPolicy(context)

	// Check all files and their imports
	for _, file := range context.RepositoryFiles {
		fromLayer := detectLayerWithPolicy(file.Path, policy)

		for _, imp := range file.Imports {
			toLayer := detectLayerWithPolicy(imp, policy)

			// Check if this is an upward import (forbidden)
			if isUpwardImportWithProfile(fromLayer, toLayer, profile, policy.rank) {
				violations = append(violations, model.Violation{
					RuleID:      r.ID(),
					Severity:    model.SeverityError,
					Message:     formatLayerViolation(file.Path, imp, fromLayer, toLayer, profile),
					File:        file.Path,
					Line:        0,
					ScoreImpact: -5.0,
				})
			}
		}
	}

	return violations
}

func resolveArchitectureProfile(context AnalysisContext) string {
	if context.Configuration == nil {
		return "layered"
	}
	raw, ok := context.Configuration["architectureProfile"]
	if !ok {
		return "layered"
	}
	profile, ok := raw.(string)
	if !ok || strings.TrimSpace(profile) == "" {
		return "layered"
	}
	return strings.ToLower(strings.TrimSpace(profile))
}

func resolveLayerPolicy(context AnalysisContext) layerPolicy {
	policy := defaultLayerPolicy()
	if context.Configuration == nil {
		return policy
	}

	customOrder := parseCustomLayerOrder(context.Configuration["customLayerOrder"])
	if len(customOrder) < 2 {
		return policy
	}

	customKeywords := parseCustomLayerKeywords(context.Configuration["customLayerKeywords"])
	compiledKeywords := make(map[LayerConvention][]string, len(customOrder))
	for _, layer := range customOrder {
		aliases := customKeywords[layer]
		if len(aliases) == 0 {
			aliases = []string{string(layer)}
		}
		compiledKeywords[layer] = aliases
	}

	customRank := make(map[LayerConvention]int, len(customOrder))
	for idx, layer := range customOrder {
		customRank[layer] = idx
	}

	defaultLayer := customOrder[0]
	if len(customOrder) > 1 {
		defaultLayer = customOrder[1]
	}

	return layerPolicy{order: customOrder, rank: customRank, keywords: compiledKeywords, defaultFor: defaultLayer}
}

func defaultLayerPolicy() layerPolicy {
	return layerPolicy{
		order:      []LayerConvention{LayerHandler, LayerService, LayerRepo},
		rank:       layerOrder,
		keywords:   map[LayerConvention][]string{LayerHandler: []string{"handler"}, LayerService: []string{"service"}, LayerRepo: []string{"repo"}},
		defaultFor: LayerService,
	}
}

// detectLayerWithPolicy detects the layer of a package based on policy keywords.
func detectLayerWithPolicy(pkgPath string, policy layerPolicy) LayerConvention {
	for _, layer := range policy.order {
		aliases := policy.keywords[layer]
		for _, alias := range aliases {
			if containsLayerKeyword(pkgPath, alias) {
				return layer
			}
		}
	}

	return policy.defaultFor
}

// containsLayerKeyword checks if a path contains a layer keyword
func containsLayerKeyword(path, keyword string) bool {
	path = strings.ToLower(path)
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return false
	}

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
func isUpwardImport(from, to LayerConvention, rank map[LayerConvention]int) bool {
	fromLevel, fromExists := rank[from]
	toLevel, toExists := rank[to]

	if !fromExists || !toExists {
		return false
	}

	// Upward import: from lower layer (higher number) to higher layer (lower number)
	return toLevel < fromLevel
}

func isUpwardImportWithProfile(from, to LayerConvention, profile string, rank map[LayerConvention]int) bool {
	if profile == "modular-monolith" {
		return false
	}
	return isUpwardImport(from, to, rank)
}

func parseCustomLayerOrder(raw interface{}) []LayerConvention {
	layers, ok := raw.([]string)
	if !ok || len(layers) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(layers))
	parsed := make([]LayerConvention, 0, len(layers))
	for _, layer := range layers {
		normalized := strings.ToLower(strings.TrimSpace(layer))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		parsed = append(parsed, LayerConvention(normalized))
	}
	if len(parsed) < 2 {
		return nil
	}
	return parsed
}

func parseCustomLayerKeywords(raw interface{}) map[LayerConvention][]string {
	keywords, ok := raw.(map[string][]string)
	if !ok || len(keywords) == 0 {
		return nil
	}
	parsed := make(map[LayerConvention][]string, len(keywords))
	for layer, aliases := range keywords {
		normalizedLayer := strings.ToLower(strings.TrimSpace(layer))
		if normalizedLayer == "" {
			continue
		}
		seenAlias := map[string]bool{}
		normalizedAliases := make([]string, 0, len(aliases))
		for _, alias := range aliases {
			normalizedAlias := strings.ToLower(strings.TrimSpace(alias))
			if normalizedAlias == "" || seenAlias[normalizedAlias] {
				continue
			}
			seenAlias[normalizedAlias] = true
			normalizedAliases = append(normalizedAliases, normalizedAlias)
		}
		if len(normalizedAliases) > 0 {
			parsed[LayerConvention(normalizedLayer)] = normalizedAliases
		}
	}
	if len(parsed) == 0 {
		return nil
	}
	return parsed
}

// formatLayerViolation formats a layer violation message
func formatLayerViolation(from, to string, fromLayer, toLayer LayerConvention, profile string) string {
	if strings.TrimSpace(profile) == "" {
		profile = "layered"
	}
	return fmt.Sprintf("%s (%s) -> %s (%s): upward import not allowed [profile:%s]", from, fromLayer, to, toLayer, profile)
}
