package languages

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"RepoDoctor/internal/model"
)

func TestGoAdapter_ParallelBuildDependencyGraph_Deterministic(t *testing.T) {
	repo := t.TempDir()
	writeGoFixture(t, repo, "go.mod", "module example.com/parallel\n\ngo 1.21\n")
	writeGoFixture(t, repo, "a.go", "package main\nimport \"example.com/parallel/pkg\"\nfunc a(){}\n")
	writeGoFixture(t, repo, "pkg/p.go", "package pkg\nimport \"fmt\"\nfunc p(){}\n")
	writeGoFixture(t, repo, "pkg/q.go", "package pkg\nimport \"example.com/parallel/internal/x\"\nfunc q(){}\n")
	writeGoFixture(t, repo, "internal/x/x.go", "package x\nfunc X(){}\n")

	adapter := NewGoAdapter()
	adapter.parseWorkers = 4

	files, err := adapter.DetectFiles(repo)
	if err != nil {
		t.Fatalf("DetectFiles failed: %v", err)
	}
	sort.Strings(files)

	first, err := adapter.BuildDependencyGraph(files)
	if err != nil {
		t.Fatalf("BuildDependencyGraph failed: %v", err)
	}
	baseline := canonicalGraphSnapshot(first)

	for i := 0; i < 20; i++ {
		next, nextErr := adapter.BuildDependencyGraph(files)
		if nextErr != nil {
			t.Fatalf("BuildDependencyGraph iteration %d failed: %v", i, nextErr)
		}
		if got := canonicalGraphSnapshot(next); got != baseline {
			t.Fatalf("non-deterministic graph output on iteration %d\nwant=%s\ngot=%s", i, baseline, got)
		}
	}
}

func BenchmarkGoAdapter_ParallelParsing(b *testing.B) {
	benchmarkGoAdapterParsing(b, 4)
}

func BenchmarkGoAdapter_SerialParsing(b *testing.B) {
	benchmarkGoAdapterParsing(b, 1)
}

func benchmarkGoAdapterParsing(b *testing.B, workers int) {
	repo := b.TempDir()
	writeGoFixture(b, repo, "go.mod", "module example.com/bench\n\ngo 1.21\n")

	for i := 0; i < 250; i++ {
		content := "package bench\n"
		if i > 0 {
			content += "import \"fmt\"\n"
		}
		content += fmt.Sprintf("func F%d(){}\n", i)
		writeGoFixture(b, repo, filepath.Join("pkg", fmt.Sprintf("f_%03d.go", i)), content)
	}

	adapter := NewGoAdapter()
	adapter.parseWorkers = workers

	files, err := adapter.DetectFiles(repo)
	if err != nil {
		b.Fatalf("DetectFiles failed: %v", err)
	}
	sort.Strings(files)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := adapter.CollectMetrics(files); err != nil {
			b.Fatalf("CollectMetrics failed: %v", err)
		}
		if _, err := adapter.BuildDependencyGraph(files); err != nil {
			b.Fatalf("BuildDependencyGraph failed: %v", err)
		}
	}
}

func canonicalGraphSnapshot(graph *model.DependencyGraph) string {
	nodes := graph.GetNodes()
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })

	parts := make([]string, 0, len(nodes))
	for _, node := range nodes {
		deps := append([]string(nil), graph.GetDependencies(node.ID)...)
		sort.Strings(deps)
		parts = append(parts, node.ID+"=>"+strings.Join(deps, ","))
	}

	return strings.Join(parts, "|")
}

func writeGoFixture(tb testing.TB, root, relPath, content string) {
	tb.Helper()
	fullPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		tb.Fatalf("failed creating fixture dir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		tb.Fatalf("failed writing fixture file %s: %v", relPath, err)
	}
}
