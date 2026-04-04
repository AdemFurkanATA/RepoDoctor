package main

import (
	"fmt"
	"os"

	analysispkg "RepoDoctor/internal/analysis"
	metricsPkg "RepoDoctor/internal/metrics"
)

type AnalyzeRequest struct {
	Path            string
	Format          string
	Verbose         bool
	ColorEnabled    bool
	ExitOnViolation bool
	Profiling       profilingRequest
}

type AnalysisService struct{}

func NewAnalysisService() *AnalysisService {
	return &AnalysisService{}
}

func (s *AnalysisService) Run(request AnalyzeRequest) int {
	absPath := validatePath(request.Path)
	InitColorFormatter(request.ColorEnabled)
	profiler, profileErr := startProfiling(request.Profiling)
	if profileErr != nil {
		fmt.Fprintf(os.Stderr, "%s", ColorError(fmt.Sprintf("Error: profiling setup failed: %v\n", profileErr)))
		return 1
	}

	if profiler != nil {
		defer func() {
			if stopErr := profiler.Stop(); stopErr != nil {
				fmt.Fprintf(os.Stderr, "%s", ColorWarn(fmt.Sprintf("Warning: profiling finalization failed: %v\n", stopErr)))
			}
		}()
	}

	progress := NewProgressReporter(!request.Verbose)
	progress.Start("Scanning repository", getStageCount("Scanning repository", absPath))
	if request.Verbose {
		fmt.Printf(ColorInfo("Extracting imports from: ")+"%s\n", absPath)
	}

	analysisResult, err := runAdapterPipeline(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", ColorError(fmt.Sprintf("Error: analysis pipeline failed: %v\n", err)))
		return 1
	}

	if request.Verbose {
		fmt.Printf(ColorInfo("Selected adapter: ")+"%s\n", analysisResult.AdapterName)
	}

	graph := s.reportAdapterGraph(progress, analysisResult, request.Verbose)

	progress.Start("Collecting metrics", getStageCount("Collecting metrics", absPath))
	totalFiles, goFiles, totalLines := scanDirectory(absPath, false)
	_ = totalFiles
	_ = goFiles
	_ = totalLines
	progress.SetProgress(progress.totalSteps)
	progress.Complete()

	progress.Start("Building dependency graph", getStageCount("Building dependency graph", absPath))
	progress.SetProgress(progress.totalSteps)
	progress.Complete()

	config := loadConfiguration(absPath, request.Verbose)

	progress.Start("Running rules", getStageCount("Running rules", absPath))
	var architecture *ArchitectureConfig
	if config != nil {
		architecture = config.Architecture
	}
	ruleSummary := runInternalRulePipelineWithProfile(absPath, graph, analysisResult.AdapterName, architecture)
	progress.SetProgress(progress.totalSteps / 2)

	report := generateRuleEngineReport(absPath, request.Format, request.Verbose, request.ColorEnabled, config, ruleSummary)
	report.Language = collectLanguageEvidenceSummary(absPath, analysisResult.AdapterName)
	progress.SetProgress(progress.totalSteps)
	progress.Complete()

	handleTrendAnalysis(absPath, report, request.Verbose)
	if metricsErr := exportMetricsSnapshot(absPath, request.Format, analysisResult, report); metricsErr != nil {
		fmt.Fprintf(os.Stderr, "%s", ColorWarn(fmt.Sprintf("Warning: metrics export skipped: %v\n", metricsErr)))
	}

	exitCode := determineExitCode(report)
	return exitCode
}

func exportMetricsSnapshot(absPath, outputFormat string, result *analysispkg.Result, report *StructuralReport) error {
	metricsPath, enabled, err := resolveMetricsExportPath(absPath)
	if err != nil || !enabled {
		return err
	}

	snapshot := metricsPkg.Snapshot{
		Version:         version,
		Adapter:         result.AdapterName,
		OutputFormat:    outputFormat,
		FilesDetected:   len(result.Files),
		GraphNodes:      0,
		GraphEdges:      0,
		CircularCount:   len(report.Circular),
		LayerCount:      len(report.Layer),
		SizeCount:       len(report.Size),
		GodObjectCount:  len(report.GodObject),
		TotalViolations: len(report.Circular) + len(report.Layer) + len(report.Size) + len(report.GodObject),
	}

	if result.Graph != nil {
		snapshot.GraphNodes = result.Graph.NodeCount()
		snapshot.GraphEdges = result.Graph.EdgeCount()
	}

	return metricsPkg.Write(metricsPath, snapshot)
}

func (s *AnalysisService) reportAdapterGraph(progress *ProgressReporter, result *analysispkg.Result, verbose bool) Graph {
	progress.SetProgress(progress.totalSteps / 2)
	graph := buildDependencyGraphFromModel(result.Graph, verbose)
	progress.SetProgress(progress.totalSteps)
	progress.Complete()
	return graph
}
