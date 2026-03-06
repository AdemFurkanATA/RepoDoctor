package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the root configuration structure
type Config struct {
	Size      *SizeConfig      `yaml:"size,omitempty"`
	GodObject *GodObjectConfig `yaml:"god_object,omitempty"`
	Rules     *RulesConfig     `yaml:"rules,omitempty"`
	Weights   *WeightsConfig   `yaml:"weights,omitempty"`
}

// SizeConfig holds size rule configuration
type SizeConfig struct {
	MaxFileLines     int `yaml:"max_file_lines,omitempty"`
	MaxFunctionLines int `yaml:"max_function_lines,omitempty"`
	Enabled          *bool `yaml:"enabled,omitempty"`
	Severity         string `yaml:"severity,omitempty"`
}

// GodObjectConfig holds god object rule configuration
type GodObjectConfig struct {
	MaxFields  int    `yaml:"max_fields,omitempty"`
	MaxMethods int    `yaml:"max_methods,omitempty"`
	Enabled    *bool  `yaml:"enabled,omitempty"`
	Severity   string `yaml:"severity,omitempty"`
	Exclude    []string `yaml:"exclude,omitempty"`
}

// RulesConfig holds rule enable/disable states
type RulesConfig struct {
	EnableSizeRule      *bool `yaml:"enable_size_rule,omitempty"`
	EnableGodObjectRule *bool `yaml:"enable_god_object_rule,omitempty"`
	EnableCircularRule  *bool `yaml:"enable_circular_rule,omitempty"`
	EnableLayerRule     *bool `yaml:"enable_layer_rule,omitempty"`
}

// WeightsConfig holds penalty weights for scoring
type WeightsConfig struct {
	Circular  float64 `yaml:"circular,omitempty"`
	Layer     float64 `yaml:"layer,omitempty"`
	Size      float64 `yaml:"size,omitempty"`
	GodObject float64 `yaml:"god_object,omitempty"`
}

// ConfigLoader handles loading and validating configuration
type ConfigLoader struct {
	configPath string
	config     *Config
}

// NewConfigLoader creates a new config loader
func NewConfigLoader(configPath string) *ConfigLoader {
	return &ConfigLoader{
		configPath: configPath,
		config:     nil,
	}
}

// Load loads configuration from file or returns defaults
func (l *ConfigLoader) Load() (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		// Return default config
		l.config = l.getDefaultConfig()
		return l.config, nil
	}

	// Read config file
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid YAML in config file: %w", err)
	}

	// Validate configuration
	if err := l.validate(&config); err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	// Validate and merge with defaults
	l.config = l.mergeWithDefaults(&config)

	return l.config, nil
}

// GetConfig returns the loaded config
func (l *ConfigLoader) GetConfig() *Config {
	return l.config
}

// validate validates the configuration and returns an error if invalid
func (l *ConfigLoader) validate(cfg *Config) error {
	// Validate severity values if provided
	validSeverities := map[string]bool{
		"info":      true,
		"warning":   true,
		"error":     true,
		"critical":  true,
	}

	if cfg.Size != nil && cfg.Size.Severity != "" {
		if !validSeverities[cfg.Size.Severity] {
			return fmt.Errorf("invalid severity '%s' for size rule (must be: info, warning, error, critical)", cfg.Size.Severity)
		}
	}

	if cfg.GodObject != nil && cfg.GodObject.Severity != "" {
		if !validSeverities[cfg.GodObject.Severity] {
			return fmt.Errorf("invalid severity '%s' for god object rule (must be: info, warning, error, critical)", cfg.GodObject.Severity)
		}
	}

	// Validate weights are non-negative
	if cfg.Weights != nil {
		if cfg.Weights.Circular < 0 {
			return fmt.Errorf("circular weight must be non-negative, got: %.2f", cfg.Weights.Circular)
		}
		if cfg.Weights.Layer < 0 {
			return fmt.Errorf("layer weight must be non-negative, got: %.2f", cfg.Weights.Layer)
		}
		if cfg.Weights.Size < 0 {
			return fmt.Errorf("size weight must be non-negative, got: %.2f", cfg.Weights.Size)
		}
		if cfg.Weights.GodObject < 0 {
			return fmt.Errorf("god object weight must be non-negative, got: %.2f", cfg.Weights.GodObject)
		}
	}

	return nil
}

