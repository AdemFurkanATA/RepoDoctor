package rules

import (
	"fmt"
	"sync"
)

// RuleRegistry manages all available rules in the system.
// It provides centralized registration, discovery, and filtering of rules.
type RuleRegistry struct {
	mu    sync.RWMutex
	rules map[string]Rule
}

// NewRuleRegistry creates a new empty rule registry
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		rules: make(map[string]Rule),
	}
}

// Register adds a rule to the registry.
// Returns an error if a rule with the same ID already exists.
func (r *RuleRegistry) Register(rule Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := rule.ID()
	if _, exists := r.rules[id]; exists {
		return fmt.Errorf("rule with ID '%s' is already registered", id)
	}

	r.rules[id] = rule
	return nil
}

// GetByID retrieves a rule by its unique identifier.
// Returns nil if no rule with the given ID exists.
func (r *RuleRegistry) GetByID(id string) Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.rules[id]
}

// GetAll returns all registered rules.
// Returns an empty slice if no rules are registered.
func (r *RuleRegistry) GetAll() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules := make([]Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}
	return rules
}

// GetByCategory returns all rules belonging to a specific category.
// Returns an empty slice if no rules match the category.
func (r *RuleRegistry) GetByCategory(category string) []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Rule
	for _, rule := range r.rules {
		if rule.Category() == category {
			result = append(result, rule)
		}
	}
	return result
}

// Count returns the total number of registered rules
func (r *RuleRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.rules)
}

// ListIDs returns a list of all registered rule IDs
func (r *RuleRegistry) ListIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.rules))
	for id := range r.rules {
		ids = append(ids, id)
	}
	return ids
}

// MustRegister registers a rule and panics if registration fails.
// Useful for initialization code where duplicate rules indicate a programming error.
func (r *RuleRegistry) MustRegister(rule Rule) {
	if err := r.Register(rule); err != nil {
		panic(err)
	}
}
