package dto

type CreateRSSRequest struct {
	URL                  string `json:"url"`
	Title                string `json:"title"`
	Category             string `json:"category"`
	FetchIntervalMinutes int    `json:"fetch_interval_minutes"`
}

type UpdateRSSRequest struct {
	URL                  string `json:"url"`
	Title                string `json:"title"`
	Category             string `json:"category"`
	Status               string `json:"status"`
	FetchIntervalMinutes int    `json:"fetch_interval_minutes"`
}

type FetchRSSResponse struct {
	FeedID     int64    `json:"feed_id"`
	Fetched    int      `json:"fetched"`
	Inserted   int      `json:"inserted"`
	Duplicates int      `json:"duplicates"`
	Errors     []string `json:"errors,omitempty"`
}

type TestRSSResponse struct {
	FeedID   int64    `json:"feed_id,omitempty"`
	URL      string   `json:"url"`
	OK       bool     `json:"ok"`
	Fetched  int      `json:"fetched"`
	Title    string   `json:"title,omitempty"`
	Errors   []string `json:"errors,omitempty"`
	Duration string   `json:"duration"`
}

type FetchGitHubReleasesRequest struct {
	URL string `json:"url"`
}

type FetchSourceResponse struct {
	SourceType string   `json:"source_type"`
	SourceURL  string   `json:"source_url"`
	Fetched    int      `json:"fetched"`
	Inserted   int      `json:"inserted"`
	Duplicates int      `json:"duplicates"`
	Errors     []string `json:"errors,omitempty"`
}
