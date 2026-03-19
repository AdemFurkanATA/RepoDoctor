package main

import (
	"os"
	"sort"

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

	for _, v := range violations {
		switch v.RuleID {
		case "rule.circular-dependency":
			report.Circular = append(report.Circular, CycleViolation{Path: []string{v.File}, Severity: string(v.Severity)})
		case "rule.layer-validation":
			report.Layer = append(report.Layer, LayerViolation{From: v.File, To: "", Message: v.Message})
		case "rule.size":
			report.Size = append(report.Size, SizeViolation{File: v.File, Function: "", Lines: 0, Threshold: 0})
		case "rule.god-object":
			report.GodObject = append(report.GodObject, GodObjectViolation{StructName: v.Message, File: v.File})
		}
	}

	report.HasViolations = len(violations) > 0
	report.Score = calculateScoreFromViolations(cfg, report)
	return report
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
