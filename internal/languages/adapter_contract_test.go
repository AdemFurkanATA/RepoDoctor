package languages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoAdapter_CapabilitiesAndNormalizeImport(t *testing.T) {
	adapter := NewGoAdapter()
	caps := adapter.Capabilities()

	if !caps.SupportsDependencyGraph || !caps.SupportsMetrics {
		t.Fatal("expected Go adapter to support graph and metrics")
	}

	if got := adapter.NormalizeImport("  github.com/foo/bar "); got != "github.com/foo/bar" {
		t.Fatalf("unexpected normalized import: %q", got)
	}
}

func TestPythonAdapter_CapabilitiesAndNormalizeImport(t *testing.T) {
	adapter := NewPythonAdapter()
	caps := adapter.Capabilities()

	if !caps.SupportsDependencyGraph || !caps.SupportsMetrics {
		t.Fatal("expected Python adapter to support graph and metrics")
	}

	if got := adapter.NormalizeImport(" requests.sessions "); got != "requests" {
		t.Fatalf("unexpected normalized import: %q", got)
	}
}

func TestJavaScriptAdapter_CapabilitiesAndNormalizeImport(t *testing.T) {
	adapter := NewJavaScriptAdapter()
	caps := adapter.Capabilities()

	if !caps.SupportsDependencyGraph || !caps.SupportsMetrics {
		t.Fatal("expected JavaScript adapter to support graph and metrics")
	}

	if got := adapter.NormalizeImport(" @scope/pkg/utils "); got != "@scope/pkg" {
		t.Fatalf("unexpected normalized JS import: %q", got)
	}
}

func TestTypeScriptAdapter_CapabilitiesAndNormalizeImport(t *testing.T) {
	adapter := NewTypeScriptAdapter()
	caps := adapter.Capabilities()

	if !caps.SupportsDependencyGraph || !caps.SupportsMetrics {
		t.Fatal("expected TypeScript adapter to support graph and metrics")
	}

	if got := adapter.NormalizeImport(" node:fs "); got != "fs" {
		t.Fatalf("unexpected normalized TS import: %q", got)
	}
}

func TestAdapterContract_DeterministicDetectFilesAcrossRuns(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "a.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed writing go fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "b.py"), []byte("print('x')\n"), 0o644); err != nil {
		t.Fatalf("failed writing python fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "c.js"), []byte("import x from 'react'\n"), 0o644); err != nil {
		t.Fatalf("failed writing js fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "d.ts"), []byte("export type X = string\n"), 0o644); err != nil {
		t.Fatalf("failed writing ts fixture: %v", err)
	}

	adapters := []LanguageAdapter{NewGoAdapter(), NewPythonAdapter(), NewJavaScriptAdapter(), NewTypeScriptAdapter()}
	for _, adapter := range adapters {
		first, err := adapter.DetectFiles(repo)
		if err != nil {
			t.Fatalf("%s DetectFiles failed: %v", adapter.Name(), err)
		}
		for i := 0; i < 10; i++ {
			next, nextErr := adapter.DetectFiles(repo)
			if nextErr != nil {
				t.Fatalf("%s DetectFiles failed on iteration %d: %v", adapter.Name(), i, nextErr)
			}
			if strings.Join(first, "|") != strings.Join(next, "|") {
				t.Fatalf("%s DetectFiles not deterministic\nfirst=%v\nnext=%v", adapter.Name(), first, next)
			}
		}
	}
}
