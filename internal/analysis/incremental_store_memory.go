package analysis

import (
	"fmt"
	"sync"
)

// InMemoryIncrementalSnapshotStore is a deterministic test-friendly store
// implementation kept within analysis boundaries.
type InMemoryIncrementalSnapshotStore struct {
	snapshots map[string]IncrementalCacheSnapshot
	mu        sync.RWMutex
}

func NewInMemoryIncrementalSnapshotStore() *InMemoryIncrementalSnapshotStore {
	return &InMemoryIncrementalSnapshotStore{
		snapshots: make(map[string]IncrementalCacheSnapshot),
	}
}

func (s *InMemoryIncrementalSnapshotStore) Load(cacheKey string) (IncrementalCacheSnapshot, bool, error) {
	if s == nil {
		return IncrementalCacheSnapshot{}, false, fmt.Errorf("snapshot store is required")
	}
	if cacheKey == "" {
		return IncrementalCacheSnapshot{}, false, fmt.Errorf("cache key is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot, ok := s.snapshots[cacheKey]
	if !ok {
		return IncrementalCacheSnapshot{}, false, nil
	}
	return cloneSnapshot(snapshot), true, nil
}

func (s *InMemoryIncrementalSnapshotStore) Save(snapshot IncrementalCacheSnapshot) error {
	if s == nil {
		return fmt.Errorf("snapshot store is required")
	}
	if err := snapshot.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshots[snapshot.CacheKey] = cloneSnapshot(snapshot)
	return nil
}

func cloneSnapshot(snapshot IncrementalCacheSnapshot) IncrementalCacheSnapshot {
	cloned := make(map[string]string, len(snapshot.Fingerprints))
	for path, hash := range snapshot.Fingerprints {
		cloned[path] = hash
	}
	return IncrementalCacheSnapshot{
		SchemaVersion: snapshot.SchemaVersion,
		CacheKey:      snapshot.CacheKey,
		Fingerprints:  cloned,
	}
}
