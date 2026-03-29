package main

import (
	"RepoDoctor/internal/domain"
	"RepoDoctor/internal/languages"
	"math"
	"sort"
)

func collectLanguageEvidenceSummary(absPath, detectedLanguage string) LanguageEvidenceSummary {
	stats, err := collectLanguageStats(absPath)
	if err != nil || len(stats) == 0 {
		return LanguageEvidenceSummary{DetectedLanguage: detectedLanguage, Confidence: 0.0, ReasonCodes: []string{"STATS_UNAVAILABLE"}}
	}

	ranked := rankLanguageStats(stats, loadLanguageTieBreak(absPath))
	if len(ranked) == 0 {
		return LanguageEvidenceSummary{DetectedLanguage: detectedLanguage, Confidence: 0.0, ReasonCodes: []string{"STATS_UNAVAILABLE"}}
	}

	winner := ranked[0]
	if detectedLanguage == "" {
		detectedLanguage = winner.Language
	}

	reasons := []string{"SCORING_ORDER_RESOLVED"}
	confidence := 1.0
	if len(ranked) > 1 {
		runner := ranked[1]
		if winner.ProductScore > runner.ProductScore {
			reasons = append(reasons, "PRODUCT_SCORE_ADVANTAGE")
		}
		if winner.Score > runner.Score {
			reasons = append(reasons, "WEIGHTED_SCORE_ADVANTAGE")
		}
		if winner.Lines > runner.Lines {
			reasons = append(reasons, "LINE_COUNT_ADVANTAGE")
		}
		if winner.Count > runner.Count {
			reasons = append(reasons, "FILE_COUNT_ADVANTAGE")
		}

		top := winner.ProductScore + winner.Score
		second := runner.ProductScore + runner.Score
		if top+second > 0 {
			confidence = top / (top + second)
		}
	}

	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}
	confidence = math.Round(confidence*1000) / 1000

	return LanguageEvidenceSummary{
		DetectedLanguage: detectedLanguage,
		Confidence:       confidence,
		ReasonCodes:      reasons,
	}
}

func collectLanguageStats(absPath string) ([]languages.LanguageStat, error) {
	ignoreStrategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	policy := languages.DetectionPolicy{}
	config := loadConfiguration(absPath, false)
	if config != nil && config.LanguageDetection != nil {
		policy.LanguageWeights = config.LanguageDetection.Weights
		policy.TieBreakOrder = config.LanguageDetection.TieBreakOrder
		policy.SegmentWeights = config.LanguageDetection.SegmentWeights
	}

	detector := languages.NewRepositoryLanguageDetectorWithPolicy(ignoreStrategy, policy)
	registerCoreAdapters(detector, config)

	return detector.GetLanguageStats(absPath)
}

func loadLanguageTieBreak(absPath string) []string {
	config := loadConfiguration(absPath, false)
	if config != nil && config.LanguageDetection != nil && len(config.LanguageDetection.TieBreakOrder) > 0 {
		return append([]string(nil), config.LanguageDetection.TieBreakOrder...)
	}
	return []string{"Python", "TypeScript", "JavaScript", "Go"}
}

func rankLanguageStats(stats []languages.LanguageStat, tieBreak []string) []languages.LanguageStat {
	ranked := append([]languages.LanguageStat(nil), stats...)
	priority := make(map[string]int, len(tieBreak))
	for i, lang := range tieBreak {
		priority[lang] = i
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		left := ranked[i]
		right := ranked[j]
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
		lp, lok := priority[left.Language]
		rp, rok := priority[right.Language]
		if lok && rok && lp != rp {
			return lp < rp
		}
		return left.Language < right.Language
	})

	return ranked
}
