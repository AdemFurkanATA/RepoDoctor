package main

import (
	"fmt"
	"strings"
	"time"
)

// ProgressReporter handles progress tracking for long-running operations
type ProgressReporter struct {
	currentStage string
	totalSteps   int
	currentStep  int
	enabled      bool
	startTime    time.Time
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(enabled bool) *ProgressReporter {
	return &ProgressReporter{
		enabled:   enabled,
		startTime: time.Now(),
	}
}

// Start begins tracking a new stage
func (p *ProgressReporter) Start(stage string, totalSteps int) {
	if !p.enabled {
		return
	}
	p.currentStage = stage
	p.totalSteps = totalSteps
	p.currentStep = 0
	p.printProgress()
}

// Update advances the progress by one step
func (p *ProgressReporter) Update() {
	if !p.enabled {
		return
	}
	p.currentStep++
	p.printProgress()
}

// SetProgress sets the progress to a specific step
func (p *ProgressReporter) SetProgress(step int) {
	if !p.enabled {
		return
	}
	p.currentStep = step
	p.printProgress()
}

// Complete marks the current stage as complete
func (p *ProgressReporter) Complete() {
	if !p.enabled {
		return
	}
	p.currentStep = p.totalSteps
	p.printProgress()
	fmt.Println() // New line after completion
}

// printProgress displays the current progress
func (p *ProgressReporter) printProgress() {
	if p.totalSteps == 0 {
		fmt.Printf("\r%s ...", p.currentStage)
		return
	}

	percentage := float64(p.currentStep) / float64(p.totalSteps) * 100
	bar := p.renderBar(percentage, 20)
	
	fmt.Printf("\r%s [%s] %3.0f%%", p.currentStage, bar, percentage)
}

// renderBar creates a visual progress bar
func (p *ProgressReporter) renderBar(percentage float64, width int) string {
	filled := int(float64(width) * percentage / 100)
	empty := width - filled

	bar := strings.Repeat("█", filled)
	if filled < width {
		bar += "░"
		empty--
	}
	bar += strings.Repeat("░", empty)

	return bar
}

// GetElapsedTime returns the elapsed time since the reporter started
func (p *ProgressReporter) GetElapsedTime() time.Duration {
	return time.Since(p.startTime)
}

// getStageCount returns the number of steps for a given stage
func getStageCount(stage string, repoPath string) int {
	switch stage {
	case "Scanning repository":
		return countFiles(repoPath)
	case "Collecting metrics":
		return 10 // Approximate progress steps
	case "Building dependency graph":
		return 10 // Approximate progress steps
	case "Running rules":
		return 4 // One for each rule type
	}
	return 10
}

// countFiles estimates the number of Go files to scan
func countFiles(path string) int {
	// Simplified - actual implementation would count files
	// For now, return a reasonable default
	return 50
}

// renderProgressBar renders a single-line progress bar
func renderProgressBar(stage string, current, total int) string {
	percentage := float64(current) / float64(total) * 100
	width := 20
	
	filled := int(float64(width) * percentage / 100)
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	
	return fmt.Sprintf("%s [%s] %3.0f%%", stage, bar, percentage)
}
