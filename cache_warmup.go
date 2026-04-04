package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	analysispkg "RepoDoctor/internal/analysis"
)

const (
	cacheWarmupEnv      = "REPODOCTOR_CACHE_WARMUP"
	cacheWarmupLimitEnv = "REPODOCTOR_CACHE_WARMUP_LIMIT"
)

func maybeWarmupIncrementalCache(absPath string) ([]string, error) {
	if strings.TrimSpace(os.Getenv(cacheWarmupEnv)) != "1" {
		return nil, nil
	}

	cacheDir := filepath.Join(absPath, ".repodoctor", "cache")
	if _, err := os.Stat(cacheDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	store, err := analysispkg.NewFileIncrementalSnapshotStore(cacheDir)
	if err != nil {
		return nil, err
	}

	limit := 128
	if raw := strings.TrimSpace(os.Getenv(cacheWarmupLimitEnv)); raw != "" {
		if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}

	warmed, warnings, err := store.WarmupFromFilesystem(limit)
	if err != nil {
		return nil, err
	}

	info := fmt.Sprintf("cache warmup loaded %d snapshot(s)", warmed)
	return append([]string{info}, warnings...), nil
}
