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

const defaultInterfaceBloatMaxMethods = 10

type InterfaceBloatRule struct {
	MaxMethods int
}

func NewInterfaceBloatRule() *InterfaceBloatRule {
	return &InterfaceBloatRule{MaxMethods: defaultInterfaceBloatMaxMethods}
}

func (r *InterfaceBloatRule) ID() string { return "rule.interface-bloat" }

func (r *InterfaceBloatRule) Category() string { return string(CategoryMaintainability) }

func (r *InterfaceBloatRule) Severity() string { return string(model.SeverityWarning) }

func (r *InterfaceBloatRule) Capabilities() RuleCapabilities {
	return RuleCapabilities{SupportedLanguages: []string{"Go"}, SupportsMultipleLanguages: false}
}

func (r *InterfaceBloatRule) Evaluate(context AnalysisContext) []model.Violation {
	maxMethods := r.resolveMaxMethods(context)
	fset := token.NewFileSet()
	violations := make([]model.Violation, 0)

	for _, file := range context.RepositoryFiles {
		if shouldSkipErrorHandlingFile(file.Path) {
			continue
		}

		node, err := parser.ParseFile(fset, file.Path, file.Content, 0)
		if err != nil {
			continue
		}

		for _, decl := range node.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				iface, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}

				methodCount := countInterfaceMethods(iface)
				if methodCount <= maxMethods {
					continue
				}

				line := fset.Position(typeSpec.Pos()).Line
				violations = append(violations, model.Violation{
					RuleID:      r.ID(),
					Severity:    model.SeverityWarning,
					Message:     "Function 'Interface " + typeSpec.Name.Name + "' has " + strconv.Itoa(methodCount) + " lines (threshold: " + strconv.Itoa(maxMethods) + ")",
					File:        file.Path,
					Line:        line,
					ScoreImpact: -2.0,
				})
			}
		}
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

func (r *InterfaceBloatRule) resolveMaxMethods(context AnalysisContext) int {
	maxMethods := r.MaxMethods
	if maxMethods <= 0 {
		maxMethods = defaultInterfaceBloatMaxMethods
	}
	if context.Configuration == nil {
		return maxMethods
	}
	raw, ok := context.Configuration["interfaceBloatMaxMethods"]
	if !ok {
		return maxMethods
	}

	switch typed := raw.(type) {
	case int:
		if typed > 0 {
			return typed
		}
	case int64:
		if typed > 0 {
			return int(typed)
		}
	case float64:
		if typed > 0 {
			return int(typed)
		}
	case string:
		v, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil && v > 0 {
			return v
		}
	}

	return maxMethods
}

func countInterfaceMethods(iface *ast.InterfaceType) int {
	if iface == nil || iface.Methods == nil {
		return 0
	}
	count := 0
	for _, method := range iface.Methods.List {
		if len(method.Names) > 0 {
			count += len(method.Names)
		}
	}
	return count
}
