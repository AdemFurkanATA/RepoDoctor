package main

import (
	"errors"
	"strings"
	"testing"
)

func TestExecuteCommand_VersionAndHelpCommands(t *testing.T) {
	versionOutput := captureStdout(t, func() {
		if err := executeCommand("version", nil); err != nil {
			t.Fatalf("version command failed: %v", err)
		}
	})
	if !strings.Contains(versionOutput, "RepoDoctor v1.1.0") {
		t.Fatalf("expected version output to include current version, got %q", versionOutput)
	}

	helpOutput := captureStdout(t, func() {
		if err := executeCommand("help", nil); err != nil {
			t.Fatalf("help command failed: %v", err)
		}
	})
	if !strings.Contains(helpOutput, "RepoDoctor - Static Architecture Intelligence") {
		t.Fatalf("expected help output header, got %q", helpOutput)
	}
}

func TestExecuteCommand_UnknownCommandReturnsCLIErrorWithSuggestion(t *testing.T) {
	err := executeCommand("ver", nil)
	if err == nil {
		t.Fatal("expected unknown command error")
	}

	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError type, got %T", err)
	}
	if cliErr.Category != ErrorCLIUsage {
		t.Fatalf("expected ErrorCLIUsage category, got %s", cliErr.Category)
	}
	if !strings.Contains(cliErr.Suggestion, "Did you mean 'version'?") {
		t.Fatalf("expected suggestion to include version hint, got %q", cliErr.Suggestion)
	}
}

func TestParseAnalyzeFlags_InvalidArgumentReturnsUsageError(t *testing.T) {
	_, err := parseAnalyzeFlags([]string{"-unknown-flag"})
	if err == nil {
		t.Fatal("expected parseAnalyzeFlags to return usage error")
	}

	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError type, got %T", err)
	}
	if cliErr.Category != ErrorCLIUsage {
		t.Fatalf("expected ErrorCLIUsage, got %s", cliErr.Category)
	}
}

func TestNormalizeAnalyzePathInput_EmptyRejected(t *testing.T) {
	_, err := normalizeAnalyzePathInput("   ")
	if err == nil {
		t.Fatal("expected empty analyze path to be rejected")
	}
}

func TestHasExplicitPathFlag_DetectionVariants(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "short flag", args: []string{"-path", "."}, want: true},
		{name: "long flag", args: []string{"--path", "."}, want: true},
		{name: "short equals", args: []string{"-path=."}, want: true},
		{name: "long equals", args: []string{"--path=."}, want: true},
		{name: "no flag", args: []string{"."}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasExplicitPathFlag(tc.args); got != tc.want {
				t.Fatalf("hasExplicitPathFlag(%v)=%v, want %v", tc.args, got, tc.want)
			}
		})
	}
}
