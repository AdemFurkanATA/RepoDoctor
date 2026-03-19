package languages

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"RepoDoctor/internal/model"
)

// PythonAdapter implements LanguageAdapter for Python programming language
type PythonAdapter struct {
	importPatterns []*regexp.Regexp
}

const maxPythonFileBytes = 2 * 1024 * 1024

// pythonImportExtractor helps extract imports from Python files
type pythonImportExtractor struct {
	patterns []*regexp.Regexp
}

// newPythonImportExtractor creates a new import extractor
func newPythonImportExtractor() *pythonImportExtractor {
	return &pythonImportExtractor{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^\s*import\s+([\w.]+)`),
			regexp.MustCompile(`^\s*from\s+([\w.]+)\s+import`),
		},
	}
}

// extractImportsFromFile extracts imports from a single Python file
func (e *pythonImportExtractor) extractImportsFromFile(path string) ([]string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	var imports []string
	pkgName := ""
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		for _, pattern := range e.patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				importPath := matches[1]
				parts := strings.Split(importPath, ".")
				if len(parts) > 0 {
					imports = append(imports, parts[0])
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, "", err
	}

	if pkgName == "" {
		base := filepath.Base(path)
		pkgName = strings.TrimSuffix(base, ".py")
	}

	return imports, pkgName, nil
}

// NewPythonAdapter creates a new Python language adapter
func NewPythonAdapter() *PythonAdapter {
	return &PythonAdapter{
		importPatterns: []*regexp.Regexp{
			regexp.MustCompile(`^\s*import\s+([\w.]+)`),
			regexp.MustCompile(`^\s*from\s+([\w.]+)\s+import`),
		},
	}
}

// Name returns the language name
func (a *PythonAdapter) Name() string {
	return "Python"
}

// FileExtensions returns supported file extensions
func (a *PythonAdapter) FileExtensions() []string {
	return []string{".py"}
}

// DetectFiles scans the repository and returns all Python files
func (a *PythonAdapter) DetectFiles(repoPath string) ([]string, error) {
	var pythonFiles []string

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and __pycache__
		if strings.HasPrefix(info.Name(), ".") || info.Name() == "__pycache__" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.py") || strings.HasPrefix(info.Name(), "test_") {
			return nil
		}

		// Check if it's a Python file
		if strings.HasSuffix(path, ".py") {
			pythonFiles = append(pythonFiles, path)
		}

		return nil
	})

	return pythonFiles, err
}

// CollectMetrics extracts Python-specific metrics from source files
func (a *PythonAdapter) CollectMetrics(files []string) (*model.RepositoryMetrics, error) {
	metrics := model.NewRepositoryMetrics()

	for _, file := range files {
		fileMetrics, err := a.collectFileMetrics(file)
		if err != nil {
			continue
		}
		metrics.AddFileMetrics(*fileMetrics)
	}

	return metrics, nil
}

// collectFileMetrics extracts metrics from a single Python file
func (a *PythonAdapter) collectFileMetrics(path string) (*model.FileMetrics, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(content) > maxPythonFileBytes {
		return nil, fmt.Errorf("python file too large for safe parsing: %s", path)
	}

	lines := strings.Split(string(content), "\n")
	fm := &model.FileMetrics{
		Path:      path,
		Lines:     len(lines),
		Functions: 0,
		Imports:   0,
	}

	localMetrics := model.NewRepositoryMetrics()

	// Count imports, functions, classes
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Count imports
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") {
			fm.Imports++
		}

		// Count function definitions
		if strings.HasPrefix(trimmed, "def ") {
			fm.Functions++
			funcMetrics := pyExtractFunctionMetrics(trimmed, path, i+1)
			localMetrics.AddFunctionMetrics(*funcMetrics)
		}

		// Count class definitions
		if strings.HasPrefix(trimmed, "class ") {
			structMetrics := pyExtractClassMetrics(trimmed, path, i+1)
			localMetrics.AddStructMetrics(*structMetrics)
		}
	}

	return fm, nil
}

// pyExtractFunctionMetrics extracts metrics from a Python function definition line.
// Package-level helper to keep PythonAdapter method count within SRP bounds.
func pyExtractFunctionMetrics(line, path string, lineNum int) *model.FunctionMetrics {
	// Parse: def function_name(params):
	name := "unknown"
	params := 0

	if strings.HasPrefix(line, "def ") {
		// Extract function name
		start := 4 // len("def ")
		end := strings.Index(line, "(")
		if end > start {
			name = line[start:end]

			// Count parameters
			paramStr := line[end+1:]
			closeParen := strings.Index(paramStr, ")")
			if closeParen > 0 {
				paramStr = paramStr[:closeParen]
				if paramStr != "" && paramStr != "self" && paramStr != "cls" {
					// Simple parameter counting
					params = strings.Count(paramStr, ",") + 1
					if strings.TrimSpace(paramStr) == "self" || strings.TrimSpace(paramStr) == "cls" {
						params = 0
					}
				}
			}
		}
	}

	return &model.FunctionMetrics{
		Name:       name,
		File:       path,
		Line:       lineNum,
		Parameters: params,
		Lines:      0, // Would need more sophisticated parsing
	}
}

// pyExtractClassMetrics extracts metrics from a Python class definition line.
// Package-level helper to keep PythonAdapter method count within SRP bounds.
func pyExtractClassMetrics(line, path string, lineNum int) *model.StructMetrics {
	// Parse: class ClassName:
	name := "Unknown"

	if strings.HasPrefix(line, "class ") {
		start := 6 // len("class ")
		end := strings.Index(line, ":")
		if end > start {
			namePart := strings.TrimSpace(line[start:end])
			// Remove base classes if present
			if parenIdx := strings.Index(namePart, "("); parenIdx > 0 {
				name = namePart[:parenIdx]
			} else {
				name = namePart
			}
		}
	}

	return &model.StructMetrics{
		Name:     name,
		File:     path,
		Line:     lineNum,
		Fields:   0, // Would need __init__ parsing
		Methods:  0, // Would need class body parsing
		Exported: len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z',
	}
}

// BuildDependencyGraph constructs a dependency graph from Python imports
func (a *PythonAdapter) BuildDependencyGraph(files []string) (*model.DependencyGraph, error) {
	graph := model.NewDependencyGraph()

	for _, file := range files {
		imports, pkgName, err := a.extractImports(file)
		if err != nil {
			continue
		}

		nodeID := file
		graphNode := graph.AddNode(nodeID, file, pkgName)

		// Add edges for imports
		for _, imp := range imports {
			graphNode.Imports = append(graphNode.Imports, imp)
			graph.AddEdge(nodeID, imp)
		}
	}

	return graph, nil
}

// extractImports extracts import statements from a Python file
func (a *PythonAdapter) extractImports(path string) ([]string, string, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}

	if fileInfo.Size() > maxPythonFileBytes {
		return nil, "", fmt.Errorf("python file too large for safe parsing: %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	var imports []string
	pkgName := ""
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Try to match import patterns
		for _, pattern := range a.importPatterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				importPath := matches[1]

				// Get base module
				parts := strings.Split(importPath, ".")
				if len(parts) > 0 {
					imports = append(imports, parts[0])
				}
			}
		}

	}

	if err := scanner.Err(); err != nil {
		return nil, "", err
	}

	// Default package name from filename
	if pkgName == "" {
		base := filepath.Base(path)
		pkgName = strings.TrimSuffix(base, ".py")
	}

	return imports, pkgName, nil
}

// IsStdlibPackage checks if a module is part of Python standard library
func (a *PythonAdapter) IsStdlibPackage(importPath string) bool {
	baseModule := strings.Split(importPath, ".")[0]

	// Common Python standard library modules
	stdlibModules := map[string]bool{
		"os": true, "sys": true, "re": true, "json": true,
		"collections": true, "itertools": true, "functools": true,
		"pathlib": true, "typing": true, "abc": true,
		"io": true, "time": true, "datetime": true,
		"math": true, "random": true, "hashlib": true,
		"string": true, "textwrap": true, "unicodedata": true,
		"struct": true, "codecs": true, "base64": true,
		"http": true, "urllib": true, "socket": true,
		"asyncio": true, "threading": true, "multiprocessing": true,
		"logging": true, "unittest": true, "pytest": false,
		"csv": true, "configparser": true, "argparse": true,
		"xml": true, "email": true, "subprocess": true,
		"tempfile": true, "inspect": true,
	}

	return stdlibModules[baseModule]
}

// Capabilities returns Python adapter capabilities.
func (a *PythonAdapter) Capabilities() AdapterCapabilities {
	return AdapterCapabilities{
		SupportsDependencyGraph: true,
		SupportsMetrics:         true,
		UsesASTParsing:          false,
	}
}

// NormalizeImport normalizes Python import module names.
func (a *PythonAdapter) NormalizeImport(importPath string) string {
	trimmed := strings.TrimSpace(importPath)
	if trimmed == "" {
		return ""
	}
	return strings.Split(trimmed, ".")[0]
}

// DetectPythonVersion attempts to detect Python version from the repository.
// Package-level helper to keep PythonAdapter method count within SRP bounds.
func DetectPythonVersion(repoPath string) (string, error) {
	// Check for .python-version file
	versionFile := filepath.Join(repoPath, ".python-version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	// Check for pyproject.toml
	pyprojectFile := filepath.Join(repoPath, "pyproject.toml")
	if content, err := os.ReadFile(pyprojectFile); err == nil {
		// Look for requires-python
		// This is a simplified check
		contentStr := string(content)
		if strings.Contains(contentStr, "requires-python") {
			return "from pyproject.toml", nil
		}
	}

	return "", fmt.Errorf("Python version not detected")
}
