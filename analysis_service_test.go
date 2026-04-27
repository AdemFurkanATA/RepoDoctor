package main

import (
	analysispkg "RepoDoctor/internal/analysis"
	"RepoDoctor/internal/model"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalysisService_RunAndRunAnalyze_SuccessPath(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatalf("failed writing go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "main.go"), []byte("package main\nfunc main(){}\n"), 0o644); err != nil {
		t.Fatalf("failed writing main.go: %v", err)
	}

	service := NewAnalysisService()
	code := service.Run(AnalyzeRequest{
		Path:            tmp,
		Format:          "json-v1",
		Verbose:         false,
		ColorEnabled:    false,
		ExitOnViolation: false,
	})
	if code != 0 {
		t.Fatalf("expected analysis service success code 0, got %d", code)
	}

	if code := runAnalyze(tmp, "text", false, false, false, profilingRequest{}, vulnerabilityCheckRequest{}); code != 0 {
		t.Fatalf("expected runAnalyze wrapper success code 0, got %d", code)
	}
}

func TestAnalysisService_ReportAdapterGraphHandlesNilAndPopulatedGraph(t *testing.T) {
	service := NewAnalysisService()
	progress := NewProgressReporter(false)

	empty := service.reportAdapterGraph(progress, &analysispkg.Result{AdapterName: "Go", Graph: nil}, false)
	if empty.GetNodeCount() != 0 || empty.GetEdgeCount() != 0 {
		t.Fatalf("expected empty graph from nil adapter graph, got nodes=%d edges=%d", empty.GetNodeCount(), empty.GetEdgeCount())
	}

	adapterGraph := model.NewDependencyGraph()
	adapterGraph.AddEdge("a", "b")
	built := service.reportAdapterGraph(progress, &analysispkg.Result{AdapterName: "Go", Graph: adapterGraph}, false)
	if built.GetNodeCount() != 2 || built.GetEdgeCount() != 1 {
		t.Fatalf("expected converted graph with 2 nodes/1 edge, got nodes=%d edges=%d", built.GetNodeCount(), built.GetEdgeCount())
	}
}
