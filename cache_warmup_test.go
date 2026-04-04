package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	analysispkg "RepoDoctor/internal/analysis"
)

func TestMaybeWarmupIncrementalCache_DefaultOff(t *testing.T) {
	warnings, err := maybeWarmupIncrementalCache(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warmup warnings when disabled, got %v", warnings)
	}
}

func TestMaybeWarmupIncrementalCache_FailSoftAndBounded(t *testing.T) {
	repo := t.TempDir()
	cacheDir := filepath.Join(repo, ".repodoctor", "cache")
	store, err := analysispkg.NewFileIncrementalSnapshotStore(cacheDir)
	if err != nil {
		t.Fatalf("failed creating cache store: %v", err)
	}

	valid := analysispkg.NewIncrementalCacheSnapshot("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", map[string]string{
		"a.go": "1111111111111111111111111111111111111111111111111111111111111111",
	})
	if err := store.Save(valid); err != nil {
		t.Fatalf("failed saving valid snapshot: %v", err)
	}

	if err := os.WriteFile(filepath.Join(cacheDir, "bad.json"), []byte("not-json"), 0o644); err != nil {
		t.Fatalf("failed writing malformed entry: %v", err)
	}

	t.Setenv(cacheWarmupEnv, "1")
	t.Setenv(cacheWarmupLimitEnv, "1")
	warnings, warmErr := maybeWarmupIncrementalCache(repo)
	if warmErr != nil {
		t.Fatalf("expected fail-soft warmup, got %v", warmErr)
	}
	if len(warnings) == 0 {
		t.Fatal("expected warmup summary output")
	}
	if !strings.Contains(warnings[0], "loaded") {
		t.Fatalf("expected summary warning, got %v", warnings)
	}
}
