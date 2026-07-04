package service

import (
	"techpulse/internal/fetcher"
	"techpulse/internal/parser"
	"techpulse/internal/rag"
	"techpulse/internal/search"
)

type FetchRequest struct {
	Source fetcher.Source `json:"source"`
}

type FetchResponse struct {
	Items []fetcher.FetchedItem `json:"items"`
}

type ParseRequest struct {
	Item fetcher.FetchedItem `json:"item"`
}

type ParseResponse struct {
	Article *parser.ParsedArticle `json:"article"`
}

type ProcessRequest struct {
	Article *parser.ParsedArticle `json:"article"`
}

type ProcessResponse struct {
	Article     any       `json:"article"`
	Summary     any       `json:"summary"`
	Tags        []string  `json:"tags"`
	Keywords    []string  `json:"keywords"`
	Embedding   []float64 `json:"embedding"`
	Translation string    `json:"translation"`
}

type IndexRequest struct {
	Document search.ArticleSearchDocument `json:"document"`
}

type SearchRequest struct {
	Query search.SearchQuery `json:"query"`
}

type ChatRequest struct {
	Question string `json:"question"`
}

type ChatResponse = rag.Answer
