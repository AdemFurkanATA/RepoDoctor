package languages

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"RepoDoctor/internal/model"
)

const (
	maxJavaFileBytes  = 2 * 1024 * 1024
	maxJavaFilesScan  = 10000
	maxJavaImportRows = 50000
)

var (
	javaPackagePattern = regexp.MustCompile(`^\s*package\s+([A-Za-z_][\w.]*)\s*;`)
	javaImportPattern  = regexp.MustCompile(`^\s*import\s+(?:static\s+)?([A-Za-z_][\w.*]*)\s*;`)
	javaTypePattern    = regexp.MustCompile(`\b(class|interface|enum|record)\s+([A-Za-z_][\w]*)`)
	javaMethodPattern  = regexp.MustCompile(`\b(?:public|protected|private)?\s*(?:static\s+)?(?:final\s+)?[A-Za-z_][\w<>\[\]]*\s+([A-Za-z_][\w]*)\s*\(`)
)

type JavaAdapter struct{}

func NewJavaAdapter() LanguageAdapter {
	return &JavaAdapter{}
}

func (a *JavaAdapter) Name() string {
	return "Java"
}

func (a *JavaAdapter) FileExtensions() []string {
	return []string{".java"}
}

func (a *JavaAdapter) DetectFiles(repoPath string) ([]string, error) {
	javaFiles := make([]string, 0)
	seen := 0
	err := filepath.WalkDir(repoPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := strings.ToLower(d.Name())
			if strings.HasPrefix(name, ".") || name == "target" || name == "build" || name == "out" || name == ".gradle" {
				if path != repoPath {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) == ".java" {
			seen++
			if seen > maxJavaFilesScan {
				return nil
			}
			javaFiles = append(javaFiles, filepath.Clean(path))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(javaFiles)
	return javaFiles, nil
}

func (a *JavaAdapter) CollectMetrics(files []string) (*model.RepositoryMetrics, error) {
	metrics := model.NewRepositoryMetrics()
	for _, file := range files {
		fm, functions, structs, err := a.collectFileMetrics(file)
		if err != nil {
			continue
		}
		metrics.AddFileMetrics(*fm)
		for _, fn := range functions {
			metrics.AddFunctionMetrics(fn)
		}
		for _, st := range structs {
			metrics.AddStructMetrics(st)
		}
	}
	return metrics, nil
}

func (a *JavaAdapter) collectFileMetrics(path string) (*model.FileMetrics, []model.FunctionMetrics, []model.StructMetrics, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(content) > maxJavaFileBytes {
		return nil, nil, nil, fmt.Errorf("java file too large for safe parsing: %s", path)
	}

	lines := strings.Split(string(content), "\n")
	functions, imports := countJavaFunctionsAndImports(lines)
	functionMetrics, structMetrics := extractJavaSymbolMetrics(path, lines)

	return &model.FileMetrics{
		Path:      path,
		Lines:     len(lines),
		Functions: functions,
		Imports:   imports,
	}, functionMetrics, structMetrics, nil
}

func countJavaFunctionsAndImports(lines []string) (int, int) {
	functions := 0
	imports := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") {
			continue
		}
		if javaImportPattern.MatchString(trimmed) {
			imports++
		}
		if javaMethodPattern.MatchString(trimmed) {
			functions++
		}
	}
	return functions, imports
}

func extractJavaSymbolMetrics(path string, lines []string) ([]model.FunctionMetrics, []model.StructMetrics) {
	functionMetrics := make([]model.FunctionMetrics, 0)
	structMetrics := make([]model.StructMetrics, 0)
	structMethods := make(map[string]int)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") {
			continue
		}
		if m := javaTypePattern.FindStringSubmatch(trimmed); len(m) == 3 {
			structMetrics = append(structMetrics, model.StructMetrics{Name: m[2], File: path, Line: i + 1, Exported: true})
			continue
		}
		if m := javaMethodPattern.FindStringSubmatch(trimmed); len(m) == 2 {
			functionMetrics = append(functionMetrics, model.FunctionMetrics{Name: m[1], File: path, Line: i + 1, Parameters: javaMethodParameterCount(trimmed)})
			if len(structMetrics) > 0 {
				owner := structMetrics[len(structMetrics)-1].Name
				structMethods[owner]++
			}
		}
	}

	for i := range structMetrics {
		structMetrics[i].Methods = structMethods[structMetrics[i].Name]
	}
	return functionMetrics, structMetrics
}

func javaMethodParameterCount(signature string) int {
	open := strings.Index(signature, "(")
	close := strings.Index(signature, ")")
	if open < 0 || close <= open {
		return 0
	}
	params := strings.TrimSpace(signature[open+1 : close])
	if params == "" {
		return 0
	}
	return strings.Count(params, ",") + 1
}

func (a *JavaAdapter) BuildDependencyGraph(files []string) (*model.DependencyGraph, error) {
	graph := model.NewDependencyGraph()
	for _, file := range files {
		pkgName, imports, err := a.extractFilePackageAndImports(file)
		if err != nil {
			continue
		}
		node := graph.AddNode(file, file, pkgName)
		for _, imp := range imports {
			normalized := a.NormalizeImport(imp)
			if normalized == "" {
				continue
			}
			if node.Metadata == nil {
				node.Metadata = make(map[string]string)
			}
			node.Metadata["import_class:"+normalized] = "internal"
			graph.AddEdge(file, normalized)
		}
	}
	return graph, nil
}

func (a *JavaAdapter) extractFilePackageAndImports(path string) (string, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	pkg := ""
	imports := make([]string, 0, 8)
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	rows := 0
	for scanner.Scan() {
		rows++
		if rows > maxJavaImportRows {
			break
		}
		line := scanner.Text()
		if strings.ContainsRune(line, '\x00') {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if pkg == "" {
			if m := javaPackagePattern.FindStringSubmatch(trimmed); len(m) == 2 {
				pkg = m[1]
			}
		}
		if m := javaImportPattern.FindStringSubmatch(trimmed); len(m) == 2 {
			imports = append(imports, m[1])
		}
	}
	if err := scanner.Err(); err != nil {
		return "", nil, err
	}
	sort.Strings(imports)
	return pkg, imports, nil
}

func (a *JavaAdapter) IsStdlibPackage(importPath string) bool {
	normalized := strings.TrimSpace(importPath)
	return strings.HasPrefix(normalized, "java.") || strings.HasPrefix(normalized, "javax.") || strings.HasPrefix(normalized, "jdk.") || strings.HasPrefix(normalized, "sun.")
}

func (a *JavaAdapter) Capabilities() AdapterCapabilities {
	return AdapterCapabilities{
		SupportsDependencyGraph: true,
		SupportsMetrics:         true,
		UsesASTParsing:          false,
	}
}

func (a *JavaAdapter) NormalizeImport(importPath string) string {
	normalized := strings.TrimSpace(importPath)
	normalized = strings.TrimSuffix(normalized, ";")
	normalized = strings.TrimPrefix(normalized, "static ")
	normalized = strings.TrimSuffix(normalized, ".*")
	return normalized
}
