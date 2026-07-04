package fetch

import (
	"time"

	"github.com/mmcdole/gofeed"
)

type Article struct {
	ID      int64
	Title   string
	Content string
	Summary string

	URL          string
	CanonicalURL string

	Author      string
	PublishedAt time.Time
	UpdatedAt   time.Time

	Tags       []string
	Categories []string

	Image string
}
type Worker struct {
	e *Extractor
}

func NewWorker() *Worker {
	return &Worker{NewExtractor()}
}
func (w *Worker) ItemToArticle(item *gofeed.Item) *Article {
	extract, err := w.e.Extract(item.Link, item.Content)
	if err != nil {
		return nil
	}
	return &Article{
		ID:           0,
		Title:        extract.Title,
		Content:      extract.Markdown,
		Summary:      "",
		URL:          extract.SourceURL,
		CanonicalURL: "",
		Author:       extract.Author,
		PublishedAt:  extract.PublishedAt,
		UpdatedAt:    time.Now(),
	}

}
