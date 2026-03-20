package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigLoader_DefaultConfig(t *testing.T) {
	loader := NewConfigLoader("")
	config := loader.getDefaultConfig()

	if config.Size.MaxFileLines != 500 {
		t.Errorf("Expected MaxFileLines to be 500, got %d", config.Size.MaxFileLines)
	}

	if config.Size.MaxFunctionLines != 80 {
		t.Errorf("Expected MaxFunctionLines to be 80, got %d", config.Size.MaxFunctionLines)
	}

	if config.GodObject.MaxFields != 15 {
		t.Errorf("Expected MaxFields to be 15, got %d", config.GodObject.MaxFields)
	}

	if config.GodObject.MaxMethods != 10 {
		t.Errorf("Expected MaxMethods to be 10, got %d", config.GodObject.MaxMethods)
	}

	if config.Rules == nil {
		t.Error("Expected Rules config to be non-nil")
	}

	if config.Rules.EnableSizeRule == nil || !*config.Rules.EnableSizeRule {
		t.Error("Expected EnableSizeRule to be true by default")
	}

	if config.Rules.EnableGodObjectRule == nil || !*config.Rules.EnableGodObjectRule {
		t.Error("Expected EnableGodObjectRule to be true by default")
	}
}

func TestConfigLoader_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yaml")

	loader := NewConfigLoader(configPath)
	config, err := loader.Load()

	if err != nil {
		t.Errorf("Expected no error for non-existent file, got: %v", err)
	}

	if config == nil {
		t.Error("Expected config to be non-nil")
	}
}

func TestConfigLoader_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewConfigLoader(configPath)
	_, err = loader.Load()

	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestConfigLoader_MergeWithDefaults(t *testing.T) {
	loader := NewConfigLoader("")

	// Partial config
	partial := &Config{
		Size: &SizeConfig{
			MaxFileLines: 1000, // Override only this
		},
	}

	merged := loader.mergeWithDefaults(partial)

	if merged.Size.MaxFileLines != 1000 {
		t.Errorf("Expected MaxFileLines to be 1000, got %d", merged.Size.MaxFileLines)
	}

	// Should have default value
	if merged.Size.MaxFunctionLines != 80 {
		t.Errorf("Expected MaxFunctionLines to be 80 (default), got %d", merged.Size.MaxFunctionLines)
	}

	// Should have default god object config
	if merged.GodObject.MaxFields != 15 {
		t.Errorf("Expected MaxFields to be 15 (default), got %d", merged.GodObject.MaxFields)
	}
}

func TestConfigLoader_MergeWithDefaults_AllSectionsParity(t *testing.T) {
	loader := NewConfigLoader("")
	partial := &Config{
		Size: &SizeConfig{
			MaxFileLines: 900,
		},
		GodObject: &GodObjectConfig{
			MaxMethods: 20,
		},
		Rules: &RulesConfig{},
		Weights: &WeightsConfig{
			Size: 9.5,
		},
		LanguageDetection: &LanguageDetectionConfig{
			Weights: map[string]float64{"Go": 2.5},
		},
	}

	merged := loader.mergeWithDefaults(partial)
	defaults := loader.getDefaultConfig()

	if merged.Size.MaxFileLines != 900 {
		t.Fatalf("expected overridden size.max_file_lines, got %d", merged.Size.MaxFileLines)
	}
	if merged.Size.MaxFunctionLines != defaults.Size.MaxFunctionLines {
		t.Fatalf("expected default size.max_function_lines, got %d", merged.Size.MaxFunctionLines)
	}

	if merged.GodObject.MaxMethods != 20 {
		t.Fatalf("expected overridden god_object.max_methods, got %d", merged.GodObject.MaxMethods)
	}
	if merged.GodObject.MaxFields != defaults.GodObject.MaxFields {
		t.Fatalf("expected default god_object.max_fields, got %d", merged.GodObject.MaxFields)
	}

	if merged.Rules.EnableSizeRule == nil || merged.Rules.EnableGodObjectRule == nil || merged.Rules.EnableCircularRule == nil || merged.Rules.EnableLayerRule == nil {
		t.Fatal("expected missing rules flags to be defaulted")
	}

	if merged.Weights.Size != 9.5 {
		t.Fatalf("expected overridden weights.size, got %.1f", merged.Weights.Size)
	}
	if merged.Weights.Circular != defaults.Weights.Circular || merged.Weights.Layer != defaults.Weights.Layer || merged.Weights.GodObject != defaults.Weights.GodObject {
		t.Fatal("expected unspecified weights to be defaulted")
	}

	if len(merged.LanguageDetection.Weights) != 1 || merged.LanguageDetection.Weights["Go"] != 2.5 {
		t.Fatalf("expected explicit language_detection.weights to be preserved, got %#v", merged.LanguageDetection.Weights)
	}
	if len(merged.LanguageDetection.TieBreakOrder) == 0 || len(merged.LanguageDetection.SegmentWeights) == 0 {
		t.Fatal("expected missing language_detection sections to be defaulted")
	}
}

