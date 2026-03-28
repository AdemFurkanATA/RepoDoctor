package main

import (
	"path/filepath"
	"testing"

	"RepoDoctor/internal/model"
)

func TestSortedStringCopy_DoesNotMutateInput(t *testing.T) {
	input := []string{"c", "a", "b"}
	got := sortedStringCopy(input)

	if got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("expected sorted copy [a b c], got %v", got)
	}

	if input[0] != "c" || input[1] != "a" || input[2] != "b" {
		t.Fatalf("expected input slice to stay unchanged, got %v", input)
	}
}

func TestSortModelViolationsDeterministic_Order(t *testing.T) {
	violations := []model.Violation{
		{RuleID: "rule.size", File: "b.go", Line: 12, Message: "z"},
		{RuleID: "rule.god-object", File: "a.go", Line: 1, Message: "a"},
		{RuleID: "rule.size", File: "a.go", Line: 2, Message: "b"},
		{RuleID: "rule.size", File: "a.go", Line: 1, Message: "c"},
	}

	sortModelViolationsDeterministic(violations)

	if violations[0].RuleID != "rule.god-object" || violations[1].Line != 1 || violations[2].Line != 2 || violations[3].File != "b.go" {
		t.Fatalf("unexpected deterministic violation order: %+v", violations)
	}
}

func TestNormalizeReportPathDeterministic_SlashNormalization(t *testing.T) {
	path := filepath.Join(".", "demo", "repo")
	got := normalizeReportPathDeterministic(path)
	if got != "demo/repo" {
		t.Fatalf("expected normalized slash path demo/repo, got %s", got)
	}
}
