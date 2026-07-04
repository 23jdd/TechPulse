package search

import (
	"context"

	"techpulse/internal/model"
)

func DocumentFromArticle(article model.Article, summary string, tags []string) ArticleSearchDocument {
	published := article.CreatedAt
	if article.PublishedAt != nil {
		published = *article.PublishedAt
	}
	return ArticleSearchDocument{
		ID:          article.ID,
		Title:       article.Title,
		URL:         article.URL,
		SourceType:  article.SourceType,
		SourceID:    article.SourceID,
		Author:      article.Author,
		Content:     article.CleanContent,
		Summary:     summary,
		Tags:        tags,
		PublishedAt: published,
	}
}

type Indexer struct {
	engine SearchEngine
}

func NewIndexer(engine SearchEngine) *Indexer {
	return &Indexer{engine: engine}
}

func (i *Indexer) Index(ctx context.Context, article model.Article, summary string, tags []string) error {
	return i.engine.IndexArticle(ctx, DocumentFromArticle(article, summary, tags))
}
