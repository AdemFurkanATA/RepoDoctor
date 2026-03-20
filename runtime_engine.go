package main

import (
	"os"
	"regexp"
	"sort"
	"strconv"

	"RepoDoctor/internal/engine"
	"RepoDoctor/internal/model"
	"RepoDoctor/internal/rules"
)

type runtimeRuleSummary struct {
	result       *engine.ExecutionResult
	rulesInScope int
}

func runInternalRulePipeline(absPath string, graph Graph) *runtimeRuleSummary {
	registry := rules.NewRuleRegistry()
	for _, rule := range rules.GetDefaultRegistry().GetAll() {
		registry.MustRegister(rule)
	}
	registry.MustRegister(rules.NewCircularDependencyRule(toRulesDependencyGraph(graph)))

	executor := engine.NewRuleExecutor(registry)
	context := buildRulesAnalysisContext(absPath, graph)
	result := executor.Execute(context)
	sortViolations(result.Violations)

	return &runtimeRuleSummary{
		result:       result,
		rulesInScope: registry.Count(),
	}
}

func buildRulesAnalysisContext(absPath string, graph Graph) rules.AnalysisContext {
	nodes := graph.GetAllNodes()
	sort.Strings(nodes)

	repoFiles := make([]rules.RepositoryFile, 0, len(nodes))
	for _, node := range nodes {
		content := ""
		if data, err := os.ReadFile(node); err == nil {
			content = string(data)
		}

		repoFiles = append(repoFiles, rules.RepositoryFile{
			Path:    node,
			Content: content,
			Imports: graph.GetDependencies(node),
		})
	}

	return rules.AnalysisContext{
		RepositoryFiles: repoFiles,
		DependencyGraph: toRulesDependencyGraph(graph),
		Configuration:   rules.Configuration{"repositoryPath": absPath},
		Languages:       []string{"Go", "Python", "JavaScript", "TypeScript"},
	}
}

func toRulesDependencyGraph(graph Graph) rules.DependencyGraph {
	nodes := graph.GetAllNodes()
	sort.Strings(nodes)
	edges := make(map[string][]string, len(nodes))

	for _, node := range nodes {
		deps := append([]string(nil), graph.GetDependencies(node)...)
		sort.Strings(deps)
		edges[node] = deps
	}

	return rules.DependencyGraph{Nodes: nodes, Edges: edges}
}

func sortViolations(violations []model.Violation) {
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].RuleID != violations[j].RuleID {
			return violations[i].RuleID < violations[j].RuleID
		}
		if violations[i].File != violations[j].File {
			return violations[i].File < violations[j].File
		}
		if violations[i].Line != violations[j].Line {
			return violations[i].Line < violations[j].Line
		}
		return violations[i].Message < violations[j].Message
	})
}

func buildReportFromRuleViolations(path string, version string, cfg *Config, violations []model.Violation) *StructuralReport {
	report := &StructuralReport{Version: version, Path: path}

	// Accumulate god object violations by file+struct so field and method
	// violations for the same struct merge into a single report entry.
	godObjectMap := make(map[string]*GodObjectViolation)

	for _, v := range violations {
		switch v.RuleID {
		case "rule.circular-dependency":
			report.Circular = append(report.Circular, CycleViolation{Path: []string{v.File}, Severity: string(v.Severity)})
		case "rule.layer-validation":
			report.Layer = append(report.Layer, LayerViolation{From: v.File, To: "", Message: v.Message})
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
)

// parseSizeViolation extracts Lines, Threshold, and Function from a size
// violation message instead of using hardcoded placeholder values.
func parseSizeViolation(v model.Violation) SizeViolation {
	sv := SizeViolation{File: v.File}

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
		}
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
