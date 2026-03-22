package main

import (
	"RepoDoctor/internal/analysis"
	"RepoDoctor/internal/domain"
	"RepoDoctor/internal/languages"
	"RepoDoctor/internal/model"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const version = "0.5.0-dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	if err := executeCommand(os.Args[1], os.Args[2:]); err != nil {
		ExitWithError(err)
	}
}

func executeCommand(cmd string, args []string) error {
	switch cmd {
	case "analyze":
		return handleAnalyzeCommand(args)

	case "extract":
		return handleExtractCommand(args)

	case "report":
		return handleReportCommand(args)

	case "history":
		return handleHistoryCommand(args)

	case "interactive":
		return handleInteractiveCommand()

	case "generate":
		return handleGenerateCommand(args)

	case "version":
		return handleVersionCommand()

	case "help", "-h", "--help":
		return handleHelpCommand()

	default:
		printUsage()
		suggestion := getCommandSuggestion(cmd)
		return NewCLIError(
			ErrorCLIUsage,
			fmt.Sprintf("Unknown command: %s", cmd),
			suggestion,
			nil,
		)
	}
}

func handleAnalyzeCommand(args []string) error {
	req, err := composeAnalyzeRequest(args)
	if err != nil {
		return err
	}

	if req.watch {
		runWatch(req.path)
		return nil
	}

	runAnalyze(req.path, req.format, req.verbose, req.colorEnabled, true)
	return nil
}

type analyzeCommandRequest struct {
	path         string
	format       string
	verbose      bool
	colorEnabled bool
	watch        bool
}

func composeAnalyzeRequest(args []string) (*analyzeCommandRequest, error) {
	parsed, err := parseAnalyzeFlags(args)
	if err != nil {
		return nil, err
	}

	resolvedPath := resolveAnalyzePathArg(args, parsed.pathFlag, parsed.positional)
	normalizedPath, normalizeErr := normalizeAnalyzePathInput(resolvedPath)
	if normalizeErr != nil {
		return nil, normalizeErr
	}

	return &analyzeCommandRequest{
		path:         normalizedPath,
		format:       parsed.outputFormat,
		verbose:      parsed.verbose,
		colorEnabled: !parsed.noColor,
		watch:        parsed.watch,
	}, nil
}

type analyzeFlagInput struct {
	pathFlag     string
	outputFormat string
	verbose      bool
	watch        bool
	noColor      bool
	positional   []string
}

func parseAnalyzeFlags(args []string) (*analyzeFlagInput, error) {
	analyzeCmd := flag.NewFlagSet("analyze", flag.ContinueOnError)
	analyzeCmd.SetOutput(os.Stderr)

	path := analyzeCmd.String("path", ".", "Path to analyze")
	format := analyzeCmd.String("format", "text", "Output format (text, json, json-v1)")
	verbose := analyzeCmd.Bool("verbose", false, "Enable verbose output")
	jsonOut := analyzeCmd.Bool("json", false, "Output in JSON format")
	watch := analyzeCmd.Bool("watch", false, "Enable watch mode for continuous analysis")
	noColor := analyzeCmd.Bool("no-color", false, "Disable colored output")

	if err := analyzeCmd.Parse(args); err != nil {
		return nil, NewCLIError(
			ErrorCLIUsage,
			fmt.Sprintf("Invalid analyze arguments: %v", err),
			"Run 'repodoctor help' to review analyze command usage",
			err,
		)
	}

	outputFormat := *format
	if *jsonOut {
		outputFormat = "json"
	}

	return &analyzeFlagInput{
		pathFlag:     *path,
		outputFormat: outputFormat,
		verbose:      *verbose,
		watch:        *watch,
		noColor:      *noColor,
		positional:   analyzeCmd.Args(),
	}, nil
}

func normalizeAnalyzePathInput(pathArg string) (string, error) {
	if strings.TrimSpace(pathArg) == "" {
		return "", NewCLIError(
			ErrorInvalidArgument,
			"Analyze path cannot be empty",
			"Provide a valid repository path with -path or positional argument",
			nil,
		)
	}

	cleaned := filepath.Clean(pathArg)
	absPath, err := filepath.Abs(cleaned)
	if err != nil {
		return "", HandleInvalidPathError(pathArg, err)
	}
	absPath = filepath.Clean(absPath)

	if resolvedPath, err := filepath.EvalSymlinks(absPath); err == nil {
		return filepath.Clean(resolvedPath), nil
	}

	return absPath, nil
}

