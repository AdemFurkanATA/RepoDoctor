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

	return GoImportExternal
}
