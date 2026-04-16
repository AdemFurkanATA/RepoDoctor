package rules

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"RepoDoctor/internal/model"
)

const (
	defaultDuplicationWindowLines = 6
	defaultDuplicationMaxWindows  = 50000
)

type CodeDuplicationRule struct {
	WindowLines int
	MaxWindows  int
}

func NewCodeDuplicationRule() *CodeDuplicationRule {
	return &CodeDuplicationRule{WindowLines: defaultDuplicationWindowLines, MaxWindows: defaultDuplicationMaxWindows}
}

func (r *CodeDuplicationRule) ID() string { return "rule.code-duplication" }

func (r *CodeDuplicationRule) Category() string { return string(CategoryMaintainability) }

func (r *CodeDuplicationRule) Severity() string { return string(model.SeverityWarning) }

func (r *CodeDuplicationRule) Capabilities() RuleCapabilities {
	return RuleCapabilities{SupportedLanguages: []string{"Go", "Python", "JavaScript", "TypeScript", "Java"}, SupportsMultipleLanguages: true}
}

func (r *CodeDuplicationRule) Evaluate(context AnalysisContext) []model.Violation {
	window := r.WindowLines
	if window < 4 {
		window = defaultDuplicationWindowLines
	}
	maxWindows := r.MaxWindows
	if maxWindows <= 0 {
		maxWindows = defaultDuplicationMaxWindows
	}

	type windowInstance struct {
		file      string
		startLine int
		length    int
	}

	buckets := make(map[string][]windowInstance)
	totalWindows := 0

	for _, file := range context.RepositoryFiles {
		if shouldSkipDuplicationScan(file.Path) {
			continue
		}
		normalizedLines, sourceLines := normalizeContentForDuplication(file.Content)
		if len(normalizedLines) < window {
			continue
		}
		for idx := 0; idx+window <= len(normalizedLines); idx++ {
			if totalWindows >= maxWindows {
				break
			}
			signature := hashWindow(normalizedLines[idx : idx+window])
			buckets[signature] = append(buckets[signature], windowInstance{
				file:      file.Path,
				startLine: sourceLines[idx],
				length:    window,
			})
			totalWindows++
		}
		if totalWindows >= maxWindows {
			break
		}
	}

	violations := make([]model.Violation, 0)
	seen := map[string]bool{}

	for _, instances := range buckets {
		if len(instances) < 2 {
			continue
		}
		sort.Slice(instances, func(i, j int) bool {
			if instances[i].file != instances[j].file {
				return instances[i].file < instances[j].file
			}
			return instances[i].startLine < instances[j].startLine
		})

		for i := 0; i < len(instances)-1; i++ {
			left := instances[i]
			right := instances[i+1]
			if left.file == right.file {
				continue
			}
			fingerprint := left.file + ":" + fmt.Sprint(left.startLine) + "->" + right.file + ":" + fmt.Sprint(right.startLine)
			if seen[fingerprint] {
				continue
			}
			seen[fingerprint] = true

			violations = append(violations,
				model.Violation{
					RuleID:      r.ID(),
					Severity:    model.SeverityWarning,
					Message:     fmt.Sprintf("File %s has %d lines (threshold: %d) duplicated with %s", left.file, left.length, window, right.file),
					File:        left.file,
					Line:        left.startLine,
					ScoreImpact: -2.0,
				},
				model.Violation{
					RuleID:      r.ID(),
					Severity:    model.SeverityWarning,
					Message:     fmt.Sprintf("File %s has %d lines (threshold: %d) duplicated with %s", right.file, right.length, window, left.file),
					File:        right.file,
					Line:        right.startLine,
					ScoreImpact: -2.0,
				},
			)
		}
	}

	sort.Slice(violations, func(i, j int) bool {
		if violations[i].File != violations[j].File {
			return violations[i].File < violations[j].File
		}
		if violations[i].Line != violations[j].Line {
			return violations[i].Line < violations[j].Line
		}
		return violations[i].Message < violations[j].Message
	})

	return violations
}

func shouldSkipDuplicationScan(path string) bool {
	lower := strings.ToLower(path)
	if strings.Contains(lower, "/testdata/") || strings.Contains(lower, "\\testdata\\") {
		return true
	}
	if strings.Contains(lower, "/vendor/") || strings.Contains(lower, "\\vendor\\") {
		return true
	}
	if strings.Contains(lower, "/fixtures/") || strings.Contains(lower, "\\fixtures\\") {
		return true
	}
	if strings.Contains(lower, "/docs/") || strings.Contains(lower, "\\docs\\") {
		return true
	}
	if strings.HasPrefix(lower, "docs/") || strings.HasPrefix(lower, "docs\\") {
		return true
	}
	if strings.Contains(lower, ".min.") {
		return true
	}
	return false
}

func normalizeContentForDuplication(content string) ([]string, []int) {
	lines := strings.Split(content, "\n")
	normalized := make([]string, 0, len(lines))
	sourceLines := make([]int, 0, len(lines))

	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		compact := strings.Join(strings.Fields(trimmed), " ")
		normalized = append(normalized, compact)
		sourceLines = append(sourceLines, idx+1)
	}

	return normalized, sourceLines
}

func hashWindow(lines []string) string {
	joined := strings.Join(lines, "\n")
	sum := sha1.Sum([]byte(joined))
	return hex.EncodeToString(sum[:])
}
