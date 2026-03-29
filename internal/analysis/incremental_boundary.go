package analysis

import "fmt"

// IncrementalSnapshotStore isolates incremental cache persistence from orchestration.
// Concrete storage implementations remain internal to the analysis layer.
type IncrementalSnapshotStore interface {
	Load(cacheKey string) (IncrementalCacheSnapshot, bool, error)
	Save(snapshot IncrementalCacheSnapshot) error
}

// IncrementalFingerprintProvider isolates fingerprint collection/diff logic.
type IncrementalFingerprintProvider interface {
	Build(repoPath string) (map[string]string, error)
	Diff(previous, current map[string]string) (changed []string, removed []string)
}

type GoIncrementalFingerprintProvider struct{}

func (p GoIncrementalFingerprintProvider) Build(repoPath string) (map[string]string, error) {
	return BuildGoFingerprintMap(repoPath)
}

func (p GoIncrementalFingerprintProvider) Diff(previous, current map[string]string) ([]string, []string) {
	return DiffFingerprintStates(previous, current)
}

type IncrementalDelta struct {
	Snapshot    IncrementalCacheSnapshot
	Changed     []string
	Removed     []string
	CacheHit    bool
	PreviousKey string
}

// IncrementalBoundaryService computes incremental deltas without leaking
// storage implementation details outside internal/analysis.
type IncrementalBoundaryService struct {
	store    IncrementalSnapshotStore
	provider IncrementalFingerprintProvider
}

func NewIncrementalBoundaryService(store IncrementalSnapshotStore, provider IncrementalFingerprintProvider) (*IncrementalBoundaryService, error) {
	if store == nil {
		return nil, fmt.Errorf("incremental snapshot store is required")
	}
	if provider == nil {
		return nil, fmt.Errorf("incremental fingerprint provider is required")
	}
	return &IncrementalBoundaryService{store: store, provider: provider}, nil
}

func (s *IncrementalBoundaryService) Compute(repoPath, cacheKey string) (IncrementalDelta, error) {
	if s == nil {
		return IncrementalDelta{}, fmt.Errorf("incremental boundary service is required")
	}
	if cacheKey == "" {
		return IncrementalDelta{}, fmt.Errorf("cache key is required")
	}

	fingerprints, err := s.provider.Build(repoPath)
	if err != nil {
		return IncrementalDelta{}, fmt.Errorf("build fingerprints: %w", err)
	}
	currentSnapshot := NewIncrementalCacheSnapshot(cacheKey, fingerprints)

	previous, found, err := s.store.Load(cacheKey)
	if err != nil {
		return IncrementalDelta{}, fmt.Errorf("load incremental snapshot: %w", err)
	}

	changed := make([]string, 0)
	removed := make([]string, 0)
	previousKey := ""
	if found {
		changed, removed = s.provider.Diff(previous.Fingerprints, currentSnapshot.Fingerprints)
		previousKey = previous.CacheKey
	}

	if err := s.store.Save(currentSnapshot); err != nil {
		return IncrementalDelta{}, fmt.Errorf("save incremental snapshot: %w", err)
	}

	return IncrementalDelta{
		Snapshot:    currentSnapshot,
		Changed:     append([]string(nil), changed...),
		Removed:     append([]string(nil), removed...),
		CacheHit:    found,
		PreviousKey: previousKey,
	}, nil
}
