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

func TestLayerValidationRule_JSTSProfileAlignment(t *testing.T) {
	rule := NewLayerValidationRule()
	ctx := AnalysisContext{
		Configuration: Configuration{"architectureProfile": "layered"},
		RepositoryFiles: []RepositoryFile{{
			Path:    "repo/repo/user_repo.ts",
			Imports: []string{"handler/user_handler.ts"},
		}},
		Languages: []string{"TypeScript"},
	}

	violations := rule.Evaluate(ctx)
	if len(violations) == 0 {
		t.Fatal("expected layer violation for layered profile in js/ts context")
	}
	if !strings.Contains(violations[0].Message, "[profile:layered]") {
		t.Fatalf("expected layered profile marker in violation message, got %q", violations[0].Message)
	}
}

func TestLayerValidationRule_CustomLayerPolicyDetectsUpwardImport(t *testing.T) {
	rule := NewLayerValidationRule()
	ctx := AnalysisContext{
		Configuration: Configuration{
			"architectureProfile": "layered",
			"customLayerOrder":    []string{"api", "service", "data"},
			"customLayerKeywords": map[string][]string{
				"api":     []string{"api", "controller"},
				"service": []string{"service"},
				"data":    []string{"repo", "data"},
			},
		},
		RepositoryFiles: []RepositoryFile{{
			Path:    "data/user_repo.go",
			Imports: []string{"api/user_controller.go"},
		}},
	}

	violations := rule.Evaluate(ctx)
	if len(violations) != 1 {
		t.Fatalf("expected one custom-policy upward import violation, got %d (%v)", len(violations), violations)
	}
	if !strings.Contains(violations[0].Message, "data") || !strings.Contains(violations[0].Message, "api") {
		t.Fatalf("expected custom layer names in violation message, got %q", violations[0].Message)
	}
}

func TestLayerValidationRule_CustomLayerPolicyAllowsDownwardImport(t *testing.T) {
	rule := NewLayerValidationRule()
	ctx := AnalysisContext{
		Configuration: Configuration{
			"architectureProfile": "layered",
			"customLayerOrder":    []string{"api", "service", "data"},
			"customLayerKeywords": map[string][]string{
				"api":     []string{"api", "controller"},
				"service": []string{"service"},
				"data":    []string{"repo", "data"},
			},
		},
		RepositoryFiles: []RepositoryFile{{
			Path:    "api/user_controller.go",
			Imports: []string{"data/user_repo.go"},
		}},
	}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected no violation for downward custom import, got %v", violations)
	}
}
