package languages

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"RepoDoctor/internal/model"
)

const (
	maxJavaFileBytes = 2 * 1024 * 1024
)

var (
	javaPackagePattern = regexp.MustCompile(`^\s*package\s+([A-Za-z_][\w.]*)\s*;`)
	javaImportPattern  = regexp.MustCompile(`^\s*import\s+(?:static\s+)?([A-Za-z_][\w.*]*)\s*;`)
	javaMethodPattern  = regexp.MustCompile(`\b(?:public|protected|private)?\s*(?:static\s+)?(?:final\s+)?[A-Za-z_][\w<>\[\]]*\s+([A-Za-z_][\w]*)\s*\(`)
)

type JavaAdapter struct{}

func NewJavaAdapter() LanguageAdapter {
	return &JavaAdapter{}
}

func (a *JavaAdapter) Name() string {
	return "Java"
}

func (a *JavaAdapter) FileExtensions() []string {
	return []string{".java"}
}

func (a *JavaAdapter) DetectFiles(repoPath string) ([]string, error) {
	javaFiles := make([]string, 0)
	err := filepath.WalkDir(repoPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := strings.ToLower(d.Name())
			if strings.HasPrefix(name, ".") || name == "target" || name == "build" || name == "out" || name == ".gradle" {
				if path != repoPath {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) == ".java" {
			javaFiles = append(javaFiles, filepath.Clean(path))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(javaFiles)
	return javaFiles, nil
}

func (a *JavaAdapter) CollectMetrics(files []string) (*model.RepositoryMetrics, error) {
	metrics := model.NewRepositoryMetrics()
	for _, file := range files {
		fm, err := a.collectFileMetrics(file)
		if err != nil {
			continue
		}
		metrics.AddFileMetrics(*fm)
	}
	return metrics, nil
}

func (a *JavaAdapter) collectFileMetrics(path string) (*model.FileMetrics, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(content) > maxJavaFileBytes {
		return nil, fmt.Errorf("java file too large for safe parsing: %s", path)
	}

	lines := strings.Split(string(content), "\n")
	functions := 0
	imports := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") {
			continue
		}
		if javaImportPattern.MatchString(trimmed) {
			imports++
		}
		if javaMethodPattern.MatchString(trimmed) {
			functions++
		}
	}

	return &model.FileMetrics{
		Path:      path,
		Lines:     len(lines),
		Functions: functions,
		Imports:   imports,
	}, nil
}

func (a *JavaAdapter) BuildDependencyGraph(files []string) (*model.DependencyGraph, error) {
	graph := model.NewDependencyGraph()
	for _, file := range files {
		pkgName, imports, err := a.extractFilePackageAndImports(file)
		if err != nil {
			continue
		}
		node := graph.AddNode(file, file, pkgName)
		for _, imp := range imports {
			normalized := a.NormalizeImport(imp)
			if normalized == "" {
				continue
			}
			node.Imports = append(node.Imports, normalized)
			if node.Metadata == nil {
				node.Metadata = make(map[string]string)
			}
			node.Metadata["import_class:"+normalized] = "internal"
			graph.AddEdge(file, normalized)
		}
	}
	return graph, nil
}

func (a *JavaAdapter) extractFilePackageAndImports(path string) (string, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	pkg := ""
	imports := make([]string, 0, 8)
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.ContainsRune(line, '\x00') {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if pkg == "" {
			if m := javaPackagePattern.FindStringSubmatch(trimmed); len(m) == 2 {
				pkg = m[1]
			}
		}
		if m := javaImportPattern.FindStringSubmatch(trimmed); len(m) == 2 {
			imports = append(imports, m[1])
		}
	}
	if err := scanner.Err(); err != nil {
		return "", nil, err
	}
	sort.Strings(imports)
	return pkg, imports, nil
}

func (a *JavaAdapter) IsStdlibPackage(importPath string) bool {
	normalized := strings.TrimSpace(importPath)
	return strings.HasPrefix(normalized, "java.") || strings.HasPrefix(normalized, "javax.") || strings.HasPrefix(normalized, "jdk.") || strings.HasPrefix(normalized, "sun.")
}

func (a *JavaAdapter) Capabilities() AdapterCapabilities {
	return AdapterCapabilities{
		SupportsDependencyGraph: true,
		SupportsMetrics:         true,
		UsesASTParsing:          false,
	}
}

func (a *JavaAdapter) NormalizeImport(importPath string) string {
	normalized := strings.TrimSpace(importPath)
	normalized = strings.TrimSuffix(normalized, ";")
	normalized = strings.TrimPrefix(normalized, "static ")
	normalized = strings.TrimSuffix(normalized, ".*")
	return normalized
}
