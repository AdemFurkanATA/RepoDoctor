package analysis

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"RepoDoctor/internal/model"
)

const (
	apiStabilityBaselineEnv        = "REPODOCTOR_API_STABILITY_BASELINE"
	defaultAPIBaselineRelativePath = ".repodoctor/api-baseline/go-public-api.json"
	apiStabilitySnapshotSchema     = "v1"
	apiBreakTypeRemoved            = "removed"
	apiBreakTypeKindChanged        = "kind-changed"
	apiBreakTypeSignatureChanged   = "signature-changed"
)

// ComputeAPIBreakingChanges compares Go public API snapshots and returns
// deterministic breaking changes. This function is fail-soft and reports
// warnings for missing baseline or parse issues.
func ComputeAPIBreakingChanges(repoPath, adapterName string, files []string) ([]model.APIBreakingChange, []string) {
	if strings.TrimSpace(adapterName) != "Go" {
		return nil, []string{"api stability analysis skipped: adapter does not support public snapshot (current: " + strings.TrimSpace(adapterName) + ")"}
	}

	baselinePath := resolveAPIBaselinePath(repoPath)
	baseline, baselineWarnings := loadAPIBaselineSnapshot(baselinePath)
	if baseline == nil {
		return nil, baselineWarnings
	}

	current, currentWarnings := buildGoPublicAPISnapshot(repoPath, files)
	if current == nil {
		warnings := append([]string{}, baselineWarnings...)
		warnings = append(warnings, currentWarnings...)
		warnings = append(warnings, "api stability analysis skipped: could not build current public API snapshot")
		return nil, warnings
	}

	breaking := diffAPISnapshots(baseline, current)
	warnings := append([]string{}, baselineWarnings...)
	warnings = append(warnings, currentWarnings...)
	return breaking, warnings
}

func resolveAPIBaselinePath(repoPath string) string {
	raw := strings.TrimSpace(os.Getenv(apiStabilityBaselineEnv))
	if raw == "" {
		raw = defaultAPIBaselineRelativePath
	}
	if filepath.IsAbs(raw) {
		return filepath.Clean(raw)
	}
	return filepath.Join(repoPath, raw)
}

func loadAPIBaselineSnapshot(path string) (*model.APISnapshot, []string) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, []string{"api stability baseline not found: " + filepath.ToSlash(path)}
		}
		return nil, []string{"api stability baseline read failed: " + err.Error()}
	}

	var snapshot model.APISnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, []string{"api stability baseline parse failed: " + err.Error()}
	}
	if strings.TrimSpace(snapshot.Language) != "Go" {
		return nil, []string{"api stability baseline skipped: unsupported language " + strings.TrimSpace(snapshot.Language)}
	}

	stabilizeSnapshotOrdering(&snapshot)
	return &snapshot, nil
}

func buildGoPublicAPISnapshot(repoPath string, files []string) (*model.APISnapshot, []string) {
	orderedFiles := append([]string(nil), files...)
	sort.Strings(orderedFiles)

	fset := token.NewFileSet()
	symbols := make([]model.APISymbol, 0, len(orderedFiles)*4)
	warnings := make([]string, 0)

	for _, file := range orderedFiles {
		if shouldSkipGoSnapshotFile(file) {
			continue
		}

		node, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			warnings = append(warnings, "api stability parse skipped: "+normalizePathForSnapshot(repoPath, file)+" ("+err.Error()+")")
			continue
		}

		pkgPath := normalizePackagePath(repoPath, file, node.Name.Name)
		relFile := normalizePathForSnapshot(repoPath, file)

		for _, decl := range node.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				symbol, ok := goFunctionSymbol(fset, pkgPath, relFile, d)
				if ok {
					symbols = append(symbols, symbol)
				}
			case *ast.GenDecl:
				symbols = append(symbols, goDeclSymbols(fset, pkgPath, relFile, d)...)
			}
		}
	}

	snapshot := &model.APISnapshot{
		SchemaVersion: apiStabilitySnapshotSchema,
		Language:      "Go",
		Symbols:       symbols,
	}
	stabilizeSnapshotOrdering(snapshot)
	return snapshot, warnings
}

