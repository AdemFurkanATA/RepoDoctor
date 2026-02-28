package main

// Graph defines the interface for a directed dependency graph
type Graph interface {
	AddNode(name string)
	AddEdge(from, to string)
	GetDependencies(name string) []string
	DetectCycles() [][]string
	GetAllNodes() []string
	GetNodeCount() int
	GetEdgeCount() int
}

// DependencyGraph implements Graph using adjacency list
type DependencyGraph struct {
	nodes     map[string]bool
	adjacency map[string]map[string]bool
}

// NewDependencyGraph creates a new empty dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes:     make(map[string]bool),
		adjacency: make(map[string]map[string]bool),
	}
}

// AddNode adds a node to the graph
func (g *DependencyGraph) AddNode(name string) {
	if !g.nodes[name] {
		g.nodes[name] = true
		if g.adjacency[name] == nil {
			g.adjacency[name] = make(map[string]bool)
		}
	}
}

// AddEdge adds a directed edge from 'from' to 'to'
func (g *DependencyGraph) AddEdge(from, to string) {
	// Ensure both nodes exist
	g.AddNode(from)
	g.AddNode(to)

	// Add edge if it doesn't exist
	if !g.adjacency[from][to] {
		g.adjacency[from][to] = true
	}
}

// GetDependencies returns all dependencies (outgoing edges) for a node
func (g *DependencyGraph) GetDependencies(name string) []string {
	neighbors := g.adjacency[name]
	if neighbors == nil {
		return []string{}
	}

	deps := make([]string, 0, len(neighbors))
	for dep := range neighbors {
		deps = append(deps, dep)
	}

	return deps
}

// GetAllNodes returns all nodes in the graph
func (g *DependencyGraph) GetAllNodes() []string {
	nodes := make([]string, 0, len(g.nodes))
	for node := range g.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNodeCount returns the number of nodes in the graph
func (g *DependencyGraph) GetNodeCount() int {
	return len(g.nodes)
}

// GetEdgeCount returns the number of edges in the graph
func (g *DependencyGraph) GetEdgeCount() int {
	count := 0
	for _, neighbors := range g.adjacency {
		count += len(neighbors)
	}
	return count
}

// DetectCycles finds all cycles in the graph using DFS
// Returns a slice of cycles, where each cycle is a slice of node names
func (g *DependencyGraph) DetectCycles() [][]string {
	cycles := [][]string{}
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, dep := range g.GetDependencies(node) {
			if !visited[dep] {
				dfs(dep)
			} else if recStack[dep] {
				// Found a cycle - extract it from path
				cycleStart := -1
				for i, n := range path {
					if n == dep {
						cycleStart = i
						break
					}
				}
				if cycleStart != -1 {
					cycle := append([]string{}, path[cycleStart:]...)
					cycles = append(cycles, cycle)
				}
			}
		}

		// Backtrack
		path = path[:len(path)-1]
		recStack[node] = false
	}

	// Run DFS from each unvisited node
	for node := range g.nodes {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}
