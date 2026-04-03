package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestReporter_JSONV2_ContainsSchemaAndSummary(t *testing.T) {
	reporter := NewReporter(FormatJSON)
	report := &StructuralReport{
		Version:       "1.1.0",
		SchemaVersion: "v2",
		Path:          "/repo/demo",
		Score: &StructuralScore{
			TotalScore: 95, MaxScore: 100,
		},
		Summary:  ReportSummary{TotalViolations: 1, Circular: 0, Layer: 0, Size: 1, GodObject: 0},
		Language: LanguageEvidenceSummary{DetectedLanguage: "Go", Confidence: 0.99, ReasonCodes: []string{"SCORING_ORDER_RESOLVED"}},
	}

	jsonOut := reporter.Format(report)
	if !strings.Contains(jsonOut, "\"schemaVersion\": \"v2\"") {
		t.Fatalf("expected v2 schema marker in output: %s", jsonOut)
	}
	if !strings.Contains(jsonOut, "\"summary\"") {
		t.Fatalf("expected summary section in output: %s", jsonOut)
	}
	if !strings.Contains(jsonOut, "\"language\"") {
		t.Fatalf("expected language section in output: %s", jsonOut)
	}
	if !strings.Contains(jsonOut, "\"reasonCodes\"") {
		t.Fatalf("expected reasonCodes in output: %s", jsonOut)
	}
}

func TestReporter_JSONV1_CompatibilitySwitch(t *testing.T) {
	reporter := NewReporter(FormatJSONV1)
	report := &StructuralReport{
		Version: "1.1.0",
		Path:    "/repo/demo",
		Score:   &StructuralScore{TotalScore: 90, MaxScore: 100},
	}

	jsonOut := reporter.Format(report)
	if strings.Contains(jsonOut, "schemaVersion") {
		t.Fatalf("v1 output must not include schemaVersion: %s", jsonOut)
	}
	if strings.Contains(jsonOut, "\"summary\"") {
		t.Fatalf("v1 output must not include summary section: %s", jsonOut)
	}
}

