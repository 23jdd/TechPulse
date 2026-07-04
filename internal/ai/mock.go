package ai

import (
	"context"
	"math"
	"strings"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) Name() string { return "mock" }

func (m *MockProvider) Summarize(_ context.Context, content string) (*Summary, error) {
	text := compact(content)
	if len(text) > 240 {
		text = text[:240]
	}
	if text == "" {
		text = "No article content was available."
	}
	return &Summary{
		OneSentence:  text,
		ShortSummary: text,
		LongSummary:  compact(content),
		BulletPoints: []string{"Key technical update extracted from the article.", "Review the citation for full context."},
		TLDR:         text,
	}, nil
}

func (m *MockProvider) Translate(_ context.Context, content, targetLanguage string) (string, error) {
	if targetLanguage == "" {
		targetLanguage = "zh-CN"
	}
	return "[" + targetLanguage + " mock translation] " + content, nil
}

func (m *MockProvider) GenerateTags(_ context.Context, content string) ([]string, error) {
	lower := strings.ToLower(content)
	tags := []string{"Tech"}
	for keyword, tag := range map[string]string{
		"go ": "Go", "golang": "Go", "kubernetes": "Kubernetes", "linux": "Linux",
		"database": "Database", "security": "Security", "ai": "AI", "cloud": "Cloud Native",
	} {
		if strings.Contains(lower, keyword) {
			tags = append(tags, tag)
		}
	}
	return unique(tags), nil
}

func (m *MockProvider) GenerateKeywords(_ context.Context, content string) ([]string, error) {
	words := strings.Fields(strings.ToLower(content))
	out := make([]string, 0, 8)
	seen := map[string]bool{}
	for _, word := range words {
		word = strings.Trim(word, ".,:;!?()[]{}\"'")
		if len(word) < 5 || seen[word] {
			continue
		}
		seen[word] = true
		out = append(out, word)
		if len(out) == 8 {
			break
		}
	}
	if len(out) == 0 {
		out = []string{"techpulse"}
	}
	return out, nil
}

func (m *MockProvider) GenerateEmbedding(_ context.Context, content string) ([]float64, error) {
	vector := make([]float64, 16)
	for i, r := range content {
		vector[i%len(vector)] += float64(r%97) / 97
	}
	var norm float64
	for _, v := range vector {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		return vector, nil
	}
	for i := range vector {
		vector[i] = vector[i] / norm
	}
	return vector, nil
}

func (m *MockProvider) ChatCompletion(_ context.Context, messages []ChatMessage) (string, error) {
	if len(messages) == 0 {
		return "I need a question before I can answer.", nil
	}
	last := messages[len(messages)-1].Content
	return "Mock answer based on retrieved TechPulse articles: " + compact(last), nil
}

func compact(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func unique(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
