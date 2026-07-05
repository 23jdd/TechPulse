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
	HealthStatus  string     `db:"health_status" json:"health_status"`
	HealthScore   int        `db:"health_score" json:"health_score"`
	Failures      int        `db:"consecutive_failures" json:"consecutive_failures"`
	LastError     string     `db:"last_error" json:"last_error,omitempty"`
	LastDuration  int64      `db:"last_duration_ms" json:"last_duration_ms"`
	LastCheckedAt *time.Time `db:"last_checked_at" json:"last_checked_at,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}
