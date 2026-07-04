package fetcher

import (
	"context"
	"strings"
)

type ArxivFetcher struct{}

func (ArxivFetcher) Name() string { return "arxiv" }
func (ArxivFetcher) Supports(sourceType string) bool {
	return strings.EqualFold(sourceType, "arxiv")
}
func (ArxivFetcher) Fetch(context.Context, Source) ([]FetchedItem, error) {
	return nil, nil
}
