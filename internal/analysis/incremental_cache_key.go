package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

type IncrementalCacheKeyInput struct {
	AnalyzerVersion   string
	CacheSchema       string
	RepositoryPath    string
	ConfigFingerprint string
	RuleFingerprint   string
}

func BuildIncrementalCacheKey(input IncrementalCacheKeyInput) (string, error) {
	version := strings.TrimSpace(input.AnalyzerVersion)
	if version == "" {
		return "", fmt.Errorf("analyzer version is required")
	}

	normalizedPath, err := normalizeCachePath(input.RepositoryPath)
	if err != nil {
		return "", err
	}

	configHash := strings.TrimSpace(input.ConfigFingerprint)
	if configHash == "" {
		configHash = HashConfigBytes(nil)
	}

	ruleHash := strings.TrimSpace(input.RuleFingerprint)
	if ruleHash == "" {
		ruleHash = HashConfigBytes(nil)
	}

	cacheSchema := strings.TrimSpace(input.CacheSchema)
	if cacheSchema == "" {
		cacheSchema = IncrementalCacheSchemaVersion
	}

	payload := strings.Join([]string{version, cacheSchema, normalizedPath, configHash, ruleHash}, "|")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:]), nil
}

func HashConfigBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func normalizeCachePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("repository path is required")
	}
	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("failed to normalize repository path: %w", err)
	}
	normalized := filepath.ToSlash(filepath.Clean(absPath))
	if runtime.GOOS == "windows" {
		normalized = strings.ToLower(normalized)
	}
	return normalized, nil
}
