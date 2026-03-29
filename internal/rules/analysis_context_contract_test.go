package rules

import (
	"reflect"
	"testing"
)

func TestAnalysisContextContract_VersionLocked(t *testing.T) {
	if AnalysisContextContractVersion != "v1" {
		t.Fatalf("analysis context contract version drifted: got %q want %q", AnalysisContextContractVersion, "v1")
	}
}

func TestAnalysisContextContract_ShapeLocked(t *testing.T) {
	ctxType := reflect.TypeOf(AnalysisContext{})

	expectedFields := []struct {
		name string
		typ  reflect.Type
	}{
		{name: "RepositoryFiles", typ: reflect.TypeOf([]RepositoryFile{})},
		{name: "RepositoryMetrics", typ: reflect.TypeOf(RepositoryMetrics{})},
		{name: "DependencyGraph", typ: reflect.TypeOf(DependencyGraph{})},
		{name: "Configuration", typ: reflect.TypeOf(Configuration{})},
		{name: "Languages", typ: reflect.TypeOf([]string{})},
	}

	if ctxType.NumField() != len(expectedFields) {
		t.Fatalf("analysis context shape drifted: got %d fields want %d", ctxType.NumField(), len(expectedFields))
	}

	for index, expected := range expectedFields {
		field := ctxType.Field(index)
		if field.Name != expected.name {
			t.Fatalf("analysis context field[%d] name drifted: got %q want %q", index, field.Name, expected.name)
		}
		if field.Type != expected.typ {
			t.Fatalf("analysis context field[%d] type drifted: got %s want %s", index, field.Type, expected.typ)
		}
	}
}
