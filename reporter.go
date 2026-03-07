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

// ColoredReporter extends Reporter with colored output support
type ColoredReporter struct {
	*Reporter
	formatter *ColorFormatter
}

// NewColoredReporter creates a new reporter with colored output
func NewColoredReporter(format OutputFormat, colorEnabled bool) *ColoredReporter {
	return &ColoredReporter{
		Reporter:  NewReporter(format),
		formatter: NewColorFormatter(colorEnabled),
	}
}

// StructuralReport represents the complete analysis report
type StructuralReport struct {
	Version       string
	Path          string
	Score         *StructuralScore
	Circular      []CycleViolation
	Layer         []LayerViolation
	Size          []SizeViolation
	GodObject     []GodObjectViolation
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
		Version:       version,
		Path:          path,
		Score:         scorer.CalculateScore(),
		Circular:      violations.Circular,
		Layer:         violations.Layer,
		Size:          violations.Size,
		GodObject:     violations.GodObject,
		HasViolations: len(violations.Circular) > 0 || len(violations.Layer) > 0 || len(violations.Size) > 0 || len(violations.GodObject) > 0,
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

	writeHeader(&sb)
	writeScoreSection(&sb, report)
	writeViolationsSummary(&sb, report)
	writeCircularViolations(&sb, report)
	writeLayerViolations(&sb, report)
	writeSizeViolations(&sb, report)
	writeGodObjectViolations(&sb, report)
	writeScoreBreakdown(&sb, report)

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
	
	r.formatScoreSection(&sb, report)
	r.formatViolationsSection(&sb, report)
	r.formatCircularViolations(&sb, report)
	r.formatLayerViolations(&sb, report)
	r.formatSizeViolations(&sb, report)
	r.formatGodObjectViolations(&sb, report)
	
	sb.WriteString("}\n")

	return sb.String()
}

// formatScoreSection formats the score section of JSON output
func (r *Reporter) formatScoreSection(sb *strings.Builder, report *StructuralReport) {
	sb.WriteString("  \"score\": {\n")
	sb.WriteString(fmt.Sprintf("    \"total\": %.2f,\n", report.Score.TotalScore))
	sb.WriteString(fmt.Sprintf("    \"max\": %.2f,\n", report.Score.MaxScore))
	sb.WriteString(fmt.Sprintf("    \"circularPenalty\": %.2f,\n", report.Score.CircularPenalty))
	sb.WriteString(fmt.Sprintf("    \"layerPenalty\": %.2f,\n", report.Score.LayerPenalty))
	sb.WriteString(fmt.Sprintf("    \"sizePenalty\": %.2f,\n", report.Score.SizePenalty))
	sb.WriteString(fmt.Sprintf("    \"godObjectPenalty\": %.2f\n", report.Score.GodObjectPenalty))
	sb.WriteString("  },\n")
}

// formatViolationsSection formats the violations summary section
func (r *Reporter) formatViolationsSection(sb *strings.Builder, report *StructuralReport) {
	sb.WriteString("  \"violations\": {\n")
	sb.WriteString(fmt.Sprintf("    \"circular\": %d,\n", report.Score.CircularCount))
	sb.WriteString(fmt.Sprintf("    \"layer\": %d,\n", report.Score.LayerCount))
	sb.WriteString(fmt.Sprintf("    \"size\": %d,\n", report.Score.SizeCount))
	sb.WriteString(fmt.Sprintf("    \"godObject\": %d\n", report.Score.GodObjectCount))
	sb.WriteString("  },\n")
}

// formatCircularViolations formats circular dependency violations
func (r *Reporter) formatCircularViolations(sb *strings.Builder, report *StructuralReport) {
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
}

// formatLayerViolations formats layer violations
func (r *Reporter) formatLayerViolations(sb *strings.Builder, report *StructuralReport) {
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
}

// formatSizeViolations formats size violations
func (r *Reporter) formatSizeViolations(sb *strings.Builder, report *StructuralReport) {
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
	sb.WriteString("  ],\n")
}

// formatGodObjectViolations formats god object violations
func (r *Reporter) formatGodObjectViolations(sb *strings.Builder, report *StructuralReport) {
	sb.WriteString("  \"godObjectViolations\": [\n")
	for i, v := range report.GodObject {
		sb.WriteString("    {\n")
		sb.WriteString(fmt.Sprintf("      \"struct\": \"%s\",\n", v.StructName))
		sb.WriteString(fmt.Sprintf("      \"file\": \"%s\",\n", v.File))
		sb.WriteString(fmt.Sprintf("      \"fields\": %d,\n", v.FieldCount))
		sb.WriteString(fmt.Sprintf("      \"methods\": %d,\n", v.MethodCount))
		sb.WriteString("    }")
		if i < len(report.GodObject)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("  ]\n")
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
