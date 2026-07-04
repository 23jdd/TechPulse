package search

import (
	"context"
	"testing"

	"techpulse/internal/ai"
)

func TestHybridSearchDelegatesAndRanks(t *testing.T) {
	base, err := NewBleveEngine(t.TempDir())
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	defer base.Close()
	engine := NewHybridEngine(base, ai.NewMockProvider())
	_ = engine.IndexArticle(context.Background(), ArticleSearchDocument{ID: 1, Title: "Go testing", Summary: "toolchain"})
	_ = engine.IndexArticle(context.Background(), ArticleSearchDocument{ID: 2, Title: "Kubernetes", Summary: "cluster"})
	result, err := engine.Search(context.Background(), SearchQuery{Query: "Go", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if result.Total == 0 {
		t.Fatal("expected results")
	}
}
