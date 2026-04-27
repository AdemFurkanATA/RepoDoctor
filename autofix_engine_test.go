package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFixFlags_DryRunDefaultAndPacks(t *testing.T) {
	request, err := parseFixFlags([]string{"-path", "."})
	if err != nil {
		t.Fatalf("parseFixFlags failed: %v", err)
	}
	if request.Apply {
		t.Fatal("expected dry-run default (apply=false)")
	}
	if len(request.Packs) != 3 {
		t.Fatalf("expected three safe default packs, got %d", len(request.Packs))
	}
}

func TestRunAutoFix_DryRunDoesNotModifyFiles(t *testing.T) {
	repo := t.TempDir()
	file := filepath.Join(repo, "sample.go")
	original := "package demo\n\nimport \"fmt\"\n\nfunc wrap(err error) error {\n\treturn fmt.Errorf(\"oops: %v\", err)\n}\n"
	if err := os.WriteFile(file, []byte(original), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	result, err := runAutoFix(AutoFixRequest{RepositoryPath: repo, Apply: false, Packs: []AutoFixPack{AutoFixPackErrorWrap}})
	if err != nil {
		t.Fatalf("runAutoFix failed: %v", err)
	}
	if result.FilesChanged == 0 || len(result.Preview) == 0 {
		t.Fatalf("expected dry-run changes and preview, got %+v", result)
	}

	after, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if string(after) != original {
		t.Fatal("expected dry-run not to modify file")
	}
}

func TestRunAutoFix_ApplyWritesSafeSubsetChanges(t *testing.T) {
	repo := t.TempDir()
	file := filepath.Join(repo, "sample.go")
	original := "package demo\n\nimport \"fmt\"\n\nfunc wrap(err error) error {\n\treturn fmt.Errorf(\"oops: %v\", err)\n}\n"
	if err := os.WriteFile(file, []byte(original), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	result, err := runAutoFix(AutoFixRequest{RepositoryPath: repo, Apply: true, Packs: []AutoFixPack{AutoFixPackErrorWrap}})
	if err != nil {
		t.Fatalf("runAutoFix failed: %v", err)
	}
	if result.DryRun {
		t.Fatal("expected apply mode")
	}

	after, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if !strings.Contains(string(after), "%w") {
		t.Fatalf("expected wrapped error conversion in file, got %q", string(after))
	}
}
