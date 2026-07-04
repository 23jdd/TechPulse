package fetcher

import (
	"context"
	"strings"
)

type GitHubReleaseFetcher struct{}

func (GitHubReleaseFetcher) Name() string { return "github_release" }
func (GitHubReleaseFetcher) Supports(sourceType string) bool {
	return strings.EqualFold(sourceType, "github")
}
func (GitHubReleaseFetcher) Fetch(context.Context, Source) ([]FetchedItem, error) {
	return nil, nil
}
