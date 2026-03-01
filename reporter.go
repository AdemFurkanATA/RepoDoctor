package main

import (
	"fmt"
	"strings"
)

// OutputFormat defines the output format type
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

// StructuralReport represents the complete analysis report
type StructuralReport struct {
	Version       string
	Path          string
	Score         *StructuralScore
	Circular      []CycleViolation
	Layer         []LayerViolation
	Size          []SizeViolation
	HasViolations bool
}

// Reporter handles formatting and displaying structural analysis results
type Reporter struct {
	format OutputFormat
}

// NewReporter creates a new reporter with the specified format
func NewReporter(format OutputFormat) *Reporter {
	return &Reporter{
		format: format,
	}
}

// GenerateReport creates a structural report from a scorer
func (r *Reporter) GenerateReport(scorer *StructuralScorer, path, version string) *StructuralReport {
	violations := scorer.GetAllViolations()

	return &StructuralReport{
		Version:  version,
		Path:     path,
		Score:    scorer.CalculateScore(),
		Circular: violations.Circular,
		Layer:    violations.Layer,
		Size:     violations.Size,
		HasViolations: len(violations.Circular) > 0 || len(violations.Layer) > 0 || len(violations.Size) > 0,
	}
}

// Format formats the report according to the output format
func (r *Reporter) Format(report *StructuralReport) string {
	switch r.format {
	case FormatJSON:
		return r.formatJSON(report)
	default:
		return r.formatText(report)
	}
}

// formatText formats the report as human-readable text
func (r *Reporter) formatText(report *StructuralReport) string {
	var sb strings.Builder

	sb.WriteString("╔═══════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║          RepoDoctor Structural Analysis Report           ║\n")
	sb.WriteString("╚═══════════════════════════════════════════════════════════╝\n\n")

	sb.WriteString(fmt.Sprintf("Version: %s\n", report.Version))
	sb.WriteString(fmt.Sprintf("Path: %s\n\n", report.Path))

	// Score section
	sb.WriteString("┌───────────────────────────────────────────────────────────┐\n")
	sb.WriteString("│  STRUCTURAL HEALTH SCORE                                  │\n")
	sb.WriteString("└───────────────────────────────────────────────────────────┘\n")
	
	scoreIndicator := "✓"
	if report.Score.TotalScore < 70 {
		scoreIndicator = "⚠"
	}
	if report.Score.TotalScore < 50 {
		scoreIndicator = "✗"
	}

	sb.WriteString(fmt.Sprintf("%s Score: %.1f / 100.0\n\n", scoreIndicator, report.Score.TotalScore))

	// Violations summary
	sb.WriteString("┌───────────────────────────────────────────────────────────┐\n")
	sb.WriteString("│  VIOLATIONS SUMMARY                                       │\n")
	sb.WriteString("└───────────────────────────────────────────────────────────┘\n")
	sb.WriteString(fmt.Sprintf("Total Violations: %d\n", report.Score.ViolationCount))
	sb.WriteString(fmt.Sprintf("  - Circular Dependencies: %d\n", report.Score.CircularCount))
	sb.WriteString(fmt.Sprintf("  - Layer Violations: %d\n", report.Score.LayerCount))
	sb.WriteString(fmt.Sprintf("  - Size Violations: %d\n\n", report.Score.SizeCount))

	// Circular dependencies
	if len(report.Circular) > 0 {
		sb.WriteString("┌───────────────────────────────────────────────────────────┐\n")
		sb.WriteString("│  CIRCULAR DEPENDENCIES [CRITICAL]                         │\n")
		sb.WriteString("└───────────────────────────────────────────────────────────┘\n")
		
		for i, v := range report.Circular {
			sb.WriteString(fmt.Sprintf("[%d] ", i+1))
			sb.WriteString(formatCyclePath(v.Path))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Layer violations
	if len(report.Layer) > 0 {
		sb.WriteString("┌───────────────────────────────────────────────────────────┐\n")
		sb.WriteString("│  LAYER VIOLATIONS [HIGH]                                  │\n")
		sb.WriteString("└───────────────────────────────────────────────────────────┘\n")
		
		for i, v := range report.Layer {
			sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, v.Message))
		}
		sb.WriteString("\n")
	}

	// Size violations
	if len(report.Size) > 0 {
		sb.WriteString("┌───────────────────────────────────────────────────────────┐\n")
		sb.WriteString("│  SIZE VIOLATIONS [LOW]                                    │\n")
		sb.WriteString("└───────────────────────────────────────────────────────────┘\n")
		
		for i, v := range report.Size {
			if v.Function != "" {
				sb.WriteString(fmt.Sprintf("[%d] Function '%s' in %s: %d lines (threshold: %d)\n", 
					i+1, v.Function, v.File, v.Lines, v.Threshold))
			} else {
				sb.WriteString(fmt.Sprintf("[%d] File %s: %d lines (threshold: %d)\n", 
					i+1, v.File, v.Lines, v.Threshold))
			}
		}
		sb.WriteString("\n")
	}

	// Score breakdown
	if report.HasViolations {
		sb.WriteString("┌───────────────────────────────────────────────────────────┐\n")
		sb.WriteString("│  SCORE BREAKDOWN                                          │\n")
		sb.WriteString("└───────────────────────────────────────────────────────────┘\n")
		sb.WriteString(fmt.Sprintf("Base Score:           100.0\n"))
		sb.WriteString(fmt.Sprintf("Circular Penalty:     -%.1f (%d violations x 10.0)\n", 
			report.Score.CircularPenalty, report.Score.CircularCount))
		sb.WriteString(fmt.Sprintf("Layer Penalty:        -%.1f (%d violations x 5.0)\n", 
			report.Score.LayerPenalty, report.Score.LayerCount))
		sb.WriteString(fmt.Sprintf("Size Penalty:         -%.1f (%d violations x 3.0)\n", 
			report.Score.SizePenalty, report.Score.SizeCount))
		sb.WriteString(fmt.Sprintf("─────────────────────────────────────────────────\n"))
		sb.WriteString(fmt.Sprintf("Final Score:          %.1f\n\n", report.Score.TotalScore))
	}

	if !report.HasViolations {
		sb.WriteString("✨ No structural violations detected! Your architecture is clean.\n\n")
	}

	return sb.String()
}

