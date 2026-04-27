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

func TestCITemplateGenerator_GenerateGitHubTemplate(t *testing.T) {
	baseDir := t.TempDir()
	generator := NewCITemplateGenerator(baseDir)

	if err := generator.Generate("github", false); err != nil {
		t.Fatalf("expected github template generation to succeed, got: %v", err)
	}

	targetPath := filepath.Join(baseDir, ".github", "workflows", "repodoctor.yml")
	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("expected github template to be written, got: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "name: repodoctor") || !strings.Contains(text, "go run . analyze -path .") {
		t.Fatalf("expected deterministic github template markers, got: %s", text)
	}
}

func TestCITemplateGenerator_UnsupportedProvider(t *testing.T) {
	generator := NewCITemplateGenerator(t.TempDir())
	if err := generator.Generate("bitbucket", false); err == nil {
		t.Fatal("expected unsupported provider to fail")
	}
}

func TestCITemplateGenerator_ExistingFileRequiresForce(t *testing.T) {
	baseDir := t.TempDir()
	targetPath := filepath.Join(baseDir, ".gitlab-ci.yml")
	if err := os.WriteFile(targetPath, []byte("sentinel"), 0o644); err != nil {
		t.Fatalf("failed creating existing file: %v", err)
	}

	generator := NewCITemplateGenerator(baseDir)
	if err := generator.Generate("gitlab", false); err == nil {
		t.Fatal("expected existing file without force to fail")
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed reading existing file: %v", err)
	}
	if string(content) != "sentinel" {
		t.Fatalf("expected existing file not to be modified without force, got: %q", string(content))
	}
}

func TestCITemplateGenerator_ForceOverwriteReplacesFile(t *testing.T) {
	baseDir := t.TempDir()
	targetPath := filepath.Join(baseDir, "azure-pipelines.yml")
	if err := os.WriteFile(targetPath, []byte("sentinel"), 0o644); err != nil {
		t.Fatalf("failed creating existing file: %v", err)
	}

	generator := NewCITemplateGenerator(baseDir)
	if err := generator.Generate("azure", true); err != nil {
		t.Fatalf("expected force overwrite to succeed, got: %v", err)
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed reading overwritten file: %v", err)
	}
	if strings.Contains(string(content), "sentinel") {
		t.Fatal("expected force overwrite to replace existing content")
	}
}

func TestRunGenerate_CIMissingProviderReturnsUsageError(t *testing.T) {
	err := runGenerate([]string{"ci"})
	if err == nil {
		t.Fatal("expected usage error for missing ci provider")
	}
}

func TestRunGenerate_RulePathRemainsBackwardCompatible(t *testing.T) {
	baseDir := t.TempDir()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed reading working directory: %v", err)
	}
	if err := os.Chdir(baseDir); err != nil {
		t.Fatalf("failed switching to temp directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	if err := runGenerate([]string{"rule", "quality-gate"}); err != nil {
		t.Fatalf("expected generate rule path to remain compatible, got: %v", err)
	}

	targetPath := filepath.Join(baseDir, "rules", "quality_gate_rule.go")
	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("expected generated rule file at %s, got error: %v", targetPath, err)
	}
}

func TestDocsGenerator_GenerateMarkdownCreatesFile(t *testing.T) {
	baseDir := t.TempDir()
	generator := NewDocsGenerator(baseDir)

	if err := generator.Generate("markdown"); err != nil {
		t.Fatalf("expected docs markdown generation to succeed, got: %v", err)
	}

	targetPath := filepath.Join(baseDir, "docs", "architecture.generated.md")
	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("expected markdown docs file to exist, got: %v", err)
	}
	if !strings.Contains(string(content), "RepoDoctor Architecture (Generated)") {
		t.Fatalf("unexpected markdown docs content: %s", string(content))
	}
}

func TestDocsGenerator_NonDestructiveWritesPreviewInsteadOfOverwrite(t *testing.T) {
	baseDir := t.TempDir()
	targetPath := filepath.Join(baseDir, "docs", "architecture.generated.mmd")
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("failed creating docs directory: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("legacy"), 0o644); err != nil {
		t.Fatalf("failed writing existing docs file: %v", err)
	}

	generator := NewDocsGenerator(baseDir)
	if err := generator.Generate("mermaid"); err != nil {
		t.Fatalf("expected mermaid docs generation to succeed, got: %v", err)
	}

	original, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("failed reading original target: %v", err)
	}
	if string(original) != "legacy" {
		t.Fatalf("expected original file to remain untouched, got: %s", string(original))
	}

	previewPath := targetPath + ".new"
	if _, err := os.Stat(previewPath); err != nil {
		t.Fatalf("expected preview file to be created, got: %v", err)
	}
}
