package rules

import "testing"

func TestDeadCodeRule_FlagsUnusedPrivateFunction(t *testing.T) {
	rule := NewDeadCodeRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{{
		Path:    "pkg/service/a.go",
		Content: "package service\nfunc init() { used() }\nfunc used() { helper() }\nfunc helper() {}\nfunc unusedThing() {\nprintln(\"x\")\n}\n",
	}}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 1 {
		t.Fatalf("expected one dead-code violation, got %d", len(violations))
	}
	if violations[0].RuleID != "rule.dead-code" {
		t.Fatalf("unexpected rule id %q", violations[0].RuleID)
	}
}

func TestDeadCodeRule_DoesNotFlagPrivateFunctionCalledFromAnotherFile(t *testing.T) {
	rule := NewDeadCodeRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{
		{Path: "pkg/service/a.go", Content: "package service\nfunc helper() {}\n"},
		{Path: "pkg/service/b.go", Content: "package service\nfunc init() { helper() }\n"},
	}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected no dead-code violations for cross-file call, got %d", len(violations))
	}
}

func TestDeadCodeRule_RespectsKeepDirectiveAndReflectionException(t *testing.T) {
	rule := NewDeadCodeRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{
		{Path: "pkg/service/keep.go", Content: "package service\n// repodoctor:keep\nfunc maybeUsedElsewhere() {}\n"},
		{Path: "pkg/service/reflective.go", Content: "package service\nimport \"reflect\"\nfunc discoveredAtRuntime() {}\nfunc x() { _ = reflect.ValueOf(discoveredAtRuntime) }\n"},
	}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected no violations due to keep/reflection exceptions, got %d", len(violations))
	}
}
