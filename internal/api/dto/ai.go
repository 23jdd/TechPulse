package dto

type TextRequest struct {
	Text           string `json:"text"`
	TargetLanguage string `json:"target_language"`
}

type ChatRequest struct {
	Question       string `json:"question"`
	ConversationID int64  `json:"conversation_id"`
}

type PromptRequest struct {
	Name      string `json:"name"`
	Content   string `json:"content"`
	IsDefault bool   `json:"is_default"`
}

type ReportRequest struct {
	Title     string `json:"title"`
	SendEmail bool   `json:"send_email"`
	EmailTo   string `json:"email_to"`
}

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
