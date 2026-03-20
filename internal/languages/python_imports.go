package languages

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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
