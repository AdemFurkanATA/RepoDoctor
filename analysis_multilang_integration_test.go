package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInternalRulePipeline_MultiLanguageMixedFixtureDeterministic(t *testing.T) {
	repo := t.TempDir()

	files := map[string]string{
		"app/main.go":       "package main\nimport \"fmt\"\nfunc main(){fmt.Println(1)}\n",
		"py/service.py":     "def run():\n    return 1\n",
		"web/index.ts":      "export const x = 1\n",
		"web/helpers.js":    "module.exports = {}\n",
		"web/feature.tsx":   "export function Feature(){ return null }\n",
		"scripts/build.mjs": "import x from 'vite'\n",
	}

	for rel, content := range files {
		path := filepath.Join(repo, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed creating dir for %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed writing fixture %s: %v", rel, err)
		}
	}

	graph := NewDependencyGraph()
	for rel := range files {
		abs := filepath.Join(repo, rel)
		graph.AddNode(abs)
	}

	first := runInternalRulePipeline(repo, graph, "TypeScript")
	if first == nil || first.result == nil {
		t.Fatal("expected non-nil runtime summary")
	}

	for i := 0; i < 10; i++ {
		next := runInternalRulePipeline(repo, graph, "TypeScript")
		if next == nil || next.result == nil {
			t.Fatalf("expected non-nil summary on iteration %d", i)
		}
		if first.rulesInScope != next.rulesInScope {
			t.Fatalf("rules in scope changed on iteration %d: %d vs %d", i, first.rulesInScope, next.rulesInScope)
		}
		if first.result.RulesExecuted != next.result.RulesExecuted {
			t.Fatalf("rules executed changed on iteration %d: %d vs %d", i, first.result.RulesExecuted, next.result.RulesExecuted)
		}
		if len(first.result.Violations) != len(next.result.Violations) {
			t.Fatalf("violation count changed on iteration %d: %d vs %d", i, len(first.result.Violations), len(next.result.Violations))
		}
	}
}
