package main

import (
	"RepoDoctor/internal/rules"
)

type runtimeAnalysisContextInput struct {
	RepositoryPath      string
	Graph               Graph
	PrimaryLanguage     string
	ArchitectureProfile string
}

func buildUnifiedRulesAnalysisContext(input runtimeAnalysisContextInput) rules.AnalysisContext {
	repositoryFiles := buildContextRepositoryFiles(input.Graph)
	languages := resolveContextLanguages(input.PrimaryLanguage)

	return rules.AnalysisContext{
		RepositoryFiles: repositoryFiles,
		DependencyGraph: toRulesDependencyGraph(input.Graph),
		Configuration:   rules.Configuration{"repositoryPath": input.RepositoryPath, "architectureProfile": input.ArchitectureProfile},
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
