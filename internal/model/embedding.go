package model

import "time"

type Embedding struct {
	ID        int64     `db:"id" json:"id"`
	ArticleID int64     `db:"article_id" json:"article_id"`
	Provider  string    `db:"provider" json:"provider"`
	Model     string    `db:"model" json:"model"`
	Vector    string    `db:"vector" json:"vector"`
	Dimension int       `db:"dimension" json:"dimension"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
