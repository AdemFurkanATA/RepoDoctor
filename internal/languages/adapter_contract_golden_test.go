package languages

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type adapterContractGoldenCase struct {
	name       string
	golden     string
	files      map[string]string
	adapters   []LanguageAdapter
	iterations int
}

type adapterContractSnapshot struct {
	Adapter     string                    `json:"adapter"`
	DetectFiles []string                  `json:"detectFiles"`
	FileNodes   []adapterContractFileNode `json:"fileNodes"`
}

type adapterContractFileNode struct {
	Path    string   `json:"path"`
	Imports []string `json:"imports"`
}

func TestAdapterContract_GoldenSnapshots(t *testing.T) {
	cases := []adapterContractGoldenCase{
		{
			name:   "valid",
			golden: "testdata/adapter_contract_goldens/valid.golden.json",
			files: map[string]string{
				"go/app/main.go":     "package main\nimport (\n\t\"fmt\"\n\t\"net/http\"\n\t\"github.com/acme/lib\"\n)\nfunc main(){fmt.Println(http.MethodGet, lib.Version)}\n",
				"python/pkg/main.py": "import os\nfrom collections import defaultdict\nfrom app.core import run\n",
				"web/src/main.js":    "import React from 'react'\nconst util = require('@scope/pkg/utils')\nconst feature = import('./feature/module')\nexport { thing } from 'lodash/fp'\n",
				"web/src/main.ts":    "import { readFile } from 'node:fs'\nimport '@/app/bootstrap'\nexport * from '~/types/user'\nconst logger = require('#/infra/logger')\n",
			},
			adapters:   []LanguageAdapter{NewGoAdapter(), NewPythonAdapter(), NewJavaScriptAdapter(), NewTypeScriptAdapter()},
			iterations: 10,
		},
		{
			name:   "malformed",
			golden: "testdata/adapter_contract_goldens/malformed.golden.json",
			files: map[string]string{
				"go/app/broken.go":     "package main\nimport (\"fmt\"\nfunc main(){\n",
				"python/pkg/broken.py": "from . import\n",
				"web/src/broken.js":    "import from 'x'\nconst a = require(\n",
				"web/src/broken.ts":    "export { from 'pkg'\n",
			},
			adapters:   []LanguageAdapter{NewGoAdapter(), NewPythonAdapter(), NewJavaScriptAdapter(), NewTypeScriptAdapter()},
			iterations: 10,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := t.TempDir()
			writeFixtures(t, repo, tc.files)

			baseline := collectAdapterSnapshots(t, repo, tc.adapters)
			for i := 0; i < tc.iterations; i++ {
				current := collectAdapterSnapshots(t, repo, tc.adapters)
				if !equalAdapterSnapshots(baseline, current) {
					t.Fatalf("adapter snapshots are not deterministic on iteration %d\nfirst=%s\nnext=%s", i, mustMarshalPrettyJSON(t, baseline), mustMarshalPrettyJSON(t, current))
				}
			}

			got := mustMarshalPrettyJSON(t, baseline)
			wantBytes, err := os.ReadFile(filepath.Join(".", tc.golden))
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v", tc.golden, err)
			}

			want := strings.TrimSpace(string(wantBytes))
			if strings.TrimSpace(got) != want {
				t.Fatalf("golden mismatch for %s\nwant:\n%s\n\ngot:\n%s", tc.name, want, got)
			}
		})
	}
}

func writeFixtures(t *testing.T, repo string, files map[string]string) {
	t.Helper()

	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, rel := range paths {
		abs := filepath.Join(repo, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("failed to create fixture directory for %s: %v", rel, err)
		}
		if err := os.WriteFile(abs, []byte(files[rel]), 0o644); err != nil {
			t.Fatalf("failed to write fixture %s: %v", rel, err)
		}
	}
}

func collectAdapterSnapshots(t *testing.T, repo string, adapters []LanguageAdapter) []adapterContractSnapshot {
	t.Helper()

	snapshots := make([]adapterContractSnapshot, 0, len(adapters))
	for _, adapter := range adapters {
		detected, err := adapter.DetectFiles(repo)
		if err != nil {
			t.Fatalf("%s DetectFiles failed: %v", adapter.Name(), err)
		}

		normalizedDetected := make([]string, 0, len(detected))
		for _, file := range detected {
			normalizedDetected = append(normalizedDetected, normalizeContractPath(t, repo, file))
		}

		graph, err := adapter.BuildDependencyGraph(detected)
		if err != nil {
			t.Fatalf("%s BuildDependencyGraph failed: %v", adapter.Name(), err)
		}

		fileNodes := make([]adapterContractFileNode, 0, len(normalizedDetected))
		for i, file := range detected {
			node := graph.GetNode(file)
			if node == nil {
				continue
			}

			imports := append([]string(nil), node.Imports...)
			fileNodes = append(fileNodes, adapterContractFileNode{
				Path:    normalizedDetected[i],
				Imports: imports,
			})
		}

		snapshots = append(snapshots, adapterContractSnapshot{
			Adapter:     adapter.Name(),
			DetectFiles: normalizedDetected,
			FileNodes:   fileNodes,
		})
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Adapter < snapshots[j].Adapter
	})

	return snapshots
}

func equalAdapterSnapshots(a, b []adapterContractSnapshot) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Adapter != b[i].Adapter {
			return false
		}
		if !equalStringSlice(a[i].DetectFiles, b[i].DetectFiles) {
			return false
		}
		if len(a[i].FileNodes) != len(b[i].FileNodes) {
			return false
		}
		for j := range a[i].FileNodes {
			if a[i].FileNodes[j].Path != b[i].FileNodes[j].Path {
				return false
			}
			if !equalStringSlice(a[i].FileNodes[j].Imports, b[i].FileNodes[j].Imports) {
				return false
			}
		}
	}

	return true
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func normalizeContractPath(t *testing.T, root, target string) string {
	t.Helper()

	rel, err := filepath.Rel(root, target)
	if err == nil {
		relSlash := filepath.ToSlash(rel)
		if relSlash != "." && !strings.HasPrefix(relSlash, "../") {
			return relSlash
		}
	}

	targetSlash := filepath.ToSlash(target)
	for _, marker := range []string{"go/", "python/", "web/"} {
		if idx := strings.Index(targetSlash, marker); idx >= 0 {
			return targetSlash[idx:]
		}
	}

	t.Fatalf("failed to normalize deterministic contract path for %s", target)
	return ""
}

func mustMarshalPrettyJSON(t *testing.T, data interface{}) string {
	t.Helper()

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal snapshot: %v", err)
	}

	return string(bytes)
}
