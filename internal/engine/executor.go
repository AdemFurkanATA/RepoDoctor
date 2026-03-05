package engine

import (
	"RepoDoctor/internal/model"
	"RepoDoctor/internal/rules"
)

// RuleExecutor is responsible for executing all registered rules.
// It guarantees deterministic order and rule isolation during execution.
type RuleExecutor struct {
	registry *rules.RuleRegistry
}

// NewRuleExecutor creates a new rule executor with the given registry
func NewRuleExecutor(registry *rules.RuleRegistry) *RuleExecutor {
	return &RuleExecutor{
		registry: registry,
	}
}

// ExecutionResult contains the results of rule execution
type ExecutionResult struct {
	// Violations contains all violations detected by all rules
	Violations []model.Violation
	// RulesExecuted is the number of rules that were executed
	RulesExecuted int
}

// Execute runs all registered rules against the provided analysis context.
// Rules are executed sequentially to maintain deterministic output.
// Returns aggregated violations from all rules.
func (e *RuleExecutor) Execute(context rules.AnalysisContext) *ExecutionResult {
	allRules := e.registry.GetAll()
	allViolations := make([]model.Violation, 0)

	for _, rule := range allRules {
		violations := e.executeRule(rule, context)
		allViolations = append(allViolations, violations...)
	}

	return &ExecutionResult{
		Violations:    allViolations,
		RulesExecuted: len(allRules),
	}
}

// ExecuteByCategory runs only rules belonging to a specific category
func (e *RuleExecutor) ExecuteByCategory(context rules.AnalysisContext, category string) *ExecutionResult {
	categoryRules := e.registry.GetByCategory(category)
	allViolations := make([]model.Violation, 0)

	for _, rule := range categoryRules {
		violations := e.executeRule(rule, context)
		allViolations = append(allViolations, violations...)
	}

	return &ExecutionResult{
		Violations:    allViolations,
		RulesExecuted: len(categoryRules),
	}
}

// ExecuteByIDs runs only rules with the specified IDs
func (e *RuleExecutor) ExecuteByIDs(context rules.AnalysisContext, ruleIDs []string) *ExecutionResult {
	allViolations := make([]model.Violation, 0)
	executedCount := 0

	for _, id := range ruleIDs {
		rule := e.registry.GetByID(id)
		if rule != nil {
			violations := e.executeRule(rule, context)
			allViolations = append(allViolations, violations...)
			executedCount++
		}
	}

	return &ExecutionResult{
		Violations:    allViolations,
		RulesExecuted: executedCount,
	}
}

// executeRule executes a single rule and handles any panics gracefully
func (e *RuleExecutor) executeRule(rule rules.Rule, context rules.AnalysisContext) []model.Violation {
	// Recover from any panics in the rule to prevent pipeline failure
	defer func() {
		if r := recover(); r != nil {
			// Log the panic but don't propagate it
			// In production, this should be logged properly
		}
	}()

	// Execute the rule
	return rule.Evaluate(context)
}
