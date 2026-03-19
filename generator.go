package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type ruleTemplateData struct {
	RuleName string
	TypeName string
	RuleID   string
}

// RuleTemplateGenerator generates rule templates
type RuleTemplateGenerator struct {
	rulesDir string
}

// NewRuleTemplateGenerator creates a new generator
func NewRuleTemplateGenerator(rulesDir string) *RuleTemplateGenerator {
	return &RuleTemplateGenerator{
		rulesDir: rulesDir,
	}
}

// Generate creates a new rule template file
func (g *RuleTemplateGenerator) Generate(ruleName string) error {
	sanitized, err := sanitizeRuleName(ruleName)
	if err != nil {
		return err
	}

	return g.generateTemplateFiles(sanitized)
}

func (g *RuleTemplateGenerator) generateTemplateFiles(sanitized string) error {

	// Create rules directory if it doesn't exist
	if err := os.MkdirAll(g.rulesDir, 0755); err != nil {
		return fmt.Errorf("failed to create rules directory: %w", err)
	}

	// Generate file name
	fileName := strings.ReplaceAll(sanitized, "-", "_") + "_rule.go"
	filePath := filepath.Join(g.rulesDir, fileName)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("rule file already exists: %s", filePath)
	}

	// Generate template content
	content := g.generateTemplate(sanitized)

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write rule file: %w", err)
	}

	fmt.Printf("✅ Rule template created: %s\n", filePath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("1. Implement the Evaluate method\n")
	fmt.Printf("2. Add the rule to the rule registry\n")
	fmt.Printf("3. Write tests for the new rule\n\n")

	return nil
}

// generateTemplate creates the rule template content
func (g *RuleTemplateGenerator) generateTemplate(ruleName string) string {
	typeName := ruleTypeName(ruleName)
	data := g.buildRuleTemplateData(ruleName, typeName)

	content, err := g.renderRuleTemplate(data)
	if err != nil {
		return g.generateSimpleTemplate(ruleName, typeName)
	}

	return content
}

func (g *RuleTemplateGenerator) buildRuleTemplateData(ruleName, typeName string) ruleTemplateData {
	return ruleTemplateData{
		RuleName: ruleName,
		TypeName: typeName,
		RuleID:   strings.ReplaceAll(ruleName, "-", "_"),
	}
}

func (g *RuleTemplateGenerator) renderRuleTemplate(data ruleTemplateData) (string, error) {
	t, err := template.New("rule").Parse(g.ruleTemplateText())
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	if err := t.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}

func (g *RuleTemplateGenerator) ruleTemplateText() string {
	return `package rules

import (
	"RepoDoctor/internal/model"
)

// {{.TypeName}}Rule detects {{.RuleName}} violations.
type {{.TypeName}}Rule struct{}

// New{{.TypeName}}Rule creates a new {{.TypeName}}Rule instance.
func New{{.TypeName}}Rule() *{{.TypeName}}Rule {
	return &{{.TypeName}}Rule{}
}

// ID returns the unique identifier for this rule.
func (r *{{.TypeName}}Rule) ID() string {
	return "rule.{{.RuleID}}"
}

// Category returns the rule category.
func (r *{{.TypeName}}Rule) Category() string {
	return "maintainability"
}

// Severity returns the default severity.
func (r *{{.TypeName}}Rule) Severity() string {
	return "warning"
}

// Evaluate checks for {{.RuleName}} violations in the analysis context.
// This signature matches internal/rules.Rule interface.
func (r *{{.TypeName}}Rule) Evaluate(context AnalysisContext) []model.Violation {
	var violations []model.Violation
	// TODO: Implement rule evaluation logic
	return violations
}

// AnalysisContext mirrors internal/rules.AnalysisContext.
// NOTE: When moving this rule into internal/rules/, remove this type
// and use the AnalysisContext defined in that package instead.
type AnalysisContext struct {
	RepositoryFiles []RepositoryFile
}

// RepositoryFile represents a source file for analysis.
type RepositoryFile struct {
	Path    string
	Content string
	Imports []string
}
`
}

