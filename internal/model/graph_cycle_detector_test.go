package model

import (
	"reflect"
	"testing"
)

func TestGraphCycleDetector_AcyclicGraph_NoCycles(t *testing.T) {
	g := NewDependencyGraph()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")

	detector := NewGraphCycleDetector(g)
	cycles := detector.DetectCycles()
	if len(cycles) != 0 {
		t.Fatalf("expected no cycles, got %v", cycles)
	}
	if detector.HasCycles() {
		t.Fatal("expected HasCycles false for acyclic graph")
	}
}

func TestGraphCycleDetector_SimpleCycle_Detected(t *testing.T) {
	g := NewDependencyGraph()
	g.AddEdge("a", "b")
	g.AddEdge("b", "a")

	detector := NewGraphCycleDetector(g)
	cycles := detector.DetectCycles()
	if len(cycles) == 0 {
		t.Fatal("expected at least one cycle")
	}

	sigs := cycleSignatures(cycles)
	found := false
	for _, sig := range sigs {
		if sig == "a->b" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected signature a->b among %v", sigs)
	}
}

func TestGraphCycleDetector_MultipleCycles_Detected(t *testing.T) {
	g := NewDependencyGraph()
	g.AddEdge("a", "b")
	g.AddEdge("b", "a")
	g.AddEdge("x", "y")
	g.AddEdge("y", "x")

	cycles := NewGraphCycleDetector(g).DetectCycles()
	sigs := cycleSignatures(cycles)

	if len(sigs) < 2 {
		t.Fatalf("expected at least two cycle signatures, got %v", sigs)
	}

	hasAB := false
	hasXY := false
	for _, sig := range sigs {
		if sig == "a->b" {
			hasAB = true
		}
		if sig == "x->y" {
			hasXY = true
		}
	}
	if !hasAB || !hasXY {
		t.Fatalf("expected cycle signatures for both components, got %v", sigs)
	}
}

func TestGraphCycleDetector_SelfLoop_Detected(t *testing.T) {
	g := NewDependencyGraph()
	g.AddEdge("self", "self")

	cycles := NewGraphCycleDetector(g).DetectCycles()
	sigs := cycleSignatures(cycles)

	foundSelf := false
	for _, sig := range sigs {
		if sig == "self" {
			foundSelf = true
			break
		}
	}
	if !foundSelf {
		t.Fatalf("expected self-loop cycle signature, got %v", sigs)
	}
}

func TestGraphCycleDetector_HasCycles_AndGetCyclesContract(t *testing.T) {
	g := NewDependencyGraph()
	g.AddEdge("a", "b")
	g.AddEdge("b", "a")

	detector := NewGraphCycleDetector(g)
	if detector.GetCycles() == nil {
		t.Fatal("expected initialized cycle storage")
	}
	if !detector.HasCycles() {
		t.Fatal("expected HasCycles true after implicit detection")
	}
	if len(detector.GetCycles()) == 0 {
		t.Fatal("expected detected cycles to be available via GetCycles")
	}
}

func TestGraphCycleDetector_DeterministicAcross10Runs_Canonicalized(t *testing.T) {
	g := NewDependencyGraph()
	g.AddEdge("a", "b")
	g.AddEdge("b", "a")
	g.AddEdge("b", "c")
	g.AddEdge("c", "d")
	g.AddEdge("d", "b")

	baseline := cycleSignatures(NewGraphCycleDetector(g).DetectCycles())

	for i := 0; i < 10; i++ {
		candidate := cycleSignatures(NewGraphCycleDetector(g).DetectCycles())
		if !reflect.DeepEqual(candidate, baseline) {
			t.Fatalf("non-deterministic cycle signatures at run %d: baseline=%v candidate=%v", i, baseline, candidate)
		}
	}
}
