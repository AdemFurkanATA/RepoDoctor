package domain

// IgnoreStrategy defines the interface for determining if a directory should be ignored
// during directory traversal.
type IgnoreStrategy interface {
	ShouldIgnore(dirName string) bool
}

// DefaultIgnoreStrategy is a high-performance implementation of IgnoreStrategy
// using an O(1) lookup map for common skipped directories.
type DefaultIgnoreStrategy struct {
	ignoredDirs map[string]struct{}
}

// NewDefaultIgnoreStrategy creates a new DefaultIgnoreStrategy with the provided
// list of directory names to ignore.
func NewDefaultIgnoreStrategy(dirs []string) *DefaultIgnoreStrategy {
	m := make(map[string]struct{}, len(dirs))
	for _, d := range dirs {
		m[d] = struct{}{}
	}
	return &DefaultIgnoreStrategy{
		ignoredDirs: m,
	}
}

// ShouldIgnore returns true if the directory name is in the ignored list.
func (s *DefaultIgnoreStrategy) ShouldIgnore(dirName string) bool {
	_, exists := s.ignoredDirs[dirName]
	return exists
}

// DefaultIgnoredDirs contains the default list of directories to ignore
// across all repository types to prevent false positives and performance issues.
var DefaultIgnoredDirs = []string{
	"venv", "env", ".env", "__pycache__", ".pytest_cache", ".tox",
	"node_modules", "bower_components", ".npm",
	".git", ".svn", ".hg",
	"vendor",
	"bin", "obj", "out", "build", "dist", "target",
	".idea", ".vscode", ".vs",
}
