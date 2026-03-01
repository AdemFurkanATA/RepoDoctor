package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGodObjectRule_DefaultThresholds(t *testing.T) {
	rule := NewGodObjectRule()

	if rule.MaxFields != 15 {
		t.Errorf("Expected MaxFields to be 15, got %d", rule.MaxFields)
	}

	if rule.MaxMethods != 10 {
		t.Errorf("Expected MaxMethods to be 10, got %d", rule.MaxMethods)
	}
}

func TestGodObjectRule_DetectManyFields(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a Go file with struct that has many fields (20 fields)
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

type GodStruct struct {
`
	for i := 0; i < 20; i++ {
		content += "    Field" + string(rune('A'+i%26)) + " int\n"
	}
	content += "}\n"

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rule := NewGodObjectRule()
	err = rule.Check(tmpDir)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	violations := rule.Violations()

	// Check that god object violation was detected
	foundViolation := false
	for _, v := range violations {
		if v.StructName == "GodStruct" && v.FieldCount > 15 {
			foundViolation = true
		}
	}

	if !foundViolation {
		t.Error("Expected god object violation not found")
		t.Logf("All violations: %+v", violations)
	}
}

func TestGodObjectRule_DetectManyMethods(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a Go file with struct and many methods (15 methods)
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

type TestStruct struct {}

`
	for i := 0; i < 15; i++ {
		content += "func (t *TestStruct) Method" + string(rune('A'+i%26)) + "() {}\n"
	}

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rule := NewGodObjectRule()
	err = rule.Check(tmpDir)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	violations := rule.Violations()

	// Check that god object violation was detected
	foundViolation := false
	for _, v := range violations {
		if v.StructName == "TestStruct" && v.MethodCount > 10 {
			foundViolation = true
		}
	}

	if !foundViolation {
		t.Error("Expected god object violation not found")
		t.Logf("All violations: %+v", violations)
	}
}

func TestGodObjectRule_NoViolationsForNormalStructs(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a Go file with normal struct
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

type NormalStruct struct {
    Field1 int
    Field2 string
    Field3 bool
}

func (n *NormalStruct) Method1() {}
func (n *NormalStruct) Method2() {}
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rule := NewGodObjectRule()
	err = rule.Check(tmpDir)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	violations := rule.Violations()
	if len(violations) != 0 {
		t.Errorf("Expected no violations for normal struct, got %d", len(violations))
	}
}

func TestGodObjectRule_HasCriticalViolations(t *testing.T) {
	tmpDir := t.TempDir()

	// Initially no violations
	rule := NewGodObjectRule()
	if rule.HasCriticalViolations() {
		t.Error("Expected no violations initially")
	}

	// Create a god object
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test

type GodStruct struct {
`
	for i := 0; i < 20; i++ {
		content += "    Field" + string(rune('A'+i%26)) + " int\n"
	}
	content += "}\n"

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = rule.Check(tmpDir)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if !rule.HasCriticalViolations() {
		t.Error("Expected violations after checking god object")
	}
}

func TestGodObjectRule_SkipsHiddenFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a hidden Go file with god object
	hiddenFile := filepath.Join(tmpDir, ".hidden.go")
	content := `package test

type GodStruct struct {
`
	for i := 0; i < 20; i++ {
		content += "    Field" + string(rune('A'+i%26)) + " int\n"
	}
	content += "}\n"

	err := os.WriteFile(hiddenFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	rule := NewGodObjectRule()
	err = rule.Check(tmpDir)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	violations := rule.Violations()
	if len(violations) != 0 {
		t.Errorf("Expected no violations for hidden file, got %d", len(violations))
	}
}