func TestGetConfigPath(t *testing.T) {
	baseDir := "/test/dir"
	expected := filepath.Join(baseDir, ".repodoctor", "config.yaml")

	result := GetConfigPath(baseDir)

	if result != expected {
		t.Errorf("Expected config path %s, got %s", expected, result)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".repodoctor")

	err := EnsureConfigDir(tmpDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check if directory exists
	info, err := os.Stat(configDir)
	if err != nil {
		t.Errorf("Expected config directory to exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected config path to be a directory")
	}
}

func TestConfigLoader_LoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write valid YAML config
	configContent := `
size:
  max_file_lines: 1000
  max_function_lines: 100

god_object:
  max_fields: 20
  max_methods: 15

rules:
  enable_size_rule: true
  enable_god_object_rule: false
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	loader := NewConfigLoader(configPath)
	config, err := loader.Load()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if config.Size.MaxFileLines != 1000 {
		t.Errorf("Expected MaxFileLines to be 1000, got %d", config.Size.MaxFileLines)
	}

	if config.Size.MaxFunctionLines != 100 {
		t.Errorf("Expected MaxFunctionLines to be 100, got %d", config.Size.MaxFunctionLines)
	}

	if config.GodObject.MaxFields != 20 {
		t.Errorf("Expected MaxFields to be 20, got %d", config.GodObject.MaxFields)
	}

	if config.GodObject.MaxMethods != 15 {
		t.Errorf("Expected MaxMethods to be 15, got %d", config.GodObject.MaxMethods)
	}

	if config.Rules.EnableSizeRule == nil || !*config.Rules.EnableSizeRule {
		t.Error("Expected EnableSizeRule to be true")
	}

	if config.Rules.EnableGodObjectRule == nil || *config.Rules.EnableGodObjectRule {
		t.Error("Expected EnableGodObjectRule to be false")
	}
}

func TestConfigLoader_RejectsUnknownTopLevelKey(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := "unknown_key: true\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loader := NewConfigLoader(configPath)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for unknown top-level key")
	}
}

func TestConfigLoader_RejectsUnknownLanguageDetectionKey(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
language_detection:
  unknown: 1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loader := NewConfigLoader(configPath)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for unknown language_detection key")
	}
}

func TestConfigLoader_LanguageDetectionValidation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
language_detection:
  weights:
    Go: -1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loader := NewConfigLoader(configPath)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected validation error for negative language weight")
	}
}

func TestConfigLoader_MergeWithDefaults_TableDrivenInvariants(t *testing.T) {
	loader := NewConfigLoader("")
	enabled := false

	tests := []struct {
		name   string
		input  *Config
		assert func(t *testing.T, got, defaults *Config)
	}{
		{
			name:  "nil sections fallback to defaults",
			input: &Config{},
			assert: func(t *testing.T, got, defaults *Config) {
				t.Helper()
				if got.Size.MaxFileLines != defaults.Size.MaxFileLines || got.Size.MaxFunctionLines != defaults.Size.MaxFunctionLines {
					t.Fatalf("size defaults drifted: %+v", got.Size)
				}
				if got.GodObject.MaxFields != defaults.GodObject.MaxFields || got.GodObject.MaxMethods != defaults.GodObject.MaxMethods {
					t.Fatalf("god object defaults drifted: %+v", got.GodObject)
				}
			},
		},
		{
			name: "explicit false bool override must be preserved",
			input: &Config{
				Rules: &RulesConfig{EnableSizeRule: &enabled},
			},
			assert: func(t *testing.T, got, _ *Config) {
				t.Helper()
				if got.Rules.EnableSizeRule == nil || *got.Rules.EnableSizeRule {
					t.Fatal("expected explicit false to be preserved for rules.enable_size_rule")
				}
			},
		},
		{
			name: "zero values use deterministic defaults",
			input: &Config{
				Size:    &SizeConfig{MaxFileLines: 0, MaxFunctionLines: 0},
				Weights: &WeightsConfig{Circular: 0, Layer: 0, Size: 0, GodObject: 0},
			},
			assert: func(t *testing.T, got, defaults *Config) {
				t.Helper()
				if got.Size.MaxFileLines != defaults.Size.MaxFileLines || got.Size.MaxFunctionLines != defaults.Size.MaxFunctionLines {
					t.Fatalf("expected zero size values to fallback to defaults, got %+v", got.Size)
				}
				if got.Weights.Circular != defaults.Weights.Circular || got.Weights.Layer != defaults.Weights.Layer || got.Weights.Size != defaults.Weights.Size || got.Weights.GodObject != defaults.Weights.GodObject {
					t.Fatalf("expected zero weights to fallback to defaults, got %+v", got.Weights)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			merged := loader.mergeWithDefaults(tc.input)
			tc.assert(t, merged, loader.getDefaultConfig())
		})
	}
}

func TestConfigLoader_RejectsAmbiguousCasingKeys(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := "Size:\n  max_file_lines: 123\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loader := NewConfigLoader(configPath)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for ambiguous top-level key casing")
	}
	if !strings.Contains(err.Error(), "unknown config key 'Size'") {
		t.Fatalf("expected deterministic unknown-key error, got: %v", err)
	}
}

func TestConfigLoader_RejectsAmbiguousNestedAliasKeys(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
language_detection:
  Tie_Break_Order:
    - Go
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loader := NewConfigLoader(configPath)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for ambiguous nested key alias/casing")
	}
	if !strings.Contains(err.Error(), "unknown language_detection key 'Tie_Break_Order'") {
		t.Fatalf("expected deterministic nested unknown-key error, got: %v", err)
	}
}
