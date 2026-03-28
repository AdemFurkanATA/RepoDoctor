package main

import (
	"encoding/json"
	"testing"
)

func TestReleaseCompatibility_JSONSchemaRequiredFields(t *testing.T) {
	report := &StructuralReport{
		Version:       "0.17.0",
		SchemaVersion: "v2",
		Path:          "repo",
		Score:         &StructuralScore{TotalScore: 100, MaxScore: 100},
		Summary:       ReportSummary{TotalViolations: 0},
		Language:      LanguageEvidenceSummary{DetectedLanguage: "Go", Confidence: 0.99, ReasonCodes: []string{"SCORING_ORDER_RESOLVED"}},
	}

	jsonText := NewReporter(FormatJSON).Format(report)
	var payload map[string]any
	if err := json.Unmarshal([]byte(jsonText), &payload); err != nil {
		t.Fatalf("failed to parse formatted JSON: %v", err)
	}

	mustHave := []string{"version", "schemaVersion", "path", "score", "summary", "language", "circularViolations", "layerViolations", "sizeViolations", "godObjectViolations"}
	for _, key := range mustHave {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected required root key %q in payload", key)
		}
	}

	language, ok := payload["language"].(map[string]any)
	if !ok {
		t.Fatalf("expected language object, got %T", payload["language"])
	}
	if _, ok := language["detectedLanguage"]; !ok {
		t.Fatal("expected detectedLanguage in language section")
	}
	if _, ok := language["confidence"]; !ok {
		t.Fatal("expected confidence in language section")
	}
}
