package analysis

import "testing"

func TestIncrementalCacheSnapshot_MarshalRoundTrip(t *testing.T) {
	s := NewIncrementalCacheSnapshot("cache-key", map[string]string{"a.go": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"})
	data, err := MarshalIncrementalCacheSnapshot(s)
	if err != nil {
		t.Fatalf("MarshalIncrementalCacheSnapshot failed: %v", err)
	}

	roundTrip, err := UnmarshalIncrementalCacheSnapshot(data)
	if err != nil {
		t.Fatalf("UnmarshalIncrementalCacheSnapshot failed: %v", err)
	}
	if roundTrip.SchemaVersion != IncrementalCacheSchemaVersion {
		t.Fatalf("expected schema version %s, got %s", IncrementalCacheSchemaVersion, roundTrip.SchemaVersion)
	}
	if roundTrip.CacheKey != "cache-key" {
		t.Fatalf("expected cache key cache-key, got %s", roundTrip.CacheKey)
	}
}

func TestIncrementalCacheSnapshot_RejectsUnsupportedVersion(t *testing.T) {
	data := []byte(`{"schemaVersion":"v0","cacheKey":"k","fingerprints":{"a.go":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}`)
	if _, err := UnmarshalIncrementalCacheSnapshot(data); err == nil {
		t.Fatal("expected unsupported version to fail")
	}
}

func TestIncrementalCacheSnapshot_RejectsNonHexHash(t *testing.T) {
	data := []byte(`{"schemaVersion":"v1","cacheKey":"k","fingerprints":{"a.go":"not-hex"}}`)
	if _, err := UnmarshalIncrementalCacheSnapshot(data); err == nil {
		t.Fatal("expected non-hex hash to fail")
	}
}

func TestUnmarshalIncrementalCacheSnapshotSafe_FailsClosedOnCorruptPayload(t *testing.T) {
	if _, ok := UnmarshalIncrementalCacheSnapshotSafe([]byte(`{"schemaVersion":"v1",`)); ok {
		t.Fatal("expected corrupt payload to fail closed")
	}

	if _, ok := UnmarshalIncrementalCacheSnapshotSafe([]byte(`{"schemaVersion":"v0","cacheKey":"k","fingerprints":{"a.go":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}`)); ok {
		t.Fatal("expected unsupported schema payload to fail closed")
	}
}
