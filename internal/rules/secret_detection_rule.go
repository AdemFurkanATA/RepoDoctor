package rules

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"RepoDoctor/internal/model"
)

const (
	defaultSecretEntropyThreshold = 3.5
	defaultSecretMaxScanBytes     = 1_000_000
)

var (
	secretLineRegex = regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|passwd|private[_-]?key)\s*[:=]\s*["']?([A-Za-z0-9_./+=-]{12,})["']?`)
	awsKeyRegex     = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)
	githubPATRegex  = regexp.MustCompile(`\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{36}\b`)
	privateKeyRegex = regexp.MustCompile(`-----BEGIN (?:RSA|EC|DSA|OPENSSH|PGP|PRIVATE) KEY-----`)
)

// SecretDetectionRule detects likely hardcoded secrets in repository files.
// It uses bounded scanning and entropy checks to reduce false positives.
type SecretDetectionRule struct {
	EntropyThreshold float64
	MaxScanBytes     int
}

func NewSecretDetectionRule() *SecretDetectionRule {
	return &SecretDetectionRule{
		EntropyThreshold: defaultSecretEntropyThreshold,
		MaxScanBytes:     defaultSecretMaxScanBytes,
	}
}

func (r *SecretDetectionRule) ID() string {
	return "rule.secret-detection"
}

func (r *SecretDetectionRule) Category() string {
	return string(CategoryMaintainability)
}

func (r *SecretDetectionRule) Severity() string {
	return string(model.SeverityCritical)
}

func (r *SecretDetectionRule) Capabilities() RuleCapabilities {
	return RuleCapabilities{SupportedLanguages: []string{"Go", "Python", "JavaScript", "TypeScript", "Java"}, SupportsMultipleLanguages: true}
}

func (r *SecretDetectionRule) Evaluate(context AnalysisContext) []model.Violation {
	violations := make([]model.Violation, 0)
	for _, file := range context.RepositoryFiles {
		if !r.shouldScanFile(file.Path, file.Content) {
			continue
		}
		violations = append(violations, r.scanFile(file)...)
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

func (r *SecretDetectionRule) shouldScanFile(path, content string) bool {
	lowerPath := strings.ToLower(path)
	allowlistedPathTokens := []string{"/testdata/", "\\testdata\\", "/fixtures/", "\\fixtures\\", "/mocks/", "\\mocks\\", "/examples/", "\\examples\\", "/samples/", "\\samples\\", "/docs/", "\\docs\\"}
	for _, token := range allowlistedPathTokens {
		if strings.Contains(lowerPath, token) {
			return false
		}
	}
	if strings.HasSuffix(lowerPath, ".md") || strings.HasSuffix(lowerPath, ".txt") || strings.HasSuffix(lowerPath, ".svg") {
		return false
	}
	if len(content) == 0 || len(content) > r.MaxScanBytes {
		return false
	}
	if strings.IndexByte(content, 0) >= 0 {
		return false
	}
	return true
}

func (r *SecretDetectionRule) scanFile(file RepositoryFile) []model.Violation {
	lines := strings.Split(file.Content, "\n")
	violations := make([]model.Violation, 0)

	for idx, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}

		if privateKeyRegex.MatchString(line) {
			violations = append(violations, r.newViolation(file.Path, idx+1, "Private key material detected"))
			continue
		}

		if match := awsKeyRegex.FindString(line); match != "" && !isAllowlistedSecretCandidate(line, match) {
			violations = append(violations, r.newViolation(file.Path, idx+1, "AWS access key pattern detected"))
			continue
		}

		if match := githubPATRegex.FindString(line); match != "" && !isAllowlistedSecretCandidate(line, match) {
			violations = append(violations, r.newViolation(file.Path, idx+1, "GitHub token pattern detected"))
			continue
		}

		captures := secretLineRegex.FindAllStringSubmatch(line, -1)
		for _, capture := range captures {
			if len(capture) < 3 {
				continue
			}
			candidate := capture[2]
			if isAllowlistedSecretCandidate(line, candidate) {
				continue
			}
			if shannonEntropy(candidate) < r.EntropyThreshold {
				continue
			}
			violations = append(violations, r.newViolation(file.Path, idx+1, "Possible hardcoded secret assigned to "+capture[1]))
			break
		}
	}

	return violations
}

func (r *SecretDetectionRule) newViolation(path string, line int, message string) model.Violation {
	return model.Violation{
		RuleID:      r.ID(),
		Severity:    model.SeverityCritical,
		Message:     message,
		File:        path,
		Line:        line,
		ScoreImpact: -5.0,
	}
}

func isAllowlistedSecretCandidate(line, candidate string) bool {
	lowerLine := strings.ToLower(line)
	lowerCandidate := strings.ToLower(strings.TrimSpace(candidate))
	if strings.Contains(lowerLine, "repodoctor:allow-secret") {
		return true
	}
	if strings.Contains(lowerLine, "example") || strings.Contains(lowerLine, "sample") || strings.Contains(lowerLine, "dummy") || strings.Contains(lowerLine, "placeholder") || strings.Contains(lowerLine, "test") {
		return true
	}
	allowlistedValues := map[string]bool{
		"changeme":     true,
		"example":      true,
		"example123":   true,
		"dummy":        true,
		"placeholder":  true,
		"token":        true,
		"not-a-secret": true,
	}
	if allowlistedValues[lowerCandidate] {
		return true
	}
	if isRepeatedSingleRune(lowerCandidate) {
		return true
	}
	return false
}

func isRepeatedSingleRune(value string) bool {
	if len(value) < 8 {
		return false
	}
	first := rune(value[0])
	for _, ch := range value {
		if ch != first {
			return false
		}
	}
	return true
}

func shannonEntropy(value string) float64 {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) == 0 {
		return 0
	}

	counts := map[rune]float64{}
	total := 0.0
	for _, ch := range trimmed {
		if unicode.IsSpace(ch) {
			continue
		}
		counts[ch]++
		total++
	}
	if total == 0 {
		return 0
	}

	entropy := 0.0
	for _, count := range counts {
		p := count / total
		entropy -= p * math.Log2(p)
	}
	return entropy
}
