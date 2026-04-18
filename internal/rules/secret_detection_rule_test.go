package rules

import (
	"reflect"
	"testing"
)

func TestSecretDetectionRule_DetectsHighRiskSecretPatterns(t *testing.T) {
	rule := NewSecretDetectionRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{{
		Path:    "internal/config/secrets.go",
		Content: "package config\nconst GitHubToken = \"ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\"\n", //nolint:lll
	}}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 1 {
		t.Fatalf("expected one secret violation, got %d", len(violations))
	}
	if violations[0].Severity != "critical" {
		t.Fatalf("expected critical severity, got %q", violations[0].Severity)
	}
	if violations[0].Line != 2 {
		t.Fatalf("expected violation line 2, got %d", violations[0].Line)
	}
}

func TestSecretDetectionRule_AllowlistAndEntropyReduceFalsePositives(t *testing.T) {
	rule := NewSecretDetectionRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{
		{Path: "docs/examples.md", Content: "token = \"example123\"\n"},
		{Path: "app/config.go", Content: "const apiKey = \"aaaaaaaaaaaaaaaaaaaaaaaaaaaa\" // repodoctor:allow-secret\n"},
		{Path: "app/low_entropy.go", Content: "const password = \"aaaaabbbbbccccc\"\n"},
	}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected no violations for allowlisted/low-entropy examples, got %d", len(violations))
	}
}

func TestSecretDetectionRule_SkipsOversizedAndBinaryFilesFailSoft(t *testing.T) {
	rule := NewSecretDetectionRule()
	content := make([]byte, defaultSecretMaxScanBytes+10)
	for i := range content {
		content[i] = 'A'
	}

	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{
		{Path: "vendor/big_blob.go", Content: string(content)},
		{Path: "assets/blob.bin", Content: "abc\x00defghijklmnopqrstuvwxyz"},
	}}

	violations := rule.Evaluate(ctx)
	if len(violations) != 0 {
		t.Fatalf("expected fail-soft skip for oversized/binary content, got %d violations", len(violations))
	}
}

func TestSecretDetectionRule_OutputOrderingDeterministic(t *testing.T) {
	rule := NewSecretDetectionRule()
	ctx := AnalysisContext{RepositoryFiles: []RepositoryFile{
		{
			Path:    "b.go",
			Content: "const token = \"ghp_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\"\n",
		},
		{
			Path:    "a.go",
			Content: "const token = \"ghp_BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB\"\n",
		},
	}}

	first := rule.Evaluate(ctx)
	second := rule.Evaluate(ctx)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected deterministic output ordering\nfirst=%v\nsecond=%v", first, second)
	}
	if len(first) != 2 || first[0].File != "a.go" || first[1].File != "b.go" {
		t.Fatalf("expected sorted violation order by file, got %+v", first)
	}
}
