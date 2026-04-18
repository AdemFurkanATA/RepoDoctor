package analysis

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestComputeGitChurnSummary_DeterministicMetrics(t *testing.T) {
	repoPath := t.TempDir()
	repository, err := git.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	worktree, err := repository.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	commit := func(file, content, authorName, authorEmail string, when time.Time) {
		absFile := filepath.Join(repoPath, file)
		if err := os.MkdirAll(filepath.Dir(absFile), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", file, err)
		}
		if err := os.WriteFile(absFile, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
		if _, err := worktree.Add(file); err != nil {
			t.Fatalf("add %s: %v", file, err)
		}
		_, err := worktree.Commit("update "+file, &git.CommitOptions{
			Author: &object.Signature{Name: authorName, Email: authorEmail, When: when},
		})
		if err != nil {
			t.Fatalf("commit %s: %v", file, err)
		}
	}

	base := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)
	commit("a.go", "package demo\n\nfunc A() {}\n", "Alice", "alice@example.com", base)
	commit("a.go", "package demo\n\nfunc A() { println(1) }\n", "Bob", "bob@example.com", base.Add(time.Minute))
	commit("b.go", "package demo\n\nfunc B() {}\n", "Alice", "alice@example.com", base.Add(2*time.Minute))

	summary, warnings := ComputeGitChurnSummary(repoPath)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if summary.TotalCommits != 3 {
		t.Fatalf("expected 3 commits, got %d", summary.TotalCommits)
	}
	if summary.DistinctAuthors != 2 {
		t.Fatalf("expected 2 distinct authors, got %d", summary.DistinctAuthors)
	}
	if len(summary.Files) != 2 {
		t.Fatalf("expected 2 touched files, got %d", len(summary.Files))
	}
	if summary.Files[0].Path != "a.go" {
		t.Fatalf("expected deterministic ordering by churn for a.go first, got %q", summary.Files[0].Path)
	}
	if summary.Files[0].CommitTouches != 2 {
		t.Fatalf("expected a.go touches 2, got %d", summary.Files[0].CommitTouches)
	}
	if summary.Files[0].UniqueAuthors != 2 {
		t.Fatalf("expected a.go unique authors 2, got %d", summary.Files[0].UniqueAuthors)
	}
}

func TestComputeGitChurnSummary_FailSoftWhenNoGitRepository(t *testing.T) {
	summary, warnings := ComputeGitChurnSummary(t.TempDir())
	if summary.TotalCommits != 0 {
		t.Fatalf("expected zero commits for non-git path, got %d", summary.TotalCommits)
	}
	if len(warnings) == 0 {
		t.Fatal("expected warning for non-git path")
	}
}
