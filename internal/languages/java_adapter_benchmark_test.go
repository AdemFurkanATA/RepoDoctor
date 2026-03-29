package languages

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func BenchmarkJavaAdapter_DetectCollectGraph(b *testing.B) {
	repo := b.TempDir()
	seedJavaBenchmarkFixture(b, repo, 400)
	adapter := NewJavaAdapter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files, err := adapter.DetectFiles(repo)
		if err != nil {
			b.Fatalf("DetectFiles failed: %v", err)
		}
		if _, err := adapter.CollectMetrics(files); err != nil {
			b.Fatalf("CollectMetrics failed: %v", err)
		}
		if _, err := adapter.BuildDependencyGraph(files); err != nil {
			b.Fatalf("BuildDependencyGraph failed: %v", err)
		}
	}
}

func seedJavaBenchmarkFixture(b *testing.B, repo string, files int) {
	b.Helper()
	for i := 0; i < files; i++ {
		dir := filepath.Join(repo, "src", "main", "java", "pkg", strconv.Itoa(i%20))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatalf("failed creating fixture dir: %v", err)
		}
		path := filepath.Join(dir, "Class"+strconv.Itoa(i)+".java")
		content := "package pkg." + strconv.Itoa(i%20) + ";\n" +
			"import java.util.List;\n" +
			"public class Class" + strconv.Itoa(i) + " {\n" +
			"  public int value(" + "int a, int b" + ") { return a + b + " + strconv.Itoa(i) + "; }\n" +
			"}\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			b.Fatalf("failed writing fixture file: %v", err)
		}
	}
}
