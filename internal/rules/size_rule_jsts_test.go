package rules

import (
	"strings"
	"testing"
)

func TestSizeRule_Evaluate_JSTSFunctionSizeViolation(t *testing.T) {
	rule := NewSizeRule()
	rule.MaxFunctionLines = 5

	content := strings.Join([]string{
		"function tooLong() {",
		"  const a = 1",
		"  const b = 2",
		"  const c = 3",
		"  const d = 4",
		"  const e = 5",
		"  const f = 6",
		"  return a+b+c+d+e+f",
		"}",
	}, "\n")

	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{{Path: "service.ts", Content: content}}}
	violations := rule.Evaluate(ctx)

	if len(violations) == 0 {
		t.Fatal("expected JS/TS function size violation")
	}
	if violations[0].RuleID != "rule.size" {
		t.Fatalf("expected rule.size violation, got %s", violations[0].RuleID)
	}
}
