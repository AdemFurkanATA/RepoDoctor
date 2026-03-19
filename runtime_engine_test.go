package main

import (
	"testing"

	"RepoDoctor/internal/model"
)

func TestSortViolations_DeterministicOrder(t *testing.T) {
	violations := []model.Violation{
		{RuleID: "rule.size", File: "b.go", Line: 20, Message: "m2"},
		{RuleID: "rule.circular-dependency", File: "a.go", Line: 1, Message: "m1"},
		{RuleID: "rule.size", File: "a.go", Line: 5, Message: "m0"},
	}

	sortViolations(violations)

	if violations[0].RuleID != "rule.circular-dependency" {
		t.Fatalf("expected rule.circular-dependency first, got %s", violations[0].RuleID)
	}

	if violations[1].File != "a.go" || violations[2].File != "b.go" {
		t.Fatalf("expected size violations ordered by file, got %s then %s", violations[1].File, violations[2].File)
	}
}
