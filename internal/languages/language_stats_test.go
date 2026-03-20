package languages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"RepoDoctor/internal/domain"
)

func TestRepositoryLanguageDetector_GetLanguageStats_SortedAndStable(t *testing.T) {
	repo := t.TempDir()

	if err := os.MkdirAll(filepath.Join(repo, "src"), 0o755); err != nil {
		t.Fatalf("failed creating src dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "tools"), 0o755); err != nil {
		t.Fatalf("failed creating tools dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repo, "src", "app.py"), []byte("print('x')\n"), 0o644); err != nil {
		t.Fatalf("failed writing python file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "tools", "cli.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed writing go file: %v", err)
	}

	detector := NewRepositoryLanguageDetector(domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs))
	detector.RegisterAdapter(NewGoAdapter())
	detector.RegisterAdapter(NewPythonAdapter())

	first, err := detector.GetLanguageStats(repo)
	if err != nil {
		t.Fatalf("GetLanguageStats failed: %v", err)
	}
	if len(first) != 2 {
		t.Fatalf("expected 2 language stats entries, got %d", len(first))
	}
	if first[0].Language != "Go" || first[1].Language != "Python" {
		t.Fatalf("expected stable ascending language sort [Go, Python], got [%s, %s]", first[0].Language, first[1].Language)
	}

	for i := 0; i < 10; i++ {
		next, nextErr := detector.GetLanguageStats(repo)
		if nextErr != nil {
			t.Fatalf("GetLanguageStats failed on iteration %d: %v", i, nextErr)
		}
		if len(next) != len(first) {
			t.Fatalf("length changed on iteration %d: %d vs %d", i, len(next), len(first))
		}
		for idx := range first {
			if next[idx] != first[idx] {
				t.Fatalf("stats changed on iteration %d index %d: %+v vs %+v", i, idx, next[idx], first[idx])
			}
		}
	}
}

func TestRepositoryLanguageDetector_GetLanguageStats_SkipsOutsideSymlinkLoop(t *testing.T) {
	repo := t.TempDir()
	outside := t.TempDir()

	if err := os.WriteFile(filepath.Join(repo, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed writing root go file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outside, "outside.py"), []byte("print('x')\n"), 0o644); err != nil {
		t.Fatalf("failed writing outside python file: %v", err)
	}

	if err := os.Symlink(outside, filepath.Join(repo, "linked_outside")); err != nil {
		t.Skipf("symlink not supported in this environment: %v", err)
	}

	detector := NewRepositoryLanguageDetector(domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs))
	detector.RegisterAdapter(NewGoAdapter())
	detector.RegisterAdapter(NewPythonAdapter())

	stats, err := detector.GetLanguageStats(repo)
	if err != nil {
		t.Fatalf("GetLanguageStats failed: %v", err)
	}

	if len(stats) != 1 || stats[0].Language != "Go" {
		t.Fatalf("expected outside symlink to be skipped and only Go stat kept, got %+v", stats)
	}
}

func TestCountLines_StreamedMemoryBoundedBehavior(t *testing.T) {
	repo := t.TempDir()
	path := filepath.Join(repo, "large.go")

	lines := strings.Repeat("package main\n", 10000)
	if err := os.WriteFile(path, []byte(lines), 0o644); err != nil {
		t.Fatalf("failed writing large fixture: %v", err)
	}

	count, err := countLines(path)
	if err != nil {
		t.Fatalf("countLines failed: %v", err)
	}
	if count != 10000 {
		t.Fatalf("expected streamed line count 10000, got %d", count)
	}
}

func TestRepositoryLanguageDetector_IsMultiLanguageRepository_Parity(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, "pkg"), 0o755); err != nil {
		t.Fatalf("failed creating package dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repo, "pkg", "a.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatalf("failed writing go file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "pkg", "b.py"), []byte("print('ok')\n"), 0o644); err != nil {
		t.Fatalf("failed writing python file: %v", err)
	}

	detector := NewRepositoryLanguageDetector(domain.NewDefaultIgnoreStrategy(domain.DefaultIgnoredDirs))
	detector.RegisterAdapter(NewGoAdapter())
	detector.RegisterAdapter(NewPythonAdapter())

	multi, langs, err := detector.IsMultiLanguageRepository(repo)
	if err != nil {
		t.Fatalf("IsMultiLanguageRepository failed: %v", err)
	}
	if !multi {
		t.Fatal("expected repository to be multi-language")
	}
	if len(langs) != 2 || langs[0] != "Go" || langs[1] != "Python" {
		t.Fatalf("expected deterministic language list [Go, Python], got %v", langs)
	}
}
