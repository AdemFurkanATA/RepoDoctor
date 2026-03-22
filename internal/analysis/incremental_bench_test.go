package analysis

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func BenchmarkBuildGoFingerprintMap_LargeRepoFixture(b *testing.B) {
	repo := b.TempDir()
	seedLargeGoFixtureRepo(b, repo, 1500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := BuildGoFingerprintMap(repo); err != nil {
			b.Fatalf("BuildGoFingerprintMap failed: %v", err)
		}
	}
}

func BenchmarkDiffFingerprintStates_LargeMap(b *testing.B) {
	prev := make(map[string]string, 5000)
	next := make(map[string]string, 5000)
	for i := 0; i < 5000; i++ {
		path := "pkg/file_" + strconv.Itoa(i) + ".go"
		prev[path] = "hash-prev-" + strconv.Itoa(i)
		next[path] = "hash-next-" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DiffFingerprintStates(prev, next)
	}
}

func seedLargeGoFixtureRepo(b *testing.B, repo string, files int) {
	b.Helper()
	for i := 0; i < files; i++ {
		sub := filepath.Join(repo, "pkg", strconv.Itoa(i%25))
		if err := os.MkdirAll(sub, 0o755); err != nil {
			b.Fatalf("failed creating fixture subdir: %v", err)
		}
		path := filepath.Join(sub, "file_"+strconv.Itoa(i)+".go")
		content := "package p" + strconv.Itoa(i%25) + "\nfunc f" + strconv.Itoa(i) + "() int { return " + strconv.Itoa(i) + " }\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			b.Fatalf("failed writing fixture file: %v", err)
		}
	}
}
