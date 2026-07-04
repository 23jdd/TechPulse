package parser

import "testing"

func TestCleanHTML(t *testing.T) {
	got := CleanHTML("<article><h1>Go</h1><p>Fast&nbsp;tests</p></article>")
	if got != "Go Fast tests" {
		t.Fatalf("unexpected clean text %q", got)
	}
}

func BenchmarkCleanHTML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = CleanHTML("<p>Go Kubernetes Database Security AI</p>")
	}
}