func shouldSkipGoSnapshotFile(path string) bool {
	lower := strings.ToLower(path)
	if !strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, "_test.go") {
		return true
	}
	if strings.Contains(lower, "/vendor/") || strings.Contains(lower, "\\vendor\\") {
		return true
	}
	if strings.Contains(lower, "/testdata/") || strings.Contains(lower, "\\testdata\\") {
		return true
	}
	return false
}

func normalizePackagePath(repoPath, filePath, packageName string) string {
	dir := filepath.Dir(filePath)
	rel, err := filepath.Rel(repoPath, dir)
	if err != nil {
		return strings.TrimSpace(packageName)
	}
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "." {
		return strings.TrimSpace(packageName)
	}
	if strings.TrimSpace(packageName) == "" {
		return rel
	}
	return rel + "/" + strings.TrimSpace(packageName)
}

func normalizePathForSnapshot(repoPath, path string) string {
	rel, err := filepath.Rel(repoPath, path)
	if err != nil {
		return filepath.ToSlash(filepath.Clean(path))
	}
	return filepath.ToSlash(filepath.Clean(rel))
}

func goFunctionSymbol(fset *token.FileSet, pkgPath, filePath string, decl *ast.FuncDecl) (model.APISymbol, bool) {
	if decl == nil || decl.Name == nil || !decl.Name.IsExported() {
		return model.APISymbol{}, false
	}

	kind := "function"
	name := decl.Name.Name
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		receiver := receiverTypeName(decl.Recv.List[0].Type)
		if receiver == "" || !isExportedIdentifier(receiver) {
			return model.APISymbol{}, false
		}
		kind = "method"
		name = receiver + "." + name
	}

	line := fset.Position(decl.Pos()).Line
	signature := goNodeString(fset, decl.Type)
	id := canonicalSymbolID(pkgPath, kind, name)

	return model.APISymbol{
		ID:        id,
		Package:   pkgPath,
		Name:      name,
		Kind:      kind,
		Signature: signature,
		File:      filePath,
		Line:      line,
	}, true
}

func goDeclSymbols(fset *token.FileSet, pkgPath, filePath string, decl *ast.GenDecl) []model.APISymbol {
	if decl == nil {
		return nil
	}

	result := make([]model.APISymbol, 0)
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Name == nil || !s.Name.IsExported() {
				continue
			}
			kind := "type"
			line := fset.Position(s.Pos()).Line
			name := s.Name.Name
			signature := goNodeString(fset, s.Type)
			result = append(result, model.APISymbol{
				ID:        canonicalSymbolID(pkgPath, kind, name),
				Package:   pkgPath,
				Name:      name,
				Kind:      kind,
				Signature: signature,
				File:      filePath,
				Line:      line,
			})
		case *ast.ValueSpec:
			for _, name := range s.Names {
				if name == nil || !name.IsExported() {
					continue
				}
				kind := strings.ToLower(decl.Tok.String())
				line := fset.Position(name.Pos()).Line
				signature := goNodeString(fset, s.Type)
				result = append(result, model.APISymbol{
					ID:        canonicalSymbolID(pkgPath, kind, name.Name),
					Package:   pkgPath,
					Name:      name.Name,
					Kind:      kind,
					Signature: signature,
					File:      filePath,
					Line:      line,
				})
			}
		}
	}
	return result
}

func receiverTypeName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		if ident, ok := typed.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

func isExportedIdentifier(name string) bool {
	if name == "" {
		return false
	}
	first := name[0:1]
	return strings.ToUpper(first) == first
}

