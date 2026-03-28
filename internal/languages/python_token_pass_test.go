package languages

import "testing"

func TestCollectPythonSignalsBounded_IgnoresCommentsAndDocstrings(t *testing.T) {
	content := []byte(`# comment
"""module docstring
still docstring
"""
import os
from collections import defaultdict

class Service:
    def run(self):
        return 1

async def handle():
    return 2
`)

	counts := collectPythonSignalsBounded(content)
	if counts.imports != 2 {
		t.Fatalf("expected 2 imports, got %d", counts.imports)
	}
	if counts.funcs != 2 {
		t.Fatalf("expected 2 funcs, got %d", counts.funcs)
	}
	if counts.classes != 1 {
		t.Fatalf("expected 1 class, got %d", counts.classes)
	}
}
