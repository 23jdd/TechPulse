package rag

import (
	"context"
	"strings"
	"testing"

	"techpulse/internal/ai"
	"techpulse/internal/search"
)

func TestRAGReturnsCitations(t *testing.T) {
	engine, err := search.NewBleveEngine(t.TempDir())
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	defer engine.Close()
	if err := engine.IndexArticle(context.Background(), search.ArticleSearchDocument{ID: 7, Title: "Go update", URL: "https://example.com/go", Content: "Go testing update", Summary: "Testing improved", Tags: []string{"Go"}}); err != nil {
		t.Fatalf("index: %v", err)
	}
	svc := NewService(NewRetriever(engine), NewGenerator(ai.NewMockProvider()))
	answer, err := svc.Ask(context.Background(), "Go testing")
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if len(answer.Citations) != 1 || answer.Citations[0].ArticleID != 7 {
		t.Fatalf("unexpected citations: %+v", answer.Citations)
	}
	if !strings.Contains(answer.Answer, "Mock answer") {
		t.Fatalf("unexpected answer: %s", answer.Answer)
	}
}
