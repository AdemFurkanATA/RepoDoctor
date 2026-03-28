package main

import "testing"

func TestLayerValidationRule_Message_UsesDecimalIndices(t *testing.T) {
	rule := &LayerValidationRule{
		violations: []LayerViolation{{Message: "v1"}, {Message: "v2"}, {Message: "v3"}},
	}
	msg := rule.Message()
	if !(contains(msg, "[1]") && contains(msg, "[2]") && contains(msg, "[3]")) {
		t.Fatalf("expected decimal violation indices in message, got %q", msg)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
