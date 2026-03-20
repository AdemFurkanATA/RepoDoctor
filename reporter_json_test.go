package main

import (
	"strings"
	"testing"
)

func TestReporter_JSONV2_ContainsSchemaAndSummary(t *testing.T) {
	reporter := NewReporter(FormatJSON)
	report := &StructuralReport{
		Version:       "0.5.0-dev",
		SchemaVersion: "v2",
		Path:          "/repo/demo",
		Score: &StructuralScore{
			TotalScore: 95, MaxScore: 100,
		},
		Summary:  ReportSummary{TotalViolations: 1, Circular: 0, Layer: 0, Size: 1, GodObject: 0},
		Language: LanguageEvidenceSummary{DetectedLanguage: "Go", Confidence: 0.99},
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
}

func TestReporter_JSONV1_CompatibilitySwitch(t *testing.T) {
	reporter := NewReporter(FormatJSONV1)
	report := &StructuralReport{
		Version: "0.5.0-dev",
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
		Version: "0.5.0-dev",
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
		"  \"version\": \"0.5.0-dev\",\n" +
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
