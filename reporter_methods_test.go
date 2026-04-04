package main

import (
	"strings"
	"testing"
)

func TestReporterWriteSections_TextMethods(t *testing.T) {
	report := &StructuralReport{
		Version: "1.1.0",
		Path:    ".",
		Score: &StructuralScore{
			TotalScore:       85,
			MaxScore:         100,
			CircularPenalty:  10,
			LayerPenalty:     5,
			SizePenalty:      0,
			GodObjectPenalty: 0,
		},
		Circular: []CycleViolation{{Path: []string{"a", "b"}, Hint: "Break the cycle by extracting shared contracts."}},
		Layer:    []LayerViolation{{From: "repo", To: "handler", Hint: "Move dependency behind an interface boundary."}},
		Size:     []SizeViolation{{File: "big.go", Lines: 800, Threshold: 500, Hint: "Split into smaller focused units."}},
		GodObject: []GodObjectViolation{{
			File:        "god.go",
			StructName:  "God",
			FieldCount:  20,
			MethodCount: 12,
			Hint:        "Extract cohesive collaborators from God.",
		}},
		HasViolations: true,
	}

	var sb strings.Builder
	writeHeader(&sb)
	writeScoreSection(&sb, report)
	writeViolationsSummary(&sb, report)
	writeCircularViolations(&sb, report)
	writeLayerViolations(&sb, report)
	writeSizeViolations(&sb, report)
	writeGodObjectViolations(&sb, report)
	writeScoreBreakdown(&sb, report)

	out := sb.String()
	checks := []string{"STRUCTURAL HEALTH SCORE", "CIRCULAR DEPENDENCIES", "LAYER VIOLATIONS", "SIZE VIOLATIONS", "GOD OBJECT VIOLATIONS", "SCORE BREAKDOWN"}
	for _, token := range checks {
		if !strings.Contains(out, token) {
			t.Fatalf("expected output to include %q, got %q", token, out)
		}
	}
	if !strings.Contains(out, "Hint:") {
		t.Fatalf("expected actionable hint lines in output, got %q", out)
	}
}
