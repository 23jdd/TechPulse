package model

import "time"

type DailyReport struct {
	ID         int64     `db:"id" json:"id"`
	UserID     int64     `db:"user_id" json:"user_id"`
	Title      string    `db:"title" json:"title"`
	Content    string    `db:"content" json:"content"`
	ReportDate time.Time `db:"report_date" json:"report_date"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
