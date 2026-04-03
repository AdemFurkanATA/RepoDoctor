package languages

import "testing"

func FuzzParsePythonImportLine_NoPanic(f *testing.F) {
	seeds := []string{
		"import os",
		"from .service import handler",
		"importlib.import_module(\"json\")",
		"__import__(\"pkg\" + name)",
		"from .. import *",
		"\x00\x00\x00",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, line string) {
		results := parsePythonImportLine(line)
		for _, ev := range results {
			if ev.level < 0 {
				t.Fatalf("relative level must not be negative: %d", ev.level)
			}
		}
	})
}

func FuzzParsePythonDynamicImport_NoPanic(f *testing.F) {
	seeds := []string{
		"importlib.import_module('json')",
		"importlib.import_module('.local')",
		"__import__('pkg')",
		"__import__(prefix + '.mod')",
		"not a dynamic import",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, line string) {
		_, _ = parsePythonDynamicImport(line)
	})
}
