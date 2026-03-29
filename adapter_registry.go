package main

import "RepoDoctor/internal/languages"

func registerCoreAdapters(detector *languages.RepositoryLanguageDetector, config *Config) {
	detector.RegisterAdapter(languages.NewGoAdapter())
	detector.RegisterAdapter(languages.NewPythonAdapter())
	detector.RegisterAdapter(languages.NewJavaScriptAdapter())
	detector.RegisterAdapter(languages.NewTypeScriptAdapter())
	_ = isJavaPilotEnabled(config)
}

func isJavaPilotEnabled(config *Config) bool {
	if config == nil || config.LanguageDetection == nil || config.LanguageDetection.JavaPilot == nil {
		return false
	}
	return *config.LanguageDetection.JavaPilot
}
