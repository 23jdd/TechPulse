package search

import "context"

type SearchEngine interface {
	IndexArticle(context.Context, ArticleSearchDocument) error
	DeleteArticle(context.Context, int64) error
	Search(context.Context, SearchQuery) (*SearchResult, error)
	Close() error
}