func TestReporter_JSONV1_GoldenParity(t *testing.T) {
	reporter := NewReporter(FormatJSONV1)
	report := &StructuralReport{
		Version: "1.1.0",
		Path:    "demo/path",
		Score: &StructuralScore{
			TotalScore:       90,
			MaxScore:         100,
			CircularPenalty:  0,
			LayerPenalty:     0,
			SizePenalty:      3,
			GodObjectPenalty: 5,
			CircularCount:    0,
			LayerCount:       0,
			SizeCount:        1,
			GodObjectCount:   1,
		},
	}

	got := reporter.Format(report)
	want := "{\n" +
		"  \"version\": \"1.1.0\",\n" +
		"  \"path\": \"demo/path\",\n" +
		"  \"score\": {\n" +
		"    \"total\": 90.00,\n" +
		"    \"max\": 100.00,\n" +
		"    \"circularPenalty\": 0.00,\n" +
		"    \"layerPenalty\": 0.00,\n" +
		"    \"sizePenalty\": 3.00,\n" +
		"    \"godObjectPenalty\": 5.00\n" +
		"  },\n" +
		"  \"violations\": {\n" +
		"    \"circular\": 0,\n" +
		"    \"layer\": 0,\n" +
		"    \"size\": 1,\n" +
		"    \"godObject\": 1\n" +
		"  },\n" +
		"  \"circularViolations\": [\n" +
		"  ],\n" +
		"  \"layerViolations\": [\n" +
		"  ],\n" +
		"  \"sizeViolations\": [\n" +
		"  ],\n" +
		"  \"godObjectViolations\": [\n" +
		"  ]\n" +
		"}\n"

	if got != want {
		t.Fatalf("json v1 golden mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestReporter_JSONV2_GoldenStableOrderingAndSchema(t *testing.T) {
	reporter := NewReporter(FormatJSON)
	report := &StructuralReport{
		Version:       "1.1.0",
		SchemaVersion: "v2",
		Path:          filepath.Join(".", "demo", "repo"),
		Score: &StructuralScore{
			TotalScore:       88,
			MaxScore:         100,
			CircularPenalty:  10,
			LayerPenalty:     0,
			SizePenalty:      2,
			GodObjectPenalty: 0,
		},
		Summary:  ReportSummary{TotalViolations: 2, Circular: 1, Layer: 0, Size: 1, GodObject: 0},
		Language: LanguageEvidenceSummary{DetectedLanguage: "Go", Confidence: 0.91},
		Circular: []CycleViolation{{Path: []string{"b", "a"}, Severity: "critical"}, {Path: []string{"a", "b"}, Severity: "critical"}},
		Size:     []SizeViolation{{File: "z.go", Function: "f", Lines: 100, Threshold: 80}, {File: "a.go", Function: "f", Lines: 90, Threshold: 80}},
	}

	out := reporter.Format(report)
	if !strings.Contains(out, "\"schemaVersion\": \"v2\"") {
		t.Fatalf("expected v2 schema marker in golden output: %s", out)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("output must be valid JSON: %v", err)
	}

	pathValue, ok := payload["path"].(string)
	if !ok {
		t.Fatalf("expected path value to be a string, got %T", payload["path"])
	}
	if strings.Contains(pathValue, "\\") {
		t.Fatalf("expected normalized slash path, got: %v", pathValue)
	}
}

func TestReporter_JSON_DoesNotLeakAbsolutePathByDefault(t *testing.T) {
	reporter := NewReporter(FormatJSON)
	abs := filepath.Join("C:\\", "tmp", "sensitive", "repo")

	report := &StructuralReport{
		Version:       "1.1.0",
		SchemaVersion: "v2",
		Path:          abs,
		Score:         &StructuralScore{TotalScore: 100, MaxScore: 100},
	}

	out := reporter.Format(report)
	if strings.Contains(out, abs) {
		t.Fatalf("absolute path leaked in JSON output: %s", out)
	}
}

func TestReporter_JSON_EscapesUntrustedFields(t *testing.T) {
	reporter := NewReporter(FormatJSON)
	malicious := "name\"with\ncontrol"

	report := &StructuralReport{
		Version:       "1.1.0",
		SchemaVersion: "v2",
		Path:          ".",
		Score:         &StructuralScore{TotalScore: 99, MaxScore: 100},
		Size: []SizeViolation{{
			File:      malicious,
			Function:  malicious,
			Lines:     1,
			Threshold: 1,
		}},
	}

	out := reporter.Format(report)
	if strings.Contains(out, "name\"with\ncontrol") {
		t.Fatalf("expected JSON escaping for untrusted fields: %s", out)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("escaped output must still be valid json: %v", err)
	}
}

func TestReporter_JSONV1_DeterministicAcrossInputOrdering(t *testing.T) {
	reporter := NewReporter(FormatJSONV1)
	build := func(circular []CycleViolation, layer []LayerViolation, size []SizeViolation, god []GodObjectViolation) *StructuralReport {
		return &StructuralReport{
			Version: "1.1.0",
			Path:    "repo",
			Score: &StructuralScore{
				TotalScore:       80,
				MaxScore:         100,
				CircularPenalty:  10,
				LayerPenalty:     5,
				SizePenalty:      3,
				GodObjectPenalty: 2,
				CircularCount:    len(circular),
				LayerCount:       len(layer),
				SizeCount:        len(size),
				GodObjectCount:   len(god),
			},
			Circular:  circular,
			Layer:     layer,
			Size:      size,
			GodObject: god,
		}
	}

	first := build(
		[]CycleViolation{{Path: []string{"pkg/z", "pkg/a"}, Severity: "critical"}, {Path: []string{"pkg/a", "pkg/b"}, Severity: "critical"}},
		[]LayerViolation{{From: "infrastructure", To: "domain", Message: "x"}, {From: "application", To: "domain", Message: "a"}},
		[]SizeViolation{{File: "z.go", Function: "b", Lines: 90, Threshold: 80}, {File: "a.go", Function: "a", Lines: 95, Threshold: 80}},
		[]GodObjectViolation{{File: "z.go", StructName: "Z", FieldCount: 20, MethodCount: 11}, {File: "a.go", StructName: "A", FieldCount: 18, MethodCount: 12}},
	)

	second := build(
		[]CycleViolation{{Path: []string{"pkg/a", "pkg/b"}, Severity: "critical"}, {Path: []string{"pkg/z", "pkg/a"}, Severity: "critical"}},
		[]LayerViolation{{From: "application", To: "domain", Message: "a"}, {From: "infrastructure", To: "domain", Message: "x"}},
		[]SizeViolation{{File: "a.go", Function: "a", Lines: 95, Threshold: 80}, {File: "z.go", Function: "b", Lines: 90, Threshold: 80}},
		[]GodObjectViolation{{File: "a.go", StructName: "A", FieldCount: 18, MethodCount: 12}, {File: "z.go", StructName: "Z", FieldCount: 20, MethodCount: 11}},
	)

	outA := reporter.Format(first)
	outB := reporter.Format(second)
	if outA != outB {
		t.Fatalf("json-v1 output must be deterministic regardless of input ordering\nA:\n%s\nB:\n%s", outA, outB)
	}

	if strings.Contains(outA, "schemaVersion") || strings.Contains(outA, "\"summary\"") || strings.Contains(outA, "\"language\"") {
		t.Fatalf("json-v1 compatibility contract broken, unexpected v2 fields in output: %s", outA)
	}
}

func TestReporter_JSONV1_WithViolations_StaysValidJSON(t *testing.T) {
	reporter := NewReporter(FormatJSONV1)
	report := &StructuralReport{
		Version: "1.1.0",
		Path:    "repo",
		Score: &StructuralScore{
			TotalScore:       92,
			MaxScore:         100,
			CircularPenalty:  0,
			LayerPenalty:     0,
			SizePenalty:      3,
			GodObjectPenalty: 5,
			CircularCount:    0,
			LayerCount:       0,
			SizeCount:        1,
			GodObjectCount:   1,
		},
		Size:      []SizeViolation{{File: "a.go", Function: "f", Lines: 90, Threshold: 80}},
		GodObject: []GodObjectViolation{{File: "a.go", StructName: "Svc", FieldCount: 20, MethodCount: 11}},
	}

	out := reporter.Format(report)
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("json-v1 output must remain valid JSON: %v\noutput:\n%s", err, out)
	}

	if _, hasSchema := payload["schemaVersion"]; hasSchema {
		t.Fatalf("json-v1 output must not include schemaVersion: %s", out)
	}
}
