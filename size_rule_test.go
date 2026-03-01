package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSizeRule_DefaultThresholds(t *testing.T) {
	rule := NewSizeRule()
	
	if rule.MaxFileLines != 500 {
		t.Errorf("Expected MaxFileLines to be 500, got %d", rule.MaxFileLines)
	}
	
	if rule.MaxFunctionLines != 80 {
		t.Errorf("Expected MaxFunctionLines to be 80, got %d", rule.MaxFunctionLines)
	}
}

func TestSizeRule_DetectLargeFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	
	// Create a large Go file (600 lines)
	largeFile := filepath.Join(tmpDir, "large.go")
	content := "package test\n\n"
	for i := 0; i < 600; i++ {
		content += "var dummy" + string(rune('a'+i%26)) + " = " + string(rune('0'+i%10)) + "\n"
	}
	
	err := os.WriteFile(largeFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	rule := NewSizeRule()
	err = rule.Check(tmpDir)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	
	violations := rule.Violations()
	if len(violations) == 0 {
		t.Error("Expected at least one violation for large file")
	}
	
	// Check that file violation was detected
	foundFileViolation := false
	for _, v := range violations {
		if v.Function == "" && v.Lines > 500 {
			foundFileViolation = true
			if v.Threshold != 500 {
				t.Errorf("Expected threshold 500, got %d", v.Threshold)
			}
		}
	}
	
	if !foundFileViolation {
		t.Error("Expected file size violation not found")
	}
}

func TestSizeRule_DetectLargeFunction(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	
	// Create a Go file with large function (100 lines)
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

func largeFunction() {
`
	for i := 0; i < 100; i++ {
		content += "    _ = " + string(rune('a'+i%26)) + "\n"
	}
	content += "}\n"
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	rule := NewSizeRule()
	err = rule.Check(tmpDir)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	
	violations := rule.Violations()
	
	// Check that function violation was detected
	foundFunctionViolation := false
	for _, v := range violations {
		if v.Function == "largeFunction" && v.Lines > 80 {
			foundFunctionViolation = true
			if v.Threshold != 80 {
				t.Errorf("Expected threshold 80, got %d", v.Threshold)
			}
		}
	}
	
	if !foundFunctionViolation {
		t.Error("Expected function size violation not found")
		t.Logf("All violations: %+v", violations)
	}
}

func TestSizeRule_NoViolationsForSmallFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	
	// Create a small Go file
	smallFile := filepath.Join(tmpDir, "small.go")
	content := `package test

func smallFunction() {
    x := 1
    y := 2
    _ = x + y
}
`
	
	err := os.WriteFile(smallFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	rule := NewSizeRule()
	err = rule.Check(tmpDir)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	
	violations := rule.Violations()
	if len(violations) != 0 {
		t.Errorf("Expected no violations for small file, got %d", len(violations))
	}
}

func TestSizeRule_SkipsHiddenFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	
	// Create a hidden Go file with many lines
	hiddenFile := filepath.Join(tmpDir, ".hidden.go")
	content := "package test\n\n"
	for i := 0; i < 600; i++ {
		content += "var dummy" + string(rune('a'+i%26)) + " = " + string(rune('0'+i%10)) + "\n"
	}
	
	err := os.WriteFile(hiddenFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	rule := NewSizeRule()
	err = rule.Check(tmpDir)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	
	violations := rule.Violations()
	if len(violations) != 0 {
		t.Errorf("Expected no violations for hidden file, got %d", len(violations))
	}
}

func TestSizeRule_HasCriticalViolations(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Initially no violations
	rule := NewSizeRule()
	if rule.HasCriticalViolations() {
		t.Error("Expected no violations initially")
	}
	
	// Create a large file
	largeFile := filepath.Join(tmpDir, "large.go")
	content := "package test\n\n"
	for i := 0; i < 600; i++ {
		content += "var dummy" + string(rune('a'+i%26)) + " = " + string(rune('0'+i%10)) + "\n"
	}
	
	err := os.WriteFile(largeFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	err = rule.Check(tmpDir)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	
	if !rule.HasCriticalViolations() {
		t.Error("Expected violations after checking large file")
	}
}
