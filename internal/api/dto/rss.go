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
