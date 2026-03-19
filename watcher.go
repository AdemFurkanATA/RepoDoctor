package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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
	if event.Op&fsnotify.Create == fsnotify.Create {
		w.addDirectoryIfNeeded(event.Name)
	}

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
	if code := runAnalyze(w.path, "text", false, true, false); code != 0 {
		fmt.Printf("Analysis finished with exit code %d (watch continues).\n", code)
	}
}

func (w *Watcher) addDirectoryIfNeeded(path string) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return
	}

	if err := filepath.Walk(path, func(current string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			if current == path {
				return filepath.SkipDir
			}
			return nil
		}
		switch info.Name() {
		case "node_modules", "vendor", "dist", "build", "bin", "obj":
			return filepath.SkipDir
		}
		_ = w.watcher.Add(current)
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Watcher add directory error: %v\n", err)
	}
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

// WatchAndAnalyze starts watch mode and runs initial analysis.
// It listens for OS signals (SIGINT, SIGTERM) for graceful shutdown.
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
	if code := runAnalyze(path, "text", false, true, false); code != 0 {
		fmt.Printf("Initial analysis finished with exit code %d (watch continues).\n", code)
	}

	// Start watching
	if err := watcher.Start(); err != nil {
		return err
	}

	// Graceful shutdown: listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n\nShutting down watch mode...")
	if err := watcher.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error stopping watcher: %v\n", err)
	}
	fmt.Println("Watch mode stopped. Goodbye!")
	return nil
}
