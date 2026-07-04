package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type RSSFetcher struct {
	client *http.Client
}

func NewRSSFetcher(client *http.Client) *RSSFetcher {
	return &RSSFetcher{client: client}
}

func (r *RSSFetcher) Name() string {
	return "rss"
}

func (r *RSSFetcher) Supports(sourceType string) bool {
	sourceType = strings.ToLower(sourceType)
	return sourceType == "rss" || sourceType == "atom" || sourceType == "jsonfeed"
}

func (r *RSSFetcher) Fetch(ctx context.Context, source Source) ([]FetchedItem, error) {
	parser := gofeed.NewParser()
	parser.Client = r.client
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "TechPulse/0.1 (+https://github.com/techpulse)")
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch rss %s: status %d", source.URL, resp.StatusCode)
	}
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	items := make([]FetchedItem, 0, len(feed.Items))
	for _, item := range feed.Items {
		content := firstNonEmpty(item.Content, item.Description)
		var published *time.Time
		if item.PublishedParsed != nil {
			published = item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			published = item.UpdatedParsed
		}
		author := ""
		if item.Author != nil {
			author = item.Author.Name
		}
		image := ""
		if item.Image != nil {
			image = item.Image.URL
		}
		items = append(items, FetchedItem{
			SourceID:    source.ID,
			SourceType:  "rss",
			Title:       firstNonEmpty(item.Title, feed.Title),
			URL:         item.Link,
			Author:      author,
			Content:     content,
			Description: item.Description,
			Image:       image,
			Categories:  item.Categories,
			PublishedAt: published,
		})
	}
	return items, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
