package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"RepoDoctor/internal/engine"
	"RepoDoctor/internal/model"
	"RepoDoctor/internal/rules"
)

type runtimeRuleSummary struct {
	result       *engine.ExecutionResult
	rulesInScope int
}

func runInternalRulePipeline(absPath string, graph Graph, primaryLanguage string) *runtimeRuleSummary {
	return runInternalRulePipelineWithProfile(absPath, graph, primaryLanguage, nil)
}

func runInternalRulePipelineWithProfile(absPath string, graph Graph, primaryLanguage string, architecture *ArchitectureConfig) *runtimeRuleSummary {
	profile := ""
	customLayerOrder := []string{}
	customLayerKeywords := map[string][]string{}
	if architecture != nil {
		profile = strings.TrimSpace(architecture.Profile)
		customLayerOrder = normalizeLayerOrder(architecture.CustomLayerOrder)
		customLayerKeywords = normalizeLayerKeywords(architecture.CustomLayerKeywords)
	}

	registry := rules.NewRuleRegistry()
	for _, rule := range rules.GetDefaultRegistry().GetAll() {
		registry.MustRegister(rule)
	}
	registry.MustRegister(rules.NewCircularDependencyRule(toRulesDependencyGraph(graph)))

	executor := engine.NewRuleExecutor(registry)
	context := buildUnifiedRulesAnalysisContext(runtimeAnalysisContextInput{
		RepositoryPath:      absPath,
		Graph:               graph,
		PrimaryLanguage:     primaryLanguage,
		ArchitectureProfile: profile,
		CustomLayerOrder:    customLayerOrder,
		CustomLayerKeywords: customLayerKeywords,
	})
	result := executor.Execute(context)
	sortViolations(result.Violations)

	return &runtimeRuleSummary{
		result:       result,
		rulesInScope: registry.Count(),
	}
}

