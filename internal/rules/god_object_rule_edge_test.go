package rules

import (
	"strings"
	"testing"
)

func TestGodObjectRule_MetadataAndCapabilities(t *testing.T) {
	rule := NewGodObjectRule()

	if rule.ID() != "rule.god-object" {
		t.Fatalf("unexpected id: %s", rule.ID())
	}
	if rule.Category() != string(CategoryMaintainability) {
		t.Fatalf("unexpected category: %s", rule.Category())
	}
	if rule.Severity() != "warning" {
		t.Fatalf("unexpected severity: %s", rule.Severity())
	}

	caps := rule.Capabilities()
	if caps.SupportsMultipleLanguages {
		t.Fatal("god object rule should be single-language")
	}
	if len(caps.SupportedLanguages) != 1 || caps.SupportedLanguages[0] != "Go" {
		t.Fatalf("unexpected supported languages: %v", caps.SupportedLanguages)
	}
}

func TestGodObjectRule_ThresholdBoundary_NoViolationAtExactThreshold(t *testing.T) {
	rule := NewGodObjectRule()
	rule.MaxFields = 2
	rule.MaxMethods = 2

	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{{
		Path: "pkg/exact.go",
		Content: `package pkg
type User struct {
	ID int
	Name string
}
func (u User) First() {}
func (u *User) Second() {}`,
	}}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected no violations at exact thresholds, got %v", violations)
	}
}

func TestGodObjectRule_ThresholdBoundary_ViolatesWhenExceeded(t *testing.T) {
	rule := NewGodObjectRule()
	rule.MaxFields = 2
	rule.MaxMethods = 2

	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{{
		Path: "pkg/exceed.go",
		Content: `package pkg
type User struct {
	A int
	B int
	C int
}
func (u User) One() {}
func (u User) Two() {}
func (u User) Three() {}`,
	}}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 2 {
		t.Fatalf("expected field and method violations, got %d (%v)", len(violations), violations)
	}

	joined := violations[0].Message + "\n" + violations[1].Message
	if !strings.Contains(joined, "fields (threshold: 2)") {
		t.Fatalf("expected field threshold detail in messages, got %q", joined)
	}
	if !strings.Contains(joined, "methods (threshold: 2)") {
		t.Fatalf("expected method threshold detail in messages, got %q", joined)
	}
}

func TestGodObjectRule_Boundary_StructNameCollisionAcrossDirs(t *testing.T) {
	rule := NewGodObjectRule()
	rule.MaxFields = 10
	rule.MaxMethods = 1

	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{
		{
			Path: "module/a/user.go",
			Content: `package a
type User struct{}
func (u User) A() {}`,
		},
		{
			Path: "module/b/user.go",
			Content: `package b
type User struct{}
func (u User) B() {}`,
		},
	}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected no collision-induced violations across directories, got %v", violations)
	}
}

func TestGodObjectRule_Evaluate_MalformedFileIgnored(t *testing.T) {
	rule := NewGodObjectRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{{
		Path:    "pkg/bad.go",
		Content: "package broken\ntype X struct {\nfunc(",
	}}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected malformed file to be ignored, got %v", violations)
	}
}
