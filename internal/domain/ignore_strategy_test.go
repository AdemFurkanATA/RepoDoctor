package domain_test

import (
	"testing"

	"RepoDoctor/internal/domain"
)

func TestDefaultIgnoreStrategy_ShouldIgnore(t *testing.T) {
	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)

	tests := []struct {
		name     string
		dir      string
		expected bool
	}{
		{name: "Python virtual env", dir: "venv", expected: true},
		{name: "Node modules", dir: "node_modules", expected: true},
		{name: "Git directory", dir: ".git", expected: true},
		{name: "Regular source directory", dir: "src", expected: false},
		{name: "Domain model directory", dir: "domain", expected: false},
		{name: "Go vendor", dir: "vendor", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.ShouldIgnore(tt.dir, tt.dir)
			if result != tt.expected {
				t.Errorf("ShouldIgnore(%q) = %v; expected %v", tt.dir, result, tt.expected)
			}
		})
	}
}

func BenchmarkDefaultIgnoreStrategy_ShouldIgnore(b *testing.B) {
	strategy := domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strategy.ShouldIgnore("node_modules", "node_modules")
		strategy.ShouldIgnore("src", "src")
	}
}
