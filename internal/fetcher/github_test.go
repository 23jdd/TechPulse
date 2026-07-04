package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGitHubReleaseFetcherFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Fatal("expected user agent")
		}
		_, _ = w.Write([]byte(`[
		  {
		    "html_url":"https://github.com/example/project/releases/tag/v1.0.0",
		    "tag_name":"v1.0.0",
		    "name":"v1.0.0",
		    "body":"Breaking changes and migration guide",
		    "draft":false,
		    "prerelease":false,
		    "published_at":"2026-07-04T08:00:00Z",
		    "author":{"login":"maintainer"}
		  }
		]`))
	}))
	defer server.Close()

	f := NewGitHubReleaseFetcher(server.Client(), "")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	items, err := f.Fetch(ctx, Source{Type: "github_release", URL: server.URL})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 release, got %d", len(items))
	}
	if items[0].Title != "v1.0.0" || items[0].Author != "maintainer" {
		t.Fatalf("unexpected item: %+v", items[0])
	}
	if items[0].SourceType != "github_release" {
		t.Fatalf("unexpected source type: %s", items[0].SourceType)
	}
}

func TestGitHubReleasesURLFromRepo(t *testing.T) {
	apiURL, repo, err := githubReleasesURL("https://github.com/golang/go")
	if err != nil {
		t.Fatalf("url: %v", err)
	}
	if repo != "golang/go" {
		t.Fatalf("repo = %q", repo)
	}
	if apiURL != "https://api.github.com/repos/golang/go/releases" {
		t.Fatalf("apiURL = %q", apiURL)
	}
}
