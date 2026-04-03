package analysis

import (
	"fmt"
	"sort"

	"RepoDoctor/internal/languages"
	"RepoDoctor/internal/logger"
	"RepoDoctor/internal/model"
)

// Orchestrator coordinates language detection and adapter-driven analysis steps.
type Orchestrator struct {
	detector languages.LanguageDetector
	logger   logger.Logger
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
	return NewOrchestratorWithLogger(detector, logger.NewNoop())
}

// NewOrchestratorWithLogger creates a new analysis orchestrator with structured logger wiring.
func NewOrchestratorWithLogger(detector languages.LanguageDetector, log logger.Logger) *Orchestrator {
	if log == nil {
		log = logger.NewNoop()
	}
	return &Orchestrator{detector: detector, logger: log}
}

// Analyze executes the runtime pipeline: detect adapter -> detect files -> metrics -> graph.
func (o *Orchestrator) Analyze(repoPath string) (*Result, error) {
	if o.detector == nil {
		return nil, fmt.Errorf("language detector is required")
	}

	o.logger.Debug("analysis.start", "repo_path", repoPath)

	adapter, err := o.detector.DetectLanguage(repoPath)
	if err != nil {
		o.logger.Error("analysis.language_detection_failed", "repo_path", repoPath, "error", err)
		return nil, fmt.Errorf("language detection failed: %w", err)
	}
	o.logger.Info("analysis.language_detected", "adapter", adapter.Name())

	capabilities := adapter.Capabilities()
	if !capabilities.SupportsDependencyGraph {
		return nil, fmt.Errorf("adapter %s does not support dependency graph capability", adapter.Name())
	}
	if !capabilities.SupportsMetrics {
		return nil, fmt.Errorf("adapter %s does not support metrics capability", adapter.Name())
	}

	files, err := adapter.DetectFiles(repoPath)
	if err != nil {
		o.logger.Error("analysis.file_detection_failed", "adapter", adapter.Name(), "error", err)
		return nil, fmt.Errorf("file detection failed for %s: %w", adapter.Name(), err)
	}
	sort.Strings(files)
	o.logger.Debug("analysis.files_detected", "adapter", adapter.Name(), "count", len(files))

	metrics, err := adapter.CollectMetrics(files)
	if err != nil {
		o.logger.Error("analysis.metrics_failed", "adapter", adapter.Name(), "error", err)
		return nil, fmt.Errorf("metrics collection failed for %s: %w", adapter.Name(), err)
	}

	graph, err := adapter.BuildDependencyGraph(files)
	if err != nil {
		o.logger.Error("analysis.graph_failed", "adapter", adapter.Name(), "error", err)
		return nil, fmt.Errorf("dependency graph build failed for %s: %w", adapter.Name(), err)
	}
	o.logger.Info("analysis.completed", "adapter", adapter.Name(), "file_count", len(files))

	return &Result{
		AdapterName: adapter.Name(),
		Files:       files,
		Metrics:     metrics,
		Graph:       graph,
	}, nil
}
