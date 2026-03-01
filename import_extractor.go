package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ImportMetadata holds package-level import information
type ImportMetadata struct {
	Package string
	Imports []string
}

// ImportExtractor extracts import metadata from Go source files
type ImportExtractor struct {
	modulePath    string
	stdlibPrefixs map[string]bool
}

// NewImportExtractor creates a new ImportExtractor
func NewImportExtractor(modulePath string) *ImportExtractor {
	return &ImportExtractor{
		modulePath:    modulePath,
		stdlibPrefixs: buildStdlibPrefixs(),
	}
}

// ExtractFromDir extracts import metadata from all .go files in a directory
func (e *ImportExtractor) ExtractFromDir(rootPath string) (map[string]*ImportMetadata, error) {
	result := make(map[string]*ImportMetadata)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			// Skip vendor, node_modules, and docs
			if info.Name() == "vendor" || info.Name() == "node_modules" || info.Name() == "docs" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// Parse the Go file
		metadata, err := e.ExtractFromFile(path)
		if err != nil {
			// Gracefully handle invalid files - skip them
			return nil
		}

		if metadata != nil {
			result[path] = metadata
		}

		return nil
	})

	return result, err
}

// ExtractFromFile extracts import metadata from a single Go file
func (e *ImportExtractor) ExtractFromFile(filePath string) (*ImportMetadata, error) {
	fset := token.NewFileSet()

	// Parse the file with parser.ParseFile which handles errors gracefully
	file, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		// Gracefully handle malformed files
		return nil, nil
	}

	// Get the package name
	packageName := file.Name.Name

	// Extract imports
	imports := e.extractImports(file)

	return &ImportMetadata{
		Package: packageName,
		Imports: imports,
	}, nil
}

// extractImports extracts and normalizes import paths from an AST file
func (e *ImportExtractor) extractImports(file *ast.File) []string {
	importMap := make(map[string]bool)

	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		// Skip standard library imports
		if e.isStdlibImport(importPath) {
			continue
		}

		// Normalize the import path (remove aliases, etc.)
		normalized := e.normalizeImport(importPath)

		if normalized != "" && !importMap[normalized] {
			importMap[normalized] = true
		}
	}

	// Convert map to slice
	imports := make([]string, 0, len(importMap))
	for imp := range importMap {
		imports = append(imports, imp)
	}

	return imports
}

// isStdlibImport checks if an import path is from the standard library
func (e *ImportExtractor) isStdlibImport(importPath string) bool {
	// Standard library imports don't contain a dot in the first path component
	// and are not the current module
	parts := strings.Split(importPath, "/")
	if len(parts) == 0 {
		return true
	}

	firstPart := parts[0]

	// Standard library packages typically don't have dots or hyphens in their names
	// and are not relative paths
	if !strings.Contains(firstPart, ".") && !strings.Contains(firstPart, "-") {
		// Check against known stdlib prefixes
		return e.stdlibPrefixs[firstPart]
	}

	return false
}

// normalizeImport normalizes an import path relative to the module
func (e *ImportExtractor) normalizeImport(importPath string) string {
	// Remove module prefix if it's an internal import
	if strings.HasPrefix(importPath, e.modulePath+"/") {
		relative := strings.TrimPrefix(importPath, e.modulePath+"/")
		return "./" + relative
	} else if importPath == e.modulePath {
		return "./"
	}

	// Return external imports as-is
	return importPath
}

// buildStdlibPrefixs builds a set of common stdlib package prefixes
func buildStdlibPrefixs() map[string]bool {
	return map[string]bool{
		// A
		"archive": true,
		"bufio":   true,
		"bytes":   true,
		// C
		"compress":  true,
		"container": true,
		"context":   true,
		"crypto":    true,
		// D
		"database": true,
		"debug":    true,
		// E
		"encoding": true,
		"errors":   true,
		// F
		"flag": true,
		"fmt":  true,
		// G
		"go": true,
		// H
		"hash": true,
		"html": true,
		// I
		"image":    true,
		"index":    true,
		"internal": true,
		"io":       true,
		// L
		"log": true,
		// M
		"math": true,
		"mime": true,
		// N
		"net": true,
		// O
		"os": true,
		// P
		"path":   true,
		"plugin": true,
		"pprof":  true,
		// R
		"reflect": true,
		"regexp":  true,
		"runtime": true,
		// S
		"sort":    true,
		"strconv": true,
		"strings": true,
		"sync":    true,
		"syscall": true,
		// T
		"testing": true,
		"text":    true,
		"time":    true,
		// U
		"unicode": true,
		// V
		"unsafe": true,
	}
}
