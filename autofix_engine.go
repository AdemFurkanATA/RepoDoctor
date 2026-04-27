package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type AutoFixPack string

const (
	AutoFixPackSize      AutoFixPack = "size"
	AutoFixPackImports   AutoFixPack = "imports"
	AutoFixPackErrorWrap AutoFixPack = "error-wrap"
)

type AutoFixRequest struct {
	RepositoryPath string
	Apply          bool
	Packs          []AutoFixPack
	DisabledPacks  []string
	UnknownPacks   []string
}

type AutoFixResult struct {
	DryRun         bool
	FilesScanned   int
	FilesChanged   int
	TotalEdits     int
	Preview        []string
	ChangedFiles   []string
	SkippedReasons []string
}

type fileEdit struct {
	path     string
	original []byte
	updated  []byte
	edits    int
}

func runAutoFix(request AutoFixRequest) (*AutoFixResult, error) {
	absPath := validatePath(request.RepositoryPath)
	packs := resolveAutoFixPacks(request.Packs)
	if len(packs) == 0 {
		return nil, NewCLIError(
			ErrorInvalidArgument,
			"No safe auto-fix packs selected",
			"Use --packs size,imports,error-wrap",
			nil,
		)
	}

	files, err := collectSafeGoFiles(absPath)
	if err != nil {
		return nil, WrapError(err, ErrorRuntime, "auto-fix scan failed", GetSuggestion(err.Error()))
	}

	result := &AutoFixResult{DryRun: !request.Apply, Preview: []string{}, ChangedFiles: []string{}, SkippedReasons: []string{}}
	for _, item := range request.DisabledPacks {
		result.SkippedReasons = append(result.SkippedReasons, "high-risk pack disabled in auto mode: "+item)
	}
	for _, item := range request.UnknownPacks {
		result.SkippedReasons = append(result.SkippedReasons, "unknown pack skipped: "+item)
	}
	result.FilesScanned = len(files)

	edits := make([]fileEdit, 0)
	for _, file := range files {
		content, readErr := os.ReadFile(file)
		if readErr != nil {
			result.SkippedReasons = append(result.SkippedReasons, "read skipped: "+normalizeReportPathDeterministic(file))
			continue
		}

		updated, count := applySafePacks(content, packs)
		if count == 0 || bytes.Equal(content, updated) {
			continue
		}

		edits = append(edits, fileEdit{path: file, original: content, updated: updated, edits: count})
	}

	sort.SliceStable(edits, func(i, j int) bool { return edits[i].path < edits[j].path })

	for _, edit := range edits {
		result.FilesChanged++
		result.TotalEdits += edit.edits
		rel := normalizeReportPathDeterministic(edit.path)
		result.ChangedFiles = append(result.ChangedFiles, rel)
		result.Preview = append(result.Preview, buildPatchPreview(rel, edit.original, edit.updated, 4)...)
		if request.Apply {
			if writeErr := os.WriteFile(edit.path, edit.updated, 0o644); writeErr != nil {
				return nil, WrapError(writeErr, ErrorRuntime, "auto-fix apply failed", GetSuggestion(writeErr.Error()))
			}
		}
	}

	return result, nil
}

func resolveAutoFixPacks(input []AutoFixPack) []AutoFixPack {
	if len(input) == 0 {
		return []AutoFixPack{AutoFixPackSize, AutoFixPackImports, AutoFixPackErrorWrap}
	}
	allowed := map[AutoFixPack]bool{AutoFixPackSize: true, AutoFixPackImports: true, AutoFixPackErrorWrap: true}
	set := map[AutoFixPack]bool{}
	resolved := make([]AutoFixPack, 0, len(input))
	for _, pack := range input {
		normalized := AutoFixPack(strings.ToLower(strings.TrimSpace(string(pack))))
		if !allowed[normalized] || set[normalized] {
			continue
		}
		set[normalized] = true
		resolved = append(resolved, normalized)
	}
	return resolved
}

