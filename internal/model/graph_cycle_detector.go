package model

// GraphCycleDetector performs cycle detection on a DependencyGraph.
// Extracted from DependencyGraph to satisfy SRP — the graph stores
// structure, the detector runs analysis algorithms over it.
type GraphCycleDetector struct {
	graph  *DependencyGraph
	cycles [][]string
}

// NewGraphCycleDetector creates a cycle detector bound to the given graph.
func NewGraphCycleDetector(graph *DependencyGraph) *GraphCycleDetector {
	return &GraphCycleDetector{
		graph:  graph,
		cycles: make([][]string, 0),
	}
}

// DetectCycles performs DFS-based cycle detection and returns all cycles found.
func (d *GraphCycleDetector) DetectCycles() [][]string {
	d.cycles = make([][]string, 0)

	nodes := d.graph.GetNodes()
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range d.graph.GetDependencies(node) {
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
					d.cycles = append(d.cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
		return false
	}

	for _, node := range nodes {
		if !visited[node.ID] {
			dfs(node.ID)
		}
	}

	return d.cycles
}

// GetCycles returns previously detected cycles without re-running detection.
func (d *GraphCycleDetector) GetCycles() [][]string {
	return d.cycles
}

// HasCycles runs detection if needed and returns whether cycles exist.
func (d *GraphCycleDetector) HasCycles() bool {
	if len(d.cycles) == 0 {
		d.DetectCycles()
	}
	return len(d.cycles) > 0
}