// getDefaultConfig returns the default configuration
func (l *ConfigLoader) getDefaultConfig() *Config {
	enableSize := true
	enableGodObject := true
	enableCircular := true
	enableLayer := true

	return &Config{
		Size: &SizeConfig{
			MaxFileLines:     500,
			MaxFunctionLines: 80,
			Enabled:          &enableSize,
			Severity:         "warning",
		},
		GodObject: &GodObjectConfig{
			MaxFields:  15,
			MaxMethods: 20,
			Enabled:    &enableGodObject,
			Severity:   "warning",
			// Exclude internal implementation files from strict checks
			Exclude: []string{"internal/"},
		},
		Rules: &RulesConfig{
			EnableSizeRule:      &enableSize,
			EnableGodObjectRule: &enableGodObject,
			EnableCircularRule:  &enableCircular,
			EnableLayerRule:     &enableLayer,
		},
		Weights: &WeightsConfig{
			Circular:  10.0,
			Layer:     5.0,
			Size:      3.0,
			GodObject: 5.0,
		},
	}
}

// mergeWithDefaults merges provided config with defaults
func (l *ConfigLoader) mergeWithDefaults(cfg *Config) *Config {
	defaults := l.getDefaultConfig()

	// Merge size config
	if cfg.Size == nil {
		cfg.Size = defaults.Size
	} else {
		if cfg.Size.MaxFileLines == 0 {
			cfg.Size.MaxFileLines = defaults.Size.MaxFileLines
		}
		if cfg.Size.MaxFunctionLines == 0 {
			cfg.Size.MaxFunctionLines = defaults.Size.MaxFunctionLines
		}
		if cfg.Size.Enabled == nil {
			cfg.Size.Enabled = defaults.Size.Enabled
		}
		if cfg.Size.Severity == "" {
			cfg.Size.Severity = defaults.Size.Severity
		}
	}

	// Merge god object config
	if cfg.GodObject == nil {
		cfg.GodObject = defaults.GodObject
	} else {
		if cfg.GodObject.MaxFields == 0 {
			cfg.GodObject.MaxFields = defaults.GodObject.MaxFields
		}
		if cfg.GodObject.MaxMethods == 0 {
			cfg.GodObject.MaxMethods = defaults.GodObject.MaxMethods
		}
		if cfg.GodObject.Enabled == nil {
			cfg.GodObject.Enabled = defaults.GodObject.Enabled
		}
		if cfg.GodObject.Severity == "" {
			cfg.GodObject.Severity = defaults.GodObject.Severity
		}
	}

	// Merge rules config
	if cfg.Rules == nil {
		cfg.Rules = defaults.Rules
	} else {
		if cfg.Rules.EnableSizeRule == nil {
			cfg.Rules.EnableSizeRule = defaults.Rules.EnableSizeRule
		}
		if cfg.Rules.EnableGodObjectRule == nil {
			cfg.Rules.EnableGodObjectRule = defaults.Rules.EnableGodObjectRule
		}
		if cfg.Rules.EnableCircularRule == nil {
			cfg.Rules.EnableCircularRule = defaults.Rules.EnableCircularRule
		}
		if cfg.Rules.EnableLayerRule == nil {
			cfg.Rules.EnableLayerRule = defaults.Rules.EnableLayerRule
		}
	}

	// Merge weights config
	if cfg.Weights == nil {
		cfg.Weights = defaults.Weights
	} else {
		if cfg.Weights.Circular == 0 {
			cfg.Weights.Circular = defaults.Weights.Circular
		}
		if cfg.Weights.Layer == 0 {
			cfg.Weights.Layer = defaults.Weights.Layer
		}
		if cfg.Weights.Size == 0 {
			cfg.Weights.Size = defaults.Weights.Size
		}
		if cfg.Weights.GodObject == 0 {
			cfg.Weights.GodObject = defaults.Weights.GodObject
		}
	}

	return cfg
}

// GetConfigPath returns the default config path for a given directory
func GetConfigPath(baseDir string) string {
	return filepath.Join(baseDir, ".repodoctor", "config.yaml")
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir(baseDir string) error {
	configDir := filepath.Join(baseDir, ".repodoctor")
	return os.MkdirAll(configDir, 0755)
}
