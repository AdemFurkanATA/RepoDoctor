package languages

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"RepoDoctor/internal/model"
)

const maxMetadataBytes = 2 * 1024 * 1024
const maxMetadataDepth = 8

type metadataCacheEntry struct {
	packageJSON bool
	tsConfig    bool
	warning     string
}

type jsTsAdapter struct {
	name       string
	extensions []string
	cache      map[string]metadataCacheEntry
}

func NewJavaScriptAdapter() LanguageAdapter {
	return &jsTsAdapter{name: "JavaScript", extensions: []string{".js", ".jsx"}, cache: map[string]metadataCacheEntry{}}
}

func NewTypeScriptAdapter() LanguageAdapter {
	return &jsTsAdapter{name: "TypeScript", extensions: []string{".ts", ".tsx"}, cache: map[string]metadataCacheEntry{}}
}

func (a *jsTsAdapter) Name() string { return a.name }

func (a *jsTsAdapter) FileExtensions() []string { return append([]string(nil), a.extensions...) }

func (a *jsTsAdapter) DetectFiles(repoPath string) ([]string, error) {
	root, err := normalizeRepoRoot(repoPath)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		normalizedPath, ok := normalizePathWithinRoot(root, path)
		if !ok {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if normalizedPath == root {
				return nil
			}
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(normalizedPath))
		for _, supported := range a.extensions {
			if ext == supported {
				files = append(files, normalizedPath)
				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

func (a *jsTsAdapter) CollectMetrics(files []string) (*model.RepositoryMetrics, error) {
	metrics := model.NewRepositoryMetrics()
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		lines := strings.Count(string(content), "\n") + 1
		imports := strings.Count(string(content), "import ") + strings.Count(string(content), "require(")
		metrics.AddFileMetrics(model.FileMetrics{Path: file, Lines: lines, Imports: imports})
	}
	return metrics, nil
}

func (a *jsTsAdapter) BuildDependencyGraph(files []string) (*model.DependencyGraph, error) {
	graph := model.NewDependencyGraph()
	for _, file := range files {
		node := graph.AddNode(file, file, filepath.Base(filepath.Dir(file)))
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "import ") && strings.Contains(trimmed, " from ") {
				parts := strings.Split(trimmed, " from ")
				candidate := strings.Trim(parts[len(parts)-1], " ';\"")
				if candidate != "" {
					normalized := a.NormalizeImport(candidate)
					node.Imports = append(node.Imports, normalized)
					graph.AddEdge(file, normalized)
				}
			}
		}
	}
	return graph, nil
}

func (a *jsTsAdapter) IsStdlibPackage(importPath string) bool {
	stdlib := map[string]bool{"fs": true, "path": true, "http": true, "https": true, "url": true, "os": true, "util": true}
	return stdlib[a.NormalizeImport(importPath)]
}

func (a *jsTsAdapter) Capabilities() AdapterCapabilities {
	return AdapterCapabilities{SupportsDependencyGraph: true, SupportsMetrics: true, UsesASTParsing: false}
}

func (a *jsTsAdapter) NormalizeImport(importPath string) string {
	trimmed := strings.TrimSpace(importPath)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.TrimPrefix(trimmed, "node:")
	if strings.HasPrefix(trimmed, "./") || strings.HasPrefix(trimmed, "../") {
		return strings.TrimSpace(trimmed)
	}
	parts := strings.Split(trimmed, "/")
	if strings.HasPrefix(trimmed, "@") && len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

func (a *jsTsAdapter) CollectEvidence(repoPath string, files []string) ([]EvidenceSignal, []string, error) {
	root, err := normalizeRepoRoot(repoPath)
	if err != nil {
		return nil, nil, err
	}
	signals := make([]EvidenceSignal, 0)
	warnings := make([]string, 0)

	for _, file := range files {
		normalizedPath, ok := normalizePathWithinRoot(root, file)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("skipped path outside repo root: %s", file))
			continue
		}
		signals = append(signals, EvidenceSignal{Language: a.name, SignalType: "source_file", WeightInput: 1.2, SourcePath: normalizedPath})

		dir := filepath.Dir(normalizedPath)
		entry, warn := a.loadMetadataOnce(root, dir)
		if warn != "" {
			warnings = append(warnings, warn)
		}
		if entry.packageJSON {
			signals = append(signals, EvidenceSignal{Language: a.name, SignalType: "package_json", WeightInput: 1.0, SourcePath: dir})
		}
		if a.name == "TypeScript" && entry.tsConfig {
			signals = append(signals, EvidenceSignal{Language: a.name, SignalType: "tsconfig", WeightInput: 1.5, SourcePath: dir})
		}
	}

	sort.SliceStable(signals, func(i, j int) bool {
		if signals[i].SourcePath != signals[j].SourcePath {
			return signals[i].SourcePath < signals[j].SourcePath
		}
		if signals[i].SignalType != signals[j].SignalType {
			return signals[i].SignalType < signals[j].SignalType
		}
		return signals[i].Language < signals[j].Language
	})

	return signals, warnings, nil
}

func (a *jsTsAdapter) loadMetadataOnce(root, dir string) (metadataCacheEntry, string) {
	if entry, ok := a.cache[dir]; ok {
		return entry, entry.warning
	}

	entry := metadataCacheEntry{}
	warning := ""

	if ok, warn := validateMetadataFile(filepath.Join(dir, "package.json")); ok {
		entry.packageJSON = true
		warning = warn
	} else if warn != "" {
		warning = warn
	}

	if ok, warn := validateMetadataFile(filepath.Join(dir, "tsconfig.json")); ok {
		entry.tsConfig = true
		if warning == "" {
			warning = warn
		}
	} else if warning == "" && warn != "" {
		warning = warn
	}

	if rel, err := filepath.Rel(root, dir); err == nil && strings.HasPrefix(rel, "..") {
		warning = "metadata outside root skipped"
		entry = metadataCacheEntry{}
	}

	entry.warning = warning
	a.cache[dir] = entry
	return entry, warning
}

func validateMetadataFile(path string) (bool, string) {
	info, err := os.Stat(path)
	if err != nil {
		return false, ""
	}
	if info.Size() > maxMetadataBytes {
		return false, fmt.Sprintf("metadata too large: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Sprintf("cannot read metadata: %s", path)
	}

	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return false, fmt.Sprintf("invalid metadata json: %s", path)
	}
	if depthExceeds(parsed, maxMetadataDepth, 0) {
		return false, fmt.Sprintf("metadata nested too deeply: %s", path)
	}

	obj, ok := parsed.(map[string]interface{})
	if !ok {
		return false, fmt.Sprintf("metadata root must be object: %s", path)
	}
	for key := range obj {
		if len(key) > 128 {
			return false, fmt.Sprintf("metadata key too long: %s", path)
		}
	}
	return true, ""
}

func depthExceeds(v interface{}, limit, depth int) bool {
	if depth > limit {
		return true
	}
	switch typed := v.(type) {
	case map[string]interface{}:
		for _, inner := range typed {
			if depthExceeds(inner, limit, depth+1) {
				return true
			}
		}
	case []interface{}:
		for _, inner := range typed {
			if depthExceeds(inner, limit, depth+1) {
				return true
			}
		}
	}
	return false
}
