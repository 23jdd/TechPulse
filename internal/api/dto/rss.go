package dto

type CreateRSSRequest struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

type FetchRSSResponse struct {
	FeedID     int64    `json:"feed_id"`
	Fetched    int      `json:"fetched"`
	Inserted   int      `json:"inserted"`
	Duplicates int      `json:"duplicates"`
	Errors     []string `json:"errors,omitempty"`
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
