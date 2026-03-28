package analysis

import (
	"encoding/json"
	"fmt"
)

const IncrementalCacheSchemaVersion = "v1"

type IncrementalCacheSnapshot struct {
	SchemaVersion string            `json:"schemaVersion"`
	CacheKey      string            `json:"cacheKey"`
	Fingerprints  map[string]string `json:"fingerprints"`
}

func NewIncrementalCacheSnapshot(cacheKey string, fingerprints map[string]string) IncrementalCacheSnapshot {
	cloned := make(map[string]string, len(fingerprints))
	for path, hash := range fingerprints {
		cloned[path] = hash
	}
	return IncrementalCacheSnapshot{
		SchemaVersion: IncrementalCacheSchemaVersion,
		CacheKey:      cacheKey,
		Fingerprints:  cloned,
	}
}

func (s IncrementalCacheSnapshot) Validate() error {
	if s.SchemaVersion != IncrementalCacheSchemaVersion {
		return fmt.Errorf("unsupported incremental cache schema version: %s", s.SchemaVersion)
	}
	if s.CacheKey == "" {
		return fmt.Errorf("incremental cache snapshot cacheKey cannot be empty")
	}
	if s.Fingerprints == nil {
		return fmt.Errorf("incremental cache snapshot fingerprints cannot be nil")
	}
	return nil
}

func MarshalIncrementalCacheSnapshot(snapshot IncrementalCacheSnapshot) ([]byte, error) {
	if err := snapshot.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(snapshot)
}

func UnmarshalIncrementalCacheSnapshot(data []byte) (IncrementalCacheSnapshot, error) {
	var snapshot IncrementalCacheSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return IncrementalCacheSnapshot{}, err
	}
	if err := snapshot.Validate(); err != nil {
		return IncrementalCacheSnapshot{}, err
	}
	return snapshot, nil
}
