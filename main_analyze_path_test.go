package main

import (
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

func TestResolveAnalyzePathArg_UsesLongPathFlagWhenProvided(t *testing.T) {
	resolved := resolveAnalyzePathArg([]string{"./a", "--path", "./b"}, "./b", []string{"./a"})
	if resolved != "./b" {
		t.Fatalf("expected --path to win, got %q", resolved)
	}
}

func TestResolveAnalyzePathArg_UsesLongPathEqualsSyntax(t *testing.T) {
	resolved := resolveAnalyzePathArg([]string{"./a", "--path=./b"}, "./b", []string{"./a"})
	if resolved != "./b" {
		t.Fatalf("expected --path= to win, got %q", resolved)
	}
}
