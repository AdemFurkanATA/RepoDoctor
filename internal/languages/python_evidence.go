package languages

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type importEvidence struct {
	modulePath string
	relative   bool
	level      int
}

func parsePythonImportEvidence(path string) ([]importEvidence, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make([]importEvidence, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "import ") {
			entry := strings.TrimSpace(strings.TrimPrefix(line, "import "))
			for _, part := range strings.Split(entry, ",") {
				modulePath := strings.TrimSpace(strings.Split(strings.TrimSpace(part), " as ")[0])
				if modulePath == "" {
					continue
				}
				result = append(result, importEvidence{modulePath: modulePath})
			}
			continue
		}

		if strings.HasPrefix(line, "from ") {
			remainder := strings.TrimSpace(strings.TrimPrefix(line, "from "))
			parts := strings.SplitN(remainder, " import ", 2)
			if len(parts) != 2 {
				continue
			}
			fromModule := strings.TrimSpace(parts[0])
			level := 0
			for level < len(fromModule) && fromModule[level] == '.' {
				level++
			}
			modulePath := strings.TrimPrefix(fromModule, strings.Repeat(".", level))
			result = append(result, importEvidence{
				modulePath: modulePath,
				relative:   level > 0,
				level:      level,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func detectPythonModuleRoot(repoRoot, filePath string) string {
	current := filepath.Dir(filePath)
	moduleRoot := current

	for {
		candidate := filepath.Join(current, "__init__.py")
		if _, err := os.Stat(candidate); err != nil {
			break
		}
		moduleRoot = current
		if current == repoRoot {
			break
		}
		next := filepath.Dir(current)
		if next == current {
			break
		}
		current = next
	}

	return moduleRoot
}

func normalizePythonImport(item importEvidence, filePath, repoRoot, moduleRoot string) string {
	if !item.relative {
		normalized := strings.TrimSpace(item.modulePath)
		if normalized == "" {
			return ""
		}
		parts := strings.Split(normalized, ".")
		return strings.TrimSpace(parts[0])
	}

	base := filepath.Dir(filePath)
	for i := 1; i < item.level; i++ {
		next := filepath.Dir(base)
		if next == base || len(next) < len(moduleRoot) {
			break
		}
		base = next
	}

	rel, err := filepath.Rel(repoRoot, base)
	if err != nil {
		return ""
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "../") || rel == ".." {
		return ""
	}

	if strings.TrimSpace(item.modulePath) != "" {
		if rel == "." {
			rel = item.modulePath
		} else {
			rel = rel + "." + item.modulePath
		}
	}

	if rel == "." || rel == "" {
		return ""
	}
	root := strings.Split(strings.ReplaceAll(rel, "/", "."), ".")[0]
	return strings.TrimSpace(root)
}
