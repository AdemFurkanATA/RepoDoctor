package main

import (
	"fmt"
	"strings"
)

func writeHeader(sb *strings.Builder) {
	sb.WriteString("╔═══════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║          RepoDoctor Structural Analysis Report           ║\n")
	sb.WriteString("╚═══════════════════════════════════════════════════════════╝\n\n")
}

func writeScoreSection(sb *strings.Builder, report *StructuralReport) {
	sb.WriteString(fmt.Sprintf("Version: %s\n", report.Version))
	sb.WriteString(fmt.Sprintf("Path: %s\n\n", report.Path))

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
}

func writeViolationsSummary(sb *strings.Builder, report *StructuralReport) {
	sb.WriteString("┌───────────────────────────────────────────────────────────┐\n")
	sb.WriteString("│  VIOLATIONS SUMMARY                                       │\n")
	sb.WriteString("└───────────────────────────────────────────────────────────┘\n")
	sb.WriteString(fmt.Sprintf("Total Violations: %d\n", report.Score.ViolationCount))
	sb.WriteString(fmt.Sprintf("  - Circular Dependencies: %d\n", report.Score.CircularCount))
	sb.WriteString(fmt.Sprintf("  - Layer Violations: %d\n", report.Score.LayerCount))
	sb.WriteString(fmt.Sprintf("  - Size Violations: %d\n", report.Score.SizeCount))
	sb.WriteString(fmt.Sprintf("  - God Objects: %d\n\n", report.Score.GodObjectCount))
}

func writeCircularViolations(sb *strings.Builder, report *StructuralReport) {
	if len(report.Circular) == 0 {
		return
	}

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

func writeLayerViolations(sb *strings.Builder, report *StructuralReport) {
	if len(report.Layer) == 0 {
		return
	}

	sb.WriteString("┌───────────────────────────────────────────────────────────┐\n")
	sb.WriteString("│  LAYER VIOLATIONS [HIGH]                                  │\n")
	sb.WriteString("└───────────────────────────────────────────────────────────┘\n")

	for i, v := range report.Layer {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, v.Message))
	}
	sb.WriteString("\n")
}

func writeSizeViolations(sb *strings.Builder, report *StructuralReport) {
	if len(report.Size) == 0 {
		return
	}

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

func writeGodObjectViolations(sb *strings.Builder, report *StructuralReport) {
	if len(report.GodObject) == 0 {
		return
	}

	sb.WriteString("┌───────────────────────────────────────────────────────────┐\n")
	sb.WriteString("│  GOD OBJECT VIOLATIONS [MEDIUM]                           │\n")
	sb.WriteString("└───────────────────────────────────────────────────────────┘\n")

	for i, v := range report.GodObject {
		sb.WriteString(fmt.Sprintf("[%d] Struct '%s' in %s: %d fields, %d methods\n",
			i+1, v.StructName, v.File, v.FieldCount, v.MethodCount))
	}
	sb.WriteString("\n")
}

func writeScoreBreakdown(sb *strings.Builder, report *StructuralReport) {
	if !report.HasViolations {
		sb.WriteString("✨ No structural violations detected! Your architecture is clean.\n\n")
		return
	}

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
	sb.WriteString(fmt.Sprintf("God Object Penalty:   -%.1f (%d violations x 5.0)\n",
		report.Score.GodObjectPenalty, report.Score.GodObjectCount))
	sb.WriteString(fmt.Sprintf("─────────────────────────────────────────────────\n"))
	sb.WriteString(fmt.Sprintf("Final Score:          %.1f\n\n", report.Score.TotalScore))
}
