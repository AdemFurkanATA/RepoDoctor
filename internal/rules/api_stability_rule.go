package rules

import (
	"sort"
	"strings"

	"RepoDoctor/internal/model"
)

// APIStabilityRule reports deterministic, actionable API breaking changes.
// Snapshot comparison is prepared by orchestration and injected through context configuration.
type APIStabilityRule struct{}

func NewAPIStabilityRule() *APIStabilityRule {
	return &APIStabilityRule{}
}

func (r *APIStabilityRule) ID() string { return "rule.api-stability" }

func (r *APIStabilityRule) Category() string { return string(CategoryArchitecture) }

func (r *APIStabilityRule) Severity() string { return string(model.SeverityCritical) }

func (r *APIStabilityRule) Capabilities() RuleCapabilities {
	return RuleCapabilities{SupportedLanguages: []string{"Go"}, SupportsMultipleLanguages: false}
}

func (r *APIStabilityRule) Evaluate(context AnalysisContext) []model.Violation {
	changes := resolveAPIBreakingChangesFromContext(context.Configuration)
	if len(changes) == 0 {
		return nil
	}

	violations := make([]model.Violation, 0, len(changes))
	for _, change := range changes {
		violations = append(violations, model.Violation{
			RuleID:      r.ID(),
			Severity:    model.SeverityCritical,
			Message:     buildAPIBreakingMessage(change),
			File:        strings.TrimSpace(change.File),
			Line:        change.Line,
			ScoreImpact: -10.0,
		})
	}

	sort.SliceStable(violations, func(i, j int) bool {
		if violations[i].File != violations[j].File {
			return violations[i].File < violations[j].File
		}
		if violations[i].Line != violations[j].Line {
			return violations[i].Line < violations[j].Line
		}
		return violations[i].Message < violations[j].Message
	})

	return violations
}

func resolveAPIBreakingChangesFromContext(configuration Configuration) []model.APIBreakingChange {
	if configuration == nil {
		return nil
	}
	raw, ok := configuration["apiStabilityBreakingChanges"]
	if !ok {
		return nil
	}

	typed, ok := raw.([]model.APIBreakingChange)
	if !ok {
		return nil
	}

	result := make([]model.APIBreakingChange, 0, len(typed))
	for _, change := range typed {
		if strings.TrimSpace(change.SymbolID) == "" {
			continue
		}
		result = append(result, change)
	}
	return result
}

func buildAPIBreakingMessage(change model.APIBreakingChange) string {
	changeType := strings.TrimSpace(change.ChangeType)
	symbolID := strings.TrimSpace(change.SymbolID)
	before := strings.TrimSpace(change.Before)
	after := strings.TrimSpace(change.After)

	switch changeType {
	case "removed":
		if before != "" {
			return "Public API removed: " + symbolID + " (baseline: " + before + "). Action: restore symbol or introduce a compatibility shim."
		}
		return "Public API removed: " + symbolID + ". Action: restore symbol or introduce a compatibility shim."
	case "kind-changed":
		return "Public API kind changed: " + symbolID + " (" + before + " -> " + after + "). Action: keep original symbol kind or release a major version bump."
	case "signature-changed":
		return "Public API signature changed: " + symbolID + " (" + before + " -> " + after + "). Action: provide backward-compatible overload/wrapper or release a major version bump."
	default:
		return "Public API breaking change: " + symbolID + ". Action: review compatibility contract before release."
	}
}
