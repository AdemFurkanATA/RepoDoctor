package main

import (
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
	analyzeCmd := flag.NewFlagSet("analyze", flag.ExitOnError)
	path := analyzeCmd.String("path", ".", "Path to analyze")
	format := analyzeCmd.String("format", "text", "Output format (text, json)")
	verbose := analyzeCmd.Bool("verbose", false, "Enable verbose output")
	jsonOut := analyzeCmd.Bool("json", false, "Output in JSON format")
	watch := analyzeCmd.Bool("watch", false, "Enable watch mode for continuous analysis")
	noColor := analyzeCmd.Bool("no-color", false, "Disable colored output")
	analyzeCmd.Parse(args)

	outputFormat := *format
	if *jsonOut {
		outputFormat = "json"
	}
	if *watch {
		runWatch(*path)
		return nil
	}

	runAnalyze(*path, outputFormat, *verbose, !*noColor, true)
	return nil
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
    -format    Output format: text, json (default: text)
    -verbose   Enable verbose output
    -watch     Enable watch mode for continuous analysis
    -no-color  Disable colored output (default: enabled)

  extract [options]
    -path      Directory path to extract imports from (default: current directory)
    -module    Module path for import normalization (default: RepoDoctor)
    -verbose   Enable verbose output

  report [options]
    -path      Path to JSON report file (default: repodoctor-report.json)
    -format    Output format: text, json (default: text)

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
	// Validate and resolve path
	absPath := validatePath(path)

	// Initialize color formatter
	InitColorFormatter(colorEnabled)

	// Extract imports and build dependency graph
	// Create progress reporter (enabled when not verbose)
	progress := NewProgressReporter(!verbose)

	// Stage 1: Repository scanning
	progress.Start("Scanning repository", getStageCount("Scanning repository", absPath))
	if verbose {
		fmt.Printf(ColorInfo("Extracting imports from: ")+"%s\n", absPath)
	}

	// Stage 2: Import extraction and dependency graph
	imports := extractImports(absPath, verbose)
	progress.SetProgress(progress.totalSteps / 2)

	graph := buildDependencyGraph(imports, verbose)
	progress.SetProgress(progress.totalSteps)
	progress.Complete()

	// Stage 2: Metrics collection
	progress.Start("Collecting metrics", getStageCount("Collecting metrics", absPath))
	totalFiles, goFiles, totalLines := scanDirectory(absPath, false)
	_ = totalFiles
	_ = goFiles
	_ = totalLines
	progress.SetProgress(progress.totalSteps)
	progress.Complete()

	// Stage 3: Dependency graph building
	progress.Start("Building dependency graph", getStageCount("Building dependency graph", absPath))
	progress.SetProgress(progress.totalSteps)
	progress.Complete()

	// Load configuration
	config := loadConfiguration(absPath, verbose)

	// Create scorer and run analysis
	progress.Start("Running rules", getStageCount("Running rules", absPath))
	scorer := NewStructuralScorer(graph, config, absPath)
	progress.SetProgress(progress.totalSteps / 2)

	// Generate and display report
	report := generateReport(scorer, absPath, format, verbose, colorEnabled)
	progress.SetProgress(progress.totalSteps)
	progress.Complete()

	// Trend analysis
	handleTrendAnalysis(absPath, report, verbose)

	// Exit with appropriate code based on violations
	exitCode := determineExitCode(report)
	if exitOnViolation && exitCode != 0 {
		os.Exit(exitCode)
	}

	return exitCode
}

// determineExitCode returns the appropriate exit code based on report
// 0 = success (no violations)
// 1 = warnings (low/medium severity violations)
// 2 = critical violations (circular dependencies or layer violations)
func determineExitCode(report *StructuralReport) int {
	if !report.HasViolations {
		return 0
	}

	// Critical violations: circular dependencies or layer violations
	if len(report.Circular) > 0 || len(report.Layer) > 0 {
		return 2
	}

	// Warnings: size or god object violations
	return 1
}

func validatePath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		err := HandleInvalidPathError(path, err)
		err.Display()
		fmt.Fprintf(os.Stderr, ColorError(fmt.Sprintf("Error: Could not resolve path: %v\n", err)))
		fmt.Fprintf(os.Stderr, ColorInfo("Suggestion: Use an absolute or valid relative path\n"))
		os.Exit(1)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		err := HandleFileNotFoundError(absPath, err)
		err.Display()
		fmt.Fprintf(os.Stderr, ColorError(fmt.Sprintf("Error: Path does not exist: %s\n", absPath)))
		fmt.Fprintf(os.Stderr, ColorInfo("Suggestion: Check if the path is correct and accessible\n"))
		os.Exit(1)
	}

	if !info.IsDir() {
		err := NewCLIError(
			ErrorInvalidArgument,
			fmt.Sprintf("Path is not a directory: %s", absPath),
			"Provide a directory path instead of a file",
			nil,
		)
		err.Display()
		fmt.Fprintf(os.Stderr, ColorError(fmt.Sprintf("Error: Path is not a directory: %s\n", absPath)))
		fmt.Fprintf(os.Stderr, ColorInfo("Suggestion: Provide a directory path instead of a file\n"))
		os.Exit(1)
	}

	return absPath
}