func goNodeString(fset *token.FileSet, node interface{}) string {
	if node == nil {
		return ""
	}
	var builder strings.Builder
	if err := format.Node(&builder, fset, node); err != nil {
		return ""
	}
	return strings.TrimSpace(builder.String())
}

func canonicalSymbolID(pkgPath, kind, name string) string {
	return strings.TrimSpace(pkgPath) + "|" + strings.TrimSpace(kind) + "|" + strings.TrimSpace(name)
}

func stabilizeSnapshotOrdering(snapshot *model.APISnapshot) {
	if snapshot == nil {
		return
	}
	sort.SliceStable(snapshot.Symbols, func(i, j int) bool {
		if snapshot.Symbols[i].ID != snapshot.Symbols[j].ID {
			return snapshot.Symbols[i].ID < snapshot.Symbols[j].ID
		}
		if snapshot.Symbols[i].File != snapshot.Symbols[j].File {
			return snapshot.Symbols[i].File < snapshot.Symbols[j].File
		}
		return snapshot.Symbols[i].Line < snapshot.Symbols[j].Line
	})
}

func diffAPISnapshots(baseline, current *model.APISnapshot) []model.APIBreakingChange {
	if baseline == nil || current == nil {
		return nil
	}

	baselineMap := make(map[string]model.APISymbol, len(baseline.Symbols))
	for _, symbol := range baseline.Symbols {
		baselineMap[symbol.ID] = symbol
	}
	currentMap := make(map[string]model.APISymbol, len(current.Symbols))
	for _, symbol := range current.Symbols {
		currentMap[symbol.ID] = symbol
	}

	changes := make([]model.APIBreakingChange, 0)
	for id, before := range baselineMap {
		after, exists := currentMap[id]
		if !exists {
			changes = append(changes, model.APIBreakingChange{
				ChangeType: apiBreakTypeRemoved,
				SymbolID:   id,
				File:       before.File,
				Line:       before.Line,
				Before:     before.Signature,
			})
			continue
		}
		if before.Kind != after.Kind {
			changes = append(changes, model.APIBreakingChange{
				ChangeType: apiBreakTypeKindChanged,
				SymbolID:   id,
				File:       after.File,
				Line:       after.Line,
				Before:     before.Kind,
				After:      after.Kind,
			})
			continue
		}
		if before.Signature != after.Signature {
			changes = append(changes, model.APIBreakingChange{
				ChangeType: apiBreakTypeSignatureChanged,
				SymbolID:   id,
				File:       after.File,
				Line:       after.Line,
				Before:     before.Signature,
				After:      after.Signature,
			})
		}
	}

	sort.SliceStable(changes, func(i, j int) bool {
		leftWeight := breakTypeWeight(changes[i].ChangeType)
		rightWeight := breakTypeWeight(changes[j].ChangeType)
		if leftWeight != rightWeight {
			return leftWeight < rightWeight
		}
		if changes[i].SymbolID != changes[j].SymbolID {
			return changes[i].SymbolID < changes[j].SymbolID
		}
		if changes[i].File != changes[j].File {
			return changes[i].File < changes[j].File
		}
		return changes[i].Line < changes[j].Line
	})

	return changes
}

func breakTypeWeight(changeType string) int {
	switch changeType {
	case apiBreakTypeRemoved:
		return 0
	case apiBreakTypeKindChanged:
		return 1
	case apiBreakTypeSignatureChanged:
		return 2
	default:
		return 9
	}
}

// SerializeCurrentGoPublicAPISnapshot returns the current snapshot as pretty JSON.
// It is intended for fixture generation and baseline updates.
func SerializeCurrentGoPublicAPISnapshot(repoPath string, files []string) ([]byte, error) {
	snapshot, _ := buildGoPublicAPISnapshot(repoPath, files)
	if snapshot == nil {
		return nil, fmt.Errorf("failed to build Go public API snapshot")
	}
	return json.MarshalIndent(snapshot, "", "  ")
}
