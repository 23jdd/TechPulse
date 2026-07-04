package service

import (
	"net/http"
	"strings"

	"techpulse/internal/ai"
	"techpulse/internal/config"
)

func AIProvider(cfg config.Config, client *http.Client) ai.Provider {
	if strings.EqualFold(cfg.AIProvider, "mock") || cfg.AIAPIKey == "" {
		return ai.NewMockProvider()
	}
	return ai.NewOpenAICompatibleProvider(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel, client)
}
