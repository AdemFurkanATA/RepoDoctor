package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"RepoDoctor/internal/model"
)

type DeadCodeRule struct{}

func NewDeadCodeRule() *DeadCodeRule {
	return &DeadCodeRule{}
}

func (r *DeadCodeRule) ID() string { return "rule.dead-code" }

func (r *DeadCodeRule) Category() string { return string(CategoryMaintainability) }

func (r *DeadCodeRule) Severity() string { return string(model.SeverityWarning) }

func (r *DeadCodeRule) Capabilities() RuleCapabilities {
	return RuleCapabilities{SupportedLanguages: []string{"Go"}, SupportsMultipleLanguages: false}
}

type goPackageFacts struct {
	filePaths    []string
	declared     map[string]deadDecl
	called       map[string]bool
	skipAnalysis bool
}

type deadDecl struct {
	file      string
	line      int
	funcLines int
}

func (r *DeadCodeRule) Evaluate(context AnalysisContext) []model.Violation {
	fset := token.NewFileSet()
	packages := map[string]*goPackageFacts{}

	for _, file := range context.RepositoryFiles {
		if shouldSkipDeadCodeFile(file.Path) {
			continue
		}
		node, err := parser.ParseFile(fset, file.Path, file.Content, parser.ParseComments)
		if err != nil {
			continue
		}

		pkgKey := filepath.Dir(file.Path) + ":" + node.Name.Name
		facts, ok := packages[pkgKey]
		if !ok {
			facts = &goPackageFacts{declared: map[string]deadDecl{}, called: map[string]bool{}}
			packages[pkgKey] = facts
		}
		facts.filePaths = append(facts.filePaths, file.Path)

		if importsReflect(node) || strings.Contains(file.Content, "reflect.") {
			facts.skipAnalysis = true
		}

		ast.Inspect(node, func(n ast.Node) bool {
			switch typed := n.(type) {
			case *ast.FuncDecl:
				if typed.Recv != nil {
					return true
				}
				name := typed.Name.Name
				if shouldSkipDeadCodeFunction(name, typed.Doc) {
					return true
				}
				start := fset.Position(typed.Pos()).Line
				end := fset.Position(typed.End()).Line
				facts.declared[name] = deadDecl{file: file.Path, line: start, funcLines: end - start + 1}
			case *ast.CallExpr:
				switch fn := typed.Fun.(type) {
				case *ast.Ident:
					facts.called[fn.Name] = true
				case *ast.SelectorExpr:
					if fn.Sel != nil {
						facts.called[fn.Sel.Name] = true
					}
				}
			}
			return true
		})
	}

	violations := make([]model.Violation, 0)
	keys := make([]string, 0, len(packages))
	for key := range packages {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		facts := packages[key]
		if facts.skipAnalysis {
			continue
		}
		names := make([]string, 0, len(facts.declared))
		for name := range facts.declared {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			if facts.called[name] {
				continue
			}
			decl := facts.declared[name]
			violations = append(violations, model.Violation{
				RuleID:      r.ID(),
				Severity:    model.SeverityWarning,
				Message:     "Function '" + name + "' has " + strconv.Itoa(decl.funcLines) + " lines (threshold: 1) and appears unused",
				File:        decl.file,
				Line:        decl.line,
				ScoreImpact: -2.0,
			})
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

func shouldSkipDeadCodeFile(path string) bool {
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
	return false
}

func shouldSkipDeadCodeFunction(name string, doc *ast.CommentGroup) bool {
	if name == "main" || name == "init" {
		return true
	}
	if strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "Benchmark") || strings.HasPrefix(name, "Example") {
		return true
	}
	if len(name) > 0 && strings.ToUpper(name[:1]) == name[:1] {
		return true
	}
	if doc == nil {
		return false
	}
	for _, c := range doc.List {
		if strings.Contains(strings.ToLower(c.Text), "repodoctor:keep") {
			return true
		}
	}
	return false
}

func importsReflect(node *ast.File) bool {
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		if path == "reflect" {
			return true
		}
	}
	return false
}
