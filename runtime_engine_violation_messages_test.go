package main

import (
	"testing"

	"RepoDoctor/internal/model"
)

func TestParseCyclePath_ExtractsDeterministicNodes(t *testing.T) {
	message := "a/service → b/repo → c/handler → a/service"
	path := parseCyclePath(message)
	if len(path) != 3 {
		t.Fatalf("expected 3 nodes after cycle closure trim, got %v", path)
	}
	if path[0] != "a/service" || path[1] != "b/repo" || path[2] != "c/handler" {
		t.Fatalf("unexpected parsed cycle path: %v", path)
	}
}

func TestParseCircularViolation_FallbackToFileWhenMessageUnavailable(t *testing.T) {
	v := model.Violation{RuleID: "rule.circular-dependency", File: "a.go", Severity: model.SeverityCritical, Message: ""}
	parsed := parseCircularViolation(v)
	if len(parsed.Path) != 1 || parsed.Path[0] != "a.go" {
		t.Fatalf("expected fallback path from file, got %v", parsed.Path)
	}
}

func TestParseLayerViolation_ImprovesMessageAndEndpoints(t *testing.T) {
	v := model.Violation{
		RuleID:   "rule.layer-validation",
		File:     "src/handler/user.go",
		Severity: model.SeverityError,
		Message:  "src/handler/user.go (handler) -> src/repo/user.go (repo): upward import not allowed [profile:layered]",
	}
	parsed := parseLayerViolation(v)
	if parsed.From != "src/handler/user.go" || parsed.To != "src/repo/user.go" {
		t.Fatalf("expected from/to extraction, got from=%q to=%q", parsed.From, parsed.To)
	}
	if parsed.Message != "handler -> repo: upward import not allowed [profile:layered]" {
		t.Fatalf("expected normalized layer message, got %q", parsed.Message)
	}
}

func TestApplyConfiguredSeverity_UsesRuleOverrides(t *testing.T) {
	cfg := &Config{Rules: &RulesConfig{
		CircularSeverity:  "warning",
		LayerSeverity:     "critical",
		SizeSeverity:      "error",
		GodObjectSeverity: "info",
	}}

	tests := []struct {
		ruleID string
		want   model.Severity
	}{
		{ruleID: "rule.circular-dependency", want: model.SeverityWarning},
		{ruleID: "rule.layer-validation", want: model.SeverityCritical},
		{ruleID: "rule.size", want: model.SeverityError},
		{ruleID: "rule.god-object", want: model.SeverityInfo},
	}

	for _, tc := range tests {
		base := model.Violation{RuleID: tc.ruleID, Severity: model.SeverityError}
		got := applyConfiguredSeverity(base, cfg)
		if got.Severity != tc.want {
			t.Fatalf("expected %s severity for %s, got %s", tc.want, tc.ruleID, got.Severity)
		}
	}
}
