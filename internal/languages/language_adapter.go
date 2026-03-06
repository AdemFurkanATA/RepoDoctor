package languages

import (
	"RepoDoctor/internal/model"
)

// LanguageAdapter defines the interface for language-specific analysis.
// Each programming language (Go, Python, JavaScript, etc.) must implement
// this interface to be analyzed by RepoDoctor's language-agnostic core.
//
// The adapter pattern allows RepoDoctor to:
// - Support multiple programming languages
// - Keep language-specific logic isolated
// - Maintain a unified analysis pipeline
type LanguageAdapter interface {
	// Name returns the programming language name (e.g., "Go", "Python")
	Name() string

	// FileExtensions returns the file extensions this language handles
	// Example: []string{".go"} for Go, []string{".py"} for Python
	FileExtensions() []string

	// DetectFiles scans the repository and returns all files belonging to this language
	DetectFiles(repoPath string) ([]string, error)

	// CollectMetrics extracts structural metrics from source files
	// Returns language-agnostic metrics that feed the Rule Engine
	CollectMetrics(files []string) (*model.RepositoryMetrics, error)

	// BuildDependencyGraph constructs a dependency graph from imports/dependencies
	// Returns a language-agnostic graph structure
	BuildDependencyGraph(files []string) (*model.DependencyGraph, error)

	// IsStdlibPackage determines if a package/module is part of the standard library
	// This helps filter out external dependencies from internal code analysis
	IsStdlibPackage(importPath string) bool
}

// LanguageDetector is responsible for detecting the primary language of a repository
type LanguageDetector interface {
	// DetectLanguage analyzes the repository and returns the detected language
	// Returns the adapter for the detected language
	DetectLanguage(repoPath string) (LanguageAdapter, error)

	// GetSupportedLanguages returns a list of all supported language names
	GetSupportedLanguages() []string
}
