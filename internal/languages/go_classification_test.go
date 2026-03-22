package languages

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClassifyGoImport(t *testing.T) {
	tests := []struct {
		name  string
		input string
		class GoImportClass
	}{
		{name: "stdlib", input: "fmt", class: GoImportStdlib},
		{name: "stdlib nested", input: "net/http", class: GoImportStdlib},
		{name: "internal segment", input: "github.com/org/repo/internal/service", class: GoImportInternal},
		{name: "external", input: "github.com/pkg/errors", class: GoImportExternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyGoImport(tt.input); got != tt.class {
				t.Fatalf("expected %s, got %s", tt.class, got)
			}
		})
	}
}

func TestGoAdapter_BuildDependencyGraph_StoresImportClassificationMetadata(t *testing.T) {
	repo := t.TempDir()
	path := filepath.Join(repo, "main.go")
	content := `package main
import (
  "fmt"
  "github.com/pkg/errors"
  "github.com/myorg/repo/internal/service"
)
func main() { _, _, _ = fmt.Println, errors.New, service.Run }
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed writing go fixture: %v", err)
	}

	adapter := NewGoAdapter()
	graph, err := adapter.BuildDependencyGraph([]string{path})
	if err != nil {
		t.Fatalf("BuildDependencyGraph failed: %v", err)
	}

	node := graph.GetNode(path)
	if node == nil {
		t.Fatalf("expected graph node for %s", path)
	}

	if got := node.Metadata["import_class:fmt"]; got != string(GoImportStdlib) {
		t.Fatalf("expected stdlib classification for fmt, got %q", got)
	}
	if got := node.Metadata["import_class:github.com/pkg/errors"]; got != string(GoImportExternal) {
		t.Fatalf("expected external classification for github.com/pkg/errors, got %q", got)
	}
	if got := node.Metadata["import_class:github.com/myorg/repo/internal/service"]; got != string(GoImportInternal) {
		t.Fatalf("expected internal classification for internal import, got %q", got)
	}
}
