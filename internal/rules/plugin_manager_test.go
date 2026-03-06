package rules

import (
	"os"
	"testing"

	"RepoDoctor/internal/model"
)

// MockPlugin for testing
type MockPlugin struct {
	name        string
	version     string
	description string
	rules       []Rule
}

func (m *MockPlugin) Name() string        { return m.name }
func (m *MockPlugin) Version() string     { return m.version }
func (m *MockPlugin) Description() string { return m.description }
func (m *MockPlugin) RegisterRules(registry *RuleRegistry) {
	for _, rule := range m.rules {
		registry.MustRegister(rule)
	}
}

// MockRule for testing
type MockRule struct {
	id       string
	category string
	severity string
}

func (m *MockRule) ID() string       { return m.id }
func (m *MockRule) Category() string { return m.category }
func (m *MockRule) Severity() string { return m.severity }
func (m *MockRule) Evaluate(ctx AnalysisContext) []model.Violation {
	return []model.Violation{}
}

func TestPluginManager_NewPluginManager(t *testing.T) {
	registry := NewRuleRegistry()
	pm := NewPluginManager("/tmp/plugins", registry)

	if pm == nil {
		t.Fatal("expected plugin manager, got nil")
	}

	if pm.pluginsDir != "/tmp/plugins" {
		t.Errorf("expected pluginsDir '/tmp/plugins', got '%s'", pm.pluginsDir)
	}

	if pm.GetPluginCount() != 0 {
		t.Errorf("expected 0 plugins, got %d", pm.GetPluginCount())
	}
}

func TestPluginManager_RegisterPlugin(t *testing.T) {
	registry := NewRuleRegistry()
	pm := NewPluginManager("/tmp/plugins", registry)

	mockRule := &MockRule{
		id:       "mock.rule",
		category: "testing",
		severity: "info",
	}

	mockPlugin := &MockPlugin{
		name:        "TestPlugin",
		version:     "1.0.0",
		description: "Test plugin for testing",
		rules:       []Rule{mockRule},
	}

	err := pm.RegisterPlugin(mockPlugin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pm.GetPluginCount() != 1 {
		t.Errorf("expected 1 plugin, got %d", pm.GetPluginCount())
	}

	// Verify rule was registered
	if registry.Count() != 1 {
		t.Errorf("expected 1 rule, got %d", registry.Count())
	}
}

func TestPluginManager_GetPluginInfo(t *testing.T) {
	registry := NewRuleRegistry()
	pm := NewPluginManager("/tmp/plugins", registry)

	mockPlugin := &MockPlugin{
		name:        "InfoPlugin",
		version:     "2.0.0",
		description: "Info test plugin",
		rules:       []Rule{},
	}

	err := pm.RegisterPlugin(mockPlugin)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	infos := pm.GetPluginInfo()
	if len(infos) != 1 {
		t.Errorf("expected 1 plugin info, got %d", len(infos))
	}

	info := infos[0]
	if info.Name != "InfoPlugin" {
		t.Errorf("expected name 'InfoPlugin', got '%s'", info.Name)
	}
	if info.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got '%s'", info.Version)
	}
	if info.Description != "Info test plugin" {
		t.Errorf("expected description 'Info test plugin', got '%s'", info.Description)
	}
	if !info.Loaded {
		t.Error("expected plugin to be loaded")
	}
}

func TestPluginManager_DiscoverAndLoad_NoDirectory(t *testing.T) {
	registry := NewRuleRegistry()
	pm := NewPluginManager("/nonexistent/path/plugins", registry)

	err := pm.DiscoverAndLoad()
	if err != nil {
		t.Fatalf("unexpected error when plugins dir doesn't exist: %v", err)
	}

	if pm.GetPluginCount() != 0 {
		t.Errorf("expected 0 plugins, got %d", pm.GetPluginCount())
	}
}

func TestPluginManager_DiscoverAndLoad_EmptyDirectory(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "plugins-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	registry := NewRuleRegistry()
	pm := NewPluginManager(tmpDir, registry)

	err = pm.DiscoverAndLoad()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pm.GetPluginCount() != 0 {
		t.Errorf("expected 0 plugins, got %d", pm.GetPluginCount())
	}
}

func TestPluginManager_MultiplePlugins(t *testing.T) {
	registry := NewRuleRegistry()
	pm := NewPluginManager("/tmp/plugins", registry)

	plugin1 := &MockPlugin{
		name:        "Plugin1",
		version:     "1.0.0",
		description: "First plugin",
		rules: []Rule{
			&MockRule{id: "rule1", category: "testing", severity: "info"},
		},
	}

	plugin2 := &MockPlugin{
		name:        "Plugin2",
		version:     "2.0.0",
		description: "Second plugin",
		rules: []Rule{
			&MockRule{id: "rule2", category: "testing", severity: "warning"},
			&MockRule{id: "rule3", category: "testing", severity: "error"},
		},
	}

	if err := pm.RegisterPlugin(plugin1); err != nil {
		t.Fatalf("failed to register plugin1: %v", err)
	}
	if err := pm.RegisterPlugin(plugin2); err != nil {
		t.Fatalf("failed to register plugin2: %v", err)
	}

	if pm.GetPluginCount() != 2 {
		t.Errorf("expected 2 plugins, got %d", pm.GetPluginCount())
	}

	if registry.Count() != 3 {
		t.Errorf("expected 3 rules, got %d", registry.Count())
	}
}

func TestPluginInfo_Structure(t *testing.T) {
	info := PluginInfo{
		Name:        "Test",
		Version:     "1.0.0",
		Description: "Test description",
		Loaded:      true,
		Error:       nil,
	}

	if info.Name != "Test" {
		t.Errorf("expected name 'Test', got '%s'", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", info.Version)
	}
	if info.Description != "Test description" {
		t.Errorf("expected description 'Test description', got '%s'", info.Description)
	}
	if !info.Loaded {
		t.Error("expected loaded to be true")
	}
}
