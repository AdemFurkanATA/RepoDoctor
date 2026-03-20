package languages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectPythonVersion attempts to detect Python version from the repository.
// Package-level helper to keep PythonAdapter method count within SRP bounds.
func DetectPythonVersion(repoPath string) (string, error) {
	// Check for .python-version file
	versionFile := filepath.Join(repoPath, ".python-version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	// Check for pyproject.toml
	pyprojectFile := filepath.Join(repoPath, "pyproject.toml")
	if content, err := os.ReadFile(pyprojectFile); err == nil {
		// Look for requires-python
		// This is a simplified check
		contentStr := string(content)
		if strings.Contains(contentStr, "requires-python") {
			return "from pyproject.toml", nil
		}
	}

	return "", fmt.Errorf("Python version not detected")
}
