package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// GodObjectViolation represents a god object detection violation
type GodObjectViolation struct {
	StructName  string
	File        string
	FieldCount  int
	MethodCount int
}

// GodObjectRule detects structs that violate single responsibility principle
type GodObjectRule struct {
	MaxFields  int
	MaxMethods int
	violations []GodObjectViolation
	fset       *token.FileSet
}

// NewGodObjectRule creates a new god object detection rule
func NewGodObjectRule() *GodObjectRule {
	return &GodObjectRule{
		MaxFields:  15,
		MaxMethods: 10,
		violations: make([]GodObjectViolation, 0),
		fset:       token.NewFileSet(),
	}
}

// Check analyzes the given directory for god object violations
func (r *GodObjectRule) Check(dirPath string) error {
	r.violations = make([]GodObjectViolation, 0)

	// Map to track methods per struct (struct name -> method count)
	structMethods := make(map[string]*structInfo)

	// First pass: collect all struct definitions and their fields
	err := r.walkDir(dirPath, func(filePath string) error {
		return r.collectStructs(filePath, structMethods)
	})
	if err != nil {
		return err
	}

	// Second pass: collect all method declarations
	err = r.walkDir(dirPath, func(filePath string) error {
		return r.collectMethods(filePath, structMethods)
	})
	if err != nil {
		return err
	}

	// Check for violations
	for structName, info := range structMethods {
		isViolation := false
		fieldCount := info.FieldCount
		methodCount := info.MethodCount

		// Check field threshold
		if fieldCount > r.MaxFields {
			isViolation = true
		}

		// Check method threshold
		if methodCount > r.MaxMethods {
			isViolation = true
		}

		if isViolation {
			r.violations = append(r.violations, GodObjectViolation{
				StructName:  structName,
				File:        info.File,
				FieldCount:  fieldCount,
				MethodCount: methodCount,
			})
		}
	}

	return nil
}

// structInfo holds information about a struct
type structInfo struct {
	File        string
	FieldCount  int
	MethodCount int
}

// Violations returns all detected god object violations
func (r *GodObjectRule) Violations() []GodObjectViolation {
	return r.violations
}

// walkDir walks through a directory and calls the callback for each Go file
func (r *GodObjectRule) walkDir(root string, callback func(string) error) error {
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

// collectStructs collects all struct definitions and their field counts
func (r *GodObjectRule) collectStructs(filePath string, structMethods map[string]*structInfo) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	node, err := parser.ParseFile(r.fset, filePath, content, 0)
	if err != nil {
		return nil // Skip malformed files
	}

	// Walk through all declarations
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		// Count fields
		fieldCount := 0
		if structType.Fields != nil {
			fieldCount = structType.Fields.NumFields()
		}

		structName := typeSpec.Name.Name
		structMethods[structName] = &structInfo{
			File:        filePath,
			FieldCount:  fieldCount,
			MethodCount: 0,
		}

		return true
	})

	return nil
}

// collectMethods collects all method declarations for each struct
func (r *GodObjectRule) collectMethods(filePath string, structMethods map[string]*structInfo) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	node, err := parser.ParseFile(r.fset, filePath, content, 0)
	if err != nil {
		return nil // Skip malformed files
	}

	// Walk through all declarations
	ast.Inspect(node, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		// Check if this is a method (has receiver)
		if funcDecl.Recv == nil {
			return true
		}

		// Get receiver type
		for _, field := range funcDecl.Recv.List {
			recvType := field.Type

			// Handle pointer receivers (*T)
			if starExpr, ok := recvType.(*ast.StarExpr); ok {
				recvType = starExpr.X
			}

			// Get the type name
			if ident, ok := recvType.(*ast.Ident); ok {
				structName := ident.Name
				if info, exists := structMethods[structName]; exists {
					info.MethodCount++
				}
			}
		}

		return true
	})

	return nil
}

// HasCriticalViolations returns true if any god object violations found
func (r *GodObjectRule) HasCriticalViolations() bool {
	return len(r.violations) > 0
}
