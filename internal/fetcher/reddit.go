package fetcher

import (
	"context"
	"strings"
)

type RedditFetcher struct{}

func (RedditFetcher) Name() string { return "reddit" }
func (RedditFetcher) Supports(sourceType string) bool {
	return strings.EqualFold(sourceType, "reddit")
}
func (RedditFetcher) Fetch(context.Context, Source) ([]FetchedItem, error) {
	return nil, nil
}
