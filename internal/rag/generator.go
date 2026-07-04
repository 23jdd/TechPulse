package rag

import (
	"context"
	"strings"

	"techpulse/internal/ai"
	"techpulse/internal/model"
	"techpulse/internal/search"
)

type Generator struct {
	provider ai.Provider
}

func NewGenerator(provider ai.Provider) *Generator {
	return &Generator{provider: provider}
}

func (g *Generator) Generate(ctx context.Context, question string, hits []search.SearchHit) (string, error) {
	return g.GenerateWithHistory(ctx, question, hits, nil)
}

func (g *Generator) GenerateWithHistory(ctx context.Context, question string, hits []search.SearchHit, history []model.Message) (string, error) {
	var b strings.Builder
	if len(history) > 0 {
		b.WriteString("Recent conversation:\n")
		for _, msg := range history {
			b.WriteString(msg.Role)
			b.WriteString(": ")
			b.WriteString(msg.Content)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("Question: ")
	b.WriteString(question)
	b.WriteString("\nUse these cited articles:\n")
	for i, hit := range hits {
		b.WriteString("[")
		b.WriteString(string(rune('1' + i)))
		b.WriteString("] ")
		b.WriteString(hit.Title)
		b.WriteString(": ")
		b.WriteString(hit.Summary)
		b.WriteString("\n")
	}
	return g.provider.ChatCompletion(ctx, []ai.ChatMessage{
		{Role: "system", Content: "Answer as a concise developer intelligence analyst. Include citation references when useful."},
		{Role: "user", Content: b.String()},
	})
}
