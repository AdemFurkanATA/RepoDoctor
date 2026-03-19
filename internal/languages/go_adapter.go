package languages

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"RepoDoctor/internal/model"
)

// GoAdapter implements LanguageAdapter for Go programming language
type GoAdapter struct {
	fset *token.FileSet
}

// NewGoAdapter creates a new Go language adapter
func NewGoAdapter() *GoAdapter {
	return &GoAdapter{
		fset: token.NewFileSet(),
	}
}

// Name returns the language name
func (a *GoAdapter) Name() string {
	return "Go"
}

// FileExtensions returns supported file extensions
func (a *GoAdapter) FileExtensions() []string {
	return []string{".go"}
}

// DetectFiles scans the repository and returns all Go files
func (a *GoAdapter) DetectFiles(repoPath string) ([]string, error) {
	var goFiles []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Check if it's a Go file
		if strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}

		return nil
	})

	return goFiles, err
}

// CollectMetrics extracts Go-specific metrics from source files
func (a *GoAdapter) CollectMetrics(files []string) (*model.RepositoryMetrics, error) {
	metrics := model.NewRepositoryMetrics()

	for _, file := range files {
		fileMetrics, err := a.collectFileMetrics(file)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		metrics.AddFileMetrics(*fileMetrics)
	}

	return metrics, nil
}

// collectFileMetrics extracts metrics from a single Go file
func (a *GoAdapter) collectFileMetrics(path string) (*model.FileMetrics, error) {
	node, err := parser.ParseFile(a.fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	fm := &model.FileMetrics{
		Path:      path,
		Functions: 0,
		Imports:   len(node.Imports),
	}

	// Count lines
	file := a.fset.File(node.Pos())
	if file != nil {
		fm.Lines = file.LineCount()
	}

	// Create metrics collector for this file
	metrics := model.NewRepositoryMetrics()

	// Walk AST to collect function and struct metrics
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			fm.Functions++
			funcMetrics := a.extractFunctionMetrics(x, path)
			metrics.AddFunctionMetrics(*funcMetrics)

		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				structMetrics := a.extractStructMetrics(x, structType, path)
				metrics.AddStructMetrics(*structMetrics)
			}
		}
		return true
	})

	// Update function count
	fm.Functions = len(metrics.Functions)

	return fm, nil
}

// extractFunctionMetrics extracts metrics from a function declaration
func (a *GoAdapter) extractFunctionMetrics(funcDecl *ast.FuncDecl, path string) *model.FunctionMetrics {
	fm := &model.FunctionMetrics{
		Name: funcDecl.Name.Name,
		File: path,
		Line: a.fset.Position(funcDecl.Pos()).Line,
	}

	// Count parameters
	if funcDecl.Type.Params != nil {
		fm.Parameters = funcDecl.Type.Params.NumFields()
	}

	// Estimate lines (rough approximation)
	startPos := a.fset.Position(funcDecl.Pos())
	endPos := a.fset.Position(funcDecl.End())
	fm.Lines = endPos.Line - startPos.Line + 1

	return fm
}

// extractStructMetrics extracts metrics from a struct type
func (a *GoAdapter) extractStructMetrics(typeSpec *ast.TypeSpec, structType *ast.StructType, path string) *model.StructMetrics {
	sm := &model.StructMetrics{
		Name:     typeSpec.Name.Name,
		File:     path,
		Line:     a.fset.Position(typeSpec.Pos()).Line,
		Fields:   structType.Fields.NumFields(),
		Methods:  0, // Methods are counted separately
		Exported: typeSpec.Name.IsExported(),
	}

	return sm
}

// BuildDependencyGraph constructs a dependency graph from Go imports
func (a *GoAdapter) BuildDependencyGraph(files []string) (*model.DependencyGraph, error) {
	graph := model.NewDependencyGraph()

	for _, file := range files {
		node, err := a.parseFileAndAddToGraph(file, graph)
		if err != nil {
			continue
		}

		// Add edges for imports
		if node != nil {
			for _, imp := range node.Imports {
				graph.AddEdge(node.ID, imp)
			}
		}
	}

	return graph, nil
}

// parseFileAndAddToGraph parses a Go file and adds it to the graph
func (a *GoAdapter) parseFileAndAddToGraph(path string, graph *model.DependencyGraph) (*model.Node, error) {
	node, err := parser.ParseFile(a.fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	fileInfo := a.fset.File(node.Pos())
	if fileInfo == nil {
		return nil, nil
	}

	pkgName := node.Name.Name
	nodeID := path

	graphNode := graph.AddNode(nodeID, path, pkgName)

	// Extract imports
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		graphNode.Imports = append(graphNode.Imports, importPath)
	}

	return graphNode, nil
}

// IsStdlibPackage checks if a package is part of Go standard library
func (a *GoAdapter) IsStdlibPackage(importPath string) bool {
	// Standard library packages don't contain dots in their import paths
	// (with some exceptions for internal packages)
	return !strings.Contains(importPath, ".")
}

// Capabilities returns Go adapter capabilities.
func (a *GoAdapter) Capabilities() AdapterCapabilities {
	return AdapterCapabilities{
		SupportsDependencyGraph: true,
		SupportsMetrics:         true,
		UsesASTParsing:          true,
	}
}

// NormalizeImport normalizes Go import declarations.
func (a *GoAdapter) NormalizeImport(importPath string) string {
	return strings.TrimSpace(importPath)
}
