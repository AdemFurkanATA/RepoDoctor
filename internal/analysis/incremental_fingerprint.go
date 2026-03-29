package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type FileFingerprint struct {
	Path string
	Hash string
}

type IncrementalFingerprintState struct {
	Files map[string]string
}

type RenameCandidate struct {
	From string
	To   string
}

type IncrementalDiffSummary struct {
	Added    []string
	Modified []string
	Removed  []string
	Renamed  []RenameCandidate
}

func BuildGoFingerprintMap(repoPath string) (map[string]string, error) {
	goFiles, err := collectGoFiles(repoPath)
	if err != nil {
		return nil, err
	}

	state := make(map[string]string, len(goFiles))
	if len(goFiles) == 0 {
		return state, nil
	}

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	if workers > len(goFiles) {
		workers = len(goFiles)
	}

	paths := make(chan string, len(goFiles))
	results := make(chan FileFingerprint, len(goFiles))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range paths {
				data, readErr := os.ReadFile(path)
				if readErr != nil {
					continue
				}
				hash := sha256.Sum256(data)
				results <- FileFingerprint{Path: path, Hash: hex.EncodeToString(hash[:])}
			}
		}()
	}

	for _, path := range goFiles {
		paths <- path
	}
	close(paths)
	wg.Wait()
	close(results)

	for fp := range results {
		state[fp.Path] = fp.Hash
	}

	return state, nil
}

func collectGoFiles(repoPath string) ([]string, error) {
	goFiles := make([]string, 0, 256)
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
		goFiles = append(goFiles, filepath.Clean(path))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(goFiles)
	return goFiles, nil
}

func DiffFingerprintStates(previous, current map[string]string) (changed, removed []string) {
	summary := DiffFingerprintStatesDetailed(previous, current)
	changed = append(append([]string(nil), summary.Added...), summary.Modified...)
	removed = append([]string(nil), summary.Removed...)
	return changed, removed
}

func DiffFingerprintStatesDetailed(previous, current map[string]string) IncrementalDiffSummary {
	added := make(map[string]struct{})
	modified := make(map[string]struct{})
	removed := make(map[string]struct{})

	for path, oldHash := range previous {
		newHash, ok := current[path]
		if !ok {
			removed[path] = struct{}{}
			continue
		}
		if newHash != oldHash {
			modified[path] = struct{}{}
		}
	}

	for path := range current {
		if _, existed := previous[path]; !existed {
			added[path] = struct{}{}
		}
	}

	renamed := detectRenameCandidates(previous, current, added, removed)

	return IncrementalDiffSummary{
		Added:    setToSortedSlice(added),
		Modified: setToSortedSlice(modified),
		Removed:  setToSortedSlice(removed),
		Renamed:  renamed,
	}
}

func detectRenameCandidates(previous, current map[string]string, addedSet, removedSet map[string]struct{}) []RenameCandidate {
	removedByHash := make(map[string][]string)
	for path := range removedSet {
		hash := previous[path]
		removedByHash[hash] = append(removedByHash[hash], path)
	}

	addedByHash := make(map[string][]string)
	for path := range addedSet {
		hash := current[path]
		addedByHash[hash] = append(addedByHash[hash], path)
	}

	renames := make([]RenameCandidate, 0)
	for hash, fromPaths := range removedByHash {
		toPaths, ok := addedByHash[hash]
		if !ok {
			continue
		}
		sort.Strings(fromPaths)
		sort.Strings(toPaths)
		pairs := minInt(len(fromPaths), len(toPaths))
		for i := 0; i < pairs; i++ {
			renames = append(renames, RenameCandidate{From: fromPaths[i], To: toPaths[i]})
		}
	}

	sort.SliceStable(renames, func(i, j int) bool {
		if renames[i].From != renames[j].From {
			return renames[i].From < renames[j].From
		}
		return renames[i].To < renames[j].To
	})
	return renames
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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
