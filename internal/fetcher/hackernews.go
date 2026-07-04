package fetcher

import (
	"context"
	"strings"
)

type HackerNewsFetcher struct{}

func (HackerNewsFetcher) Name() string { return "hackernews" }
func (HackerNewsFetcher) Supports(sourceType string) bool {
	return strings.EqualFold(sourceType, "hackernews")
}
func (HackerNewsFetcher) Fetch(context.Context, Source) ([]FetchedItem, error) {
	return nil, nil
}
