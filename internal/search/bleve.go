package search

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	blevequery "github.com/blevesearch/bleve/v2/search/query"
)

type BleveEngine struct {
	index bleve.Index
}

func NewBleveEngine(path string) (*BleveEngine, error) {
	index, err := bleve.Open(path)
	if err != nil && (errors.Is(err, bleve.ErrorIndexPathDoesNotExist) || os.IsNotExist(err) || strings.Contains(err.Error(), "metadata missing")) {
		_ = os.RemoveAll(path)
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(path, mapping)
	}
	if err != nil {
		return nil, err
	}
	return &BleveEngine{index: index}, nil
}

func (b *BleveEngine) IndexArticle(_ context.Context, article ArticleSearchDocument) error {
	if article.PublishedAt.IsZero() {
		article.PublishedAt = time.Now()
	}
	return b.index.Index(strconv.FormatInt(article.ID, 10), article)
}

func (b *BleveEngine) DeleteArticle(_ context.Context, articleID int64) error {
	return b.index.Delete(strconv.FormatInt(articleID, 10))
}

func (b *BleveEngine) Search(_ context.Context, query SearchQuery) (*SearchResult, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	var q blevequery.Query
	if query.Query == "" {
		q = bleve.NewMatchAllQuery()
	} else {
		titleQuery := bleve.NewMatchQuery(query.Query)
		titleQuery.SetField("title")
		titleQuery.SetBoost(3)
		contentQuery := bleve.NewMatchQuery(query.Query)
		contentQuery.SetField("content")
		summaryQuery := bleve.NewMatchQuery(query.Query)
		summaryQuery.SetField("summary")
		tagsQuery := bleve.NewMatchQuery(query.Query)
		tagsQuery.SetField("tags")
		q = bleve.NewDisjunctionQuery(titleQuery, contentQuery, summaryQuery, tagsQuery)
	}
	conjuncts := []blevequery.Query{q}
	if query.Tag != "" {
		tagQuery := bleve.NewMatchQuery(query.Tag)
		tagQuery.SetField("tags")
		conjuncts = append(conjuncts, tagQuery)
	}
	if query.Author != "" {
		authorQuery := bleve.NewMatchQuery(query.Author)
		authorQuery.SetField("author")
		conjuncts = append(conjuncts, authorQuery)
	}
	if len(conjuncts) > 1 {
		q = bleve.NewConjunctionQuery(conjuncts...)
	}
	req := bleve.NewSearchRequestOptions(q, query.PageSize, (query.Page-1)*query.PageSize, false)
	req.Fields = []string{"id", "title", "url", "author", "summary", "tags"}
	req.Highlight = bleve.NewHighlight()
	res, err := b.index.Search(req)
	if err != nil {
		return nil, err
	}
	out := &SearchResult{Query: query.Query, Total: res.Total, Page: query.Page, PageSize: query.PageSize, Hits: make([]SearchHit, 0, len(res.Hits))}
	for _, hit := range res.Hits {
		id, _ := strconv.ParseInt(hit.ID, 10, 64)
		out.Hits = append(out.Hits, SearchHit{
			ID:        id,
			Title:     stringField(hit.Fields["title"]),
			URL:       stringField(hit.Fields["url"]),
			Author:    stringField(hit.Fields["author"]),
			Summary:   stringField(hit.Fields["summary"]),
			Tags:      stringSliceField(hit.Fields["tags"]),
			Score:     hit.Score,
			Highlight: hit.Fragments,
		})
	}
	return out, nil
}

func (b *BleveEngine) Close() error {
	return b.index.Close()
}

func stringField(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func stringSliceField(value any) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case string:
		return []string{v}
	default:
		return nil
	}
}
