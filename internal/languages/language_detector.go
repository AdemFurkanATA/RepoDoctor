package languages

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"RepoDoctor/internal/domain"
)

// RepositoryLanguageDetector detects the primary language of a repository
// based on deterministic extension, layout and adapter evidence.
type RepositoryLanguageDetector struct {
	adapters       map[string]LanguageAdapter
	ignoreStrategy domain.IgnoreStrategy
	policy         DetectionPolicy
}

type DetectionPolicy struct {
	LanguageWeights map[string]float64
	TieBreakOrder   []string
	SegmentWeights  map[string]float64
}

// LanguageStat holds statistics about detected language files.
type LanguageStat struct {
	Language     string
	Count        int
	Lines        int
	Score        float64
	ProductScore float64
}

type pathRole int

const (
	roleUnknown pathRole = iota
	roleProduct
	roleTest
	roleTooling
)

func roleWeight(role pathRole) float64 {
	switch role {
	case roleProduct:
		return 1.0
	case roleTest:
		return 0.6
	case roleTooling:
		return 0.2
	default:
		return 0.8
	}
}

func classifyPathRole(path string) pathRole {
	normalized := strings.ToLower(filepath.ToSlash(path))
	segments := strings.Split(normalized, "/")

	for _, seg := range segments {
		switch seg {
		case "src", "app", "pkg":
			return roleProduct
		case "test", "tests":
			return roleTest
		case "tools", "scripts", "hack", "examples", "third_party":
			return roleTooling
		}
	}

	return roleUnknown
}

func markerBoost(repoPath, language string) float64 {
	if language == "Python" {
		pythonMarkers := []string{"pyproject.toml", "requirements.txt", "setup.py"}
		for _, marker := range pythonMarkers {
			if _, err := os.Stat(filepath.Join(repoPath, marker)); err == nil {
				return 5.0
			}
		}
	}

	if language == "Go" {
		if _, err := os.Stat(filepath.Join(repoPath, "go.mod")); err == nil {
			return 5.0
		}
	}

	if language == "TypeScript" {
		if _, err := os.Stat(filepath.Join(repoPath, "tsconfig.json")); err == nil {
			return 5.0
		}
	}

	if language == "JavaScript" {
		if _, err := os.Stat(filepath.Join(repoPath, "package.json")); err == nil {
			return 3.0
		}
	}

	return 0
}

// NewRepositoryLanguageDetector creates a new language detector.
func NewRepositoryLanguageDetector(ignoreStrategy domain.IgnoreStrategy) *RepositoryLanguageDetector {
	return &RepositoryLanguageDetector{
		adapters:       make(map[string]LanguageAdapter),
		ignoreStrategy: ignoreStrategy,
		policy:         defaultDetectionPolicy(),
	}
}

func NewRepositoryLanguageDetectorWithPolicy(ignoreStrategy domain.IgnoreStrategy, policy DetectionPolicy) *RepositoryLanguageDetector {
	merged := defaultDetectionPolicy()
	if policy.LanguageWeights != nil {
		merged.LanguageWeights = policy.LanguageWeights
	}
	if len(policy.TieBreakOrder) > 0 {
		merged.TieBreakOrder = policy.TieBreakOrder
	}
	if policy.SegmentWeights != nil {
		merged.SegmentWeights = policy.SegmentWeights
	}

	return &RepositoryLanguageDetector{
		adapters:       make(map[string]LanguageAdapter),
		ignoreStrategy: ignoreStrategy,
		policy:         merged,
	}
}

func defaultDetectionPolicy() DetectionPolicy {
	return DetectionPolicy{
		LanguageWeights: map[string]float64{"Go": 1, "Python": 1, "JavaScript": 1, "TypeScript": 1},
		TieBreakOrder:   []string{"Python", "TypeScript", "JavaScript", "Go"},
		SegmentWeights:  map[string]float64{"src": 1.0, "app": 1.0, "pkg": 1.0, "tools": 0.2, "scripts": 0.2},
	}
}

func (d *RepositoryLanguageDetector) languageWeight(language string) float64 {
	if weight, ok := d.policy.LanguageWeights[language]; ok && weight > 0 {
		return weight
	}
	return 1.0
}

func (d *RepositoryLanguageDetector) segmentWeight(path string) float64 {
	normalized := strings.ToLower(filepath.ToSlash(path))
	for segment, weight := range d.policy.SegmentWeights {
		if strings.Contains(normalized, "/"+strings.ToLower(segment)+"/") || strings.HasSuffix(normalized, "/"+strings.ToLower(segment)) {
			if weight > 0 {
				return weight
			}
		}
	}
	return 1.0
}

