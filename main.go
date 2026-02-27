package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const version = "0.1.0-dev"

func main() {
	// Command flags
	analyzeCmd := flag.NewFlagSet("analyze", flag.ExitOnError)
	analyzePath := analyzeCmd.String("path", ".", "Path to analyze")
	analyzeFormat := analyzeCmd.String("format", "text", "Output format (text, json)")
	analyzeVerbose := analyzeCmd.Bool("verbose", false, "Enable verbose output")

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
  version    Show version information
  help       Show this help message

Arguments:
  analyze [options]
    -path      Directory path to analyze (default: current directory)
    -format    Output format: text, json (default: text)
    -verbose   Enable verbose output

Examples:
  repodoctor analyze .
  repodoctor analyze -path ./myproject -format json
  repodoctor version`)
}

func runAnalyze(path, format string, verbose bool) {
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
	fmt.Printf("Analyzing: %s\n\n", absPath)

	// Perform basic analysis
	files, goFiles, totalLines := scanDirectory(absPath, verbose)

	// Display results
	if format == "json" {
		fmt.Printf("{\n")
		fmt.Printf("  \"version\": \"%s\",\n", version)
		fmt.Printf("  \"path\": \"%s\",\n", absPath)
		fmt.Printf("  \"totalFiles\": %d,\n", files)
		fmt.Printf("  \"goFiles\": %d,\n", goFiles)
		fmt.Printf("  \"totalLines\": %d,\n", totalLines)
		fmt.Printf("  \"status\": \"v0.1-dev\"\n")
		fmt.Printf("}\n")
	} else {
		fmt.Println("📊 Analysis Results")
		fmt.Println(strings.Repeat("─", 40))
		fmt.Printf("📁 Total Files:      %d\n", files)
		fmt.Printf("📄 Go Files:         %d\n", goFiles)
		fmt.Printf("📝 Total Lines:      %d\n", totalLines)
		fmt.Println(strings.Repeat("─", 40))
		fmt.Println("Status: v0.1-dev (Early Development)")
		fmt.Println("✨ Full analysis engine coming soon...")
		fmt.Println()
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
