package ai

import "context"

type Summary struct {
	OneSentence  string
	ShortSummary string
	LongSummary  string
	BulletPoints []string
	TLDR         string
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Provider interface {
	Name() string
	Summarize(context.Context, string) (*Summary, error)
	Translate(context.Context, string, string) (string, error)
	GenerateTags(context.Context, string) ([]string, error)
	GenerateKeywords(context.Context, string) ([]string, error)
	GenerateEmbedding(context.Context, string) ([]float64, error)
	ChatCompletion(context.Context, []ChatMessage) (string, error)
}
