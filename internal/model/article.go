package model

import "time"

type Article struct {
	ID           int64      `db:"id" json:"id"`
	SourceType   string     `db:"source_type" json:"source_type"`
	SourceID     int64      `db:"source_id" json:"source_id"`
	Title        string     `db:"title" json:"title"`
	URL          string     `db:"url" json:"url"`
	URLHash      string     `db:"url_hash" json:"url_hash"`
	ContentHash  string     `db:"content_hash" json:"content_hash"`
	Author       string     `db:"author" json:"author"`
	Language     string     `db:"language" json:"language"`
	RawContent   string     `db:"raw_content" json:"raw_content,omitempty"`
	CleanContent string     `db:"clean_content" json:"clean_content,omitempty"`
	CoverImage   string     `db:"cover_image" json:"cover_image,omitempty"`
	PublishedAt  *time.Time `db:"published_at" json:"published_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}
