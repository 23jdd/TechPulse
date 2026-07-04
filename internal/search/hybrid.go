package search

import (
	"context"
	"math"
	"sort"

	"techpulse/internal/ai"
)

type HybridEngine struct {
	base     SearchEngine
	provider ai.Provider
}

func NewHybridEngine(base SearchEngine, provider ai.Provider) *HybridEngine {
	return &HybridEngine{base: base, provider: provider}
}

func (h *HybridEngine) IndexArticle(ctx context.Context, article ArticleSearchDocument) error {
	return h.base.IndexArticle(ctx, article)
}

func (h *HybridEngine) DeleteArticle(ctx context.Context, articleID int64) error {
	return h.base.DeleteArticle(ctx, articleID)
}

func (h *HybridEngine) Search(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	result, err := h.base.Search(ctx, query)
	if err != nil || h.provider == nil || query.Query == "" || len(result.Hits) < 2 {
		return result, err
	}
	qv, err := h.provider.GenerateEmbedding(ctx, query.Query)
	if err != nil {
		return result, nil
	}
	for i := range result.Hits {
		text := result.Hits[i].Title + " " + result.Hits[i].Summary
		hv, err := h.provider.GenerateEmbedding(ctx, text)
		if err != nil {
			continue
		}
		result.Hits[i].Score = result.Hits[i].Score + cosine(qv, hv)
	}
	sort.SliceStable(result.Hits, func(i, j int) bool {
		return result.Hits[i].Score > result.Hits[j].Score
	})
	return result, nil
}

func (h *HybridEngine) Close() error {
	return h.base.Close()
}

func cosine(a, b []float64) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}
	var dot, ma, mb float64
	for i := 0; i < n; i++ {
		dot += a[i] * b[i]
		ma += a[i] * a[i]
		mb += b[i] * b[i]
	}
	if ma == 0 || mb == 0 {
		return 0
	}
	return dot / (math.Sqrt(ma) * math.Sqrt(mb))
}
