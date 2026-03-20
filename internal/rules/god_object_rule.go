package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"RepoDoctor/internal/model"
)

// GodObjectRule detects structs that violate single responsibility principle
type GodObjectRule struct {
	MaxFields  int
	MaxMethods int
	fset       *token.FileSet
}

// NewGodObjectRule creates a new god object detection rule
func NewGodObjectRule() *GodObjectRule {
	return &GodObjectRule{
		MaxFields:  15,
		MaxMethods: 10,
		fset:       token.NewFileSet(),
	}
}

// ID returns the unique identifier for this rule
func (r *GodObjectRule) ID() string {
	return "rule.god-object"
}

// Category returns the category for this rule
func (r *GodObjectRule) Category() string {
	return string(CategoryMaintainability)
}

// Severity returns the severity level for this rule
func (r *GodObjectRule) Severity() string {
	return string(model.SeverityWarning)
}

func (r *GodObjectRule) Capabilities() RuleCapabilities {
	return RuleCapabilities{SupportedLanguages: []string{"Go"}, SupportsMultipleLanguages: false}
}

// Evaluate executes the rule logic against the provided context
func (r *GodObjectRule) Evaluate(context AnalysisContext) []model.Violation {
	var violations []model.Violation

	// Map to track methods per struct (package-qualified key -> info)
	// Keys are Dir(filePath)+"#"+structName to prevent cross-package
	// name collisions (e.g. main.DependencyGraph vs model.DependencyGraph).
	structMethods := make(map[string]*structInfo)

	// First pass: collect all struct definitions and their fields
	for _, file := range context.RepositoryFiles {
		r.collectStructs(file, structMethods)
	}

	// Second pass: collect all method declarations
	for _, file := range context.RepositoryFiles {
		r.collectMethods(file, structMethods)
	}

	// Check for violations
	for _, info := range structMethods {
		fieldCount := info.FieldCount
		methodCount := info.MethodCount

		// Check field threshold
		if fieldCount > r.MaxFields {
			violations = append(violations, model.Violation{
				RuleID:      r.ID(),
				Severity:    model.SeverityWarning,
				Message:     info.Name + " has " + strconv.Itoa(fieldCount) + " fields (threshold: " + strconv.Itoa(r.MaxFields) + ")",
				File:        info.File,
				Line:        0,
				ScoreImpact: -5.0,
			})
		}

		// Check method threshold
		if methodCount > r.MaxMethods {
			violations = append(violations, model.Violation{
				RuleID:      r.ID(),
				Severity:    model.SeverityWarning,
				Message:     info.Name + " has " + strconv.Itoa(methodCount) + " methods (threshold: " + strconv.Itoa(r.MaxMethods) + ")",
				File:        info.File,
				Line:        0,
				ScoreImpact: -5.0,
			})
		}
	}

	return violations
}

// structInfo holds information about a struct
type structInfo struct {
	Name        string // bare struct name for display
	File        string
	FieldCount  int
	MethodCount int
}

// structKey returns a package-qualified key for a struct to avoid
// cross-package name collisions. Go requires methods to reside in the
// same package (directory) as their receiver type, so Dir+Name is unique.
func structKey(filePath, structName string) string {
	return filepath.Dir(filePath) + "#" + structName
}

// collectStructs collects all struct definitions and their field counts
func (r *GodObjectRule) collectStructs(file RepositoryFile, structMethods map[string]*structInfo) {
	node, err := parser.ParseFile(r.fset, file.Path, file.Content, 0)
	if err != nil {
		return // Skip malformed files
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
		key := structKey(file.Path, structName)
		structMethods[key] = &structInfo{
			Name:        structName,
			File:        file.Path,
			FieldCount:  fieldCount,
			MethodCount: 0,
		}

		return true
	})
}

// collectMethods collects all method declarations for each struct
func (r *GodObjectRule) collectMethods(file RepositoryFile, structMethods map[string]*structInfo) {
	node, err := parser.ParseFile(r.fset, file.Path, file.Content, 0)
	if err != nil {
		return // Skip malformed files
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

			// Get the type name and look up with package-qualified key
			if ident, ok := recvType.(*ast.Ident); ok {
				key := structKey(file.Path, ident.Name)
				if info, exists := structMethods[key]; exists {
					info.MethodCount++
				}
			}
		}

		return true
	})
}

// LoadFromDir loads Go files from a directory and returns them as RepositoryFile slices
func LoadFromDir(root string) ([]RepositoryFile, error) {
	var files []RepositoryFile

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		files = append(files, RepositoryFile{
			Path:    path,
			Content: string(content),
		})

		return nil
	})

	return files, err
}
