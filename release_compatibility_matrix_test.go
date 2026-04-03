package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompatibilityMatrix_ConfigArchitectureProfiles(t *testing.T) {
	tests := []struct {
		name      string
		profile   string
		wantError bool
	}{
		{name: "clean", profile: "clean"},
		{name: "layered", profile: "layered"},
		{name: "modular-monolith", profile: "modular-monolith"},
		{name: "invalid", profile: "hexagonal", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfgPath := filepath.Join(dir, "config.yaml")
			content := []byte("architecture:\n  profile: " + tt.profile + "\n")
			if err := os.WriteFile(cfgPath, content, 0o644); err != nil {
				t.Fatalf("failed writing config fixture: %v", err)
			}

			loader := NewConfigLoader(cfgPath)
			cfg, err := loader.Load()
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error for invalid architecture profile")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected load error: %v", err)
			}
			if cfg.Architecture == nil || cfg.Architecture.Profile != tt.profile {
				t.Fatalf("expected architecture profile %q, got %#v", tt.profile, cfg.Architecture)
			}
		})
	}
}

func TestCompatibilityMatrix_ConfigDefaultsRemainStable(t *testing.T) {
	loader := NewConfigLoader(filepath.Join(t.TempDir(), "missing.yaml"))
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected default config load error: %v", err)
	}

	if cfg.Architecture == nil || cfg.Architecture.Profile != "layered" {
		t.Fatalf("expected default architecture profile layered, got %#v", cfg.Architecture)
	}
	if cfg.LanguageDetection == nil || len(cfg.LanguageDetection.Weights) == 0 {
		t.Fatal("expected default language detection weights to be present")
	}
	if cfg.Rules == nil || cfg.Rules.EnableCircularRule == nil || cfg.Rules.EnableLayerRule == nil {
		t.Fatal("expected default rule toggles to include circular/layer rules")
	}
}

func TestCompatibilityMatrix_ReportJSONFormats(t *testing.T) {
	base := &StructuralReport{
		Version:       "1.1.0",
		SchemaVersion: "v2",
		Path:          "repo",
		Score:         &StructuralScore{TotalScore: 100, MaxScore: 100},
		Summary:       ReportSummary{TotalViolations: 0},
		Language:      LanguageEvidenceSummary{DetectedLanguage: "Go", Confidence: 1},
	}

	v2 := NewReporter(FormatJSON).Format(base)
	if !containsAll(v2, "\"schemaVersion\"", "\"summary\"", "\"language\"") {
		t.Fatalf("v2 output missing required compatibility sections: %s", v2)
	}

	v1 := NewReporter(FormatJSONV1).Format(base)
	if containsAny(v1, "\"schemaVersion\"", "\"summary\"", "\"language\"") {
		t.Fatalf("v1 compatibility broken, contains v2-only fields: %s", v1)
	}
}

func containsAll(input string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(input, part) {
			return false
		}
	}
	return true
}

func containsAny(input string, parts ...string) bool {
	for _, part := range parts {
		if strings.Contains(input, part) {
			return true
		}
	}
	return false
}
