package rules

import (
	"strings"
	"testing"

	"RepoDoctor/internal/model"
)

func TestAPIStabilityRule_Evaluate_ProducesCriticalActionableViolations(t *testing.T) {
	rule := NewAPIStabilityRule()
	ctx := AnalysisContext{
		Configuration: Configuration{
			"apiStabilityBreakingChanges": []model.APIBreakingChange{
				{ChangeType: "removed", SymbolID: "demo|function|Legacy", File: "service.go", Line: 10, Before: "func()"},
			},
		},
	}

	violations := rule.Evaluate(ctx)
	if len(violations) != 1 {
		t.Fatalf("expected 1 api stability violation, got %d", len(violations))
	}
	v := violations[0]
	if v.Severity != model.SeverityCritical {
		t.Fatalf("expected critical severity, got %s", v.Severity)
	}
	if !strings.Contains(v.Message, "Action:") {
		t.Fatalf("expected actionable violation message, got %q", v.Message)
	}
}

func TestAPIStabilityRule_Evaluate_HandlesMissingConfiguration(t *testing.T) {
	rule := NewAPIStabilityRule()
	violations := rule.Evaluate(AnalysisContext{Configuration: Configuration{"apiStabilityBreakingChanges": "invalid"}})
	if len(violations) != 0 {
		t.Fatalf("expected no violations for invalid payload, got %d", len(violations))
	}
}
