package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/fsnotify.v1"
)

// Watcher handles filesystem watching for continuous analysis
type Watcher struct {
	watcher      *fsnotify.Watcher
	path         string
	debounceTime time.Duration
	lastChange   time.Time
	mu           sync.Mutex
	running      bool
	stopChan     chan struct{}
}

// NewWatcher creates a new filesystem watcher
func NewWatcher(path string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &Watcher{
		watcher:      watcher,
		path:         path,
		debounceTime: 500 * time.Millisecond,
		stopChan:     make(chan struct{}),
	}, nil
}

// Start begins watching for file changes
func (w *Watcher) Start() error {
	// Walk through directory and add all subdirectories to watcher
	err := filepath.Walk(w.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common non-source directories
		if info.IsDir() {
			switch info.Name() {
			case "node_modules", "vendor", "dist", "build", "bin", "obj":
				return filepath.SkipDir
			}
		}

		// Add directories to watcher
		if info.IsDir() {
			return w.watcher.Add(path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	w.running = true
	go w.watchLoop()

	return nil
}

// watchLoop monitors filesystem events
func (w *Watcher) watchLoop() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)

		case <-w.stopChan:
			return
		}
	}
}

// handleEvent processes a filesystem event
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Only process Go files
	if !strings.HasSuffix(event.Name, ".go") {
		return
	}

	// Skip hidden files
	if strings.HasPrefix(filepath.Base(event.Name), ".") {
		return
	}

	// Check if file is relevant (modification, creation, deletion)
	switch event.Op {
	case fsnotify.Write, fsnotify.Create, fsnotify.Remove, fsnotify.Rename:
		w.scheduleAnalysis(event)
	}
}

// scheduleAnalysis schedules an analysis run with debouncing
func (w *Watcher) scheduleAnalysis(event fsnotify.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	w.lastChange = now

	// Debounce - wait for changes to settle
	time.AfterFunc(w.debounceTime, func() {
		w.mu.Lock()
		lastChange := w.lastChange
		w.mu.Unlock()

		if lastChange.Equal(now) {
			// No new changes, run analysis
			w.runAnalysis(event.Name)
		}
	})
}

// runAnalysis executes the analysis
func (w *Watcher) runAnalysis(changedFile string) {
	fmt.Printf("\n\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("Change detected: %s\n", filepath.Base(changedFile))
	fmt.Println("Re-running analysis...")
	fmt.Println(strings.Repeat("=", 60))

	// Run analysis
	runAnalyze(w.path, "text", false, true)
}

// Stop stops the watcher
func (w *Watcher) Stop() error {
	close(w.stopChan)
	w.running = false
	return w.watcher.Close()
}

// IsRunning returns true if the watcher is active
func (w *Watcher) IsRunning() bool {
	return w.running
}

// WatchAndAnalyze starts watch mode and runs initial analysis
func WatchAndAnalyze(path string) error {
	watcher, err := NewWatcher(path)
	if err != nil {
		return err
	}

	fmt.Println("Watch Mode")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("Watching: %s\n", path)
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println()

	// Run initial analysis
	fmt.Println("Running initial analysis...")
	fmt.Println()
	runAnalyze(path, "text", false, true)

	// Start watching
	if err := watcher.Start(); err != nil {
		return err
	}

	// Keep running until interrupted
	select {}
}
