package service

import (
	"net/http"
	"strings"

	"techpulse/internal/ai"
	"techpulse/internal/config"
)

func AIProvider(cfg config.Config, client *http.Client) ai.Provider {
	provider := strings.ToLower(strings.TrimSpace(cfg.AIProvider))
	var base ai.Provider
	switch provider {
	case "", "mock":
		base = ai.NewMockProvider()
	case "ollama":
		base = ai.NewOllamaCompatibleProvider(cfg.AIBaseURL, cfg.AIModel, client)
	default:
		if cfg.AIAPIKey == "" {
			base = ai.NewMockProvider()
			break
		}
		base = ai.NewOpenAICompatibleProvider(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel, client)
	}
	return ai.NewResilientProvider(base, cfg.RequestTimeout, 2)
}
