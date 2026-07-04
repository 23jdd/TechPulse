package parser

import (
	"context"

	"techpulse/internal/fetcher"
)

type RSSParser struct{}

func (RSSParser) Parse(_ context.Context, item fetcher.FetchedItem) (*ParsedArticle, error) {
	clean := CleanHTML(item.Content)
	if clean == "" {
		clean = CleanHTML(item.Description)
	}
	return &ParsedArticle{
		SourceID:     item.SourceID,
		SourceType:   item.SourceType,
		Title:        item.Title,
		URL:          item.URL,
		Author:       item.Author,
		Language:     DetectLanguage(clean),
		RawContent:   item.Content,
		CleanContent: clean,
		CoverImage:   item.Image,
		Tags:         item.Categories,
		PublishedAt:  item.PublishedAt,
	}, nil
}
