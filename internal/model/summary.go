package model

import "time"

type Summary struct {
	ID           int64     `db:"id" json:"id"`
	ArticleID    int64     `db:"article_id" json:"article_id"`
	OneSentence  string    `db:"one_sentence" json:"one_sentence"`
	ShortSummary string    `db:"short_summary" json:"short_summary"`
	LongSummary  string    `db:"long_summary" json:"long_summary"`
	BulletPoints string    `db:"bullet_points" json:"bullet_points"`
	TLDR         string    `db:"tldr" json:"tldr"`
	Language     string    `db:"language" json:"language"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
