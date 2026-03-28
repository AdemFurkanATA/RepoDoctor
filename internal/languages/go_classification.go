package languages

import "strings"

// GoImportClass defines deterministic import edge categories.
type GoImportClass string

const (
	GoImportStdlib   GoImportClass = "stdlib"
	GoImportInternal GoImportClass = "internal"
	GoImportExternal GoImportClass = "third_party"
)

func classifyGoImport(importPath string) GoImportClass {
	return classifyGoImportWithModule(importPath, "")
}

func classifyGoImportWithModule(importPath, modulePath string) GoImportClass {
	normalized := strings.TrimSpace(importPath)
	if normalized == "" {
		return GoImportExternal
	}

	if !strings.Contains(normalized, ".") {
		return GoImportStdlib
	}

	if strings.Contains(normalized, "/internal/") || strings.HasSuffix(normalized, "/internal") {
		return GoImportInternal
	}

	modulePath = strings.TrimSpace(modulePath)
	if modulePath != "" {
		if normalized == modulePath || strings.HasPrefix(normalized, modulePath+"/") {
			return GoImportInternal
		}
	}

	return GoImportExternal
}
