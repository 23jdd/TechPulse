package websocket

import "time"

type Event struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	ArticleID int64     `json:"article_id,omitempty"`
	TaskID    int64     `json:"task_id,omitempty"`
	Time      time.Time `json:"time"`
}