func normalizeLayerOrder(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(raw))
	normalized := make([]string, 0, len(raw))
	for _, layer := range raw {
		value := strings.ToLower(strings.TrimSpace(layer))
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizeLayerKeywords(raw map[string][]string) map[string][]string {
	if len(raw) == 0 {
		return nil
	}
	normalized := make(map[string][]string, len(raw))
	for layer, aliases := range raw {
		normalizedLayer := strings.ToLower(strings.TrimSpace(layer))
		if normalizedLayer == "" {
			continue
		}
		seenAliases := map[string]bool{}
		normalizedAliases := make([]string, 0, len(aliases))
		for _, alias := range aliases {
			value := strings.ToLower(strings.TrimSpace(alias))
			if value == "" || seenAliases[value] {
				continue
			}
			seenAliases[value] = true
			normalizedAliases = append(normalizedAliases, value)
		}
		if len(normalizedAliases) > 0 {
			normalized[normalizedLayer] = normalizedAliases
		}
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func buildRulesAnalysisContext(absPath string, graph Graph, primaryLanguage string) rules.AnalysisContext {
	// Backward-compatible delegate retained for tests/legacy callers.
	return buildUnifiedRulesAnalysisContext(runtimeAnalysisContextInput{
		RepositoryPath:  absPath,
		Graph:           graph,
		PrimaryLanguage: primaryLanguage,
	})
}

func readRepositoryFileForContext(graph Graph, node string) rules.RepositoryFile {
	content := ""
	if data, err := os.ReadFile(node); err == nil {
		content = string(data)
	}

	return rules.RepositoryFile{
		Path:    node,
		Content: content,
		Imports: graph.GetDependencies(node),
	}
}

func legacyBuildRulesAnalysisContext(absPath string, graph Graph, primaryLanguage string) rules.AnalysisContext {
	nodes := graph.GetAllNodes()
	nodes = sortedStringCopy(nodes)

	repoFiles := make([]rules.RepositoryFile, 0, len(nodes))
	for _, node := range nodes {
		repoFiles = append(repoFiles, readRepositoryFileForContext(graph, node))
	}

	languages := []string{"Go", "Python", "JavaScript", "TypeScript", "Java"}
	if primaryLanguage != "" {
		languages = []string{primaryLanguage}
	}

	return rules.AnalysisContext{
		RepositoryFiles: repoFiles,
		DependencyGraph: toRulesDependencyGraph(graph),
		Configuration:   rules.Configuration{"repositoryPath": absPath},
		Languages:       languages,
	}
}

func toRulesDependencyGraph(graph Graph) rules.DependencyGraph {
	nodes := graph.GetAllNodes()
	nodes = sortedStringCopy(nodes)
	edges := make(map[string][]string, len(nodes))

	for _, node := range nodes {
		deps := sortedStringCopy(graph.GetDependencies(node))
		edges[node] = deps
	}

	return rules.DependencyGraph{Nodes: nodes, Edges: edges}
}

func sortViolations(violations []model.Violation) {
	sortModelViolationsDeterministic(violations)
}

func buildReportFromRuleViolations(path string, version string, cfg *Config, violations []model.Violation) *StructuralReport {
	report := &StructuralReport{Version: version, Path: path}

	// Accumulate god object violations by file+struct so field and method
	// violations for the same struct merge into a single report entry.
	godObjectMap := make(map[string]*GodObjectViolation)

	for _, v := range violations {
		v = applyConfiguredSeverity(v, cfg)
		switch v.RuleID {
		case "rule.circular-dependency":
			report.Circular = append(report.Circular, parseCircularViolation(v))
		case "rule.layer-validation":
			report.Layer = append(report.Layer, parseLayerViolation(v))
		case "rule.secret-detection":
			report.Layer = append(report.Layer, parseLayerViolation(v))
		case "rule.code-duplication":
			report.Size = append(report.Size, parseSizeViolation(v))
		case "rule.size":
			report.Size = append(report.Size, parseSizeViolation(v))
		case "rule.god-object":
			mergeGodObjectViolation(godObjectMap, v)
		}
	}

	for _, gov := range godObjectMap {
		report.GodObject = append(report.GodObject, *gov)
	}

	report.HasViolations = len(violations) > 0
	report.Score = calculateScoreFromViolations(cfg, report)
	return report
}

// Regex patterns for parsing violation messages produced by internal rules.
// Size:       "File <path> has <N> lines (threshold: <T>)"
//
//	"Function '<name>' has <N> lines (threshold: <T>)"
//
// GodObject:  "<Struct> has <N> fields (threshold: <T>)"
//
//	"<Struct> has <N> methods (threshold: <T>)"
var (
	sizeFileRe  = regexp.MustCompile(`has (\d+) lines \(threshold: (\d+)\)`)
	sizeFuncRe  = regexp.MustCompile(`^Function '([^']+)' has (\d+) lines \(threshold: (\d+)\)`)
	godFieldRe  = regexp.MustCompile(`^(.+) has (\d+) fields \(threshold: \d+\)`)
	godMethodRe = regexp.MustCompile(`^(.+) has (\d+) methods \(threshold: \d+\)`)
	layerMsgRe  = regexp.MustCompile(`^(.+?) \(([^)]+)\) -> (.+?) \(([^)]+)\): (.+)$`)
)

func parseCircularViolation(v model.Violation) CycleViolation {
	path := parseCyclePath(v.Message)
	if len(path) == 0 {
		path = []string{v.File}
	}
	return CycleViolation{Path: path, Severity: string(v.Severity), Hint: remediationHintForViolation(v)}
}

func parseCyclePath(message string) []string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "→")
	path := make([]string, 0, len(parts))
	for _, part := range parts {
		node := strings.TrimSpace(part)
		if node == "" {
			continue
		}
		path = append(path, node)
	}
	if len(path) > 1 && path[0] == path[len(path)-1] {
		path = path[:len(path)-1]
	}
	return path
}

func parseLayerViolation(v model.Violation) LayerViolation {
	lv := LayerViolation{From: v.File, To: "", Message: v.Message, Hint: remediationHintForViolation(v)}
	if m := layerMsgRe.FindStringSubmatch(v.Message); len(m) == 6 {
		fromPath := strings.TrimSpace(m[1])
		fromLayer := strings.TrimSpace(m[2])
		toPath := strings.TrimSpace(m[3])
		toLayer := strings.TrimSpace(m[4])
		reason := strings.TrimSpace(m[5])

		lv.From = fromPath
		lv.To = toPath
		lv.Message = fromLayer + " -> " + toLayer + ": " + reason
	}
	return lv
}

// parseSizeViolation extracts Lines, Threshold, and Function from a size
// violation message instead of using hardcoded placeholder values.
func parseSizeViolation(v model.Violation) SizeViolation {
	sv := SizeViolation{File: v.File, Hint: remediationHintForViolation(v)}

	// Try function-level match first (more specific)
	if m := sizeFuncRe.FindStringSubmatch(v.Message); len(m) == 4 {
		sv.Function = m[1]
		sv.Lines, _ = strconv.Atoi(m[2])
		sv.Threshold, _ = strconv.Atoi(m[3])
		return sv
	}

	// Fall back to file-level match
	if m := sizeFileRe.FindStringSubmatch(v.Message); len(m) == 3 {
		sv.Lines, _ = strconv.Atoi(m[1])
		sv.Threshold, _ = strconv.Atoi(m[2])
	}

	return sv
}

// mergeGodObjectViolation accumulates field and method counts for the same
// struct into a single GodObjectViolation entry keyed by file + struct name.
func mergeGodObjectViolation(m map[string]*GodObjectViolation, v model.Violation) {
	structName := ""
	fieldCount := 0
	methodCount := 0

	if match := godFieldRe.FindStringSubmatch(v.Message); len(match) == 3 {
		structName = match[1]
		fieldCount, _ = strconv.Atoi(match[2])
	} else if match := godMethodRe.FindStringSubmatch(v.Message); len(match) == 3 {
		structName = match[1]
		methodCount, _ = strconv.Atoi(match[2])
	} else {
		// Unrecognised format — preserve raw message as struct name
		structName = v.Message
	}

	key := v.File + "#" + structName
	if existing, ok := m[key]; ok {
		existing.FieldCount += fieldCount
		existing.MethodCount += methodCount
	} else {
		m[key] = &GodObjectViolation{
			StructName:  structName,
			File:        v.File,
			FieldCount:  fieldCount,
			MethodCount: methodCount,
			Hint:        remediationHintForViolation(v),
		}
	}
}

func remediationHintForViolation(v model.Violation) string {
	switch v.RuleID {
	case "rule.circular-dependency":
		return "Break the dependency cycle by moving shared contracts to a lower-level package and injecting dependencies inward."
	case "rule.layer-validation":
		return "Move the dependency to an allowed lower layer or introduce an interface boundary to preserve dependency direction."
	case "rule.size":
		return "Split oversized files/functions into focused units with one responsibility each."
	case "rule.god-object":
		return "Extract cohesive responsibilities into dedicated types and keep each object focused on one concern."
	case "rule.secret-detection":
		return "Move credentials to environment variables or secret stores and rotate any leaked keys immediately."
	case "rule.code-duplication":
		return "Extract shared logic into reusable helpers/components and keep one source of truth for duplicated blocks."
	default:
		return ""
	}
}

func applyConfiguredSeverity(v model.Violation, cfg *Config) model.Violation {
	if cfg == nil || cfg.Rules == nil {
		return v
	}
	severity := ""
	switch v.RuleID {
	case "rule.circular-dependency":
		severity = cfg.Rules.CircularSeverity
	case "rule.layer-validation":
		severity = cfg.Rules.LayerSeverity
	case "rule.secret-detection":
		severity = cfg.Rules.LayerSeverity
	case "rule.code-duplication":
		severity = cfg.Rules.SizeSeverity
	case "rule.size":
		severity = cfg.Rules.SizeSeverity
	case "rule.god-object":
		severity = cfg.Rules.GodObjectSeverity
	}
	parsed, ok := parseSeverity(severity)
	if !ok {
		return v
	}
	v.Severity = parsed
	return v
}

func parseSeverity(raw string) (model.Severity, bool) {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "info":
		return model.SeverityInfo, true
	case "warning":
		return model.SeverityWarning, true
	case "error":
		return model.SeverityError, true
	case "critical":
		return model.SeverityCritical, true
	default:
		return "", false
	}
}

