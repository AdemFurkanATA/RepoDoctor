package rules

// DefaultRegistry is the global registry containing all built-in rules
var DefaultRegistry = NewRuleRegistry()

func init() {
	// Register all built-in rules
	DefaultRegistry.MustRegister(NewGodObjectRule())
	DefaultRegistry.MustRegister(NewSizeRule())
	DefaultRegistry.MustRegister(NewLayerValidationRule())
	DefaultRegistry.MustRegister(NewSecretDetectionRule())
	// Note: CircularDependencyRule requires a graph parameter, so it's registered separately
}

// GetDefaultRegistry returns the global default registry with all built-in rules
func GetDefaultRegistry() *RuleRegistry {
	return DefaultRegistry
}
