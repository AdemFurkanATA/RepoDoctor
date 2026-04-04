package analysis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileIncrementalSnapshotStore_WarmupFromFilesystem_FailSoft(t *testing.T) {
	baseDir := t.TempDir()
	store, err := NewFileIncrementalSnapshotStore(baseDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	validKey := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	valid := NewIncrementalCacheSnapshot(validKey, map[string]string{"a.go": "1111111111111111111111111111111111111111111111111111111111111111"})
	if err := store.Save(valid); err != nil {
		t.Fatalf("failed saving valid snapshot: %v", err)
	}

	if err := os.WriteFile(filepath.Join(baseDir, "not-cache.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("failed writing ignored file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed writing malformed json file: %v", err)
	}

	warmed, warnings, warmErr := store.WarmupFromFilesystem(10)
	if warmErr != nil {
		t.Fatalf("warmup should be fail-soft, got error: %v", warmErr)
	}
	if warmed != 1 {
		t.Fatalf("expected one warmed snapshot, got %d", warmed)
	}
	if len(warnings) == 0 {
		t.Fatal("expected warnings for malformed cache entries")
	}
}

func TestFileIncrementalSnapshotStore_WarmupFromFilesystem_RespectsLimit(t *testing.T) {
	baseDir := t.TempDir()
	store, err := NewFileIncrementalSnapshotStore(baseDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	keys := []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
	}
	for i, key := range keys {
		snapshot := NewIncrementalCacheSnapshot(key, map[string]string{"f.go": "1111111111111111111111111111111111111111111111111111111111111111"})
		snapshot.Fingerprints["f.go"] = string('1'+rune(i)) + "111111111111111111111111111111111111111111111111111111111111111"
		if err := store.Save(snapshot); err != nil {
			t.Fatalf("failed saving snapshot %d: %v", i, err)
		}
	}

	warmed, warnings, warmErr := store.WarmupFromFilesystem(2)
	if warmErr != nil {
		t.Fatalf("warmup failed: %v", warmErr)
	}
	if warmed != 2 {
		t.Fatalf("expected warmup limit 2, got %d", warmed)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for valid snapshots, got %v", warnings)
	}
}

func TestParseSnapshotFileName(t *testing.T) {
	key, ok := parseSnapshotFileName("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.json")
	if !ok || key == "" {
		t.Fatal("expected valid snapshot filename to parse")
	}

	if _, ok := parseSnapshotFileName("../escape.json"); ok {
		t.Fatal("expected invalid filename to be rejected")
	}
	if _, ok := parseSnapshotFileName("nothex.json"); ok {
		t.Fatal("expected non-hex key to be rejected")
	}
}
