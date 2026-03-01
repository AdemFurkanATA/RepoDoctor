package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const version = "0.2.0-dev"

func main() {
	// Command flags
	analyzeCmd := flag.NewFlagSet("analyze", flag.ExitOnError)
	analyzePath := analyzeCmd.String("path", ".", "Path to analyze")
	analyzeFormat := analyzeCmd.String("format", "text", "Output format (text, json)")
	analyzeVerbose := analyzeCmd.Bool("verbose", false, "Enable verbose output")

	// Extract imports command
	extractCmd := flag.NewFlagSet("extract", flag.ExitOnError)
	extractPath := extractCmd.String("path", ".", "Path to extract imports from")
	extractModule := extractCmd.String("module", "RepoDoctor", "Module path for normalization")

	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	// Main command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "analyze":
		analyzeCmd.Parse(os.Args[2:])
		runAnalyze(*analyzePath, *analyzeFormat, *analyzeVerbose)
	case "extract":
		extractCmd.Parse(os.Args[2:])
		runExtract(*extractPath, *extractModule, *analyzeVerbose)
	case "version":
		versionCmd.Parse(os.Args[2:])
		fmt.Printf("RepoDoctor v%s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`RepoDoctor - Static Architecture Intelligence for Go Repositories

Usage:
  repodoctor <command> [options]

Commands:
  analyze    Analyze repository architecture and health
  extract    Extract Go package imports from source files
  version    Show version information
  help       Show this help message

Arguments:
  analyze [options]
    -path      Directory path to analyze (default: current directory)
    -format    Output format: text, json (default: text)
    -verbose   Enable verbose output

  extract [options]
    -path      Directory path to extract imports from (default: current directory)
    -module    Module path for import normalization (default: RepoDoctor)
    -verbose   Enable verbose output

Examples:
  repodoctor analyze .
  repodoctor analyze -path ./myproject -format json
  repodoctor extract .
  repodoctor extract -path ./src -module github.com/myorg/myrepo
  repodoctor version`)
}

func runAnalyze(path, format string, verbose bool) {
	// Validate and resolve path
	absPath := validatePath(path)

	// Extract imports and build dependency graph
	if verbose {
		fmt.Printf("Extracting imports from: %s\n", absPath)
	}

	imports := extractImports(absPath, verbose)
	graph := buildDependencyGraph(imports, verbose)

	// Load configuration
	config := loadConfiguration(absPath, verbose)

	// Create scorer and run analysis
	scorer := NewStructuralScorer(graph, config, absPath)
	
	// Generate and display report
	report := generateReport(scorer, absPath, format, verbose)

	// Trend analysis
	handleTrendAnalysis(absPath, report, verbose)

	// Exit with error code if critical violations found
	if report.HasViolations {
		os.Exit(1)
	}
}

func validatePath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: path does not exist: %s\n", absPath)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: path is not a directory: %s\n", absPath)
		os.Exit(1)
	}

	return absPath
}

func extractImports(absPath string, verbose bool) map[string]*ImportMetadata {
	moduleName := "RepoDoctor"
	extractor := NewImportExtractor(moduleName)
	imports, err := extractor.ExtractFromDir(absPath)
	if err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: error extracting imports: %v\n", err)
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
		fmt.Printf("Built dependency graph with %d nodes and %d edges\n", 
			graph.GetNodeCount(), graph.GetEdgeCount())
	}
	return graph
}

func loadConfiguration(absPath string, verbose bool) *Config {
	configPath := GetConfigPath(absPath)
	configLoader := NewConfigLoader(configPath)
	config, err := configLoader.Load()
	if err != nil {
		if verbose {
			fmt.Printf("Warning: error loading config: %v\n", err)
		}
		config = configLoader.getDefaultConfig()
	}

	if verbose {
		fmt.Printf("Configuration loaded from: %s\n", configPath)
	}
	return config
}

func generateReport(scorer *StructuralScorer, absPath, format string, verbose bool) *StructuralReport {
	reporter := NewReporter(OutputFormat(format))
	report := reporter.GenerateReport(scorer, absPath, version)

	if format == "json" {
		fmt.Println(reporter.Format(report))
	} else {
		fmt.Println(reporter.Format(report))
	}
	return report
}

func handleTrendAnalysis(absPath string, report *StructuralReport, verbose bool) {
	trendAnalyzer := NewTrendAnalyzer(absPath)
	if err := trendAnalyzer.LoadHistory(); err != nil && verbose {
		fmt.Printf("Warning: could not load history: %v\n", err)
	}
	
	if verbose {
		fmt.Println()
		fmt.Println(trendAnalyzer.GetTrendSummary(report.Score.TotalScore))
	}
	
	if err := trendAnalyzer.AppendScore(report.Score.TotalScore); err != nil && verbose {
		fmt.Printf("Warning: could not save to history: %v\n", err)
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

func runExtract(path, module string, verbose bool) {
	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: path does not exist: %s\n", absPath)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: path is not a directory: %s\n", absPath)
		os.Exit(1)
	}

	fmt.Printf("RepoDoctor v%s\n", version)
	fmt.Printf("Extracting imports from: %s\n", absPath)
	fmt.Printf("Module path: %s\n\n", module)

	// Create extractor and extract imports
	extractor := NewImportExtractor(module)
	imports, err := extractor.ExtractFromDir(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting imports: %v\n", err)
		os.Exit(1)
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
}
