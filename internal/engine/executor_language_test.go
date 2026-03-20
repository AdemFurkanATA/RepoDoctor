package engine

import (
	"testing"

	"RepoDoctor/internal/model"
	"RepoDoctor/internal/rules"
)

type stubRule struct {
	id   string
	caps rules.RuleCapabilities
	hits *int
}

func (r *stubRule) ID() string                           { return r.id }
func (r *stubRule) Category() string                     { return "testing" }
func (r *stubRule) Severity() string                     { return "info" }
func (r *stubRule) Capabilities() rules.RuleCapabilities { return r.caps }
func (r *stubRule) Evaluate(context rules.AnalysisContext) []model.Violation {
	*r.hits = *r.hits + 1
	return nil
}

func TestRuleExecutor_SelectEligibleRules_ByLanguageContext(t *testing.T) {
	registry := rules.NewRuleRegistry()

	goHits := 0
	pyHits := 0
	sharedHits := 0

	registry.MustRegister(&stubRule{id: "rule.go-only", caps: rules.RuleCapabilities{SupportedLanguages: []string{"Go"}}, hits: &goHits})
	registry.MustRegister(&stubRule{id: "rule.py-only", caps: rules.RuleCapabilities{SupportedLanguages: []string{"Python"}}, hits: &pyHits})
	registry.MustRegister(&stubRule{id: "rule.shared", caps: rules.RuleCapabilities{SupportsMultipleLanguages: true}, hits: &sharedHits})

	executor := NewRuleExecutor(registry)
	executor.Execute(rules.AnalysisContext{Languages: []string{"Go", "TypeScript"}})

	if goHits != 1 {
		t.Fatalf("expected go rule to execute once, got %d", goHits)
	}
	if pyHits != 0 {
		t.Fatalf("expected python rule to be skipped, got %d", pyHits)
	}
	if sharedHits != 1 {
		t.Fatalf("expected shared rule to execute once, got %d", sharedHits)
	}
}
