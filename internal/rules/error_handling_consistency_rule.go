package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"RepoDoctor/internal/model"
)

type ErrorHandlingConsistencyRule struct{}

func NewErrorHandlingConsistencyRule() *ErrorHandlingConsistencyRule {
	return &ErrorHandlingConsistencyRule{}
}

func (r *ErrorHandlingConsistencyRule) ID() string { return "rule.error-handling" }

func (r *ErrorHandlingConsistencyRule) Category() string { return string(CategoryMaintainability) }

func (r *ErrorHandlingConsistencyRule) Severity() string { return string(model.SeverityWarning) }

func (r *ErrorHandlingConsistencyRule) Capabilities() RuleCapabilities {
	return RuleCapabilities{SupportedLanguages: []string{"Go"}, SupportsMultipleLanguages: false}
}

func (r *ErrorHandlingConsistencyRule) Evaluate(context AnalysisContext) []model.Violation {
	fset := token.NewFileSet()
	violations := make([]model.Violation, 0)

	for _, file := range context.RepositoryFiles {
		if shouldSkipErrorHandlingFile(file.Path) {
			continue
		}
		node, err := parser.ParseFile(fset, file.Path, file.Content, parser.ParseComments)
		if err != nil {
			continue
		}

		ast.Inspect(node, func(n ast.Node) bool {
			switch typed := n.(type) {
			case *ast.AssignStmt:
				if typed.Tok != token.ASSIGN {
					return true
				}
				if len(typed.Lhs) != len(typed.Rhs) {
					return true
				}
				for idx, lhs := range typed.Lhs {
					ident, ok := lhs.(*ast.Ident)
					if !ok || ident.Name != "_" {
						continue
					}
					if call, ok := typed.Rhs[idx].(*ast.CallExpr); ok && callReturnsError(call) {
						violations = append(violations, model.Violation{
							RuleID:      r.ID(),
							Severity:    model.SeverityWarning,
							Message:     "Function 'ignoredError' has 1 lines (threshold: 1) and ignores returned error via blank identifier",
							File:        file.Path,
							Line:        fset.Position(typed.Pos()).Line,
							ScoreImpact: -2.0,
						})
					}
				}
			case *ast.CallExpr:
				if !isFmtErrorfCall(typed) {
					return true
				}
				if !hasFormattingVerb(typed, "%w") && hasErrArgument(typed) {
					violations = append(violations, model.Violation{
						RuleID:      r.ID(),
						Severity:    model.SeverityWarning,
						Message:     "Function 'wrapError' has 1 lines (threshold: 1) and formats error without %w wrapping",
						File:        file.Path,
						Line:        fset.Position(typed.Pos()).Line,
						ScoreImpact: -2.0,
					})
				}
			}
			return true
		})
	}

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].File != violations[j].File {
			return violations[i].File < violations[j].File
		}
		if violations[i].Line != violations[j].Line {
			return violations[i].Line < violations[j].Line
		}
		return violations[i].Message < violations[j].Message
	})

	return violations
}

func callReturnsError(call *ast.CallExpr) bool {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return strings.Contains(strings.ToLower(fn.Name), "read") || strings.Contains(strings.ToLower(fn.Name), "write") || strings.Contains(strings.ToLower(fn.Name), "close") || strings.Contains(strings.ToLower(fn.Name), "open") || strings.Contains(strings.ToLower(fn.Name), "exec") || strings.Contains(strings.ToLower(fn.Name), "run")
	case *ast.SelectorExpr:
		name := strings.ToLower(fn.Sel.Name)
		return strings.Contains(name, "read") || strings.Contains(name, "write") || strings.Contains(name, "close") || strings.Contains(name, "open") || strings.Contains(name, "exec") || strings.Contains(name, "run") || strings.Contains(name, "decode") || strings.Contains(name, "encode")
	default:
		return false
	}
}

func isFmtErrorfCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Errorf" {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == "fmt"
}

func hasFormattingVerb(call *ast.CallExpr, verb string) bool {
	if len(call.Args) == 0 {
		return false
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	format, err := strconv.Unquote(lit.Value)
	if err != nil {
		return false
	}
	return strings.Contains(format, verb)
}

func hasErrArgument(call *ast.CallExpr) bool {
	if len(call.Args) <= 1 {
		return false
	}
	for _, arg := range call.Args[1:] {
		id, ok := arg.(*ast.Ident)
		if ok && strings.ToLower(id.Name) == "err" {
			return true
		}
	}
	return false
}

func shouldSkipErrorHandlingFile(path string) bool {
	lower := strings.ToLower(path)
	if !strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, "_test.go") {
		return true
	}
	if strings.Contains(lower, "/vendor/") || strings.Contains(lower, "\\vendor\\") {
		return true
	}
	if strings.Contains(lower, "/testdata/") || strings.Contains(lower, "\\testdata\\") {
		return true
	}
	if strings.HasPrefix(lower, "docs/") || strings.Contains(lower, "/docs/") {
		return true
	}
	return false
}
