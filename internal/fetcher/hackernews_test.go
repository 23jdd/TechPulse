package fetcher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHackerNewsFetcherFetchesStoriesAndComments(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/topstories.json", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode([]int64{1})
	})
	mux.HandleFunc("/item/1.json", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          1,
			"type":        "story",
			"by":          "alice",
			"time":        1720000000,
			"title":       "Go runtime gets faster",
			"url":         "https://example.com/go",
			"text":        "Story body",
			"score":       123,
			"descendants": 7,
			"kids":        []int64{2},
		})
	})
	mux.HandleFunc("/item/2.json", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":   2,
			"type": "comment",
			"by":   "bob",
			"text": "Useful discussion",
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	fetcher := HackerNewsFetcher{client: server.Client(), baseURL: server.URL}
	items, err := fetcher.Fetch(context.Background(), Source{Type: "hackernews", URL: "top?limit=1"})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len=%d", len(items))
	}
	item := items[0]
	if item.SourceType != "hackernews" || item.Title != "Go runtime gets faster" || item.Author != "alice" {
		t.Fatalf("unexpected item: %+v", item)
	}
	if !strings.Contains(item.Content, "Useful discussion") {
		t.Fatalf("content did not include comment: %q", item.Content)
	}
}

func TestParseHackerNewsSource(t *testing.T) {
	feed, limit := parseHackerNewsSource("best:100")
	if feed != "best" || limit != 50 {
		t.Fatalf("feed=%s limit=%d", feed, limit)
	}
	feed, limit = parseHackerNewsSource("new?limit=3")
	if feed != "new" || limit != 3 {
		t.Fatalf("feed=%s limit=%d", feed, limit)
	}
}