func resolveAnalyzePathArg(rawArgs []string, pathFlag string, positional []string) string {
	if hasExplicitPathFlag(rawArgs) {
		return pathFlag
	}

	if len(positional) > 0 {
		return positional[0]
	}

	return pathFlag
}

func hasExplicitPathFlag(rawArgs []string) bool {
	for _, arg := range rawArgs {
		if arg == "-path" || arg == "--path" || strings.HasPrefix(arg, "-path=") || strings.HasPrefix(arg, "--path=") {
			return true
		}
	}

	return false
}

func handleExtractCommand(args []string) error {
	extractCmd := flag.NewFlagSet("extract", flag.ExitOnError)
	path := extractCmd.String("path", ".", "Path to extract imports from")
	module := extractCmd.String("module", "RepoDoctor", "Module path for normalization")
	verbose := extractCmd.Bool("verbose", false, "Enable verbose output")
	jsonOut := extractCmd.Bool("json", false, "Output in JSON format")
	extractCmd.Parse(args)

	return runExtract(*path, *module, *verbose, *jsonOut)
}

func handleReportCommand(args []string) error {
	reportCmd := flag.NewFlagSet("report", flag.ExitOnError)
	path := reportCmd.String("path", "repodoctor-report.json", "Path to report file")
	format := reportCmd.String("format", "text", "Output format (text, json)")
	jsonOut := reportCmd.Bool("json", false, "Output in JSON format")
	reportCmd.Parse(args)

	outputFormat := *format
	if *jsonOut {
		outputFormat = "json"
	}

	return runReport(*path, outputFormat)
}

func handleHistoryCommand(args []string) error {
	historyCmd := flag.NewFlagSet("history", flag.ExitOnError)
	path := historyCmd.String("path", ".", "Path to repository")
	historyCmd.Parse(args)

	return runHistory(*path)
}

func handleInteractiveCommand() error {
	runInteractive()
	return nil
}

func handleGenerateCommand(args []string) error {
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	generateCmd.Parse(args)

	return runGenerate(generateCmd.Args())
}

func handleVersionCommand() error {
	fmt.Printf("RepoDoctor v%s\n", version)
	return nil
}

func handleHelpCommand() error {
	printUsage()
	return nil
}

