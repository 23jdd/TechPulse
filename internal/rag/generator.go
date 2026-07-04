package rag

import (
	"context"
	"strings"

	"techpulse/internal/ai"
	"techpulse/internal/search"
)

type Generator struct {
	provider ai.Provider
}

func NewGenerator(provider ai.Provider) *Generator {
	return &Generator{provider: provider}
}

func (g *Generator) Generate(ctx context.Context, question string, hits []search.SearchHit) (string, error) {
	var b strings.Builder
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
