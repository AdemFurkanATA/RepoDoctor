package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type packageImports struct {
	pkg     string
	imports []string
	files   []string
}

func TestArchitecture_InternalDependencyBoundaries(t *testing.T) {
	modulePath := readModulePath(t)
	root := projectRoot(t)
	pkgMap := collectPackageImports(t, root, modulePath)

	violations := make([]string, 0)
	for _, pkg := range sortedPackageKeys(pkgMap) {
		info := pkgMap[pkg]
		for _, imp := range info.imports {
			if !strings.HasPrefix(imp, modulePath) {
				continue
			}

			if violation, ok := boundaryViolation(pkg, imp, modulePath); ok {
				violations = append(violations, violation)
			}
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("architecture dependency boundary violations:\n%s", strings.Join(violations, "\n"))
	}
}

func TestArchitecture_HighRiskImportsForbiddenInModelAndDomain(t *testing.T) {
	modulePath := readModulePath(t)
	root := projectRoot(t)
	pkgMap := collectPackageImports(t, root, modulePath)

	banned := map[string]struct{}{"os/exec": {}, "unsafe": {}}
	violations := make([]string, 0)

	for _, pkg := range sortedPackageKeys(pkgMap) {
		if pkg != modulePath+"/internal/model" && pkg != modulePath+"/internal/domain" {
			continue
		}

		for _, imp := range pkgMap[pkg].imports {
			if _, blocked := banned[imp]; blocked {
				violations = append(violations, fmt.Sprintf("%s imports forbidden package %q", pkg, imp))
			}
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("security boundary violations:\n%s", strings.Join(violations, "\n"))
	}
}

func boundaryViolation(fromPkg, toPkg, modulePath string) (string, bool) {
	mainPkg := modulePath
	analysisPkg := modulePath + "/internal/analysis"
	languagesPkg := modulePath + "/internal/languages"
	rulesPkg := modulePath + "/internal/rules"
	enginePkg := modulePath + "/internal/engine"
	modelPkg := modulePath + "/internal/model"
	domainPkg := modulePath + "/internal/domain"

	if !strings.HasPrefix(fromPkg, modulePath+"/internal/") {
		return "", false
	}

	if toPkg == mainPkg {
		return fmt.Sprintf("%s must not import main package %s", fromPkg, toPkg), true
	}

	if strings.HasPrefix(fromPkg, modelPkg) {
		if toPkg == analysisPkg || toPkg == languagesPkg || toPkg == rulesPkg || toPkg == enginePkg {
			return fmt.Sprintf("%s must not import higher-level package %s", fromPkg, toPkg), true
		}
		return "", false
	}

	if fromPkg == analysisPkg {
		if toPkg != languagesPkg && toPkg != modelPkg && toPkg != domainPkg {
			return fmt.Sprintf("%s imports forbidden internal dependency %s", fromPkg, toPkg), true
		}
		return "", false
	}

	if fromPkg == rulesPkg {
		if toPkg != modelPkg && toPkg != domainPkg {
			return fmt.Sprintf("%s imports forbidden internal dependency %s", fromPkg, toPkg), true
		}
		return "", false
	}

	if fromPkg == enginePkg {
		// Engine may depend on rules abstractions and core model/domain.
		if toPkg != rulesPkg && toPkg != modelPkg && toPkg != domainPkg {
			return fmt.Sprintf("%s imports forbidden internal dependency %s", fromPkg, toPkg), true
		}
		return "", false
	}

	if fromPkg == languagesPkg {
		if toPkg != modelPkg && toPkg != domainPkg {
			return fmt.Sprintf("%s imports forbidden internal dependency %s", fromPkg, toPkg), true
		}
		return "", false
	}

	return "", false
}

func collectPackageImports(t *testing.T, root, modulePath string) map[string]packageImports {
	t.Helper()

	fset := token.NewFileSet()
	pkgMap := make(map[string]packageImports)

	goFiles := collectGoFiles(t, root)
	for _, file := range goFiles {
		relDir, pkgPath := packagePathForFile(t, root, modulePath, file)

		parsed, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("failed to parse imports for %s: %v", file, err)
		}

		entry := pkgMap[pkgPath]
		entry.pkg = pkgPath
		entry.files = append(entry.files, filepath.ToSlash(filepath.Join(relDir, filepath.Base(file))))
		for _, imp := range parsed.Imports {
			entry.imports = append(entry.imports, strings.Trim(imp.Path.Value, "\""))
		}
		pkgMap[pkgPath] = entry
	}

	for pkg, info := range pkgMap {
		sort.Strings(info.files)
		sort.Strings(info.imports)
		info.imports = compactSorted(info.imports)
		pkgMap[pkg] = info
	}

	return pkgMap
}

func collectGoFiles(t *testing.T, root string) []string {
	t.Helper()

	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()
		if d.IsDir() {
			if name == ".git" || name == ".repodoctor" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to scan repository files: %v", err)
	}

	sort.Strings(files)
	return files
}

func packagePathForFile(t *testing.T, root, modulePath, file string) (string, string) {
	t.Helper()

	dir := filepath.Dir(file)
	relDir, err := filepath.Rel(root, dir)
	if err != nil {
		t.Fatalf("failed to resolve package path for %s: %v", file, err)
	}

	if relDir == "." {
		return "", modulePath
	}

	relDir = filepath.ToSlash(relDir)
	return relDir, modulePath + "/" + relDir
}

func sortedPackageKeys(pkgMap map[string]packageImports) []string {
	keys := make([]string, 0, len(pkgMap))
	for key := range pkgMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func compactSorted(values []string) []string {
	if len(values) == 0 {
		return values
	}
	out := values[:1]
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1] {
			out = append(out, values[i])
		}
	}
	return out
}

func readModulePath(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") {
			modulePath := strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
			if modulePath == "" {
				break
			}
			return modulePath
		}
	}

	t.Fatal("module path not found in go.mod")
	return ""
}

func projectRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to resolve working directory: %v", err)
	}
	return wd
}
