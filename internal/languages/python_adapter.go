package languages

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"RepoDoctor/internal/model"
)

// PythonAdapter implements LanguageAdapter for Python programming language
type PythonAdapter struct {
	importPatterns []*regexp.Regexp
}

const maxPythonFileBytes = 2 * 1024 * 1024
const maxPythonEvidenceFiles = 5000

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

// CollectEvidence extracts normalized import evidence for language detection.
func (a *PythonAdapter) CollectEvidence(repoPath string, files []string) ([]EvidenceSignal, []string, error) {
	root, err := normalizeRepoRoot(repoPath)
	if err != nil {
		return nil, nil, err
	}

	sortedFiles := append([]string(nil), files...)
	sort.Strings(sortedFiles)

	if len(sortedFiles) > maxPythonEvidenceFiles {
		sortedFiles = sortedFiles[:maxPythonEvidenceFiles]
	}

	signals := make([]EvidenceSignal, 0, len(sortedFiles)*2)
	warnings := make([]string, 0)

	for _, file := range sortedFiles {
		normalizedPath, ok := normalizePathWithinRoot(root, file)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("skipped path outside repo root: %s", file))
			continue
		}

		info, statErr := os.Lstat(normalizedPath)
		if statErr != nil {
			warnings = append(warnings, fmt.Sprintf("failed to stat python file: %s", normalizedPath))
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			warnings = append(warnings, fmt.Sprintf("skipped symlink python file: %s", normalizedPath))
			continue
		}
		if info.Size() > maxPythonFileBytes {
			warnings = append(warnings, fmt.Sprintf("skipped oversized python file: %s", normalizedPath))
			continue
		}

		evidence, parseErr := parsePythonImportEvidence(normalizedPath)
		if parseErr != nil {
			warnings = append(warnings, fmt.Sprintf("failed to parse python imports: %s", normalizedPath))
			continue
		}

		moduleRoot := detectPythonModuleRoot(root, normalizedPath)
		for _, item := range evidence {
			weight := 0.40
			signalType := "python_import_absolute"
			if item.relative {
				signalType = "python_import_relative"
				weight = 2.10
			}

			normalizedImport := normalizePythonImport(item, normalizedPath, root, moduleRoot)
			if normalizedImport == "" {
				continue
			}

			signals = append(signals, EvidenceSignal{
				Language:    "Python",
				SignalType:  signalType,
				WeightInput: weight,
				SourcePath:  normalizedPath,
			})
		}
	}

	return signals, warnings, nil
}
