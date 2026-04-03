package main

import (
	"RepoDoctor/internal/engine"
	"strings"
	"testing"
)

func TestGenerateReport_OutputModes_TextJSONJSONV1(t *testing.T) {
	scorer := NewStructuralScorer(NewDependencyGraph(), nil, "")

	t.Run("text", func(t *testing.T) {
		output := captureStdout(t, func() {
			_ = generateReport(scorer, ".", "text", false, false)
		})
		if !strings.Contains(output, "RepoDoctor Structural Analysis Report") {
			t.Fatalf("expected text report header, got %q", output)
		}
	})

	t.Run("json", func(t *testing.T) {
		output := captureStdout(t, func() {
			_ = generateReport(scorer, ".", "json", false, false)
		})
		if !strings.Contains(output, "\"schemaVersion\": \"v2\"") {
			t.Fatalf("expected json output to include schema version, got %q", output)
		}
	})

	t.Run("json-v1", func(t *testing.T) {
		output := captureStdout(t, func() {
			_ = generateReport(scorer, ".", "json-v1", false, false)
		})
		if !strings.Contains(output, "\"version\"") {
			t.Fatalf("expected json-v1 output to include version field, got %q", output)
		}
		if strings.Contains(output, "\"schemaVersion\"") {
			t.Fatalf("expected json-v1 output to omit v2 schema field, got %q", output)
		}
	})
}

func TestGenerateRuleEngineReport_JSONV1Path(t *testing.T) {
	summary := &runtimeRuleSummary{
		result:       &engine.ExecutionResult{Violations: nil, RulesExecuted: 0, TimedOut: false},
		rulesInScope: 0,
	}

	output := captureStdout(t, func() {
		_ = generateRuleEngineReport(".", "json-v1", false, false, nil, summary)
	})

	if !strings.Contains(output, "\"version\"") {
		t.Fatalf("expected json-v1 rule-engine output to include version, got %q", output)
	}
	if strings.Contains(output, "\"schemaVersion\"") {
		t.Fatalf("expected json-v1 rule-engine output to omit schemaVersion, got %q", output)
	}
}
