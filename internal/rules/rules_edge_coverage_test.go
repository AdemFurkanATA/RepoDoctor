package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCategories_AllAndValidations(t *testing.T) {
	categories := AllCategories()
	if len(categories) < 5 {
		t.Fatalf("expected all standard categories, got %v", categories)
	}

	for _, c := range categories {
		if !IsValidCategory(string(c)) {
			t.Fatalf("expected category %q to be valid", c)
		}
	}

	if IsValidCategory("unknown") {
		t.Fatal("expected unknown category to be invalid")
	}
}

func TestExampleRule_ImplementsContract(t *testing.T) {
	rule := &ExampleRule{}
	if rule.ID() != "rule.example" {
		t.Fatalf("unexpected id: %s", rule.ID())
	}
	if rule.Category() != "structural" {
		t.Fatalf("unexpected category: %s", rule.Category())
	}
	if rule.Severity() != "info" {
		t.Fatalf("unexpected severity: %s", rule.Severity())
	}

	violations := rule.Evaluate(AnalysisContext{})
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %v", violations)
	}
}

func TestLoadFromDir_FiltersHiddenAndNonGoFiles(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, "ok.go"), []byte("package sample\n"), 0o644); err != nil {
		t.Fatalf("failed writing go file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("failed writing non-go file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".hidden.go"), []byte("package hidden\n"), 0o644); err != nil {
		t.Fatalf("failed writing hidden file: %v", err)
	}

	hiddenDir := filepath.Join(root, ".git")
	if err := os.MkdirAll(hiddenDir, 0o755); err != nil {
		t.Fatalf("failed creating hidden dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "ignored.go"), []byte("package git\n"), 0o644); err != nil {
		t.Fatalf("failed writing hidden-dir go file: %v", err)
	}

	files, err := LoadFromDir(root)
	if err != nil {
		t.Fatalf("LoadFromDir returned error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected exactly one visible go file, got %d (%v)", len(files), files)
	}
	if filepath.Base(files[0].Path) != "ok.go" {
		t.Fatalf("expected ok.go, got %s", files[0].Path)
	}
}
