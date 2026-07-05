package model

import "time"

type GitHubRepo struct {
	ID               int64      `db:"id" json:"id"`
	UserID           int64      `db:"user_id" json:"user_id"`
	Owner            string     `db:"owner" json:"owner"`
	Name             string     `db:"name" json:"name"`
	HTMLURL          string     `db:"html_url" json:"html_url"`
	Description      string     `db:"description" json:"description"`
	Stars            int64      `db:"stars" json:"stars"`
	OpenIssues       int64      `db:"open_issues" json:"open_issues"`
	DefaultBranch    string     `db:"default_branch" json:"default_branch"`
	LatestRelease    string     `db:"latest_release" json:"latest_release"`
	LatestReleaseURL string     `db:"latest_release_url" json:"latest_release_url"`
	LatestReleaseAt  *time.Time `db:"latest_release_at" json:"latest_release_at,omitempty"`
	BreakingChange   bool       `db:"breaking_change" json:"breaking_change"`
	SecurityUpdate   bool       `db:"security_update" json:"security_update"`
	LastCheckedAt    *time.Time `db:"last_checked_at" json:"last_checked_at,omitempty"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}
