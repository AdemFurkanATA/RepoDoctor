package engine

import (
	"fmt"
	"sync"
	"testing"

	"RepoDoctor/internal/model"
	"RepoDoctor/internal/rules"
)

type stubRule struct {
	id   string
	caps rules.RuleCapabilities
	hits *int
}

func TestRuleExecutor_ExecuteByCategory_DeterministicAndFiltered(t *testing.T) {
	registry := rules.NewRuleRegistry()

	firstHits := 0
	secondHits := 0
	otherHits := 0

	registry.MustRegister(&stubRule{id: "rule.cat.2", caps: rules.RuleCapabilities{}, hits: &secondHits})
	registry.MustRegister(&stubRule{id: "rule.cat.1", caps: rules.RuleCapabilities{}, hits: &firstHits})
	registry.MustRegister(&executorCategorizedRule{stubRule: stubRule{id: "rule.other", caps: rules.RuleCapabilities{}, hits: &otherHits}, category: "other"})

	// Replace category for first two rules with testing category via wrappers.
	registry = rules.NewRuleRegistry()
	registry.MustRegister(&executorCategorizedRule{stubRule: stubRule{id: "rule.cat.2", caps: rules.RuleCapabilities{}, hits: &secondHits}, category: "testing"})
	registry.MustRegister(&executorCategorizedRule{stubRule: stubRule{id: "rule.cat.1", caps: rules.RuleCapabilities{}, hits: &firstHits}, category: "testing"})
	registry.MustRegister(&executorCategorizedRule{stubRule: stubRule{id: "rule.other", caps: rules.RuleCapabilities{}, hits: &otherHits}, category: "other"})

	executor := NewRuleExecutor(registry)
	result := executor.ExecuteByCategory(rules.AnalysisContext{}, "testing")

	if result.RulesExecuted != 2 {
		t.Fatalf("expected 2 testing rules executed, got %d", result.RulesExecuted)
	}
	if firstHits != 1 || secondHits != 1 {
		t.Fatalf("expected both testing rules to execute once, got first=%d second=%d", firstHits, secondHits)
	}
	if otherHits != 0 {
		t.Fatalf("expected non-matching category to be skipped, got %d", otherHits)
	}
}

func TestRuleExecutor_ExecuteByIDs_CountInvariant(t *testing.T) {
	registry := rules.NewRuleRegistry()
	hitA := 0
	hitB := 0

	registry.MustRegister(&stubRule{id: "rule.a", caps: rules.RuleCapabilities{}, hits: &hitA})
	registry.MustRegister(&stubRule{id: "rule.b", caps: rules.RuleCapabilities{}, hits: &hitB})

	executor := NewRuleExecutor(registry)
	result := executor.ExecuteByIDs(rules.AnalysisContext{}, []string{"rule.a", "rule.missing", "rule.b"})

	if result.RulesExecuted != 2 {
		t.Fatalf("expected 2 executed rules, got %d", result.RulesExecuted)
	}
	if hitA != 1 || hitB != 1 {
		t.Fatalf("expected both existing IDs to execute once, got a=%d b=%d", hitA, hitB)
	}
}

func TestRuleExecutor_ConcurrentExecute_NoRaceOnSharedRegistry(t *testing.T) {
	registry := rules.NewRuleRegistry()
	const ruleCount = 20
	const workers = 16

	hits := make([]int, ruleCount)
	for i := 0; i < ruleCount; i++ {
		registry.MustRegister(&stubRule{id: fmt.Sprintf("rule.concurrent.%03d", i), caps: rules.RuleCapabilities{SupportsMultipleLanguages: true}, hits: &hits[i]})
	}

	executor := NewRuleExecutor(registry)
	ctx := rules.AnalysisContext{Languages: []string{"Go", "Python", "TypeScript"}}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := executor.Execute(ctx)
			if res.RulesExecuted != ruleCount {
				t.Errorf("expected %d executed rules, got %d", ruleCount, res.RulesExecuted)
			}
		}()
	}
	wg.Wait()
}

type executorCategorizedRule struct {
	stubRule
	category string
}

func (r *executorCategorizedRule) Category() string { return r.category }

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
