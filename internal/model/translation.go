package model

import "time"

type Translation struct {
	ID                int64     `db:"id" json:"id"`
	ArticleID         int64     `db:"article_id" json:"article_id"`
	TargetLanguage    string    `db:"target_language" json:"target_language"`
	TranslatedTitle   string    `db:"translated_title" json:"translated_title"`
	TranslatedContent string    `db:"translated_content" json:"translated_content"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}
