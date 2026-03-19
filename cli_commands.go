package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
		cliErr := WrapError(err, ErrorRuntime, "Watch mode failed", "Check the target path and try again")
		cliErr.Display()
		os.Exit(1)
	}
}
