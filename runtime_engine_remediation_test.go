package main

import (
	"testing"

	"RepoDoctor/internal/model"
)

func TestBuildReportFromRuleViolations_AttachesRemediationHints(t *testing.T) {
	violations := []model.Violation{
		{RuleID: "rule.layer-validation", File: "repo/repo.go", Severity: model.SeverityError, Message: "x"},
		{RuleID: "rule.size", File: "a.go", Severity: model.SeverityWarning, Message: "Function 'big' has 101 lines (threshold: 80)"},
		{RuleID: "rule.god-object", File: "obj.go", Severity: model.SeverityWarning, Message: "Manager has 12 methods (threshold: 10)"},
	}

	report := buildReportFromRuleViolations(".", "0.5.0-dev", nil, violations)
	if len(report.Layer) == 0 || report.Layer[0].Hint == "" {
		t.Fatal("expected layer violation hint to be populated")
	}
	if len(report.Size) == 0 || report.Size[0].Hint == "" {
		t.Fatal("expected size violation hint to be populated")
	}
	if len(report.GodObject) == 0 || report.GodObject[0].Hint == "" {
		t.Fatal("expected god object hint to be populated")
	}
}
