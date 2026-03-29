package analysis

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
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
	for path, hash := range s.Fingerprints {
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf("incremental cache snapshot contains empty fingerprint path")
		}
		if filepath.Clean(path) != path {
			return fmt.Errorf("incremental cache snapshot path must be clean: %s", path)
		}
		if len(hash) != 64 || !isLowerHex(hash) {
			return fmt.Errorf("incremental cache snapshot hash must be 64-char lowercase hex")
		}
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

// UnmarshalIncrementalCacheSnapshotSafe never panics and returns ok=false
// for malformed, unsupported, or suspicious cache payloads.
func UnmarshalIncrementalCacheSnapshotSafe(data []byte) (snapshot IncrementalCacheSnapshot, ok bool) {
	snapshot, err := UnmarshalIncrementalCacheSnapshot(data)
	if err != nil {
		return IncrementalCacheSnapshot{}, false
	}
	return snapshot, true
}

func isLowerHex(value string) bool {
	for _, ch := range value {
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			return false
		}
	}
	return true
}