// RegisterAdapter registers a language adapter for detection.
func (d *RepositoryLanguageDetector) RegisterAdapter(adapter LanguageAdapter) {
	d.adapters[adapter.Name()] = adapter
}

// DetectLanguage analyzes the repository and returns the primary language adapter.
func (d *RepositoryLanguageDetector) DetectLanguage(repoPath string) (LanguageAdapter, error) {
	normalizedRepoPath, err := normalizeRepoRoot(repoPath)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]*LanguageStat)
	matchedFiles := make(map[string][]string)
	adapters := d.orderedAdapters()

	err = filepath.WalkDir(normalizedRepoPath, func(path string, dEntry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		normalizedPath, ok := normalizePathWithinRoot(normalizedRepoPath, path)
		if !ok {
			if dEntry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if dEntry.IsDir() {
			if normalizedPath == normalizedRepoPath {
				return nil
			}
			if strings.HasPrefix(dEntry.Name(), ".") || (d.ignoreStrategy != nil && d.ignoreStrategy.ShouldIgnore(dEntry.Name())) {
				return filepath.SkipDir
			}
			return nil
		}

		if dEntry.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		if strings.HasPrefix(dEntry.Name(), ".") {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(normalizedPath))
		role := classifyPathRole(normalizedPath)
		weight := roleWeight(role)
		weight *= d.segmentWeight(normalizedPath)

		for _, adapter := range adapters {
			for _, supportedExt := range adapter.FileExtensions() {
				if ext != strings.ToLower(supportedExt) {
					continue
				}

				lang := adapter.Name()
				if stats[lang] == nil {
					stats[lang] = &LanguageStat{Language: lang}
				}

				stats[lang].Count++
				stats[lang].Score += 0.35 * weight * d.languageWeight(lang)
				matchedFiles[lang] = append(matchedFiles[lang], normalizedPath)

				lines, _ := countLines(normalizedPath)
				stats[lang].Lines += lines
				stats[lang].Score += float64(lines) * weight * d.languageWeight(lang)
				if role == roleProduct {
					stats[lang].ProductScore += float64(lines) + 0.35
				}
				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error scanning repository: %w", err)
	}

	d.applyAdapterEvidence(normalizedRepoPath, adapters, matchedFiles, stats)

	for lang, stat := range stats {
		stat.Score += markerBoost(normalizedRepoPath, lang)
	}

	return d.findDominantLanguage(stats)
}

func (d *RepositoryLanguageDetector) orderedAdapters() []LanguageAdapter {
	ordered := make([]LanguageAdapter, 0, len(d.adapters))
	for _, adapter := range d.adapters {
		ordered = append(ordered, adapter)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Name() < ordered[j].Name()
	})
	return ordered
}

func (d *RepositoryLanguageDetector) applyAdapterEvidence(repoPath string, adapters []LanguageAdapter, matchedFiles map[string][]string, stats map[string]*LanguageStat) {
	evidence := make([]EvidenceSignal, 0)

	for _, adapter := range adapters {
		provider, ok := adapter.(EvidenceProvider)
		if !ok {
			continue
		}

		files := append([]string(nil), matchedFiles[adapter.Name()]...)
		sort.Strings(files)
		signals, _, err := provider.CollectEvidence(repoPath, files)
		if err != nil {
			continue
		}
		evidence = append(evidence, signals...)
	}

	sort.SliceStable(evidence, func(i, j int) bool {
		if evidence[i].SourcePath != evidence[j].SourcePath {
			return evidence[i].SourcePath < evidence[j].SourcePath
		}
		if evidence[i].SignalType != evidence[j].SignalType {
			return evidence[i].SignalType < evidence[j].SignalType
		}
		if evidence[i].Language != evidence[j].Language {
			return evidence[i].Language < evidence[j].Language
		}
		return evidence[i].WeightInput < evidence[j].WeightInput
	})

	for _, signal := range evidence {
		if signal.Language == "" || signal.WeightInput <= 0 {
			continue
		}
		if stats[signal.Language] == nil {
			stats[signal.Language] = &LanguageStat{Language: signal.Language}
		}
		stats[signal.Language].Score += signal.WeightInput
	}
}

func normalizeRepoRoot(repoPath string) (string, error) {
	if strings.TrimSpace(repoPath) == "" {
		return "", fmt.Errorf("repository path is required")
	}
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve repository path: %w", err)
	}
	cleaned := filepath.Clean(absPath)
	resolved, err := filepath.EvalSymlinks(cleaned)
	if err == nil {
		cleaned = filepath.Clean(resolved)
	}
	return cleaned, nil
}

func normalizePathWithinRoot(root, candidate string) (string, bool) {
	cleaned := filepath.Clean(candidate)
	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return "", false
	}
	abs = filepath.Clean(abs)
	resolved := abs
	if eval, err := filepath.EvalSymlinks(abs); err == nil {
		resolved = filepath.Clean(eval)
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil {
		return "", false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return resolved, true
}

// findDominantLanguage returns the adapter for the strongest language.
func (d *RepositoryLanguageDetector) findDominantLanguage(stats map[string]*LanguageStat) (LanguageAdapter, error) {
	if len(stats) == 0 {
		return nil, fmt.Errorf("no supported language files found in repository")
	}

	candidates := make([]LanguageStat, 0, len(stats))
	for _, stat := range stats {
		candidates = append(candidates, *stat)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		left := candidates[i]
		right := candidates[j]

		if left.ProductScore != right.ProductScore {
			return left.ProductScore > right.ProductScore
		}
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if left.Lines != right.Lines {
			return left.Lines > right.Lines
		}
		if left.Count != right.Count {
			return left.Count > right.Count
		}

		priority := make(map[string]int, len(d.policy.TieBreakOrder))
		for i, lang := range d.policy.TieBreakOrder {
			priority[lang] = i
		}
		lp, lok := priority[left.Language]
		rp, rok := priority[right.Language]
		if lok && rok && lp != rp {
			return lp < rp
		}

		return left.Language < right.Language
	})

	dominantLang := candidates[0].Language
	if dominantLang == "" {
		return nil, fmt.Errorf("could not determine dominant language")
	}

	adapter, exists := d.adapters[dominantLang]
	if !exists {
		return nil, fmt.Errorf("no adapter found for language: %s", dominantLang)
	}

	return adapter, nil
}

// GetSupportedLanguages returns a list of all supported language names.
func (d *RepositoryLanguageDetector) GetSupportedLanguages() []string {
	languages := make([]string, 0, len(d.adapters))
	for lang := range d.adapters {
		languages = append(languages, lang)
	}
	sort.Strings(languages)
	return languages
}

// countLines counts the number of lines in a file.
func countLines(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return len(strings.Split(string(data), "\n")), nil
}

// GetLanguageStats returns detailed statistics about languages in the repository.
func (d *RepositoryLanguageDetector) GetLanguageStats(repoPath string) ([]LanguageStat, error) {
	normalizedRepoPath, err := normalizeRepoRoot(repoPath)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]*LanguageStat)
	matchedFiles := make(map[string][]string)
	adapters := d.orderedAdapters()

	err = filepath.WalkDir(normalizedRepoPath, func(path string, dEntry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		normalizedPath, ok := normalizePathWithinRoot(normalizedRepoPath, path)
		if !ok {
			if dEntry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if dEntry.IsDir() {
			if normalizedPath == normalizedRepoPath {
				return nil
			}
			if strings.HasPrefix(dEntry.Name(), ".") || (d.ignoreStrategy != nil && d.ignoreStrategy.ShouldIgnore(dEntry.Name())) {
				return filepath.SkipDir
			}
			return nil
		}

		if dEntry.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		if strings.HasPrefix(dEntry.Name(), ".") {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(normalizedPath))
		role := classifyPathRole(normalizedPath)
		weight := roleWeight(role)
		weight *= d.segmentWeight(normalizedPath)

		for _, adapter := range adapters {
			for _, supportedExt := range adapter.FileExtensions() {
				if ext != strings.ToLower(supportedExt) {
					continue
				}

				lang := adapter.Name()
				if stats[lang] == nil {
					stats[lang] = &LanguageStat{Language: lang}
				}

				stats[lang].Count++
				matchedFiles[lang] = append(matchedFiles[lang], normalizedPath)
				lines, _ := countLines(normalizedPath)
				stats[lang].Lines += lines
				stats[lang].Score += (0.35*weight + float64(lines)*weight) * d.languageWeight(lang)
				if role == roleProduct {
					stats[lang].ProductScore += float64(lines) + 0.35
				}
				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error scanning repository: %w", err)
	}

	d.applyAdapterEvidence(normalizedRepoPath, adapters, matchedFiles, stats)

	for lang, stat := range stats {
		stat.Score += markerBoost(normalizedRepoPath, lang)
	}

	result := make([]LanguageStat, 0, len(stats))
	for _, stat := range stats {
		result = append(result, *stat)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Language < result[j].Language })

	return result, nil
}

// IsMultiLanguageRepository checks if the repository contains multiple supported languages.
func (d *RepositoryLanguageDetector) IsMultiLanguageRepository(repoPath string) (bool, []string, error) {
	stats, err := d.GetLanguageStats(repoPath)
	if err != nil {
		return false, nil, err
	}

	if len(stats) <= 1 {
		return false, nil, nil
	}

	languages := make([]string, len(stats))
	for i, stat := range stats {
		languages[i] = stat.Language
	}

	return true, languages, nil
}
