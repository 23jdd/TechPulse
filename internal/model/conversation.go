package model

import "time"

type Conversation struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Title     string    `db:"title" json:"title"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Message struct {
	ID             int64     `db:"id" json:"id"`
	ConversationID int64     `db:"conversation_id" json:"conversation_id"`
	Role           string    `db:"role" json:"role"`
	Content        string    `db:"content" json:"content"`
	Citations      string    `db:"citations" json:"citations"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
