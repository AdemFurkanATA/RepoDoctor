package rules

import (
	"fmt"
	"os"
	"path/filepath"
)

// PluginManager handles plugin discovery, loading, and lifecycle management
type PluginManager struct {
	pluginsDir string
	plugins    []Plugin
	infos      []PluginInfo
	registry   *RuleRegistry
}

// NewPluginManager creates a new plugin manager
// pluginsDir: directory to search for plugins (e.g., .repodoctor/plugins/)
func NewPluginManager(pluginsDir string, registry *RuleRegistry) *PluginManager {
	return &PluginManager{
		pluginsDir: pluginsDir,
		plugins:    make([]Plugin, 0),
		infos:      make([]PluginInfo, 0),
		registry:   registry,
	}
}

// DiscoverAndLoad discovers plugins in the plugins directory and loads them
func (pm *PluginManager) DiscoverAndLoad() error {
	// Check if plugins directory exists
	if _, err := os.Stat(pm.pluginsDir); os.IsNotExist(err) {
		// No plugins directory, nothing to load
		return nil
	}

	// Walk through plugins directory
	err := filepath.Walk(pm.pluginsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only load .go files (in production, these would be compiled .dll/.so files)
		// For v0.5, we'll use a simpler approach with plugin registry
		if filepath.Ext(path) == ".go" || filepath.Ext(path) == ".yaml" {
			pluginInfo := PluginInfo{
				Name:   filepath.Base(path),
				Loaded: false,
				Error:  fmt.Errorf("plugin loading requires compiled plugin binary"),
			}
			pm.infos = append(pm.infos, pluginInfo)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error discovering plugins: %w", err)
	}

	return nil
}

// RegisterPlugin manually registers a plugin with the manager
// This is used for built-in plugins or programmatic registration
func (pm *PluginManager) RegisterPlugin(plugin Plugin) error {
	info := PluginInfo{
		Name:        plugin.Name(),
		Version:     plugin.Version(),
		Description: plugin.Description(),
		Loaded:      true,
	}

	// Register rules from plugin
	if err := pm.registerPluginRules(plugin); err != nil {
		info.Error = err
		info.Loaded = false
		pm.infos = append(pm.infos, info)
		return err
	}

	pm.plugins = append(pm.plugins, plugin)
	pm.infos = append(pm.infos, info)

	return nil
}

// registerPluginRules calls the plugin's RegisterRules method
func (pm *PluginManager) registerPluginRules(plugin Plugin) error {
	plugin.RegisterRules(pm.registry)
	return nil
}

// GetPlugins returns all loaded plugins
func (pm *PluginManager) GetPlugins() []Plugin {
	return pm.plugins
}

// GetPluginInfo returns information about all discovered plugins
func (pm *PluginManager) GetPluginInfo() []PluginInfo {
	return pm.infos
}

// GetPluginCount returns the number of successfully loaded plugins
func (pm *PluginManager) GetPluginCount() int {
	return len(pm.plugins)
}

// GetPluginsDir returns the plugins directory path
func (pm *PluginManager) GetPluginsDir() string {
	return pm.pluginsDir
}
