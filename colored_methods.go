package main

import (
	"fmt"
	"strings"
)

// writeHeaderWithColor writes the report header with colors
func writeHeaderWithColor(sb *strings.Builder, formatter *ColorFormatter) {
	header := "╔═══════════════════════════════════════════════════════════╗"
	title := "║          RepoDoctor Structural Analysis Report           ║"
	footer := "╚═══════════════════════════════════════════════════════════╝"

	sb.WriteString(formatter.Color(header, ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color(title, ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color(footer, ColorCyan))
	sb.WriteString("\n\n")
}

// writeScoreSectionWithColor writes the score section with colors
func writeScoreSectionWithColor(sb *strings.Builder, report *StructuralReport, formatter *ColorFormatter) {
	sb.WriteString(fmt.Sprintf("Version: %s\n", report.Version))
	sb.WriteString(fmt.Sprintf("Path: %s\n\n", report.Path))

	sb.WriteString(formatter.Color("┌───────────────────────────────────────────────────────────┐", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("│  STRUCTURAL HEALTH SCORE                                  │", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("└───────────────────────────────────────────────────────────┘", ColorCyan))
	sb.WriteString("\n")

	scoreIndicator := formatter.Success("✓")
	if report.Score.TotalScore < 70 {
		scoreIndicator = formatter.Warn("⚠")
	}
	if report.Score.TotalScore < 50 {
		scoreIndicator = formatter.Error("✗")
	}

	sb.WriteString(fmt.Sprintf("%s Score: %s\n\n", scoreIndicator, formatter.Bold(fmt.Sprintf("%.1f / 100.0", report.Score.TotalScore))))
}

// writeViolationsSummaryWithColor writes the violations summary with colors
func writeViolationsSummaryWithColor(sb *strings.Builder, report *StructuralReport, formatter *ColorFormatter) {
	sb.WriteString(formatter.Color("┌───────────────────────────────────────────────────────────┐", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("│  VIOLATIONS SUMMARY                                       │", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("└───────────────────────────────────────────────────────────┘", ColorCyan))
	sb.WriteString("\n")

	totalViolations := report.Score.ViolationCount
	if totalViolations == 0 {
		sb.WriteString(formatter.Success("✓ No violations detected") + "\n")
	} else {
		sb.WriteString(fmt.Sprintf("Total Violations: %s\n", formatter.Error(fmt.Sprintf("%d", totalViolations))))
		sb.WriteString(fmt.Sprintf("  - Circular Dependencies: %s\n", formatter.Error(fmt.Sprintf("%d", report.Score.CircularCount))))
		sb.WriteString(fmt.Sprintf("  - Layer Violations: %s\n", formatter.Warn(fmt.Sprintf("%d", report.Score.LayerCount))))
		sb.WriteString(fmt.Sprintf("  - Size Violations: %s\n", formatter.Info(fmt.Sprintf("%d", report.Score.SizeCount))))
		sb.WriteString(fmt.Sprintf("  - God Objects: %s\n\n", formatter.Info(fmt.Sprintf("%d", report.Score.GodObjectCount))))
	}
}

// writeCircularViolationsWithColor writes circular dependency violations with colors
func writeCircularViolationsWithColor(sb *strings.Builder, report *StructuralReport, formatter *ColorFormatter) {
	if len(report.Circular) == 0 {
		return
	}

	sb.WriteString(formatter.Color("┌───────────────────────────────────────────────────────────┐", ColorRed))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("│  CIRCULAR DEPENDENCIES [CRITICAL]                         │", ColorRed))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("└───────────────────────────────────────────────────────────┘", ColorRed))
	sb.WriteString("\n")

	for i, v := range report.Circular {
		sb.WriteString(formatter.Error(fmt.Sprintf("[%d] ", i+1)))
		sb.WriteString(formatter.Color(formatCyclePath(v.Path), ColorRed))
		sb.WriteString("\n")
		if v.Hint != "" {
			sb.WriteString(formatter.Info(fmt.Sprintf("    Hint: %s\n", v.Hint)))
		}
	}
	sb.WriteString("\n")
}

// writeLayerViolationsWithColor writes layer violations with colors
func writeLayerViolationsWithColor(sb *strings.Builder, report *StructuralReport, formatter *ColorFormatter) {
	if len(report.Layer) == 0 {
		return
	}

	sb.WriteString(formatter.Color("┌───────────────────────────────────────────────────────────┐", ColorYellow))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("│  LAYER VIOLATIONS [HIGH]                                  │", ColorYellow))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("└───────────────────────────────────────────────────────────┘", ColorYellow))
	sb.WriteString("\n")

	for i, v := range report.Layer {
		sb.WriteString(formatter.Warn(fmt.Sprintf("[%d] %s\n", i+1, v.Message)))
		if v.Hint != "" {
			sb.WriteString(formatter.Info(fmt.Sprintf("    Hint: %s\n", v.Hint)))
		}
	}
	sb.WriteString("\n")
}

// writeSizeViolationsWithColor writes size violations with colors
func writeSizeViolationsWithColor(sb *strings.Builder, report *StructuralReport, formatter *ColorFormatter) {
	if len(report.Size) == 0 {
		return
	}

	sb.WriteString(formatter.Color("┌───────────────────────────────────────────────────────────┐", ColorBlue))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("│  SIZE VIOLATIONS [LOW]                                    │", ColorBlue))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("└───────────────────────────────────────────────────────────┘", ColorBlue))
	sb.WriteString("\n")

	for i, v := range report.Size {
		if v.Function != "" {
			sb.WriteString(formatter.Info(fmt.Sprintf("[%d] Function '%s' in %s: %d lines (threshold: %d)\n",
				i+1, v.Function, v.File, v.Lines, v.Threshold)))
		} else {
			sb.WriteString(formatter.Info(fmt.Sprintf("[%d] File %s: %d lines (threshold: %d)\n",
				i+1, v.File, v.Lines, v.Threshold)))
		}
		if v.Hint != "" {
			sb.WriteString(formatter.Info(fmt.Sprintf("    Hint: %s\n", v.Hint)))
		}
	}
	sb.WriteString("\n")
}

