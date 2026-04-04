package languages

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"RepoDoctor/internal/model"
)

// GoAdapter implements LanguageAdapter for Go programming language
type GoAdapter struct {
	parseWorkers int
}

const maxGoParseWorkers = 8

// NewGoAdapter creates a new Go language adapter
func NewGoAdapter() *GoAdapter {
	return &GoAdapter{
		parseWorkers: defaultGoParseWorkers(),
	}
}

func defaultGoParseWorkers() int {
	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		workers = 1
	}
	if workers > maxGoParseWorkers {
		workers = maxGoParseWorkers
	}
	return workers
}

func goWorkerCount(configuredWorkers, total int) int {
	if total <= 1 {
		return 1
	}
	workers := configuredWorkers
	if workers < 1 {
		workers = 1
	}
	if workers > total {
		workers = total
	}
	return workers
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
	if len(files) == 0 {
		return metrics, nil
	}

	type result struct {
		fm  *model.FileMetrics
		err error
	}
	results := make([]result, len(files))

	jobs := make(chan int, len(files))
	for idx := range files {
		jobs <- idx
	}
	close(jobs)

	var wg sync.WaitGroup
	for worker := 0; worker < goWorkerCount(a.parseWorkers, len(files)); worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				fm, err := a.collectFileMetrics(files[idx])
				results[idx] = result{fm: fm, err: err}
			}
		}()
	}
	wg.Wait()

	for _, r := range results {
		if r.err != nil || r.fm == nil {
			continue // Skip files that can't be parsed
		}
		metrics.AddFileMetrics(*r.fm)
	}

	return metrics, nil
}

// collectFileMetrics extracts metrics from a single Go file
func (a *GoAdapter) collectFileMetrics(path string) (*model.FileMetrics, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	fm := &model.FileMetrics{
		Path:      path,
		Functions: 0,
		Imports:   len(node.Imports),
	}

	// Count lines
	file := fset.File(node.Pos())
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
			funcMetrics := goExtractFunctionMetrics(fset, x, path)
			metrics.AddFunctionMetrics(*funcMetrics)

		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				structMetrics := goExtractStructMetrics(fset, x, structType, path)
				metrics.AddStructMetrics(*structMetrics)
			}
		}
		return true
	})

	// Update function count
	fm.Functions = len(metrics.Functions)

	return fm, nil
}

// goExtractFunctionMetrics extracts metrics from a Go function declaration.
// Package-level helper to keep GoAdapter method count within SRP bounds.
func goExtractFunctionMetrics(fset *token.FileSet, funcDecl *ast.FuncDecl, path string) *model.FunctionMetrics {
	fm := &model.FunctionMetrics{
		Name: funcDecl.Name.Name,
		File: path,
		Line: fset.Position(funcDecl.Pos()).Line,
	}

	// Count parameters
	if funcDecl.Type.Params != nil {
		fm.Parameters = funcDecl.Type.Params.NumFields()
	}

	// Estimate lines (rough approximation)
	startPos := fset.Position(funcDecl.Pos())
	endPos := fset.Position(funcDecl.End())
	fm.Lines = endPos.Line - startPos.Line + 1

	return fm
}

// goExtractStructMetrics extracts metrics from a Go struct type.
// Package-level helper to keep GoAdapter method count within SRP bounds.
func goExtractStructMetrics(fset *token.FileSet, typeSpec *ast.TypeSpec, structType *ast.StructType, path string) *model.StructMetrics {
	return &model.StructMetrics{
		Name:     typeSpec.Name.Name,
		File:     path,
		Line:     fset.Position(typeSpec.Pos()).Line,
		Fields:   structType.Fields.NumFields(),
		Methods:  0, // Methods are counted separately
		Exported: typeSpec.Name.IsExported(),
	}
}

// BuildDependencyGraph constructs a dependency graph from Go imports
func (a *GoAdapter) BuildDependencyGraph(files []string) (*model.DependencyGraph, error) {
	graph := model.NewDependencyGraph()
	modulePath := resolveGoModulePath(files)
	entries := parseGoImportEntriesParallel(files, a.parseWorkers)

	for idx, file := range files {
		entry := entries[idx]
		if entry == nil {
			continue
		}

		node := graph.AddNode(file, file, entry.packageName)
		for _, imp := range entry.imports {
			node.Imports = append(node.Imports, imp)
			if node.Metadata == nil {
				node.Metadata = make(map[string]string)
			}
			node.Metadata["import_class:"+imp] = string(classifyGoImportWithModule(imp, modulePath))
		}
		for _, imp := range entry.imports {
			graph.AddEdge(node.ID, imp)
		}
	}

	return graph, nil
}

type goImportEntry struct {
	packageName string
	imports     []string
}

func parseGoImportEntriesParallel(files []string, configuredWorkers int) []*goImportEntry {
	entries := make([]*goImportEntry, len(files))
	if len(files) == 0 {
		return entries
	}

	jobs := make(chan int, len(files))
	for idx := range files {
		jobs <- idx
	}
	close(jobs)

	var wg sync.WaitGroup
	for worker := 0; worker < goWorkerCount(configuredWorkers, len(files)); worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				entries[idx] = parseGoImportEntry(files[idx])
			}
		}()
	}
	wg.Wait()

	return entries
}

func parseGoImportEntry(path string) *goImportEntry {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil
	}

	imports := make([]string, 0, len(node.Imports))
	for _, imp := range node.Imports {
		imports = append(imports, strings.Trim(imp.Path.Value, "\""))
	}

	return &goImportEntry{
		packageName: node.Name.Name,
		imports:     imports,
	}
}

func resolveGoModulePath(files []string) string {
	if len(files) == 0 {
		return ""
	}
	start := filepath.Dir(files[0])
	for {
		goModPath := filepath.Join(start, "go.mod")
		if module := readGoModuleName(goModPath); module != "" {
			return module
		}
		parent := filepath.Dir(start)
		if parent == start {
			break
		}
		start = parent
	}
	return ""
}

func readGoModuleName(goModPath string) string {
	file, err := os.Open(goModPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
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
