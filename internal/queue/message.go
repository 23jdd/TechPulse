package queue

type MessageType string

const (
	FetchJob       MessageType = "fetch_job"
	ParseJob       MessageType = "parse_job"
	AIJob          MessageType = "ai_job"
	IndexJob       MessageType = "index_job"
	DailyReportJob MessageType = "daily_report_job"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload []byte      `json:"payload"`
}
