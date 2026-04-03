package model

import "testing"

func TestNewRepositoryMetrics_InitialState(t *testing.T) {
	m := NewRepositoryMetrics()

	if m == nil {
		t.Fatal("expected non-nil metrics instance")
	}
	if len(m.Files) != 0 || len(m.Functions) != 0 || len(m.Structs) != 0 {
		t.Fatalf("expected empty slices, got files=%d functions=%d structs=%d", len(m.Files), len(m.Functions), len(m.Structs))
	}
	if m.TotalLines != 0 || m.TotalFiles != 0 || m.TotalFunctions != 0 || m.TotalStructs != 0 {
		t.Fatalf("expected zero totals, got %+v", *m)
	}
}

func TestRepositoryMetrics_AddFileMetrics_UpdatesTotals(t *testing.T) {
	m := NewRepositoryMetrics()

	m.AddFileMetrics(FileMetrics{Path: "a.go", Lines: 10, Functions: 2})
	m.AddFileMetrics(FileMetrics{Path: "b.go", Lines: 5, Functions: 1})

	if m.TotalFiles != 2 {
		t.Fatalf("expected TotalFiles=2, got %d", m.TotalFiles)
	}
	if m.TotalLines != 15 {
		t.Fatalf("expected TotalLines=15, got %d", m.TotalLines)
	}
	if m.TotalFunctions != 3 {
		t.Fatalf("expected TotalFunctions=3, got %d", m.TotalFunctions)
	}
}

func TestRepositoryMetrics_AddFunctionMetrics_AppendsOnly(t *testing.T) {
	m := NewRepositoryMetrics()
	m.AddFunctionMetrics(FunctionMetrics{Name: "A", File: "a.go", Lines: 10})
	m.AddFunctionMetrics(FunctionMetrics{Name: "B", File: "b.go", Lines: 5})

	if len(m.Functions) != 2 {
		t.Fatalf("expected 2 function metrics, got %d", len(m.Functions))
	}
	if m.TotalFunctions != 0 {
		t.Fatalf("expected TotalFunctions unchanged by AddFunctionMetrics, got %d", m.TotalFunctions)
	}
}

func TestRepositoryMetrics_AddStructMetrics_AppendsAndIncrementsTotal(t *testing.T) {
	m := NewRepositoryMetrics()
	m.AddStructMetrics(StructMetrics{Name: "A", File: "a.go", Fields: 2, Methods: 1})
	m.AddStructMetrics(StructMetrics{Name: "B", File: "b.go", Fields: 1, Methods: 0})

	if len(m.Structs) != 2 {
		t.Fatalf("expected 2 struct metrics, got %d", len(m.Structs))
	}
	if m.TotalStructs != 2 {
		t.Fatalf("expected TotalStructs=2, got %d", m.TotalStructs)
	}
}
