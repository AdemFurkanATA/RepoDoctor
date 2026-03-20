package rules

import (
	"fmt"
	"sync"
	"testing"

	"RepoDoctor/internal/model"
)

type registryStubRule struct {
	id       string
	category string
}

func (r *registryStubRule) ID() string       { return r.id }
func (r *registryStubRule) Category() string { return r.category }
func (r *registryStubRule) Severity() string { return "info" }
func (r *registryStubRule) Evaluate(AnalysisContext) []model.Violation {
	return nil
}

func TestRuleRegistry_GetAll_DeterministicOrder(t *testing.T) {
	registry := NewRuleRegistry()
	registry.MustRegister(&registryStubRule{id: "rule.zeta", category: "testing"})
	registry.MustRegister(&registryStubRule{id: "rule.alpha", category: "testing"})
	registry.MustRegister(&registryStubRule{id: "rule.beta", category: "architecture"})

	ordered := registry.GetAll()
	if len(ordered) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(ordered))
	}
	if ordered[0].ID() != "rule.alpha" || ordered[1].ID() != "rule.beta" || ordered[2].ID() != "rule.zeta" {
		t.Fatalf("unexpected deterministic order: [%s, %s, %s]", ordered[0].ID(), ordered[1].ID(), ordered[2].ID())
	}
}

func TestRuleRegistry_GetByCategory_DeterministicOrder(t *testing.T) {
	registry := NewRuleRegistry()
	registry.MustRegister(&registryStubRule{id: "rule.zeta", category: "architecture"})
	registry.MustRegister(&registryStubRule{id: "rule.alpha", category: "architecture"})
	registry.MustRegister(&registryStubRule{id: "rule.beta", category: "testing"})

	arch := registry.GetByCategory("architecture")
	if len(arch) != 2 {
		t.Fatalf("expected 2 architecture rules, got %d", len(arch))
	}
	if arch[0].ID() != "rule.alpha" || arch[1].ID() != "rule.zeta" {
		t.Fatalf("unexpected category ordering: [%s, %s]", arch[0].ID(), arch[1].ID())
	}
}

func TestRuleRegistry_Register_DuplicateIDFails(t *testing.T) {
	registry := NewRuleRegistry()
	rule := &registryStubRule{id: "rule.duplicate", category: "testing"}
	registry.MustRegister(rule)

	err := registry.Register(&registryStubRule{id: "rule.duplicate", category: "architecture"})
	if err == nil {
		t.Fatal("expected duplicate registration to fail")
	}
}

func TestRuleRegistry_Register_ConcurrentAccessInvariant(t *testing.T) {
	registry := NewRuleRegistry()
	const ruleCount = 50

	var wg sync.WaitGroup
	errCh := make(chan error, ruleCount)

	for i := 0; i < ruleCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id := fmt.Sprintf("rule.concurrent.%03d", idx)
			if err := registry.Register(&registryStubRule{id: id, category: "testing"}); err != nil {
				errCh <- err
			}

			_ = registry.GetAll()
			_ = registry.ListIDs()
			_ = registry.GetByCategory("testing")
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("unexpected concurrent registration error: %v", err)
	}

	if registry.Count() != ruleCount {
		t.Fatalf("expected %d rules after concurrent register, got %d", ruleCount, registry.Count())
	}

	ids := registry.ListIDs()
	if len(ids) != ruleCount {
		t.Fatalf("expected %d sorted IDs, got %d", ruleCount, len(ids))
	}
}
