package model

import "time"

type Task struct {
	ID           int64      `db:"id" json:"id"`
	Type         string     `db:"type" json:"type"`
	Status       string     `db:"status" json:"status"`
	Payload      string     `db:"payload" json:"payload"`
	RetryCount   int        `db:"retry_count" json:"retry_count"`
	ErrorMessage string     `db:"error_message" json:"error_message,omitempty"`
	ScheduledAt  *time.Time `db:"scheduled_at" json:"scheduled_at,omitempty"`
	StartedAt    *time.Time `db:"started_at" json:"started_at,omitempty"`
	FinishedAt   *time.Time `db:"finished_at" json:"finished_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}
