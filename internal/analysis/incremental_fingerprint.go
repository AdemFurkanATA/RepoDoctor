package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileFingerprint struct {
	Path string
	Hash string
}

type IncrementalFingerprintState struct {
	Files map[string]string
}

func BuildGoFingerprintMap(repoPath string) (map[string]string, error) {
	state := make(map[string]string)
	err := filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".go" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		hash := sha256.Sum256(data)
		state[filepath.Clean(path)] = hex.EncodeToString(hash[:])
		return nil
	})
	if err != nil {
		return nil, err
	}
	return state, nil
}

func DiffFingerprintStates(previous, current map[string]string) (changed, removed []string) {
	changedSet := make(map[string]struct{})
	removedSet := make(map[string]struct{})

	for path, oldHash := range previous {
		newHash, ok := current[path]
		if !ok {
			removedSet[path] = struct{}{}
			continue
		}
		if newHash != oldHash {
			changedSet[path] = struct{}{}
		}
	}

	for path := range current {
		if _, existed := previous[path]; !existed {
			changedSet[path] = struct{}{}
		}
	}

	changed = setToSortedSlice(changedSet)
	removed = setToSortedSlice(removedSet)
	return changed, removed
}

func setToSortedSlice(input map[string]struct{}) []string {
	out := make([]string, 0, len(input))
	for key := range input {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func MarshalFingerprintState(state map[string]string) IncrementalFingerprintState {
	cloned := make(map[string]string, len(state))
	for path, hash := range state {
		cloned[path] = hash
	}
	return IncrementalFingerprintState{Files: cloned}
}

func ValidateFingerprintState(state IncrementalFingerprintState) error {
	if state.Files == nil {
		return fmt.Errorf("fingerprint state files map cannot be nil")
	}
	for path, hash := range state.Files {
		if strings.TrimSpace(path) == "" || strings.TrimSpace(hash) == "" {
			return fmt.Errorf("fingerprint state contains empty path/hash entry")
		}
	}
	return nil
}
