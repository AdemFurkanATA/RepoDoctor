package analysis

import "testing"

func TestBuildIncrementalCacheKey_Deterministic(t *testing.T) {
	input := IncrementalCacheKeyInput{
		AnalyzerVersion:   "0.14.0-dev",
		RepositoryPath:    ".",
		ConfigFingerprint: HashConfigBytes([]byte("a: 1")),
	}

	first, err := BuildIncrementalCacheKey(input)
	if err != nil {
		t.Fatalf("BuildIncrementalCacheKey failed: %v", err)
	}
	for i := 0; i < 10; i++ {
		next, nextErr := BuildIncrementalCacheKey(input)
		if nextErr != nil {
			t.Fatalf("BuildIncrementalCacheKey failed on iteration %d: %v", i, nextErr)
		}
		if next != first {
			t.Fatalf("cache key must be deterministic, got %s vs %s", first, next)
		}
	}
}

func TestBuildIncrementalCacheKey_ChangesWithVersionOrConfig(t *testing.T) {
	base := IncrementalCacheKeyInput{AnalyzerVersion: "0.14.0-dev", RepositoryPath: ".", ConfigFingerprint: HashConfigBytes([]byte("a:1"))}
	keyA, err := BuildIncrementalCacheKey(base)
	if err != nil {
		t.Fatalf("BuildIncrementalCacheKey failed: %v", err)
	}

	variantVersion := base
	variantVersion.AnalyzerVersion = "0.14.1-dev"
	keyB, err := BuildIncrementalCacheKey(variantVersion)
	if err != nil {
		t.Fatalf("BuildIncrementalCacheKey failed: %v", err)
	}
	if keyA == keyB {
		t.Fatal("expected key to change when version changes")
	}

	variantConfig := base
	variantConfig.ConfigFingerprint = HashConfigBytes([]byte("a:2"))
	keyC, err := BuildIncrementalCacheKey(variantConfig)
	if err != nil {
		t.Fatalf("BuildIncrementalCacheKey failed: %v", err)
	}
	if keyA == keyC {
		t.Fatal("expected key to change when config hash changes")
	}
}
