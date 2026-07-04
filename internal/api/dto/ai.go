package dto

type TextRequest struct {
	Text           string `json:"text"`
	TargetLanguage string `json:"target_language"`
}

type ChatRequest struct {
	Question       string `json:"question"`
	ConversationID int64  `json:"conversation_id"`
}
