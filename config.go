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
}

// SizeConfig holds size rule configuration
type SizeConfig struct {
	MaxFileLines     int `yaml:"max_file_lines,omitempty"`
	MaxFunctionLines int `yaml:"max_function_lines,omitempty"`
}

// GodObjectConfig holds god object rule configuration
type GodObjectConfig struct {
	MaxFields  int `yaml:"max_fields,omitempty"`
	MaxMethods int `yaml:"max_methods,omitempty"`
}

// RulesConfig holds rule enable/disable states
type RulesConfig struct {
	EnableSizeRule      *bool `yaml:"enable_size_rule,omitempty"`
	EnableGodObjectRule *bool `yaml:"enable_god_object_rule,omitempty"`
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

	// Validate and merge with defaults
	l.config = l.mergeWithDefaults(&config)

	return l.config, nil
}

// GetConfig returns the loaded config
func (l *ConfigLoader) GetConfig() *Config {
	return l.config
}

// getDefaultConfig returns the default configuration
func (l *ConfigLoader) getDefaultConfig() *Config {
	enableSize := true
	enableGodObject := true

	return &Config{
		Size: &SizeConfig{
			MaxFileLines:     500,
			MaxFunctionLines: 80,
		},
		GodObject: &GodObjectConfig{
			MaxFields:  15,
			MaxMethods: 10,
		},
		Rules: &RulesConfig{
			EnableSizeRule:      &enableSize,
			EnableGodObjectRule: &enableGodObject,
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
