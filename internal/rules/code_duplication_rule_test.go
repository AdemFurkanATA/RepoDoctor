package rules

import (
	"reflect"
	"testing"
)

func TestCodeDuplicationRule_DetectsDuplicateBlocksAcrossFiles(t *testing.T) {
	rule := NewCodeDuplicationRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{
		{
			Path:    "a.go",
			Content: "package p\nfunc a() {\nvalue := 1\nvalue = value + 2\nvalue = value + 3\nvalue = value + 4\nvalue = value + 5\nvalue = value + 6\n}\n",
		},
		{
			Path:    "b.go",
			Content: "package p\nfunc b() {\nvalue := 1\nvalue = value + 2\nvalue = value + 3\nvalue = value + 4\nvalue = value + 5\nvalue = value + 6\n}\n",
		},
	}}

	violations := rule.Evaluate(ctx)
	if len(violations) == 0 {
		t.Fatal("expected duplication violations across two files")
	}
	if violations[0].RuleID != "rule.code-duplication" {
		t.Fatalf("unexpected rule id %q", violations[0].RuleID)
	}
}

func TestCodeDuplicationRule_SkipsAllowlistedPaths(t *testing.T) {
	rule := NewCodeDuplicationRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{
		{Path: "docs/snippets.go", Content: "a\nb\nc\nd\ne\nf\n"},
		{Path: "vendor/lib.go", Content: "a\nb\nc\nd\ne\nf\n"},
	}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected no violations for allowlisted paths, got %d", len(violations))
	}
}

func TestCodeDuplicationRule_IsDeterministic(t *testing.T) {
	rule := NewCodeDuplicationRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{
		{
			Path:    "z.go",
			Content: "package p\nfunc z() {\na\nb\nc\nd\ne\nf\n}\n",
		},
		{
			Path:    "a.go",
			Content: "package p\nfunc a() {\na\nb\nc\nd\ne\nf\n}\n",
		},
	}}

	first := rule.Evaluate(ctx)
	second := rule.Evaluate(ctx)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic output ordering\nfirst=%v\nsecond=%v", first, second)
	}
}
