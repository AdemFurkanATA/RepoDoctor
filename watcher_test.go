package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/fsnotify.v1"
)

func TestWatcher_NewStartStopAndIsRunning(t *testing.T) {
	tmp := t.TempDir()
	w, err := NewWatcher(tmp)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}

	if w.IsRunning() {
		t.Fatal("new watcher should not be running")
	}

	if err := w.Start(); err != nil {
		t.Fatalf("watcher start failed: %v", err)
	}
	if !w.IsRunning() {
		t.Fatal("watcher expected running after start")
	}

	if err := w.Stop(); err != nil {
		t.Fatalf("watcher stop failed: %v", err)
	}
	if w.IsRunning() {
		t.Fatal("watcher should not be running after stop")
	}
}

func TestWatcher_HandleEventNonGoAndCreateDirPaths(t *testing.T) {
	tmp := t.TempDir()
	childDir := filepath.Join(tmp, "child")
	if err := os.MkdirAll(childDir, 0o755); err != nil {
		t.Fatalf("failed creating child dir: %v", err)
	}

	w, err := NewWatcher(tmp)
	if err != nil {
		t.Fatalf("NewWatcher failed: %v", err)
	}
	defer func() { _ = w.Stop() }()

	// keep debounce from triggering runAnalyze during this test
	w.debounceTime = 5 * time.Second

	if err := w.Start(); err != nil {
		t.Fatalf("watcher start failed: %v", err)
	}

	// Non-go file should be ignored early.
	w.handleEvent(fsnotify.Event{Name: filepath.Join(tmp, "notes.txt"), Op: fsnotify.Write})

	// Create event for directory should traverse addDirectoryIfNeeded path.
	w.handleEvent(fsnotify.Event{Name: childDir, Op: fsnotify.Create})
}
