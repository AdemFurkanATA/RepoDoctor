package analysis

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestIncrementalCorrectness_ColdAndIncrementalFingerprintStateParity(t *testing.T) {
	repo := t.TempDir()
	goFile := filepath.Join(repo, "main.go")
	if err := os.WriteFile(goFile, []byte("package main\nfunc main(){}\n"), 0o644); err != nil {
		t.Fatalf("failed writing fixture: %v", err)
	}

	key, err := BuildIncrementalCacheKey(IncrementalCacheKeyInput{
		AnalyzerVersion:   "0.14.0-dev",
		RepositoryPath:    repo,
		ConfigFingerprint: HashConfigBytes([]byte("rules:default")),
	})
	if err != nil {
		t.Fatalf("BuildIncrementalCacheKey failed: %v", err)
	}

	coldState, err := BuildGoFingerprintMap(repo)
	if err != nil {
		t.Fatalf("BuildGoFingerprintMap cold failed: %v", err)
	}
	incrementalState, err := BuildGoFingerprintMap(repo)
	if err != nil {
		t.Fatalf("BuildGoFingerprintMap incremental failed: %v", err)
	}

	if len(coldState) != len(incrementalState) {
		t.Fatalf("state size mismatch cold=%d incremental=%d", len(coldState), len(incrementalState))
	}
	for path, coldHash := range coldState {
		if incrementalState[path] != coldHash {
			t.Fatalf("hash mismatch for path %s cold=%s incremental=%s", path, coldHash, incrementalState[path])
		}
	}

	if bytes.Equal([]byte(key), []byte("")) {
		t.Fatal("cache key must be non-empty")
	}
}

func TestIncrementalCorrectness_DetectsChangeAndPreservesUnchanged(t *testing.T) {
	repo := t.TempDir()
	goA := filepath.Join(repo, "a.go")
	goB := filepath.Join(repo, "b.go")
	if err := os.WriteFile(goA, []byte("package main\nfunc a(){}\n"), 0o644); err != nil {
		t.Fatalf("failed writing a.go: %v", err)
	}
	if err := os.WriteFile(goB, []byte("package main\nfunc b(){}\n"), 0o644); err != nil {
		t.Fatalf("failed writing b.go: %v", err)
	}

	before, err := BuildGoFingerprintMap(repo)
	if err != nil {
		t.Fatalf("BuildGoFingerprintMap before failed: %v", err)
	}

	if err := os.WriteFile(goB, []byte("package main\nfunc b(){println(1)}\n"), 0o644); err != nil {
		t.Fatalf("failed mutating b.go: %v", err)
	}

	after, err := BuildGoFingerprintMap(repo)
	if err != nil {
		t.Fatalf("BuildGoFingerprintMap after failed: %v", err)
	}

	changed, removed := DiffFingerprintStates(before, after)
	if len(removed) != 0 {
		t.Fatalf("expected no removed files, got %v", removed)
	}
	if len(changed) != 1 {
		t.Fatalf("expected exactly one changed file, got %v", changed)
	}
	if filepath.Clean(changed[0]) != filepath.Clean(goB) {
		t.Fatalf("expected changed file %s, got %s", goB, changed[0])
	}
}
