package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// SizeViolation represents a violation of size thresholds
type SizeViolation struct {
	File      string
	Function  string
	Lines     int
	Threshold int
}

// SizeRule checks file and function size thresholds
type SizeRule struct {
	MaxFileLines     int
	MaxFunctionLines int
	violations       []SizeViolation
	fset             *token.FileSet
}

// NewSizeRule creates a new size rule checker with default thresholds
func NewSizeRule() *SizeRule {
	return &SizeRule{
		MaxFileLines:     500,
		MaxFunctionLines: 80,
		violations:       make([]SizeViolation, 0),
		fset:             token.NewFileSet(),
	}
}

// Check analyzes the given directory for size violations
func (s *SizeRule) Check(dirPath string) error {
	s.violations = make([]SizeViolation, 0)

	err := s.checkFilesInDir(dirPath)
	if err != nil {
		return err
	}

	return nil
}

// Violations returns all detected size violations
func (s *SizeRule) Violations() []SizeViolation {
	return s.violations
}

// checkFilesInDir walks through the directory and checks all Go files
func (s *SizeRule) checkFilesInDir(dirPath string) error {
	return s.walkDir(dirPath, func(filePath string) error {
		return s.checkFile(filePath)
	})
}

// walkDir walks through a directory and calls the callback for each Go file
func (s *SizeRule) walkDir(root string, callback func(string) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip non-Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		return callback(path)
	})
}

// checkFile checks a single file for size violations
func (s *SizeRule) checkFile(filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Check file LOC
	fileLines := s.countNonEmptyLines(string(content))
	if fileLines > s.MaxFileLines {
		s.violations = append(s.violations, SizeViolation{
			File:      filePath,
			Function:  "",
			Lines:     fileLines,
			Threshold: s.MaxFileLines,
		})
	}

	// Check function LOC
	s.checkFunctions(filePath, content)

	return nil
}

// countNonEmptyLines counts non-empty lines in a file
func (s *SizeRule) countNonEmptyLines(content string) int {
	lines := strings.Split(content, "\n")
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			count++
		}
	}
	return count
}

// checkFunctions checks function sizes in a file
func (s *SizeRule) checkFunctions(filePath string, content []byte) {
	// Parse AST
	node, err := parser.ParseFile(s.fset, filePath, content, 0)
	if err != nil {
		return // Skip malformed files
	}

	// Walk through all declarations
	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Calculate function lines
		startLine := s.fset.Position(funcDecl.Pos()).Line
		endLine := s.fset.Position(funcDecl.End()).Line
		funcLines := endLine - startLine + 1

		if funcLines > s.MaxFunctionLines {
			s.violations = append(s.violations, SizeViolation{
				File:      filePath,
				Function:  funcDecl.Name.Name,
				Lines:     funcLines,
				Threshold: s.MaxFunctionLines,
			})
		}

		return true
	})
}

// HasCriticalViolations returns true if any size violations found
func (s *SizeRule) HasCriticalViolations() bool {
	return len(s.violations) > 0
}
