package metrics

import (
	"path/filepath"
	"testing"
)

func TestSnapshotMarshal_Deterministic(t *testing.T) {
	snapshot := Snapshot{
		Version:          "1.1.0",
		Adapter:          "Go",
		OutputFormat:     "text",
		FilesDetected:    3,
		GraphNodes:       4,
		GraphEdges:       2,
		CircularCount:    0,
		LayerCount:       1,
		SizeCount:        2,
		GodObjectCount:   0,
		ComplexityLow:    3,
		ComplexityMedium: 1,
		ComplexityHigh:   0,
		TotalViolations:  3,
	}

	first, err := snapshot.Marshal()
	if err != nil {
		t.Fatalf("first marshal failed: %v", err)
	}
	second, err := snapshot.Marshal()
	if err != nil {
		t.Fatalf("second marshal failed: %v", err)
	}

	if string(first) != string(second) {
		t.Fatalf("expected deterministic serialization\nfirst=%s\nsecond=%s", string(first), string(second))
	}
}

func TestWrite_CreatesSnapshotFile(t *testing.T) {
	output := filepath.Join(t.TempDir(), ".repodoctor", "metrics", "snapshot.json")

	err := Write(output, Snapshot{Version: "1.1.0", Adapter: "Go", OutputFormat: "json"})
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
