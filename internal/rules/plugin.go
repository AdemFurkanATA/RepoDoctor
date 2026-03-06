package rules

// Plugin defines the interface that all RepoDoctor plugins must implement.
// Plugins allow external developers to extend RepoDoctor with custom rules
// without modifying core code.
//
// Plugins are discovered from .repodoctor/plugins/ directory
// and loaded during engine initialization.
type Plugin interface {
	// Name returns the human-readable name of the plugin.
	// Example: "CustomArchitectureRules"
	Name() string

	// Version returns the plugin version string.
	// Should follow semantic versioning (e.g., "1.0.0").
	Version() string

	// Description returns a brief description of what the plugin does.
	Description() string

	// RegisterRules registers custom rules with the provided registry.
	// Plugins can add one or more rules to extend analysis capabilities.
	// Example:
	//   func (p *MyPlugin) RegisterRules(registry *RuleRegistry) {
	//     registry.MustRegister(&MyCustomRule{})
	//   }
	RegisterRules(registry *RuleRegistry)
}

// PluginInfo contains metadata about a loaded plugin
type PluginInfo struct {
	Name        string
	Version     string
	Description string
	Loaded      bool
	Error       error
}
