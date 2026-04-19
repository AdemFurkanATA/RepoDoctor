package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ComplexityBandSummary struct {
	Low    int `json:"low"`
	Medium int `json:"medium"`
	High   int `json:"high"`
}

type GoFileComplexity struct {
	Path            string
	TotalComplexity int
	FunctionCount   int
}

func collectCyclomaticComplexitySummary(root string) ComplexityBandSummary {
	summary, _ := collectCyclomaticComplexitySummaryWithFiles(root)
	return summary
}

func collectGoFileComplexitySnapshot(root string) []GoFileComplexity {
	_, files := collectCyclomaticComplexitySummaryWithFiles(root)
	return files
}

func collectCyclomaticComplexitySummaryWithFiles(root string) (ComplexityBandSummary, []GoFileComplexity) {
	paths := make([]string, 0)
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := strings.ToLower(d.Name())
			if name == ".git" || name == "vendor" || name == "testdata" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		lower := strings.ToLower(path)
		if !strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, "_test.go") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})

	sort.Strings(paths)
	summary := ComplexityBandSummary{}
	fileComplexities := make([]GoFileComplexity, 0, len(paths))
	fset := token.NewFileSet()

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		node, err := parser.ParseFile(fset, path, content, 0)
		if err != nil {
			continue
		}

		fileTotal := 0
		functionCount := 0
		for _, decl := range node.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			complexity := 1
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch typed := n.(type) {
				case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause, *ast.CommClause:
					complexity++
				case *ast.BinaryExpr:
					if typed.Op.String() == "&&" || typed.Op.String() == "||" {
						complexity++
					}
				}
				return true
			})

			fileTotal += complexity
			functionCount++

			switch {
			case complexity <= 10:
				summary.Low++
			case complexity <= 20:
				summary.Medium++
			default:
				summary.High++
			}
		}

		relPath := normalizeComplexityPath(root, path)
		if functionCount > 0 {
			fileComplexities = append(fileComplexities, GoFileComplexity{
				Path:            relPath,
				TotalComplexity: fileTotal,
				FunctionCount:   functionCount,
			})
		}
	}

	sort.SliceStable(fileComplexities, func(i, j int) bool {
		return fileComplexities[i].Path < fileComplexities[j].Path
	})

	return summary, fileComplexities
}

func normalizeComplexityPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(filepath.Clean(path))
	}
	return filepath.ToSlash(filepath.Clean(rel))
}
