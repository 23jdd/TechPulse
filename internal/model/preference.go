package model

import "time"

type UserPreference struct {
	UserID             int64     `db:"user_id" json:"user_id"`
	InterestedTags     string    `db:"interested_tags" json:"interested_tags"`
	DailyReportTime    string    `db:"daily_report_time" json:"daily_report_time"`
	DailyReportEmail   string    `db:"daily_report_email" json:"daily_report_email"`
	DailyReportEnabled bool      `db:"daily_report_enabled" json:"daily_report_enabled"`
	Timezone           string    `db:"timezone" json:"timezone"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}
