package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// HistoryEntry represents a single historical score entry
type HistoryEntry struct {
	Timestamp string  `json:"timestamp"`
	Score     float64 `json:"score"`
}

// TrendAnalyzer handles historical score tracking and trend analysis
type TrendAnalyzer struct {
	historyPath string
	history     []HistoryEntry
}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer(baseDir string) *TrendAnalyzer {
	historyPath := filepath.Join(baseDir, ".repodoctor", "history.json")
	return &TrendAnalyzer{
		historyPath: historyPath,
		history:     make([]HistoryEntry, 0),
	}
}

// LoadHistory loads the score history from file
func (t *TrendAnalyzer) LoadHistory() error {
	// Check if file exists
	if _, err := os.Stat(t.historyPath); os.IsNotExist(err) {
		// No history file yet, start fresh
		t.history = make([]HistoryEntry, 0)
		return nil
	}

	// Read history file
	data, err := os.ReadFile(t.historyPath)
	if err != nil {
		return fmt.Errorf("failed to read history file: %w", err)
	}

	// Parse JSON
	var history []HistoryEntry
	if err := json.Unmarshal(data, &history); err != nil {
		// Malformed file, start fresh
		t.history = make([]HistoryEntry, 0)
		return nil
	}

	t.history = history
	return nil
}

// AppendScore appends a new score entry to the history
func (t *TrendAnalyzer) AppendScore(score float64) error {
	// Ensure directory exists
	configDir := filepath.Dir(t.historyPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	// Create new entry
	entry := HistoryEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Score:     score,
	}

	// Append to history
	t.history = append(t.history, entry)

	// Write to file
	return t.saveHistory()
}

// saveHistory writes the history to disk
func (t *TrendAnalyzer) saveHistory() error {
	data, err := json.MarshalIndent(t.history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(t.historyPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// GetPreviousScore returns the previous score (if available)
func (t *TrendAnalyzer) GetPreviousScore() (float64, bool) {
	if len(t.history) < 2 {
		return 0, false
	}
	// Return second to last score (last is the one we're about to add)
	return t.history[len(t.history)-1].Score, true
}

// GetLastEntry returns the last history entry
func (t *TrendAnalyzer) GetLastEntry() (*HistoryEntry, bool) {
	if len(t.history) == 0 {
		return nil, false
	}
	return &t.history[len(t.history)-1], true
}

// CalculateDelta calculates the score delta from the previous run
func (t *TrendAnalyzer) CalculateDelta(currentScore float64) (delta float64, trend string, hasPrevious bool) {
	prevScore, ok := t.GetPreviousScore()
	if !ok {
		return 0, "no previous data", false
	}

	delta = currentScore - prevScore

	if delta > 0 {
		trend = "increased"
	} else if delta < 0 {
		trend = "decreased"
	} else {
		trend = "unchanged"
	}

	return delta, trend, true
}

// GetTrendSummary returns a human-readable trend summary
func (t *TrendAnalyzer) GetTrendSummary(currentScore float64) string {
	delta, trend, hasPrevious := t.CalculateDelta(currentScore)

	if !hasPrevious {
		return "Current Score: No previous data available"
	}

	prevScore, _ := t.GetPreviousScore()

	summary := fmt.Sprintf("Current Score: %.1f\n", currentScore)
	summary += fmt.Sprintf("Previous Score: %.1f\n", prevScore)
	summary += fmt.Sprintf("Delta: %+.1f (%s)", delta, trend)

	return summary
}

// GetHistoryLength returns the number of entries in history
func (t *TrendAnalyzer) GetHistoryLength() int {
	return len(t.history)
}

// GetAllHistory returns the complete history
func (t *TrendAnalyzer) GetAllHistory() []HistoryEntry {
	return t.history
}
