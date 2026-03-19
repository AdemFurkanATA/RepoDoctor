package model

// FindRoots returns nodes with no incoming edges (root packages).
// This is a standalone function rather than a method on DependencyGraph
// to keep the graph struct focused on storage and traversal.
func FindRoots(g *DependencyGraph) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	roots := make([]*Node, 0)
	for id, dependents := range g.reverse {
		if len(dependents) == 0 {
			if node, ok := g.nodes[id]; ok {
				roots = append(roots, node)
			}
		}
	}
	return roots
}

// FindLeaves returns nodes with no outgoing edges (leaf packages).
// This is a standalone function rather than a method on DependencyGraph
// to keep the graph struct focused on storage and traversal.
func FindLeaves(g *DependencyGraph) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	leaves := make([]*Node, 0)
	for id, deps := range g.edges {
		if len(deps) == 0 {
			if node, ok := g.nodes[id]; ok {
				leaves = append(leaves, node)
			}
		}
	}
	return leaves
}
