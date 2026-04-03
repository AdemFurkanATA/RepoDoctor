package main

import (
	"errors"
	"strings"
	"testing"
)

func TestCLIError_ErrorAndDisplay(t *testing.T) {
	original := errors.New("boom")
	cliErr := NewCLIError(ErrorRuntime, "runtime failed", "try again", original)

	if !strings.Contains(cliErr.Error(), "runtime failed") {
		t.Fatalf("expected Error() to include message, got %q", cliErr.Error())
	}

	output := captureStderr(t, func() {
		cliErr.Display()
	})
	if !strings.Contains(output, "Suggestion") || !strings.Contains(output, "runtime failed") {
		t.Fatalf("expected display output with suggestion and message, got %q", output)
	}
}

func TestErrorHelpersAndSuggestionLookup(t *testing.T) {
	if got := GetSuggestion("file not found"); !strings.Contains(strings.ToLower(got), "path") {
		t.Fatalf("expected file-not-found suggestion, got %q", got)
	}
	if got := GetSuggestion("unmatched error"); !strings.Contains(got, "--help") {
		t.Fatalf("expected fallback help suggestion, got %q", got)
	}

	if err := HandleFileNotFoundError("/tmp/missing", errors.New("x")); err.Category != ErrorFileNotFound {
		t.Fatalf("expected file-not-found category, got %s", err.Category)
	}
	if err := HandleInvalidPathError("::bad", errors.New("x")); err.Category != ErrorInvalidArgument {
		t.Fatalf("expected invalid-argument category, got %s", err.Category)
	}
	if err := HandleConfigNotFoundError("cfg"); err.Category != ErrorConfiguration {
		t.Fatalf("expected config category, got %s", err.Category)
	}
	if err := HandleUnknownRuleError("foo", []string{"a", "b"}); err.Category != ErrorInvalidArgument {
		t.Fatalf("expected invalid-argument category for unknown rule, got %s", err.Category)
	}
	if err := HandleRuntimeError("oops", nil); err.Category != ErrorRuntime {
		t.Fatalf("expected runtime category, got %s", err.Category)
	}

	if msg := FormatErrorMessage(ErrorRuntime, "x"); !strings.Contains(msg, "Runtime Error") {
		t.Fatalf("unexpected formatted message: %q", msg)
	}
}

func TestPrintError_PrintsCLIAndGenericErrors(t *testing.T) {
	cliOut := captureStderr(t, func() {
		PrintError(NewCLIError(ErrorAnalysis, "bad", "fix", nil))
	})
	if !strings.Contains(cliOut, "bad") {
		t.Fatalf("expected CLI error output, got %q", cliOut)
	}

	genericOut := captureStderr(t, func() {
		PrintError(errors.New("generic"))
	})
	if !strings.Contains(genericOut, "Suggestion") {
		t.Fatalf("expected generic error suggestion output, got %q", genericOut)
	}
}