func getCommandSuggestion(cmd string) string {
	commands := []string{"analyze", "extract", "report", "history", "interactive", "generate", "version", "help"}
	closest := ""
	for _, candidate := range commands {
		if strings.HasPrefix(candidate, strings.ToLower(cmd[:min(1, len(cmd))])) || strings.Contains(candidate, strings.ToLower(cmd)) {
			closest = candidate
			break
		}
	}
	if closest != "" {
		return fmt.Sprintf("Did you mean '%s'? Run 'repodoctor help' for available commands", closest)
	}
	return "Run 'repodoctor help' for available commands"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func printUsage() {
	fmt.Println(`RepoDoctor - Static Architecture Intelligence for Go Repositories

Usage:
  repodoctor <command> [options]

Commands:
  analyze      Analyze repository architecture and health
  extract      Extract Go package imports from source files
  report       Display existing analysis report
  history      Show score trend history
  interactive  Start interactive mode for guided analysis
  generate     Generate rule templates and other files
  version      Show version information
  help         Show this help message

Arguments:
  analyze [options]
    -path      Directory path to analyze (default: current directory)
    -format    Output format: text, json, json-v1 (default: text)
    -verbose   Enable verbose output
    -watch     Enable watch mode for continuous analysis
    -no-color  Disable colored output (default: enabled)

  extract [options]
    -path      Directory path to extract imports from (default: current directory)
    -module    Module path for import normalization (default: RepoDoctor)
    -verbose   Enable verbose output

  report [options]
    -path      Path to JSON report file (default: repodoctor-report.json)
    -format    Output format: text, json, json-v1 (default: text)

  history [options]
    -path      Path to repository (default: current directory)

Examples:
  repodoctor analyze .
  repodoctor analyze -path ./myproject -format json
  repodoctor analyze -path . --json
  repodoctor extract .
  repodoctor extract -path ./src -module github.com/myorg/myrepo
  repodoctor report -path ./report.json
  repodoctor history -path .
  repodoctor version`)
}

func runAnalyze(path, format string, verbose bool, colorEnabled bool, exitOnViolation bool) int {
	service := NewAnalysisService()
	return service.Run(AnalyzeRequest{
		Path:            path,
		Format:          format,
		Verbose:         verbose,
		ColorEnabled:    colorEnabled,
		ExitOnViolation: exitOnViolation,
	})
}

// determineExitCode returns the appropriate exit code based on report
// 0 = success (no violations)
// 2 = critical violations (circular dependencies or layer violations)
func determineExitCode(report *StructuralReport) int {
	if !report.HasViolations {
		return 0
	}

	// Critical violations: circular dependencies or layer violations
	if len(report.Circular) > 0 || len(report.Layer) > 0 {
		return 2
	}

	// Non-critical warnings (size/god-object) should not fail CI pipelines.
	return 0
}

func validatePath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		cliErr := HandleInvalidPathError(path, err)
		cliErr.Display()
		os.Exit(1)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		cliErr := HandleFileNotFoundError(absPath, err)
		cliErr.Display()
		os.Exit(1)
	}

	if !info.IsDir() {
		cliErr := NewCLIError(
			ErrorInvalidArgument,
			fmt.Sprintf("Path is not a directory: %s", absPath),
			"Provide a directory path instead of a file",
			nil,
		)
		cliErr.Display()
		os.Exit(1)
	}

	canonicalPath := absPath
	if resolvedPath, resolveErr := filepath.EvalSymlinks(absPath); resolveErr == nil {
		canonicalPath = resolvedPath
	}

	return canonicalPath
}

func extractImports(absPath string, verbose bool) map[string]*ImportMetadata {
	moduleName := "RepoDoctor"
	extractor := NewImportExtractor(moduleName)
	imports, err := extractor.ExtractFromDir(absPath)
	if err != nil && verbose {
		fmt.Fprintf(os.Stderr, "%s", ColorWarn(fmt.Sprintf("Warning: error extracting imports: %v\n", err)))
	}
	return imports
}

func runAdapterPipeline(absPath string) (*analysis.Result, error) {
	ignoreStrategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	config := loadConfiguration(absPath, false)
	policy := languages.DetectionPolicy{}
	if config != nil && config.LanguageDetection != nil {
		policy.LanguageWeights = config.LanguageDetection.Weights
		policy.TieBreakOrder = config.LanguageDetection.TieBreakOrder
		policy.SegmentWeights = config.LanguageDetection.SegmentWeights
	}
	detector := languages.NewRepositoryLanguageDetectorWithPolicy(ignoreStrategy, policy)
	detector.RegisterAdapter(languages.NewGoAdapter())
	detector.RegisterAdapter(languages.NewPythonAdapter())
	detector.RegisterAdapter(languages.NewJavaScriptAdapter())
	detector.RegisterAdapter(languages.NewTypeScriptAdapter())

	orchestrator := analysis.NewOrchestrator(detector)
	return orchestrator.Analyze(absPath)
}

func buildDependencyGraphFromModel(languageGraph *model.DependencyGraph, verbose bool) Graph {
	graph := NewDependencyGraph()
	if languageGraph == nil {
		return graph
	}

	for _, node := range languageGraph.GetNodes() {
		graph.AddNode(node.ID)
		for _, dep := range languageGraph.GetDependencies(node.ID) {
			graph.AddEdge(node.ID, dep)
		}
	}

	if verbose {
		fmt.Printf("%s", ColorInfo(fmt.Sprintf("Built dependency graph with %d nodes and %d edges\n",
			graph.GetNodeCount(), graph.GetEdgeCount())))
	}

	return graph
}

