package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallManagedPreCommitHook_Idempotent(t *testing.T) {
	root := t.TempDir()
	hookDir := filepath.Join(root, ".git", "hooks")
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}

	if err := installManagedPreCommitHook(root); err != nil {
		t.Fatalf("first install failed: %v", err)
	}
	first, err := os.ReadFile(filepath.Join(hookDir, "pre-commit"))
	if err != nil {
		t.Fatalf("read hook after first install: %v", err)
	}

	if err := installManagedPreCommitHook(root); err != nil {
		t.Fatalf("second install should be idempotent, got error: %v", err)
	}
	second, err := os.ReadFile(filepath.Join(hookDir, "pre-commit"))
	if err != nil {
		t.Fatalf("read hook after second install: %v", err)
	}

	if string(first) != string(second) {
		t.Fatal("expected hook content to stay unchanged across repeated installs")
	}
	if !strings.Contains(string(second), preCommitMarker) {
		t.Fatalf("expected managed marker in hook: %s", string(second))
	}
}

func TestInstallManagedPreCommitHook_RejectsUnmanagedExistingHook(t *testing.T) {
	root := t.TempDir()
	hookDir := filepath.Join(root, ".git", "hooks")
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	hookPath := filepath.Join(hookDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho custom\n"), 0o755); err != nil {
		t.Fatalf("write unmanaged hook: %v", err)
	}

	err := installManagedPreCommitHook(root)
	if err == nil {
		t.Fatal("expected unmanaged hook installation to be rejected")
	}
}
