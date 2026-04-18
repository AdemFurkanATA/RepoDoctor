package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectCyclomaticComplexitySummary_BandsDeterministic(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package demo\nfunc low(){if true {}}\nfunc med(){if true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\n}\n"), 0o644); err != nil {
		t.Fatalf("write a.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.go"), []byte("package demo\nfunc high(){if true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\nif true {}\n}\n"), 0o644); err != nil {
		t.Fatalf("write b.go: %v", err)
	}

	first := collectCyclomaticComplexitySummary(root)
	second := collectCyclomaticComplexitySummary(root)
	if first != second {
		t.Fatalf("expected deterministic summary, got %v vs %v", first, second)
	}
	if first.Low == 0 || first.Medium == 0 || first.High == 0 {
		t.Fatalf("expected all bands to be populated, got %+v", first)
	}
}
