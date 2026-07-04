package model

import "time"

type RSSFeed struct {
	ID            int64      `db:"id" json:"id"`
	UserID        int64      `db:"user_id" json:"user_id"`
	URL           string     `db:"url" json:"url"`
	Title         string     `db:"title" json:"title"`
	Category      string     `db:"category" json:"category"`
	Status        string     `db:"status" json:"status"`
	FetchInterval int        `db:"fetch_interval_minutes" json:"fetch_interval_minutes"`
	LastFetchedAt *time.Time `db:"last_fetched_at" json:"last_fetched_at,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}
