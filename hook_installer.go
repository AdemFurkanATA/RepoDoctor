package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const preCommitMarker = "# repodoctor-managed:pre-commit"

func handleInstallHookCommand(args []string) error {
	hookType := "pre-commit"
	repoPath := "."

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--type", "-type":
			if i+1 >= len(args) {
				return HandleCLIUsageError("Usage: repodoctor install-hook --type pre-commit [-path .]", nil)
			}
			hookType = strings.TrimSpace(args[i+1])
			i++
		case "--path", "-path":
			if i+1 >= len(args) {
				return HandleCLIUsageError("Usage: repodoctor install-hook --type pre-commit [-path .]", nil)
			}
			repoPath = strings.TrimSpace(args[i+1])
			i++
		default:
			return HandleCLIUsageError("Usage: repodoctor install-hook --type pre-commit [-path .]", nil)
		}
	}

	if hookType != "pre-commit" {
		return NewCLIError(ErrorInvalidArgument, "unsupported hook type: "+hookType, "Only --type pre-commit is supported", nil)
	}

	abs, err := normalizeAnalyzePathInput(repoPath)
	if err != nil {
		return err
	}

	if err := installManagedPreCommitHook(abs); err != nil {
		return err
	}

	fmt.Println("Pre-commit hook installed successfully.")
	return nil
}

func installManagedPreCommitHook(repoRoot string) error {
	hookPath := filepath.Join(repoRoot, ".git", "hooks", "pre-commit")
	hookDir := filepath.Dir(hookPath)
	if _, err := os.Stat(hookDir); err != nil {
		return NewCLIError(ErrorInvalidArgument, "git hooks directory not found", "Run inside a git repository with a valid .git/hooks directory", err)
	}

	wanted := managedPreCommitHookScript()
	current, readErr := os.ReadFile(hookPath)
	if readErr == nil {
		currentText := string(current)
		if currentText == wanted {
			return nil
		}
		if !strings.Contains(currentText, preCommitMarker) {
			return NewCLIError(ErrorCLIUsage, "existing pre-commit hook is not managed by RepoDoctor", "Backup the existing hook, then re-run installation", errors.New("unmanaged pre-commit hook"))
		}
	}

	return os.WriteFile(hookPath, []byte(wanted), 0o755)
}

func managedPreCommitHookScript() string {
	return "#!/bin/sh\n" +
		preCommitMarker + "\n" +
		"set -e\n\n" +
		"STAGED=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\\.(go|py|js|jsx|ts|tsx|java)$' || true)\n" +
		"if [ -z \"$STAGED\" ]; then\n" +
		"  exit 0\n" +
		"fi\n\n" +
		"if command -v repodoctor >/dev/null 2>&1; then\n" +
		"  repodoctor analyze -path . -no-color\n" +
		"else\n" +
		"  go run . analyze -path . -no-color\n" +
		"fi\n"
}
