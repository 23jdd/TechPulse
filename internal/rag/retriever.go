package rag

import (
	"context"

	"techpulse/internal/search"
)

type Retriever struct {
	engine search.SearchEngine
}

func NewRetriever(engine search.SearchEngine) *Retriever {
	return &Retriever{engine: engine}
}

func (r *Retriever) Retrieve(ctx context.Context, question string, limit int) ([]search.SearchHit, error) {
	result, err := r.engine.Search(ctx, search.SearchQuery{Query: question, Page: 1, PageSize: limit})
	if err != nil {
		return nil, err
	}
	return result.Hits, nil
}
