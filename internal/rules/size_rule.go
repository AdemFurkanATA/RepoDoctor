package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"RepoDoctor/internal/model"
)

// SizeRule checks file and function size thresholds
type SizeRule struct {
	MaxFileLines     int
	MaxFunctionLines int
	fset             *token.FileSet
}

// NewSizeRule creates a new size rule checker with default thresholds
func NewSizeRule() *SizeRule {
	return &SizeRule{
		MaxFileLines:     500,
		MaxFunctionLines: 80,
		fset:             token.NewFileSet(),
	}
}

// ID returns the unique identifier for this rule
func (r *SizeRule) ID() string {
	return "rule.size"
}

// Category returns the category for this rule
func (r *SizeRule) Category() string {
	return string(CategorySize)
}

// Severity returns the severity level for this rule
func (r *SizeRule) Severity() string {
	return string(model.SeverityWarning)
}

// Evaluate executes the rule logic against the provided context
func (r *SizeRule) Evaluate(context AnalysisContext) []model.Violation {
	var violations []model.Violation

	for _, file := range context.RepositoryFiles {
		r.checkFile(file, &violations)
	}

	return violations
}

// checkFile checks a single file for size violations
func (r *SizeRule) checkFile(file RepositoryFile, violations *[]model.Violation) {
	// Check file LOC
	fileLines := r.countNonEmptyLines(file.Content)
	if fileLines > r.MaxFileLines {
		*violations = append(*violations, model.Violation{
			RuleID:      r.ID(),
			Severity:    model.SeverityWarning,
			Message:     "File " + file.Path + " has " + strconv.Itoa(fileLines) + " lines (threshold: " + strconv.Itoa(r.MaxFileLines) + ")",
			File:        file.Path,
			Line:        0,
			ScoreImpact: -3.0,
		})
	}

	// Check function LOC
	r.checkFunctions(file, violations)
}

// countNonEmptyLines counts non-empty lines in a file
func (r *SizeRule) countNonEmptyLines(content string) int {
	lines := strings.Split(content, "\n")
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			count++
		}
	}
	return count
}

// checkFunctions checks function sizes in a file
func (r *SizeRule) checkFunctions(file RepositoryFile, violations *[]model.Violation) {
	// Parse AST
	node, err := parser.ParseFile(r.fset, file.Path, file.Content, 0)
	if err != nil {
		return // Skip malformed files
	}

	// Walk through all declarations
	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Calculate function lines
		startLine := r.fset.Position(funcDecl.Pos()).Line
		endLine := r.fset.Position(funcDecl.End()).Line
		funcLines := endLine - startLine + 1

		if funcLines > r.MaxFunctionLines {
			*violations = append(*violations, model.Violation{
				RuleID:      r.ID(),
				Severity:    model.SeverityWarning,
				Message:     "Function '" + funcDecl.Name.Name + "' has " + strconv.Itoa(funcLines) + " lines (threshold: " + strconv.Itoa(r.MaxFunctionLines) + ")",
				File:        file.Path,
				Line:        startLine,
				ScoreImpact: -3.0,
			})
		}

		return true
	})
}
