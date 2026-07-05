package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultHackerNewsBaseURL = "https://hacker-news.firebaseio.com/v0"

type HackerNewsFetcher struct {
	client  *http.Client
	baseURL string
}

func NewHackerNewsFetcher(client *http.Client) *HackerNewsFetcher {
	return &HackerNewsFetcher{client: client, baseURL: defaultHackerNewsBaseURL}
}

func (HackerNewsFetcher) Name() string { return "hackernews" }
func (HackerNewsFetcher) Supports(sourceType string) bool {
	sourceType = strings.ToLower(strings.TrimSpace(sourceType))
	return sourceType == "hackernews" || sourceType == "hn"
}

type hackerNewsItem struct {
	ID          int64   `json:"id"`
	Type        string  `json:"type"`
	By          string  `json:"by"`
	Time        int64   `json:"time"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Text        string  `json:"text"`
	Score       int     `json:"score"`
	Descendants int     `json:"descendants"`
	Kids        []int64 `json:"kids"`
	Deleted     bool    `json:"deleted"`
	Dead        bool    `json:"dead"`
}

func (h HackerNewsFetcher) Fetch(ctx context.Context, source Source) ([]FetchedItem, error) {
	client := h.client
	if client == nil {
		client = http.DefaultClient
	}
	baseURL := strings.TrimRight(h.baseURL, "/")
	if baseURL == "" {
		baseURL = defaultHackerNewsBaseURL
	}
	feed, limit := parseHackerNewsSource(source.URL)
	endpoint := hackerNewsEndpoint(feed)
	ids, err := h.fetchIDs(ctx, client, baseURL, endpoint)
	if err != nil {
		return nil, err
	}
	if limit > len(ids) {
		limit = len(ids)
	}
	items := make([]FetchedItem, 0, limit)
	for _, id := range ids[:limit] {
		item, err := h.fetchItem(ctx, client, baseURL, id)
		if err != nil || item == nil || item.Deleted || item.Dead {
			continue
		}
		if item.Type != "story" && item.Type != "job" && item.Type != "poll" {
			continue
		}
		title := firstNonEmpty(item.Title, source.Title, "Hacker News item")
		link := item.URL
		if strings.TrimSpace(link) == "" {
			link = fmt.Sprintf("https://news.ycombinator.com/item?id=%d", item.ID)
		}
		published := time.Unix(item.Time, 0)
		commentText := h.fetchTopComments(ctx, client, baseURL, item.Kids, 3)
		content := strings.TrimSpace(item.Text)
		if commentText != "" {
			content = strings.TrimSpace(content + "\n\nTop comments:\n" + commentText)
		}
		if content == "" {
			content = fmt.Sprintf("%s\n\nHN score: %d, comments: %d", title, item.Score, item.Descendants)
		}
		items = append(items, FetchedItem{
			SourceID:    source.ID,
			SourceType:  "hackernews",
			Title:       title,
			URL:         link,
			Author:      item.By,
			Content:     content,
			Description: fmt.Sprintf("HN score: %d, comments: %d", item.Score, item.Descendants),
			Categories:  []string{"Hacker News", strings.Title(feed), item.Type},
			PublishedAt: &published,
		})
	}
	return items, nil
}

func (h HackerNewsFetcher) fetchIDs(ctx context.Context, client *http.Client, baseURL, endpoint string) ([]int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/"+endpoint+".json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "TechPulse/0.1")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch hacker news %s: status %d", endpoint, resp.StatusCode)
	}
	var ids []int64
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func (h HackerNewsFetcher) fetchItem(ctx context.Context, client *http.Client, baseURL string, id int64) (*hackerNewsItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/item/%d.json", baseURL, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "TechPulse/0.1")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch hacker news item %d: status %d", id, resp.StatusCode)
	}
	var item hackerNewsItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (h HackerNewsFetcher) fetchTopComments(ctx context.Context, client *http.Client, baseURL string, ids []int64, limit int) string {
	if limit > len(ids) {
		limit = len(ids)
	}
	var comments []string
	for _, id := range ids[:limit] {
		item, err := h.fetchItem(ctx, client, baseURL, id)
		if err != nil || item == nil || item.Deleted || item.Dead || strings.TrimSpace(item.Text) == "" {
			continue
		}
		comments = append(comments, "- "+item.Text)
	}
	return strings.Join(comments, "\n")
}

func parseHackerNewsSource(raw string) (string, int) {
	feed := "top"
	limit := 20
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return feed, limit
	}
	if parsed, err := url.Parse(raw); err == nil {
		if parsed.Scheme != "" && parsed.Host != "" {
			if parsed.Query().Get("feed") != "" {
				feed = parsed.Query().Get("feed")
			}
			if parsed.Query().Get("limit") != "" {
				if n, err := strconv.Atoi(parsed.Query().Get("limit")); err == nil {
					limit = n
				}
			}
			return normalizeHackerNewsFeed(feed), clampHackerNewsLimit(limit)
		}
		if parsed.Path != "" {
			feed = parsed.Path
		}
		if parsed.Query().Get("limit") != "" {
			if n, err := strconv.Atoi(parsed.Query().Get("limit")); err == nil {
				limit = n
			}
		}
	}
	parts := strings.Split(raw, ":")
	if len(parts) == 2 {
		feed = parts[0]
		if n, err := strconv.Atoi(parts[1]); err == nil {
			limit = n
		}
	} else if raw != "" && !strings.Contains(raw, "?") {
		feed = raw
	}
	return normalizeHackerNewsFeed(feed), clampHackerNewsLimit(limit)
}

func normalizeHackerNewsFeed(feed string) string {
	feed = strings.ToLower(strings.Trim(strings.TrimSpace(feed), "/"))
	switch feed {
	case "new", "newstories":
		return "new"
	case "best", "beststories":
		return "best"
	case "ask", "askstories":
		return "ask"
	case "show", "showstories":
		return "show"
	case "job", "jobs", "jobstories":
		return "job"
	default:
		return "top"
	}
}

func hackerNewsEndpoint(feed string) string {
	switch normalizeHackerNewsFeed(feed) {
	case "new":
		return "newstories"
	case "best":
		return "beststories"
	case "ask":
		return "askstories"
	case "show":
		return "showstories"
	case "job":
		return "jobstories"
	default:
		return "topstories"
	}
}

func clampHackerNewsLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 50 {
		return 50
	}
	return limit
}
