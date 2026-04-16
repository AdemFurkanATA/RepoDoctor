package rules

import "testing"

func TestErrorHandlingConsistencyRule_FlagsIgnoredErrorAndMissingWrap(t *testing.T) {
	rule := NewErrorHandlingConsistencyRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{{
		Path:    "pkg/app/service.go",
		Content: "package app\nimport \"fmt\"\nfunc x(err error) error {\n_ = writeData()\nreturn fmt.Errorf(\"failed: %v\", err)\n}\nfunc writeData() error { return nil }\n",
	}}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}
}

func TestErrorHandlingConsistencyRule_DoesNotFlagWrappedErrors(t *testing.T) {
	rule := NewErrorHandlingConsistencyRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{{
		Path:    "pkg/app/service.go",
		Content: "package app\nimport \"fmt\"\nfunc x(err error) error {\nif err != nil {\nreturn fmt.Errorf(\"failed: %w\", err)\n}\nreturn nil\n}\n",
	}}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected no violations for wrapped error handling, got %d", len(violations))
	}
}
