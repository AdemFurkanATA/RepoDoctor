package model

import (
	"reflect"
	"testing"
)

func TestFindRoots_ReturnsExpectedSet(t *testing.T) {
	g := NewDependencyGraph()
	g.AddEdge("a", "b")
	g.AddNode("c", "c.go", "c")

	got := sortedNodeIDs(FindRoots(g))
	want := []string{"a", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected roots: got %v want %v", got, want)
	}
}

func TestFindLeaves_ReturnsExpectedSet(t *testing.T) {
	g := NewDependencyGraph()
	g.AddEdge("a", "b")
	g.AddNode("c", "c.go", "c")

	got := sortedNodeIDs(FindLeaves(g))
	want := []string{"b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected leaves: got %v want %v", got, want)
	}
}

func TestFindRootsAndLeaves_IsolatedNode_BothRootAndLeaf(t *testing.T) {
	g := NewDependencyGraph()
	g.AddNode("isolated", "isolated.go", "isolated")

	roots := sortedNodeIDs(FindRoots(g))
	leaves := sortedNodeIDs(FindLeaves(g))

	if !reflect.DeepEqual(roots, []string{"isolated"}) {
		t.Fatalf("unexpected roots for isolated graph: %v", roots)
	}
	if !reflect.DeepEqual(leaves, []string{"isolated"}) {
		t.Fatalf("unexpected leaves for isolated graph: %v", leaves)
	}
}
