package main

import "testing"

func TestDetermineExitCode_NoViolationsReturnsZero(t *testing.T) {
	report := &StructuralReport{HasViolations: false}
	if got := determineExitCode(report); got != 0 {
		t.Fatalf("expected exit code 0 for no violations, got %d", got)
	}
}

func TestDetermineExitCode_CircularOrLayerViolationReturnsTwo(t *testing.T) {
	tests := []struct {
		name   string
		report *StructuralReport
	}{
		{
			name: "circular violation",
			report: &StructuralReport{
				HasViolations: true,
				Circular:      []CycleViolation{{Path: []string{"a", "b"}}},
			},
		},
		{
			name: "layer violation",
			report: &StructuralReport{
				HasViolations: true,
				Layer:         []LayerViolation{{From: "repo", To: "handler"}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := determineExitCode(tc.report); got != 2 {
				t.Fatalf("expected exit code 2 for %s, got %d", tc.name, got)
			}
		})
	}
}

func TestDetermineExitCode_SizeAndGodObjectViolationsRemainNonBlocking(t *testing.T) {
	report := &StructuralReport{
		HasViolations: true,
		Size:          []SizeViolation{{File: "big.go", Lines: 900, Threshold: 500}},
		GodObject:     []GodObjectViolation{{File: "god.go", StructName: "GodType", FieldCount: 25}},
	}

	if got := determineExitCode(report); got != 0 {
		t.Fatalf("expected exit code 0 for size/god-object only violations, got %d", got)
	}
}