// writeGodObjectViolationsWithColor writes god object violations with colors
func writeGodObjectViolationsWithColor(sb *strings.Builder, report *StructuralReport, formatter *ColorFormatter) {
	if len(report.GodObject) == 0 {
		return
	}

	sb.WriteString(formatter.Color("┌───────────────────────────────────────────────────────────┐", ColorYellow))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("│  GOD OBJECT VIOLATIONS [MEDIUM]                           │", ColorYellow))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("└───────────────────────────────────────────────────────────┘", ColorYellow))
	sb.WriteString("\n")

	for i, v := range report.GodObject {
		sb.WriteString(formatter.Warn(fmt.Sprintf("[%d] Struct '%s' in %s: %d fields, %d methods\n",
			i+1, v.StructName, v.File, v.FieldCount, v.MethodCount)))
		if v.Hint != "" {
			sb.WriteString(formatter.Info(fmt.Sprintf("    Hint: %s\n", v.Hint)))
		}
	}
	sb.WriteString("\n")
}

func writeComplexityBandsWithColor(sb *strings.Builder, report *StructuralReport, formatter *ColorFormatter) {
	sb.WriteString(formatter.Color("┌───────────────────────────────────────────────────────────┐", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("│  CYCLOMATIC COMPLEXITY BANDS                              │", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("└───────────────────────────────────────────────────────────┘", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Info(fmt.Sprintf("Low (<=10): %d\n", report.Complexity.Low)))
	sb.WriteString(formatter.Info(fmt.Sprintf("Medium (11-20): %d\n", report.Complexity.Medium)))
	sb.WriteString(formatter.Info(fmt.Sprintf("High (>20): %d\n\n", report.Complexity.High)))
}

func writeTechnicalDebtSummaryWithColor(sb *strings.Builder, report *StructuralReport, formatter *ColorFormatter) {
	sb.WriteString(formatter.Color("┌───────────────────────────────────────────────────────────┐", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("│  TECHNICAL DEBT ESTIMATE                                  │", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("└───────────────────────────────────────────────────────────┘", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Info(fmt.Sprintf("Circular: %dh\n", report.Debt.CircularHours)))
	sb.WriteString(formatter.Info(fmt.Sprintf("Layer: %dh\n", report.Debt.LayerHours)))
	sb.WriteString(formatter.Info(fmt.Sprintf("Size: %dh\n", report.Debt.SizeHours)))
	sb.WriteString(formatter.Info(fmt.Sprintf("God Object: %dh\n", report.Debt.GodObjectHours)))
	sb.WriteString(formatter.Info(fmt.Sprintf("Complexity: %dh\n", report.Debt.ComplexityHours)))
	sb.WriteString(formatter.Info(fmt.Sprintf("Total: %dh\n\n", report.Debt.TotalHours)))
}

// writeScoreBreakdownWithColor writes the score breakdown with colors
func writeScoreBreakdownWithColor(sb *strings.Builder, report *StructuralReport, formatter *ColorFormatter) {
	if !report.HasViolations {
		sb.WriteString(formatter.Success("✨ No structural violations detected! Your architecture is clean.") + "\n\n")
		return
	}

	sb.WriteString(formatter.Color("┌───────────────────────────────────────────────────────────┐", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("│  SCORE BREAKDOWN                                          │", ColorCyan))
	sb.WriteString("\n")
	sb.WriteString(formatter.Color("└───────────────────────────────────────────────────────────┘", ColorCyan))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("Base Score:           100.0\n"))
	sb.WriteString(fmt.Sprintf("Circular Penalty:     %s\n", formatter.Error(fmt.Sprintf("-%.1f (%d violations x 10.0)", report.Score.CircularPenalty, report.Score.CircularCount))))
	sb.WriteString(fmt.Sprintf("Layer Penalty:        %s\n", formatter.Warn(fmt.Sprintf("-%.1f (%d violations x 5.0)", report.Score.LayerPenalty, report.Score.LayerCount))))
	sb.WriteString(fmt.Sprintf("Size Penalty:         %s\n", formatter.Info(fmt.Sprintf("-%.1f (%d violations x 3.0)", report.Score.SizePenalty, report.Score.SizeCount))))
	sb.WriteString(fmt.Sprintf("God Object Penalty:   %s\n", formatter.Info(fmt.Sprintf("-%.1f (%d violations x 5.0)", report.Score.GodObjectPenalty, report.Score.GodObjectCount))))
	sb.WriteString(formatter.Color("─────────────────────────────────────────────────", ColorCyan) + "\n")
	sb.WriteString(fmt.Sprintf("Final Score:          %s\n\n", formatter.Bold(fmt.Sprintf("%.1f", report.Score.TotalScore))))
}
