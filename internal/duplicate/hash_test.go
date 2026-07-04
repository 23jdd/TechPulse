package duplicate

import "testing"

func TestURLHashStable(t *testing.T) {
	if URLHash(" HTTPS://EXAMPLE.COM/a ") != URLHash("https://example.com/a") {
		t.Fatal("url hash should normalize case and whitespace")
	}
}

func TestContentHashStable(t *testing.T) {
	if ContentHash("hello") == ContentHash("world") {
		t.Fatal("different content should have different hashes")
	}
}

func BenchmarkContentHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ContentHash("Go Kubernetes Database Security AI")
	}
}
