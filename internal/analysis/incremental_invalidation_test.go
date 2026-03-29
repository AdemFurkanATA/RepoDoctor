package analysis

import "testing"

func TestDiffFingerprintStatesDetailed_TracksAddedModifiedRemovedAndRenamed(t *testing.T) {
	previous := map[string]string{
		"a.go": "hash-a",
		"b.go": "hash-b",
		"c.go": "hash-c",
	}
	current := map[string]string{
		"a.go":       "hash-a-mod",
		"renamed.go": "hash-b",
		"new.go":     "hash-new",
	}

	diff := DiffFingerprintStatesDetailed(previous, current)

	if len(diff.Modified) != 1 || diff.Modified[0] != "a.go" {
		t.Fatalf("expected modified [a.go], got %v", diff.Modified)
	}
	if len(diff.Added) != 2 || diff.Added[0] != "new.go" || diff.Added[1] != "renamed.go" {
		t.Fatalf("expected added [new.go renamed.go], got %v", diff.Added)
	}
	if len(diff.Removed) != 2 || diff.Removed[0] != "b.go" || diff.Removed[1] != "c.go" {
		t.Fatalf("expected removed [b.go c.go], got %v", diff.Removed)
	}
	if len(diff.Renamed) != 1 || diff.Renamed[0].From != "b.go" || diff.Renamed[0].To != "renamed.go" {
		t.Fatalf("expected rename b.go->renamed.go, got %v", diff.Renamed)
	}
}

func TestDiffFingerprintStates_BackwardCompatibilityChangedAndRemoved(t *testing.T) {
	previous := map[string]string{"same.go": "h1", "old.go": "h2"}
	current := map[string]string{"same.go": "h1-mod", "new.go": "h3"}

	changed, removed := DiffFingerprintStates(previous, current)
	if len(changed) != 2 || changed[0] != "new.go" || changed[1] != "same.go" {
		t.Fatalf("expected changed [new.go same.go], got %v", changed)
	}
	if len(removed) != 1 || removed[0] != "old.go" {
		t.Fatalf("expected removed [old.go], got %v", removed)
	}
}
