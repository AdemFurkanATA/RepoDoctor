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

// Description returns what this rule checks.
func (r *{{.TypeName}}Rule) Description() string {
	return "TODO: describe {{.RuleName}} rule behavior"
}

// Category returns the rule category.
func (r *{{.TypeName}}Rule) Category() string {
	return "maintainability"
}

// Severity returns the default severity.
func (r *{{.TypeName}}Rule) Severity() string {
	return "warning"
}

// Enabled indicates whether this rule is enabled by default.
func (r *{{.TypeName}}Rule) Enabled() bool {
	return r.enabled
}

// SetEnabled toggles the rule state.
func (r *{{.TypeName}}Rule) SetEnabled(enabled bool) {
	r.enabled = enabled
}

// Configure validates and applies optional settings.
func (r *{{.TypeName}}Rule) Configure(config map[string]interface{}) error {
	_ = config
	return nil
}

// Evaluate checks for {{.RuleName}} violations in the repository
func (r *{{.TypeName}}Rule) Evaluate(rootPath string) ([]Violation, error) {
	if rootPath == "" {
		return nil, fmt.Errorf("rootPath cannot be empty")
	}

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
}

const simpleRuleTemplate = `package rules

import (
	"fmt"
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
	_ = config
	return nil
}

// Evaluate checks for %s violations in the repository
func (r *%sRule) Evaluate(rootPath string) ([]Violation, error) {
	if rootPath == "" {
		return nil, fmt.Errorf("rootPath cannot be empty")
	}

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

// generateSimpleTemplate creates a simpler template if template rendering fails
func (g *RuleTemplateGenerator) generateSimpleTemplate(ruleName, typeName string) string {
	ruleID := strings.ReplaceAll(ruleName, "-", "_")

	return fmt.Sprintf(
		simpleRuleTemplate,
		typeName, ruleName, typeName,
		typeName, typeName, typeName, typeName, typeName, typeName,
		typeName, ruleID,
		typeName, ruleName,
		typeName, ruleName,
		typeName,
		typeName,
		typeName,
		typeName,
		typeName,
		ruleName,
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
	
	expected := "%s"
	if rule.ID() != expected {
		t.Errorf("Expected ID %%s, got %%s", expected, rule.ID())
	}
}

func Test%sRule_Name(t *testing.T) {
	rule := New%sRule()
	
	expected := "%s"
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
	
	if _, err := rule.Evaluate(""); err == nil {
		t.Fatal("expected error for empty rootPath")
	}
	
	if _, err := rule.Evaluate("."); err != nil {
		t.Fatalf("unexpected error: %%v", err)
	}
}
`, typeName, typeName, strings.ReplaceAll(ruleName, "-", "_"),
		typeName, typeName, ruleName,
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
