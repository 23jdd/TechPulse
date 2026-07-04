package search

import (
	"context"
	"testing"
	"time"
)

func TestBleveSearch(t *testing.T) {
	engine, err := NewBleveEngine(t.TempDir())
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	defer engine.Close()
	err = engine.IndexArticle(context.Background(), ArticleSearchDocument{
		ID: 1, Title: "Go release", URL: "https://example.com", Content: "Golang toolchain update",
		Summary: "Go release summary", Tags: []string{"Go"}, PublishedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	result, err := engine.Search(context.Background(), SearchQuery{Query: "toolchain", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if result.Total != 1 || result.Hits[0].ID != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func BenchmarkBleveSearch(b *testing.B) {
	engine, err := NewBleveEngine(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer engine.Close()
	_ = engine.IndexArticle(context.Background(), ArticleSearchDocument{ID: 1, Title: "Go release", Content: "Golang toolchain update"})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Search(context.Background(), SearchQuery{Query: "go", Page: 1, PageSize: 10})
	}
}
