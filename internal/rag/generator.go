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
		{Role: "system", Content: "Answer as a concise developer intelligence analyst. Use only the cited articles. If the citations do not support an answer, say that the knowledge base does not contain enough evidence. Include citation references when useful."},
		{Role: "user", Content: b.String()},
	})
}

func (g *Generator) RewriteQuery(ctx context.Context, question string) string {
	question = strings.TrimSpace(question)
	if question == "" {
		return question
	}
	if g.provider.Name() == "mock" {
		return question
	}
	out, err := g.provider.ChatCompletion(ctx, []ai.ChatMessage{
		{Role: "system", Content: "Rewrite the user's question into one short search query for a technical article knowledge base. Return only the query."},
		{Role: "user", Content: question},
	})
	out = strings.TrimSpace(strings.Trim(out, "\"'`"))
	if err != nil || out == "" || len(out) > 160 {
		return question
	}
	return out
}
