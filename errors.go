package main

import (
	"fmt"
	"os"
	"strings"
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	ErrorCLIUsage        ErrorCategory = "CLI Usage Error"
	ErrorConfiguration   ErrorCategory = "Configuration Error"
	ErrorAnalysis        ErrorCategory = "Analysis Error"
	ErrorRuntime         ErrorCategory = "Runtime Error"
	ErrorFileNotFound    ErrorCategory = "File Not Found"
	ErrorInvalidArgument ErrorCategory = "Invalid Argument"
)

// CLIError represents a structured CLI error
type CLIError struct {
	Category    ErrorCategory
	Code        string
	Message     string
	Suggestion  string
	OriginalErr error
}

// Error implements the error interface
func (e *CLIError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("%s: %s: %v", e.Category, e.Message, e.OriginalErr)
	}
	return fmt.Sprintf("%s: %s", e.Category, e.Message)
}

// Display prints the error with formatting and suggestions
func (e *CLIError) Display() {
	fmt.Fprintf(os.Stderr, "\n%s: %s\n", e.Category, e.Message)
	
	if e.Suggestion != "" {
		fmt.Fprintf(os.Stderr, "\n💡 Suggestion: %s\n", e.Suggestion)
	}
	
	if e.OriginalErr != nil {
		fmt.Fprintf(os.Stderr, "\nDetails: %v\n", e.OriginalErr)
	}
	fmt.Fprintf(os.Stderr, "\n")
}

// NewCLIError creates a new CLI error with suggestion
func NewCLIError(category ErrorCategory, message, suggestion string, originalErr error) *CLIError {
	return &CLIError{
		Category:    category,
		Code:        string(category),
		Message:     message,
		Suggestion:  suggestion,
		OriginalErr: originalErr,
	}
}

// Error suggestions for common issues
var errorSuggestions = map[string]string{
	"file not found":      "Check if the file path is correct and the file exists",
	"directory not found": "Verify the directory path and ensure it's accessible",
	"permission denied":   "Check file permissions or run with appropriate privileges",
	"invalid path":        "Use an absolute path or a valid relative path",
	"not a directory":     "Provide a directory path instead of a file",
	"config not found":    "Run 'repodoctor init' to create a configuration file",
	"invalid config":      "Check the configuration file syntax and required fields",
	"unknown command":     "Run 'repodoctor --help' to see available commands",
	"missing flag":        "Run 'repodoctor <command> --help' to see required flags",
	"invalid format":      "Use a supported format: text or json",
}

// GetSuggestion returns a suggestion for a given error message
func GetSuggestion(errMessage string) string {
	errLower := strings.ToLower(errMessage)
	
	for key, suggestion := range errorSuggestions {
		if strings.Contains(errLower, key) {
			return suggestion
		}
	}
	
	return "Run 'repodoctor --help' for usage information"
}

// HandleFileNotFoundError creates a file not found error with suggestion
func HandleFileNotFoundError(path string, originalErr error) *CLIError {
	return NewCLIError(
		ErrorFileNotFound,
		fmt.Sprintf("File or directory not found: %s", path),
		"Verify the path exists and is accessible. Use 'ls' or 'dir' to list files.",
		originalErr,
	)
}

// HandleInvalidPathError creates an invalid path error with suggestion
func HandleInvalidPathError(path string, originalErr error) *CLIError {
	return NewCLIError(
		ErrorInvalidArgument,
		fmt.Sprintf("Invalid path: %s", path),
		"Use an absolute path or a valid relative path from current directory",
		originalErr,
	)
}

// HandleConfigNotFoundError creates a config not found error with suggestion
func HandleConfigNotFoundError(configPath string) *CLIError {
	return NewCLIError(
		ErrorConfiguration,
		fmt.Sprintf("Configuration file not found: %s", configPath),
		"Run 'repodoctor init' or create a .repodoctor directory with config",
		nil,
	)
}

// HandleUnknownRuleError creates an unknown rule error with suggestions
func HandleUnknownRuleError(ruleName string, availableRules []string) *CLIError {
	suggestion := "Available rules: " + strings.Join(availableRules, ", ")
	
	return NewCLIError(
		ErrorInvalidArgument,
		fmt.Sprintf("Unknown rule: %s", ruleName),
		suggestion,
		nil,
	)
}

// HandleCLIUsageError creates a CLI usage error with suggestion
func HandleCLIUsageError(message string, originalErr error) *CLIError {
	return NewCLIError(
		ErrorCLIUsage,
		message,
		"Run 'repodoctor <command> --help' for usage information",
		originalErr,
	)
}

// HandleRuntimeError creates a runtime error with suggestion
func HandleRuntimeError(message string, originalErr error) *CLIError {
	return NewCLIError(
		ErrorRuntime,
		message,
		"If this error persists, please report it on GitHub",
		originalErr,
	)
}

// FormatErrorMessage formats an error message with category
func FormatErrorMessage(category ErrorCategory, message string) string {
	return fmt.Sprintf("[%s] %s", category, message)
}

// PrintError prints an error to stderr with formatting
func PrintError(err error) {
	if cliErr, ok := err.(*CLIError); ok {
		cliErr.Display()
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Suggestion: %s\n", GetSuggestion(err.Error()))
	}
}

// ExitWithError prints an error and exits with code 1
func ExitWithError(err error) {
	PrintError(err)
	os.Exit(1)
}

// WrapError wraps an existing error with additional context
func WrapError(originalErr error, category ErrorCategory, message, suggestion string) *CLIError {
	return NewCLIError(category, message, suggestion, originalErr)
}
