package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrendAnalyzer_NewAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()
	analyzer := NewTrendAnalyzer(tmpDir)

	if analyzer.historyPath != filepath.Join(tmpDir, ".repodoctor", "history.json") {
		t.Errorf("Expected history path %s, got %s", filepath.Join(tmpDir, ".repodoctor", "history.json"), analyzer.historyPath)
	}

	if len(analyzer.history) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(analyzer.history))
	}
}

func TestTrendAnalyzer_LoadHistory_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	analyzer := NewTrendAnalyzer(tmpDir)

	err := analyzer.LoadHistory()
	if err != nil {
		t.Errorf("Expected no error when file doesn't exist, got: %v", err)
	}

	if len(analyzer.history) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(analyzer.history))
	}
}

func TestTrendAnalyzer_AppendScore(t *testing.T) {
	tmpDir := t.TempDir()
	analyzer := NewTrendAnalyzer(tmpDir)

	// Load (empty) history
	err := analyzer.LoadHistory()
	if err != nil {
		t.Errorf("Expected no error loading empty history: %v", err)
	}

	// Append score
	err = analyzer.AppendScore(85.5)
	if err != nil {
		t.Errorf("Expected no error appending score: %v", err)
	}

	// Verify history
	if len(analyzer.history) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(analyzer.history))
	}

	if analyzer.history[0].Score != 85.5 {
		t.Errorf("Expected score 85.5, got %.1f", analyzer.history[0].Score)
	}

	// Verify file exists
	historyPath := filepath.Join(tmpDir, ".repodoctor", "history.json")
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		t.Error("Expected history file to exist")
	}
}

func TestTrendAnalyzer_CalculateDelta(t *testing.T) {
	tmpDir := t.TempDir()
	analyzer := NewTrendAnalyzer(tmpDir)

	// No history
	delta, trend, hasPrevious := analyzer.CalculateDelta(80.0)
	if hasPrevious {
		t.Error("Expected no previous data")
	}
	if trend != "no previous data" {
		t.Errorf("Expected 'no previous data', got '%s'", trend)
	}
	if delta != 0 {
		t.Errorf("Expected delta 0, got %.1f", delta)
	}

	// Add first entry
	analyzer.AppendScore(75.0)

	// Still no previous (only one entry)
	delta, trend, hasPrevious = analyzer.CalculateDelta(80.0)
	if hasPrevious {
		t.Error("Expected no previous data with only one entry")
	}

	// Add second entry
	analyzer.AppendScore(75.0)

	// Now we have previous
	delta, trend, hasPrevious = analyzer.CalculateDelta(80.0)
	if !hasPrevious {
		t.Error("Expected previous data")
	}
	if delta != 5.0 {
		t.Errorf("Expected delta 5.0, got %.1f", delta)
	}
	if trend != "increased" {
		t.Errorf("Expected 'increased', got '%s'", trend)
	}
}

func TestTrendAnalyzer_DecreasedTrend(t *testing.T) {
	tmpDir := t.TempDir()
	analyzer := NewTrendAnalyzer(tmpDir)

	// Add entries
	analyzer.AppendScore(90.0)
	analyzer.AppendScore(90.0)

	delta, trend, hasPrevious := analyzer.CalculateDelta(85.0)

	if !hasPrevious {
		t.Error("Expected previous data")
	}
	if delta != -5.0 {
		t.Errorf("Expected delta -5.0, got %.1f", delta)
	}
	if trend != "decreased" {
		t.Errorf("Expected 'decreased', got '%s'", trend)
	}
}

func TestTrendAnalyzer_UnchangedTrend(t *testing.T) {
	tmpDir := t.TempDir()
	analyzer := NewTrendAnalyzer(tmpDir)

	// Add entries
	analyzer.AppendScore(85.0)
	analyzer.AppendScore(85.0)

	delta, trend, hasPrevious := analyzer.CalculateDelta(85.0)

	if !hasPrevious {
		t.Error("Expected previous data")
	}
	if delta != 0 {
		t.Errorf("Expected delta 0, got %.1f", delta)
	}
	if trend != "unchanged" {
		t.Errorf("Expected 'unchanged', got '%s'", trend)
	}
}

func TestTrendAnalyzer_GetTrendSummary(t *testing.T) {
	tmpDir := t.TempDir()
	analyzer := NewTrendAnalyzer(tmpDir)

	// No history
	summary := analyzer.GetTrendSummary(80.0)
	if summary != "Current Score: No previous data available" {
		t.Errorf("Expected no previous data message, got: %s", summary)
	}

	// Add history
	analyzer.AppendScore(75.0)
	analyzer.AppendScore(75.0)

	summary = analyzer.GetTrendSummary(80.0)

	if summary != "Current Score: 80.0\nPrevious Score: 75.0\nDelta: +5.0 (increased)" {
		t.Errorf("Expected trend summary with increase, got: %s", summary)
	}
}

func TestTrendAnalyzer_GetHistoryLength(t *testing.T) {
	tmpDir := t.TempDir()
	analyzer := NewTrendAnalyzer(tmpDir)

	if analyzer.GetHistoryLength() != 0 {
		t.Errorf("Expected history length 0, got %d", analyzer.GetHistoryLength())
	}

	analyzer.AppendScore(80.0)
	analyzer.AppendScore(85.0)

	if analyzer.GetHistoryLength() != 2 {
		t.Errorf("Expected history length 2, got %d", analyzer.GetHistoryLength())
	}
}

func TestTrendAnalyzer_GetLastEntry(t *testing.T) {
	tmpDir := t.TempDir()
	analyzer := NewTrendAnalyzer(tmpDir)

	// No history
	entry, ok := analyzer.GetLastEntry()
	if ok {
		t.Error("Expected no last entry")
	}
	if entry != nil {
		t.Error("Expected nil entry")
	}

	// Add entry
	analyzer.AppendScore(85.5)

	entry, ok = analyzer.GetLastEntry()
	if !ok {
		t.Error("Expected last entry")
	}
	if entry.Score != 85.5 {
		t.Errorf("Expected score 85.5, got %.1f", entry.Score)
	}
}

func TestTrendAnalyzer_LoadHistory_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, ".repodoctor", "history.json")

	// Create history file
	historyContent := `[
  {
    "timestamp": "2026-02-28T10:00:00Z",
    "score": 82.5
  },
  {
    "timestamp": "2026-03-01T10:00:00Z",
    "score": 85.0
  }
]`

	err := os.MkdirAll(filepath.Dir(historyPath), 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	err = os.WriteFile(historyPath, []byte(historyContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create history file: %v", err)
	}

	analyzer := NewTrendAnalyzer(tmpDir)
	err = analyzer.LoadHistory()
	if err != nil {
		t.Errorf("Expected no error loading history: %v", err)
	}

	if len(analyzer.history) != 2 {
		t.Errorf("Expected 2 history entries, got %d", len(analyzer.history))
	}

	if analyzer.history[0].Score != 82.5 {
		t.Errorf("Expected first score 82.5, got %.1f", analyzer.history[0].Score)
	}

	if analyzer.history[1].Score != 85.0 {
		t.Errorf("Expected second score 85.0, got %.1f", analyzer.history[1].Score)
	}
}

func TestTrendAnalyzer_EnsureConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".repodoctor")

	// Directory shouldn't exist yet
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatal("Expected config directory to not exist initially")
	}

	analyzer := NewTrendAnalyzer(tmpDir)
	analyzer.AppendScore(80.0)

	// Now it should exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Expected config directory to be created")
	}
}
