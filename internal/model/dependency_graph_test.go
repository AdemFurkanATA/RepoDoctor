package model

import "testing"

func TestDependencyGraph_AddNode_Idempotent(t *testing.T) {
	g := NewDependencyGraph()
	first := g.AddNode("a", "a.go", "pkg/a")
	second := g.AddNode("a", "ignored.go", "ignored")

	if first != second {
		t.Fatal("expected AddNode to return existing node for duplicate id")
	}
	if g.NodeCount() != 1 {
		t.Fatalf("expected node count 1, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Fatalf("expected edge count 0, got %d", g.EdgeCount())
	}
}

func TestDependencyGraph_AddEdge_AutoCreatesNodes(t *testing.T) {
	g := NewDependencyGraph()
	g.AddEdge("a", "b")

	if g.NodeCount() != 2 {
		t.Fatalf("expected 2 auto-created nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 1 {
		t.Fatalf("expected 1 edge, got %d", g.EdgeCount())
	}

	deps := g.GetDependencies("a")
	if len(deps) != 1 || deps[0] != "b" {
		t.Fatalf("unexpected dependencies for a: %v", deps)
	}
	dependents := g.GetDependents("b")
	if len(dependents) != 1 || dependents[0] != "a" {
		t.Fatalf("unexpected dependents for b: %v", dependents)
	}
}

func TestDependencyGraph_AddEdge_DeduplicatesEdgeAndImports(t *testing.T) {
	g := NewDependencyGraph()
	g.AddNode("svc", "svc.go", "service")
	g.AddNode("repo", "repo.go", "repository")

	g.AddEdge("svc", "repo")
	g.AddEdge("svc", "repo")

	if g.EdgeCount() != 1 {
		t.Fatalf("expected de-duplicated edge count 1, got %d", g.EdgeCount())
	}

	node := g.GetNode("svc")
	if node == nil {
		t.Fatal("expected source node to exist")
	}
	if len(node.Imports) != 1 || node.Imports[0] != "repo" {
		t.Fatalf("expected imports [repo], got %v", node.Imports)
	}
}

func TestDependencyGraph_CountsAndAccessors(t *testing.T) {
	g := NewDependencyGraph()
	g.AddNode("a", "a.go", "a")
	g.AddNode("b", "b.go", "b")
	g.AddNode("c", "c.go", "c")
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")

	if g.NodeCount() != 3 {
		t.Fatalf("expected node count 3, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 2 {
		t.Fatalf("expected edge count 2, got %d", g.EdgeCount())
	}

	node := g.GetNode("a")
	if node == nil || node.Path != "a.go" || node.Package != "a" {
		t.Fatalf("unexpected node metadata: %+v", node)
	}
}

func TestDependencyGraph_UnknownNodeAccessors(t *testing.T) {
	g := NewDependencyGraph()

	if g.GetNode("missing") != nil {
		t.Fatal("expected nil for unknown node")
	}
	if len(g.GetDependencies("missing")) != 0 {
		t.Fatalf("expected empty dependencies for unknown node")
	}
	if len(g.GetDependents("missing")) != 0 {
		t.Fatalf("expected empty dependents for unknown node")
	}
}
