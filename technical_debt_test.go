package main

import "testing"

func TestEstimateTechnicalDebt_DeterministicBreakdown(t *testing.T) {
	report := &StructuralReport{
		Circular:  []CycleViolation{{}, {}},
		Layer:     []LayerViolation{{}},
		Size:      []SizeViolation{{}, {}, {}},
		GodObject: []GodObjectViolation{{}},
		Complexity: ComplexityBandSummary{
			Medium: 2,
			High:   1,
		},
	}

	debt := estimateTechnicalDebt(report)
	if debt.CircularHours != 8 || debt.LayerHours != 3 || debt.SizeHours != 3 || debt.GodObjectHours != 2 || debt.ComplexityHours != 4 {
		t.Fatalf("unexpected debt breakdown: %+v", debt)
	}
	if debt.TotalHours != 20 {
		t.Fatalf("expected total 20, got %d", debt.TotalHours)
	}
}
