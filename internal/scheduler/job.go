package scheduler

import "time"

type Job struct {
	ID          int64
	Type        string
	Payload     []byte
	ScheduledAt time.Time
}
