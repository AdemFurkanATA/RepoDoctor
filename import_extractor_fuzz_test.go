package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportExtractor_ExtractFromFile_MalformedFailSoft(t *testing.T) {
	extractor := NewImportExtractor("github.com/example/repo")
	dir := t.TempDir()
	malformed := filepath.Join(dir, "broken.go")

	if err := os.WriteFile(malformed, []byte("package broken\nimport (\n\t\"fmt\"\n"), 0o644); err != nil {
		t.Fatalf("failed writing malformed file: %v", err)
	}

	meta, err := extractor.ExtractFromFile(malformed)
	if err != nil {
		t.Fatalf("expected fail-soft nil error, got: %v", err)
	}
	if meta != nil {
		t.Fatalf("expected nil metadata for malformed file, got: %+v", meta)
	}
}

func TestImportExtractor_ExtractFromDir_SkipsMalformedKeepsValid(t *testing.T) {
	extractor := NewImportExtractor("github.com/example/repo")
	dir := t.TempDir()

	valid := filepath.Join(dir, "ok.go")
	invalid := filepath.Join(dir, "bad.go")

	if err := os.WriteFile(valid, []byte("package ok\nimport \"github.com/example/repo/internal/app\"\n"), 0o644); err != nil {
		t.Fatalf("failed writing valid file: %v", err)
	}
	if err := os.WriteFile(invalid, []byte("package bad\nimport (\n\t\"fmt\"\n"), 0o644); err != nil {
		t.Fatalf("failed writing malformed file: %v", err)
	}

	imports, err := extractor.ExtractFromDir(dir)
	if err != nil {
		t.Fatalf("expected fail-soft directory extraction, got error: %v", err)
	}
	if len(imports) != 1 {
		t.Fatalf("expected only valid file metadata, got %d entries", len(imports))
	}
	if meta := imports[valid]; meta == nil || meta.Package != "ok" {
		t.Fatalf("expected valid file metadata for %s, got %+v", valid, meta)
	}
}

func FuzzImportExtractor_ExtractFromFile_NoPanic(f *testing.F) {
	seeds := [][]byte{
		[]byte("package p\nimport \"fmt\"\n"),
		[]byte("package p\nimport (\"github.com/acme/x\";\"os\")\n"),
		[]byte("package p\nimport \"github.com/acme/x\n"),
		[]byte("\x00\x00\x00"),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, content []byte) {
		extractor := NewImportExtractor("github.com/acme/project")
		dir := t.TempDir()
		path := filepath.Join(dir, "input.go")

		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("failed writing fuzz input: %v", err)
		}

		meta, err := extractor.ExtractFromFile(path)
		if err != nil {
			t.Fatalf("expected fail-soft parse behavior, got error: %v", err)
		}
		if meta != nil && meta.Package == "" {
			t.Fatalf("unexpected empty package in metadata: %+v", meta)
		}
	})
}
