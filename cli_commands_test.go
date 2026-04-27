package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReport_JSONAndJSONV1EmitRawJSON(t *testing.T) {
	tmp := t.TempDir()
	reportPath := filepath.Join(tmp, "report.json")
	raw := "{\"ok\":true}\n"
	if err := os.WriteFile(reportPath, []byte(raw), 0o644); err != nil {
		t.Fatalf("failed writing report fixture: %v", err)
	}

	for _, format := range []string{"json", "json-v1"} {
		t.Run(format, func(t *testing.T) {
			output := captureStdout(t, func() {
				if err := runReport(reportPath, format); err != nil {
					t.Fatalf("runReport failed for format %s: %v", format, err)
				}
			})
			if strings.TrimSpace(normalizeNewlines(output)) != strings.TrimSpace(raw) {
				t.Fatalf("expected raw JSON passthrough for %s, got %q", format, output)
			}
		})
	}
}

func TestRunReport_TextModeWrapsOutput(t *testing.T) {
	tmp := t.TempDir()
	reportPath := filepath.Join(tmp, "report.json")
	raw := "{\"score\":100}"
	if err := os.WriteFile(reportPath, []byte(raw), 0o644); err != nil {
		t.Fatalf("failed writing report fixture: %v", err)
	}

	output := captureStdout(t, func() {
		if err := runReport(reportPath, "text"); err != nil {
			t.Fatalf("runReport text mode failed: %v", err)
		}
	})
	if !strings.Contains(output, "RepoDoctor Analysis Report") || !strings.Contains(output, raw) {
		t.Fatalf("expected wrapped text report output, got %q", output)
	}
}

func TestRunReport_MissingFileReturnsError(t *testing.T) {
	err := runReport(filepath.Join(t.TempDir(), "missing.json"), "json")
	if err == nil {
		t.Fatal("expected missing report file error")
	}
}

func TestRunHistory_SucceedsWithoutHistoryFile(t *testing.T) {
	output := captureStdout(t, func() {
		if err := runHistory(t.TempDir()); err != nil {
			t.Fatalf("runHistory failed for empty history: %v", err)
		}
	})
	if !strings.Contains(output, "Score Trend History") {
		t.Fatalf("expected history header, got %q", output)
	}
}

func TestRunExtract_ErrorsForInvalidPathAndFilePath(t *testing.T) {
	if err := runExtract(filepath.Join(t.TempDir(), "missing"), "RepoDoctor", false, false); err == nil {
		t.Fatal("expected runExtract to fail for missing path")
	}

	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "file.go")
	if err := os.WriteFile(filePath, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed writing file fixture: %v", err)
	}
	if err := runExtract(filePath, "RepoDoctor", false, false); err == nil {
		t.Fatal("expected runExtract to fail for file path")
	}
}

func TestScanDirectory_CountsGoFilesAndSkipsDocsAndHidden(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "a.go"), []byte("package main\nfunc a(){}\n"), 0o644); err != nil {
		t.Fatalf("failed writing top go file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatalf("failed writing non-go file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "docs"), 0o755); err != nil {
		t.Fatalf("failed creating docs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "docs", "ignored.go"), []byte("package docs\n"), 0o644); err != nil {
		t.Fatalf("failed writing docs go file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatalf("failed creating hidden dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".git", "ignored.go"), []byte("package git\n"), 0o644); err != nil {
		t.Fatalf("failed writing hidden go file: %v", err)
	}

	totalFiles, goFiles, totalLines := scanDirectory(tmp, false)
	if totalFiles != 2 {
		t.Fatalf("expected total files 2 (a.go + b.txt), got %d", totalFiles)
	}
	if goFiles != 1 {
		t.Fatalf("expected go files 1, got %d", goFiles)
	}
	if totalLines < 2 {
		t.Fatalf("expected at least 2 lines counted for a.go, got %d", totalLines)
	}
}

func TestParseGenerateCIArgs_SupportedAndInvalidOptions(t *testing.T) {
	provider, force, err := parseGenerateCIArgs([]string{"github"})
	if err != nil {
		t.Fatalf("expected github provider to parse, got: %v", err)
	}
	if provider != "github" || force {
		t.Fatalf("unexpected parse result: provider=%s force=%v", provider, force)
	}

	provider, force, err = parseGenerateCIArgs([]string{"gitlab", "--force"})
	if err != nil {
		t.Fatalf("expected gitlab with force to parse, got: %v", err)
	}
	if provider != "gitlab" || !force {
		t.Fatalf("unexpected parse result with force: provider=%s force=%v", provider, force)
	}

	if _, _, err := parseGenerateCIArgs([]string{"azure", "--bad-flag"}); err == nil {
		t.Fatal("expected invalid flag to fail parse")
	}
}

func TestParseGenerateDocsArgs_ValidAndInvalid(t *testing.T) {
	format, err := parseGenerateDocsArgs([]string{"--format", "markdown"})
	if err != nil {
		t.Fatalf("expected markdown docs args to parse, got: %v", err)
	}
	if format != "markdown" {
		t.Fatalf("expected format markdown, got %s", format)
	}

	format, err = parseGenerateDocsArgs([]string{"--format=mermaid"})
	if err != nil {
		t.Fatalf("expected mermaid docs args to parse, got: %v", err)
	}
	if format != "mermaid" {
		t.Fatalf("expected format mermaid, got %s", format)
	}

	if _, err := parseGenerateDocsArgs([]string{"--format", "html"}); err == nil {
		t.Fatal("expected unsupported format to fail")
	}
}
