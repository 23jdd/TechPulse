package rag

type Citation struct {
	ArticleID int64  `json:"article_id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
}
