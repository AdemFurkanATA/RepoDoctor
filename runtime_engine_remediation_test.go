package main

import (
	"testing"

	"RepoDoctor/internal/model"
)

func TestBuildReportFromRuleViolations_AttachesRemediationHints(t *testing.T) {
	violations := []model.Violation{
		{RuleID: "rule.layer-validation", File: "repo/repo.go", Severity: model.SeverityError, Message: "x"},
		{RuleID: "rule.secret-detection", File: "repo/secrets.go", Severity: model.SeverityCritical, Message: "GitHub token pattern detected"},
		{RuleID: "rule.code-duplication", File: "repo/dup_a.go", Severity: model.SeverityWarning, Message: "File repo/dup_a.go has 8 lines (threshold: 6) duplicated with repo/dup_b.go"},
		{RuleID: "rule.dead-code", File: "repo/dead.go", Severity: model.SeverityWarning, Message: "Function 'unusedThing' has 3 lines (threshold: 1) and appears unused"},
		{RuleID: "rule.error-handling", File: "repo/err.go", Severity: model.SeverityWarning, Message: "Function 'wrapError' has 1 lines (threshold: 1) and formats error without %w wrapping"},
		{RuleID: "rule.size", File: "a.go", Severity: model.SeverityWarning, Message: "Function 'big' has 101 lines (threshold: 80)"},
		{RuleID: "rule.god-object", File: "obj.go", Severity: model.SeverityWarning, Message: "Manager has 12 methods (threshold: 10)"},
	}

	report := buildReportFromRuleViolations(".", "1.1.0", nil, violations)
	if len(report.Layer) == 0 || report.Layer[0].Hint == "" {
		t.Fatal("expected layer violation hint to be populated")
	}
	if len(report.Layer) < 2 || report.Layer[1].Hint == "" {
		t.Fatal("expected secret-detection remediation hint to be populated")
	}
	if len(report.Size) == 0 || report.Size[0].Hint == "" {
		t.Fatal("expected size violation hint to be populated")
	}
	if len(report.GodObject) == 0 || report.GodObject[0].Hint == "" {
		t.Fatal("expected god object hint to be populated")
	}
}
