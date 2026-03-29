package analysis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileIncrementalSnapshotStore_RoundTrip(t *testing.T) {
	store, err := NewFileIncrementalSnapshotStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create file store: %v", err)
	}

	cacheKey := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	snapshot := NewIncrementalCacheSnapshot(cacheKey, map[string]string{
		"a.go": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	})

	if err := store.Save(snapshot); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, found, err := store.Load(cacheKey)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if !found {
		t.Fatal("expected snapshot to be found")
	}
	if loaded.CacheKey != cacheKey {
		t.Fatalf("expected cache key %s, got %s", cacheKey, loaded.CacheKey)
	}
	if loaded.Fingerprints["a.go"] != snapshot.Fingerprints["a.go"] {
		t.Fatalf("expected persisted fingerprint %s, got %s", snapshot.Fingerprints["a.go"], loaded.Fingerprints["a.go"])
	}
}

func TestFileIncrementalSnapshotStore_LoadCorruptPayloadFailsClosed(t *testing.T) {
	baseDir := t.TempDir()
	store, err := NewFileIncrementalSnapshotStore(baseDir)
	if err != nil {
		t.Fatalf("failed to create file store: %v", err)
	}

	cacheKey := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	corruptPath := filepath.Join(baseDir, cacheKey+".json")
	if err := os.WriteFile(corruptPath, []byte(`{"schemaVersion":"v1",`), 0o644); err != nil {
		t.Fatalf("failed writing corrupt payload: %v", err)
	}

	_, found, err := store.Load(cacheKey)
	if err != nil {
		t.Fatalf("expected corrupt payload to fail-closed without error, got: %v", err)
	}
	if found {
		t.Fatal("expected corrupt payload to be treated as cache miss")
	}
}

func TestFileIncrementalSnapshotStore_RejectsInvalidCacheKeyFormat(t *testing.T) {
	store, err := NewFileIncrementalSnapshotStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create file store: %v", err)
	}

	if _, _, err := store.Load("cache-key"); err == nil {
		t.Fatal("expected invalid cache key format to be rejected")
	}
}

func TestFileIncrementalSnapshotStore_SaveAtomicReplace(t *testing.T) {
	baseDir := t.TempDir()
	store, err := NewFileIncrementalSnapshotStore(baseDir)
	if err != nil {
		t.Fatalf("failed to create file store: %v", err)
	}

	cacheKey := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	first := NewIncrementalCacheSnapshot(cacheKey, map[string]string{"a.go": "1111111111111111111111111111111111111111111111111111111111111111"})
	second := NewIncrementalCacheSnapshot(cacheKey, map[string]string{"a.go": "2222222222222222222222222222222222222222222222222222222222222222"})

	if err := store.Save(first); err != nil {
		t.Fatalf("first save failed: %v", err)
	}
	if err := store.Save(second); err != nil {
		t.Fatalf("second save failed: %v", err)
	}

	loaded, found, err := store.Load(cacheKey)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if !found {
		t.Fatal("expected snapshot to be found")
	}
	if loaded.Fingerprints["a.go"] != second.Fingerprints["a.go"] {
		t.Fatalf("expected latest snapshot hash %s, got %s", second.Fingerprints["a.go"], loaded.Fingerprints["a.go"])
	}
}
