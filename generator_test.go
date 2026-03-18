package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuleTemplateGenerator_GenerateCreatesCompilableTemplate(t *testing.T) {
	tempDir := t.TempDir()
	generator := NewRuleTemplateGenerator(tempDir)

	if err := generator.Generate("large-interface"); err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	generatedFile := filepath.Join(tempDir, "large_interface_rule.go")
	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("failed reading generated file: %v", err)
	}

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, generatedFile, content, parser.AllErrors); err != nil {
		t.Fatalf("generated template has invalid Go syntax: %v", err)
	}
}

func TestRuleTemplateGenerator_GenerateSanitizesRuleName(t *testing.T) {
	tempDir := t.TempDir()
	generator := NewRuleTemplateGenerator(tempDir)

	if err := generator.Generate("  Large Interface  "); err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	generatedFile := filepath.Join(tempDir, "large_interface_rule.go")
	if _, err := os.Stat(generatedFile); err != nil {
		t.Fatalf("expected sanitized file path to exist: %v", err)
	}

	content, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("failed reading generated file: %v", err)
	}

	if !strings.Contains(string(content), "type LargeInterfaceRule struct") {
		t.Fatalf("expected generated type name to be LargeInterfaceRule")
	}
}

func TestSanitizeRuleName_RejectsInvalidCharacters(t *testing.T) {
	_, err := sanitizeRuleName("invalid/rule")
	if err == nil {
		t.Fatal("expected error for invalid rule name")
	}
}
