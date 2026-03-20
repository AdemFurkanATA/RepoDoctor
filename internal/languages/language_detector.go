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
// based on file extension distribution.
type RepositoryLanguageDetector struct {
	adapters       map[string]LanguageAdapter
	ignoreStrategy domain.IgnoreStrategy
}

// LanguageStat holds statistics about detected language files
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

	return 0
}

// NewRepositoryLanguageDetector creates a new language detector
func NewRepositoryLanguageDetector(ignoreStrategy domain.IgnoreStrategy) *RepositoryLanguageDetector {
	return &RepositoryLanguageDetector{
		adapters:       make(map[string]LanguageAdapter),
		ignoreStrategy: ignoreStrategy,
	}
}

// RegisterAdapter registers a language adapter for detection
func (d *RepositoryLanguageDetector) RegisterAdapter(adapter LanguageAdapter) {
	d.adapters[adapter.Name()] = adapter
}

// DetectLanguage analyzes the repository and returns the primary language adapter
func (d *RepositoryLanguageDetector) DetectLanguage(repoPath string) (LanguageAdapter, error) {
	stats := make(map[string]*LanguageStat)

	// WalkDir is faster than Walk because it doesn't call os.Lstat for every file/directory
	err := filepath.WalkDir(repoPath, func(path string, dEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Directory handling
		if dEntry.IsDir() {
			// Do not skip the root directory itself
			if path == repoPath {
				return nil
			}

			// Check ignore strategy or hidden directories
			if strings.HasPrefix(dEntry.Name(), ".") || (d.ignoreStrategy != nil && d.ignoreStrategy.ShouldIgnore(path, dEntry.Name())) {
				return filepath.SkipDir
			}
			return nil
		}

		// Security: skip symlinks to prevent infinite loops and path traversal
		if dEntry.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(dEntry.Name(), ".") {
			return nil
		}

		// Check file extension against registered adapters
		ext := strings.ToLower(filepath.Ext(path))
		role := classifyPathRole(path)
		weight := roleWeight(role)
		for _, adapter := range d.adapters {
			for _, supportedExt := range adapter.FileExtensions() {
				if ext == strings.ToLower(supportedExt) {
					lang := adapter.Name()
					if stats[lang] == nil {
						stats[lang] = &LanguageStat{
							Language:     lang,
							Count:        0,
							Lines:        0,
							Score:        0,
							ProductScore: 0,
						}
					}
					stats[lang].Count++
					stats[lang].Score += 0.35 * weight

					// Count lines (rough estimate)
					lines, _ := countLines(path)
					stats[lang].Lines += lines
					stats[lang].Score += float64(lines) * weight
					if role == roleProduct {
						stats[lang].ProductScore += float64(lines) + 0.35
					}
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning repository: %w", err)
	}

	for lang, stat := range stats {
		stat.Score += markerBoost(repoPath, lang)
	}

	// Find the dominant language
	return d.findDominantLanguage(stats)
}

// findDominantLanguage returns the adapter for the most common language
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

		priority := map[string]int{
			"Python": 0,
			"Go":     1,
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

// GetSupportedLanguages returns a list of all supported language names
func (d *RepositoryLanguageDetector) GetSupportedLanguages() []string {
	languages := make([]string, 0, len(d.adapters))
	for lang := range d.adapters {
		languages = append(languages, lang)
	}
	return languages
}

// countLines counts the number of lines in a file
func countLines(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return len(strings.Split(string(data), "\n")), nil
}

// GetLanguageStats returns detailed statistics about languages in the repository
func (d *RepositoryLanguageDetector) GetLanguageStats(repoPath string) ([]LanguageStat, error) {
	stats := make(map[string]*LanguageStat)

	err := filepath.WalkDir(repoPath, func(path string, dEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Directory handling
		if dEntry.IsDir() {
			if path == repoPath {
				return nil
			}
			if strings.HasPrefix(dEntry.Name(), ".") || (d.ignoreStrategy != nil && d.ignoreStrategy.ShouldIgnore(path, dEntry.Name())) {
				return filepath.SkipDir
			}
			return nil
		}

		// Security: skip symlinks
		if dEntry.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		if strings.HasPrefix(dEntry.Name(), ".") {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		role := classifyPathRole(path)
		weight := roleWeight(role)
		for _, adapter := range d.adapters {
			for _, supportedExt := range adapter.FileExtensions() {
				if ext == strings.ToLower(supportedExt) {
					lang := adapter.Name()
					if stats[lang] == nil {
						stats[lang] = &LanguageStat{
							Language:     lang,
							Count:        0,
							Lines:        0,
							Score:        0,
							ProductScore: 0,
						}
					}
					stats[lang].Count++
					lines, _ := countLines(path)
					stats[lang].Lines += lines
					stats[lang].Score += 0.35*weight + float64(lines)*weight
					if role == roleProduct {
						stats[lang].ProductScore += float64(lines) + 0.35
					}
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning repository: %w", err)
	}

	for lang, stat := range stats {
		stat.Score += markerBoost(repoPath, lang)
	}

	result := make([]LanguageStat, 0, len(stats))
	for _, stat := range stats {
		result = append(result, *stat)
	}

	return result, nil
}

// IsMultiLanguageRepository checks if the repository contains multiple supported languages
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
