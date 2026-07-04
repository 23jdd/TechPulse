package parser

import (
	"context"
	"time"

	"techpulse/internal/fetcher"
)

type ParsedArticle struct {
	SourceID     int64
	SourceType   string
	Title        string
	URL          string
	Author       string
	Language     string
	RawContent   string
	CleanContent string
	CoverImage   string
	Tags         []string
	PublishedAt  *time.Time
}

type Parser interface {
	Parse(context.Context, fetcher.FetchedItem) (*ParsedArticle, error)
}

type Service struct {
	rss Parser
}

func NewService() *Service {
	return &Service{rss: RSSParser{}}
}

func (s *Service) Parse(ctx context.Context, item fetcher.FetchedItem) (*ParsedArticle, error) {
	return s.rss.Parse(ctx, item)
}