func parseAutoFixPacks(raw string) (safe []AutoFixPack, disabled []string, unknown []string) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(raw)), ",")

	safeSet := map[AutoFixPack]bool{}
	safe = make([]AutoFixPack, 0, len(parts))
	disabled = make([]string, 0)
	unknown = make([]string, 0)

	safeCatalog := map[string]AutoFixPack{
		"size":       AutoFixPackSize,
		"imports":    AutoFixPackImports,
		"error-wrap": AutoFixPackErrorWrap,
	}
	highRiskCatalog := map[string]bool{
		"extract-method": true,
		"layer-refactor": true,
		"api-rewrite":    true,
	}

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if pack, ok := safeCatalog[trimmed]; ok {
			if !safeSet[pack] {
				safeSet[pack] = true
				safe = append(safe, pack)
			}
			continue
		}
		if highRiskCatalog[trimmed] {
			disabled = append(disabled, trimmed)
			continue
		}
		unknown = append(unknown, trimmed)
	}

	return safe, disabled, unknown
}

func collectSafeGoFiles(root string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			name := strings.ToLower(d.Name())
			if name == ".git" || name == "vendor" || name == "testdata" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		lower := strings.ToLower(path)
		if !strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func applySafePacks(content []byte, packs []AutoFixPack) ([]byte, int) {
	updated := append([]byte(nil), content...)
	edits := 0
	for _, pack := range packs {
		switch pack {
		case AutoFixPackSize:
			next, delta := applySizePack(updated)
			updated, edits = next, edits+delta
		case AutoFixPackImports:
			next, delta := applyImportsPack(updated)
			updated, edits = next, edits+delta
		case AutoFixPackErrorWrap:
			next, delta := applyErrorWrapPack(updated)
			updated, edits = next, edits+delta
		}
	}
	return updated, edits
}

func applySizePack(content []byte) ([]byte, int) {
	text := strings.ReplaceAll(string(content), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	edits := 0
	for i, line := range lines {
		trimmed := strings.TrimRight(line, " \t")
		if trimmed != line {
			lines[i] = trimmed
			edits++
		}
	}

	collapsed := make([]string, 0, len(lines))
	blanks := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blanks++
			if blanks > 2 {
				edits++
				continue
			}
		} else {
			blanks = 0
		}
		collapsed = append(collapsed, line)
	}

	return []byte(strings.Join(collapsed, "\n")), edits
}

func applyImportsPack(content []byte) ([]byte, int) {
	formatted, err := format.Source(content)
	if err != nil {
		return content, 0
	}
	if bytes.Equal(formatted, content) {
		return content, 0
	}
	return formatted, 1
}

var errorWrapRegex = regexp.MustCompile(`fmt\.Errorf\("([^"]*?)%v([^"]*?)",\s*err\)`) // safe subset only

func applyErrorWrapPack(content []byte) ([]byte, int) {
	text := string(content)
	matches := errorWrapRegex.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return content, 0
	}
	replaced := errorWrapRegex.ReplaceAllString(text, `fmt.Errorf("$1%w$2", err)`)
	return []byte(replaced), len(matches)
}

func buildPatchPreview(path string, before, after []byte, maxLines int) []string {
	left := strings.Split(strings.ReplaceAll(string(before), "\r\n", "\n"), "\n")
	right := strings.Split(strings.ReplaceAll(string(after), "\r\n", "\n"), "\n")
	limit := len(left)
	if len(right) > limit {
		limit = len(right)
	}
	lines := []string{fmt.Sprintf("--- %s", path), fmt.Sprintf("+++ %s", path)}
	for i, used := 0, 0; i < limit && used < maxLines; i++ {
		var l, r string
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		if l == r {
			continue
		}
		if l != "" {
			lines = append(lines, fmt.Sprintf("- %d:%s", i+1, l))
			used++
		}
		if r != "" && used < maxLines {
			lines = append(lines, fmt.Sprintf("+ %d:%s", i+1, r))
			used++
		}
	}
	return lines
}

func printAutoFixResult(result *AutoFixResult) {
	if result == nil {
		return
	}
	mode := "APPLY"
	if result.DryRun {
		mode = "DRY-RUN"
	}
	fmt.Printf("Auto-fix mode: %s\n", mode)
	fmt.Printf("Files scanned: %d\n", result.FilesScanned)
	fmt.Printf("Files changed: %d\n", result.FilesChanged)
	fmt.Printf("Total edits: %d\n", result.TotalEdits)
	for _, line := range result.Preview {
		fmt.Println(line)
	}
	for _, reason := range result.SkippedReasons {
		fmt.Printf("Skipped: %s\n", reason)
	}
	if result.DryRun {
		fmt.Println("Dry-run is default. Re-run with --apply to write safe fixes.")
	}
}
