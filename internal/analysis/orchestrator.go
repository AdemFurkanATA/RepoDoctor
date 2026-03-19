package analysis

import (
	"fmt"
	"sort"

	"RepoDoctor/internal/languages"
	"RepoDoctor/internal/model"
)

// Orchestrator coordinates language detection and adapter-driven analysis steps.
type Orchestrator struct {
	detector languages.LanguageDetector
}

// Result contains adapter-driven analysis output.
type Result struct {
	AdapterName string
	Files       []string
	Metrics     *model.RepositoryMetrics
	Graph       *model.DependencyGraph
}

// NewOrchestrator creates a new analysis orchestrator.
func NewOrchestrator(detector languages.LanguageDetector) *Orchestrator {
	return &Orchestrator{detector: detector}
}

// Analyze executes the runtime pipeline: detect adapter -> detect files -> metrics -> graph.
func (o *Orchestrator) Analyze(repoPath string) (*Result, error) {
	if o.detector == nil {
		return nil, fmt.Errorf("language detector is required")
	}

	adapter, err := o.detector.DetectLanguage(repoPath)
	if err != nil {
		return nil, fmt.Errorf("language detection failed: %w", err)
	}

	files, err := adapter.DetectFiles(repoPath)
	if err != nil {
		return nil, fmt.Errorf("file detection failed for %s: %w", adapter.Name(), err)
	}
	sort.Strings(files)

	metrics, err := adapter.CollectMetrics(files)
	if err != nil {
		return nil, fmt.Errorf("metrics collection failed for %s: %w", adapter.Name(), err)
	}

	graph, err := adapter.BuildDependencyGraph(files)
	if err != nil {
		return nil, fmt.Errorf("dependency graph build failed for %s: %w", adapter.Name(), err)
	}

	return &Result{
		AdapterName: adapter.Name(),
		Files:       files,
		Metrics:     metrics,
		Graph:       graph,
	}, nil
}
