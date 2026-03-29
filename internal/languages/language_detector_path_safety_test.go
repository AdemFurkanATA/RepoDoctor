package languages

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNormalizePathWithinRoot_RejectsAbsoluteOutsidePath(t *testing.T) {
	root := t.TempDir()
	normalizedRoot, err := normalizeRepoRoot(root)
	if err != nil {
		t.Fatalf("failed to normalize root: %v", err)
	}
	outside := t.TempDir()

	if normalized, ok := normalizePathWithinRoot(normalizedRoot, outside); ok {
		t.Fatalf("expected outside absolute path to be rejected, got %q", normalized)
	}
}

func TestNormalizePathWithinRoot_RejectsParentTraversalEscape(t *testing.T) {
	root := t.TempDir()
	normalizedRoot, err := normalizeRepoRoot(root)
	if err != nil {
		t.Fatalf("failed to normalize root: %v", err)
	}
	parent := filepath.Dir(root)
	escapeCandidate := filepath.Join(parent, "escape.go")

	if normalized, ok := normalizePathWithinRoot(normalizedRoot, escapeCandidate); ok {
		t.Fatalf("expected parent traversal escape to be rejected, got %q", normalized)
	}
}

func TestNormalizePathWithinRoot_AllowsPathInsideRoot(t *testing.T) {
	root := t.TempDir()
	normalizedRoot, err := normalizeRepoRoot(root)
	if err != nil {
		t.Fatalf("failed to normalize root: %v", err)
	}
	inside := filepath.Join(root, "pkg", "main.go")
	if err := os.MkdirAll(filepath.Dir(inside), 0o755); err != nil {
		t.Fatalf("failed creating inside dir: %v", err)
	}
	if err := os.WriteFile(inside, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed writing inside file: %v", err)
	}

	normalized, ok := normalizePathWithinRoot(normalizedRoot, inside)
	if !ok {
		t.Fatal("expected inside path to be allowed")
	}
	if rel, err := filepath.Rel(normalizedRoot, normalized); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		t.Fatalf("normalized inside path escaped root: %q (err=%v)", normalized, err)
	}
}

func TestNormalizePathWithinRoot_RejectsSymlinkOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation is not reliable in all Windows CI environments")
	}

	root := t.TempDir()
	normalizedRoot, rootErr := normalizeRepoRoot(root)
	if rootErr != nil {
		t.Fatalf("failed to normalize root: %v", rootErr)
	}
	outside := t.TempDir()
	target := filepath.Join(outside, "danger.py")
	if err := os.WriteFile(target, []byte("print('x')\n"), 0o644); err != nil {
		t.Fatalf("failed writing outside target: %v", err)
	}

	symlinkPath := filepath.Join(root, "linked_outside.py")
	if err := os.Symlink(target, symlinkPath); err != nil {
		t.Skipf("symlink not supported in this environment: %v", err)
	}

	if normalized, ok := normalizePathWithinRoot(normalizedRoot, symlinkPath); ok {
		t.Fatalf("expected outside symlink target to be rejected, got %q", normalized)
	}
}

func TestNormalizePathWithinRoot_SymlinkLoopFailsSoftWithoutEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink loop creation is not reliable in all Windows CI environments")
	}

	root := t.TempDir()
	normalizedRoot, rootErr := normalizeRepoRoot(root)
	if rootErr != nil {
		t.Fatalf("failed to normalize root: %v", rootErr)
	}
	loop := filepath.Join(root, "loop")
	if err := os.Symlink(loop, loop); err != nil {
		t.Skipf("symlink loop not supported in this environment: %v", err)
	}

	normalized, ok := normalizePathWithinRoot(normalizedRoot, loop)
	if !ok {
		t.Fatal("expected symlink loop candidate inside root to fail-soft as in-root path")
	}
	if rel, err := filepath.Rel(normalizedRoot, normalized); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		t.Fatalf("normalized path escaped root for symlink loop: %q (err=%v)", normalized, err)
	}
}
