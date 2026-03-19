package model

import "sync"

// DependencyGraph represents a language-agnostic dependency graph.
// Nodes represent files or modules, edges represent import/dependency relationships.
// Cycle detection and topology queries live in separate types (GraphCycleDetector,
// FindRoots, FindLeaves) to keep this struct focused on storage and traversal.
type DependencyGraph struct {
	mu      sync.RWMutex
	nodes   map[string]*Node
	edges   map[string][]string // adjacency list: node -> [dependencies]
	reverse map[string][]string // reverse adjacency: node -> [dependents]
}

// Node represents a node in the dependency graph
type Node struct {
	ID         string            // Unique identifier (file path or module path)
	Path       string            // File system path
	Package    string            // Package/module name
	Imports    []string          // Import paths
	IsInternal bool              // Whether this is internal code (vs external dependency)
	Metadata   map[string]string // Additional metadata
}

// NewDependencyGraph creates a new empty dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes:   make(map[string]*Node),
		edges:   make(map[string][]string),
		reverse: make(map[string][]string),
	}
}

// AddNode adds a node to the graph
func (g *DependencyGraph) AddNode(id string, path, pkg string) *Node {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.addNodeLocked(id, path, pkg)
}

func (g *DependencyGraph) addNodeLocked(id string, path, pkg string) *Node {

	if existing, ok := g.nodes[id]; ok {
		return existing
	}

	node := &Node{
		ID:         id,
		Path:       path,
		Package:    pkg,
		Imports:    make([]string, 0),
		IsInternal: true,
		Metadata:   make(map[string]string),
	}

	g.nodes[id] = node
	g.edges[id] = make([]string, 0)
	g.reverse[id] = make([]string, 0)

	return node
}

// AddEdge adds a dependency edge from source to target
func (g *DependencyGraph) AddEdge(source, target string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Ensure both nodes exist
	if _, exists := g.nodes[source]; !exists {
		g.addNodeLocked(source, source, source)
	}
	if _, exists := g.nodes[target]; !exists {
		g.addNodeLocked(target, target, target)
	}

	// Check if edge already exists
	for _, edge := range g.edges[source] {
		if edge == target {
			return
		}
	}

	g.edges[source] = append(g.edges[source], target)
	g.reverse[target] = append(g.reverse[target], source)

	// Update node's imports
	if node, ok := g.nodes[source]; ok {
		node.Imports = append(node.Imports, target)
	}
}

// GetNode retrieves a node by ID
func (g *DependencyGraph) GetNode(id string) *Node {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.nodes[id]
}

// GetNodes returns all nodes in the graph
func (g *DependencyGraph) GetNodes() []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes := make([]*Node, 0, len(g.nodes))
	for _, node := range g.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetDependencies returns the dependencies of a node
func (g *DependencyGraph) GetDependencies(id string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.edges[id]
}

// GetDependents returns nodes that depend on the given node
func (g *DependencyGraph) GetDependents(id string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.reverse[id]
}

// NodeCount returns the total number of nodes
func (g *DependencyGraph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// EdgeCount returns the total number of edges
func (g *DependencyGraph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	count := 0
	for _, edges := range g.edges {
		count += len(edges)
	}
	return count
}

// DetectCycles is a convenience delegate to GraphCycleDetector.
// Prefer using NewGraphCycleDetector(g).DetectCycles() for new code.
func (g *DependencyGraph) DetectCycles() [][]string {
	return NewGraphCycleDetector(g).DetectCycles()
}
