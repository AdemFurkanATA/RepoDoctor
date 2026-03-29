package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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

	const fingerprintBatchSize = 128
	buffer := make([]byte, 32*1024)

	for start := 0; start < len(goFiles); start += fingerprintBatchSize {
		end := minInt(start+fingerprintBatchSize, len(goFiles))
		for _, path := range goFiles[start:end] {
			hash, readErr := hashFileSHA256StreamingWithBuffer(path, buffer)
			if readErr != nil {
				continue
			}
			state[path] = hash
		}
	}

	return state, nil
}

func hashFileSHA256Streaming(path string) (string, error) {
	return hashFileSHA256StreamingWithBuffer(path, make([]byte, 32*1024))
}

func hashFileSHA256StreamingWithBuffer(path string, buffer []byte) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if len(buffer) == 0 {
		buffer = make([]byte, 32*1024)
	}

	hasher := sha256.New()
	for {
		readBytes, readErr := file.Read(buffer)
		if readBytes > 0 {
			if _, writeErr := hasher.Write(buffer[:readBytes]); writeErr != nil {
				return "", writeErr
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", readErr
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
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
