package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

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
	// Create rules directory if it doesn't exist
	if err := os.MkdirAll(g.rulesDir, 0755); err != nil {
		return fmt.Errorf("failed to create rules directory: %w", err)
	}

	// Generate file name
	fileName := strings.ToLower(ruleName) + "_rule.go"
	filePath := filepath.Join(g.rulesDir, fileName)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("rule file already exists: %s", filePath)
	}

	// Generate template content
	content := g.generateTemplate(ruleName)

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
	typeName := strings.Title(strings.ReplaceAll(ruleName, "-", "_"))
	typeName = strings.ReplaceAll(typeName, "_", "")

	tmpl := `package rules

import "fmt"

// {{.TypeName}}Rule detects {{.RuleName}} violations.
type {{.TypeName}}Rule struct {
	enabled bool
}

// New{{.TypeName}}Rule creates a new {{.TypeName}}Rule instance.
func New{{.TypeName}}Rule() *{{.TypeName}}Rule {
	return &{{.TypeName}}Rule{
		enabled: true,
	}
}

// ID returns the unique identifier for this rule.
func (r *{{.TypeName}}Rule) ID() string {
	return "{{.RuleID}}"
}

// Name returns the human-readable name of the rule.
func (r *{{.TypeName}}Rule) Name() string {
	return "{{.RuleName}}"
}

// Enabled indicates whether this rule is enabled by default.
func (r *{{.TypeName}}Rule) Enabled() bool {
	return r.enabled
}

// Evaluate checks for {{.RuleName}} violations in the repository
func (r *{{.TypeName}}Rule) Evaluate(graph internal.DependencyGraph, rootPath string) ([]Violation, error) {
	var violations []Violation
	// TODO: Implement rule evaluation logic
	return violations, nil
}

// Violation represents a rule violation
type Violation struct {
	RuleID   string
	File     string
	Message  string
	Severity string
	Line     int
}
`

	t := template.Must(template.New("rule").Parse(tmpl))
	
	data := struct {
		RuleName string
		TypeName string
		RuleID   string
	}{
		RuleName: ruleName,
		TypeName: typeName,
		RuleID:   strings.ReplaceAll(ruleName, "-", "_"),
	}

	var builder strings.Builder
	if err := t.Execute(&builder, data); err != nil {
		// Fallback to simple template if template fails
		return g.generateSimpleTemplate(ruleName, typeName)
	}

	return builder.String()
}

// generateSimpleTemplate creates a simpler template if template rendering fails
func (g *RuleTemplateGenerator) generateSimpleTemplate(ruleName, typeName string) string {
	ruleID := strings.ReplaceAll(ruleName, "-", "_")
	
	return fmt.Sprintf(`package rules

import (
	"RepoDoctor/internal"
)

// %sRule detects %s violations
type %sRule struct {
	enabled bool
}

// New%sRule creates a new %sRule instance
func New%sRule() *%sRule {
	return &%sRule{
		enabled: true,
	}
}

// ID returns the unique identifier for this rule
func (r *%sRule) ID() string {
	return "%s"
}

// Name returns the human-readable name of the rule
func (r *%sRule) Name() string {
	return "%s Rule"
}

// Description returns a detailed description of what the rule checks
func (r *%sRule) Description() string {
	return "TODO: Add description for %s rule"
}

// Category returns the category of this rule
func (r *%sRule) Category() string {
	return "TODO"
}

// Severity returns the severity level of violations
func (r *%sRule) Severity() string {
	return "warning"
}

// Enabled returns whether the rule is enabled
func (r *%sRule) Enabled() bool {
	return r.enabled
}

// SetEnabled enables or disables the rule
func (r *%sRule) SetEnabled(enabled bool) {
	r.enabled = enabled
}

// Configure sets rule-specific configuration
func (r *%sRule) Configure(config map[string]interface{}) error {
	return nil
}

// Evaluate checks for %s violations in the repository
func (r *%sRule) Evaluate(graph internal.DependencyGraph, rootPath string) ([]Violation, error) {
	var violations []Violation

	// TODO: Implement rule evaluation logic

	return violations, nil
}

// Violation represents a rule violation
type Violation struct {
	RuleID   string
	File     string
	Message  string
	Severity string
	Line     int
}
`, typeName, ruleName, typeName, typeName, typeName, typeName, typeName, typeName, 
	typeName, ruleID, typeName, typeName, typeName, typeName, typeName, typeName, 
	typeName, ruleName, typeName)
}

// GenerateWithTest generates both rule template and test file
func (g *RuleTemplateGenerator) GenerateWithTest(ruleName string) error {
	if err := g.Generate(ruleName); err != nil {
		return err
	}

	// Generate test file
	testFileName := strings.ToLower(ruleName) + "_rule_test.go"
	testFilePath := filepath.Join(g.rulesDir, testFileName)
	
	testContent := g.generateTestTemplate(ruleName)
	
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		return fmt.Errorf("failed to write test file: %w", err)
	}

	fmt.Printf("✅ Test template created: %s\n", testFilePath)

	return nil
}

// generateTestTemplate creates a test template
func (g *RuleTemplateGenerator) generateTestTemplate(ruleName string) string {
	typeName := strings.Title(strings.ReplaceAll(ruleName, "-", "_"))
	typeName = strings.ReplaceAll(typeName, "_", "")
	
	return fmt.Sprintf(`package rules

import (
	"testing"
)

func Test%sRule_ID(t *testing.T) {
	rule := New%sRule()
	
	expected := "%s"
	if rule.ID() != expected {
		t.Errorf("Expected ID %%s, got %%s", expected, rule.ID())
	}
}

func Test%sRule_Name(t *testing.T) {
	rule := New%sRule()
	
	expected := "%s Rule"
	if rule.Name() != expected {
		t.Errorf("Expected Name %%s, got %%s", expected, rule.Name())
	}
}

func Test%sRule_Enabled(t *testing.T) {
	rule := New%sRule()
	
	if !rule.Enabled() {
		t.Error("Expected rule to be enabled by default")
	}
	
	rule.SetEnabled(false)
	if rule.Enabled() {
		t.Error("Expected rule to be disabled after SetEnabled(false)")
	}
}

func Test%sRule_Evaluate(t *testing.T) {
	rule := New%sRule()
	
	// TODO: Add test cases for Evaluate method
	//violations, err := rule.Evaluate(nil, "")
	//if err != nil {
	//	t.Fatalf("Unexpected error: %%v", err)
	//}
	
	// Add assertions based on expected behavior
	_ = rule
}
`, typeName, typeName, strings.ReplaceAll(ruleName, "-", "_"), 
	typeName, typeName, typeName, 
	typeName, typeName, 
	typeName, typeName)
}
