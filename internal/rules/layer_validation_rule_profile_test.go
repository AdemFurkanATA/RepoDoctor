package rules

import (
	"strings"
	"testing"
)

func TestLayerValidationRule_ProfileModularMonolithAllowsUpwardImport(t *testing.T) {
	rule := NewLayerValidationRule()
	ctx := AnalysisContext{
		Configuration: Configuration{"architectureProfile": "modular-monolith"},
		RepositoryFiles: []RepositoryFile{{
			Path:    "repo/user_repo.go",
			Imports: []string{"handler/user_handler.go"},
		}},
	}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected no layer violations for modular-monolith profile, got %v", violations)
	}
}

func TestLayerValidationRule_ProfileLayeredIncludesReasonMarker(t *testing.T) {
	rule := NewLayerValidationRule()
	ctx := AnalysisContext{
		Configuration: Configuration{"architectureProfile": "layered"},
		RepositoryFiles: []RepositoryFile{{
			Path:    "repo/user_repo.go",
			Imports: []string{"handler/user_handler.go"},
		}},
	}

	violations := rule.Evaluate(ctx)
	if len(violations) == 0 {
		t.Fatal("expected a layer violation for layered profile")
	}
	if got := violations[0].Message; got == "" || !strings.Contains(got, "[profile:layered]") {
		t.Fatalf("expected profile marker in violation message, got %q", got)
	}
}

func TestLayerValidationRule_PythonProfileAlignment(t *testing.T) {
	rule := NewLayerValidationRule()
	ctx := AnalysisContext{
		Configuration: Configuration{"architectureProfile": "clean"},
		RepositoryFiles: []RepositoryFile{{
			Path:    "repo/repo/user_repo.py",
			Imports: []string{"handler/user_handler.py"},
		}},
		Languages: []string{"Python"},
	}

	violations := rule.Evaluate(ctx)
	if len(violations) == 0 {
		t.Fatal("expected layer violation for clean profile in python context")
	}
	if !strings.Contains(violations[0].Message, "[profile:clean]") {
		t.Fatalf("expected clean profile marker in violation message, got %q", violations[0].Message)
	}
}
