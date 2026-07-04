package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OpenAICompatibleProvider struct {
	BaseURL        string
	APIKey         string
	Model          string
	EmbeddingModel string
	Client         *http.Client
}

func NewOpenAICompatibleProvider(baseURL, apiKey, model string, client *http.Client) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{BaseURL: baseURL, APIKey: apiKey, Model: model, EmbeddingModel: model, Client: client}
}

func (p *OpenAICompatibleProvider) Name() string { return "openai-compatible" }

func (p *OpenAICompatibleProvider) Summarize(ctx context.Context, content string) (*Summary, error) {
	answer, err := p.ChatCompletion(ctx, []ChatMessage{{Role: "user", Content: "Summarize this technical article:\n" + content}})
	if err != nil {
		return nil, err
	}
	return &Summary{OneSentence: answer, ShortSummary: answer, LongSummary: answer, TLDR: answer}, nil
}

func (p *OpenAICompatibleProvider) Translate(ctx context.Context, content, targetLanguage string) (string, error) {
	return p.chat(ctx, "Translate to "+targetLanguage+":\n"+content)
}

func (p *OpenAICompatibleProvider) GenerateTags(ctx context.Context, content string) ([]string, error) {
	answer, err := p.chat(ctx, "Return 3-6 comma separated technical tags for:\n"+content)
	if err != nil {
		return nil, err
	}
	return splitCSV(answer), nil
}

func (p *OpenAICompatibleProvider) GenerateKeywords(ctx context.Context, content string) ([]string, error) {
	answer, err := p.chat(ctx, "Return 5-10 comma separated technical keywords for:\n"+content)
	if err != nil {
		return nil, err
	}
	return splitCSV(answer), nil
}

func (p *OpenAICompatibleProvider) GenerateEmbedding(ctx context.Context, content string) ([]float64, error) {
	reqBody := map[string]any{"model": p.EmbeddingModel, "input": content}
	raw, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/embeddings", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding request failed: status %d", resp.StatusCode)
	}
	var decoded struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	if len(decoded.Data) == 0 {
		return nil, fmt.Errorf("openai-compatible embedding response had no vectors")
	}
	return decoded.Data[0].Embedding, nil
}

func (p *OpenAICompatibleProvider) ChatCompletion(ctx context.Context, messages []ChatMessage) (string, error) {
	reqBody := map[string]any{"model": p.Model, "messages": messages}
	raw, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	resp, err := p.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var decoded struct {
		Choices []struct {
			Message ChatMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("openai-compatible response had no choices")
	}
	return decoded.Choices[0].Message.Content, nil
}

func (p *OpenAICompatibleProvider) chat(ctx context.Context, prompt string) (string, error) {
	return p.ChatCompletion(ctx, []ChatMessage{{Role: "user", Content: prompt}})
}
