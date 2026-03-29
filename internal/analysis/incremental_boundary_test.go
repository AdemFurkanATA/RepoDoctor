package analysis

import (
	"errors"
	"testing"
)

type stubFingerprintProvider struct {
	state map[string]string
	err   error
}

func (s stubFingerprintProvider) Build(repoPath string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	result := make(map[string]string, len(s.state))
	for path, hash := range s.state {
		result[path] = hash
	}
	return result, nil
}

func (s stubFingerprintProvider) Diff(previous, current map[string]string) ([]string, []string) {
	return DiffFingerprintStates(previous, current)
}

func TestIncrementalBoundaryService_RequiresDependencies(t *testing.T) {
	_, err := NewIncrementalBoundaryService(nil, GoIncrementalFingerprintProvider{})
	if err == nil {
		t.Fatal("expected error for nil store")
	}

	store := NewInMemoryIncrementalSnapshotStore()
	_, err = NewIncrementalBoundaryService(store, nil)
	if err == nil {
		t.Fatal("expected error for nil provider")
	}
}

func TestIncrementalBoundaryService_FirstRunCacheMissThenHitWithDiff(t *testing.T) {
	store := NewInMemoryIncrementalSnapshotStore()
	service, err := NewIncrementalBoundaryService(store, stubFingerprintProvider{
		state: map[string]string{
			"a.go": "1111111111111111111111111111111111111111111111111111111111111111",
			"b.go": "2222222222222222222222222222222222222222222222222222222222222222",
		},
	})
	if err != nil {
		t.Fatalf("failed to create boundary service: %v", err)
	}

	first, err := service.Compute("/repo", "cache-key")
	if err != nil {
		t.Fatalf("first compute failed: %v", err)
	}
	if first.CacheHit {
		t.Fatal("first compute should be cache miss")
	}
	if len(first.Changed) != 0 || len(first.Removed) != 0 {
		t.Fatalf("first compute must not return stale diffs, changed=%v removed=%v", first.Changed, first.Removed)
	}

	service.provider = stubFingerprintProvider{state: map[string]string{
		"a.go": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"c.go": "3333333333333333333333333333333333333333333333333333333333333333",
	}}
	second, err := service.Compute("/repo", "cache-key")
	if err != nil {
		t.Fatalf("second compute failed: %v", err)
	}
	if !second.CacheHit {
		t.Fatal("second compute should be cache hit")
	}

	if len(second.Changed) != 2 {
		t.Fatalf("expected two changed entries (a.go + c.go), got %v", second.Changed)
	}
	got := map[string]bool{}
	for _, path := range second.Changed {
		got[path] = true
	}
	if !got["a.go"] || !got["c.go"] {
		t.Fatalf("unexpected changed list content: %v", second.Changed)
	}
	if len(second.Removed) != 1 || second.Removed[0] != "b.go" {
		t.Fatalf("unexpected removed list: %v", second.Removed)
	}
}

func TestIncrementalBoundaryService_PropagatesBuildErrors(t *testing.T) {
	store := NewInMemoryIncrementalSnapshotStore()
	service, err := NewIncrementalBoundaryService(store, stubFingerprintProvider{err: errors.New("boom")})
	if err != nil {
		t.Fatalf("failed to create boundary service: %v", err)
	}

	if _, err := service.Compute("/repo", "cache-key"); err == nil {
		t.Fatal("expected build error to be returned")
	}
}

func TestInMemoryIncrementalSnapshotStore_ClonesOnSaveAndLoad(t *testing.T) {
	store := NewInMemoryIncrementalSnapshotStore()
	original := NewIncrementalCacheSnapshot("cache-key", map[string]string{"a.go": "1111111111111111111111111111111111111111111111111111111111111111"})
	if err := store.Save(original); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	original.Fingerprints["a.go"] = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	loaded, found, err := store.Load("cache-key")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if !found {
		t.Fatal("expected stored snapshot to be found")
	}
	if loaded.Fingerprints["a.go"] != "1111111111111111111111111111111111111111111111111111111111111111" {
		t.Fatalf("store must clone on save, got %s", loaded.Fingerprints["a.go"])
	}

	loaded.Fingerprints["a.go"] = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	reloaded, found, err := store.Load("cache-key")
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if !found {
		t.Fatal("expected snapshot on reload")
	}
	if reloaded.Fingerprints["a.go"] != "1111111111111111111111111111111111111111111111111111111111111111" {
		t.Fatalf("store must clone on load, got %s", reloaded.Fingerprints["a.go"])
	}
}
