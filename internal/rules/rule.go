package rules

import "RepoDoctor/internal/model"

// AnalysisContext provides read-only access to repository data for rules.
// It encapsulates all information needed for rule evaluation without
// allowing mutation of the repository code.
type AnalysisContext struct {
	// RepositoryFiles contains all Go files in the repository
	RepositoryFiles []RepositoryFile
	// RepositoryMetrics contains computed metrics for the repository
	RepositoryMetrics RepositoryMetrics
	// DependencyGraph contains the dependency graph for the repository
	DependencyGraph DependencyGraph
	// Configuration contains rule configuration settings
	Configuration Configuration
	// Languages contains detected language context for multi-language-aware rule dispatch.
	Languages []string
}

type RuleCapabilities struct {
	SupportedLanguages        []string
	SupportsMultipleLanguages bool
}

// LanguageAwareRule is an optional extension for rules that participate in
// language-targeted dispatch while keeping core engine language-agnostic.
type LanguageAwareRule interface {
	Rule
	Capabilities() RuleCapabilities
}

// RepositoryFile represents a Go file in the repository
type RepositoryFile struct {
	// Path is the file path relative to repository root
	Path string
	// Content contains the file content
	Content string
	// Imports contains the list of import paths
	Imports []string
}

// RepositoryMetrics contains computed metrics for analysis
type RepositoryMetrics struct {
	// TotalLines is the total number of lines across all files
	TotalLines int
	// FileCount is the number of Go files
	FileCount int
	// FunctionCount is the number of functions
	FunctionCount int
}

// DependencyGraph represents the dependency relationships between packages
type DependencyGraph struct {
	// Nodes contains all package nodes
	Nodes []string
	// Edges contains dependency relationships
	Edges map[string][]string
}

// Configuration contains rule-specific configuration
type Configuration map[string]interface{}

// Rule defines the interface that all analysis rules must implement.
// Rules are pure analysis components that evaluate repository data
// and return violations without mutating the repository.
type Rule interface {
	// ID returns the unique identifier for the rule.
	// IDs must be stable across versions (e.g., "rule.god-object").
	ID() string

	// Category returns the rule category for organization and filtering.
	// Examples: "structural", "architecture", "maintainability", "size".
	Category() string

	// Severity returns the severity level of the rule.
	// Expected values: "info", "warning", "error", "critical".
	Severity() string

	// Evaluate executes the rule logic against the provided context.
	// It must handle missing data safely and never panic.
	// Returns a slice of violations found (may be empty).
	Evaluate(context AnalysisContext) []model.Violation
}

// ExampleRule demonstrates a minimal rule implementation.
// This serves as documentation for how to implement the Rule interface.
type ExampleRule struct{}

// ID returns the unique identifier for this rule
func (r *ExampleRule) ID() string {
	return "rule.example"
}

// Category returns the category for this rule
func (r *ExampleRule) Category() string {
	return "structural"
}

// Severity returns the severity level for this rule
func (r *ExampleRule) Severity() string {
	return "info"
}

// Evaluate executes the rule logic (placeholder implementation)
func (r *ExampleRule) Evaluate(context AnalysisContext) []model.Violation {
	// Placeholder implementation - always returns no violations
	return []model.Violation{}
}
