package search

import "time"

type ArticleSearchDocument struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	SourceType  string    `json:"source_type"`
	SourceID    int64     `json:"source_id"`
	Author      string    `json:"author"`
	Content     string    `json:"content"`
	Summary     string    `json:"summary"`
	Tags        []string  `json:"tags"`
	PublishedAt time.Time `json:"published_at"`
}

type SearchQuery struct {
	Query    string     `json:"query"`
	Tag      string     `json:"tag"`
	Source   string     `json:"source"`
	SourceID int64      `json:"source_id"`
	Author   string     `json:"author"`
	DateFrom *time.Time `json:"date_from,omitempty"`
	DateTo   *time.Time `json:"date_to,omitempty"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type SearchHit struct {
	ID        int64               `json:"id"`
	Title     string              `json:"title"`
	URL       string              `json:"url"`
	Author    string              `json:"author"`
	Summary   string              `json:"summary"`
	Tags      []string            `json:"tags"`
	Score     float64             `json:"score"`
	Highlight map[string][]string `json:"highlight,omitempty"`
}

type SearchResult struct {
	Query    string      `json:"query"`
	Total    uint64      `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Hits     []SearchHit `json:"hits"`
}
