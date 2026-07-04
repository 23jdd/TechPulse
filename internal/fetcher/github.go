package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type GitHubReleaseFetcher struct {
	client *http.Client
	token  string
}

func NewGitHubReleaseFetcher(client *http.Client, token string) *GitHubReleaseFetcher {
	return &GitHubReleaseFetcher{client: client, token: token}
}

func (g *GitHubReleaseFetcher) Name() string { return "github_release" }

func (g *GitHubReleaseFetcher) Supports(sourceType string) bool {
	sourceType = strings.ToLower(sourceType)
	return sourceType == "github" || sourceType == "github_release"
}

type githubRelease struct {
	HTMLURL     string    `json:"html_url"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Author      struct {
		Login string `json:"login"`
	} `json:"author"`
}

func (g *GitHubReleaseFetcher) Fetch(ctx context.Context, source Source) ([]FetchedItem, error) {
	client := g.client
	if client == nil {
		client = http.DefaultClient
	}
	apiURL, repo, err := githubReleasesURL(source.URL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "TechPulse/0.1")
	if g.token != "" {
		req.Header.Set("Authorization", "Bearer "+g.token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch github releases %s: status %d", source.URL, resp.StatusCode)
	}
	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	items := make([]FetchedItem, 0, len(releases))
	for _, release := range releases {
		if release.Draft {
			continue
		}
		title := release.Name
		if strings.TrimSpace(title) == "" {
			title = release.TagName
		}
		if repo != "" {
			title = repo + " " + title
		}
		published := release.PublishedAt
		content := release.Body
		if strings.TrimSpace(content) == "" {
			content = "GitHub release " + release.TagName
		}
		categories := []string{"GitHub", "Release"}
		if release.Prerelease {
			categories = append(categories, "Prerelease")
		}
		items = append(items, FetchedItem{
			SourceID:    source.ID,
			SourceType:  "github_release",
			Title:       title,
			URL:         release.HTMLURL,
			Author:      release.Author.Login,
			Content:     content,
			Description: release.TagName,
			Categories:  categories,
			PublishedAt: &published,
		})
	}
	return items, nil
}

func githubReleasesURL(raw string) (apiURL, repo string, err error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", err
	}
	if parsed.Host == "api.github.com" {
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) >= 4 && parts[0] == "repos" {
			repo = parts[1] + "/" + parts[2]
		}
		return raw, repo, nil
	}
	if parsed.Host != "github.com" {
		return raw, "", nil
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("github repository URL must look like https://github.com/{owner}/{repo}")
	}
	owner := parts[0]
	repoName := parts[1]
	repo = owner + "/" + repoName
	return "https://api.github.com/repos/" + path.Join(owner, repoName, "releases"), repo, nil
}
