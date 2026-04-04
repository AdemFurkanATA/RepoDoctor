package main

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseBoolEnv_DefaultOffAndTruthyValues(t *testing.T) {
	t.Setenv(interactiveFixSuggestionsEnv, "")
	if parseBoolEnv(interactiveFixSuggestionsEnv) {
		t.Fatal("expected empty env to keep interactive suggestions disabled")
	}

	t.Setenv(interactiveFixSuggestionsEnv, "yes")
	if !parseBoolEnv(interactiveFixSuggestionsEnv) {
		t.Fatal("expected 'yes' to enable interactive suggestions")
	}
}

func TestBuildInteractiveFixSuggestions_UsesHintsWithFallbacks(t *testing.T) {
	report := &StructuralReport{
		HasViolations: true,
		Circular:      []CycleViolation{{Hint: "cycle hint"}},
		Layer:         []LayerViolation{{Hint: ""}},
		Size:          []SizeViolation{{Hint: "size hint"}},
		GodObject:     []GodObjectViolation{{Hint: ""}},
	}

	got := buildInteractiveFixSuggestions(report)
	if len(got) != 4 {
		t.Fatalf("expected one suggestion per violation class, got %d", len(got))
	}
	if got[0] != "cycle hint" {
		t.Fatalf("expected explicit circular hint first, got %q", got[0])
	}
	if !strings.Contains(got[1], "interface") {
		t.Fatalf("expected layer fallback suggestion, got %q", got[1])
	}
}

func TestInteractiveFixSuggestionAdvisor_MaybePrintAdvisoryOnly(t *testing.T) {
	advisor := &InteractiveFixSuggestionAdvisor{
		Enabled: true,
		io:      &interactiveIO{reader: bufio.NewReader(strings.NewReader("yes\n"))},
	}
	report := &StructuralReport{
		HasViolations: true,
		Size:          []SizeViolation{{Hint: "split long function"}},
	}

	out := captureStdout(t, func() {
		advisor.MaybePrint(report)
	})
	if !strings.Contains(out, "advisory only") {
		t.Fatalf("expected advisory-only wording in output, got %q", out)
	}
	if !strings.Contains(out, "split long function") {
		t.Fatalf("expected rendered suggestion in output, got %q", out)
	}
}
