package fetcher

import (
	"context"
	"time"
)

type Source struct {
	ID       int64
	Type     string
	URL      string
	Title    string
	Category string
}

type FetchedItem struct {
	SourceID    int64
	SourceType  string
	Title       string
	URL         string
	Author      string
	Content     string
	Description string
	Image       string
	Categories  []string
	PublishedAt *time.Time
}

type Fetcher interface {
	Name() string
	Supports(sourceType string) bool
	Fetch(ctx context.Context, source Source) ([]FetchedItem, error)
}

type Registry struct {
	fetchers []Fetcher
}

func NewRegistry(fetchers ...Fetcher) *Registry {
	return &Registry{fetchers: fetchers}
}

func (r *Registry) FetcherFor(sourceType string) Fetcher {
	for _, fetcher := range r.fetchers {
		if fetcher.Supports(sourceType) {
			return fetcher
		}
	}
	return nil
}
