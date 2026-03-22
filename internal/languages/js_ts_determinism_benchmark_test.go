package languages

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestJSTSParity_DeterministicAcrossRepeatedRuns(t *testing.T) {
	repo := t.TempDir()
	file := filepath.Join(repo, "index.ts")
	content := strings.Join([]string{
		"import React from 'react'",
		"export * from '@scope/pkg/utils'",
		"const axios = require('axios/lib/core')",
		"async function load() { return import('dayjs/plugin/utc') }",
		"function longFn() {",
		"  const a = 1",
		"  const b = 2",
		"  const c = 3",
		"  const d = 4",
		"  const e = 5",
		"  const f = 6",
		"  return a + b + c + d + e + f",
		"}",
	}, "\n")
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing fixture: %v", err)
	}

	adapter := NewTypeScriptAdapter()
	files, err := adapter.DetectFiles(repo)
	if err != nil {
		t.Fatalf("DetectFiles failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected one TypeScript file, got %d", len(files))
	}
	target := files[0]

	var baseline string
	for i := 0; i < 20; i++ {
		graph, graphErr := adapter.BuildDependencyGraph(files)
		if graphErr != nil {
			t.Fatalf("BuildDependencyGraph failed on iteration %d: %v", i, graphErr)
		}
		node := graph.GetNode(target)
		if node == nil {
			t.Fatalf("missing node on iteration %d", i)
		}
		current := strings.Join(node.Imports, "|")
		if i == 0 {
			baseline = current
			continue
		}
		if current != baseline {
			t.Fatalf("non-deterministic imports on iteration %d\nbase=%s\nnow=%s", i, baseline, current)
		}
	}
}

func BenchmarkJSTSParity_TypeScriptAdapterGraphBuild(b *testing.B) {
	repo := b.TempDir()
	for i := 0; i < 200; i++ {
		file := filepath.Join(repo, "m"+strconv.Itoa(i)+".ts")
		content := "import x from 'react'\nexport * from '@scope/pkg/utils'\nconst y = require('axios/lib/core')\n"
		if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
			b.Fatalf("failed writing fixture: %v", err)
		}
	}

	adapter := NewTypeScriptAdapter()
	files, err := adapter.DetectFiles(repo)
	if err != nil {
		b.Fatalf("DetectFiles failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := adapter.BuildDependencyGraph(files); err != nil {
			b.Fatalf("BuildDependencyGraph failed: %v", err)
		}
	}
}
