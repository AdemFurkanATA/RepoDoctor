package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"RepoDoctor/internal/model"
)

func sortedStringCopy(in []string) []string {
	result := append([]string(nil), in...)
	sort.Strings(result)
	return result
}

func stableSortedCopy[T any](in []T, less func(left, right T) bool) []T {
	result := append([]T(nil), in...)
	sort.SliceStable(result, func(i, j int) bool {
		return less(result[i], result[j])
	})
	return result
}

func normalizeReportPathDeterministic(path string) string {
	cleaned := filepath.ToSlash(filepath.Clean(path))
	if wd, err := os.Getwd(); err == nil {
		if rel, relErr := filepath.Rel(wd, cleaned); relErr == nil && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
	}
	return cleaned
}

func sortedCircularViolations(in []CycleViolation) []CycleViolation {
	return stableSortedCopy(in, func(left, right CycleViolation) bool {
		return strings.Join(left.Path, "/") < strings.Join(right.Path, "/")
	})
}

func sortedLayerViolations(in []LayerViolation) []LayerViolation {
	return stableSortedCopy(in, func(left, right LayerViolation) bool {
		if left.From != right.From {
			return left.From < right.From
		}
		if left.To != right.To {
			return left.To < right.To
		}
		return left.Message < right.Message
	})
}

func sortedSizeViolations(in []SizeViolation) []SizeViolation {
	return stableSortedCopy(in, func(left, right SizeViolation) bool {
		if left.File != right.File {
			return left.File < right.File
		}
		if left.Function != right.Function {
			return left.Function < right.Function
		}
		return left.Lines < right.Lines
	})
}

func sortedGodObjectViolations(in []GodObjectViolation) []GodObjectViolation {
	return stableSortedCopy(in, func(left, right GodObjectViolation) bool {
		if left.File != right.File {
			return left.File < right.File
		}
		return left.StructName < right.StructName
	})
}

func sortedAPIViolations(in []APIStabilityViolation) []APIStabilityViolation {
	return stableSortedCopy(in, func(left, right APIStabilityViolation) bool {
		if left.File != right.File {
			return left.File < right.File
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		return left.Message < right.Message
	})
}

func sortedGitChurnFiles(in []model.GitChurnFile) []model.GitChurnFile {
	return stableSortedCopy(in, func(left, right model.GitChurnFile) bool {
		if left.CommitTouches != right.CommitTouches {
			return left.CommitTouches > right.CommitTouches
		}
		if left.RecentTouches != right.RecentTouches {
			return left.RecentTouches > right.RecentTouches
		}
		if left.UniqueAuthors != right.UniqueAuthors {
			return left.UniqueAuthors > right.UniqueAuthors
		}
		return left.Path < right.Path
	})
}

func sortedHotspotEntries(in []HotspotEntry) []HotspotEntry {
	return stableSortedCopy(in, func(left, right HotspotEntry) bool {
		if left.RiskScore != right.RiskScore {
			return left.RiskScore > right.RiskScore
		}
		return left.File < right.File
	})
}

func sortedVulnerabilityFindings(in []model.VulnerabilityFinding) []model.VulnerabilityFinding {
	return stableSortedCopy(in, func(left, right model.VulnerabilityFinding) bool {
		if left.Package != right.Package {
			return left.Package < right.Package
		}
		if left.Version != right.Version {
			return left.Version < right.Version
		}
		return left.ID < right.ID
	})
}

func sortModelViolationsDeterministic(violations []model.Violation) {
	sort.SliceStable(violations, func(i, j int) bool {
		if violations[i].RuleID != violations[j].RuleID {
			return violations[i].RuleID < violations[j].RuleID
		}
		if violations[i].File != violations[j].File {
			return violations[i].File < violations[j].File
		}
		if violations[i].Line != violations[j].Line {
			return violations[i].Line < violations[j].Line
		}
		return violations[i].Message < violations[j].Message
	})
}
