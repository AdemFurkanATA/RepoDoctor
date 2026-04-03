package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInteractiveIO_ReadChoiceAndReadPositiveInt(t *testing.T) {
	io := &interactiveIO{reader: bufio.NewReader(strings.NewReader("2\n42\n-1\n"))}

	if got := io.readChoice(); got != 2 {
		t.Fatalf("expected choice 2, got %d", got)
	}
	value, ok := io.readPositiveInt("Enter value")
	if !ok || value != 42 {
		t.Fatalf("expected positive int 42, got value=%d ok=%v", value, ok)
	}
	_, ok = io.readPositiveInt("Enter value")
	if ok {
		t.Fatal("expected negative number to be rejected")
	}
}

func TestInteractiveIO_ReadStringAndConfirm(t *testing.T) {
	io := &interactiveIO{reader: bufio.NewReader(strings.NewReader("  hello  \nyes\nno\n"))}

	if got := io.readString("Prompt: "); got != "hello" {
		t.Fatalf("expected trimmed readString result 'hello', got %q", got)
	}
	if !io.confirm("Confirm") {
		t.Fatal("expected 'yes' to confirm true")
	}
	if io.confirm("Confirm") {
		t.Fatal("expected 'no' to confirm false")
	}
}

func TestInteractiveConfigController_TogglesAndThresholds(t *testing.T) {
	cfg := (&ConfigLoader{}).getDefaultConfig()
	io := &interactiveIO{reader: bufio.NewReader(strings.NewReader("100\n50\n20\n15\n"))}
	controller := NewInteractiveConfigController(io)

	controller.toggleSizeRule(cfg)
	if *cfg.Rules.EnableSizeRule {
		t.Fatal("expected size rule toggle to disable")
	}
	controller.toggleGodObjectRule(cfg)
	if *cfg.Rules.EnableGodObjectRule {
		t.Fatal("expected god-object rule toggle to disable")
	}

	controller.setMaxFileLines(cfg)
	controller.setMaxFunctionLines(cfg)
	controller.setMaxFields(cfg)
	controller.setMaxMethods(cfg)

	if cfg.Size.MaxFileLines != 100 || cfg.Size.MaxFunctionLines != 50 {
		t.Fatalf("unexpected size thresholds: %+v", cfg.Size)
	}
	if cfg.GodObject.MaxFields != 20 || cfg.GodObject.MaxMethods != 15 {
		t.Fatalf("unexpected god-object thresholds: %+v", cfg.GodObject)
	}

	menu := captureStdout(t, func() {
		controller.showConfigMenu(cfg)
	})
	if !strings.Contains(menu, "Rule Configuration") {
		t.Fatalf("expected config menu output, got %q", menu)
	}
}

func TestSaveConfig_WritesConfigFile(t *testing.T) {
	tmp := t.TempDir()
	cfg := (&ConfigLoader{}).getDefaultConfig()

	if err := saveConfig(tmp, cfg); err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	configPath := GetConfigPath(tmp)
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file at %s, got error: %v", configPath, err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed reading saved config: %v", err)
	}
	if !strings.Contains(string(data), "size:") {
		t.Fatalf("expected yaml config content, got %q", string(data))
	}

	if !filepath.IsAbs(configPath) && strings.TrimSpace(configPath) == "" {
		t.Fatalf("unexpected config path value: %q", configPath)
	}
}

func TestBoolLabelValues(t *testing.T) {
	if got := boolLabel(true); got != "Enabled" {
		t.Fatalf("expected Enabled, got %q", got)
	}
	if got := boolLabel(false); got != "Disabled" {
		t.Fatalf("expected Disabled, got %q", got)
	}
}
