package languages

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type importEvidence struct {
	modulePath        string
	relative          bool
	level             int
	unsupportedReason string
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
		result = append(result, parsePythonImportLine(scanner.Text())...)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func parsePythonImportLine(raw string) []importEvidence {
	line := strings.TrimSpace(raw)
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	if dynamic, reason := parsePythonDynamicImport(line); dynamic != "" || reason != "" {
		return []importEvidence{{modulePath: dynamic, unsupportedReason: reason}}
	}

	if strings.HasPrefix(line, "import ") {
		return parsePythonDirectImportLine(line)
	}

	if strings.HasPrefix(line, "from ") {
		return parsePythonFromImportLine(line)
	}

	return nil
}

func parsePythonDirectImportLine(line string) []importEvidence {
	entry := strings.TrimSpace(strings.TrimPrefix(line, "import "))
	parts := strings.Split(entry, ",")
	result := make([]importEvidence, 0, len(parts))
	for _, part := range parts {
		modulePath := strings.TrimSpace(strings.Split(strings.TrimSpace(part), " as ")[0])
		if modulePath == "" {
			continue
		}
		result = append(result, importEvidence{modulePath: modulePath})
	}
	return result
}

func parsePythonFromImportLine(line string) []importEvidence {
	remainder := strings.TrimSpace(strings.TrimPrefix(line, "from "))
	parts := strings.SplitN(remainder, " import ", 2)
	if len(parts) != 2 {
		return nil
	}
	fromModule := strings.TrimSpace(parts[0])
	targets := parsePythonImportTargets(parts[1])
	level := parsePythonRelativeLevel(fromModule)
	modulePath := strings.TrimPrefix(fromModule, strings.Repeat(".", level))

	if len(targets) == 0 {
		return []importEvidence{{
			modulePath: modulePath,
			relative:   level > 0,
			level:      level,
		}}
	}

	result := make([]importEvidence, 0, len(targets))
	for _, target := range targets {
		if target == "*" {
			continue
		}
		combined := modulePath
		if combined == "" {
			combined = target
		} else {
			combined = combined + "." + target
		}
		result = append(result, importEvidence{
			modulePath: combined,
			relative:   level > 0,
			level:      level,
		})
	}
	return result
}

func parsePythonRelativeLevel(fromModule string) int {
	level := 0
	for level < len(fromModule) && fromModule[level] == '.' {
		level++
	}
	return level
}

func parsePythonImportTargets(importPart string) []string {
	parts := strings.Split(importPart, ",")
	targets := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(strings.Split(strings.TrimSpace(part), " as ")[0])
		if trimmed == "" {
			continue
		}
		targets = append(targets, trimmed)
	}
	return targets
}

func parsePythonDynamicImport(line string) (string, string) {
	if !strings.Contains(line, "import_module(") && !strings.Contains(line, "__import__(") {
		return "", ""
	}

	for _, prefix := range []string{"importlib.import_module", "__import__"} {
		idx := strings.Index(line, prefix+"(")
		if idx < 0 {
			continue
		}
		args := line[idx+len(prefix)+1:]
		for _, quote := range []string{"'", "\""} {
			start := strings.Index(args, quote)
			if start < 0 {
				continue
			}
			end := strings.Index(args[start+1:], quote)
			if end < 0 {
				continue
			}
			candidate := strings.TrimSpace(args[start+1 : start+1+end])
			if candidate == "" || strings.HasPrefix(candidate, ".") {
				return "", "RELATIVE_DYNAMIC_IMPORT_UNSUPPORTED"
			}
			return candidate, ""
		}

		if strings.Contains(args, "+") || strings.Contains(args, "%") || strings.Contains(args, "format(") {
			return "", "DYNAMIC_IMPORT_EXPRESSION_UNSUPPORTED"
		}

		return "", "DYNAMIC_IMPORT_UNPARSEABLE"
	}

	return "", "DYNAMIC_IMPORT_UNPARSEABLE"
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
