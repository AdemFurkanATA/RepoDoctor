package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Snapshot is a deterministic metrics export payload for one analyze run.
type Snapshot struct {
	Version         string `json:"version"`
	Adapter         string `json:"adapter"`
	OutputFormat    string `json:"outputFormat"`
	FilesDetected   int    `json:"filesDetected"`
	GraphNodes      int    `json:"graphNodes"`
	GraphEdges      int    `json:"graphEdges"`
	CircularCount   int    `json:"circularCount"`
	LayerCount      int    `json:"layerCount"`
	SizeCount       int    `json:"sizeCount"`
	GodObjectCount  int    `json:"godObjectCount"`
	TotalViolations int    `json:"totalViolations"`
}

// Marshal serializes the snapshot in deterministic key order.
func (s Snapshot) Marshal() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// Write writes the snapshot to disk using deterministic JSON representation.
func Write(path string, snapshot Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	payload, err := snapshot.Marshal()
	if err != nil {
		return err
	}

	return os.WriteFile(path, payload, 0o644)
}