func calculateScoreFromViolations(cfg *Config, report *StructuralReport) *StructuralScore {
	weights := DefaultScoringWeights()
	if cfg != nil && cfg.Weights != nil {
		weights.CircularDependencyPenalty = cfg.Weights.Circular
		weights.LayerViolationPenalty = cfg.Weights.Layer
		weights.SizeViolationPenalty = cfg.Weights.Size
		weights.GodObjectPenalty = cfg.Weights.GodObject
	}

	score := &StructuralScore{MaxScore: 100.0}
	score.CircularCount = len(report.Circular)
	score.LayerCount = len(report.Layer)
	score.SizeCount = len(report.Size)
	score.GodObjectCount = len(report.GodObject)

	score.CircularPenalty = float64(score.CircularCount) * weights.CircularDependencyPenalty
	score.LayerPenalty = float64(score.LayerCount) * weights.LayerViolationPenalty
	score.SizePenalty = float64(score.SizeCount) * weights.SizeViolationPenalty
	score.GodObjectPenalty = float64(score.GodObjectCount) * weights.GodObjectPenalty

	score.ViolationCount = score.CircularCount + score.LayerCount + score.SizeCount + score.GodObjectCount
	penalty := score.CircularPenalty + score.LayerPenalty + score.SizePenalty + score.GodObjectPenalty
	score.TotalScore = score.MaxScore - penalty
	if score.TotalScore < 0 {
		score.TotalScore = 0
	}

	return score
}
