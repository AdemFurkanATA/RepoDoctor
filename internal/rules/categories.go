package rules

// RuleCategory represents a category for organizing rules
type RuleCategory string

// Standard rule categories for RepoDoctor
const (
	// CategoryStructural represents rules that analyze structural integrity
	CategoryStructural RuleCategory = "structural"

	// CategoryArchitecture represents rules that enforce architectural boundaries
	CategoryArchitecture RuleCategory = "architecture"

	// CategoryMaintainability represents rules that detect maintainability issues
	CategoryMaintainability RuleCategory = "maintainability"

	// CategorySize represents rules that check file/function size thresholds
	CategorySize RuleCategory = "size"

	// CategoryTesting represents rules that analyze test coverage and quality
	CategoryTesting RuleCategory = "testing"
)

// AllCategories returns a list of all supported rule categories
func AllCategories() []RuleCategory {
	return []RuleCategory{
		CategoryStructural,
		CategoryArchitecture,
		CategoryMaintainability,
		CategorySize,
		CategoryTesting,
	}
}

// IsValidCategory checks if a given string is a valid rule category
func IsValidCategory(category string) bool {
	for _, c := range AllCategories() {
		if string(c) == category {
			return true
		}
	}
	return false
}
