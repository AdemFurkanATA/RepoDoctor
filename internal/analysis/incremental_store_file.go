package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileIncrementalSnapshotStore persists incremental snapshots on disk.
// It is fail-closed for malformed data (treated as cache miss).
type FileIncrementalSnapshotStore struct {
	baseDir string
}

func NewFileIncrementalSnapshotStore(baseDir string) (*FileIncrementalSnapshotStore, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, fmt.Errorf("base directory is required")
	}
	cleanBase := filepath.Clean(baseDir)
	if err := os.MkdirAll(cleanBase, 0o755); err != nil {
		return nil, fmt.Errorf("create incremental cache directory: %w", err)
	}
	return &FileIncrementalSnapshotStore{baseDir: cleanBase}, nil
}

func (s *FileIncrementalSnapshotStore) Load(cacheKey string) (IncrementalCacheSnapshot, bool, error) {
	if s == nil {
		return IncrementalCacheSnapshot{}, false, fmt.Errorf("snapshot store is required")
	}
	if strings.TrimSpace(cacheKey) == "" {
		return IncrementalCacheSnapshot{}, false, fmt.Errorf("cache key is required")
	}

	path, err := s.snapshotPath(cacheKey)
	if err != nil {
		return IncrementalCacheSnapshot{}, false, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return IncrementalCacheSnapshot{}, false, nil
		}
		return IncrementalCacheSnapshot{}, false, fmt.Errorf("read incremental snapshot: %w", err)
	}

	snapshot, ok := UnmarshalIncrementalCacheSnapshotSafe(data)
	if !ok {
		return IncrementalCacheSnapshot{}, false, nil
	}

	return cloneSnapshot(snapshot), true, nil
}

func (s *FileIncrementalSnapshotStore) Save(snapshot IncrementalCacheSnapshot) error {
	if s == nil {
		return fmt.Errorf("snapshot store is required")
	}
	if err := snapshot.Validate(); err != nil {
		return err
	}

	path, err := s.snapshotPath(snapshot.CacheKey)
	if err != nil {
		return err
	}

	data, err := MarshalIncrementalCacheSnapshot(snapshot)
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write temporary incremental snapshot: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace incremental snapshot atomically: %w", err)
	}

	return nil
}

func (s *FileIncrementalSnapshotStore) snapshotPath(cacheKey string) (string, error) {
	normalized := strings.TrimSpace(cacheKey)
	if normalized == "" {
		return "", fmt.Errorf("cache key is required")
	}
	if len(normalized) != 64 || !isLowerHex(normalized) {
		return "", fmt.Errorf("cache key must be 64-char lowercase hex")
	}
	return filepath.Join(s.baseDir, normalized+".json"), nil
}
