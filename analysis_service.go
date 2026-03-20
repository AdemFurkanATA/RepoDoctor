package main

import (
	"fmt"
	"os"

	analysispkg "RepoDoctor/internal/analysis"
)

type AnalyzeRequest struct {
	Path            string
	Format          string
	Verbose         bool
	ColorEnabled    bool
	ExitOnViolation bool
}

type AnalysisService struct{}

func NewAnalysisService() *AnalysisService {
	return &AnalysisService{}
}

func (s *AnalysisService) Run(request AnalyzeRequest) int {
	absPath := validatePath(request.Path)
	InitColorFormatter(request.ColorEnabled)

	progress := NewProgressReporter(!request.Verbose)
	progress.Start("Scanning repository", getStageCount("Scanning repository", absPath))
	if request.Verbose {
		fmt.Printf(ColorInfo("Extracting imports from: ")+"%s\n", absPath)
	}

	analysisResult, err := runAdapterPipeline(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", ColorError(fmt.Sprintf("Error: analysis pipeline failed: %v\n", err)))
		if request.ExitOnViolation {
			os.Exit(1)
		}
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
	ruleSummary := runInternalRulePipeline(absPath, graph)
	progress.SetProgress(progress.totalSteps / 2)

	report := generateRuleEngineReport(absPath, request.Format, request.Verbose, request.ColorEnabled, config, ruleSummary)
	progress.SetProgress(progress.totalSteps)
	progress.Complete()

	handleTrendAnalysis(absPath, report, request.Verbose)

	exitCode := determineExitCode(report)
	if request.ExitOnViolation && exitCode != 0 {
		os.Exit(exitCode)
	}

	return exitCode
}

func (s *AnalysisService) reportAdapterGraph(progress *ProgressReporter, result *analysispkg.Result, verbose bool) Graph {
	progress.SetProgress(progress.totalSteps / 2)
	graph := buildDependencyGraphFromModel(result.Graph, verbose)
	progress.SetProgress(progress.totalSteps)
	progress.Complete()
	return graph
}
