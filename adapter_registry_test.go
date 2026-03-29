package main

import (
	"testing"

	"RepoDoctor/internal/domain"
	"RepoDoctor/internal/languages"
)

func TestIsJavaPilotEnabled_DefaultFalse(t *testing.T) {
	if isJavaPilotEnabled(nil) {
		t.Fatal("nil config must not enable java pilot")
	}

	if isJavaPilotEnabled(&Config{LanguageDetection: &LanguageDetectionConfig{}}) {
		t.Fatal("missing java pilot flag must default to false")
	}
}

func TestIsJavaPilotEnabled_ExplicitTrue(t *testing.T) {
	enabled := true
	cfg := &Config{LanguageDetection: &LanguageDetectionConfig{JavaPilot: &enabled}}
	if !isJavaPilotEnabled(cfg) {
		t.Fatal("explicit java pilot flag must be honored")
	}
}

func TestRegisterCoreAdapters_JavaPilotGate(t *testing.T) {
	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)
	baseDetector := languages.NewRepositoryLanguageDetector(strategy)
	registerCoreAdapters(baseDetector, &Config{LanguageDetection: &LanguageDetectionConfig{}})
	if got := len(baseDetector.GetSupportedLanguages()); got != 4 {
		t.Fatalf("expected 4 languages when java pilot disabled, got %d", got)
	}

	detector := languages.NewRepositoryLanguageDetector(strategy)
	enabled := true
	cfg := &Config{LanguageDetection: &LanguageDetectionConfig{JavaPilot: &enabled}}
	registerCoreAdapters(detector, cfg)

	supported := detector.GetSupportedLanguages()
	if len(supported) != 5 {
		t.Fatalf("expected Java to be registered when pilot enabled, got %v", supported)
	}
}