const simpleRuleTemplate = `package rules

import (
	"RepoDoctor/internal/model"
)

// %sRule detects %s violations.
type %sRule struct{}

// New%sRule creates a new %sRule instance.
func New%sRule() *%sRule {
	return &%sRule{}
}

// ID returns the unique identifier for this rule.
func (r *%sRule) ID() string {
	return "rule.%s"
}

// Category returns the rule category.
func (r *%sRule) Category() string {
	return "maintainability"
}

// Severity returns the default severity.
func (r *%sRule) Severity() string {
	return "warning"
}

// Evaluate checks for %s violations in the analysis context.
func (r *%sRule) Evaluate(context AnalysisContext) []model.Violation {
	var violations []model.Violation
	// TODO: Implement rule evaluation logic
	return violations
}

// AnalysisContext mirrors internal/rules.AnalysisContext.
type AnalysisContext struct {
	RepositoryFiles []RepositoryFile
}

// RepositoryFile represents a source file for analysis.
type RepositoryFile struct {
	Path    string
	Content string
	Imports []string
}
`

// generateSimpleTemplate creates a simpler template if template rendering fails
func (g *RuleTemplateGenerator) generateSimpleTemplate(ruleName, typeName string) string {
	ruleID := strings.ReplaceAll(ruleName, "-", "_")

	return fmt.Sprintf(
		simpleRuleTemplate,
		typeName, ruleName, typeName,
		typeName, typeName, typeName, typeName, typeName,
		typeName, ruleID,
		typeName,
		typeName,
		ruleName,
		typeName,
	)
}

// GenerateWithTest generates both rule template and test file
func (g *RuleTemplateGenerator) GenerateWithTest(ruleName string) error {
	sanitized, err := sanitizeRuleName(ruleName)
	if err != nil {
		return err
	}

	if err := g.generateTemplateFiles(sanitized); err != nil {
		return err
	}

	// Generate test file
	testFileName := strings.ReplaceAll(sanitized, "-", "_") + "_rule_test.go"
	testFilePath := filepath.Join(g.rulesDir, testFileName)

	testContent := g.generateTestTemplate(sanitized)

	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		return fmt.Errorf("failed to write test file: %w", err)
	}

	fmt.Printf("✅ Test template created: %s\n", testFilePath)

	return nil
}

// generateTestTemplate creates a test template
func (g *RuleTemplateGenerator) generateTestTemplate(ruleName string) string {
	typeName := ruleTypeName(ruleName)

	return fmt.Sprintf(`package rules

import (
	"testing"
)

func Test%sRule_ID(t *testing.T) {
	rule := New%sRule()
	
	expected := "rule.%s"
	if rule.ID() != expected {
		t.Errorf("Expected ID %%s, got %%s", expected, rule.ID())
	}
}

func Test%sRule_Category(t *testing.T) {
	rule := New%sRule()
	
	if rule.Category() == "" {
		t.Error("Expected non-empty category")
	}
}

func Test%sRule_Severity(t *testing.T) {
	rule := New%sRule()
	
	if rule.Severity() == "" {
		t.Error("Expected non-empty severity")
	}
}

func Test%sRule_Evaluate(t *testing.T) {
	rule := New%sRule()
	
	violations := rule.Evaluate(AnalysisContext{})
	if violations == nil {
		t.Fatal("expected non-nil violations slice")
	}
}
`, typeName, typeName, strings.ReplaceAll(ruleName, "-", "_"),
		typeName, typeName,
		typeName, typeName,
		typeName, typeName)
}

func sanitizeRuleName(ruleName string) (string, error) {
	trimmed := strings.TrimSpace(strings.ToLower(ruleName))
	if trimmed == "" {
		return "", fmt.Errorf("rule name cannot be empty")
	}

	normalized := strings.ReplaceAll(trimmed, " ", "-")
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.Trim(normalized, "-")

	validName := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !validName.MatchString(normalized) {
		return "", fmt.Errorf("invalid rule name %q: use only letters, numbers, spaces, underscore or hyphen", ruleName)
	}

	return normalized, nil
}

func ruleTypeName(ruleName string) string {
	parts := strings.FieldsFunc(ruleName, func(r rune) bool {
		return r == '-' || r == '_'
	})

	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}

	if builder.Len() == 0 {
		return "Custom"
	}

	return builder.String()
}
