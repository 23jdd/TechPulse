package model

import "time"

type User struct {
	ID        int64     `db:"id" json:"id"`
	GitHubID  string    `db:"github_id" json:"github_id"`
	Username  string    `db:"username" json:"username"`
	Email     string    `db:"email" json:"email"`
	AvatarURL string    `db:"avatar_url" json:"avatar_url"`
	APIToken  string    `db:"api_token" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
