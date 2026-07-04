package fetcher

import (
	"context"
	"strings"
)

type YouTubeFetcher struct{}

func (YouTubeFetcher) Name() string { return "youtube" }
func (YouTubeFetcher) Supports(sourceType string) bool {
	return strings.EqualFold(sourceType, "youtube")
}
func (YouTubeFetcher) Fetch(context.Context, Source) ([]FetchedItem, error) {
	return nil, nil
}
