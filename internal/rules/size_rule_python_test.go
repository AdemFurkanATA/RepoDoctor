package rules

import (
	"strings"
	"testing"
)

func TestSizeRule_Evaluate_PythonFunctionSizeViolation(t *testing.T) {
	rule := NewSizeRule()
	rule.MaxFunctionLines = 5

	content := strings.Join([]string{
		"def too_long():",
		"    a = 1",
		"    b = 2",
		"    c = 3",
		"    d = 4",
		"    e = 5",
		"    f = 6",
		"    return a + b + c + d + e + f",
	}, "\n")

	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{{Path: "service.py", Content: content}}}
	violations := rule.Evaluate(ctx)

	if len(violations) == 0 {
		t.Fatal("expected python function size violation")
	}
	if violations[0].RuleID != "rule.size" {
		t.Fatalf("expected rule.size violation, got %s", violations[0].RuleID)
	}
}
