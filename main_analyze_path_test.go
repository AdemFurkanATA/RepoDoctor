package main

import (
	"path/filepath"
	"testing"
)

func TestResolveAnalyzePathArg_UsesPositionalWhenNoPathFlag(t *testing.T) {
	resolved := resolveAnalyzePathArg([]string{"."}, ".", []string{"."})
	if resolved != "." {
		t.Fatalf("expected positional path, got %q", resolved)
	}
}

func TestResolveAnalyzePathArg_UsesPathFlagWhenProvided(t *testing.T) {
	resolved := resolveAnalyzePathArg([]string{"./a", "-path", "./b"}, "./b", []string{"./a"})
	if resolved != "./b" {
		t.Fatalf("expected -path to win, got %q", resolved)
	}
}

func TestResolveAnalyzePathArg_UsesPathEqualsSyntax(t *testing.T) {
	resolved := resolveAnalyzePathArg([]string{"./a", "-path=./b"}, "./b", []string{"./a"})
	if resolved != "./b" {
		t.Fatalf("expected -path= to win, got %q", resolved)
	}
}

func TestIsWithinRoot(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "sub")
	outsideBase := t.TempDir()
	outside := filepath.Join(outsideBase, "other")

	if !isWithinRoot(root, inside) {
		t.Fatalf("expected sub path to be within root")
	}

	if isWithinRoot(root, outside) {
		t.Fatalf("expected outside path to be rejected")
	}
}
