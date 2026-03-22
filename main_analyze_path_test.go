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

func TestComposeAnalyzeRequest_PathParityVariants(t *testing.T) {
	abs, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("failed to resolve abs path: %v", err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{name: "dot slash", args: []string{"-path", "./"}},
		{name: "dot backslash", args: []string{"-path", ".\\"}},
		{name: "absolute", args: []string{"-path", abs}},
	}

	var baseline *analyzeCommandRequest
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, composeErr := composeAnalyzeRequest(tc.args)
			if composeErr != nil {
				t.Fatalf("composeAnalyzeRequest failed: %v", composeErr)
			}
			if req.format != "text" {
				t.Fatalf("expected default format text, got %s", req.format)
			}
			if i == 0 {
				baseline = req
				return
			}
			if req.path != baseline.path {
				t.Fatalf("expected path parity, baseline=%q got=%q", baseline.path, req.path)
			}
			if req.format != baseline.format || req.verbose != baseline.verbose || req.colorEnabled != baseline.colorEnabled || req.watch != baseline.watch {
				t.Fatalf("expected request parity across path forms")
			}
		})
	}
}

func TestComposeAnalyzeRequest_AllowsParentPathWhenExists(t *testing.T) {
	req, err := composeAnalyzeRequest([]string{"-path", ".."})
	if err != nil {
		t.Fatalf("expected parent path to be allowed, got error: %v", err)
	}
	if req.path == "" {
		t.Fatal("expected normalized path to be non-empty")
	}
}

func TestComposeAnalyzeRequest_JSONFlagOverridesFormat(t *testing.T) {
	req, err := composeAnalyzeRequest([]string{"-format", "text", "-json"})
	if err != nil {
		t.Fatalf("composeAnalyzeRequest failed: %v", err)
	}
	if req.format != "json" {
		t.Fatalf("expected json flag to override format, got %s", req.format)
	}
}
