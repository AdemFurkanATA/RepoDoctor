package analysis

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"RepoDoctor/internal/model"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	defaultGitChurnCommitLimit  = 1000
	defaultGitChurnRecentWindow = 50
	gitChurnCommitLimitEnv      = "REPODOCTOR_GIT_CHURN_COMMIT_LIMIT"
	gitChurnRecentWindowEnv     = "REPODOCTOR_GIT_CHURN_RECENT_WINDOW"
)

type churnAccumulator struct {
	entry   model.GitChurnFile
	authors map[string]bool
}

// ComputeGitChurnSummary extracts deterministic churn metrics from local git history.
// The function is fail-soft and returns warnings when churn data cannot be collected.
func ComputeGitChurnSummary(repoPath string) (model.GitChurnSummary, []string) {
	result := model.GitChurnSummary{Files: []model.GitChurnFile{}}

	repository, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return result, []string{"git churn skipped: repository metadata not available"}
	}

	head, err := repository.Head()
	if err != nil {
		return result, []string{"git churn skipped: HEAD not available"}
	}

	commitLimit := resolvePositiveIntEnv(gitChurnCommitLimitEnv, defaultGitChurnCommitLimit)
	recentWindow := resolvePositiveIntEnv(gitChurnRecentWindowEnv, defaultGitChurnRecentWindow)

	iter, err := repository.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return result, []string{"git churn skipped: commit log unavailable"}
	}
	defer iter.Close()

	accumulator := map[string]*churnAccumulator{}
	distinctAuthors := map[string]bool{}
	warnings := make([]string, 0)
	processed := 0

	_ = iter.ForEach(func(commit *object.Commit) error {
		if commit == nil {
			return nil
		}
		if processed >= commitLimit {
			return nil
		}

		processed++
		authorID := normalizeAuthorID(commit.Author)
		distinctAuthors[authorID] = true

		stats, statsErr := commit.Stats()
		if statsErr != nil {
			warnings = append(warnings, "git churn stats skipped for commit "+commit.Hash.String()+": "+statsErr.Error())
			return nil
		}

		sort.SliceStable(stats, func(i, j int) bool {
			return stats[i].Name < stats[j].Name
		})

		for _, stat := range stats {
			path := normalizeChurnPath(stat.Name)
			if path == "" {
				continue
			}
			entry := accumulator[path]
			if entry == nil {
				entry = &churnAccumulator{entry: model.GitChurnFile{Path: path}, authors: map[string]bool{}}
				accumulator[path] = entry
			}

			entry.entry.CommitTouches++
			entry.entry.AddedLines += stat.Addition
			entry.entry.DeletedLines += stat.Deletion
			if processed <= recentWindow {
				entry.entry.RecentTouches++
			}
			entry.authors[authorID] = true
		}

		return nil
	})

	result.TotalCommits = processed
	if processed < recentWindow {
		result.RecentWindowCommits = processed
	} else {
		result.RecentWindowCommits = recentWindow
	}
	result.DistinctAuthors = len(distinctAuthors)

	files := make([]model.GitChurnFile, 0, len(accumulator))
	for _, aggregated := range accumulator {
		aggregated.entry.UniqueAuthors = len(aggregated.authors)
		files = append(files, aggregated.entry)
	}

	sort.SliceStable(files, func(i, j int) bool {
		if files[i].CommitTouches != files[j].CommitTouches {
			return files[i].CommitTouches > files[j].CommitTouches
		}
		if files[i].RecentTouches != files[j].RecentTouches {
			return files[i].RecentTouches > files[j].RecentTouches
		}
		if files[i].UniqueAuthors != files[j].UniqueAuthors {
			return files[i].UniqueAuthors > files[j].UniqueAuthors
		}
		return files[i].Path < files[j].Path
	})

	result.Files = files
	return result, warnings
}

func resolvePositiveIntEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func normalizeAuthorID(signature object.Signature) string {
	if strings.TrimSpace(signature.Email) != "" {
		return strings.ToLower(strings.TrimSpace(signature.Email))
	}
	if strings.TrimSpace(signature.Name) != "" {
		return strings.ToLower(strings.TrimSpace(signature.Name))
	}
	return "unknown"
}

func normalizeChurnPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	cleaned := filepath.ToSlash(filepath.Clean(path))
	if cleaned == "." || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	return cleaned
}
