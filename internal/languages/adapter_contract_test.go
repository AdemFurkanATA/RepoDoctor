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

func TestAdapterContract_DeterministicBuildDependencyGraphAcrossRuns(t *testing.T) {
	repo := t.TempDir()
	filesByAdapter := map[string]string{
		"Go":         filepath.Join(repo, "main.go"),
		"Python":     filepath.Join(repo, "main.py"),
		"JavaScript": filepath.Join(repo, "main.js"),
		"TypeScript": filepath.Join(repo, "main.ts"),
	}

	fixtures := map[string]string{
		"Go":         "package main\nimport (\"fmt\"; \"net/http\")\nfunc main(){fmt.Println(http.MethodGet)}\n",
		"Python":     "import os\nfrom collections import defaultdict\n",
		"JavaScript": "import x from 'react'\nconst y = require('lodash')\n",
		"TypeScript": "export {x} from '@scope/pkg/utils'\n",
	}

	for name, file := range filesByAdapter {
		if err := os.WriteFile(file, []byte(fixtures[name]), 0o644); err != nil {
			t.Fatalf("failed writing %s fixture: %v", name, err)
		}
	}

	adapters := []LanguageAdapter{NewGoAdapter(), NewPythonAdapter(), NewJavaScriptAdapter(), NewTypeScriptAdapter()}
	for _, adapter := range adapters {
		file := filesByAdapter[adapter.Name()]
		firstGraph, err := adapter.BuildDependencyGraph([]string{file})
		if err != nil {
			t.Fatalf("%s BuildDependencyGraph failed: %v", adapter.Name(), err)
		}
		firstNode := firstGraph.GetNode(file)
		if firstNode == nil {
			t.Fatalf("%s expected node for %s", adapter.Name(), file)
		}
		baseline := strings.Join(firstNode.Imports, "|")

		for i := 0; i < 10; i++ {
			nextGraph, nextErr := adapter.BuildDependencyGraph([]string{file})
			if nextErr != nil {
				t.Fatalf("%s BuildDependencyGraph failed on iteration %d: %v", adapter.Name(), i, nextErr)
			}
			nextNode := nextGraph.GetNode(file)
			if nextNode == nil {
				t.Fatalf("%s expected node for %s on iteration %d", adapter.Name(), file, i)
			}
			current := strings.Join(nextNode.Imports, "|")
			if baseline != current {
				t.Fatalf("%s BuildDependencyGraph not deterministic\nfirst=%v\nnext=%v", adapter.Name(), baseline, current)
			}
		}
	}
}

func TestAdapterContract_MalformedCorpusFailSoftNoPanic(t *testing.T) {
	repo := t.TempDir()
	filesByAdapter := map[string]string{
		"Go":         filepath.Join(repo, "broken.go"),
		"Python":     filepath.Join(repo, "broken.py"),
		"JavaScript": filepath.Join(repo, "broken.js"),
		"TypeScript": filepath.Join(repo, "broken.ts"),
	}

	malformed := map[string]string{
		"Go":         "package main\nimport (\"fmt\"\nfunc main() {",
		"Python":     "from . import\n",
		"JavaScript": "import from 'x'\nconst a = require(\n",
		"TypeScript": "export { from 'pkg'\n",
	}

	for name, file := range filesByAdapter {
		if err := os.WriteFile(file, []byte(malformed[name]), 0o644); err != nil {
			t.Fatalf("failed writing malformed %s fixture: %v", name, err)
		}
	}

	adapters := []LanguageAdapter{NewGoAdapter(), NewPythonAdapter(), NewJavaScriptAdapter(), NewTypeScriptAdapter()}
	for _, adapter := range adapters {
		file := filesByAdapter[adapter.Name()]
		_, _ = adapter.CollectMetrics([]string{file})
		_, _ = adapter.BuildDependencyGraph([]string{file})

		if provider, ok := adapter.(EvidenceProvider); ok {
			_, _, _ = provider.CollectEvidence(repo, []string{file})
		}
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