func extractImports(absPath string, verbose bool) map[string]*ImportMetadata {
	moduleName := "RepoDoctor"
	extractor := NewImportExtractor(moduleName)
	imports, err := extractor.ExtractFromDir(absPath)
	if err != nil && verbose {
		fmt.Fprintf(os.Stderr, ColorWarn(fmt.Sprintf("Warning: error extracting imports: %v\n", err)))
	}
	return imports
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
		fmt.Printf(ColorInfo(fmt.Sprintf("Built dependency graph with %d nodes and %d edges\n",
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
			fmt.Printf(ColorWarn(fmt.Sprintf("Warning: error loading config: %v\n", err)))
		}
		config = configLoader.getDefaultConfig()
	}

	if verbose {
		fmt.Printf(ColorInfo(fmt.Sprintf("Configuration loaded from: %s\n", configPath)))
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

func handleTrendAnalysis(absPath string, report *StructuralReport, verbose bool) {
	trendAnalyzer := NewTrendAnalyzer(absPath)
	if err := trendAnalyzer.LoadHistory(); err != nil && verbose {
		fmt.Printf(ColorWarn(fmt.Sprintf("Warning: could not load history: %v\n", err)))
	}

	if verbose {
		fmt.Println()
		fmt.Println(ColorInfo(trendAnalyzer.GetTrendSummary(report.Score.TotalScore)))
	}

	if err := trendAnalyzer.AppendScore(report.Score.TotalScore); err != nil && verbose {
		fmt.Printf(ColorWarn(fmt.Sprintf("Warning: could not save to history: %v\n", err)))
	}
}

func scanDirectory(path string, verbose bool) (totalFiles, goFiles, totalLines int) {
	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories and files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip docs directory (as per user request)
		if info.IsDir() && info.Name() == "docs" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			if verbose {
				fmt.Printf("📂 Scanning: %s\n", filePath)
			}
			return nil
		}

		totalFiles++

		// Count Go files
		if strings.HasSuffix(info.Name(), ".go") {
			goFiles++

			// Count lines
			data, err := os.ReadFile(filePath)
			if err == nil {
				lines := strings.Split(string(data), "\n")
				totalLines += len(lines)

				if verbose {
					fmt.Printf("  📄 %s (%d lines)\n", info.Name(), len(lines))
				}
			}
		}

		return nil
	})

	return
}

func runReport(reportPath, format string) error {
	// Read report file
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return WrapError(err, ErrorAnalysis, fmt.Sprintf("Error reading report file: %s", reportPath), GetSuggestion(err.Error()))
	}

	// Parse report based on format
	if format == "json" {
		// Output JSON as-is
		fmt.Println(string(data))
	} else {
		// For text format, parse JSON and format
		fmt.Println("📊 RepoDoctor Analysis Report")
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println(string(data))
		fmt.Println(strings.Repeat("─", 60))
		fmt.Println("✨ Report displayed successfully")
	}

	return nil
}

func runHistory(repoPath string) error {
	// Resolve path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return HandleInvalidPathError(repoPath, err)
	}

	// Load trend history
	trendAnalyzer := NewTrendAnalyzer(absPath)
	if err := trendAnalyzer.LoadHistory(); err != nil {
		return WrapError(err, ErrorRuntime, "Error loading history", GetSuggestion(err.Error()))
	}

	// Display history
	fmt.Println("📈 Score Trend History")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println(trendAnalyzer.GetTrendSummary(0))
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("✨ History retrieved successfully")

	return nil
}

func runExtract(path, module string, verbose bool, jsonOutput bool) error {
	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return HandleInvalidPathError(path, err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return HandleFileNotFoundError(absPath, err)
	}

	if !info.IsDir() {
		return NewCLIError(
			ErrorInvalidArgument,
			fmt.Sprintf("Path is not a directory: %s", absPath),
			"Provide a directory path instead of a file",
			nil,
		)
	}

	fmt.Printf("RepoDoctor v%s\n", version)
	fmt.Printf("Extracting imports from: %s\n", absPath)
	fmt.Printf("Module path: %s\n\n", module)

	// Create extractor and extract imports
	extractor := NewImportExtractor(module)
	imports, err := extractor.ExtractFromDir(absPath)
	if err != nil {
		return WrapError(err, ErrorAnalysis, "Error extracting imports", GetSuggestion(err.Error()))
	}

	// Display results
	fmt.Println("📊 Import Extraction Results")
	fmt.Println(strings.Repeat("─", 60))

	totalImports := 0
	for filePath, metadata := range imports {
		relPath, _ := filepath.Rel(absPath, filePath)
		if relPath == "" {
			relPath = filePath
		}

		fmt.Printf("\n📄 %s (package: %s)\n", relPath, metadata.Package)
		if len(metadata.Imports) > 0 {
			for _, imp := range metadata.Imports {
				fmt.Printf("   • %s\n", imp)
				totalImports++
			}
		} else {
			fmt.Printf("   (no external imports)\n")
		}

		if verbose {
			fmt.Printf("   └─ Absolute: %s\n", filePath)
		}
	}

	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("📦 Total files analyzed: %d\n", len(imports))
	fmt.Printf("📥 Total unique imports: %d\n", totalImports)
	fmt.Println("✨ Import extraction completed successfully")
	fmt.Println()

	_ = jsonOutput
	return nil
}

func runGenerate(args []string) error {
	if len(args) < 2 {
		return HandleCLIUsageError("Usage: repodoctor generate rule <rule-name>", nil)
	}

	if args[0] != "rule" {
		return NewCLIError(
			ErrorInvalidArgument,
			fmt.Sprintf("Unknown generate type: %s", args[0]),
			"Available types: rule",
			nil,
		)
	}

	ruleName := args[1]
	generator := NewRuleTemplateGenerator("rules")

	if err := generator.Generate(ruleName); err != nil {
		return WrapError(err, ErrorRuntime, "Error generating rule", GetSuggestion(err.Error()))
	}

	return nil
}

func runWatch(path string) {
	if err := WatchAndAnalyze(path); err != nil {
		fmt.Fprintf(os.Stderr, "Error in watch mode: %v\n", err)
		os.Exit(1)
	}
}
