package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRSSFetcherFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>TechPulse Feed</title>
  <entry>
    <title>Go testing</title>
    <link href="https://example.com/go-testing"/>
    <updated>2026-07-04T08:00:00Z</updated>
    <content type="html">&lt;p&gt;Go parser and RSS fetcher test&lt;/p&gt;</content>
  </entry>
</feed>`))
	}))
	defer server.Close()

	f := NewRSSFetcher(server.Client())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	items, err := f.Fetch(ctx, Source{ID: 1, Type: "rss", URL: server.URL})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Go testing" || items[0].URL == "" {
		t.Fatalf("unexpected item: %+v", items[0])
	}
}
