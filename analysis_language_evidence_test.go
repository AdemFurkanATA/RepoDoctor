package main

import (
	"reflect"
	"testing"

	"RepoDoctor/internal/languages"
)

func TestRankLanguageStats_DeterministicAndReasonableOrder(t *testing.T) {
	stats := []languages.LanguageStat{
		{Language: "Go", ProductScore: 10, Score: 22, Lines: 100, Count: 5},
		{Language: "Python", ProductScore: 11, Score: 21, Lines: 99, Count: 4},
		{Language: "TypeScript", ProductScore: 8, Score: 30, Lines: 120, Count: 6},
	}

	ranked := rankLanguageStats(stats, []string{"Python", "TypeScript", "JavaScript", "Go", "Java"})
	if len(ranked) != 3 {
		t.Fatalf("expected 3 ranked stats, got %d", len(ranked))
	}
	if ranked[0].Language != "Python" {
		t.Fatalf("expected python to win by product score, got %s", ranked[0].Language)
	}
}

func TestCollectLanguageEvidenceSummary_UnknownPathFailsSoft(t *testing.T) {
	summary := collectLanguageEvidenceSummary("/path/does/not/exist", "Go")
	if summary.DetectedLanguage != "Go" {
		t.Fatalf("expected fallback detected language Go, got %s", summary.DetectedLanguage)
	}
	if len(summary.ReasonCodes) == 0 {
		t.Fatal("expected non-empty reason codes on fail-soft summary")
	}
}

func TestLoadLanguageTieBreak_DefaultOrderWhenNoConfig(t *testing.T) {
	tieBreak := loadLanguageTieBreak(t.TempDir())
	want := []string{"Python", "TypeScript", "JavaScript", "Go", "Java"}
	if !reflect.DeepEqual(tieBreak, want) {
		t.Fatalf("unexpected default tie break order: got %v want %v", tieBreak, want)
	}
}

func TestRankLanguageStats_UsesTieBreakPriorityOnEqualScores(t *testing.T) {
	stats := []languages.LanguageStat{
		{Language: "Go", ProductScore: 10, Score: 10, Lines: 10, Count: 10},
		{Language: "Python", ProductScore: 10, Score: 10, Lines: 10, Count: 10},
	}

	ranked := rankLanguageStats(stats, []string{"Python", "Go"})
	if ranked[0].Language != "Python" {
		t.Fatalf("expected tie-break priority to rank Python first, got %s", ranked[0].Language)
	}
}
