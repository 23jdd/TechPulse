package ai

import (
	"context"
	"testing"
)

func TestMockProvider(t *testing.T) {
	p := NewMockProvider()
	summary, err := p.Summarize(context.Background(), "Go adds useful testing and toolchain updates.")
	if err != nil {
		t.Fatalf("summarize: %v", err)
	}
	if summary.TLDR == "" {
		t.Fatal("expected tldr")
	}
	tags, err := p.GenerateTags(context.Background(), "Golang and Kubernetes")
	if err != nil {
		t.Fatalf("tags: %v", err)
	}
	if len(tags) == 0 {
		t.Fatal("expected tags")
	}
	embedding, err := p.GenerateEmbedding(context.Background(), "hello")
	if err != nil || len(embedding) != 16 {
		t.Fatalf("embedding len=%d err=%v", len(embedding), err)
	}
}
