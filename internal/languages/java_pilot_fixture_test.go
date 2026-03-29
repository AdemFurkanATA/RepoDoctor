package languages

import (
	"path/filepath"
	"testing"
)

func TestJavaPilotFixture_DeterministicEvidence(t *testing.T) {
	fixture := filepath.Join("testdata", "java_pilot_fixture")
	adapter := NewJavaAdapter()

	firstFiles, err := adapter.DetectFiles(fixture)
	if err != nil {
		t.Fatalf("DetectFiles failed: %v", err)
	}
	if len(firstFiles) < 2 {
		t.Fatalf("expected at least 2 Java files in fixture, got %d", len(firstFiles))
	}

	firstMetrics, err := adapter.CollectMetrics(firstFiles)
	if err != nil {
		t.Fatalf("CollectMetrics failed: %v", err)
	}
	if firstMetrics.TotalFiles == 0 || firstMetrics.TotalFunctions == 0 {
		t.Fatalf("expected non-empty Java metrics from fixture, got %+v", firstMetrics)
	}

	firstGraph, err := adapter.BuildDependencyGraph(firstFiles)
	if err != nil {
		t.Fatalf("BuildDependencyGraph failed: %v", err)
	}
	firstNodes := firstGraph.NodeCount()
	firstEdges := firstGraph.EdgeCount()

	for i := 0; i < 20; i++ {
		nextFiles, nextErr := adapter.DetectFiles(fixture)
		if nextErr != nil {
			t.Fatalf("DetectFiles failed on iteration %d: %v", i, nextErr)
		}
		if len(nextFiles) != len(firstFiles) {
			t.Fatalf("detect files parity drift at %d: first=%d next=%d", i, len(firstFiles), len(nextFiles))
		}

		nextGraph, graphErr := adapter.BuildDependencyGraph(nextFiles)
		if graphErr != nil {
			t.Fatalf("BuildDependencyGraph failed on iteration %d: %v", i, graphErr)
		}
		if nextGraph.NodeCount() != firstNodes || nextGraph.EdgeCount() != firstEdges {
			t.Fatalf("graph parity drift at %d: nodes %d/%d edges %d/%d", i, firstNodes, nextGraph.NodeCount(), firstEdges, nextGraph.EdgeCount())
		}
	}
}
