package service

import (
	"net/http"
	"strings"

	"techpulse/internal/ai"
	"techpulse/internal/config"
)

func AIProvider(cfg config.Config, client *http.Client) ai.Provider {
	provider := strings.ToLower(strings.TrimSpace(cfg.AIProvider))
	switch provider {
	case "", "mock":
		return ai.NewMockProvider()
	case "ollama":
		return ai.NewOllamaCompatibleProvider(cfg.AIBaseURL, cfg.AIModel, client)
	default:
		if cfg.AIAPIKey == "" {
			return ai.NewMockProvider()
		}
		return ai.NewOpenAICompatibleProvider(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel, client)
	}
}