func buildDependencyGraph(imports map[string]*ImportMetadata, verbose bool) Graph {
	graph := NewDependencyGraph()
	for filePath, importMeta := range imports {
		graph.AddNode(filePath)
		for _, imp := range importMeta.Imports {
			graph.AddEdge(filePath, imp)
		}
	}

	if verbose {
		fmt.Printf("%s", ColorInfo(fmt.Sprintf("Built dependency graph with %d nodes and %d edges\n",
			graph.GetNodeCount(), graph.GetEdgeCount())))
	}
	return graph
}

func loadConfiguration(absPath string, verbose bool) *Config {
	configPath := GetConfigPath(absPath)
	configLoader := NewConfigLoader(configPath)
	config, err := configLoader.Load()
	if err != nil {
		if verbose {
			fmt.Printf("%s", ColorWarn(fmt.Sprintf("Warning: error loading config: %v\n", err)))
		}
		config = configLoader.getDefaultConfig()
	}

	if verbose {
		fmt.Printf("%s", ColorInfo(fmt.Sprintf("Configuration loaded from: %s\n", configPath)))
	}
	return config
}

func generateReport(scorer *StructuralScorer, absPath, format string, verbose bool, colorEnabled bool) *StructuralReport {
	reporter := NewColoredReporter(OutputFormat(format), colorEnabled)
	report := reporter.GenerateReport(scorer, absPath, version)

	if format == "json" {
		fmt.Println(reporter.Format(report))
	} else {
		// Use colored output for text format
		var sb strings.Builder
		writeHeaderWithColor(&sb, reporter.formatter)
		writeScoreSectionWithColor(&sb, report, reporter.formatter)
		writeViolationsSummaryWithColor(&sb, report, reporter.formatter)
		writeCircularViolationsWithColor(&sb, report, reporter.formatter)
		writeLayerViolationsWithColor(&sb, report, reporter.formatter)
		writeSizeViolationsWithColor(&sb, report, reporter.formatter)
		writeGodObjectViolationsWithColor(&sb, report, reporter.formatter)
		writeScoreBreakdownWithColor(&sb, report, reporter.formatter)
		fmt.Println(sb.String())
	}
	return report
}

func generateRuleEngineReport(absPath, format string, verbose bool, colorEnabled bool, cfg *Config, summary *runtimeRuleSummary) *StructuralReport {
	report := buildReportFromRuleViolations(absPath, version, cfg, summary.result.Violations)

	if verbose {
		fmt.Printf(ColorInfo("Rules in registry: ")+"%d\n", summary.rulesInScope)
		fmt.Printf(ColorInfo("Rules executed: ")+"%d\n", summary.result.RulesExecuted)
	}

	reporter := NewColoredReporter(OutputFormat(format), colorEnabled)
	if format == "json" {
		fmt.Println(reporter.Format(report))
	} else {
		var sb strings.Builder
		writeHeaderWithColor(&sb, reporter.formatter)
		writeScoreSectionWithColor(&sb, report, reporter.formatter)
		writeViolationsSummaryWithColor(&sb, report, reporter.formatter)
		writeCircularViolationsWithColor(&sb, report, reporter.formatter)
		writeLayerViolationsWithColor(&sb, report, reporter.formatter)
		writeSizeViolationsWithColor(&sb, report, reporter.formatter)
		writeGodObjectViolationsWithColor(&sb, report, reporter.formatter)
		writeScoreBreakdownWithColor(&sb, report, reporter.formatter)
		fmt.Println(sb.String())
	}

	return report
}

func handleTrendAnalysis(absPath string, report *StructuralReport, verbose bool) {
	trendAnalyzer := NewTrendAnalyzer(absPath)
	if err := trendAnalyzer.LoadHistory(); err != nil && verbose {
		fmt.Printf("%s", ColorWarn(fmt.Sprintf("Warning: could not load history: %v\n", err)))
	}

	if verbose {
		fmt.Println()
		fmt.Println(ColorInfo(trendAnalyzer.GetTrendSummary(report.Score.TotalScore)))
	}

	if err := trendAnalyzer.AppendScore(report.Score.TotalScore); err != nil && verbose {
		fmt.Printf("%s", ColorWarn(fmt.Sprintf("Warning: could not save to history: %v\n", err)))
	}
}