// formatCyclePath formats a cycle path for display
func formatCyclePath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	
	result := ""
	for i, pkg := range path {
		result += pkg
		if i < len(path)-1 {
			result += " → "
		}
	}
	// Complete the cycle
	result += " → " + path[0]
	
	return result
}

// formatJSON formats the report as JSON
func (r *Reporter) formatJSON(report *StructuralReport) string {
	var sb strings.Builder
	
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("  \"version\": \"%s\",\n", report.Version))
	sb.WriteString(fmt.Sprintf("  \"path\": \"%s\",\n", report.Path))
	sb.WriteString("  \"score\": {\n")
	sb.WriteString(fmt.Sprintf("    \"total\": %.2f,\n", report.Score.TotalScore))
	sb.WriteString(fmt.Sprintf("    \"max\": %.2f,\n", report.Score.MaxScore))
	sb.WriteString(fmt.Sprintf("    \"circularPenalty\": %.2f,\n", report.Score.CircularPenalty))
	sb.WriteString(fmt.Sprintf("    \"layerPenalty\": %.2f,\n", report.Score.LayerPenalty))
	sb.WriteString(fmt.Sprintf("    \"sizePenalty\": %.2f\n", report.Score.SizePenalty))
	sb.WriteString("  },\n")
	sb.WriteString("  \"violations\": {\n")
	sb.WriteString(fmt.Sprintf("    \"circular\": %d,\n", report.Score.CircularCount))
	sb.WriteString(fmt.Sprintf("    \"layer\": %d,\n", report.Score.LayerCount))
	sb.WriteString(fmt.Sprintf("    \"size\": %d\n", report.Score.SizeCount))
	sb.WriteString("  },\n")
	
	// Circular violations
	sb.WriteString("  \"circularViolations\": [\n")
	for i, v := range report.Circular {
		sb.WriteString("    {\n")
		sb.WriteString(fmt.Sprintf("      \"path\": %s,\n", formatStringArray(v.Path)))
		sb.WriteString(fmt.Sprintf("      \"severity\": \"%s\"\n", v.Severity))
		sb.WriteString("    }")
		if i < len(report.Circular)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("  ],\n")
	
	// Layer violations
	sb.WriteString("  \"layerViolations\": [\n")
	for i, v := range report.Layer {
		sb.WriteString("    {\n")
		sb.WriteString(fmt.Sprintf("      \"from\": \"%s\",\n", v.From))
		sb.WriteString(fmt.Sprintf("      \"to\": \"%s\",\n", v.To))
		sb.WriteString(fmt.Sprintf("      \"message\": \"%s\"\n", v.Message))
		sb.WriteString("    }")
		if i < len(report.Layer)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("  ],\n")
	
	// Size violations
	sb.WriteString("  \"sizeViolations\": [\n")
	for i, v := range report.Size {
		sb.WriteString("    {\n")
		sb.WriteString(fmt.Sprintf("      \"file\": \"%s\",\n", v.File))
		sb.WriteString(fmt.Sprintf("      \"function\": \"%s\",\n", v.Function))
		sb.WriteString(fmt.Sprintf("      \"lines\": %d,\n", v.Lines))
		sb.WriteString(fmt.Sprintf("      \"threshold\": %d\n", v.Threshold))
		sb.WriteString("    }")
		if i < len(report.Size)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}\n")
	
	return sb.String()
}

// formatStringArray formats a string array as JSON
func formatStringArray(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	
	result := "["
	for i, s := range arr {
		result += fmt.Sprintf("\"%s\"", s)
		if i < len(arr)-1 {
			result += ", "
		}
	}
	result += "]"
	return result
}
