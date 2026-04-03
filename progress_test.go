package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProgressReporter_RenderAndLifecycle(t *testing.T) {
	reporter := NewProgressReporter(true)

	output := captureStdout(t, func() {
		reporter.Start("Scanning", 4)
		reporter.Update()
		reporter.SetProgress(3)
		reporter.Complete()
	})

	if !strings.Contains(output, "Scanning") {
		t.Fatalf("expected stage name in progress output, got %q", output)
	}
	if reporter.currentStep != reporter.totalSteps {
		t.Fatalf("expected reporter to complete all steps, current=%d total=%d", reporter.currentStep, reporter.totalSteps)
	}
	_ = reporter.GetElapsedTime()
}

func TestProgressReporter_RenderBarAndProgressBarHelpers(t *testing.T) {
	reporter := NewProgressReporter(true)
	bar := reporter.renderBar(50, 10)
	if len([]rune(bar)) != 10 {
		t.Fatalf("expected renderBar width 10, got %q", bar)
	}

	line := renderProgressBar("Rules", 1, 2)
	if !strings.Contains(line, "Rules") || !strings.Contains(line, "50%") {
		t.Fatalf("unexpected progress bar line: %q", line)
	}
}

func TestCountFilesAndStageCount(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "a.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to write go file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to write non-go file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".hidden"), 0o755); err != nil {
		t.Fatalf("failed creating hidden dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".hidden", "ignored.go"), []byte("package hidden\n"), 0o644); err != nil {
		t.Fatalf("failed writing hidden go file: %v", err)
	}

	count := countFiles(tmp)
	if count != 1 {
		t.Fatalf("expected one visible go file, got %d", count)
	}

	if got := getStageCount("Scanning repository", tmp); got != 1 {
		t.Fatalf("expected scanning stage count 1, got %d", got)
	}
	if got := getStageCount("Unknown", tmp); got != 10 {
		t.Fatalf("expected default stage count 10, got %d", got)
	}
}
