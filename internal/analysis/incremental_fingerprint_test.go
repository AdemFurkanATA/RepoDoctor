package analysis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildGoFingerprintMap_OnlyGoFiles(t *testing.T) {
	repo := t.TempDir()
	goPath := filepath.Join(repo, "a.go")
	pyPath := filepath.Join(repo, "b.py")
	if err := os.WriteFile(goPath, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed writing go file: %v", err)
	}
	if err := os.WriteFile(pyPath, []byte("print('x')\n"), 0o644); err != nil {
		t.Fatalf("failed writing py file: %v", err)
	}

	state, err := BuildGoFingerprintMap(repo)
	if err != nil {
		t.Fatalf("BuildGoFingerprintMap failed: %v", err)
	}
	if len(state) != 1 {
		t.Fatalf("expected one go fingerprint entry, got %d", len(state))
	}
	if _, ok := state[filepath.Clean(goPath)]; !ok {
		t.Fatalf("expected fingerprint for go file path %s", goPath)
	}
}

func TestDiffFingerprintStates_DetectsChangedAddedRemoved(t *testing.T) {
	prev := map[string]string{"a.go": "h1", "b.go": "h2"}
	next := map[string]string{"a.go": "h1-new", "c.go": "h3"}

	changed, removed := DiffFingerprintStates(prev, next)
	if len(changed) != 2 {
		t.Fatalf("expected changed set size 2, got %d (%v)", len(changed), changed)
	}
	if len(removed) != 1 || removed[0] != "b.go" {
		t.Fatalf("expected removed [b.go], got %v", removed)
	}
}

func TestValidateFingerprintState(t *testing.T) {
	if err := ValidateFingerprintState(IncrementalFingerprintState{Files: nil}); err == nil {
		t.Fatal("expected nil files map validation failure")
	}
	if err := ValidateFingerprintState(IncrementalFingerprintState{Files: map[string]string{"": "h"}}); err == nil {
		t.Fatal("expected empty path validation failure")
	}
	if err := ValidateFingerprintState(IncrementalFingerprintState{Files: map[string]string{"a.go": "h"}}); err != nil {
		t.Fatalf("expected valid state, got %v", err)
	}
}
