package model

// Severity represents the severity level of a rule violation
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Violation represents a standardized rule violation
type Violation struct {
	// RuleID is the unique identifier of the rule that detected the violation
	RuleID string
	// Severity is the severity level of the violation
	Severity Severity
	// Message is a human-readable description of the violation
	Message string
	// File is the path to the file where the violation occurred
	File string
	// Line is the line number where the violation occurred (0 if not applicable)
	Line int
	// ScoreImpact is the impact on the structural health score
	ScoreImpact float64
}
