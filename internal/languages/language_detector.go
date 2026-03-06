package languages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RepositoryLanguageDetector detects the primary language of a repository
// based on file extension distribution.
type RepositoryLanguageDetector struct {
	adapters map[string]LanguageAdapter
}

// LanguageStat holds statistics about detected language files
type LanguageStat struct {
	Language string
	Count    int
	Lines    int
}

// NewRepositoryLanguageDetector creates a new language detector
func NewRepositoryLanguageDetector() *RepositoryLanguageDetector {
	return &RepositoryLanguageDetector{
		adapters: make(map[string]LanguageAdapter),
	}
}

// RegisterAdapter registers a language adapter for detection
func (d *RepositoryLanguageDetector) RegisterAdapter(adapter LanguageAdapter) {
	d.adapters[adapter.Name()] = adapter
}

// DetectLanguage analyzes the repository and returns the primary language adapter
func (d *RepositoryLanguageDetector) DetectLanguage(repoPath string) (LanguageAdapter, error) {
	stats := make(map[string]*LanguageStat)

	// Walk through repository and count files by language
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check file extension against registered adapters
		ext := strings.ToLower(filepath.Ext(path))
		for _, adapter := range d.adapters {
			for _, supportedExt := range adapter.FileExtensions() {
				if ext == strings.ToLower(supportedExt) {
					lang := adapter.Name()
					if stats[lang] == nil {
						stats[lang] = &LanguageStat{
							Language: lang,
							Count:    0,
							Lines:    0,
						}
					}
					stats[lang].Count++

					// Count lines (rough estimate)
					lines, _ := countLines(path)
					stats[lang].Lines += lines
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning repository: %w", err)
	}

	// Find the dominant language
	return d.findDominantLanguage(stats)
}

// findDominantLanguage returns the adapter for the most common language
func (d *RepositoryLanguageDetector) findDominantLanguage(stats map[string]*LanguageStat) (LanguageAdapter, error) {
	if len(stats) == 0 {
		return nil, fmt.Errorf("no supported language files found in repository")
	}

	var dominantLang string
	maxCount := 0

	for lang, stat := range stats {
		if stat.Count > maxCount {
			maxCount = stat.Count
			dominantLang = lang
		}
	}

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

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, adapter := range d.adapters {
			for _, supportedExt := range adapter.FileExtensions() {
				if ext == strings.ToLower(supportedExt) {
					lang := adapter.Name()
					if stats[lang] == nil {
						stats[lang] = &LanguageStat{
							Language: lang,
							Count:    0,
							Lines:    0,
						}
					}
					stats[lang].Count++
					lines, _ := countLines(path)
					stats[lang].Lines += lines
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning repository: %w", err)
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
