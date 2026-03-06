package model

import "sync"

// DependencyGraph represents a language-agnostic dependency graph
// Nodes represent files or modules, edges represent import/dependency relationships.
type DependencyGraph struct {
	mu         sync.RWMutex
	nodes      map[string]*Node
	edges      map[string][]string // adjacency list: node -> [dependencies]
	reverse    map[string][]string // reverse adjacency: node -> [dependents]
	cycles     [][]string          // detected cycles
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
		cycles:  make([][]string, 0),
	}
}

// AddNode adds a node to the graph
func (g *DependencyGraph) AddNode(id string, path, pkg string) *Node {
	g.mu.Lock()
	defer g.mu.Unlock()

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
		g.AddNode(source, source, source)
	}
	if _, exists := g.nodes[target]; !exists {
		g.AddNode(target, target, target)
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

// DetectCycles performs cycle detection using DFS
func (g *DependencyGraph) DetectCycles() [][]string {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.cycles = make([][]string, 0)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range g.edges[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				// Found a cycle
				cycleStart := -1
				for i, n := range path {
					if n == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart != -1 {
					cycle := append([]string{}, path[cycleStart:]...)
					g.cycles = append(g.cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
		return false
	}

	for node := range g.nodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return g.cycles
}

// GetCycles returns detected cycles
func (g *DependencyGraph) GetCycles() [][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.cycles
}

// HasCycles returns true if the graph contains cycles
func (g *DependencyGraph) HasCycles() bool {
	if len(g.cycles) == 0 {
		g.DetectCycles()
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.cycles) > 0
}

// GetRoots returns nodes with no incoming edges (root packages)
func (g *DependencyGraph) GetRoots() []*Node {
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

// GetLeaves returns nodes with no outgoing edges (leaf packages)
func (g *DependencyGraph) GetLeaves() []*Node {
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
