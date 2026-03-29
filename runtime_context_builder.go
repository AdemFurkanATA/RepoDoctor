package main

import (
	"RepoDoctor/internal/rules"
)

type runtimeAnalysisContextInput struct {
	RepositoryPath      string
	Graph               Graph
	PrimaryLanguage     string
	ArchitectureProfile string
	CustomLayerOrder    []string
	CustomLayerKeywords map[string][]string
}

func buildUnifiedRulesAnalysisContext(input runtimeAnalysisContextInput) rules.AnalysisContext {
	repositoryFiles := buildContextRepositoryFiles(input.Graph)
	languages := resolveContextLanguages(input.PrimaryLanguage)

	configuration := rules.Configuration{
		"repositoryPath":      input.RepositoryPath,
		"architectureProfile": input.ArchitectureProfile,
	}
	if len(input.CustomLayerOrder) > 0 {
		configuration["customLayerOrder"] = append([]string{}, input.CustomLayerOrder...)
	}
	if len(input.CustomLayerKeywords) > 0 {
		copiedKeywords := make(map[string][]string, len(input.CustomLayerKeywords))
		for layer, aliases := range input.CustomLayerKeywords {
			copiedKeywords[layer] = append([]string{}, aliases...)
		}
		configuration["customLayerKeywords"] = copiedKeywords
	}

	return rules.AnalysisContext{
		RepositoryFiles: repositoryFiles,
		DependencyGraph: toRulesDependencyGraph(input.Graph),
		Configuration:   configuration,
		Languages:       languages,
	}
}

func buildContextRepositoryFiles(graph Graph) []rules.RepositoryFile {
	nodes := graph.GetAllNodes()
	nodes = sortedStringCopy(nodes)

	repositoryFiles := make([]rules.RepositoryFile, 0, len(nodes))
	for _, node := range nodes {
		repositoryFiles = append(repositoryFiles, readRepositoryFileForContext(graph, node))
	}

	return repositoryFiles
}

func resolveContextLanguages(primaryLanguage string) []string {
	if primaryLanguage != "" {
		return []string{primaryLanguage}
	}
	return []string{"Go", "Python", "JavaScript", "TypeScript", "Java"}
}
